<template>
  <!-- 登录/初始化页 -->
  <div v-if="!isAuthed" class="auth-container">
    <div class="auth-box">
      <div class="auth-title">管理后台</div>
      <div class="auth-subtitle">API: {{ apiBaseUrl }}</div>

      <div v-if="mode === 'login'" class="space-y">
        <div class="input-group">
          <label class="label">用户名</label>
          <input v-model.trim="loginForm.username" class="input" placeholder="请输入用户名" @keyup.enter="handleLogin" />
        </div>
        <div class="input-group">
          <label class="label">密码</label>
          <input v-model="loginForm.password" type="password" class="input" placeholder="请输入密码" @keyup.enter="handleLogin" />
        </div>
        
        <button class="btn btn-primary" style="width: 100%" @click="handleLogin" :disabled="busy">登录</button>
        
        <div class="flex-between" style="margin-top: 16px;">
          <button class="btn btn-sm" @click="tryRestoreToken" :disabled="busy">刷新登录态</button>
          <button class="btn btn-sm" @click="mode = 'setup'">初始化账号</button>
        </div>
        
        <div v-if="error" class="text-danger text-sm" style="text-align: center; margin-top: 12px;">{{ error }}</div>
      </div>

      <div v-else class="space-y">
        <div class="input-group">
          <label class="label">用户名</label>
          <input v-model.trim="setupForm.username" class="input" />
        </div>
        <div class="input-group">
          <label class="label">密码</label>
          <input v-model="setupForm.password" type="password" class="input" />
        </div>
        <div class="input-group">
          <label class="label">SetupKey <span class="optional">(后端配置)</span></label>
          <input v-model.trim="setupForm.setupKey" class="input" />
        </div>
        
        <button class="btn btn-primary" style="width: 100%" @click="handleSetup" :disabled="busy">初始化并登录</button>
        <button class="btn" style="width: 100%; margin-top: 8px;" @click="mode = 'login'">返回登录</button>
        
        <div v-if="error" class="text-danger text-sm" style="text-align: center; margin-top: 12px;">{{ error }}</div>
      </div>
    </div>
  </div>

  <!-- 主界面 -->
  <div v-else class="app-layout">
    <!-- 侧边栏 -->
    <aside class="sidebar">
      <div class="sidebar-header">
        Bre News
      </div>
      <nav class="sidebar-nav">
        <a class="nav-item" :class="{ active: tab === 'users' }" @click="switchTab('users')">
          <span class="nav-icon">👤</span> 用户管理
        </a>
        <a class="nav-item" :class="{ active: tab === 'sites' }" @click="switchTab('sites')">
          <span class="nav-icon">🌐</span> 网站管理
        </a>
        <a class="nav-item" :class="{ active: tab === 'batches' }" @click="switchTab('batches')">
          <span class="nav-icon">📦</span> 批次管理
        </a>
        <a class="nav-item" :class="{ active: tab === 'news' }" @click="switchTab('news')">
          <span class="nav-icon">📰</span> 新闻管理
        </a>
        <a class="nav-item" :class="{ active: tab === 'analysis' }" @click="switchTab('analysis')">
          <span class="nav-icon">📈</span> 分析管理
        </a>
      </nav>
      <div class="sidebar-footer">
        <div class="text-xs text-muted" style="margin-bottom: 8px;">API: {{ apiBaseUrl }}</div>
        <button class="btn btn-sm btn-danger" style="width: 100%" @click="handleLogout" :disabled="busy">退出登录</button>
      </div>
    </aside>

    <!-- 内容区 -->
    <main class="main-content">
      <header class="top-bar">
        <div class="font-bold text-lg">
          {{ tabName }}
        </div>
        <div class="flex-center space-x">
          <span class="text-sm text-muted">管理员</span>
          <div class="badge badge-blue">Admin</div>
        </div>
      </header>

      <div class="page-content">
        <div v-if="error" class="card" style="border-left: 4px solid var(--danger-color);">
          <div class="text-danger">{{ error }}</div>
        </div>

        <!-- 用户管理 -->
        <div v-if="tab === 'users'" class="space-y">
          <div class="flex-between mb-4">
             <h2 class="text-lg font-bold">用户列表</h2>
             <div class="space-x">
               <button class="btn btn-primary" @click="openCreateUserModal" :disabled="busy">新增用户</button>
               <button class="btn" @click="loadUsers" :disabled="busy">刷新数据</button>
             </div>
          </div>

          <div class="card" style="padding: 0;">
            <div class="table-container">
              <table>
                <thead>
                  <tr>
                    <th style="width: 80px;">ID</th>
                    <th>用户名</th>
                    <th style="width: 220px;">创建时间</th>
                    <th style="width: 140px;">操作</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="u in users" :key="u.id">
                    <td>{{ u.id }}</td>
                    <td><span class="font-bold">{{ u.username }}</span></td>
                    <td class="text-muted text-sm">{{ formatTime(u.created_at) }}</td>
                    <td>
                      <div class="space-x">
                        <button class="btn btn-sm" @click="openSetPasswordModal(u)" :disabled="busy">修改密码</button>
                        <button class="btn btn-sm btn-danger" @click="handleDeleteUser(u.id)" :disabled="busy">删除</button>
                      </div>
                    </td>
                  </tr>
                  <tr v-if="users.length === 0">
                    <td colspan="4" class="text-muted" style="text-align: center; padding: 32px;">暂无数据</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>

        <!-- 网站管理 -->
        <div v-else-if="tab === 'sites'" class="space-y">
          <!-- 分类部分 -->
          <div class="card space-y">
            <div class="card-header">
              <div class="card-title">网站分类</div>
              <div class="space-x">
                <button class="btn btn-primary btn-sm" @click="openCreateCategoryModal" :disabled="busy">新增分类</button>
                <button class="btn btn-sm" @click="loadSiteCategories" :disabled="busy">刷新</button>
              </div>
            </div>

            <div class="table-container mt-4">
              <table>
                <thead>
                  <tr>
                    <th style="width: 80px;">ID</th>
                    <th>名称</th>
                    <th style="width: 80px;">排序</th>
                    <th style="width: 140px;">操作</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="c in categories" :key="c.id">
                    <td>{{ c.id }}</td>
                    <td class="font-bold">{{ c.name }}</td>
                    <td>{{ c.sort }}</td>
                    <td>
                      <div class="space-x">
                        <button class="btn btn-sm" @click="openEditCategoryModal(c)" :disabled="busy">编辑</button>
                        <button class="btn btn-sm btn-danger" @click="handleDeleteCategory(c.id)" :disabled="busy">删除</button>
                      </div>
                    </td>
                  </tr>
                  <tr v-if="categories.length === 0">
                    <td colspan="4" class="text-muted text-center">暂无数据</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>

          <!-- 网站列表部分 -->
          <div class="card space-y">
            <div class="card-header">
              <div class="card-title">网站列表</div>
              <div class="space-x flex-center">
                <button class="btn btn-primary btn-sm" @click="openCreateSiteModal" :disabled="busy">新增网站</button>
                <select v-model="siteFilterCategoryId" class="select" style="width: 150px;">
                  <option :value="''">全部分类</option>
                  <option v-for="c in categories" :key="c.id" :value="String(c.id)">{{ c.name }}</option>
                </select>
                <button class="btn btn-sm" @click="loadSites" :disabled="busy">刷新</button>
              </div>
            </div>

            <div class="table-container mt-4">
              <table>
                <thead>
                  <tr>
                    <th style="width: 60px;">ID</th>
                    <th style="width: 80px;">分类</th>
                    <th style="width: 150px;">名称</th>
                    <th>URL</th>
                    <th style="width: 60px;">排序</th>
                    <th style="width: 100px;">操作</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="s in sites" :key="s.id">
                    <td>{{ s.id }}</td>
                    <td><span class="badge badge-gray">{{ s.category_id }}</span></td>
                    <td class="font-bold">{{ s.name }}</td>
                    <td class="text-sm text-muted" style="word-break: break-all;">{{ s.url }}</td>
                    <td>{{ s.sort }}</td>
                    <td>
                      <div class="space-x">
                        <button class="btn btn-sm" @click="openEditSiteModal(s)" :disabled="busy">编辑</button>
                        <button class="btn btn-sm btn-danger" @click="handleDeleteSite(s.id)" :disabled="busy">删除</button>
                      </div>
                    </td>
                  </tr>
                  <tr v-if="sites.length === 0">
                    <td colspan="6" class="text-muted text-center">暂无数据</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>

        <!-- 批次管理 -->
        <div v-else-if="tab === 'batches'" class="space-y">
          <div class="card space-y">
            <div class="card-header">
              <div class="card-title">批次筛选</div>
              <button class="btn btn-sm" @click="loadBatches" :disabled="busy">刷新</button>
            </div>
            <div class="grid-4">
              <div class="input-group">
                <label class="label">类型</label>
                <select v-model="batchFilters.type" class="select">
                  <option value="">全部</option>
                  <option value="morning">Morning</option>
                  <option value="noon">Noon</option>
                  <option value="evening">Evening</option>
                </select>
              </div>
              <div class="input-group">
                <label class="label">开始时间</label>
                <input v-model.trim="batchFilters.createdAtStart" class="input" placeholder="2006-01-02" />
              </div>
              <div class="input-group">
                <label class="label">结束时间</label>
                <input v-model.trim="batchFilters.createdAtEnd" class="input" placeholder="2006-01-02" />
              </div>
              <div class="input-group" style="display: flex; align-items: flex-end;">
                <button class="btn btn-primary" style="width: 100%" @click="loadBatches" :disabled="busy">查询</button>
              </div>
            </div>
          </div>

          <div class="card" style="padding: 0;">
            <div class="table-container">
              <table>
                <thead>
                  <tr>
                    <th style="width: 80px;">ID</th>
                    <th style="width: 110px;">类型</th>
                    <th style="width: 140px;">日期</th>
                    <th style="width: 200px;">创建时间</th>
                    <th style="width: 220px;">操作</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="b in batches" :key="b.id">
                    <td>{{ b.id }}</td>
                    <td>
                      <span class="badge" :class="b.type === 'morning' ? 'badge-green' : 'badge-blue'">{{ b.type }}</span>
                    </td>
                    <td>{{ b.date }}</td>
                    <td class="text-sm text-muted">{{ formatTime(b.created_at) }}</td>
                    <td class="space-x">
                      <button class="btn btn-sm" @click="loadBatchNews(b.id)" :disabled="busy">查看新闻</button>
                      <button class="btn btn-sm btn-danger" @click="handleDeleteBatch(b.id)" :disabled="busy">删除</button>
                    </td>
                  </tr>
                  <tr v-if="batches.length === 0">
                    <td colspan="5" class="text-muted text-center">暂无数据</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>

          <div v-if="batchNews.length" class="card space-y">
            <div class="card-header">
              <div class="card-title">批次新闻 (ID: {{ activeBatchId }})</div>
              <button class="btn btn-sm" @click="batchNews = []">关闭</button>
            </div>
            <div class="table-container">
              <table>
                <thead>
                  <tr>
                    <th style="width: 80px;">ID</th>
                    <th>标题</th>
                    <th style="width: 240px;">URL</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="n in batchNews" :key="n.id">
                    <td>{{ n.id }}</td>
                    <td class="font-bold">{{ n.title }}</td>
                    <td class="text-sm text-muted" style="word-break: break-all;">
                      <a :href="n.url" target="_blank" style="color: var(--primary-color);">链接</a>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>

        <!-- 新闻管理 -->
        <div v-else-if="tab === 'news'" class="space-y">
          <div class="card space-y">
            <div class="card-header">
              <div class="card-title">新闻管理</div>
              <div class="space-x">
                <button class="btn btn-primary btn-sm" @click="openCreateNewsModal" :disabled="busy">新增新闻</button>
                <button class="btn btn-sm" @click="loadNews" :disabled="busy">刷新</button>
              </div>
            </div>
            <div class="grid-4">
              <div class="input-group">
                <label class="label">Batch ID</label>
                <input v-model.number="newsFilters.batchId" type="number" class="input" />
              </div>
              <div class="input-group">
                <label class="label">关键字</label>
                <input v-model.trim="newsFilters.keyword" class="input" placeholder="标题匹配" />
              </div>
              <div class="input-group">
                <label class="label">开始时间</label>
                <input v-model.trim="newsFilters.createdAtStart" class="input" />
              </div>
              <div class="input-group">
                <label class="label">结束时间</label>
                <input v-model.trim="newsFilters.createdAtEnd" class="input" />
              </div>
            </div>
            <button class="btn btn-primary" @click="loadNews" :disabled="busy">查询</button>
          </div>

          <div class="card" style="padding: 0;">
            <div class="table-container">
              <table>
                <thead>
                  <tr>
                    <th style="width: 80px;">ID</th>
                    <th style="width: 80px;">Batch</th>
                    <th>标题</th>
                    <th style="width: 150px;">Source</th>
                    <th style="width: 200px;">时间</th>
                    <th style="width: 100px;">操作</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="n in news" :key="n.id">
                    <td>{{ n.id }}</td>
                    <td>{{ n.batch_id }}</td>
                    <td class="font-bold">{{ n.title }}</td>
                    <td class="text-sm text-muted">{{ n.source }}</td>
                    <td class="text-sm text-muted">{{ formatTime(n.created_at) }}</td>
                    <td>
                      <div class="space-x">
                        <button class="btn btn-sm" @click="openEditNewsModal(n)" :disabled="busy">编辑</button>
                        <button class="btn btn-sm btn-danger" @click="handleDeleteNews(n.id)" :disabled="busy">删除</button>
                      </div>
                    </td>
                  </tr>
                  <tr v-if="news.length === 0">
                    <td colspan="6" class="text-muted text-center">暂无数据</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>

        <!-- 分析管理 -->
        <div v-else-if="tab === 'analysis'" class="space-y">
          <div class="card space-y">
            <div class="card-header">
              <div class="card-title">分析管理</div>
              <div class="space-x">
                <button class="btn btn-primary btn-sm" @click="openCreateAnalysisModal" :disabled="busy">新增分析</button>
                <button class="btn btn-sm" @click="loadAnalysis" :disabled="busy">刷新</button>
              </div>
            </div>
            <div class="grid-4">
              <div class="input-group">
                <label class="label">Batch ID</label>
                <input v-model.number="analysisFilters.batchId" type="number" class="input" />
              </div>
              <div class="input-group">
                <label class="label">类型</label>
                <select v-model="analysisFilters.type" class="select">
                  <option value="">全部</option>
                  <option value="3_day">3 Days</option>
                  <option value="7_day">7 Days</option>
                </select>
              </div>
              <div class="input-group">
                <label class="label">开始时间</label>
                <input v-model.trim="analysisFilters.createdAtStart" class="input" />
              </div>
              <div class="input-group">
                <label class="label">结束时间</label>
                <input v-model.trim="analysisFilters.createdAtEnd" class="input" />
              </div>
            </div>
            <button class="btn btn-primary" @click="loadAnalysis" :disabled="busy">查询</button>
          </div>

          <div class="card" style="padding: 0;">
            <div class="table-container">
              <table>
                <thead>
                  <tr>
                    <th style="width: 80px;">ID</th>
                    <th style="width: 80px;">Batch</th>
                    <th style="width: 100px;">类型</th>
                    <th>内容摘要</th>
                    <th style="width: 200px;">时间</th>
                    <th style="width: 100px;">操作</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="a in analysisList" :key="a.id">
                    <td>{{ a.id }}</td>
                    <td>{{ a.batch_id }}</td>
                    <td><span class="badge badge-blue">{{ a.type }}</span></td>
                    <td class="text-sm text-muted">{{ a.content ? a.content.slice(0, 50) + '...' : '' }}</td>
                    <td class="text-sm text-muted">{{ formatTime(a.created_at) }}</td>
                    <td>
                      <div class="space-x">
                        <button class="btn btn-sm" @click="openEditAnalysisModal(a)" :disabled="busy">编辑</button>
                        <button class="btn btn-sm btn-danger" @click="handleDeleteAnalysis(a.id)" :disabled="busy">删除</button>
                      </div>
                    </td>
                  </tr>
                  <tr v-if="analysisList.length === 0">
                    <td colspan="6" class="text-muted text-center">暂无数据</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>
      </div>
    </main>
  </div>

  <div v-if="modalOpen" class="modal-backdrop" @click.self="closeModal">
    <div class="modal">
      <div class="modal-header">
        <div class="modal-title">{{ modalTitle }}</div>
        <button class="btn btn-sm" @click="closeModal" :disabled="busy || modalBusy">关闭</button>
      </div>
      <div class="modal-body space-y">
        <div v-if="modalError" class="text-danger text-sm">{{ modalError }}</div>

        <template v-if="modalKind === 'createUser'">
          <div class="input-group">
            <label class="label">用户名</label>
            <input v-model.trim="modalForm.username" class="input" />
          </div>
          <div class="input-group">
            <label class="label">密码</label>
            <input v-model="modalForm.password" type="password" class="input" />
          </div>
        </template>

        <template v-else-if="modalKind === 'setUserPassword'">
          <div class="input-group">
            <label class="label">用户ID</label>
            <input v-model.number="modalForm.id" type="number" class="input" disabled />
          </div>
          <div class="input-group">
            <label class="label">新密码</label>
            <input v-model="modalForm.password" type="password" class="input" />
          </div>
        </template>

        <template v-else-if="modalKind === 'createCategory' || modalKind === 'editCategory'">
          <div class="grid-2">
            <div class="input-group">
              <label class="label">名称</label>
              <input v-model.trim="modalForm.name" class="input" />
            </div>
            <div class="input-group">
              <label class="label">排序</label>
              <input v-model.number="modalForm.sort" type="number" class="input" />
            </div>
          </div>
        </template>

        <template v-else-if="modalKind === 'createSite' || modalKind === 'editSite'">
          <div class="grid-2">
            <div class="input-group">
              <label class="label">分类</label>
              <select v-model.number="modalForm.category_id" class="select">
                <option :value="0">请选择</option>
                <option v-for="c in categories" :key="c.id" :value="c.id">{{ c.name }}</option>
              </select>
            </div>
            <div class="input-group">
              <label class="label">排序</label>
              <input v-model.number="modalForm.sort" type="number" class="input" />
            </div>
          </div>
          <div class="grid-2">
            <div class="input-group">
              <label class="label">名称</label>
              <input v-model.trim="modalForm.name" class="input" />
            </div>
            <div class="input-group">
              <label class="label">URL</label>
              <input v-model.trim="modalForm.url" class="input" />
            </div>
          </div>
          <div class="grid-2">
            <div class="input-group">
              <label class="label">图标</label>
              <input v-model.trim="modalForm.icon" class="input" />
            </div>
            <div class="input-group">
              <label class="label">描述</label>
              <input v-model.trim="modalForm.description" class="input" />
            </div>
          </div>
        </template>

        <template v-else-if="modalKind === 'createNews' || modalKind === 'editNews'">
          <div class="grid-2">
            <div class="input-group">
              <label class="label">Batch ID</label>
              <input v-model.number="modalForm.batch_id" type="number" class="input" />
            </div>
            <div class="input-group">
              <label class="label">Source</label>
              <input v-model.trim="modalForm.source" class="input" />
            </div>
          </div>
          <div class="input-group">
            <label class="label">Title</label>
            <input v-model.trim="modalForm.title" class="input" />
          </div>
          <div class="input-group">
            <label class="label">URL</label>
            <input v-model.trim="modalForm.url" class="input" />
          </div>
          <div class="input-group">
            <label class="label">Content</label>
            <textarea v-model="modalForm.content" class="textarea"></textarea>
          </div>
        </template>

        <template v-else-if="modalKind === 'createAnalysis' || modalKind === 'editAnalysis'">
          <div class="grid-2">
            <div class="input-group">
              <label class="label">Batch ID</label>
              <input v-model.number="modalForm.batch_id" type="number" class="input" />
            </div>
            <div class="input-group">
              <label class="label">Type</label>
              <select v-model="modalForm.type" class="select">
                <option value="3_day">3_day</option>
                <option value="7_day">7_day</option>
              </select>
            </div>
          </div>
          <div class="input-group">
            <label class="label">Content</label>
            <textarea v-model="modalForm.content" class="textarea"></textarea>
          </div>
        </template>
      </div>
      <div class="modal-footer space-x">
        <button class="btn" @click="closeModal" :disabled="busy || modalBusy">取消</button>
        <button class="btn btn-primary" @click="confirmModal" :disabled="busy || modalBusy">{{ modalConfirmText }}</button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed } from 'vue';
