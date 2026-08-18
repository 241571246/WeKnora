# VONE-0.7.2.1 Release Manifest

## Metadata

- Release ID: VONE-0.7.2.1
- Status: Blocked
- Release Owner: ROCKL
- Target Environment: local staging deployed in shadow; production only after human approvals
- Scheduled At: Not scheduled because release gates are blocked
- Scope Freeze At: 2026-08-17 20:34 +08:00
- Baseline: upstream source commit 9524f7310b8d5636307afc6365966ab513c19a1f; scoped VONE commit 3693b7990176217304bc921bd01f422a8f1d5a55
- QA Independence Gate: EXCEPTION REQUIRED
- Project Board: ../../project-board.md

## Scope

| Requirement ID | Task IDs | PR / Commit / Artifact | QA Status | Include |
|---|---|---|---|---|
| REQ-2026-001 | TASK-20260817-SEC-ROCKL-01; TASK-20260817-DB-ROCKL-01; TASK-20260817-BE-ROCKL-01; TASK-20260817-AI-ROCKL-01; TASK-20260817-TREE-ROCKL-01; TASK-20260817-FE-ROCKL-01; TASK-20260817-QA-ROCKL-01 | Commit `3693b7990176217304bc921bd01f422a8f1d5a55`; migrations/vone; this release package | Technical automation passed; B48 and independent QA not passed | Yes |

## Artifacts

| Component | Version / Checksum | Source | Owner | Evidence |
|---|---|---|---|---|
| backend Linux QA candidate | SHA-256 `4D3C00EFFD5E423BBEDB11ADD94A6EDB6CB88467661902DE8F42423CE18951BB`; 334205112 bytes | `docs/vone/runtime/TASK-20260817-QA-ROCKL-01/backend/weknora-server` | ROCKL | clean Linux amd64 CGO build; Go 1.26.5; `vcs.revision=3693b7990176217304bc921bd01f422a8f1d5a55`; `vcs.modified=false` |
| frontend QA candidate | SHA-256 `189E3FEC361611A21C8F3E365F34AD81FD317D09CE6A9A9A009CF6E92B83AC4A`; 11217904 bytes | `docs/vone/runtime/TASK-20260817-QA-ROCKL-01/frontend/weknora-frontend-VONE-0.7.2.1-candidate1.zip` | ROCKL | isolated-clone build PASS; ZIP integrity PASS, 181 entries, `index.html` present, no unsafe entry |
| backend source candidate | VONE-0.7.2.1; commit `3693b7990176217304bc921bd01f422a8f1d5a55` | custom/v0.7.2 | ROCKL | ../../../plans/2026-08-17-KB-PL-知识库权限与树形管理实施计划.md |
| frontend source candidate | VONE-0.7.2.1; commit `3693b7990176217304bc921bd01f422a8f1d5a55` | custom/v0.7.2 frontend | ROCKL | execution-status.md |
| PostgreSQL migration | VONE ledger version 1 | migrations/vone/versioned/000001_kb_acl_and_collections.up.sql | ROCKL | reconciliation-postgresql.sql |
| SQLite migration | VONE ledger version 1 | migrations/vone/sqlite/000001_kb_acl_and_collections.up.sql | ROCKL | execution-status.md |
| deployment/rollback runbook | VONE-0.7.2.1 | deployment-rollback-runbook.md | ROCKL | deployment-rollback-runbook.md |
| initial deployment guide | VONE-0.7.2.1 | initial-deployment-guide.md | ROCKL | source checkout, image build, shadow migration, reconciliation, enforce and rollback procedure |

The local QA candidates were rebuilt from immutable commit `3693b7990176217304bc921bd01f422a8f1d5a55` in isolated clean clones. Matching local staging images are deployed in `shadow`; no tag, PR or production release is claimed.

## Environment Changes

- Database: Executed and Verified
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
- Exception Approver: Tim / T001; `BUG-2026-001` / `TASK-20260817-TREE-ROCKL-01` scoped approval recorded; no release-wide exception granted.
- Exception Reason: Tim does not execute this frontend regression and approves ROCKL self-verification only for the narrowly scoped tree-filter fix.
- Exception Scope: `BUG-2026-001` tree-filter frontend regression and `TASK-20260817-TREE-ROCKL-01` local fix only; excludes the full REQ-2026-001 ACL, migration, B48 identity matrix and production release.
- Compensating Controls: authenticated Chrome behavior check; targeted tests 7/7; full frontend tests 351/351; i18n 11/11; type-check and production build; local image checksum and HTTP smoke; scoped diff review; no backend, ACL or database change.
- Evidence: `DEC-2026-002`; ../../bugs/BUG-2026-001-树形节点筛选未联动本空间知识库列表.md; approval conveyed by ROCKL in the Codex task on 2026-08-18. This is self-verification evidence, not Independent QA.
- Scope Limitation: release-wide QA Independence remains `EXCEPTION REQUIRED`.

## Gate Evidence

| Gate | Status | Evidence | Owner |
|---|---|---|---|
| Scope Frozen | PASS | requirement, security contract and plan V1.4 | ROCKL |
| Technical Verification | PASS | execution-status.md; acceptance-evidence.md; upstream-merge-rehearsal.md | ROCKL |
| QA Passed | BLOCKED | no real-identity browser/API matrix signed by a distinct reviewer | RockL |
| QA Independence | BLOCKED | Tim / T001 approved only the BUG-2026-001 tree-filter self-verification exception; full REQ-2026-001 and B48 release scope remains uncovered | ROCKL |
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
| QA Exception (BUG-2026-001 only) | Tim / T001 | Approved | 2026-08-18 11:34 +08:00 | DEC-2026-002; Tim did not execute testing and approved ROCKL self-verification only |
| Release Owner | ROCKL | Pending | Not recorded | No release authorization recorded |

## Open Blockers

- [x] Create the authorized scoped source commit and rebuild checksummed QA candidates from it — Owner: ROCKL — Evidence: commit `3693b7990176217304bc921bd01f422a8f1d5a55`; backend/frontend hashes recorded above. No tag was authorized or created.
- [ ] Execute B48 with real Owner, co-owner, editor, document viewer, AI-only, organization-share, Agent and API-key identities — Owner: RockL — Exit condition: signed matrix and retained HTTP/browser/audit evidence.
- [x] Admit the local staging deployment described by `b48-staging-preflight.md` — Owner: ROCKL — Evidence: backup SHA-256 `3D4B8DD107EB0B3C756C132B01700EA0007F6B3296F1520C5A094717CF3E48D7`; matching candidate images/hash; `shadow`; VONE ledger 1 clean; reconciliation and `kb_shares=2` PASS. Identity fixtures remain a separate B48 blocker.
- [ ] Assign a distinct Independent Reviewer for the full release, or obtain a separate release-wide exception from a different named human approver — Owner: ROCKL — Exit condition: release-wide QA Independence Gate is PASS or EXCEPTION APPROVED with evidence. `DEC-2026-002` satisfies only BUG-2026-001.
- [ ] Record human TL, QA and Release Owner approvals plus business acceptance — Owner: ROCKL — Exit condition: all required approval rows are Approved with evidence.
