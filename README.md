# bre_new

## 系统功能介绍

本项目是一个集成了金融新闻自动采集、AI 分析与实时贵金属行情展示的全栈系统。

### 后端功能 (Go + Gin + Gorm)
- **定时任务调度**：
  - 每日分批次（早/中/晚）自动执行新闻采集任务。
  - 集成 AI 服务（GLM-4）自动生成每日新闻摘要。
  - 自动执行金融市场趋势分析（支持 3 日与 7 日周期）。
- **数据持久化**：
  - 使用 SQLite (`news.db`) 存储新闻条目、批次记录及分析结果。
- **API 服务**：
  - `/news/latest`: 获取最新批次的新闻列表。
  - `/analysis/latest`: 获取最新的金融分析报告（支持 `days=3` 或 `days=7` 参数）。

### 前端功能 (Vue 3 + Vite)
- **实时行情展示**：
  - **黄金价格**：实时展示上海黄金交易所（AUTD）及国际现货黄金价格。
  - **白银价格**：实时展示上海白银交易所（AGTD）及国际现货白银价格。
  - **技术实现**：采用浏览器端 Script Loading (JSONP 模式) 直接对接新浪财经接口，有效规避跨域 (CORS) 问题，无需后端代理。
- **新闻与分析展示**：
  - 展示后端生成的每日财经新闻与 AI 分析报告。

## 目录结构

- backend：Go 后端服务
- frontend：Vue3 + Vite 前端

## 后端部署（backend）

脚本：[build_push.sh](file:///Users/bre/workspace/self/bre_new/backend/build_push.sh)

- 默认输出二进制名：`bre_new`
- 默认部署目录：`/111workspace/news`
- 默认重启命令：`sudo systemctl restart bre_new.service`

常用命令：

```bash
cd backend

# 构建
./build_push.sh build

# 构建并部署（上传二进制 + 重启服务）
./build_push.sh deploy

# 默认：构建并部署
./build_push.sh
```

可用环境变量（示例）：

```bash
DEPLOY_HOST=114.132.245.76 DEPLOY_USER=root DEPLOY_PORT=22 DEPLOY_PATH=/111workspace/news ./build_push.sh
```

## 前端打包与部署（frontend）

### 仅打包

脚本：[package.sh](file:///Users/bre/workspace/self/bre_new/frontend/package.sh)

```bash
cd frontend
./package.sh
```

产物会输出为 `frontend/release/frontend-dist-YYYYmmddHHMMSS.tar.gz`，内容包含 `dist/`。

### 打包并上传覆盖解压

脚本：[package_push.sh](file:///Users/bre/workspace/self/bre_new/frontend/package_push.sh)

```bash
cd frontend

# 默认：打包 -> 上传 -> 覆盖解压到远端 dist/
./package_push.sh
```

默认参数：

- 服务器：`root@114.132.245.76`（可用 `DEPLOY_HOST`/`DEPLOY_USER`/`DEPLOY_PORT` 覆盖）
- 远端目录：`/111workspace/news/frontend`
- 上传位置：`${DEPLOY_PATH}/release/`
- 覆盖解压：删除 `${DEPLOY_PATH}/dist` 后再解压生成新的 `dist/`

常用覆盖配置示例：

```bash
# 指定远端目录（比如 Nginx 静态目录的父目录）
DEPLOY_PATH=/var/www/bre_new_frontend ./package_push.sh

# 指定 SSH 参数（例如禁用本机代理）
DEPLOY_SSH_ARGS='-o ProxyCommand=none' ./package_push.sh

# 只部署：使用 release/ 中最新的 tar.gz
./package_push.sh deploy

# 只部署：指定某个 tar.gz
ARTIFACT=release/frontend-dist-20260212152116.tar.gz ./package_push.sh deploy
```
