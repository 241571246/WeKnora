# VONE-0.7.2.1 B48 QA Execution Pack

## 1. Purpose and evidence boundary

This pack executes B48 against one deployed staging baseline. It does not authorize production. The tester must be a named human distinct from ROCKL; RockL and ROCKL are the same identity and therefore cannot sign the Independent QA field.

Record for every case: timestamp, actor user ID, tenant ID, KB ID, HTTP method/path, status, response summary, browser screenshot or recording reference, and matching audit row where required. Redact tokens and document content outside the test corpus.

## 2. Candidate and environment preflight

- Backend candidate: `docs/vone/runtime/TASK-20260817-QA-ROCKL-01/backend/weknora-server`
  - SHA-256: `D9793B497B18D3E45BA853DDAF3DCA66F9752945087B37CEA7169CB791E88896`
- Frontend candidate: `docs/vone/runtime/TASK-20260817-QA-ROCKL-01/frontend/weknora-frontend-VONE-0.7.2.1-candidate1.zip`
  - SHA-256: `C6761EC39DAACB0D6CBBDB63BDA0E5FB19314B486609B3130696368731583D33`
- Source status: `9524f731+dirty`; rebuild from an authorized immutable commit before formal QA sign-off.
- Start with `WEKNORA_VONE_KB_ACL_MODE=shadow`; record `/api/v1/system/info` fields `kb_acl_mode`, `kb_acl_metrics`, `vone_db_version`, `vone_db_error`.
- Run the read-only `reconciliation-postgresql.sql`; all anomaly counts must be zero and the retained `kb_shares` count must match the pre-migration record.
- Switch to `enforce` only after shadow decisions match expected outcomes, then repeat all denial cases.

## 3. Required identities and data

| Alias | Workspace role / relation | Required use |
|---|---|---|
| WSO | Workspace Owner, implementation-independent test actor | collection governance and whole-workspace view |
| O1 | employee KB creator / Owner | create KB and grant members |
| O2 | employee co-owner | owner lifecycle and delegation |
| ED | employee Editor | content mutation without member/settings/KB-delete rights |
| DV | employee Document Viewer | list/preview/download without mutation |
| AI | employee AI User | AI answer and citations only |
| CU | employee Custom member | independent capability combinations |
| OUT | same-workspace unassigned employee | list/tree/direct-ID denial |
| ORG | organization-share-only user | retained cross-workspace sharing semantics |

Create KB-A by O1 with one test document and at least two chunks. Create KB-B outside AI/OUT authorization. Configure one Agent with KB-A and KB-B. Create one scoped API key allowing `retrieve` for KB-A only. Use non-sensitive synthetic content with a unique citation marker.

## 4. Membership and ownership matrix

| ID | Action | Expected |
|---|---|---|
| M01 | O1 creates KB-A | creator immediately has role `owner` and all 17 capabilities |
| M02 | O1 `POST /api/v1/knowledge-bases/{KB-A}/members` with `{"user_id":"{O2}","role":"owner"}` | 201; O2 can manage ordinary members and owners |
| M03 | O2 grants ED, DV and AI using roles `editor`, `document_viewer`, `ai_user` | 201; presets match the frozen capability contract |
| M04 | O2 grants CU role `custom` with only selected `capabilities` | 201; access response contains exactly the stored custom capabilities |
| M05 | remove or demote O1 while O2 remains Owner | 200; audit records owner change |
| M06 | attempt to remove or demote the final Owner | 409; membership remains unchanged |
| M07 | member manager without `kb.owners.manage` attempts Owner mutation | 403 |
| M08 | revoke ED and immediately retry a previously allowed endpoint | denied on the next request; record elapsed time and prove under 5 seconds |

For every actor/capability pair query `GET /api/v1/knowledge-bases/{KB-A}/access?capability={capability}` and retain `allowed`, `source`, `role`, and returned capability list.

## 5. Capability and direct-ID matrix

Run both from the browser and direct HTTP using known KB/document/chunk IDs.

