# Vone Decisions

## Task ID

- TASK-20260817-FE-ROCKL-01

## Last Updated

- 2026-08-21 14:41

## Current Decisions

### D1

#### Decision

- 将 `/platform/knowledge-bases` 顶部横向项目结构调整为左侧可折叠树形导航，继续使用现有项目树接口和 ACL。

#### Reason

- 用户确认树形目录比横向节点更适合多级项目浏览，并要求基于已讨论方案直接优化调整。

#### Impact

- 仅前端布局、树展开状态和节点操作入口变化；不修改后端、数据库、迁移、权限或知识库内部 `folder_path`。

#### Follow-up

- U2 完成后执行定向测试、类型检查、i18n、生产构建和登录态浏览器验收；技术通过不自动表示 QA、业务验收或发布授权通过。

### D2

#### Decision

- 浏览器技术验证使用生产构建预览和 Playwright 会话内只读模拟数据；不使用未授权的现网账号或写接口。

#### Reason

- 隔离浏览器无现网登录态，但仍需验证真实生产组件的布局、筛选和折叠行为。

#### Impact

- Task Technical 可记录页面级技术证据；真实账号、真实后端、Owner/非 Owner 矩阵仍属于独立 QA，不因模拟预览而通过。

#### Follow-up

- `TASK-20260817-QA-ROCKL-01` 使用受控真实账号执行运行时验证并单独记录结果。
