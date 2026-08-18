# Vone Task Plan

## Task ID

- TASK-20260817-AI-ROCKL-01

## Last Updated

- 2026-08-17 12:34

## Task

- AI 与 Agent 知识库权限收口

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
- Estimated Active Effort: 7～10.5h
- Estimated Implementation Active Effort: 7～10.5h
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

- Current Batch: A1
- Run Until: A1 exit gate
- Verification Cadence: Per WP static / Per Batch targeted / Code Complete integrated / Closure L3
- State Update Cadence: lightweight checkpoint / batch boundary
- Pause Only On: unapproved decision / dangerous external write / missing authority or environment / irreducible task blocker / unsafe worktree overlap
- Continue Without User Prompt: Yes

## Execution Batch Register

| Batch | Work Packages | Entry Gate | Exit Gate | Heavy Verification |
| --- | --- | --- | --- | --- |
| A1 | B24-B30 | Dependencies satisfied | Task technical exit gate | Per canonical plan |

## Goal

- 将直接 AI、Agent、API Key、组织共享与旧 `kb_shares` 收口到统一 Authorizer。

## UI Skill Selection

- Selected UI Skill: N/A
- Integration Mode: sop
- Capability Mode: plan
- Selection Evidence: non-UI task
- Fallback Skill: vone-vue-admin-ui-builder

## Scope

### In Scope

- B24-B30：AI query、引用边界、Agent 交集、撤权、API Key/组织共享和 legacy share 兼容。

### Out of Scope

- 模型质量调优、外部 IAM 和独立运行时 QA。

## Milestones

### M1

Owner Skill: `vone-fullstack-change`

- [x] 直接 AI、引用和 Agent 交集完成（B24-B27）

### M2

Owner Skill: `vone-api-contract`

- [x] legacy share、API Key、撤权与回归完成（B28-B30）

## Risks

- Agent 或共享链路范围扩大可泄漏来源内容；由每请求交集和 IDOR 测试缓解。

## Dependencies

- S1 central Authorizer 和 D2 schema。

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