import * as api from './api';
import { format } from 'date-fns';

const mode = ref('login'); // login, setup
const tab = ref('users'); // users, sites, batches, news, analysis
const isAuthed = ref(false);
const busy = ref(false);
const error = ref('');
const apiBaseUrl = api.baseURL;

const tabName = computed(() => {
  const map = {
    users: '用户管理',
    sites: '网站管理',
    batches: '批次管理',
    news: '新闻管理',
    analysis: '分析管理'
  };
  return map[tab.value] || '管理后台';
});

const formatTime = (t) => {
  if (!t) return '-';
  try {
    return format(new Date(t), 'yyyy-MM-dd HH:mm:ss');
  } catch (e) {
    return t;
  }
};

// --- Auth Forms ---
const loginForm = reactive({ username: '', password: '' });
const setupForm = reactive({ username: '', password: '', setupKey: '' });

// --- Data ---
const users = ref([]);
const categories = ref([]);
const sites = ref([]);
const batches = ref([]);
const batchNews = ref([]);
const activeBatchId = ref(0);
const news = ref([]);
const analysisList = ref([]);

// --- User Forms ---
const siteFilterCategoryId = ref('');

// --- Batch Forms ---
const batchFilters = reactive({ type: '', createdAtStart: '', createdAtEnd: '' });

