import assert from 'node:assert/strict'
import test from 'node:test'
import { canMutateMemberRole, projectKBActionPermissions } from './kbAclPresentation'

test('AI-only access never exposes content management actions', () => {
  const result = projectKBActionPermissions(['kb.ai.query', 'kb.metadata.read'])
  assert.equal(result.aiOnly, true)
  assert.equal(result.canListDocuments, false)
  assert.equal(result.canPreviewDocument, false)
  assert.equal(result.canPreviewChunks, false)
  assert.equal(result.canDownloadDocument, false)
  assert.equal(result.canEditDocument, false)
  assert.equal(result.canDeleteDocument, false)
  assert.equal(result.canDeleteChunks, false)
})

test('document, chunk, and whole-KB deletion stay independent', () => {
  const documentOnly = projectKBActionPermissions(['kb.document.delete'])
  assert.equal(documentOnly.canDeleteDocument, true)
  assert.equal(documentOnly.canDeleteChunks, false)
  assert.equal(documentOnly.canDeleteKnowledgeBase, false)

  const chunkOnly = projectKBActionPermissions(['kb.chunk.delete'])
  assert.equal(chunkOnly.canDeleteDocument, false)
  assert.equal(chunkOnly.canDeleteChunks, true)
  assert.equal(chunkOnly.canDeleteKnowledgeBase, false)
})

test('owner mutations require owners.manage independently of members.manage', () => {
  assert.equal(canMutateMemberRole('editor', 'custom', false), true)
  assert.equal(canMutateMemberRole('editor', 'owner', false), false)
  assert.equal(canMutateMemberRole('owner', 'editor', false), false)
  assert.equal(canMutateMemberRole('owner', 'owner', true), true)
})
