# KB ACL and Collection Security Contract

## Scope

- Release: `VONE-0.7.2.1`
- Requirement: `REQ-2026-001`
- Contract owner: `ROCKL`
- Default result for unresolved identity, resource or source: deny.

## Capability Contract

The persisted capability codes are defined in `internal/types/kb_acl.go`. There are 17 independently enforceable operations. File deletion, chunk deletion and knowledge-base deletion are separate capabilities.

Preset roles are frozen as follows:

| Preset | Effective access |
|---|---|
| Owner | All 17 capabilities |
| Editor | AI, metadata, document/chunk read, download, upload/edit/delete/reparse, chunk edit/delete and internal folder management; excludes KB settings, KB deletion and member/owner management |
| Document Viewer | AI, metadata, file list, full-text preview, chunk preview and download |
| AI User | AI query and metadata only; may receive source filename and answer citation snippets, but cannot list/download/preview source files or raw chunks |
| Custom | Only explicitly persisted capabilities |

## Ownership Invariants

- A knowledge base always has at least one active Owner membership after migration.
- A KB Owner may grant another active workspace employee as co-owner.
- Removing or demoting the final Owner is rejected transactionally.
- Workspace Owner is the administrator override and can inspect/manage every KB in the workspace.
- Existing KBs migrate to the valid current `creator_id`; an invalid or empty creator falls back to the earliest active Workspace Owner.

## Decision Precedence

1. Workspace Owner override.
2. Same-workspace KB membership and its effective preset/custom capabilities.
3. Existing cross-workspace `kb_shares`, mapped to capabilities by the central Authorizer.
4. Shared Agent context constrains AI retrieval to `Agent configured KBs ∩ caller kb.ai.query KBs`; it may expose answer citations but never creates an independent KB grant.
5. API-key scopes and KB allow-list, evaluated by the same Authorizer.
6. Default deny.

Organization shares are not converted into employee membership rows and no existing `kb_shares` call path may bypass the Authorizer.

## Visibility and Tree Contract

- Non-Owner employees list only KBs they own or have been assigned.
- The collection tree returns only branches containing at least one visible KB.
- A KB has zero or one collection binding; no binding means Unclassified.
- Collections may nest to arbitrary depth in phase 1, but must remain inside one workspace and must not form cycles.
- Collection nodes do not inherit permissions. Existing `knowledge.folder_path` remains the internal document-folder tree and is not a member ACL boundary.
- A collection with child collections or KB bindings cannot be deleted.

## API Contract

- `GET /knowledge-bases/:id/access`
- `GET|POST /knowledge-bases/:id/members`
- `PATCH|DELETE /knowledge-bases/:id/members/:user_id`
- `GET|POST /knowledge-base-collections[/tree]`
- `PATCH|DELETE /knowledge-base-collections/:id`
- `PUT /knowledge-bases/:id/collection`

All endpoints are tenant-scoped. Resource absence and unauthorized cross-tenant access must not reveal resource existence.

## Audit and Revocation

- Audit: grants, revocations, role/capability changes, ownership changes, file deletion, chunk deletion, KB deletion, download, authorization denials and collection changes.
- Authorizer reads are direct or cached with a maximum three-second TTL, meeting the five-second revocation objective.
- Mutation commits invalidate the affected `(tenant, kb, user)` decision entries.

## Migration and Rollback

- Upstream migrations continue in `schema_migrations`.
- VONE extensions use `vone_schema_migrations` and `migrations/vone/{versioned,sqlite}`.
- Rollback removes only VONE bindings, collections, capabilities and memberships. It does not alter `knowledge_bases`, `knowledge.folder_path`, `kb_shares` or upstream migration history.
- Authorization rollout order is schema, backfill validation, shadow comparison, enforcement, then legacy guard removal.
