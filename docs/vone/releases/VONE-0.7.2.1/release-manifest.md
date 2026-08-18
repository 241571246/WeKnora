# VONE-0.7.2.1 Release Manifest

## Metadata

- Release ID: VONE-0.7.2.1
- Status: Blocked
- Release Owner: ROCKL
- Target Environment: staging first, then production after human approvals
- Scheduled At: Not scheduled because release gates are blocked
- Scope Freeze At: 2026-08-17 20:34 +08:00
- Baseline: upstream source commit 9524f7310b8d5636307afc6365966ab513c19a1f plus the reviewed VONE candidate diff
- QA Independence Gate: EXCEPTION REQUIRED
- Project Board: ../../project-board.md

## Scope

| Requirement ID | Task IDs | PR / Commit / Artifact | QA Status | Include |
|---|---|---|---|---|
| REQ-2026-001 | TASK-20260817-SEC-ROCKL-01; TASK-20260817-DB-ROCKL-01; TASK-20260817-BE-ROCKL-01; TASK-20260817-AI-ROCKL-01; TASK-20260817-TREE-ROCKL-01; TASK-20260817-FE-ROCKL-01; TASK-20260817-QA-ROCKL-01 | VONE candidate diff on custom/v0.7.2; migrations/vone; this release package | Technical automation in progress; independent QA not passed | Yes |

## Artifacts

| Component | Version / Checksum | Source | Owner | Evidence |
|---|---|---|---|---|
| backend Linux QA candidate | SHA-256 `D9793B497B18D3E45BA853DDAF3DCA66F9752945087B37CEA7169CB791E88896`; 336594760 bytes | `docs/vone/runtime/TASK-20260817-QA-ROCKL-01/backend/weknora-server` | ROCKL | rebuilt after Agent-intersection correction; ELF x86-64, Go 1.26.5, module `9524f731+dirty`; dependency readback passed |
| frontend QA candidate | SHA-256 `C6761EC39DAACB0D6CBBDB63BDA0E5FB19314B486609B3130696368731583D33`; 11217916 bytes | `docs/vone/runtime/TASK-20260817-QA-ROCKL-01/frontend/weknora-frontend-VONE-0.7.2.1-candidate1.zip` | ROCKL | ZIP integrity PASS, 181 entries, `index.html` present |
| backend source candidate | VONE-0.7.2.1; immutable commit still pending | custom/v0.7.2 working tree | ROCKL | ../../../plans/2026-08-17-KB-PL-知识库权限与树形管理实施计划.md |
| frontend source candidate | VONE-0.7.2.1; immutable commit still pending | frontend candidate diff | ROCKL | execution-status.md |
| PostgreSQL migration | VONE ledger version 1 | migrations/vone/versioned/000001_kb_acl_and_collections.up.sql | ROCKL | reconciliation-postgresql.sql |
| SQLite migration | VONE ledger version 1 | migrations/vone/sqlite/000001_kb_acl_and_collections.up.sql | ROCKL | execution-status.md |
| deployment/rollback runbook | VONE-0.7.2.1 | deployment-rollback-runbook.md | ROCKL | deployment-rollback-runbook.md |

The local QA candidates are checksummed but were built from a `+dirty` source tree. No image, tag, PR or immutable release commit is claimed; source-to-artifact reproducibility remains blocked until a scoped commit is authorized.

## Environment Changes

- Database: Prepared
- Nacos: None
- Gateway: None
- Permission: Prepared

### Database Details

- Script / Migration: `migrations/vone/versioned/000001_kb_acl_and_collections.up.sql`; SQLite equivalent under `migrations/vone/sqlite`; read-only reconciliation in `reconciliation-postgresql.sql`.
- Environments: fresh database, upgraded WeKnora 0.7.2 database, staging, then production.
- Order: deploy ACL-aware backend in shadow mode; run independent VONE migration; reconcile; deploy frontend; switch to enforce only after shadow review.
- Compatibility Window: additive VONE tables coexist with upstream tables; ACL-aware binaries support shadow and enforce. An upstream-only 0.7.2 binary is not access-safe after activation.
- Rollback / Forward Fix: keep additive tables during application rollback; roll back only to an ACL-aware binary. Run `.down.sql` only after approved data-loss handling and traffic isolation.
- Schema Verification: require zero ownerless KBs, invalid capabilities and orphan bindings; verify the retained `kb_shares` count and VONE ledger version.

### Nacos / Configuration Details

- Namespace / Group / DataId / Key: no Nacos change; process environment key `WEKNORA_VONE_KB_ACL_MODE` is introduced.
- Old Value / New Value: absent to `shadow` for deployment, then `enforce` after verification; invalid or empty values fail closed to `enforce`.
- Backup: preserve the previous deployment manifest and ACL-aware backend/frontend artifact pair.
- Readback Verification: `/api/v1/system/info` must report `kb_acl_mode`, `kb_acl_metrics`, `vone_db_version` and no `vone_db_error`.

## QA Independence

- QA Owner: RockL
- Implementation Owners: ROCKL
- Independent Reviewer: Not assigned
- Exception Approver: Tim (candidate nominated; stable employee code and explicit approval evidence pending)
- Exception Reason: RockL and ROCKL are the same human identity; REQ-2026-001 is High risk and cannot use a concentrated-team policy shortcut.
- Exception Scope: VONE-0.7.2.1 authorization, migration, Agent and rollback verification.
- Compensating Controls: default-deny Authorizer; route-contract and IDOR tests; PostgreSQL/SQLite rehearsals; targeted race tests; shadow-mode counters and durable denial audit; access-safe rollback invariant.
- Evidence: docs/vone/tasks/TASK-20260817-QA-ROCKL-01/verification.md and execution-status.md; these are self-verification evidence, not independent QA.