// --- News Forms ---
const newsFilters = reactive({ batchId: '', keyword: '', createdAtStart: '', createdAtEnd: '' });

// --- Analysis Forms ---
const analysisFilters = reactive({ batchId: '', type: '', createdAtStart: '', createdAtEnd: '' });

// --- Handlers ---

const handleError = (e) => {
  console.error(e);
  if (e.response && e.response.data && e.response.data.msg) {
    error.value = e.response.data.msg;
  } else {
    error.value = e.message || 'Unknown error';
  }
  setTimeout(() => { error.value = ''; }, 3000);
};

const tryRestoreToken = () => {
  const token = localStorage.getItem('admin_token');
  if (token) {
    api.setToken(token);
    isAuthed.value = true;
    loadData();
  }
};

const handleSetup = async () => {
  busy.value = true;
  error.value = '';
  try {
    const res = await api.adminSetup(setupForm);
    const token = res.data.token;
    localStorage.setItem('admin_token', token);
    api.setToken(token);
    isAuthed.value = true;
    mode.value = 'login';
    loadData();
  } catch (e) {
    handleError(e);
  } finally {
    busy.value = false;
  }
};

const handleLogin = async () => {
  busy.value = true;
  error.value = '';
  try {
    const res = await api.adminLogin(loginForm);
    const token = res.data.token;
    localStorage.setItem('admin_token', token);
    api.setToken(token);
    isAuthed.value = true;
    loadData();
  } catch (e) {
    handleError(e);
  } finally {
    busy.value = false;
  }
};

