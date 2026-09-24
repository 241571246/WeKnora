# WeKnora v0.8.2-vone.1

本版本基于腾讯 WeKnora 官方 `v0.8.2` 标签，发布分支为 `custom/v0.8.2`。这是 VONE 定制源码版本；不代表已经部署到本地或生产 Docker。

## 本次合并

- 合入官方 v0.8.2 的前后端、数据库迁移与文档更新。
- 保留 VONE 钉钉跨系统登录、VONE 品牌与系统切换入口。
- 保留知识库成员权限、集合目录、文件夹及上传交互等定制，并适配官方新组件与接口。
- 将外部身份表迁移放入独立的 `migrations/vone/*/000002_external_auth_identities`，不再占用官方主迁移编号。先运行官方迁移，再运行 VONE 独立迁移；升级前应备份并在隔离环境演练。
- 补齐各语言包所需词条，并使源码文本断言与 POSIX 专用测试适配 Windows 开发环境。

## 验证记录

- 前端 `npm run type-check`、`npm run build`、`npm run check-i18n` 通过。
- 前端全量测试：1,244 项中 1,241 通过，0 失败，3 项按环境条件跳过。
- Linux/WSL Go：钉钉定向测试，以及 `internal/database`、`internal/handler`、`internal/router` 全包测试通过。
- 未执行 Docker 镜像构建、数据库实机升级、钉钉真实账号联调或部署验收；这些是后续部署前的独立门槛。

## 升级提醒

请先备份数据库和对象存储，演练官方及 VONE 双轨迁移，并按 `docs/customization/VONE_CROSS_SYSTEM_DINGTALK_AUTH.md` 检查钉钉配置。当前发布仅提供源码与版本标签，不自动修改运行中的容器或数据。
