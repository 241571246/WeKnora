# Vone-weknora 本地 Docker 安装实施方案

> 状态：方案已设计，暂不部署。
> 基线：WeKnora `v0.7.2`，提交 `3d5d8bfcdfee`，本地分支 `custom/v0.7.2`。
> 项目目录：`F:\Docker\weknora`。
> 产品名：`Vone-weknora`。

## 1. 目标与实施边界

本方案用于在本机 Docker Desktop 中部署 Vone-weknora，并保持腾讯 WeKnora 核心能力可持续升级。

本阶段只完成仓库治理、配置设计和部署准入设计，不执行以下动作：

- 不拉取或构建镜像；
- 不创建或启动容器；
- 不执行数据库迁移；
- 不写入真实外部 API Key；
- 不实施 UI 改造。

## 2. Git 仓库与升级基线

远程仓库采用双远程：

| 名称 | 地址 | 用途 |
| --- | --- | --- |
| `origin` | `https://github.com/241571246/WeKnora.git` | Vone 定制代码的推送目标 |
| `upstream` | `https://github.com/Tencent/WeKnora.git` | 获取腾讯官方版本和安全修复 |

分支规范：

- `main`：跟踪 Vone 可发布主线；
- `upstream/main`：只读上游主线；
- `custom/v0.7.2`：当前定制基线；
- `feature/vone-*`：单项定制功能；
- `upgrade/v0.7.2-to-vX.Y.Z`：上游升级验证分支。

禁止直接在 `upstream` 对应分支开发，也不建议在 `main` 上直接试改。

建议升级流程：

1. `git fetch upstream --tags --prune`；
2. 从当前 Vone 发布分支创建 `upgrade/*`；
3. 合并目标上游 tag，而不是长期追随易变的 `latest`；
4. 先解决配置、接口和数据库迁移冲突，再处理前端冲突；
5. 执行构建、迁移演练、知识库检索、对话、文件解析和回滚验证；
6. 验证通过后再合并到 Vone 主线。

## 3. Docker 拓扑

默认最小完整部署为 5 个常驻容器：

| 容器 | 职责 | 镜像/构建位置 | 对宿主机暴露 |
| --- | --- | --- | --- |
| `WeKnora-frontend` | Vue 前端与 Nginx 反向代理 | `wechatopenai/weknora-ui`；构建上下文 `frontend/` | 默认 `80` |
| `WeKnora-app` | Go API、业务逻辑、迁移与模型调用 | `wechatopenai/weknora-app`；`docker/Dockerfile.app` | 默认 `8080` |
| `WeKnora-docreader` | 文档解析 gRPC 服务 | `wechatopenai/weknora-docreader`；`docker/Dockerfile.docreader` | 不对宿主机开放 |
| `WeKnora-postgres` | 主数据库、pgvector/ParadeDB 检索 | `paradedb/paradedb:v0.22.2-pg17` | 不对宿主机开放 |
| `WeKnora-redis` | 流式响应和异步任务队列 | `redis:7.0-alpine` | 不对宿主机开放 |

`sandbox`、SearXNG、MinIO、Neo4j、Qdrant、Milvus、Weaviate、Doris、Dex、Langfuse 和 MCP 均为可选 profile，本次首期不启用。

多个容器会增加少量进程与内存开销，但不会复制一套操作系统内核。换来的好处是组件独立升级、健康检查、故障隔离和数据持久化。对当前本机 16 CPU、约 31 GiB 内存的 Docker 环境，先运行上述 5 容器是合理的；正式资源上限要在首次导入典型文档后以监控数据校准。

## 4. 镜像、源码和数据分别存放在哪里

### 4.1 源码与配置

以下内容位于项目目录 `F:\Docker\weknora`：

- Git 源码；
- `docker-compose.yml`；
- Vone Compose 覆盖文件 `deploy/docker-compose.vone.yml`；
- 根目录 `.env`（本地私密文件，Git 已忽略）；
- `config/builtin_models.vone.yaml`（只引用环境变量，不保存明文 Key）；
- 部署文档、升级记录和备份清单。

### 4.2 Docker 镜像

镜像层由 Docker Desktop 的内部存储管理，不会以普通文件形式保存到 `F:\Docker\weknora`。可通过以下命令查看：

```powershell
docker image ls
docker system df
```

首期推荐按已检出的 `v0.7.2` 源码本地构建，并在 Vone Compose 覆盖文件中使用独立标签：