const handleLogout = async () => {
  busy.value = true;
  try {
    await api.adminLogout();
  } catch (e) {
    console.warn(e);
  }
  localStorage.removeItem('admin_token');
  isAuthed.value = false;
  busy.value = false;
};

const loadData = () => {
  if (tab.value === 'users') loadUsers();
  else if (tab.value === 'sites') { loadSiteCategories(); loadSites(); }
  else if (tab.value === 'batches') loadBatches();
  else if (tab.value === 'news') loadNews();
  else if (tab.value === 'analysis') loadAnalysis();
};

const switchTab = (t) => {
  tab.value = t;
  loadData();
};

const modalOpen = ref(false);
const modalKind = ref('');
const modalTitle = ref('');
const modalConfirmText = ref('确认');
const modalError = ref('');
const modalBusy = ref(false);
const modalForm = reactive({});

let modalConfirmHandler = null;

const openModal = ({ kind, title, confirmText, formInit, onConfirm }) => {
  modalKind.value = kind;
  modalTitle.value = title;
  modalConfirmText.value = confirmText || '确认';
  modalError.value = '';
  for (const k of Object.keys(modalForm)) delete modalForm[k];
  Object.assign(modalForm, formInit || {});
  modalConfirmHandler = onConfirm;
  modalOpen.value = true;
};

