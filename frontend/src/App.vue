<script setup>
import { ref, onMounted } from 'vue'
import { fetchLatestNews, fetchAnalysis } from './api'
import MarkdownIt from 'markdown-it'
import GoldPrice from './components/GoldPrice.vue'
import SilverPrice from './components/SilverPrice.vue'

const md = new MarkdownIt()
const newsList = ref([])
const batchInfo = ref(null)
const analysis3Day = ref(null)
const analysis7Day = ref(null)
const activeTab = ref('home') // 'home', '3day', '7day'
const loading = ref(true)

const loadData = async () => {
  loading.value = true
  try {
    // Load News
    const newsRes = await fetchLatestNews()
    if (newsRes.data.code === 200) {
      newsList.value = newsRes.data.rows
      batchInfo.value = newsRes.data.batch
    }

    // Load Analysis
    const analysis3Res = await fetchAnalysis(3)
    if (analysis3Res.data.code === 200) {
      analysis3Day.value = analysis3Res.data.data
    }

    const analysis7Res = await fetchAnalysis(7)
    if (analysis7Res.data.code === 200) {
      analysis7Day.value = analysis7Res.data.data
    }

  } catch (error) {
    console.error("Failed to load data", error)
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadData()
})
</script>

<template>
  <div class="container">
    <header class="header">
      <div class="header-left">
        <h1>每日资讯</h1>
      </div>
      <nav class="header-nav">
        <button :class="{ active: activeTab === 'home' }" @click="activeTab = 'home'">首页</button>
        <button :class="{ active: activeTab === '3day' }" @click="activeTab = '3day'">3日财经分析</button>
        <button :class="{ active: activeTab === '7day' }" @click="activeTab = '7day'">7日财经分析</button>
      </nav>
      <div v-if="batchInfo" class="header-right">
        <span class="date">{{ batchInfo.date }}</span>
        <span class="type">{{ batchInfo.type === 'morning' ? '早报' : batchInfo.type === 'noon' ? '午报' : '晚报' }}</span>
      </div>
    </header>

    <main class="main-content" :class="{ home: activeTab === 'home', analysis: activeTab !== 'home' }">
      <section v-show="activeTab === 'home'" class="news-section">
        <div class="section-header">
          <span class="blue-bar"></span>
          <h2>最新热点简送</h2>
        </div>
        <div v-if="loading" class="loading">加载中...</div>
        <ul v-else class="news-list">
          <li v-for="(item, index) in newsList" :key="item.id" class="news-item">
            <div class="news-row">
              <span class="news-index">{{ index + 1 }}.</span>
              <p class="news-text">
                <span class="news-time">[{{ batchInfo?.date }} {{ batchInfo?.type === 'morning' ? '08:00' : '12:00' }}]</span>
                <a v-if="item.url" :href="item.url" target="_blank" class="news-link">{{ item.content }}</a>
                <span v-else>{{ item.content }}</span>
              </p>
            </div>
          </li>
          <li v-if="newsList.length === 0" class="no-data">暂无数据</li>
        </ul>
      </section>

      <section v-show="activeTab === 'home'" class="price-section">
        <GoldPrice />
      </section>

      <section v-show="activeTab === 'home'" class="price-section">
        <SilverPrice />
      </section>

      <section v-show="activeTab !== 'home'" class="analysis-section full-width">
        <div class="analysis-content">
          <div v-show="activeTab === '3day'">
            <div v-if="analysis3Day" class="content-box">
              <div v-html="md.render(analysis3Day.content)" class="markdown-body"></div>
            </div>
            <div v-else class="no-data">暂无3日分析数据</div>
          </div>

          <div v-show="activeTab === '7day'">
            <div v-if="analysis7Day" class="content-box">
              <div v-html="md.render(analysis7Day.content)" class="markdown-body"></div>
            </div>
            <div v-else class="no-data">暂无7日分析数据</div>
          </div>
        </div>
      </section>
    </main>
  </div>
</template>

<style scoped>
.container {
  padding: 20px;
  font-family: 'Helvetica Neue', Helvetica, Arial, sans-serif;
  background-color: #fff;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  border-bottom: 2px solid #ddd;
  padding: 15px 0;
  margin-bottom: 30px;
  position: sticky;
  top: 0;
  background-color: #fff;
  z-index: 1000;
}

.header-left h1 {
  margin: 0;
  font-size: 1.8rem;
  color: #333;
  font-weight: bold;
}

.header-nav {
  display: flex;
  justify-content: center;
  gap: 20px;
}

