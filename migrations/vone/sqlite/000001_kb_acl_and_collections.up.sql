-- SQLite equivalent of VONE migration 000001.
CREATE TABLE IF NOT EXISTS vone_kb_memberships (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    tenant_id   INTEGER NOT NULL,
    kb_id       TEXT NOT NULL,
    user_id     TEXT NOT NULL,
    role        TEXT NOT NULL CHECK (role IN ('owner','editor','document_viewer','ai_user','custom')),
    granted_by  TEXT NOT NULL,
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at  DATETIME
);
CREATE UNIQUE INDEX IF NOT EXISTS ux_vone_kb_memberships_active_user
    ON vone_kb_memberships(kb_id, user_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS ix_vone_kb_memberships_tenant_user
    ON vone_kb_memberships(tenant_id, user_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS ix_vone_kb_memberships_kb_role
    ON vone_kb_memberships(kb_id, role) WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS vone_kb_member_capabilities (
    membership_id INTEGER NOT NULL REFERENCES vone_kb_memberships(id) ON DELETE CASCADE,
    capability    TEXT NOT NULL CHECK (capability IN (
        'kb.ai.query','kb.metadata.read','kb.documents.list','kb.document.preview',
        'kb.chunk.preview','kb.document.download','kb.document.upload','kb.document.edit',
        'kb.document.delete','kb.document.reparse','kb.chunk.edit','kb.chunk.delete',
        'kb.folder.manage','kb.settings.edit','kb.delete','kb.members.manage','kb.owners.manage'
    )),
    PRIMARY KEY (membership_id, capability)
);

CREATE TABLE IF NOT EXISTS vone_kb_collections (
    id          TEXT PRIMARY KEY,
    tenant_id   INTEGER NOT NULL,
    parent_id   TEXT REFERENCES vone_kb_collections(id),
    name        TEXT NOT NULL,
    sort_order  INTEGER NOT NULL DEFAULT 0,
    created_by  TEXT NOT NULL,
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at  DATETIME,
    CHECK (parent_id IS NULL OR parent_id <> id)
);
CREATE UNIQUE INDEX IF NOT EXISTS ux_vone_kb_collections_active_sibling
    ON vone_kb_collections(tenant_id, IFNULL(parent_id, ''), name) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS ix_vone_kb_collections_tenant_parent
    ON vone_kb_collections(tenant_id, parent_id) WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS vone_kb_collection_bindings (
    kb_id          TEXT PRIMARY KEY,
    tenant_id      INTEGER NOT NULL,
    collection_id  TEXT NOT NULL REFERENCES vone_kb_collections(id),
    updated_by     TEXT NOT NULL,
    created_at     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS ix_vone_kb_collection_bindings_collection
    ON vone_kb_collection_bindings(tenant_id, collection_id);

INSERT OR IGNORE INTO vone_kb_memberships
    (tenant_id, kb_id, user_id, role, granted_by, created_at, updated_at)
SELECT kb.tenant_id,
       kb.id,
       COALESCE(
           (SELECT tm.user_id FROM tenant_members tm
             WHERE tm.tenant_id = kb.tenant_id AND tm.user_id = kb.creator_id
               AND tm.status = 'active' AND tm.deleted_at IS NULL LIMIT 1),
           (SELECT tm.user_id FROM tenant_members tm
             WHERE tm.tenant_id = kb.tenant_id AND tm.role = 'owner'
               AND tm.status = 'active' AND tm.deleted_at IS NULL
             ORDER BY tm.joined_at ASC, tm.id ASC LIMIT 1)
       ),
       'owner',
       COALESCE(
           (SELECT tm.user_id FROM tenant_members tm
             WHERE tm.tenant_id = kb.tenant_id AND tm.user_id = kb.creator_id
               AND tm.status = 'active' AND tm.deleted_at IS NULL LIMIT 1),
           (SELECT tm.user_id FROM tenant_members tm
             WHERE tm.tenant_id = kb.tenant_id AND tm.role = 'owner'
               AND tm.status = 'active' AND tm.deleted_at IS NULL
             ORDER BY tm.joined_at ASC, tm.id ASC LIMIT 1)
       ),
       CURRENT_TIMESTAMP,
       CURRENT_TIMESTAMP
  FROM knowledge_bases kb
 WHERE kb.deleted_at IS NULL
   AND COALESCE(
       (SELECT tm.user_id FROM tenant_members tm
         WHERE tm.tenant_id = kb.tenant_id AND tm.user_id = kb.creator_id
           AND tm.status = 'active' AND tm.deleted_at IS NULL LIMIT 1),
       (SELECT tm.user_id FROM tenant_members tm
         WHERE tm.tenant_id = kb.tenant_id AND tm.role = 'owner'
           AND tm.status = 'active' AND tm.deleted_at IS NULL
         ORDER BY tm.joined_at ASC, tm.id ASC LIMIT 1)
   ) IS NOT NULL;
