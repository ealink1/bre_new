<script setup>
import { ref, onMounted, onUnmounted } from 'vue'

const prices = ref([])
const loading = ref(true)
const error = ref(null)
let timer = null

const loadScript = (src) =>
  new Promise((resolve, reject) => {
    const script = document.createElement('script')
    script.async = true
    script.src = src

    script.onload = () => {
      script.remove()
      resolve()
    }

    script.onerror = () => {
      script.remove()
      reject(new Error(`Failed to load script: ${src}`))
    }

    document.head.appendChild(script)
  })

const withCacheBust = (url) => {
  const separator = url.includes('?') ? '&' : '?'
  return `${url}${separator}_t=${Date.now()}`
}

const parseGoldData = () => {
  const result = []
  
  const extractVar = (name) => {
    const value = window[name]
    return typeof value === 'string' ? value.split(',') : null
  }

  const autd = extractVar('hq_str_gds_AUTD')
  if (autd) {
    const current = parseFloat(autd[0])
    const prevClose = parseFloat(autd[7])
    const change = current - prevClose
    const changePercent = (change / prevClose) * 100
    
    result.push({
      id: 'shanghai',
      name: '国内黄金价格',
      market: '中国上海黄金交易所 人民币/克',
      price: current.toFixed(2),
      change: change.toFixed(2),
      changePercent: changePercent.toFixed(2),
      high: autd[4],
      low: autd[5],
      open: autd[8],
      prevClose: autd[7],
      time: autd[12] + ' ' + autd[6]
    })
  }

  const ny = extractVar('hq_str_hf_GC')
  if (ny) {
    const current = parseFloat(ny[0])
    const prevClose = parseFloat(ny[7])
    const change = current - prevClose
    const changePercent = (change / prevClose) * 100

    result.push({
      id: 'ny',
      name: '纽约期货国际金价',
      market: '美国纽约商品交易所 美元/盎司',
      price: current.toFixed(3),
      change: change.toFixed(2),
      changePercent: changePercent.toFixed(2),
      high: ny[4],
      low: ny[5],
      open: ny[8],
      prevClose: ny[7],
      time: ny[12] + ' ' + ny[6]
    })
  }

  const ldn = extractVar('hq_str_hf_XAU')
  if (ldn) {
    const current = parseFloat(ldn[0])
    const prevClose = parseFloat(ldn[7])
    const change = current - prevClose
    const changePercent = (change / prevClose) * 100

    result.push({
      id: 'london',
      name: '伦敦现货黄金价格',
      market: '英国伦敦黄金交易市场 美元/盎司',
      price: current.toFixed(2),
      change: change.toFixed(2),
      changePercent: changePercent.toFixed(2),
      high: ldn[4],
      low: ldn[5],
      open: ldn[8],
      prevClose: ldn[7],
      time: ldn[12] + ' ' + ldn[6]
    })
  }

  return result
}

const fetchGoldPrices = async () => {
  try {
    error.value = null
    await loadScript(withCacheBust('https://www.huilvbiao.com/api/gold_indexApi'))
    prices.value = parseGoldData()
    loading.value = false
  } catch (err) {
    console.error('Failed to fetch gold prices:', err)
    error.value = '获取金价数据失败'
    loading.value = false
  }
}

onMounted(() => {
  fetchGoldPrices()
  // Update every 10 seconds as per the website description
  timer = setInterval(fetchGoldPrices, 10000)
})

onUnmounted(() => {
  if (timer) clearInterval(timer)
})
</script>

<template>
  <div class="gold-price-container">
    <h2 class="section-title">今日金价 (实时更新)</h2>
    
    <div v-if="loading && prices.length === 0" class="loading">加载中...</div>
    <div v-else-if="error" class="error">{{ error }}</div>
    
    <div v-else class="price-cards">
      <div v-for="item in prices" :key="item.id" class="price-card">
        <div class="card-header">
          <h3 class="card-title">{{ item.name }}</h3>
          <span class="market-info">{{ item.market }}</span>
        </div>
        
        <div class="main-price" :class="item.change >= 0 ? 'up' : 'down'">
          <span class="current-price">{{ item.price }}</span>
          <div class="changes">
            <span class="change-amount">{{ item.change > 0 ? '+' : '' }}{{ item.change }}</span>
            <span class="change-percent">{{ item.changePercent > 0 ? '+' : '' }}{{ item.changePercent }}%</span>
          </div>
        </div>
        
        <div class="details-grid">
          <div class="detail-item">
            <span class="label">最高价:</span>
            <span class="value">{{ item.high }}</span>
          </div>
          <div class="detail-item">
            <span class="label">最低价:</span>
            <span class="value">{{ item.low }}</span>
          </div>
          <div class="detail-item">
            <span class="label">开盘价:</span>
            <span class="value">{{ item.open }}</span>
          </div>
          <div class="detail-item">
            <span class="label">昨结算:</span>
            <span class="value">{{ item.prevClose }}</span>
          </div>
        </div>
        
        <div class="update-time">更新时间: {{ item.time }}</div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.gold-price-container {
  background: #fff;
  border-radius: 8px;
  padding: 0;
}

.section-title {
  font-size: 1.5rem;
  color: #2c3e50;
  margin-bottom: 20px;
  text-align: center;
  font-weight: bold;
}

.price-cards {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.price-card {
  border: 1px solid #eee;
  border-radius: 8px;
  padding: 20px;
  background-color: #fcfcfc;
  transition: box-shadow 0.3s;
}

.price-card:hover {
  box-shadow: 0 4px 12px rgba(0,0,0,0.05);
}

.card-header {
  margin-bottom: 15px;
  border-bottom: 1px solid #eee;
  padding-bottom: 10px;
}

.card-title {
  margin: 0 0 5px 0;
  font-size: 1.2rem;
  color: #333;
}

.market-info {
  font-size: 0.85rem;
  color: #999;
}

.main-price {
  display: flex;
  align-items: baseline;
  margin-bottom: 20px;
}

.current-price {
  font-size: 2.5rem;
  font-weight: bold;
  margin-right: 15px;
  font-family: 'Helvetica Neue', Helvetica, Arial, sans-serif;
}

.changes {
  display: flex;
  gap: 10px;
  font-size: 1.1rem;
  font-weight: bold;
}

.up .current-price, .up .changes {
  color: #e74c3c;
}

.down .current-price, .down .changes {
  color: #27ae60;
}

.details-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 10px;
  margin-bottom: 15px;
  font-size: 0.9rem;
}

.detail-item {
  display: flex;
  justify-content: space-between;
}

.label {
  color: #7f8c8d;
}

.value {
  font-weight: 500;
  color: #2c3e50;
}

.update-time {
  font-size: 0.8rem;
  color: #bdc3c7;
  text-align: right;
  border-top: 1px dashed #eee;
  padding-top: 10px;
}

.loading, .error {
  text-align: center;
  padding: 40px;
  color: #999;
}

.error {
  color: #e74c3c;
}
</style>
