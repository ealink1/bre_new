package services

import (
	"bre_new_backend/config"
	"bre_new_backend/models"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

func RunUpdateTask() {
	RunUpdateTaskWithDeps(config.DB, time.Now, GetDailyNews, AnalyzeNews, true)
}

type NowFunc func() time.Time
type GetDailyNewsFunc func() ([]NewsData, error)
type AnalyzeNewsFunc func(newsContent string, days int) (string, error)

func RunUpdateTaskWithDeps(db *gorm.DB, nowFn NowFunc, getDailyNews GetDailyNewsFunc, analyzeNews AnalyzeNewsFunc, runAnalysis bool) {
	fmt.Println("开始执行定时更新任务...")

	// 1. 确定批次类型 (早/中/晚)
	if db == nil {
		fmt.Println("DB 未初始化")
		return
	}
	if nowFn == nil {
		nowFn = time.Now
	}
	if getDailyNews == nil {
		getDailyNews = GetDailyNews
	}
	if analyzeNews == nil {
		analyzeNews = AnalyzeNews
	}

	now := nowFn()
	hour := now.Hour()
	var batchType models.BatchType
	if hour < 10 {
		batchType = models.BatchMorning // 早报
	} else if hour < 16 {
		batchType = models.BatchNoon // 午报
	} else {
		batchType = models.BatchEvening // 晚报
	}

	// 3. 获取新闻数据
	fmt.Println("正在从 AI 获取今日热点新闻...")
	newsItems, err := getDailyNews()
	if err != nil {
		fmt.Printf("获取新闻失败: %v\n", err)
		return
	}

	// 2. 创建批次记录
	batch := models.BatchLog{
		Type: batchType,
		Date: now.Format("2006-01-02"),
	}
	if err := db.Create(&batch).Error; err != nil {
		fmt.Printf("创建批次记录失败: %v\n", err)
		return
	}
	fmt.Printf("已创建批次 %d (%s)\n", batch.ID, batchType)

	// 保存新闻条目
	for _, item := range newsItems {
		news := models.NewsItem{
			BatchID: batch.ID,
			Title:   item.Title,
			Content: item.Title,
			Url:     item.URL,
			Source:  "AI Summary",
		}
		if err := db.Create(&news).Error; err != nil {
			fmt.Printf("保存新闻失败: %v\n", err)
		} else {
			fmt.Printf("已保存新闻: %s, URL: %s\n", news.Title, news.Url)
		}
	}

	fmt.Println("新闻条目已保存")

	// 4. 执行 3 天财经分析
	if runAnalysis {
		analyzeAndSaveWithDeps(db, analyzeNews, 3, batch.ID, now)
	}

	// 5. 执行 7 天财经分析
	if runAnalysis {
		analyzeAndSaveWithDeps(db, analyzeNews, 7, batch.ID, now)
	}

	fmt.Println("更新任务完成")
}

func analyzeAndSave(days int, batchID uint) {
	analyzeAndSaveWithDeps(config.DB, AnalyzeNews, days, batchID, time.Now())
}

func analyzeAndSaveWithDeps(db *gorm.DB, analyzeNews AnalyzeNewsFunc, days int, batchID uint, now time.Time) {
	fmt.Printf("开始 %d 天财经分析...\n", days)
	// 获取过去 N 天的新闻
	cutoff := now.AddDate(0, 0, -days)
	var recentNews []models.NewsItem
	db.Where("created_at >= ?", cutoff).Find(&recentNews)

	if len(recentNews) == 0 {
		fmt.Println("未找到分析所需的新闻数据")
		return
	}

	var sb strings.Builder
	for _, n := range recentNews {
		sb.WriteString(fmt.Sprintf("- %s\n", n.Content))
	}

	// 调用 AI 进行分析
	analysisContent, err := analyzeNews(sb.String(), days)
	if err != nil {
		fmt.Printf("分析失败: %v\n", err)
		return
	}

	analysisType := models.Analysis3Day
	if days == 7 {
		analysisType = models.Analysis7Day
	}

	// 保存分析结果
	analysis := models.Analysis{
		BatchID: batchID,
		Type:    analysisType,
		Content: analysisContent,
	}
	db.Create(&analysis)
	fmt.Printf("已保存 %d 天分析结果\n", days)
}
