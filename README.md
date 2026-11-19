# The-Blue-Book 📕

一个小红书风格的轻量图文社区（笔记分享），Go + React 前后端分离实现，功能完整、开箱即用。

![Go](https://img.shields.io/badge/Go-1.27-00ADD8?style=flat&logo=go&logoColor=white)
![React](https://img.shields.io/badge/React-19-61DAFB?style=flat&logo=react&logoColor=black)
![TypeScript](https://img.shields.io/badge/TypeScript-7-3178C6?style=flat&logo=typescript&logoColor=white)
![License](https://img.shields.io/badge/License-MIT-blue?style=flat)

## 📸 界面预览

| 登录页 | 社区首页（瀑布流） |
| --- | --- |
| ![登录页](docs/screenshots/login.webp) | ![首页](docs/screenshots/home.webp) |

| 笔记详情 | 个人主页 |
| --- | --- |
| ![笔记详情](docs/screenshots/post-detail.webp) | ![个人主页](docs/screenshots/profile.webp) |

<div align="center">
  <img src="docs/screenshots/home-mobile.webp" width="300" alt="移动端首页" />
  <p><i>移动端：底部导航 + 双列瀑布流</i></p>
</div>

## ✨ 功能特性

- 🔐 手机号 + 验证码登录（JWT 双 token，未登录访问受限页自动跳登录并回跳）
- 📝 图文发布：多图上传、话题标签、图片预览与移除
- 🏠 首页 feed：推荐 / 最新 / 热门 / 关注，四类 Tab 对应真实接口
- 🧱 瀑布流：2 / 3 / 4 / 5 列随窗口自适应，卡片高度随真实图片比例变化
- 🔍 搜索：标题 / 正文 / 话题，支持分页加载与热门关键词
- ❤️ 互动：点赞、收藏、评论（嵌套回复）、关注 / 取关
- 🔔 消息中心：点赞 / 收藏 / 评论 / 关注 / @提及 / 私信
- 💬 私信：会话列表 + 未读角标 + 气泡聊天 + 清空记录
- 👤 个人中心：我的笔记 / 我的点赞 / 我的收藏、粉丝关注统计、退出登录
- ✏️ 帖子管理：编辑 / 删除（仅作者）
- 🤖 推荐算法：热度 + 时间衰减 + 话题个性化 + 冷启动兜底
- 🖥 统一设计语言：一套 token 驱动的响应式 Web UI

## 🛠 技术栈

**后端**

| 分类 | 技术 |
| --- | --- |
| 语言 | Go 1.27 |
| Web 框架 | Gin |
| ORM | GORM |
| 数据库 | MySQL 8.0 |
| 缓存 | Redis 7 |
| 消息队列 | RabbitMQ（通知异步化，连不上自动降级为同步） |
| 鉴权 | JWT（golang-jwt/v5） |
| 对象存储 | 本地磁盘 / MinIO 可切换 |

**前端**

| 分类 | 技术 |
| --- | --- |
| 框架 | React 19 + TypeScript |
| 构建 | Vite 8 |
| 状态管理 | Zustand |
| 路由 | React Router 7 |
| 样式 | 原生 CSS（设计 token + 分层样式文件） |
| 校验 | oxlint |

## 🎨 设计系统

前端样式按职责拆分为 4 个文件，统一由 `src/index.css` 引入：

| 文件 | 职责 |
| --- | --- |
| `styles/tokens.css` | 色板、字号、圆角、阴影、动效、重置样式 |
| `styles/layout.css` | 应用骨架：固定侧栏、顶部搜索、移动端底部导航 |
| `styles/components.css` | 按钮、表单、头像、卡片、Tab、骨架屏、弹层、Toast |
| `styles/pages.css` | 登录、Feed、详情、主页、消息、私信、编辑器等页面级布局 |

核心组件（`src/components/`）：

| 组件 | 说明 |
| --- | --- |
| `AppShell` | 统一壳层：侧栏 + 吸顶搜索 + 内容区 + 移动端导航 + 关于弹窗 |
| `Sidebar` / `HeaderSearch` | 固定导航（含未读角标）、全局搜索 |
| `MasonryFeed` | 自研瀑布流：最短列放置、保留 DOM 顺序、图片加载后自动重排 |
| `PostCard` | 笔记卡片：懒加载封面、话题、作者、可交互点赞（乐观更新） |
| `FeedTabs` / `States` / `Toast` / `Avatar` / `Icon` | 分类切换、加载/空/错误态、轻提示、头像、内联图标 |

响应式断点：`≤900px` 收起侧栏改用底部导航，`≥1700px` 内容区放宽至 5 列。

## 🚀 快速开始

### 环境要求

- Go 1.27+
- Node.js 20+
- Docker + Docker Compose（或自备 MySQL 8 / Redis 7 / RabbitMQ）

### 启动

    # 1. 启动本地依赖（MySQL + Redis + RabbitMQ）
    docker compose up -d

    # 2. 启动后端（端口 8080）
    cd backend
    go run ./cmd/server

    # 3. 启动前端（端口 5173）
    cd frontend
    npm install
    npm run dev

浏览器打开 `http://localhost:5173`。

> 开发环境下，登录验证码会直接显示在页面提示里（mock 模式），无需真实短信。

> RabbitMQ 为可选依赖：未启动时后端会打印 `fallback to sync notification`，通知改为同步写入，其余功能不受影响。

### 构建检查

    cd frontend
    npm run build     # tsc -b && vite build
    npm run lint      # oxlint

## 📁 项目结构

    The-Blue-Book/
    ├── backend/                 # Go 后端
    │   ├── cmd/server/          # 程序入口
    │   └── internal/
    │       ├── config/          # 配置
    │       ├── model/           # 数据模型
    │       ├── repository/      # 数据访问层
    │       ├── service/         # 业务逻辑层
    │       ├── handler/         # HTTP 处理器
    │       ├── middleware/      # 中间件（鉴权、CORS）
    │       ├── router/          # 路由
    │       └── pkg/             # 公共库（jwt、storage、mq）
    ├── frontend/                # React 前端
    │   └── src/
    │       ├── api/             # 接口封装
    │       ├── store/           # 状态管理（Zustand）
    │       ├── components/      # 通用组件（壳层/瀑布流/卡片/状态）
    │       ├── styles/          # 设计 token 与分层样式
    │       └── pages/           # 页面
    ├── docs/screenshots/        # README 预览图
    └── docker-compose.yml       # 本地依赖（MySQL + Redis + RabbitMQ + MinIO 可选）

## 📡 API 概览

    # 认证
    POST   /api/v1/auth/send-code        发送验证码
    POST   /api/v1/auth/login            验证码登录

    # 帖子
    GET    /api/v1/posts                 首页 feed（sort=recommend|latest|hot|following）
    POST   /api/v1/posts                 发布帖子
    GET    /api/v1/posts/:id             帖子详情
    PUT    /api/v1/posts/:id             编辑帖子（仅作者）
    DELETE /api/v1/posts/:id             删除帖子（仅作者）

    # 互动
    POST   /api/v1/posts/:id/like        点赞
    DELETE /api/v1/posts/:id/like        取消点赞
    POST   /api/v1/posts/:id/favorite    收藏
    DELETE /api/v1/posts/:id/favorite    取消收藏
    POST   /api/v1/posts/:id/comments    发评论
    GET    /api/v1/posts/:id/comments    评论列表

    # 用户 / 关注
    GET    /api/v1/users/me              当前用户
    GET    /api/v1/users/:id             用户主页
    GET    /api/v1/users/:id/posts       某用户的帖子
    POST   /api/v1/users/:id/follow      关注
    DELETE /api/v1/users/:id/follow      取关

    # 消息 / 通知
    GET    /api/v1/notifications           通知列表（type=like|comment|follow|mention）
    GET    /api/v1/notifications/unread-count  未读数
    POST   /api/v1/notifications/read      全部已读
    GET    /api/v1/conversations           私信会话列表
    GET    /api/v1/messages/:id            聊天记录
    POST   /api/v1/messages/:id            发私信

    # 其他
    GET    /api/v1/search                搜索（标题/正文/话题）
    POST   /api/v1/upload                图片上传
    GET    /api/v1/health                健康检查

## ⚙️ 配置

后端通过环境变量配置，见 `backend/.env.example`：

| 变量 | 默认值 | 说明 |
| --- | --- | --- |
| PORT | 8080 | 服务端口 |
| DATABASE_URL | bluebook:bluebook@tcp(127.0.0.1:3306)/bluebook?... | MySQL 连接串 |
| REDIS_ADDR | 127.0.0.1:6379 | Redis 地址 |
| RABBITMQ_URL | amqp://guest:guest@127.0.0.1:5672/ | RabbitMQ 地址（留空则不同步通知） |
| JWT_SECRET | the-blue-book-dev-secret-change-me | JWT 密钥（生产务必修改） |
| APP_ENV | development | 环境（development 下返回验证码） |
| UPLOAD_DIR | ./uploads | 本地图片目录 |
| PUBLIC_BASE_URL | http://127.0.0.1:8080 | 图片访问前缀 |
| MINIO_ENDPOINT | （空） | 留空用本地磁盘，设置后切 MinIO |

## ⚠️ 生产部署注意

- 验证码当前为 mock 模式（开发环境直接返回），生产需接入真实短信服务
- 对象存储默认本地磁盘，生产建议切换 MinIO / OSS（设置 MINIO_ENDPOINT）
- JWT_SECRET 生产环境务必修改为强随机密钥
- 国内上线需 ICP 备案，UGC 内容需接入内容审核

## 🗺 路线图

**已完成**

- [x] 手机号验证码登录（JWT 双 token）
- [x] 图文发布（多图 + 话题标签）
- [x] 首页 feed（推荐 / 最新 / 热门 / 关注）
- [x] 搜索（标题 / 正文 / 话题）
- [x] 点赞 / 收藏 / 评论
- [x] 关注 / 关注流 / 个人主页
- [x] 帖子编辑 / 删除
- [x] 评论回复（嵌套）
- [x] 消息通知 + @提及 + 私信
- [x] 推荐算法（热度 + 时间衰减 + 话题个性化）
- [x] RabbitMQ 通知异步化（不可用时自动降级）
- [x] 前端全站 UI 重构：固定侧栏 + 顶部搜索 + 瀑布流 + 统一设计系统
- [x] 响应式适配（桌面 / 平板 / 移动端）

**规划中**

- [ ] 接入真实短信服务
- [ ] 接入 OSS / MinIO 生产对象存储
- [ ] 内容审核（敏感词 / 图片审核）

## 📄 License

[MIT](./LICENSE)
