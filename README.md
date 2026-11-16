# The-Blue-Book —— 小红书类轻量社区项目

一个轻量化的图文分享社区，参考小红书的产品形态，聚焦「内容消费 + 内容生产」的最小闭环。项目采用**前后端分离**架构，后端使用 **Go** 语言实现。

---

## 1. 项目定位

- **定位**：轻量级图文社区（笔记分享），不做视频 / 直播 / 私信等重功能。
- **目标**：跑通「注册登录 → 浏览 feed → 搜索 → 发帖 → 个人中心」完整用户链路。
- **形态**：Web 优先（响应式），后续可扩展小程序 / App 客户端。

---

## 2. 技术栈

### 2.1 后端（Go）

| 领域 | 选型 | 说明 |
| --- | --- | --- |
| Web 框架 | Gin | 轻量、生态成熟、性能好 |
| 数据访问 | GORM（备选 sqlc） | GORM 上手快；追求类型安全用 sqlc |
| 数据库 | MySQL 8.0 | 主流、生态好；中文全文检索用 ngram 解析器 |
| 缓存 | Redis | 热点 feed 缓存、验证码、限流 |
| 对象存储 | 阿里 OSS / 腾讯 COS / Cloudflare R2 | 本地开发用 MinIO |
| 鉴权 | JWT（golang-jwt）+ Refresh Token | Access / Refresh 双 token |
| 配置 | Viper | 多环境配置 |
| 日志 | zap | 结构化日志 |
| 迁移 | golang-migrate | 数据库版本管理 |
| 参数校验 | go-playground/validator | 请求参数校验 |
| 搜索 | MySQL FULLTEXT + ngram | 中文分词用 ngram 解析器；数据量大后尽早切 Meilisearch |

### 2.2 前端

| 领域 | 选型 | 说明 |
| --- | --- | --- |
| 框架 | React + TypeScript + Vite | 主流、生态好 |
| UI 组件库 | Ant Design（备选 shadcn/ui） | 快速出页面 |
| 状态管理 | Zustand | 轻量 |
| 路由 | React Router | — |
| 请求 | Axios + TanStack Query | 缓存与请求管理 |
| 图片上传 | 直传对象存储 | 前端拿签名直传，后端只存 URL |

---

## 3. 功能需求

### 3.1 用户登录体系
- 手机号 + 短信验证码登录（核心）
- 第三方登录预留：微信 / 微博 / GitHub（可选）
- JWT 双 token 机制：Access Token（短期）+ Refresh Token（长期）
- 登录态刷新、登出、token 失效处理
- 验证码：Redis 存储、有效期 5 分钟、发送频率限制

### 3.2 个人中心体系
- 个人资料：头像、昵称、简介、性别、地区
- 我的帖子列表（分页）
- 我的点赞 / 收藏列表
- 关注 / 粉丝列表
- 编辑资料、修改头像（图片上传）

### 3.3 帖子推送页面体系（feed）
- 首页 feed：默认「推荐流」，支持「最新」「热门」排序切换
- 瀑布流 / 卡片式列表（图文卡片）
- 帖子详情页：多图浏览、正文、点赞、收藏、评论、关注作者
- 关注流（第二期）：只展示已关注作者的内容
- 分页：基于游标（cursor）分页，避免深分页问题

### 3.4 搜索帖子体系
- 关键词搜索：标题 + 正文 + 话题标签
- 搜索建议（热门关键词 / 联想）
- 搜索结果分页
- 初期用 MySQL FULLTEXT + ngram 解析器（中文分词）；MySQL 全文检索较弱，数据量大后尽早迁移 Meilisearch

### 3.5 自主创新帖子体系（发帖）
- 图文发布：标题 + 正文 + 多图（最多 9 张）+ 话题标签
- 图片上传：多图、拖拽排序、上传进度、缩略图生成
- 草稿保存（可选）
- 帖子编辑 / 删除（仅作者）
- 内容审核：敏感词过滤 + 图片审核（可接第三方内容安全）

---

## 4. 数据库设计

### 4.1 核心表（v1）

    users          用户表
      - id, phone, password_hash, nickname, avatar_url, bio, gender, region
      - created_at, updated_at, deleted_at

    posts          帖子表
      - id, user_id, title, content, status(草稿/发布/下架), like_count, comment_count
      - favorite_count, created_at, updated_at, deleted_at

    post_images    帖子图片表（一帖多图）
      - id, post_id, image_url, thumb_url, sort_order, width, height

    topics         话题标签表
      - id, name, post_count, hot_score

    post_topics    帖子-话题关联表
      - post_id, topic_id

    likes          点赞表
      - id, user_id, post_id, created_at

    favorites      收藏表
      - id, user_id, post_id, created_at

    follows        关注表
      - id, follower_id, followee_id, created_at

    comments       评论表（第二期）
      - id, post_id, user_id, parent_id, content, like_count, created_at

