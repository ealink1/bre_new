package services

import (
	"bre_new_backend/config"
	"bre_new_backend/models"
	"fmt"
	"sync"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	TotalPVKey     = "TOTAL_PV"
	FlushThreshold = 10
)

var (
	trafficMu      sync.Mutex
	pendingVisits  int64
	pendingIpStats = make(map[string]int64)
)

// RecordVisit 记录一次访问（线程安全）
func RecordVisit(ip string) {
	trafficMu.Lock()
	defer trafficMu.Unlock()

	pendingVisits++
	pendingIpStats[ip]++

	if pendingVisits >= FlushThreshold {
		go flushTrafficStats()
	}
}

// flushTrafficStats 将缓冲区的统计数据写入数据库
func flushTrafficStats() {
	trafficMu.Lock()
	if pendingVisits == 0 {
		trafficMu.Unlock()
		return
	}

	// 复制当前缓冲区数据，以便尽快释放锁
	visitsToFlush := pendingVisits
	ipsToFlush := make(map[string]int64)
	for ip, count := range pendingIpStats {
		ipsToFlush[ip] = count
	}

	// 重置缓冲区
	pendingVisits = 0
	pendingIpStats = make(map[string]int64)
	trafficMu.Unlock()

	// 执行数据库更新（使用 Upsert）
	db := config.DB
	if db == nil {
		fmt.Println("DB 未初始化，跳过流量统计写入")
		return
	}

	// 1. 更新总 PV
	err := incrementStat(db, TotalPVKey, visitsToFlush)
	if err != nil {
		fmt.Printf("更新总 PV 失败: %v\n", err)
	} else {
		fmt.Printf("已更新总 PV，增加: %d\n", visitsToFlush)
	}

	// 2. 更新各 IP 访问量
	for ip, count := range ipsToFlush {
		err := incrementStat(db, ip, count)
		if err != nil {
			fmt.Printf("更新 IP %s 统计失败: %v\n", ip, err)
		}
	}
}

// incrementStat 使用 GORM 的 Upsert 功能增加计数
func incrementStat(db *gorm.DB, ip string, delta int64) error {
	// 使用 OnConflict 能够处理并发写入（虽然这里我们是单协程 flush，但多实例部署时有用）
	// SQLite 支持 ON CONFLICT
	stat := models.TrafficStat{
		IP:    ip,
		Count: delta, // 这里的 Count 初始值，实际会被 OnConflict 覆盖逻辑处理
	}

	// GORM 的 Upsert 语法
	// 如果记录存在，则 Count = Count + delta
	return db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "ip"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"count":      gorm.Expr("count + ?", delta),
			"updated_at": time.Now(),
		}),
	}).Create(&stat).Error
}

// GetTotalVisits 获取总访问量
func GetTotalVisits() (int64, error) {
	var stat models.TrafficStat
	err := config.DB.Where("ip = ?", TotalPVKey).First(&stat).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, nil
		}
		return 0, err
	}
	return stat.Count, nil
}