```text
vone/weknora-ui:0.7.2-vone.1
vone/weknora-app:0.7.2-vone.1
vone/weknora-docreader:0.7.2-vone.1
vone/weknora-sandbox:0.7.2-vone.1   # 仅需要 Skills 沙箱时构建
```

这样可以避免 `latest` 漂移，也不会把本地定制镜像伪装成腾讯官方镜像。实际构建命令将在部署批准后由 `docker-compose.vone.yml` 统一生成。

### 4.3 持久化数据

首期推荐使用 Docker 命名卷：

- `postgres-data`：PostgreSQL 数据；
- `data-files`：上传文件；
- `redis-data`：Vone 覆盖文件新增，挂载到 `/data`，使 Redis AOF 在容器重建后仍保留；
- `docreader-tmp`：文档解析临时数据，可重建。

数据库直接绑定 Windows 目录可能遇到权限、文件锁和 I/O 性能问题，因此不作为默认方案。为了让可恢复资产仍归档在项目下，后续建立：

```text
F:\Docker\weknora\runtime\backups\postgres
F:\Docker\weknora\runtime\backups\config
F:\Docker\weknora\runtime\logs
```

`runtime/` 必须加入 `.gitignore`，备份通过 `pg_dump` 产生，而不是复制运行中的 PostgreSQL 数据目录。

## 5. 数据库与 Redis 决策

### 5.1 主数据库

采用 PostgreSQL/ParadeDB，不采用 MySQL。

虽然 `.env.example` 的注释仍列出 MySQL，但 v0.7.2 实际 `initDatabase` 只接受：

- `DB_DRIVER=postgres`；
- `DB_DRIVER=sqlite`。

传入 `mysql` 会返回 `unsupported database driver`。项目里的 MySQL 驱动用于 Doris 的 MySQL 协议，不代表 WeKnora 主库支持 MySQL。

完整本地部署使用：

```dotenv
DB_DRIVER=postgres
DB_HOST=postgres
DB_PORT=5432
DB_USER=weknora
DB_PASSWORD=<部署时生成的随机强密码>
DB_NAME=weknora
```

SQLite 仅作为轻量或桌面模式备选，不用于本方案。

### 5.2 Redis

Redis 使用独立容器：

```dotenv
STREAM_MANAGER_TYPE=redis
REDIS_ADDR=redis:6379
REDIS_PASSWORD=<部署时生成的随机强密码>
REDIS_DB=0
REDIS_PREFIX=stream:
```

数据库和 Redis 不合并进 App 容器。容器只加入内部 `WeKnora-network`，首期不映射 PostgreSQL `5432` 和 Redis `6379` 到宿主机。

## 6. 外部模型配置标准

### 6.1 安全原则

- 真实 Key 只写在根目录 `.env`；
- `.env` 不提交 Git，不复制到需求文档和日志；
- `config/builtin_models.vone.yaml` 只保存 `${ENV_VAR}` 引用；
- 模型配置以稳定 ID 管理，避免升级或改名后重复建模；
- `SYSTEM_AES_KEY` 必须使用独立的 32 字节值并安全备份，否则数据库内加密凭据可能无法恢复。

已确认的模型环境变量：

```dotenv
VONE_LLM_MODEL_NAME=deepseek-v4-flash
VONE_LLM_BASE_URL=https://api.deepseek.com/v1
VONE_LLM_API_KEY=<仅写本地>
VONE_LLM_PROVIDER=deepseek

VONE_EMBEDDING_MODEL_NAME=BAAI/bge-m3
VONE_EMBEDDING_BASE_URL=https://api.siliconflow.cn/v1
VONE_EMBEDDING_API_KEY=<仅写本地>
VONE_EMBEDDING_PROVIDER=siliconflow

VONE_RERANK_MODEL_NAME=BAAI/bge-reranker-v2-m3
VONE_RERANK_BASE_URL=https://api.siliconflow.cn/v1
VONE_RERANK_API_KEY=<仅写本地，可与 Embedding 共用>
VONE_RERANK_PROVIDER=siliconflow
```

DeepSeek Chat 使用一套 Key；SiliconFlow Embedding 和 Rerank 可以共用一套 Key，但仍保留按用途拆分的变量，便于后续独立轮换和统计。

`config/builtin_models.vone.yaml` 已按以下结构激活：

