-- VONE-0.7.2.1 read-only PostgreSQL reconciliation.
-- Run after 000001_kb_acl_and_collections.up.sql and before enforce.

-- Every KB must have at least one effective Owner membership.
SELECT kb.id, kb.tenant_id
FROM knowledge_bases kb
LEFT JOIN vone_kb_memberships m
  ON m.kb_id = kb.id AND m.tenant_id = kb.tenant_id
 AND m.role = 'owner' AND m.deleted_at IS NULL
GROUP BY kb.id, kb.tenant_id
HAVING COUNT(m.id) = 0;

-- Memberships must reference an existing KB and active workspace membership.
SELECT m.*
FROM vone_kb_memberships m
LEFT JOIN knowledge_bases kb ON kb.id = m.kb_id AND kb.tenant_id = m.tenant_id
LEFT JOIN tenant_members tm ON tm.tenant_id = m.tenant_id AND tm.user_id = m.user_id
WHERE m.deleted_at IS NULL AND (kb.id IS NULL OR tm.user_id IS NULL);

-- Persisted capabilities must be from the frozen 17-code set.
SELECT c.* FROM vone_kb_member_capabilities c
WHERE c.capability NOT IN (
  'kb.ai.query','kb.metadata.read','kb.documents.list','kb.document.preview',
  'kb.chunk.preview','kb.document.download','kb.document.upload','kb.document.edit',
  'kb.document.delete','kb.document.reparse','kb.chunk.edit','kb.chunk.delete',
  'kb.folder.manage','kb.settings.edit','kb.delete','kb.members.manage','kb.owners.manage'
);

-- Collection bindings must point to same-workspace KBs and collections.
SELECT b.*
FROM vone_kb_collection_bindings b
LEFT JOIN knowledge_bases kb ON kb.id = b.kb_id AND kb.tenant_id = b.tenant_id
LEFT JOIN vone_kb_collections c ON c.id = b.collection_id AND c.tenant_id = b.tenant_id AND c.deleted_at IS NULL
WHERE kb.id IS NULL OR c.id IS NULL;

-- Inventory only: kb_shares is retained and must not be rewritten by this release.
SELECT COUNT(*) AS retained_kb_share_count FROM kb_shares;