const closeModal = () => {
  if (modalBusy.value) return;
  modalOpen.value = false;
  modalError.value = '';
  modalConfirmHandler = null;
};

const confirmModal = async () => {
  if (!modalConfirmHandler) return;
  modalBusy.value = true;
  modalError.value = '';
  try {
    await modalConfirmHandler();
    closeModal();
  } catch (e) {
    if (e?.response?.data?.msg) modalError.value = e.response.data.msg;
    else modalError.value = e?.message || 'Unknown error';
  } finally {
    modalBusy.value = false;
  }
};

const openCreateUserModal = () => {
  openModal({
    kind: 'createUser',
    title: '新增用户',
    confirmText: '确认新增',
    formInit: { username: '', password: '' },
    onConfirm: async () => {
      await handleCreateUser({
        username: String(modalForm.username || '').trim(),
        password: String(modalForm.password || ''),
      });
      await loadUsers();
    },
  });
};

const openSetPasswordModal = (u) => {
  openModal({
    kind: 'setUserPassword',
    title: `修改密码 (ID: ${u.id})`,
    confirmText: '确认修改',
    formInit: { id: u.id, password: '' },
    onConfirm: async () => {
      await handleSetPassword({ id: Number(modalForm.id), password: String(modalForm.password || '') });
      alert('密码已修改');
    },
  });
};

