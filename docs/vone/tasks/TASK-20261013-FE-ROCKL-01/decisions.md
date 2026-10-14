# Vone Decisions

## Task ID

- TASK-20261013-FE-ROCKL-01

## Last Updated

- 2026-10-13 09:25

## Current Decisions

### D1

#### Decision

- 用户确认采用 /login?autologindingtalk=true；仅自动发起既有认证，不绕过授权。

#### Reason

- 减少工作台点击，保留普通浏览器和原登录按钮。

#### Impact

- 加性前端变更；精确 true；消费参数防刷新循环；邀请/注册不自动触发。

#### Follow-up

- 自验不等于真实账号 QA；当前 SSH 不可达，不宣称发布。
