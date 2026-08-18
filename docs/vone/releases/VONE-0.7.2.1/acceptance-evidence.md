# REQ-2026-001 Acceptance Evidence

Recorded: 2026-08-18

`TECHNICAL PASS` means code, contract, migration or automated-test evidence exists. It is not independent QA, business acceptance or release authorization. `PENDING L3` requires deployed real identities/browser/API evidence.

| AC | Technical evidence | State |
|---|---|---|
| AC-01 | create flow writes creator Owner membership; role-capability contract and migration backfill tests | TECHNICAL PASS; runtime identity pending |
| AC-02 | `kb.owners.manage` handler gates plus transactional final-owner repository tests, including concurrency | TECHNICAL PASS; browser pending |
| AC-03 | `ListAccessibleKBIDs`, route IDOR test and pruned collection tree | TECHNICAL PASS; multi-user browser/direct URL pending |
| AC-04 | workspace Owner governance source contract without implicit KB Owner override | TECHNICAL PASS; runtime workspace-owner matrix pending |
| AC-05 | three independent delete capabilities in Authorizer and frontend projection tests | TECHNICAL PASS; HTTP matrix pending |
| AC-06 | sensitive route contract test, body-ID resolution, direct IDOR denial and preview/download gates | TECHNICAL PASS; deployed direct-ID smoke pending |
| AC-07 | five presets and 17 capabilities; custom capability persistence and immediate decision tests | TECHNICAL PASS; browser editor pending |
| AC-08 | central middleware/direct Authorizer tests and exact route-capability inventory | TECHNICAL PASS |
| AC-09 | AI-only grants `kb.ai.query`; citation/source answer contract retained | TECHNICAL PASS; real-provider answer evidence pending |
| AC-10 | AI-only denies document/chunk/fulltext/download/edit in backend and UI projection tests | TECHNICAL PASS; browser/network evidence pending |
| AC-11 | central Authorizer treats Agent scope as a constraint, not a grant; regression proves configured+caller allow, outside-Agent deny and no-caller-grant deny | TECHNICAL PASS; runtime Agent evidence pending |
| AC-12 | revocation removes Authorizer access immediately and Agent resolution is per request | TECHNICAL PASS; measured under-five-second runtime evidence pending |
| AC-13 | collection create/update/move/order service and UI; cycle prevention tests | TECHNICAL PASS; browser interaction pending |
| AC-14 | unique KB binding and explicit unclassified move/list behavior | TECHNICAL PASS; browser pending |
| AC-15 | non-empty/cyclic collection rejection tests | TECHNICAL PASS |
| AC-16 | visible bindings plus ancestor-only pruning; tree never grants KB access | TECHNICAL PASS; multi-user browser pending |
| AC-17 | collection schema is separate from existing `folder_path`; upstream folder behavior retained | TECHNICAL PASS; regression browser pending |
| AC-18 | creator-to-Owner and workspace-Owner fallback backfill; zero-owner reconciliation | TECHNICAL PASS on SQLite/PostgreSQL rehearsals |
| AC-19 | bindings default to unclassified; VONE migrations do not mutate knowledge/chunk/`folder_path` rows | TECHNICAL PASS on migration rehearsals |
| AC-20 | `kb_shares` retained as cross-workspace source, no DML conversion, all access forced through Authorizer | TECHNICAL PASS; runtime organization-share smoke pending |
| AC-21 | member, owner, three deletes, download, denial and tree actions emit queryable KB-scoped audit records | TECHNICAL PASS; deployed audit-query evidence pending |
| AC-22 | fresh/upgrade PostgreSQL and SQLite rehearsals pass; validator confirms upstream migrations unchanged | TECHNICAL PASS |
| AC-23 | actual 85-commit upstream merge rehearsal, conflict inventory, full build/test/race and rollback procedure | TECHNICAL PASS |

## L3 execution set

B48 must retain HTTP status/body, browser screenshots or recording, audit rows and Actor/tenant/KB identifiers for Owner, co-owner, Editor, Document Viewer, AI User, organization share, Agent and scoped API key. A distinct human QA reviewer must sign the matrix; ROCKL/RockL self-verification cannot satisfy the High-risk independence gate.
