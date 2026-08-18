# Vone Task Plan

## Task ID

- TASK-20260817-BE-ROCKL-01

## Last Updated

- 2026-08-17 12:34

## Task

- 知识库成员与能力鉴权后端

## Governance Context

- Assurance Model Version: 2
- Execution Contract Version: 2
- Project ID: PRJ-WEKNORA-VONE
- Requirement ID: REQ-2026-001
- Requirement Document: docs/vone/requirements/2026-08-17-KB-RE-知识库树形管理与精细化权限.md
- Change Document: N/A
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
- Estimated Active Effort: 10～15h
- Estimated Implementation Active Effort: 10～15h
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

- Current Batch: S1
- Run Until: S1 exit gate
- Verification Cadence: Per WP static / Per Batch targeted / Code Complete integrated / Closure L3
- State Update Cadence: lightweight checkpoint / batch boundary
- Pause Only On: unapproved decision / dangerous external write / missing authority or environment / irreducible task blocker / unsafe worktree overlap
- Continue Without User Prompt: Yes

## Execution Batch Register

| Batch | Work Packages | Entry Gate | Exit Gate | Heavy Verification |
| --- | --- | --- | --- | --- |
| S1 | B14-B23 | Dependencies satisfied | Task technical exit gate | Per canonical plan |

## Goal

- 实现知识库成员、共同 Owner、17 能力和所有敏感资源的后端强制授权。

## UI Skill Selection

- Selected UI Skill: N/A
- Integration Mode: sop
- Capability Mode: plan
- Selection Evidence: non-UI task
- Fallback Skill: vone-vue-admin-ui-builder

## Scope

### In Scope

- B14-B23：repository/service、final-owner 并发保护、Authorizer、handler/middleware/routes、三类删除与审计。

### Out of Scope

- AI/Agent 特殊调用、集合树 UI 和独立 QA。

## Milestones

### M1

Owner Skill: `vone-fullstack-change`

- [x] 成员仓储、服务和 Owner 并发不变量完成（B14-B18）

### M2

Owner Skill: `vone-api-contract`

- [x] Authorizer、路由、资源能力与审计完成（B19-B23）

## Risks

- 遗漏路由会允许 IDOR；由精确路由能力测试和直接 ID 测试缓解。

## Dependencies

- D1 合约和 D2 schema。

## Plan Phase Status

- Baseline Implementation: Completed and technically verified
- Governance Recovery: N/A
- Amendment: N/A
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