```yaml
builtin_models:
  - id: vone-default-chat
    tenant_id: 10000
    name: ${VONE_LLM_MODEL_NAME}
    type: KnowledgeQA
    source: remote
    is_default: true
    status: active
    parameters:
      base_url: ${VONE_LLM_BASE_URL}
      api_key: ${VONE_LLM_API_KEY}
      provider: ${VONE_LLM_PROVIDER}

  - id: vone-default-embedding
    tenant_id: 10000
    name: ${VONE_EMBEDDING_MODEL_NAME}
    type: Embedding
    source: remote
    is_default: true
    status: active
    parameters:
      base_url: ${VONE_EMBEDDING_BASE_URL}
      api_key: ${VONE_EMBEDDING_API_KEY}
      provider: ${VONE_EMBEDDING_PROVIDER}
      embedding_parameters:
        dimension: 1024
        truncate_prompt_tokens: 0

  - id: vone-default-rerank
    tenant_id: 10000
    name: ${VONE_RERANK_MODEL_NAME}
    type: Rerank
    source: remote
    is_default: true
    status: active
    parameters:
      base_url: ${VONE_RERANK_BASE_URL}
      api_key: ${VONE_RERANK_API_KEY}
      provider: ${VONE_RERANK_PROVIDER}
```

Compose 覆盖文件通过以下只读挂载加载模型声明：

```yaml
services:
  app:
    volumes:
      - ./config/builtin_models.vone.yaml:/app/config/builtin_models.yaml:ro
```

### 6.2 已确认的模型组合

| 用途 | Provider | 模型 | Base URL | 维度 |
| --- | --- | --- | --- | --- |
| Chat | DeepSeek | `deepseek-v4-flash` | `https://api.deepseek.com/v1` | - |
| Embedding | SiliconFlow | `BAAI/bge-m3` | `https://api.siliconflow.cn/v1` | `1024` |
| Rerank | SiliconFlow | `BAAI/bge-reranker-v2-m3` | `https://api.siliconflow.cn/v1` | - |

真实 Key 无需写入 Git 文档。可以在部署阶段由用户直接填入本机 `.env`，再用只显示“是否存在和长度”、不回显内容的检查方式验证。

## 7. 其余必要配置

部署时必须生成或确认：

| 配置 | 决策 |
| --- | --- |
| `WEKNORA_VERSION` | 不使用长期漂移的 `latest`；Vone 镜像固定 `0.7.2-vone.1` |
| `COMPOSE_PROJECT_NAME` | `vone-weknora` |
| `TZ` | `Asia/Shanghai` |
| `GIN_MODE` | `release` |
| `AUTO_MIGRATE` | 首次部署 `true`，升级前先备份并演练 |
| `FRONTEND_PORT` | 默认 `80`，实际部署前复查端口 |
| `APP_PORT` | 默认 `8080`，仅调试或 API 直连需要；公网应经前端/Nginx |
| `JWT_SECRET` | 随机强密钥，不使用示例值 |
| `SYSTEM_AES_KEY` | 严格 32 字节，单独备份，不使用示例值 |
| `DISABLE_REGISTRATION` | 首次管理员注册后建议设 `true` |
| `LANGFUSE_ENABLED` | 首期 `false` |
| Langfuse 示例 Key | `.env` 中保持空值，不使用 `.env.example` 的占位字符串 |
| PostgreSQL/Redis 端口 | 不映射到宿主机 |

首期不启用 OIDC、MinIO、图数据库、外部向量库、SearXNG、Langfuse 和 MCP，待实际需求出现后逐项开启，避免一次性增加资源和排障面。

## 8. 后续实施步骤（收到部署授权后执行）

### 阶段 A：配置落地

1. 备份当前 Git 状态并确认工作树差异；
2. 复核不含秘密的 `deploy/docker-compose.vone.yml`；
3. 激活 `config/builtin_models.vone.yaml` 中已确认的模型；
4. 从 `deploy/vone.env.example` 生成本地 `.env`，替换所有空白必填项；
5. 增加 `runtime/` 忽略规则和备份目录；
6. 运行 `deploy/validate-vone-config.ps1`，通过后才允许构建或启动。

模板结构可随时校验，不需要根 `.env`，也不会启动容器：

```powershell
powershell -NoProfile -ExecutionPolicy Bypass `
  -File deploy/validate-vone-config.ps1 `
  -Template `
  -EnvFile deploy/vone.env.example `
  -ModelConfig config/builtin_models.vone.yaml
```

未来填好根 `.env` 并激活模型 YAML 后，运行真实部署准入校验：

