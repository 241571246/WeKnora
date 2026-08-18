# Vone Task Plan

## Task ID

- TASK-20260817-QA-ROCKL-01

## Last Updated

- 2026-08-17 12:34

## Task

- 知识库 ACL 独立性待补的 QA 与发布验证

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
- QA Independence Exception Approver: Tim (candidate; stable employee code and approval evidence pending)
- Owner Assignment Gate: PASS
- QA Independence Gate: EXCEPTION REQUIRED
- Delivery Model: Concentrated Team
- Change Risk: High
- Verification Mode: Mixed
- Technical Verification Gate: PARTIAL
- Assurance Policy: N/A — High-risk work cannot use solo policy approval
- Risk Acceptance Owner: ROCKL
- Business Acceptance Gate: PENDING
- Release Authorization Gate: PENDING
- Estimate Unit: Ideal Hours
- Execution Model: Hybrid
- Primary Executor: ROCKL / Codex Agent
- Reuse Evidence: current Gin/GORM/Vue KB RBAC, kb_access, audit, folder_path and Agent selection code referenced by canonical plan
- Estimated Active Effort: 9～13.5h
- Estimated Implementation Active Effort: 9～13.5h
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

- Current Batch: V1/R1
- Run Until: V1/R1 exit gate
- Verification Cadence: Per WP static / Per Batch targeted / Code Complete integrated / Closure L3
- State Update Cadence: lightweight checkpoint / batch boundary
- Pause Only On: unapproved decision / dangerous external write / missing authority or environment / irreducible task blocker / unsafe worktree overlap
- Continue Without User Prompt: Yes

## Execution Batch Register

| Batch | Work Packages | Entry Gate | Exit Gate | Heavy Verification |
| --- | --- | --- | --- | --- |
| V1/R1 | B44-B52 | Dependencies satisfied | Task technical exit gate | Per canonical plan |

## Goal

- 聚合 V1/R1 技术证据并执行 B48 独立运行时验收和发布门禁。

## UI Skill Selection

- Selected UI Skill: N/A
- Integration Mode: sop
- Capability Mode: plan
- Selection Evidence: non-UI task
- Fallback Skill: vone-vue-admin-ui-builder

## Scope

### In Scope

- B44-B52：权限/IDOR/迁移/前端测试、B48 E2E、全量验证、诊断、部署回滚、manifest 和上游合并演练。

### Out of Scope

- 未经授权的生产部署、人工审批替代或 destructive rollback。

## Milestones

### M1

Owner Skill: `vone-verify-before-done`

- [x] 自动化 L1/L2、迁移、构建、回滚包和上游合并演练完成（B44-B47、B49-B52）

### M2

Owner Skill: `vone-release-readiness`

- [ ] 使用不可变构建物完成并签署真实身份 E2E（B48）

## Risks

- High-risk ACL 不能由同一实施者自验关闭；保持 QA Independence EXCEPTION REQUIRED。

## Dependencies

- D1-D2-S1-A1-T1-U1 技术完成，另需不可变 commit、staging identities 和 distinct human QA。

## Plan Phase Status

- Baseline Implementation: B44-B47 and B49-B52 technically complete; B48 pending
- Governance Recovery: N/A
- Amendment: N/A
- Remaining Closure: immutable scoped commit, B48 independent runtime QA, business acceptance and release authorization

## Closure Gates

- Completion Scope: Task Technical
- Technical Verification Gate: PARTIAL
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
