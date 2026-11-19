# 贡献指南

感谢你对 The Blue Book 的关注！无论是提交 Issue、修正文档，还是贡献代码，都非常欢迎。

## 如何贡献

1. Fork 本仓库并克隆到本地
2. 从 `main` 切出功能分支：`git checkout -b feat/your-feature`
3. 完成开发与自测
4. 提交前运行检查（见下文）
5. 提交 Pull Request，并描述改动内容与验证方式

## 开发环境

```bash
# 本地依赖（MySQL / Redis / RabbitMQ）
docker compose up -d

# 后端
cd backend && go run ./cmd/server

# 前端
cd frontend && npm install && npm run dev
```

开发环境（`APP_ENV=development`）下验证码会直接返回在登录页提示中，无需真实短信。

## 提交前检查

```bash
# 前端
cd frontend
npm run build
npm run lint

# 后端
cd backend
go build ./...
go vet ./...
```

请确保以上命令全部通过。若新增了接口或页面，请在 PR 描述中说明验证步骤。

## 提交信息规范

使用 [Conventional Commits](https://www.conventionalcommits.org/)：

```text
feat(frontend): 新增关注 / 粉丝列表弹窗
fix(backend): 修复列表接口作者头像为空的问题
docs: 更新 README 界面预览
refactor: 抽取瀑布流布局组件
chore: 升级依赖版本
```

- `feat` 新功能
- `fix` 缺陷修复
- `docs` 文档变更
- `refactor` 不改变行为的重构
- `chore` 构建 / 依赖 / 工具链调整

## 代码约定

### 后端（Go）

- 保持 `handler → service → repository` 的分层，不要在 handler 中直接写 SQL
- 新增接口需在 `internal/router/router.go` 注册，并遵循现有 JSON 返回风格
- 查询注意 N+1，必要时批量查询
- 错误信息面向用户，内部错误记录日志

### 前端（TypeScript / React）

- 组件放在 `src/components`，页面放在 `src/pages`
- 样式遵循 `src/styles` 的分层：tokens / layout / components / pages
- 不要引入不必要的第三方依赖，优先使用现有组件与设计 token
- 所有异步请求需要处理加载、空、错误三种状态
- 保持 TypeScript 严格模式通过（无 `any` 泛滥、无未使用变量）

## Issue 规范

提交 Issue 时请尽量包含：

- 问题描述与复现步骤
- 期望行为与实际行为
- 环境信息（操作系统、浏览器、Go / Node 版本）
- 相关日志或截图

## Pull Request 规范

- 一个 PR 只做一件事，避免混杂与功能无关的格式化改动
- 描述中说明：改了什么、为什么改、如何验证
- 涉及 UI 的改动请附截图（桌面端与移动端）
- 保持提交历史清晰，必要时可 squash

## 行为准则

请保持友善与尊重，针对事不对人。任何形式的骚扰、歧视性言论都不被接受。

---

再次感谢你的贡献！