| Actor | Positive cases | Mandatory negative cases |
|---|---|---|
| ED | list documents; preview document/chunks; download; upload/edit/reparse; delete document/chunk; manage folders | member list/mutation, settings edit, whole-KB delete |
| DV | document list, full preview, chunk preview, download, AI query | upload/edit/reparse, document delete, chunk edit/delete, folder/settings/member/KB mutation |
| AI | AI answer includes permitted filename/source and necessary quotation | document list, full preview, raw chunk, download, upload/edit/delete and settings/member APIs |
| CU | only explicitly granted operations | all 16 ungranted capabilities |
| OUT | none for KB-A | KB list/tree exposure and every direct-ID read/write |

At minimum exercise:

- `GET /api/v1/knowledge-bases/{KB}/knowledge`
- `GET /api/v1/knowledge/{DOC}` and `/preview` and `/download`
- `GET /api/v1/chunks/{DOC}` and `/api/v1/chunks/by-id/{CHUNK}`
- `PUT /api/v1/knowledge/manual/{DOC}` and `POST /api/v1/knowledge/{DOC}/reparse`
- `DELETE /api/v1/knowledge/{DOC}` and `DELETE /api/v1/chunks/{DOC}/{CHUNK}`
- `DELETE /api/v1/knowledge-bases/{KB}`

Expected denial is 403 without content leakage. A missing or cross-workspace resource may be normalized to a non-revealing 403/404 according to the route contract; record the exact response and confirm it does not reveal source tenant, filename, chunk text or existence beyond that contract.

## 6. AI, Agent, organization share and API key

| ID | Action | Expected |
|---|---|---|
| A01 | AI asks a question answered only by KB-A marker | answer may show filename, source and necessary cited excerpt; no content-management UI/API access |
| A02 | revoke `kb.ai.query`, retry direct AI and Agent | KB-A is not searched within 5 seconds |
| A03 | Agent configured with KB-A and KB-B, caller authorized only for KB-A | effective Agent targets contain KB-A only |
| A04 | caller authorized for neither configured KB | pure/no-KB result; no source leakage |
| A05 | ORG accesses KB-A only through retained organization share | permitted organization-share read/AI semantics pass through Authorizer; no employee membership row is synthesized |
| A06 | scoped API key queries KB-A | allowed only when key capability and KB allow-list both match |
| A07 | same key queries KB-B or a disallowed capability | 403 and denial audit/counter increment |

## 7. Collection tree and folder regression

| ID | Action | Expected |
|---|---|---|
| T01 | WSO creates root/child collections using `POST /api/v1/knowledge-base-collections` | 201; sort order retained |
| T02 | WSO renames/moves/sorts a node using `PATCH /api/v1/knowledge-base-collections/{ID}` | 200; no cycle allowed |
| T03 | bind KB-A using `PUT /api/v1/knowledge-bases/{KB-A}/collection` | one binding only |
| T04 | bind with empty `collection_id` | KB-A returns to Unclassified |
| T05 | delete a collection with child nodes or KB bindings | 409 and no cascade deletion |
| T06 | OUT/AI fetch `GET /api/v1/knowledge-base-collections/tree` | only branches containing visible KBs; hidden IDs absent |
| T07 | exercise the existing document `folder_path` tree inside KB-A | behavior unchanged; collection membership does not grant or inherit KB permission |

## 8. Audit, metrics and rollback proof

- Query `/api/v1/knowledge-bases/{KB-A}/activity` after grant, revoke, owner change, document delete, chunk delete, download, denial and collection binding. Confirm actor, KB scope, action, outcome and timestamp.
- Capture `kb_acl_metrics` before and after allowed, denied, shadow-denied and error scenarios; counters must move in the expected low-cardinality bucket.
- Execute the access-safe application rollback rehearsal from `deployment-rollback-runbook.md`: keep ACL enforcement, restore only an ACL-aware backend/frontend pair, retain additive VONE tables, rerun reconciliation and denial smoke.
- Do not execute the VONE down migration unless the approved destructive rollback prerequisites in the runbook are satisfied.

## 9. Sign-off record

| Field | Value |
|---|---|
| Deployed immutable commit/tag | Pending |
| Backend/frontend SHA-256 readback | Pending |
| Environment / deployment time | Pending |
| Independent human QA name/code | Pending |
| Executed cases / failures | Pending |
| Evidence directory/link | Pending |
| Technical QA result | Pending |
| Business acceptance approver/result | Pending |
| Release Owner authorization | Pending |

Any cross-workspace content leak, ownerless KB, unexpected Agent target, missing audit, migration anomaly or rollback to a non-ACL backend is a release-blocking failure.
