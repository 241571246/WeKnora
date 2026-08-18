# VONE-0.7.2.1 权限与树形结构首次发布指南

## 1. 文档目的

本文供运维人员首次发布 VONE 知识库权限与树形结构功能时执行。适用范围：

- 知识库 Owner、共同 Owner、编辑、文档查看、AI 使用和自定义能力权限；
- 员工仅查看获授权知识库；
- 文档、切片和知识库删除能力分离；
- 跨工作空间 `kb_shares` 保留原有共享语义并统一接入 Authorizer；
- 工作空间内多级知识库集合树及单库单节点绑定；
- 树形节点筛选知识库列表。

本流程默认使用 PostgreSQL 和 Docker Compose。SQLite 仅用于开发或测试，不属于本生产流程。

## 2. 发布结论边界

- 本文是执行手册，不等于生产发布授权。
- 只有 `release-manifest.md` 的 Release Ready 门禁和人工发布授权通过后，才允许执行生产切换。
- 首次部署必须从 `shadow` 开始，完成迁移、只读对账和权限矩阵后才能切换到 `enforce`。
- 不允许使用 `off` 作为正式发布或权限安全回滚方案。

## 3. 发布输入

运维必须从发布负责人取得以下不可变信息：

| 项目 | 要求 |
|---|---|
| Release ID | `VONE-0.7.2.1` |
| Git 分支 | `custom/v0.7.2` |
| Release Commit | 必须填写完整 40 位 Commit SHA |
| 后端镜像 Tag | 建议 `0.7.2.1.<Commit短SHA>` |
| 前端镜像 Tag | 建议 `0.7.2.1.<Commit短SHA>` |
| 上一版镜像 | 后端和前端 Tag 均须记录 |
| 数据库备份 | 备份文件、恢复编号和 SHA-256 |
| 发布审批 | 发布负责人、审批时间和证据 |

禁止使用未提交工作区、浮动 `latest` Tag 或清单外临时 cherry-pick 构建生产镜像。

## 4. 变更资产

### 4.1 数据库

- PostgreSQL Up：`migrations/vone/versioned/000001_kb_acl_and_collections.up.sql`
- PostgreSQL Down：`migrations/vone/versioned/000001_kb_acl_and_collections.down.sql`
- 独立迁移版本表：`vone_schema_migrations`
- 迁移后核对：`docs/vone/releases/VONE-0.7.2.1/reconciliation-postgresql.sql`

正常发布只执行 Up。Down 会删除 VONE 权限和集合树数据，不属于普通应用回滚步骤。

### 4.2 配置

首次发布必须设置：

```dotenv
AUTO_MIGRATE=true
AUTO_RECOVER_DIRTY=true
WEKNORA_VONE_KB_ACL_MODE=shadow
```

`.env` 包含密码和模型 Key，不得提交 Git、粘贴到工单或写入发布日志。

### 4.3 应用

- 后端必须为包含 VONE Authorizer 和独立迁移 Runner 的 ACL-aware 镜像。
- 前端必须与同一 Release Commit 构建。
- DocReader 无业务逻辑变更时可复用已批准镜像，但首次完整构建建议生成同版本 Tag。

## 5. 发布前检查

以下命令假设在 Windows PowerShell 和仓库根目录执行。

### 5.1 锁定源码

```powershell
Set-Location F:\Docker\weknora

$ReleaseCommit = '<填写发布 Commit 的完整 SHA>'

git fetch origin
git switch custom/v0.7.2
git pull --ff-only

if ((git rev-parse HEAD).Trim() -ne $ReleaseCommit) {
    throw '当前 HEAD 与批准的 Release Commit 不一致'
}
if (git status --porcelain) {
    throw '工作区不干净，禁止构建发布镜像'
}
```

### 5.2 配置预检

从模板生成实际 `.env`，通过安全渠道填入密码和 Key：

```powershell
Copy-Item deploy\vone.env.example .env
```

确认首次发布为 `WEKNORA_VONE_KB_ACL_MODE=shadow`，然后执行：

```powershell
powershell -NoProfile -ExecutionPolicy Bypass `
  -File deploy\validate-vone-config.ps1

powershell -NoProfile -ExecutionPolicy Bypass `
  -File docs\vone\releases\VONE-0.7.2.1\validate-release-package.ps1
```

任一命令返回非零退出码时停止发布。

### 5.3 记录现状

