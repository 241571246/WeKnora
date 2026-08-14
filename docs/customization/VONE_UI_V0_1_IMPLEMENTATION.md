# VONE知识库 UI v0.1 实施与发布说明

> Requirement：`CUST-VONE-UI-001`
> 上游基线：WeKnora `v0.7.2`
> UI 版本：`0.1.0`
> 前端镜像：`vone/weknora-ui:0.7.2-vone.2`

## 1. 本版本交付范围

- 统一品牌配置：`VONE知识库`、`Vone Knowledge`、VONE、`@vonechina 2026`；
- 蓝色/浅色 Logo 与 favicon；
- VONE 蓝色设计令牌、浅色和深色变量、TDesign 覆盖；
- 登录、注册、邀请注册和 OIDC 入口的 VONE 页面外壳；
- 登录页浅色、深色、跟随系统及语言切换；
- 全局侧栏 Logo、中文产品名；
- 浏览器标题、描述和页面启动底色；
- `Powered by Tencent WeKnora` 上游归属链接；
- `VITE_BRAND_PROFILE=upstream|vone` 构建期开关。

本版本不修改后端、数据库、Redis、模型配置、API 契约、认证 Store、路由或知识库核心业务。

## 2. 代码边界

主要定制代码集中在：

```text
frontend/src/customizations/vone/
├── assets/
├── components/
│   ├── BrandIdentity.vue
│   └── VoneLoginExperience.vue
├── brand.ts
├── index.ts
├── tokens.css
├── light.css
├── dark.css
└── tdesign-overrides.css
```

上游文件只保留以下接入点：

- `frontend/src/main.ts`：加载并应用 VONE 品牌包；
- `frontend/src/views/auth/Login.vue`：VONE/上游登录体验分流，认证处理仍使用原函数；
- `frontend/src/components/menu.vue`：替换侧栏品牌位；
- `frontend/index.html`：首屏标题、描述和无闪烁底色。

## 3. 本地构建

在仓库根目录执行：

```powershell
Set-Location F:\Docker\weknora\frontend
$env:VITE_BRAND_PROFILE='vone'
npm ci
npm run type-check
npm run build
```

构建结果位于：

```text
F:\Docker\weknora\frontend\dist
```

随后构建仅包含静态文件的前端镜像：

```powershell
Set-Location F:\Docker\weknora
docker build --file frontend\Dockerfile --tag vone/weknora-ui:0.7.2-vone.2 frontend
```

不需要因为 UI v0.1 重新构建 App、Docreader、PostgreSQL 或 Redis 镜像。

## 4. 本地 Docker 发布

根目录 `.env` 中设置：

```dotenv
VONE_IMAGE_TAG=0.7.2-vone.1
VONE_UI_IMAGE_TAG=0.7.2-vone.2
```

仅更新前端容器：

```powershell
Set-Location F:\Docker\weknora
docker compose -f docker-compose.yml -f deploy/docker-compose.vone.yml up -d --no-deps frontend
docker compose -f docker-compose.yml -f deploy/docker-compose.vone.yml ps frontend
docker image inspect vone/weknora-ui:0.7.2-vone.2
```

浏览器访问 `http://localhost:${FRONTEND_PORT}`，验证登录/注册、主题、语言、侧栏品牌和核心页面。

## 5. 回滚

将 `.env` 的 `VONE_UI_IMAGE_TAG` 改回上一个已验收标签，然后只重建前端容器：

```powershell
docker compose -f docker-compose.yml -f deploy/docker-compose.vone.yml up -d --no-deps --force-recreate frontend
```

回滚不触碰数据库卷、Redis 数据、上传文件和后端容器。

## 6. 后续上游升级策略

1. 先从上游新 tag 建升级分支，不直接覆盖已发布分支；
2. 先合并上游并完成原版构建，再重放 `customizations/vone`；
3. 优先处理四个接入点冲突，不把 VONE CSS 分散到业务页面；
4. 对照上游 `Login.vue` 的认证、邀请注册、OIDC 新逻辑，确保 VONE 外壳 props/事件同步；
5. 依次执行 type-check、build、桌面/移动端、light/dark/system 和真实认证回归；
6. 新版本采用 `vone/weknora-ui:<新上游版本>-vone.<revision>`，旧镜像至少保留一个发布周期。

## 7. 验证状态

- TypeScript/Vue 类型检查：`npm run type-check`，通过；
- 前端测试：`npm test`，341/341 通过；
- Vite 生产构建：`npm run build`，6345 个模块，构建通过；
- 桌面登录/注册视觉：1536×1024 浏览器检查通过；
- 移动端登录/注册视觉：390×844 浏览器检查通过；
- 浅色/深色/跟随系统：切换通过，匿名偏好写入 `WeKnora_anon_theme`；
- 表单错误路径：空登录提交能显示邮箱、密码必填校验；
- Docker 镜像：已于 `2026-08-11` 构建并发布为 `vone/weknora-ui:0.7.2-vone.2`；本轮提交复核时 Docker Desktop Linux Engine 未启动，因此未重复执行容器检查；
- 真实后端登录、邀请注册和 OIDC：当前静态预览未连接后端，需要部署环境与账号，由 QA/用户验收。

## 8. 原型一致性核对

| 对照项 | v0.1 结果 |
| --- | --- |
| 左右分屏与品牌叙事区 | 已实现，桌面约 52/48，移动端改为上下布局 |
| VONE Logo、双语产品名 | 已实现，深色自动使用浅色 Logo |
| 登录/注册表单和企业身份入口 | 登录、注册和 OIDC 条件入口均保留原逻辑 |
| 浅色/深色/跟随系统 | 已实现并复用上游偏好存储 |
| 蓝色设计体系 | 根变量计算值浅色 `#205895`、深色 `#6fa2dd` |
| 版权与上游归属 | `@vonechina 2026` 与 `Powered by Tencent WeKnora` 已展示 |

浏览器证据：

- `prototype/vone-knowledge-ui/screenshots/v0.1-login-desktop-light.png`；
- `prototype/vone-knowledge-ui/screenshots/v0.1-login-desktop-dark.png`；
- `prototype/vone-knowledge-ui/screenshots/v0.1-login-mobile-light.png`；
- `prototype/vone-knowledge-ui/screenshots/v0.1-register-mobile-system.png`。

## 9. 本地发布记录

- Release ID：`VONE-UI-0.1.0-WEKNORA-0.7.2`；
- 发布时间：`2026-08-11`；
- 发布环境：本机 Docker Desktop，Compose Project `vone-weknora`；
- 已部署镜像：`vone/weknora-ui:0.7.2-vone.2`；
- 镜像 ID：`sha256:faae205cd7ed7174ee833ee8ba0114c3f235eee4d5bdd92d1d9c1a1f274f9210`；
- 前端容器：`WeKnora-frontend`，容器 ID 前缀 `c833765ab7e8`；
- 发布方式：`--no-deps --force-recreate frontend`；
- 未重建服务：App、Docreader、PostgreSQL、Redis；
- 发布验证：首页与登录页 HTTP 200，认证配置接口 HTTP 200，App/PostgreSQL/Docreader healthy；
- 浏览器验证：标题、品牌配置、设计令牌和登录页展示通过；
- 发布截图：`prototype/vone-knowledge-ui/screenshots/v0.1-release-localhost.png`；
- 回滚镜像：`vone/weknora-ui:0.7.2-vone.1`，镜像 ID 前缀 `sha256:4dc80c9e`。