const openCreateCategoryModal = () => {
  openModal({
    kind: 'createCategory',
    title: '新增分类',
    confirmText: '确认新增',
    formInit: { name: '', sort: 0 },
    onConfirm: async () => {
      await handleCreateCategory({ name: String(modalForm.name || '').trim(), sort: Number(modalForm.sort || 0) });
      await loadSiteCategories();
    },
  });
};

const openEditCategoryModal = (c) => {
  openModal({
    kind: 'editCategory',
    title: `编辑分类 (ID: ${c.id})`,
    confirmText: '保存修改',
    formInit: { id: c.id, name: c.name || '', sort: Number(c.sort || 0) },
    onConfirm: async () => {
      await handleUpdateCategory({
        id: Number(modalForm.id),
        name: String(modalForm.name || '').trim(),
        sort: Number(modalForm.sort || 0),
      });
      await loadSiteCategories();
      await loadSites();
    },
  });
};

const openCreateSiteModal = () => {
  const defaultCategoryId = categories.value?.[0]?.id ? Number(categories.value[0].id) : 0;
  openModal({
    kind: 'createSite',
    title: '新增网站',
    confirmText: '确认新增',
    formInit: { category_id: defaultCategoryId, name: '', url: '', icon: '', description: '', sort: 0 },
    onConfirm: async () => {
      await handleCreateSite({
        category_id: Number(modalForm.category_id || 0),
        name: String(modalForm.name || '').trim(),
        url: String(modalForm.url || '').trim(),
        icon: String(modalForm.icon || '').trim(),
        description: String(modalForm.description || '').trim(),
        sort: Number(modalForm.sort || 0),
      });
      await loadSites();
    },
  });
};

