# VONE 跨系统导航与 WeKnora 独立钉钉登录

## 1. 设计边界

- PMS 与 WeKnora 只通过普通 `http/https` 外链切换，不携带 Token、Cookie、OAuth code、state、用户 ID 或权限参数。
- PMS 与 WeKnora 使用同一企业下两个独立钉钉应用，各自维护 Client ID、Client Secret 和回调地址。
- WeKnora 钉钉登录成功后只签发 WeKnora 自己的 access/refresh token，并继续使用 WeKnora 本地租户与 RBAC。
- 不建设 `/sso-entry`，不共享 Cookie，不做跨系统静默登录。

## 2. 配置

PMS 前端构建配置：

```env
VITE_AI_BRAIN_URL=https://brain.example.com/
```

WeKnora Docker 运行时配置：

```env
PMS_URL=https://pms.example.com/

WEKNORA_DINGTALK_AUTH_ENABLE=true
WEKNORA_DINGTALK_AUTH_CORP_ID=dingxxxxxxxx
WEKNORA_DINGTALK_AUTH_CLIENT_ID=dingxxxxxxxx
WEKNORA_DINGTALK_AUTH_CLIENT_SECRET=<仅写入部署环境>
WEKNORA_DINGTALK_AUTH_REDIRECT_URI=https://brain.example.com/api/v1/auth/dingtalk/callback
WEKNORA_DINGTALK_AUTH_TIMEOUT=30s
WEKNORA_DINGTALK_AUTH_JIT_PROVISIONING=false
WEKNORA_DINGTALK_AUTH_EMAIL_AUTO_BIND=false
WEKNORA_DINGTALK_AUTH_AUTO_JOIN_TENANT=false
WEKNORA_DINGTALK_AUTH_DEFAULT_TENANT_ID=
WEKNORA_DINGTALK_AUTH_DEFAULT_TENANT_ROLE=viewer
```

钉钉开放平台登记的回调必须与 `WEKNORA_DINGTALK_AUTH_REDIRECT_URI` 完全一致。Secret 不进入前端镜像、运行时 `config.js`、Git 或浏览器日志。

## 3. 首次账号处理

### 钉钉工作台自动入口

前端支持可选入口 `/login?autologindingtalk=true`，例如：

```text
https://co.vonechina.com:13443/login?autologindingtalk=true
```

只接受精确的 `true`。未登录时等待钉钉配置就绪后，自动调用现有登录流程：有钉钉桥接时尝试企业内免登，否则跳转现有浏览器 OAuth。参数不携带凭据，也不能绕过钉钉要求的授权确认。

参数在发起前从当前 URL 消费，失败后停留登录页，刷新不会再次自动跳转；可手动点钉钉登录重试，或重新打开工作台入口发起新尝试。不带参数的 `/login` 保持手动登录。授权回调和邀请/注册流程优先，不触发竞争的自动登录。

该参数需发布支持它的前端版本后才生效；不需要后端或数据库变更。真实钉钉客户端登录与平台授权仍需用真实账号验收。

默认关闭 JIT 和邮箱自动绑定。未建立映射的钉钉身份会保存为 `pending`，但不会登录到任何本地账号。

- 已有 WeKnora 账号：先审核身份，再把 `external_auth_identities.user_id` 绑定到目标本地用户。
- 全新用户且允许自动开户：显式开启 `WEKNORA_DINGTALK_AUTH_JIT_PROVISIONING=true`。自动创建的账号是普通用户，不是系统管理员，也没有跨租户权限。
- 如需把JIT新用户统一加入已有空间，同时开启 `WEKNORA_DINGTALK_AUTH_AUTO_JOIN_TENANT=true`，填写目标空间ID和默认角色。该模式不会创建个人空间，并把目标空间设为用户登录默认空间；建议使用最小权限 `viewer`。
- 仅在企业邮箱完整、唯一且已核验时，才考虑显式开启 `WEKNORA_DINGTALK_AUTH_EMAIL_AUTO_BIND=true`。

## 4. 发布顺序与回滚

建议顺序：

1. 备份数据库并演练 migration。
2. 上游迁移完成后运行 VONE 独立迁移轨道中的 `000002_external_auth_identities`（PostgreSQL / SQLite 各一份）；不要占用上游迁移编号。
3. 发布 WeKnora 后端，再发布 WeKnora 前端。
4. 配置 WeKnora 独立钉钉应用并用测试账号验证浏览器授权、钉钉内免登和本地权限。
5. PMS 前端独立发布并验证外链。

应用级回滚时先将 `WEKNORA_DINGTALK_AUTH_ENABLE=false`，并清空两个系统的外链 URL。映射表是加性表，可以保留；删除表属于单独的数据变更，不随应用回滚自动执行。

## 5. 上游升级检查

该实现没有修改 WeKnora 既有 Token 格式、密码登录、OIDC 表或权限模型，但以下上游接入点可能产生合并冲突：

- 后端：`internal/config/config.go`、`internal/container/container.go`、`internal/middleware/auth.go`、`internal/router/router.go`。
- 前端：`frontend/src/App.vue`、`frontend/src/api/auth/index.ts`、`frontend/src/views/auth/Login.vue`、`frontend/src/components/menu.vue`、`frontend/docker-entrypoint.sh`。
- 独立新增的 `dingtalk_*` 文件、`externalSystemUrl.*` 和 migration 通常可以直接保留。

每次升级从上游新 tag 建立 `upgrade/*` 分支，先完成上游原版构建，再重放上述接入点；随后执行钉钉定向测试、前端类型检查/i18n/build、两个外链检查和真实测试应用联调。不要直接覆盖当前 VONE 发布分支。
