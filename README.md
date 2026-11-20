<div align="center">

# The Blue Book

**一个轻量、现代的图文笔记社区 —— Go + React 全栈实现**

记录灵感、分享生活，让每一次发现都值得被看见。

[![Go](https://img.shields.io/badge/Go-1.27-00ADD8?logo=go&logoColor=white)](https://go.dev/)
[![React](https://img.shields.io/badge/React-19-087EA4?logo=react&logoColor=white)](https://react.dev/)
[![TypeScript](https://img.shields.io/badge/TypeScript-7-3178C6?logo=typescript&logoColor=white)](https://www.typescriptlang.org/)
[![Vite](https://img.shields.io/badge/Vite-8-646CFF?logo=vite&logoColor=white)](https://vitejs.dev/)
[![MySQL](https://img.shields.io/badge/MySQL-8.0-4479A1?logo=mysql&logoColor=white)](https://www.mysql.com/)
[![Redis](https://img.shields.io/badge/Redis-7-DC382D?logo=redis&logoColor=white)](https://redis.io/)
[![RabbitMQ](https://img.shields.io/badge/RabbitMQ-4.x-FF6600?logo=rabbitmq&logoColor=white)](https://www.rabbitmq.com/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](./LICENSE)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](CONTRIBUTING.md)

[功能特性](#-功能特性) · [界面预览](#-界面预览) · [快速开始](#-快速开始) · [API 概览](#-api-概览) · [贡献指南](#-贡献指南)

</div>

---

## 📖 目录

- [项目简介](#-项目简介)
- [界面预览](#-界面预览)
- [功能特性](#-功能特性)
- [技术栈](#-技术栈)
- [系统架构](#-系统架构)
- [快速开始](#-快速开始)
- [配置说明](#-配置说明)
- [项目结构](#-项目结构)
- [API 概览](#-api-概览)
- [开发与测试](#-开发与测试)
- [部署建议](#-部署建议)
- [路线图](#-路线图)
- [贡献指南](#-贡献指南)
- [许可证](#-许可证)

## 💡 项目简介

**The Blue Book** 是一个面向普通用户的轻量图文社区，支持发布图文笔记、浏览瀑布流内容、关注创作者，并通过点赞、收藏、评论和私信进行互动。

项目采用前后端分离架构：后端使用 Go + Gin + GORM，前端使用 React + TypeScript + Vite，依赖 MySQL / Redis / RabbitMQ。前端基于原生 CSS 构建了一套完整的设计系统，未引入 UI 组件库。

## 📸 界面预览

<table>
  <tr>
    <td width="50%"><img src="docs/screenshots/login.webp" alt="登录页" /></td>
    <td width="50%"><img src="docs/screenshots/home.webp" alt="社区首页" /></td>
  </tr>
  <tr>
    <td align="center"><sub>登录页 · 手机号验证码登录</sub></td>
    <td align="center"><sub>首页 · 侧边导航 + 搜索 + 瀑布流</sub></td>
  </tr>
  <tr>
    <td width="50%"><img src="docs/screenshots/discover.webp" alt="发现页" /></td>
    <td width="50%"><img src="docs/screenshots/post-detail.webp" alt="笔记详情" /></td>
  </tr>
  <tr>
    <td align="center"><sub>发现页 · 热门笔记与话题</sub></td>
    <td align="center"><sub>详情页 · 图文、点赞收藏与评论</sub></td>
  </tr>
  <tr>
    <td width="50%"><img src="docs/screenshots/profile.webp" alt="个人主页" /></td>
    <td width="50%"><img src="docs/screenshots/notifications.webp" alt="消息中心" /></td>
  </tr>
  <tr>
    <td align="center"><sub>个人主页 · 笔记 / 点赞 / 收藏 / 关注粉丝</sub></td>
    <td align="center"><sub>消息中心 · 点赞 / 评论 / 关注 / @ / 私信</sub></td>
  </tr>
</table>

<div align="center">
  <img src="docs/screenshots/home-mobile.webp" width="300" alt="移动端首页" />
  <p><sub>移动端 · 底部导航 + 双列瀑布流</sub></p>
</div>

## ✨ 功能特性

### 内容与创作

- 图文笔记发布：多图上传、话题标签、图片预览与移除；标题、正文、话题均可选
- 首页四类 Feed：推荐 / 最新 / 热门 / 关注
- 推荐算法：热度加权 + 时间衰减 + 话题个性化 + 冷启动兜底
- 瀑布流布局：2 / 3 / 4 / 5 列随窗口自适应
- 搜索：笔记（标题 / 正文 / 话题）与用户（昵称 / 简介）双分页，支持分页加载

### 互动与社交

- 点赞、收藏、评论（支持嵌套回复与 @提及）
- 关注 / 取关、关注流、粉丝与关注列表
- 笔记编辑 / 删除（仅作者）
- 他人主页可直接发起私信

### 消息通知

- 消息中心：点赞、收藏、评论、关注、@提及分类查看，分类未读角标，查看后自动清除
- 私信：会话列表、未读角标、气泡聊天、清空记录
- RabbitMQ 异步通知，不可用时自动降级为同步写入

### 账号与安全

- 手机号 + 短信验证码登录，未注册自动创建账号
- JWT 双 Token，Access 过期后自动静默续期（令牌轮换）
- 未登录访问受限页面跳转登录，登录后回跳原目标

### 体验与工程

- 统一设计系统（Design Tokens + 分层样式）
- 完整的加载 / 空 / 错误状态与骨架屏
- 响应式适配：桌面、平板、移动端
- TypeScript 严格模式 + oxlint

## 🛠 技术栈

| 分类 | 技术 |
| --- | --- |
| 后端 | Go 1.27 · Gin · GORM · golang-jwt/v5 |
| 前端 | React 19 · TypeScript · Vite 8 · React Router 7 · Zustand |
| 存储 | MySQL 8.0 · Redis 7 · 本地磁盘 / MinIO（可切换） |
| 消息 | RabbitMQ（通知异步化） |
| 样式 | 原生 CSS（Design Tokens + 分层样式）+ 自研瀑布流 |
| 工具 | oxlint · Docker Compose |

## 🏗 系统架构

```text
┌──────────────────────────────┐
│  React SPA（Vite + Router）  │
│  路由守卫 · Zustand 状态管理  │
└───────────────┬──────────────┘
                │ HTTP / JSON  /api/v1
┌───────────────▼──────────────┐
│  Gin API                     │
│  Handler → Service → Repo    │
│  JWT 鉴权 · CORS             │
└──┬─────────┬─────────┬───────┘
   │         │         │
 ┌─▼──┐   ┌──▼───┐  ┌──▼──────┐
 │MySQL│  │Redis │  │RabbitMQ │
 └────┘   └──────┘  └─────────┘
   │
 ┌─▼───────────────────────────┐
 │ 对象存储：本地磁盘 / MinIO   │
 └─────────────────────────────┘
```

- 通知默认通过 RabbitMQ 异步投递，MQ 不可用时自动降级为同步写库。
- 图片默认存本地磁盘（`UPLOAD_DIR`），配置 `MINIO_ENDPOINT` 后切换 MinIO。

## 🚀 快速开始

### 环境要求

| 依赖 | 版本 |
| --- | --- |
| Go | 1.27+ |
| Node.js | 20+ |
| MySQL | 8.0+ |
| Redis | 7+ |
| RabbitMQ | 可选（未启动时使用同步通知） |
| Docker | 可选（用于一键启动依赖） |

### 方式一：Docker Compose 启动依赖

```bash
docker compose up -d
```

该命令启动 MySQL（3306）、Redis（6379）、RabbitMQ（5672 / 15672），并自动创建 `bluebook` 数据库。

### 方式二：使用已有服务

若本机已有 MySQL / Redis，确保连接信息与 [配置说明](#-配置说明) 一致，并手动创建数据库：

```sql
CREATE DATABASE bluebook DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

### 启动后端

```bash
cd backend
go mod download
go run ./cmd/server
```

服务默认监听 `http://127.0.0.1:8080`，健康检查：

```bash
curl http://127.0.0.1:8080/api/v1/health
# {"db":"connected","status":"ok"}
```

### 启动前端

```bash
cd frontend
npm install
npm run dev
```

打开 `http://localhost:5173`，使用手机号登录。

> 开发环境（`APP_ENV=development`）不会真实发送短信，验证码会直接显示在登录页提示中。

## ⚙️ 配置说明

后端通过环境变量配置，默认值见 [`backend/.env.example`](backend/.env.example)：

| 变量 | 默认值 | 说明 |
| --- | --- | --- |
| `PORT` | `8080` | HTTP 服务端口 |
| `DATABASE_URL` | `bluebook:bluebook@tcp(127.0.0.1:3306)/bluebook?...` | MySQL 连接串 |
| `REDIS_ADDR` | `127.0.0.1:6379` | Redis 地址 |
| `RABBITMQ_URL` | `amqp://guest:guest@127.0.0.1:5672/` | RabbitMQ 地址，留空则关闭 MQ |
| `JWT_SECRET` | `the-blue-book-dev-secret-change-me` | JWT 签名密钥（生产必须修改） |
| `APP_ENV` | `development` | 环境标识，development 下返回验证码 |
| `UPLOAD_DIR` | `./uploads` | 本地图片存储目录 |
| `PUBLIC_BASE_URL` | `http://127.0.0.1:8080` | 图片访问前缀 |
| `MINIO_ENDPOINT` | *(空)* | 留空使用本地磁盘，设置后切换 MinIO |
| `MINIO_ACCESS_KEY` / `MINIO_SECRET_KEY` | `minioadmin` | MinIO 凭证 |
| `MINIO_BUCKET` | `bluebook` | MinIO Bucket |

## 📁 项目结构

```text
The-Blue-Book/
├── backend/                        # Go 后端
│   ├── cmd/server/                 # 程序入口
│   └── internal/
│       ├── config/                 # 环境变量配置
│       ├── model/                  # GORM 数据模型
│       ├── repository/             # 数据访问层
│       ├── service/                # 业务逻辑层
│       ├── handler/                # HTTP 处理器
│       ├── middleware/             # JWT 鉴权、CORS
│       ├── router/                 # 路由注册
│       └── pkg/                    # jwt / storage / mq 公共库
├── frontend/                       # React 前端
│   └── src/
│       ├── api/                    # 接口封装
│       ├── store/                  # Zustand 状态
│       ├── components/             # 通用组件
│       │   ├── AppShell.tsx        # 应用壳层（侧栏 + 搜索 + 内容区）
│       │   ├── Sidebar.tsx         # 固定侧边导航
│       │   ├── HeaderSearch.tsx    # 全局搜索
│       │   ├── MasonryFeed.tsx     # 瀑布流布局
│       │   ├── PostCard.tsx        # 笔记卡片
│       │   ├── FollowListModal.tsx # 关注 / 粉丝列表
│       │   └── States.tsx          # 加载 / 空 / 错误态
│       ├── styles/                 # 设计系统（tokens / layout / components / pages）
│       └── pages/                  # 页面：首页、发现、发布、详情、消息、私信、个人中心
├── docs/screenshots/               # README 预览图
├── docker-compose.yml              # 本地依赖（MySQL / Redis / RabbitMQ / MinIO）
└── README.md
```

## 📡 API 概览

所有接口以 `/api/v1` 为前缀，除标注外均返回 JSON。

<details>
<summary><b>认证</b></summary>

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| POST | `/auth/send-code` | 发送短信验证码（开发环境返回 `debug_code`） |
| POST | `/auth/login` | 验证码登录，返回 `access_token` / `refresh_token` |
| POST | `/auth/refresh` | 用 `refresh_token` 换取新的 access / refresh token |

</details>

<details>
<summary><b>笔记</b></summary>

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | `/posts?sort=recommend\|latest\|hot\|following` | 首页 Feed |
| POST | `/posts` | 发布笔记（需登录） |
| GET | `/posts/:id` | 笔记详情 |
| PUT | `/posts/:id` | 编辑笔记（仅作者） |
| DELETE | `/posts/:id` | 删除笔记（仅作者） |
| GET | `/search?q=` | 搜索标题 / 正文 / 话题 |
| GET | `/search/users?q=` | 搜索用户（昵称 / 简介） |

</details>

<details>
<summary><b>互动</b></summary>

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| POST / DELETE | `/posts/:id/like` | 点赞 / 取消点赞 |
| POST / DELETE | `/posts/:id/favorite` | 收藏 / 取消收藏 |
| GET / POST | `/posts/:id/comments` | 评论列表 / 发表评论 |

</details>

<details>
<summary><b>用户与关注</b></summary>

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | `/users/me` | 当前登录用户 |
| GET | `/users/:id` | 用户主页（含粉丝 / 关注数） |
| GET | `/users/:id/posts` | 用户的笔记 |
| GET | `/users/:id/following` | 关注列表 |
| GET | `/users/:id/followers` | 粉丝列表 |
| POST / DELETE | `/users/:id/follow` | 关注 / 取关 |

</details>

<details>
<summary><b>消息与私信</b></summary>

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | `/notifications?type=` | 通知列表（like / comment / follow / mention） |
| GET | `/notifications/unread-count` | 未读数（总数 / 分类明细 / 私信） |
| POST | `/notifications/read` | 标记已读，可传 `types` 按分类处理 |
| GET | `/conversations` | 私信会话列表 |
| GET / POST | `/messages/:id` | 聊天记录 / 发送私信 |
| POST | `/messages/:id/unread` | 标记会话未读 |
| POST | `/conversations/:id/hide` | 隐藏会话 |

</details>

<details>
<summary><b>其他</b></summary>

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| POST | `/upload` | 图片上传（需登录） |
| GET | `/health` | 健康检查（含数据库连通性） |

</details>

## 🧪 开发与测试

```bash
# 前端类型检查 + 生产构建
cd frontend
npm run build

# 前端代码规范检查
npm run lint

# 后端编译与静态检查
cd backend
go build ./...
go vet ./...
```

前端构建产物输出到 `frontend/dist/`，可直接配合 Nginx 部署。

## 🚢 部署建议

1. **反向代理**：使用 Nginx 托管 `frontend/dist`，并将 `/api` 与 `/uploads` 反向代理到 Go 服务。
2. **进程管理**：后端以 systemd / supervisor 托管，配置 `APP_ENV=production`。
3. **密钥与凭证**：务必修改 `JWT_SECRET`，数据库与 MinIO 使用独立账号。
4. **对象存储**：生产环境建议启用 MinIO / OSS，避免本地磁盘单点。
5. **短信服务**：接入真实短信服务商，替换当前 mock 验证码逻辑。
6. **合规**：国内部署需 ICP 备案；UGC 产品需接入内容审核与举报机制。

## 🗺 路线图

- [x] 手机号验证码登录（JWT 双 Token）
- [x] 图文发布（多图 + 话题标签）
- [x] 首页 Feed（推荐 / 最新 / 热门 / 关注）
- [x] 搜索（标题 / 正文 / 话题）
- [x] 点赞 / 收藏 / 评论（嵌套回复）
- [x] 关注 / 关注流 / 粉丝与关注列表
- [x] 消息通知 + @提及 + 私信
- [x] 推荐算法（热度 + 时间衰减 + 话题个性化）
- [x] RabbitMQ 通知异步化（不可用时自动降级）
- [x] 全站 UI：固定侧栏 + 顶部搜索 + 瀑布流 + 设计系统
- [x] 响应式适配（桌面 / 平板 / 移动端）
- [x] Access Token 过期自动静默续期
- [ ] 接入真实短信服务
- [ ] 接入 OSS / MinIO 生产对象存储
- [ ] 内容审核（敏感词 / 图片审核）
- [ ] 单元测试与接口测试

## 🤝 贡献指南

欢迎提交 Issue 和 Pull Request。开始之前请阅读 [CONTRIBUTING.md](CONTRIBUTING.md)。

简要约定：

- 提交信息遵循 [Conventional Commits](https://www.conventionalcommits.org/)（`feat:` / `fix:` / `docs:` / `refactor:`）
- 新增功能请补充测试或说明验证方式
- 提交前确保 `npm run build`、`npm run lint`、`go build ./...` 通过
- 请勿在 PR 中混入与改动无关的格式化内容

## 📄 许可证

本项目基于 [MIT License](./LICENSE) 开源。

## 🙏 致谢

- 产品交互与视觉参考自小红书 Web 端
- 感谢所有 Issue、Star 和 PR 的贡献者