const openEditSiteModal = (s) => {
  openModal({
    kind: 'editSite',
    title: `编辑网站 (ID: ${s.id})`,
    confirmText: '保存修改',
    formInit: {
      id: s.id,
      category_id: Number(s.category_id || 0),
      name: s.name || '',
      url: s.url || '',
      icon: s.icon || '',
      description: s.description || '',
      sort: Number(s.sort || 0),
    },
    onConfirm: async () => {
      await handleUpdateSite({
        id: Number(modalForm.id),
        category_id: Number(modalForm.category_id || 0),
        name: String(modalForm.name || '').trim(),
        url: String(modalForm.url || '').trim(),
        icon: String(modalForm.icon || '').trim(),
        description: String(modalForm.description || '').trim(),
        sort: Number(modalForm.sort || 0),
      });
      await loadSites();
    },
  });
};

const openCreateNewsModal = () => {
  const batchId = newsFilters.batchId ? Number(newsFilters.batchId) : 0;
  openModal({
    kind: 'createNews',
    title: '新增新闻',
    confirmText: '确认新增',
    formInit: { batch_id: batchId, source: '', title: '', url: '', content: '' },
    onConfirm: async () => {
      await handleCreateNews({
        batch_id: Number(modalForm.batch_id || 0),
        source: String(modalForm.source || '').trim(),
        title: String(modalForm.title || '').trim(),
        url: String(modalForm.url || '').trim(),
        content: String(modalForm.content || ''),
      });
      await loadNews();
    },
  });
};

const openEditNewsModal = (n) => {
  openModal({
    kind: 'editNews',
    title: `编辑新闻 (ID: ${n.id})`,
    confirmText: '保存修改',
    formInit: {
      id: n.id,
      batch_id: Number(n.batch_id || 0),
      source: n.source || '',
      title: n.title || '',
      url: n.url || '',
      content: n.content || '',
    },
    onConfirm: async () => {
      await handleUpdateNews({
        id: Number(modalForm.id),
        batch_id: Number(modalForm.batch_id || 0),
        source: String(modalForm.source || '').trim(),
        title: String(modalForm.title || '').trim(),
        url: String(modalForm.url || '').trim(),
        content: String(modalForm.content || ''),
      });
      await loadNews();
    },
  });
};

const openCreateAnalysisModal = () => {
  const batchId = analysisFilters.batchId ? Number(analysisFilters.batchId) : 0;
  openModal({
    kind: 'createAnalysis',
    title: '新增分析',
    confirmText: '确认新增',
    formInit: { batch_id: batchId, type: '3_day', content: '' },
    onConfirm: async () => {
      await handleCreateAnalysis({
        batch_id: Number(modalForm.batch_id || 0),
        type: String(modalForm.type || '3_day'),
        content: String(modalForm.content || ''),
      });
      await loadAnalysis();
    },
  });
};

const openEditAnalysisModal = (a) => {
  openModal({
    kind: 'editAnalysis',
    title: `编辑分析 (ID: ${a.id})`,
    confirmText: '保存修改',
    formInit: {
      id: a.id,
      batch_id: Number(a.batch_id || 0),
      type: a.type || '3_day',
      content: a.content || '',
    },
    onConfirm: async () => {
      await handleUpdateAnalysis({
        id: Number(modalForm.id),
        batch_id: Number(modalForm.batch_id || 0),
        type: String(modalForm.type || ''),
        content: String(modalForm.content || ''),
      });
      await loadAnalysis();
    },
  });
};

// Users
const loadUsers = async () => {
  busy.value = true;
  try {
    const res = await api.getUsers();
    users.value = res.rows || [];
  } catch (e) { handleError(e); } finally { busy.value = false; }
};
const handleCreateUser = async ({ username, password }) => {
  try {
    await api.createUser({ username, password });
  } catch (e) { handleError(e); }
};
const handleSetPassword = async ({ id, password }) => {
  try {
    await api.updateUserPassword({ id, password });
  } catch (e) { handleError(e); }
};
const handleDeleteUser = async (id) => {
  if (!confirm('确定删除?')) return;
  busy.value = true;
  try {
    await api.deleteUser(id);
    loadUsers();
  } catch (e) { handleError(e); } finally { busy.value = false; }
};

