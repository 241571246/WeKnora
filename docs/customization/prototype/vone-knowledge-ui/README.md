# VONE知识库 UI 静态原型

本目录是 `VONE知识库 / Vone Knowledge` 的第一版可点击 HTML 原型，用于冻结品牌视觉和交互方向。它没有接入 WeKnora API，也没有修改现有 Vue 业务代码。

## 页面入口

- `index.html#login`：登录页；
- `index.html#register`：注册页；
- `index.html#home`：登录后的知识库首页。

在登录页点击“登录”可进入知识库首页；点击“创建账户”可切换到注册页。页面同时提供浅色、深色、跟随系统和中英文切换。

## 本地预览

可以直接用浏览器打开 `index.html`。若浏览器限制本地资源，可在本目录执行：

```powershell
python -m http.server 4179 --bind 127.0.0.1
```

然后打开 `http://127.0.0.1:4179/index.html#login`。

## 目录说明

```text
vone-knowledge-ui/
├── index.html
├── styles.css
├── app.js
├── assets/
│   ├── favicon.ico
│   ├── vone-logo-blue.png
│   └── vone-logo-light.png
├── concepts/
│   ├── login-concept.png
│   └── knowledge-home-concept.png
└── screenshots/
    ├── login-desktop.png
    ├── register-desktop.png
    ├── knowledge-home-desktop.png
    ├── knowledge-home-dark.png
    ├── login-mobile.png
    └── knowledge-home-mobile.png
```

## 已冻结的品牌输入

| 项目 | 内容 |
| --- | --- |
| 中文产品名 | VONE知识库 |
| 英文产品名 | Vone Knowledge |
| 主色 | `#205895`，从蓝色 Logo 提取 |
| 版权 | `@vonechina 2026` |
| 上游归属 | `Powered by Tencent WeKnora` |
| Logo 规则 | 浅色界面用蓝色 Logo，深色界面用白色 Logo |
| 主题 | 浅色、深色、跟随系统 |

## 与生产 Vue 的边界

该原型只表达视觉和交互，不应整页复制到生产前端。生产实施应把品牌配置、设计令牌、TDesign 覆盖和品牌资产落入独立 `frontend/src/customizations/vone/` 扩展层，再由登录页、全局菜单和 HTML 元信息读取统一配置。

认证、路由、Pinia Store、请求层、知识库业务组件和后端 API 契约继续使用 WeKnora 上游实现。这样后续升级主要处理少量品牌接入点，而不是维护整套长期前端分叉。

## 浏览器验证结果

- 桌面视口：1440 × 960；
- 窄屏视口：390 × 844；
- 已验证登录、注册、登录进入首页、新建知识库弹窗、移动菜单、浅色/深色和中英文切换；
- 浏览器控制台错误：0；
- 本阶段仅为 UI 原型验证，不代表 WeKnora 真实认证、注册和知识库接口已联调。
