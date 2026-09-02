# Vone Task Plan

## Task ID

- TASK-20260817-FE-ROCKL-01

## Last Updated

- 2026-08-17 12:34

## Task

- 知识库权限管理前端

## Governance Context

- Assurance Model Version: 2
- Execution Contract Version: 2
- Project ID: PRJ-WEKNORA-VONE
- Requirement ID: REQ-2026-001
- Requirement Document: docs/vone/requirements/2026-08-17-KB-RE-知识库树形管理与精细化权限.md
- Change Document: docs/vone/changes/2026-08-21-UI-CH-知识库项目结构左侧树形导航.md
- Canonical Plan: docs/vone/plans/2026-08-17-KB-PL-知识库权限与树形管理实施计划.md
- Plan Type: Forecast Implementation Plan
- Current Plan Version: V1.3
- Plan Baseline: Approved Planning Baseline V1.0
- Plan Accountable Owner: ROCKL
- Historical Executor: N/A
- Technical Approver: ROCKL
- Maintenance Developer Owner: ROCKL
- QA Owner: RockL
- QA Second Reviewer: TBD — distinct human required
- QA Independence Exception Approver: TBD — distinct human required
- Owner Assignment Gate: PASS
- QA Independence Gate: EXCEPTION REQUIRED
- Delivery Model: Concentrated Team
- Change Risk: High
- Verification Mode: Mixed
- Technical Verification Gate: PASS
- Assurance Policy: N/A — High-risk work cannot use solo policy approval
- Risk Acceptance Owner: ROCKL
- Business Acceptance Gate: PENDING
- Release Authorization Gate: PENDING
- Estimate Unit: Ideal Hours
- Execution Model: Hybrid
- Primary Executor: ROCKL / Codex Agent
- Reuse Evidence: current Gin/GORM/Vue KB RBAC, kb_access, audit, folder_path and Agent selection code referenced by canonical plan
- Estimated Active Effort: 8～12h
- Estimated Implementation Active Effort: 8～12h
- Estimated Human Review / QA Effort: tracked separately in TASK-20260817-QA-ROCKL-01
- Estimated Parallel Elapsed Effort: N/A
- Expected Delivery Elapsed Time: Not Committed
- External Wait Excluded: review, test identities, deployment window and release authorization
- Estimate Confidence: Medium
- Actual Effort: Not Recorded
- Schedule Commitment: Not Committed
- Target Release: VONE-0.7.2.1
- Implementation Admission: PASS

## Process Profile

- Profile: Controlled
- Profile Reason: High-risk ACL and migration scope under REQ-2026-001
- Assessed At: 2026-08-17T20:34:20+08:00
- Assessed By: ROCKL / Codex recorder
- Confidence: Medium
- Required Gates: PLAN, IMPLEMENTATION_ADMISSION, L1_L2_VERIFICATION, INDEPENDENT_QA, RELEASE_AUTHORIZATION
- Skipped Gates: None
- Escalation Triggers: scope drift, authorization bypass, unsafe migration or rollback

## Execution Control

- Current Batch: U2
- Run Until: U2 technical exit gate
- Verification Cadence: Per WP static / Per Batch targeted / Code Complete integrated / Closure L3
- State Update Cadence: lightweight checkpoint / batch boundary
- Pause Only On: unapproved decision / dangerous external write / missing authority or environment / irreducible task blocker / unsafe worktree overlap
- Continue Without User Prompt: Yes

## Execution Batch Register

| Batch | Work Packages | Entry Gate | Exit Gate | Heavy Verification |
| --- | --- | --- | --- | --- |
| U1 | B36-B43 | Dependencies satisfied | Task technical exit gate | Per canonical plan |
| U2 | UI Amendment: left tree navigation | User-approved change baseline | Targeted tests, typecheck, i18n and production build pass | Required |

## Goal

- 交付成员权限、AI-only 隔离和集合树的能力感知前端。

## UI Skill Selection

- Selected UI Skill: vone-vue-admin-ui-builder
- Integration Mode: sop
- Capability Mode: plan
- Selection Evidence: Vue 3 + TDesign repository; no project-specific root AGENTS.md binding
- Fallback Skill: vone-vue-admin-ui-builder

## Scope

### In Scope

- B36-B43：成员设置、共同 Owner 确认、能力投影、三类删除、内容预览门控、集合树和多语言。

### Out of Scope

- 以前端隐藏代替后端安全边界、独立浏览器 QA。

## Milestones

### M1

Owner Skill: `vone-vue-admin-ui-builder`

- [x] 成员与能力管理 UI 完成（B36-B40）

### M2

Owner Skill: `vone-fullstack-change`

- [x] 资源能力投影、AI-only 和集合树 UI 完成（B41-B43）

### M3 — Amendment U2

Owner Skill: `vone-fullstack-change`

- [x] 将顶部横向项目结构改为左侧可折叠树形导航。
- [x] 保持现有筛选、Owner 操作、接口和 ACL 语义。
- [x] 完成定向测试、类型检查、i18n、生产构建和模拟数据浏览器验证。

## Risks

- UI 控件错误会误导用户但不能成为安全边界；后端 Authorizer 保持最终裁决。

## Dependencies

- 冻结 API 合约、S1/A1/T1 response shape。

## Plan Phase Status

- Baseline Implementation: Completed and technically verified
- Governance Recovery: N/A
- Amendment: U2 completed and technically verified — knowledge-base collection navigation layout only; no API, schema or permission change
- Remaining Closure: immutable scoped commit, B48 independent runtime QA, business acceptance and release authorization

## Closure Gates

- Completion Scope: Task Technical
- Technical Verification Gate: PASS
- Change Evidence Review: Pending
- Change Evidence Report: TBD
- Workload Reconciliation: Not Assessable
- QA Independence: EXCEPTION REQUIRED
- Assurance Policy / Risk Acceptance Owner: N/A
- QA Independence Exception Reason / Scope: N/A
- QA Independence Compensating Controls / Evidence: N/A
- User Acceptance: PENDING
- Release Decision: BLOCKED
- Commit / PR: Pending explicit commit authorization; regenerate the read-only Plan immediately before commit to prevent drift