```powershell
docker compose --env-file .env `
  -f docker-compose.yml `
  -f deploy/docker-compose.vone.yml `
  ps

docker inspect WeKnora-app --format '{{.Config.Image}}'
docker inspect WeKnora-frontend --format '{{.Config.Image}}'
```

同时记录当前 Git Commit、脱敏后的 Compose 配置、当前数据库迁移版本和 `kb_shares` 数量。

## 6. 数据备份与变更冻结

冻结知识库成员、权限、集合节点和知识库绑定写操作。备份 PostgreSQL：

```powershell
$BackupDir = 'runtime\backups\postgres'
New-Item -ItemType Directory -Force -Path $BackupDir | Out-Null
$BackupFile = Join-Path $BackupDir "weknora-pre-0721-$(Get-Date -Format 'yyyyMMdd-HHmmss').dump"

docker compose --env-file .env `
  -f docker-compose.yml `
  -f deploy/docker-compose.vone.yml `
  exec -T postgres sh -lc `
  'pg_dump -Fc -U "$POSTGRES_USER" -d "$POSTGRES_DB"' `
  > $BackupFile

Get-FileHash -Algorithm SHA256 $BackupFile
```

备份文件必须非空，并由运维验证可读取。生产环境应按组织制度补做恢复演练或使用平台级备份。

迁移前记录 `kb_shares`：

```powershell
docker compose --env-file .env `
  -f docker-compose.yml `
  -f deploy/docker-compose.vone.yml `
  exec -T postgres sh -lc `
  'psql -At -U "$POSTGRES_USER" -d "$POSTGRES_DB" -c "SELECT COUNT(*) FROM kb_shares;"'
```

## 7. 构建不可变镜像

### 7.1 前端验证与构建

```powershell
$ShortCommit = git rev-parse --short HEAD
$env:VITE_IS_DOCKER = 'true'
$env:VITE_BRAND_PROFILE = 'vone'
$env:VITE_FRONTEND_COMMIT = $ShortCommit

Push-Location frontend
npm ci
npm run type-check
npm test
npm run check-i18n
npm run build-only
Pop-Location
```

### 7.2 设置镜像 Tag

将 `.env` 中两个 Tag 固定为批准值，例如：

```dotenv
VONE_IMAGE_TAG=0.7.2.1.<Commit短SHA>
VONE_UI_IMAGE_TAG=0.7.2.1.<Commit短SHA>
```

### 7.3 Docker 构建

```powershell
docker compose --env-file .env `
  -f docker-compose.yml `
  -f deploy/docker-compose.vone.yml `
  build frontend app docreader
```

如本版本使用 Skills Sandbox，再构建 `sandbox`。记录所有镜像 ID：

```powershell
docker image inspect "vone/weknora-app:$env:VONE_IMAGE_TAG" --format 'ID={{.Id}} Created={{.Created}}'
docker image inspect "vone/weknora-ui:$env:VONE_UI_IMAGE_TAG" --format 'ID={{.Id}} Created={{.Created}}'
```

如果 Tag 只写入 `.env` 而未写入当前 PowerShell 环境，请直接使用 `.env` 中的实际 Tag 查询。

## 8. 首次部署顺序

### 8.1 启动基础设施

```powershell
docker compose --env-file .env `
  -f docker-compose.yml `
  -f deploy/docker-compose.vone.yml `
  up -d postgres redis docreader
```

确认健康后再启动后端。

### 8.2 以 shadow 启动后端并自动迁移

```powershell
docker compose --env-file .env `
  -f docker-compose.yml `
  -f deploy/docker-compose.vone.yml `
  up -d --no-deps app

docker logs --tail 300 WeKnora-app
```

`AUTO_MIGRATE=true` 时，后端按顺序执行上游迁移和 VONE 独立迁移。不要再手工重复执行 Up SQL。

迁移失败时后端可能继续启动，但 VONE 权限会 fail closed。必须检查日志和迁移状态，不能只看容器 Running。

### 8.3 回读迁移版本

```powershell
docker compose --env-file .env `
  -f docker-compose.yml `
  -f deploy/docker-compose.vone.yml `
  exec -T postgres sh -lc `
  'psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -c "SELECT version, dirty FROM vone_schema_migrations;"'
```

期望：`version=1`、`dirty=false`。

### 8.4 执行只读对账

```powershell
Get-Content -Raw `
  docs\vone\releases\VONE-0.7.2.1\reconciliation-postgresql.sql |
docker compose --env-file .env `
  -f docker-compose.yml `
  -f deploy/docker-compose.vone.yml `
  exec -T postgres sh -lc `
  'psql -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d "$POSTGRES_DB"'
```

