# VONE-0.7.2.1 B48 Staging Preflight

## Result

- Recorded At: 2026-08-18 10:27 +08:00
- Result: DEPLOYED_BLOCKED_FOR_B48_FIXTURES
- Evidence Mode: local staging deployment plus read-only runtime/schema reconciliation
- State Mutation: Candidate backend/frontend replaced the prior local containers; additive VONE migration version 1 executed in `shadow`; no down migration or real AI call was performed
- Repeatable Validator: `validate-b48-staging.ps1`

## Candidate Baseline

- Source commit: `3693b7990176217304bc921bd01f422a8f1d5a55`
- Backend candidate SHA-256: `4D3C00EFFD5E423BBEDB11ADD94A6EDB6CB88467661902DE8F42423CE18951BB`
- Frontend candidate SHA-256: `189E3FEC361611A21C8F3E365F34AD81FD317D09CE6A9A9A009CF6E92B83AC4A`

## Runtime Evidence

| Check | Observed | Conclusion |
|---|---|---|
| Compose project | `vone-weknora`, five containers running | Runtime exists |
| Running backend image | `vone/weknora-app:0.7.2.1-candidate.3693b799` | Candidate image active |
| Running frontend image | `vone/weknora-ui:0.7.2.1-candidate.3693b799` | Candidate frontend active |
| Running backend file | `/app/WeKnora`, SHA-256 `4D3C00EFFD5E423BBEDB11ADD94A6EDB6CB88467661902DE8F42423CE18951BB` | Exact candidate match |
| Backend health | `/health` returns HTTP 200 with proxy bypass; restart count 0 | Candidate runtime is healthy |
| System diagnostics | `/api/v1/system/info` returns HTTP 401 without token | Route is protected as expected |
| ACL mode | `WEKNORA_VONE_KB_ACL_MODE=shadow` | Initial observation mode active |
| Frontend smoke | `/` returns HTTP 200 with proxy bypass | Candidate static assets served |

## Database and Identity Evidence

| Check | Observed | B48 Impact |
|---|---:|---|
| VONE ACL/tree tables present | 5 of 5 expected; ledger `1`, `dirty=false` | Candidate migration executed |
| Users | 4 | Insufficient for the required distinct identity matrix |
| Tenants | 2 | Cross-workspace test topology is possible after preparation |
| Knowledge bases | 3 | Existing data must be protected and reconciled |
| Legacy `kb_shares` | 2 before and after migration | Retention invariant passed |
| Tenant members | 4 | Current role distribution is 2 owner and 2 viewer |
| Organization tenant members | 1 | Organization-share path has an existing fixture candidate |
| Tenant API keys | 0 | Scoped API-key cases need a test key after candidate deployment |

## Deployment Admission Conditions

1. [x] Local window, administration-write freeze and recoverable PostgreSQL backup recorded: `docs/vone/runtime/TASK-20260817-QA-ROCKL-01/staging/backups/20260818-102516/vone-pre-0721-20260818-102516.dump`, SHA-256 `3D4B8DD107EB0B3C756C132B01700EA0007F6B3296F1520C5A094717CF3E48D7`.
2. [x] Explicit local restart/replacement and additive migration authority received from ROCKL.
3. [x] Matching backend/frontend candidate images deployed; backend in-container hash matches the Candidate Baseline.
4. [ ] `shadow` is active; authenticated Workspace Owner `/api/v1/system/info` field capture remains part of B48.
5. [x] VONE migration version 1 and read-only reconciliation passed: five tables, clean ledger, zero ownerless KBs, three Owner memberships and retained `kb_shares=2`.
6. [ ] Prepare distinct WSO/O1/O2/ED/DV/AI/CU/OUT/ORG identities and non-sensitive test data. Current four-user population is not sufficient for the declared matrix.
7. [ ] Create a scoped test API key; never record its secret in tracked evidence.
8. [ ] Tim's stable identity is recorded as `Tim / T001`; explicit scoped exception approval evidence is still required, or assign a distinct independent QA reviewer.

Formal B48 execution may start only after the remaining identity and API-key fixtures are prepared. The local candidate deployment itself is admitted.

Run the read-only check from the repository root:

```powershell
& .\docs\vone\releases\VONE-0.7.2.1\validate-b48-staging.ps1
```

Exit code `0` means staging admission PASS, `2` means expected prerequisites are still BLOCKED, and `1` means the validator itself could not obtain trustworthy evidence. The post-deployment run parsed successfully and returned exit code `2` only for `B48_IDENTITIES_INSUFFICIENT` and `B48_API_KEY_FIXTURE_MISSING`.

## Deployment Notes

- Two candidate starts were rolled back before database migration: first for a Debian 12 GLIBC mismatch, then for a build-cache absolute GoJieba dictionary path. The final Ubuntu 24.04 runtime image preserves the exact candidate binary and maps the required dictionary path; the successful container is healthy with restart count 0.
- The earlier PowerShell HTTP 502 was a host proxy path; `curl --noproxy '*'` returned frontend 200, backend health 200 and unauthenticated system-info 401.
