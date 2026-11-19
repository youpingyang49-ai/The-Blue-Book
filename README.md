# The-Blue-Book 📕

小红书风格的轻量图文社区（笔记分享），Go + React 前后端分离实现，功能完整、开箱即用。

![Go](https://img.shields.io/badge/Go-1.27-00ADD8?style=flat&logo=go&logoColor=white)
![React](https://img.shields.io/badge/React-61DAFB?style=flat&logo=react&logoColor=black)
![TypeScript](https://img.shields.io/badge/TypeScript-3178C6?style=flat&logo=typescript&logoColor=white)
![License](https://img.shields.io/badge/License-MIT-blue?style=flat)

## ✨ 功能特性

- 🔐 手机号 + 验证码登录（JWT 双 token）
- 📝 图文发布：多图上传 + 话题标签
- 🏠 首页 feed：推荐 / 最新 / 热门 / 关注 + 5 列网格
- 🔍 搜索：标题 / 正文 / 话题
- ❤️ 互动：点赞、收藏、评论（嵌套回复）
- 👥 关注：关注 / 取关、关注流、个人主页
- 🔔 消息中心：点赞 / 评论 / 关注 / @提及 / 私信
- ✏️ 帖子管理：编辑 / 删除（仅作者）
- 🤖 推荐算法：热度加权 + 时间衰减
- 🖥 响应式 Web UI（桌面 5 列，窗口缩小自动降列）

## 🛠 技术栈

**后端**

| 分类 | 技术 |
| --- | --- |
| 语言 | Go 1.27 |
| Web 框架 | Gin |
| ORM | GORM |
| 数据库 | MySQL 8.0 |
| 缓存 | Redis 7 |
| 消息队列 | RabbitMQ |
| 鉴权 | JWT（golang-jwt/v5） |
| 对象存储 | 本地磁盘 / MinIO 可切换 |

**前端**

| 分类 | 技术 |
| --- | --- |
| 框架 | React + TypeScript |
| 构建 | Vite |
| 状态管理 | Zustand |
| 路由 | React Router |

## 🚀 快速开始

### 环境要求

- Go 1.27+
- Node.js 20+
- Docker + Docker Compose

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
    │       └── pkg/             # 公共库（jwt、storage）
    ├── frontend/                # React 前端
    │   └── src/
    │       ├── api/             # 接口封装
    │       ├── store/           # 状态管理
    │       ├── components/      # 组件
    │       └── pages/           # 页面
    └── docker-compose.yml       # 本地依赖（MySQL + Redis + RabbitMQ）

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
| JWT_SECRET | the-blue-book-dev-secret-change-me | JWT 密钥（生产务必修改） |
| APP_ENV | development | 环境（development 下返回验证码） |
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
- [x] 首页 feed（最新 / 热门 / 关注）
- [x] 搜索（标题 / 正文 / 话题）
- [x] 点赞 / 收藏 / 评论
- [x] 关注 / 关注流 / 个人主页
- [x] 帖子编辑 / 删除
- [x] 评论回复（嵌套）
- [x] 消息通知 + @提及 + 私信
- [x] 推荐算法
- [x] 前端全部页面

**规划中**

- [ ] 接入真实短信服务
- [ ] 接入 OSS / MinIO 生产对象存储
- [ ] 内容审核（敏感词 / 图片审核）

## 📄 License

[MIT](./LICENSE)