## Gate Evidence

| Gate | Status | Evidence | Owner |
|---|---|---|---|
| Scope Frozen | PASS | requirement, security contract and plan V1.3 | ROCKL |
| Technical Verification | PASS | execution-status.md; acceptance-evidence.md; upstream-merge-rehearsal.md | ROCKL |
| QA Passed | BLOCKED | no real-identity browser/API matrix signed by a distinct reviewer | RockL |
| QA Independence | BLOCKED | Tim is nominated as exception approver, but stable employee code and explicit approval evidence are not recorded | ROCKL |
| DB and Config Prepared | PASS | PostgreSQL/SQLite rehearsal plus reconciliation-postgresql.sql and shadow/enforce diagnostics | ROCKL |
| Rollback Ready | PASS | deployment-rollback-runbook.md preserves ACL enforcement and additive data | ROCKL |
| Smoke Plan Ready | PASS | Owner/co-owner/editor/viewer/AI-only, organization-share, Agent and API-key matrix in the canonical plan | RockL |

## Deployment Plan

| Step | Action | Owner | Success Evidence | Failure Action |
|---|---|---|---|---|
| 1 | Freeze KB/member administration writes and record DB backup identifier plus `kb_shares` count | Release Owner | change freeze record, restore identifier, share count | stop; do not migrate |
| 2 | Deploy the checksummed ACL-aware backend with `WEKNORA_VONE_KB_ACL_MODE=shadow` | Release Owner | system info reports shadow and expected source/VONE versions | restore prior ACL-aware artifact |
| 3 | Run the independent VONE migration, then read-only reconciliation | DBA / Release Owner | VONE version 1; all anomaly queries return zero; retained share count matches | keep APIs unavailable; forward-fix or restore |
| 4 | Execute representative shadow decisions and inspect counters/audit rows | QA Owner | expected shadow-denied/allowed counters and KB-scoped audit records | remain in shadow; investigate |
| 5 | Deploy the matching capability-aware frontend | Release Owner | asset checksum and staging browser smoke | restore matching prior frontend |
| 6 | Switch to `enforce` in a controlled window | Release Owner | system info reports enforce; denial smoke returns 403 | preserve enforce and roll back only to ACL-aware pair |
| 7 | Run the complete role/share/Agent/API-key matrix and monitor | Independent QA | signed matrix, logs, audit queries and reconciliation | invoke access-safe rollback |

## Rollback Plan

- Trigger: any cross-workspace/IDOR access, ownerless KB, invalid capability, orphan binding, unexpected Agent scope, migration error, or sustained ACL authorization error rate.
- Decision Owner: human Release Owner; security visibility widening triggers immediate traffic isolation without waiting for business approval.
- Steps: stop administrative writes; preserve `enforce`; remove KB API traffic if required; restore the last checksummed ACL-aware backend/frontend pair; keep additive VONE tables; reconcile and rerun denial smoke before traffic restoration.
- Data Handling: prefer forward fix. Use VONE down migration only after export, approved data-loss decision, and confirmation that no ACL-dependent binary serves traffic.
- Recovery Verification: `/system/info` readback; VONE ledger; zero reconciliation anomalies; unchanged `kb_shares`; denied direct-ID request; Owner and AI-only smoke.

## Production Verification

- Smoke Cases: execute `qa-execution-pack.md`: create KB and automatic Owner; co-owner lifecycle and last-owner conflict; separate document/chunk/KB delete; AI-only answer with citations but no content APIs; Agent configured/caller intersection; organization share; scoped API key; visible-tree pruning; unclassified binding.
- Schema Readback: Pending production deployment.
- Configuration Readback: Pending production deployment.
- Monitoring Window: minimum 60 minutes after enforce plus next scheduled audit/worker cycle.
- Business Acceptance: Pending named business approver.

## Approvals

| Role | Approver | Status | Time | Evidence |
|---|---|---|---|---|
| TL | Not assigned | Pending | Not recorded | No human approval recorded |
| QA | RockL (same identity as implementation owner) | Pending | Not recorded | Self-verification cannot satisfy independence |
| Release Owner | ROCKL | Pending | Not recorded | No release authorization recorded |

## Open Blockers

- [ ] Authorize and create the scoped source commit/tag matching the checksummed QA candidates — Owner: ROCKL — Exit condition: immutable SHA is recorded, candidates are rebuilt from it, and artifact hashes are reproduced.
- [ ] Execute B48 with real Owner, co-owner, editor, document viewer, AI-only, organization-share, Agent and API-key identities — Owner: RockL — Exit condition: signed matrix and retained HTTP/browser/audit evidence.
- [ ] Assign a distinct Independent Reviewer, or obtain a scoped exception from a different named human approver — Owner: ROCKL — Exit condition: QA Independence Gate is PASS or EXCEPTION APPROVED with evidence.
- [ ] Record human TL, QA and Release Owner approvals plus business acceptance — Owner: ROCKL — Exit condition: all required approval rows are Approved with evidence.