> 关键索引：posts(created_at)、posts(user_id)、posts(status)、likes(user_id, post_id) 唯一索引、follows(follower_id, followee_id) 唯一索引。

---

## 5. API 设计概要（RESTful）

    # 认证
    POST   /api/v1/auth/send-code       发送验证码
    POST   /api/v1/auth/login           验证码登录
    POST   /api/v1/auth/refresh         刷新 token
    POST   /api/v1/auth/logout          登出

    # 用户
    GET    /api/v1/users/:id            用户主页信息
    PUT    /api/v1/users/me             更新个人资料
    GET    /api/v1/users/:id/posts      某用户的帖子列表
    GET    /api/v1/users/:id/likes      某用户的点赞列表
    GET    /api/v1/users/:id/favorites  某用户的收藏列表
    GET    /api/v1/users/:id/followers  粉丝列表
    GET    /api/v1/users/:id/following  关注列表
    POST   /api/v1/users/:id/follow     关注
    DELETE /api/v1/users/:id/follow     取关

    # 帖子
    GET    /api/v1/posts                首页 feed（分页、排序）
    POST   /api/v1/posts                发布帖子
    GET    /api/v1/posts/:id            帖子详情
    PUT    /api/v1/posts/:id            编辑帖子（仅作者）
    DELETE /api/v1/posts/:id            删除帖子（仅作者）
    POST   /api/v1/posts/:id/like       点赞
    DELETE /api/v1/posts/:id/like       取消点赞
    POST   /api/v1/posts/:id/favorite   收藏
    DELETE /api/v1/posts/:id/favorite   取消收藏

    # 搜索
    GET    /api/v1/search               搜索帖子
    GET    /api/v1/search/suggestions   搜索联想

    # 上传
    POST   /api/v1/upload/signature     获取直传签名（或后端中转上传）

---

## 6. 目录结构建议

    xiaohongshu-lite/
    ├── README.md
    ├── backend/                    # Go 后端
    │   ├── cmd/server/main.go      # 入口
    │   ├── internal/
    │   │   ├── config/             # 配置
    │   │   ├── handler/            # HTTP 处理器
    │   │   ├── service/            # 业务逻辑
    │   │   ├── repository/         # 数据访问
    │   │   ├── model/              # 数据模型
    │   │   ├── middleware/         # 中间件（鉴权、日志、限流）
    │   │   ├── router/             # 路由
    │   │   └── pkg/                # 公共工具（jwt、redis、storage）
    │   ├── migrations/             # 数据库迁移
    │   └── go.mod
    ├── frontend/                   # React 前端
    │   ├── src/
    │   │   ├── pages/              # 页面
    │   │   ├── components/         # 组件
    │   │   ├── api/                # 接口封装
    │   │   ├── store/              # 状态管理
    │   │   └── utils/
    │   └── package.json
    └── docker-compose.yml          # 本地环境（MySQL + Redis + MinIO）

---

## 7. MVP 范围与里程碑

### 第一期（跑通核心闭环）
- ✅ 手机号验证码登录
- ✅ 发布图文（多图上传）
- ✅ 首页 feed（最新 / 热门排序）
- ✅ 帖子详情
- ✅ 点赞 / 收藏
- ✅ 个人中心（资料 + 我的帖子）
- ✅ 基础搜索

### 第二期
- 关注流 / 关注作者
- 评论
- 搜索联想 + 热门话题
- 内容审核接入
- 草稿箱

### 第三期
- 推荐算法（热度 + 个性化）
- 消息通知
- 小程序 / App 客户端
- Meilisearch 全文检索

---

## 8. 安全与合规

- **内容审核**：敏感词过滤（本地词库 + 第三方内容安全 API），图片审核。
- **鉴权**：token 加密存储，接口限流，防刷。
- **合规**：国内上线需 ICP 备案；UGC 平台需落实内容安全主体责任。
- **数据安全**：用户数据脱敏，对象存储私有读写 + 签名 URL。

---

## 9. 快速开始

### 环境要求
- Go 1.27+ · Node.js 20+ · Docker + Docker Compose

### 启动步骤

    # 1. 本地依赖（MySQL + Redis）
    docker compose up -d

    # 2. 后端
    cd backend && go run ./cmd/server
    curl http://127.0.0.1:8080/api/v1/health   # => {"db":"connected","status":"ok"}

    # 3. 前端
    cd frontend && npm install && npm run dev   # /api 已代理到后端 8080

### 配置
环境变量见 backend/.env.example（PORT / DATABASE_URL / REDIS_ADDR）。
