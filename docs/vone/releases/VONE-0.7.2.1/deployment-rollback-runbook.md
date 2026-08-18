# VONE-0.7.2.1 Deployment and Access-Safe Rollback

## Preconditions

1. Back up the database and record the restore identifier.
2. Freeze membership/KB administration writes for the migration window.
3. Verify the candidate binary contains the VONE migration runner and Authorizer.
4. Inventory `kb_shares`; this release retains them as cross-workspace shares and never converts them to employee memberships.

## Deployment sequence

1. Deploy the ACL-aware backend with `WEKNORA_VONE_KB_ACL_MODE=shadow`.
2. Run the independent VONE migration history. Do not copy the SQL into upstream migration numbering.
3. Execute `reconciliation-postgresql.sql`; require zero ownerless KBs, zero invalid capability rows and zero orphan bindings.
4. Confirm `kb_shares` row count and semantics are unchanged.
5. Configure representative memberships and verify shadow-deny logs (`[kb_acl_metric]`).
6. Deploy the capability-aware frontend.
7. Change the backend to `WEKNORA_VONE_KB_ACL_MODE=enforce` in a controlled window and restart only through the normal deployment mechanism.
8. Run Owner/co-owner/editor/viewer/AI-only, organization-share, Agent and API-key smoke tests.

Unknown or empty mode values resolve to `enforce`. `off` is an explicit emergency compatibility mode and must not be used as an access-safe rollback after enforcement.

## Access-safe rollback

1. Stop new administrative writes and preserve `enforce`.
2. Roll back only to the last ACL-aware VONE binary/frontend pair. Never deploy upstream 0.7.2 Standard after ACL activation because it does not enforce employee memberships.
3. If no ACL-aware binary can run, place KB APIs behind maintenance mode or deny traffic before considering `shadow`/`off`.
4. Keep additive VONE tables in place during an application rollback; they are harmless to upstream tables and preserve grants for forward recovery.
5. Use the `.down.sql` migration only after an approved data-loss decision, export of all VONE ACL/tree rows, and confirmation that no ACL-dependent binary is serving traffic.
6. Re-run reconciliation and restore traffic only after access behavior is verified.

## Rollback invariant

Rollback must not broaden visibility. A successful process rollback with broader KB access is a failed security rollback.