.header-nav button {
  box-sizing: border-box;
  flex-shrink: 0;
  background: none;
  border: 2px solid transparent;
  padding: 5px 0;
  width: 120px;
  text-align: center;
  font-size: 1rem;
  font-weight: bold;
  color: #333;
  cursor: pointer;
  border-radius: 4px;
}

.header-nav button.active {
  border: 2px solid #e74c3c;
  color: #e74c3c;
}

.header-left {
  flex: 1;
}

.header-right {
  flex: 1;
  display: flex;
  justify-content: flex-end;
  align-items: center;
  gap: 10px;
}

.date {
  color: #666;
  font-size: 0.9rem;
}

.type {
  background-color: #e74c3c;
  color: white;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 0.8rem;
}

/* Responsive Layout */
.main-content {
  display: flex;
  flex-direction: column;
  gap: 30px;
}

@media (min-width: 768px) {
  .main-content.home {
    display: grid;
    grid-template-columns: 2fr 1fr 1fr;
    align-items: start;
    gap: 30px;
  }

  .main-content.analysis {
    padding-left: 10%;
    padding-right: 10%;
  }
}

.section-header {
  display: flex;
  align-items: center;
  margin-bottom: 20px;
}

.blue-bar {
  display: inline-block;
  width: 4px;
  height: 24px;
  background-color: #3498db;
  margin-right: 10px;
}

.news-section h2 {
  margin: 0;
  font-size: 1.2rem;
  color: #333;
}

.news-list {
  padding: 0;
  margin: 0;
  list-style: none;
}

.news-item {
  padding: 15px 0;
  border-bottom: 1px solid #f0f0f0;
}

.news-item:last-child {
  border-bottom: none;
}

.news-row {
  display: flex;
  gap: 10px;
  align-items: flex-start;
  font-size: 0.9rem;
}

.news-index {
  color: #999;
  font-weight: bold;
  min-width: 20px;
}

.news-text {
  margin: 0;
  line-height: 1.5;
  color: #333;
}

.news-time {
  color: #999;
  margin-right: 5px;
  font-size: 0.85rem;
}

.news-link {
  color: #333;
  text-decoration: none;
}

.news-link:hover {
  color: #3498db;
}

/* Analysis Tabs Removed */

.content-box {
  background-color: #fff;
  padding: 20px;
  border-radius: 8px;
  line-height: 1.8;
  color: #333;
  /* box-shadow: 0 2px 8px rgba(0,0,0,0.05); */ /* Optional shadow */
  overflow-x: auto;
}

/* Markdown Styles */
:deep(.markdown-body) {
  text-align: left;
}

:deep(.markdown-body h1),
:deep(.markdown-body h2),
:deep(.markdown-body h3) {
  margin-top: 1.5em;
  margin-bottom: 0.5em;
  color: #2c3e50;
  border-bottom: 1px solid #eee;
  padding-bottom: 5px;
}

:deep(.markdown-body h1) {
  font-size: 1.8rem;
}

:deep(.markdown-body h2) {
  font-size: 1.4rem;
}

:deep(.markdown-body h3) {
  font-size: 1.2rem;
}

:deep(.markdown-body h1:first-child),
:deep(.markdown-body h2:first-child),
:deep(.markdown-body h3:first-child) {
  margin-top: 0;
}

:deep(.markdown-body p) {
  margin-bottom: 1em;
}

:deep(.markdown-body ul),
:deep(.markdown-body ol) {
  padding-left: 20px;
  margin-bottom: 1em;
}

:deep(.markdown-body li) {
  margin-bottom: 0.5em;
}

:deep(.markdown-body code) {
  background-color: #eee;
  padding: 2px 4px;
  border-radius: 3px;
  font-family: monospace;
}

:deep(.markdown-body table) {
  width: 100%;
  border-collapse: collapse;
  margin: 1em 0;
}

:deep(.markdown-body th),
:deep(.markdown-body td) {
  border: 1px solid #eee;
  padding: 10px 12px;
  vertical-align: top;
}

:deep(.markdown-body th) {
  background-color: #f8fafc;
  font-weight: 700;
}

:deep(.markdown-body pre) {
  background-color: #2c3e50;
  color: #fff;
  padding: 15px;
  border-radius: 5px;
  overflow-x: auto;
  margin-bottom: 1em;
}

:deep(.markdown-body pre code) {
  background-color: transparent;
  color: inherit;
  padding: 0;
}

.no-data, .loading {
  color: #999;
  text-align: center;
  padding: 20px;
}
</style>