以下异常查询必须返回零行：

- 无有效 Owner 的知识库；
- 无效知识库或工作空间成员关系；
- 不在冻结 17 项能力集合内的 Capability；
- 跨工作空间或孤立的集合绑定。

`kb_shares` 数量必须与迁移前一致。

### 8.5 启动前端

```powershell
docker compose --env-file .env `
  -f docker-compose.yml `
  -f deploy/docker-compose.vone.yml `
  up -d --no-deps frontend
```

## 9. shadow 阶段验证

至少验证并留存：

1. Workspace Owner 创建知识库后自动成为 Owner；
2. Owner 添加共同 Owner；
3. Editor 可编辑但不能执行未授权删除；
4. 文档查看者可按配置查看文件，但不能编辑；
5. AI-only 用户可以检索并显示引用，但不能调用文件、全文和切片预览接口；
6. 文件删除、切片删除和知识库删除三项权限分别生效；
7. 员工只能看到获授权知识库；
8. `kb_shares` 仍保持跨工作空间共享，且不能绕过 Authorizer；
9. 集合树新增、移动、排序、非空删除保护和单库单节点绑定正确；
10. 点击树节点只显示该节点及后代绑定的知识库，点击全部恢复完整列表；
11. Agent 和 API Key 权限不超过调用者与配置知识库范围的交集；
12. `/api/v1/system/info` 回读 `kb_acl_mode=shadow`、`vone_db_version=1` 且 `vone_db_error` 为空。

发现越权、Owner 缺失、迁移异常或绑定异常时停止切换 `enforce`。

## 10. 切换 enforce

只有权限矩阵、只读对账和发布授权通过后，才将 `.env` 修改为：

```dotenv
WEKNORA_VONE_KB_ACL_MODE=enforce
```

必须重新创建 App 容器以加载环境变量；仅执行 `docker restart` 不可靠：

```powershell
docker compose --env-file .env `
  -f docker-compose.yml `
  -f deploy/docker-compose.vone.yml `
  up -d --no-deps --force-recreate app
```

重新执行健康检查、系统信息回读、Owner 正向用例和至少一个未授权 403 用例。

## 11. 发布后监控

- 观察窗口不少于 60 分钟，并覆盖下一次审计或后台任务周期；
- 监控 App 重启次数、HTTP 5xx、授权拒绝计数和 `[kb_acl_metric]`；
- 抽查 VONE 拒绝审计记录包含工作空间、知识库、主体和能力；
- 再次执行 `reconciliation-postgresql.sql`；
- 记录业务验收人和验收结论。

## 12. 回滚

### 12.1 触发条件

- 任何跨工作空间或直接 ID 越权；
- Owner 丢失；
- 非法 Capability 或孤立绑定；
- 迁移 dirty/error；
- Agent/API Key 范围扩大；
- 持续授权异常或不可接受的业务错误。

### 12.2 安全回滚步骤

1. 冻结知识库管理写操作；
2. 保持 `WEKNORA_VONE_KB_ACL_MODE=enforce`；
3. 回滚到上一套已验证的 ACL-aware VONE 后端和前端镜像；
4. 保留 VONE 新增表和数据；
5. 重新执行对账和拒绝用例；
6. 验证不会扩大知识库可见范围后恢复流量。

禁止直接回滚到不识别新 ACL 的原始 WeKnora 0.7.2 Standard 后端。

### 12.3 Down SQL 限制

普通回滚不得执行 `000001_kb_acl_and_collections.down.sql`。仅在以下条件全部满足时才允许：

- 人工批准数据丢失；
- 已导出 VONE ACL 和集合树数据；
- 所有 ACL 依赖服务已停止；
- 已隔离知识库 API 流量；
- 已确认恢复方案。

## 13. 运维结果记录

```text
Release ID:
Release Commit:
Backend Image / ID:
Frontend Image / ID:
Previous Backend Image:
Previous Frontend Image:
Backup File / Restore ID:
Backup SHA-256:
kb_shares Before / After:
VONE Migration Version / Dirty:
Initial ACL Mode:
Enforce Switched At:
Reconciliation Result:
Authorization Matrix Result:
Smoke Result:
Monitoring Result:
Release Approver:
Operator:
Started At / Finished At:
Rollback Required: Yes / No
Notes:
```
