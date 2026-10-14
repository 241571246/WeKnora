# WeKnora v0.8.3-vone.1

本次源码发布基于官方 `v0.8.3`（`3e22822c1d2397be730c458c7539d0eb8c113a24`），发布分支为 `custom/v0.8.3`。合并提交为 `4d460ab9`，另包含后续构建与权限兼容修正。服务器部署结果以发布后的独立回读记录为准。

## 定制兼容性结论

保留 VONE 品牌、跨系统切换、钉钉 OAuth/企业内桥接登录、`autologindingtalk=true` 可选入口、知识库细分 ACL、项目树及独立 VONE 数据库迁移。原 `F:\Docker\weknora` 工作区及未提交文件保持原样；升级在独立工作树完成。证书及私钥目录没有纳入源码。

合并时有 11 个文件存在文本冲突：`Login.vue`、`KnowledgeBase.vue`、`menu.vue`、5 个既有语言包、2 个样式/终端测试文件、`internal/router/router.go`。全部逐项处理，未采用整文件覆盖定制代码的方式。

| 范围 | 处理与验证 |
| --- | --- |
| 登录与注册 | 保留钉钉配置加载、自动登录防重入、回调/邀请优先规则；采用新版 `registrationPolicy`。VONE 注册组件同样遵守新策略，不再凭邀请对象绕过禁止注册。 |
| 新画廊 | 接口同时要求 `kb.documents.list` 与 `kb.document.preview`；前端页签与加载同步检查权限，防止仅 AI 权限或仅文件列表权限看到原图与 OCR。 |
| 跨库 Wiki 搜索 | 对请求中的每个去重知识库检查 `kb.ai.query`，与既有单库搜索一致；保留原请求正文供上游处理器解析。拒绝第二个库时，不执行部分搜索。 |
| 路由与侧栏 | 同时保留 VONE ACL/项目树路由和官方图片属性路由；保留系统切换与新版资源加载。 |
| 多语言 | 按词条进行三方比较，5 种既有语言各有 89–90 项 VONE 定制，未发现同词条语义冲突；保留全部官方新词条。新增繁体中文的 VONE 专用词条使用现有英文回退。 |
| 测试 | 保留新版样式规则，并按合并后的既有 VONE 样式计数冻结基线；保留 Windows 不适用的 POSIX 测试跳过条件。 |
| 镜像构建 | 统一 Dockerfile 各阶段的 Debian 镜像源 URL 格式；DuckDB 预下载步骤复用 Go 模块缓存，避免重复下载及 IPv6 网络错误。 |

## 已执行的技术验证

- 前端 `vue-tsc --build`、VONE 生产构建通过。
- 前端全量 1,570 项：1,567 通过、0 失败、3 项环境跳过；i18n 14 项通过。
- 构建页面的 Edge Chromium 模拟验证：9 个钉钉/回调/重试/已有会话场景，4 个邀请与注册策略场景通过；覆盖 1440×900 和 390×844。浏览器插件未提供，使用现有 Playwright。API 为模拟，不是实际钉钉账号验收。
- Linux Go 的 config、types、database、middleware、router、handler、session 关键包，以及钉钉/ACL/目录定向测试通过；服务器入口编译通过。
- 新增真实 HTTP 路由拒绝测试：画廊缺少列表或预览权限返回 403；跨库 Wiki 搜索检查第二个库并保留 JSON 正文。
- 真实数据库备份在隔离 ParadeDB 实例恢复，然后在同一隔离卷上完成 0.22.6 → 0.25.11 升级；主迁移 110 → 118，VONE 迁移保持 2，均无 dirty 标记。
- 演练中 4 个租户、16 个知识库、377 个文档、55,407 个切片数量一致；画廊回填 585 张图片。关键词与向量查询可执行，无无效索引。
- 与部署前容器比较，17 个敏感运行配置保持一致，包括 `SYSTEM_AES_KEY`、`JWT_SECRET` 与签名密钥。实际注册策略为 `self_serve`，保持原值。

这是 Codex 的技术自验。独立人工 QA、真实钉钉账号/设备、模型服务回答质量及业务验收仍需分别完成。

## 升级顺序与回滚边界

1. 完成数据库、角色、文件和配置备份，验证恢复；记录原镜像、卷、Compose 组合及密钥一致性。
2. 使用固定 `0.8.3-vone.1` 镜像构建并完成隔离演练。生产切换前停止写入，补充一致性备份及 Redis/Neo4j 快照。
3. 保留原 PostgreSQL 17 数据卷，仅切换到 `paradedb/paradedb:v0.25.11-pg17`。
4. 在每个已安装扩展的可连接数据库中执行 `ALTER EXTENSION pg_search UPDATE` / `ALTER EXTENSION vector UPDATE`，回读 0.25.11 / 0.8.4 后再运行应用迁移。不要执行旧的升级到 0.22.6 的 SQL。
5. 使用官方 Compose + VONE overlay + 服务器 overlay，只重建 postgres/app/docreader/frontend；保留已有 Redis/Neo4j 卷与容器。
6. 回读主迁移 118、VONE 迁移 2、数据量、容器健康、版本、登录/授权入口和检索能力。

扩展升级后，旧镜像不能重放新版 pg_search WAL；回滚必须将升级前备份恢复到独立卷，并配套旧镜像及旧配置，不能只改镜像标签。保留升级前备份直至业务验收完成。

## 官方升级须知的适用情况

| 项目 | 本次处理 |
| --- | --- |
| PostgreSQL 111–118 / SQLite 31–36 | 当前部署使用 PostgreSQL；真实备份演练通过。SQLite 迁移随源码保留并通过相应 Go 测试。 |
| embeddings 索引 | 约 5.5 万切片的真实备份演练中，000112 正常完成；更大部署可提前创建同名并发索引。 |
| 注册模式 | 当前 `self_serve` 无需转换；`invite_only` 禁止邀请注册，需要邀请自注册时应明确选择 `invite_register`。 |
| 空工具列表 | 由 000116 将旧智能推理 Agent 默认工具显式写入；新的 API 若要默认工具需省略字段或传 null。仅 Wiki 知识库的旧 Agent 仍需业务回归。 |
| 图片向量 | 不自动更换模型或启用图片向量。开启后须重新解析目标文档；更换到 messages 编码的多模态模型时需重建旧向量。 |
| 原文定位 | 旧文档可文字回退；精确页码/段落定位需重新解析，不对所有文档自动重解析。 |
| MCP OAuth 时间戳 | 由 000118 按 UTC 转带时区类型；历史非 UTC 偏移若影响授权，需刷新或重新授权。 |
| 附件保留 | 有有效会话引用的附件保留更久，应关注文件存储增长。 |
| Docker 终端 | 断开结束当前 Shell，重连建立新 Shell；长任务应使用合适的进程管理方式。 |

官方说明：[v0.8.3 升级须知](https://weknora.weixin.qq.com/docs/07-releases/v0.8.3#upgrade-notes)、[ParadeDB 存量库升级](https://weknora.weixin.qq.com/docs/01-getting-started/06-paradedb-upgrade)。