```powershell
powershell -NoProfile -ExecutionPolicy Bypass `
  -File deploy/validate-vone-config.ps1
```

预检只输出缺失的变量名或规则，不输出变量值。任何错误均返回非零退出码，并阻止后续构建/启动步骤。

### 阶段 B：构建

使用上游 `v0.7.2` 源码和 Vone 覆盖文件构建固定标签。构建物进入 Docker Desktop 镜像存储，不进入 Git：

```powershell
$env:VITE_IS_DOCKER = 'true'
$env:VITE_FRONTEND_COMMIT = git rev-parse --short HEAD
Set-Location frontend
npm ci
npm run type-check
npm run build
Set-Location ..

docker compose --env-file .env `
  -f docker-compose.yml `
  -f deploy/docker-compose.vone.yml `
  build frontend app docreader
```

上游 `frontend/Dockerfile` 只封装现成的 `frontend/dist`，因此必须先执行前端生产构建。Vone 覆盖文件向 App 传入带故障切换的 `GOPROXY` 和 HTTPS Debian 镜像源；本地 `docker/Dockerfile.app` 对 APT 增加 5 次重试，并允许镜像参数包含协议。升级上游时需单独复核这项构建基础设施补丁。

如启用 Skills 沙箱，再单独构建 `sandbox`。

### 阶段 C：首次启动

1. 先启动 PostgreSQL、Redis 和 Docreader；
2. 健康后启动 App，观察数据库迁移和内置模型加载日志；
3. App 健康后启动 Frontend；
4. 不启动任何未批准 profile。

### 阶段 D：验证

- 5 个核心容器状态正常；
- `/health` 正常；
- 首次注册、登录和空间创建正常；
- 内置 Chat、Embedding（以及可选 Rerank）模型可见；
- 模型连通性测试通过，日志不回显 Key；
- 上传典型 PDF/Office 文档，解析、分块、Embedding 和检索成功；
- 知识库对话能够返回引用；
- PostgreSQL 备份可生成；
- 重建 App/Redis 容器后数据仍存在；
- 停止与回滚步骤经过验证。

## 9. 回滚与备份

每次升级前必须保存：

- 当前 Git commit 和 Vone 镜像标签；
- `docker compose config` 的脱敏版本；
- PostgreSQL `pg_dump`；
- `.env` 的安全离线备份；
- `SYSTEM_AES_KEY` 的安全离线备份；
- `config/builtin_models.vone.yaml`；
- 上传文件卷的备份或校验清单。

回滚时恢复上一版镜像标签和配置，再恢复与该版本匹配的数据库备份。不能只回滚应用镜像而保留未经验证的新数据库结构。

## 10. 当前准入状态

| 项目 | 状态 |
| --- | --- |
| Git 双远程 | 已完成 |
| v0.7.2 本地基线 | 已完成 |
| Docker 环境只读检查 | 已完成 |
| 外部模型配置机制 | 已确认 |
| 外部模型具体参数 | 已确认并写入非秘密模板 |
| Vone Compose 覆盖文件 | 已建立并投入运行 |
| Vone 环境变量模板 | 已建立，不含秘密 |
| 内置模型 YAML | 已激活并注册 Chat、Embedding、Rerank；数据库记录由 YAML 管理 |
| 不回显秘密的部署前预检 | 已建立；真实配置校验通过 |
| DeepSeek V4 静态兼容 | 已确认 Provider、OpenAI Chat 路径及 reasoning_content 回传代码 |
| Windows 原生 Go 针对性测试 | 受上游 `pg_query_go` CGO 依赖阻断；当前缺少 GCC，不作为通过证据 |
| Docker Linux 构建 | UI、App、DocReader 三个 Vone 镜像已构建成功 |
| 外部模型只读连通性 | DeepSeek Key 和模型目录通过；SiliconFlow 因账户未完成实名认证返回 403 |
| 镜像构建/拉取 | 已完成；PostgreSQL ParadeDB 与 Redis 镜像已拉取 |
| 容器启动和数据迁移 | 5 个核心容器运行；迁移版本 79、dirty=false；重启恢复通过 |
| 数据备份 | 已生成 `runtime/backups/postgres/weknora-initial-20260808-111421.dump` |
| 端到端业务冒烟 | 待完成 SiliconFlow 实名认证后执行注册、文档解析、Embedding、检索和知识库对话 |
| UI 定制 | 已另立未来需求，未实施 |
