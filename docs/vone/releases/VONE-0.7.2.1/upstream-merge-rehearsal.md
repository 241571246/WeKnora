# Upstream Merge Rehearsal Report

Recorded: 2026-08-18

## Rehearsal identity

- Customized baseline: `9524f7310b8d5636307afc6365966ab513c19a1f` (`custom/v0.7.2`) plus the VONE-0.7.2.1 candidate diff.
- Upstream target: `9b4f792a04d82bef630a2b2dc95344b3dad2649d` (`upstream/main`), 85 commits after the customized baseline.
- Isolated staging commit: `fd9e384699263f54e66f0b6878598b70dc59a602`.
- Rehearsal merge commit: `f226d491daa066210dab03c5a64443a182388cdf`.
- Isolation: temporary `codex/vone-0721-merge-rehearsal` worktree; the real `custom/v0.7.2` branch was not merged or committed.

## Conflict and compatibility inventory

Five content conflicts required a semantic merge:

| File | Preserved VONE behavior | Preserved upstream behavior |
|---|---|---|
| `frontend/src/components/doc-content.vue` | capability-gated document/chunk preview and chunk delete | chunk pagination and loading |
| `frontend/src/hooks/useKnowledgeBase.ts` | optional `loadChunks` permission gate | request generation and active-KB handling |
| `frontend/src/views/knowledge/KnowledgeBase.vue` | 17-capability UI projection | upstream download/list props |
| `internal/handler/session/handler.go` | central KB Authorizer dependency | artifact collector and memory service |
| `internal/handler/session/qa.go` | caller-authorized Agent KB intersection | attachment validation |

Two upstream test-infrastructure defects were corrected only in the isolated merge candidate:

- `deployment_capabilities_test.go`: normalize CRLF whitespace before removing a trailing comma.
- `tenant_api_key_test.go`: synchronize the asynchronous fake repository so `-race` tests the production path instead of racing the test double.

## Verification result

| Gate | Result | Evidence |
|---|---|---|
| unresolved conflicts / whitespace | PASS | no unmerged paths; `git diff --check` clean before commit |
| full Go suite | PASS | Linux/WSL, Go 1.26.5, CGO enabled, temporary `libsqlite3-dev` headers; `go test ./...` |
| security race selection | PASS | Authorizer, collection, middleware, HTTP, Agent intersection, route contract and API-key async touch under `go test -race` |
| frontend typecheck | PASS | `npm run type-check` |
| frontend tests | PASS | 398/398 |
| i18n contract | PASS | 11/11 |
| frontend production build | PASS | Vite, 6396 modules; existing large-chunk warnings only |
| release-package validator | PASS | migrations, retained `kb_shares`, diagnostics, rollback boundaries and SHA-256 inventory |

The first Windows-native full Go invocation was invalid because CGO was disabled. The first WSL invocation was interrupted by an Ollama module-download EOF; the next exposed a missing SQLite development header. Neither is counted as a product failure. The valid run used Linux CGO and a temporary, non-system `/tmp` header extraction.

`npm ci` reported 8 dependency audit findings inherited by the merged dependency tree (2 moderate, 6 high). They are a separate dependency-security review item; no unreviewed `npm audit fix` was applied.

## Upgrade and rollback conclusion

The customization does not prevent upstream upgrades. It is maintained as an additive VONE schema/ledger plus central Authorizer and capability-aware extensions, but every upgrade remains a controlled semantic merge rather than a blind source replacement.

For the real upgrade: create an integration branch, reproduce the five semantic resolutions, retain the two cross-platform test fixes where still applicable, run upstream and VONE migrations on a database copy, rerun the complete gates above, and deploy in `shadow` before `enforce`. On failure, preserve `enforce`, restore the last checksummed ACL-aware backend/frontend pair, retain additive VONE tables, and use the down migration only after an approved data-loss decision.
