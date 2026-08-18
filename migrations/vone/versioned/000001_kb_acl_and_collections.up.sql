-- VONE-0.7.2.1: per-knowledge-base ACL and workspace collection tree.
-- This migration has an independent version table (vone_schema_migrations)
-- so upstream WeKnora migration numbers can continue without collisions.

CREATE TABLE IF NOT EXISTS vone_kb_memberships (
    id          BIGSERIAL PRIMARY KEY,
    tenant_id   BIGINT NOT NULL,
    kb_id       VARCHAR(36) NOT NULL,
    user_id     VARCHAR(36) NOT NULL,
    role        VARCHAR(32) NOT NULL,
    granted_by  VARCHAR(36) NOT NULL,
    created_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at  TIMESTAMP WITH TIME ZONE,
    CONSTRAINT ck_vone_kb_memberships_role CHECK (
        role IN ('owner', 'editor', 'document_viewer', 'ai_user', 'custom')
    )
);

CREATE UNIQUE INDEX IF NOT EXISTS ux_vone_kb_memberships_active_user
    ON vone_kb_memberships(kb_id, user_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS ix_vone_kb_memberships_tenant_user
    ON vone_kb_memberships(tenant_id, user_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS ix_vone_kb_memberships_kb_role
    ON vone_kb_memberships(kb_id, role) WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS vone_kb_member_capabilities (
    membership_id BIGINT NOT NULL REFERENCES vone_kb_memberships(id) ON DELETE CASCADE,
    capability    VARCHAR(64) NOT NULL,
    PRIMARY KEY (membership_id, capability),
    CONSTRAINT ck_vone_kb_member_capability CHECK (capability IN (
        'kb.ai.query', 'kb.metadata.read', 'kb.documents.list',
        'kb.document.preview', 'kb.chunk.preview', 'kb.document.download',
        'kb.document.upload', 'kb.document.edit', 'kb.document.delete',
        'kb.document.reparse', 'kb.chunk.edit', 'kb.chunk.delete',
        'kb.folder.manage', 'kb.settings.edit', 'kb.delete',
        'kb.members.manage', 'kb.owners.manage'
    ))
);

CREATE TABLE IF NOT EXISTS vone_kb_collections (
    id          VARCHAR(36) PRIMARY KEY,
    tenant_id   BIGINT NOT NULL,
    parent_id   VARCHAR(36) REFERENCES vone_kb_collections(id),
    name        VARCHAR(128) NOT NULL,
    sort_order  INTEGER NOT NULL DEFAULT 0,
    created_by  VARCHAR(36) NOT NULL,
    created_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at  TIMESTAMP WITH TIME ZONE,
    CONSTRAINT ck_vone_kb_collection_not_self CHECK (parent_id IS NULL OR parent_id <> id)
);

CREATE UNIQUE INDEX IF NOT EXISTS ux_vone_kb_collections_active_sibling
    ON vone_kb_collections(tenant_id, COALESCE(parent_id, ''), name)
    WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS ix_vone_kb_collections_tenant_parent
    ON vone_kb_collections(tenant_id, parent_id) WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS vone_kb_collection_bindings (
    kb_id          VARCHAR(36) PRIMARY KEY,
    tenant_id      BIGINT NOT NULL,
    collection_id  VARCHAR(36) NOT NULL REFERENCES vone_kb_collections(id),
    updated_by     VARCHAR(36) NOT NULL,
    created_at     TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at     TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS ix_vone_kb_collection_bindings_collection
    ON vone_kb_collection_bindings(tenant_id, collection_id);

-- Existing knowledge bases belong to their valid same-workspace creator. If
-- creator_id is empty or invalid, assign the earliest active workspace Owner.
-- kb_shares is intentionally not converted; it remains cross-workspace input
-- to the central Authorizer.
WITH resolved_owner AS (
    SELECT kb.id AS kb_id,
           kb.tenant_id,
           COALESCE(
               (SELECT tm.user_id
                  FROM tenant_members tm
                 WHERE tm.tenant_id = kb.tenant_id
                   AND tm.user_id = kb.creator_id
                   AND tm.status = 'active'
                   AND tm.deleted_at IS NULL
                 LIMIT 1),
               (SELECT tm.user_id
                  FROM tenant_members tm
                 WHERE tm.tenant_id = kb.tenant_id
                   AND tm.role = 'owner'
                   AND tm.status = 'active'
                   AND tm.deleted_at IS NULL
                 ORDER BY tm.joined_at ASC, tm.id ASC
                 LIMIT 1)
           ) AS user_id
      FROM knowledge_bases kb
     WHERE kb.deleted_at IS NULL
)
INSERT INTO vone_kb_memberships
    (tenant_id, kb_id, user_id, role, granted_by, created_at, updated_at)
SELECT tenant_id, kb_id, user_id, 'owner', user_id, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
  FROM resolved_owner
 WHERE user_id IS NOT NULL
ON CONFLICT DO NOTHING;