// Sites
const loadSiteCategories = async () => {
  busy.value = true;
  try {
    const res = await api.getSiteCategories();
    categories.value = res.rows || [];
  } catch (e) { handleError(e); } finally { busy.value = false; }
};
const handleCreateCategory = async ({ name, sort }) => {
  try {
    await api.createSiteCategory({ name, sort });
  } catch (e) { handleError(e); }
};
const handleUpdateCategory = async ({ id, name, sort }) => {
  try {
    await api.updateSiteCategory({ id, name, sort });
  } catch (e) { handleError(e); }
};
const handleDeleteCategory = async (id) => {
  if (!confirm('确定删除? 会同时删除该分类下所有网站!')) return;
  busy.value = true;
  try {
    await api.deleteSiteCategory(id);
    loadSiteCategories(); loadSites();
  } catch (e) { handleError(e); } finally { busy.value = false; }
};

const loadSites = async () => {
  busy.value = true;
  try {
    const res = await api.getSites({ categoryId: siteFilterCategoryId.value });
    sites.value = res.rows || [];
  } catch (e) { handleError(e); } finally { busy.value = false; }
};
const handleCreateSite = async (payload) => {
  try {
    await api.createSite(payload);
  } catch (e) { handleError(e); }
};
const handleUpdateSite = async ({ id, ...payload }) => {
  try {
    const updates = {};
    for (const [k, v] of Object.entries(payload)) {
      if (v === undefined || v === null) continue;
      if (typeof v === 'string' && v.trim() === '') continue;
      updates[k] = v;
    }
    await api.updateSite({ id, ...updates });
  } catch (e) { handleError(e); }
};
const handleDeleteSite = async (id) => {
  if (!confirm('确定删除?')) return;
  busy.value = true;
  try {
    await api.deleteSite(id);
    loadSites();
  } catch (e) { handleError(e); } finally { busy.value = false; }
};

// Batches
const loadBatches = async () => {
  busy.value = true;
  try {
    const res = await api.getBatches(batchFilters);
    batches.value = res.rows || [];
  } catch (e) { handleError(e); } finally { busy.value = false; }
};
const loadBatchNews = async (id) => {
  busy.value = true;
  activeBatchId.value = id;
  try {
    const res = await api.getBatchNews(id);
    batchNews.value = res.rows || [];
  } catch (e) { handleError(e); } finally { busy.value = false; }
};
const handleDeleteBatch = async (id) => {
  if (!confirm('确定删除? 会删除该批次下所有新闻和分析!')) return;
  busy.value = true;
  try {
    await api.deleteBatch(id);
    loadBatches();
    if (activeBatchId.value === id) {
      batchNews.value = [];
      activeBatchId.value = 0;
    }
  } catch (e) { handleError(e); } finally { busy.value = false; }
};

// News
const loadNews = async () => {
  busy.value = true;
  try {
    const res = await api.getNews(newsFilters);
    news.value = res.rows || [];
  } catch (e) { handleError(e); } finally { busy.value = false; }
};
const handleCreateNews = async (payload) => {
  try {
    await api.createNews(payload);
  } catch (e) { handleError(e); }
};
const handleUpdateNews = async ({ id, ...payload }) => {
  try {
    const updates = {};
    for (const [k, v] of Object.entries(payload)) {
      if (v === undefined || v === null) continue;
      if (typeof v === 'string' && v.trim() === '') continue;
      updates[k] = v;
    }
    await api.updateNews({ id, ...updates });
  } catch (e) { handleError(e); }
};
const handleDeleteNews = async (id) => {
  if (!confirm('确定删除?')) return;
  busy.value = true;
  try {
    await api.deleteNews(id);
    loadNews();
  } catch (e) { handleError(e); } finally { busy.value = false; }
};

// Analysis
const loadAnalysis = async () => {
  busy.value = true;
  try {
    const res = await api.getAnalysis(analysisFilters);
    analysisList.value = res.rows || [];
  } catch (e) { handleError(e); } finally { busy.value = false; }
};
const handleCreateAnalysis = async (payload) => {
  try {
    await api.createAnalysis(payload);
  } catch (e) { handleError(e); }
};
const handleUpdateAnalysis = async ({ id, ...payload }) => {
  try {
    const updates = {};
    for (const [k, v] of Object.entries(payload)) {
      if (v === undefined || v === null) continue;
      if (typeof v === 'string' && v.trim() === '') continue;
      updates[k] = v;
    }
    await api.updateAnalysis({ id, ...updates });
  } catch (e) { handleError(e); }
};
const handleDeleteAnalysis = async (id) => {
  if (!confirm('确定删除?')) return;
  busy.value = true;
  try {
    await api.deleteAnalysis(id);
    loadAnalysis();
  } catch (e) { handleError(e); } finally { busy.value = false; }
};

onMounted(() => {
  tryRestoreToken();
});
</script>
