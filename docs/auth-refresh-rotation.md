# Refresh Token 轮换与重放防护

## 1. 问题

只做「轮换」是不够的：服务端签发新 refresh token 后，旧 token 在 JWT 过期前仍然能通过签名校验。攻击者一旦拿到旧 token，可以和正常客户端一样持续续期，服务端无法察觉。

修复思路是给每张 refresh token 一个唯一 `jti`，并在 Redis 里记录「当前唯一有效的那一张」：

- 旧 token 用过一次后立刻从白名单消失；
- 再次出现时，要么是并发刷新（可幂等），要么是凭证重放（吊销整条链）。

## 2. 令牌家族（token family）

- 每次登录生成一个随机 `fid`（family id），家族内所有 refresh token 共享它；
- 家族内 `gen` 逐次递增（仅用于排查）；
- Redis 只保存家族当前有效的 `jti`：

| Key | 值 | TTL |
| --- | --- | --- |
| `bluebook:auth:family:<fid>` | 当前有效 refresh token 的 `jti` | `REFRESH_TOKEN_TTL + 1min` |
| `bluebook:auth:family:revoked:<fid>` | `"1"` | `REFRESH_TOKEN_TTL + 1min` |
| `bluebook:auth:grace:<旧jti>` | 轮换后签发的 token 对（JSON） | 60s |

access token 里同样带 `fid`，这样家族被吊销后，该家族的 access token 也会立即失效（中间件每次请求检查一次吊销标记）。

## 3. 刷新时序

```text
客户端                         AuthService                       Redis
  │  POST /auth/refresh(R1)        │                               │
  │───────────────────────────────>│ 校验 JWT（typ=refresh, jti, fid）
  │                                │ 查吊销标记 ──────────────────>│
  │                                │ 签发 access(A2) + refresh(R2, gen+1)
  │                                │ Lua: 校验 family[fid]==jti(R1)
  │                                │      SET family[fid]=jti(R2)
  │                                │      SET grace[jti(R1)]={A2,R2} EX 60
  │                                │<──────────────────────────────│
  │<── 200 {A2, R2} ───────────────│                               │
  │  POST /auth/refresh(R1) 再次   │                               │
  │───────────────────────────────>│ Lua 返回 -3（当前 jti 已变）   │
  │                                │ 读 grace[jti(R1)] ───────────>│
  │<── 200 {A2, R2}（幂等）────────│                               │
```

Lua 脚本把「校验旧 jti + 写入新 jti + 写宽限缓存」做成一个原子操作，避免两个并发请求都轮换成功：

```lua
if redis.call('EXISTS', revoked) == 1 then return -1 end
local current = redis.call('GET', family)
if not current then return -2 end
if current ~= oldJti then return -3 end
redis.call('SET', family, newJti, 'EX', ttl)
redis.call('SET', grace, payload, 'EX', 60)
return 1
```

## 4. 为什么要 60 秒宽限窗口

多标签页、弱网重试、页面刷新都可能让同一个旧 token 被提交两次。如果第二次直接判定为攻击，正常用户会被踢下线。

因此 60 秒内重复提交**同一旧 token**时，返回第一次轮换产生的那一对 token（幂等），不吊销家族；超过 60 秒后再次出现同一旧 token，则判定为泄露：

- 写入 `family:revoked:<fid>`；
- 删除 `family:<fid>`；
- 该家族的 refresh token 与 access token 全部失效，用户需要重新登录。

这是安全性与可用性的折中：宽限窗口内攻击者最多复用一次已发出的新 token 对，且无法延长家族寿命；窗口外一律吊销整条链。

## 5. 登出

`POST /auth/logout` 接收 refresh token，解析出 `fid` 后吊销整个家族。接口是幂等的：token 已过期或格式错误也返回成功。前端退出登录时先调用该接口，再清理本地存储。

## 6. 失败模式与取舍

| 场景 | 行为 |
| --- | --- |
| Redis 不可用（登录） | 拒绝签发 token（无法保证单次使用），返回 500 |
| Redis 不可用（刷新） | 拒绝刷新，返回 500 |
| Redis 不可用（鉴权中间件查吊销标记） | 放行（fail-open），只记日志；access token 本身仍是短 TTL |
| 旧版本签发、没有 `jti` / `fid` 的 refresh token | 拒绝，要求重新登录（部署后需要重新登录一次） |
| 家族标记过期 | 按「登录过期」处理，返回 401 |

refresh token 的 TTL 默认 7 天（`REFRESH_TOKEN_TTL`），access token 默认 15 分钟（`ACCESS_TOKEN_TTL`）；两者都可通过环境变量调整。

## 7. 手工验证

```bash
BASE=http://127.0.0.1:8080/api/v1
PHONE=13900000001

CODE=$(curl -s -X POST $BASE/auth/send-code -H 'Content-Type: application/json' \
  -d "{\"phone\":\"$PHONE\"}" | jq -r .debug_code)

R1=$(curl -s -X POST $BASE/auth/login -H 'Content-Type: application/json' \
  -d "{\"phone\":\"$PHONE\",\"code\":\"$CODE\"}" | jq -r .refresh_token)

# 第一次轮换成功
curl -s -X POST $BASE/auth/refresh -H 'Content-Type: application/json' \
  -d "{\"refresh_token\":\"$R1\"}"

# 60 秒内重复使用 R1：幂等，返回同一对 token
curl -s -X POST $BASE/auth/refresh -H 'Content-Type: application/json' \
  -d "{\"refresh_token\":\"$R1\"}"

# 等 60 秒后再用 R1：401，家族被吊销
sleep 61
curl -i -X POST $BASE/auth/refresh -H 'Content-Type: application/json' \
  -d "{\"refresh_token\":\"$R1\"}"
```

相关代码：

- `backend/internal/pkg/jwt/jwt.go`：`jti` / `fid` / `gen` 声明与签发
- `backend/internal/pkg/session/session.go`：家族状态、Lua 原子轮换、宽限缓存
- `backend/internal/service/auth_service.go`：登录、刷新、登出、重放处理
- `backend/internal/middleware/auth.go`：家族吊销检查
