import type { KBCapability } from '@/api/knowledge-base'

export interface KBActionPermissions {
  aiOnly: boolean
  canQueryAI: boolean
  canReadMetadata: boolean
  canListDocuments: boolean
  canPreviewDocument: boolean
  canPreviewChunks: boolean
  canDownloadDocument: boolean
  canUploadDocument: boolean
  canEditDocument: boolean
  canDeleteDocument: boolean
  canReparseDocument: boolean
  canEditChunks: boolean
  canDeleteChunks: boolean
  canManageFolders: boolean
  canEditSettings: boolean
  canDeleteKnowledgeBase: boolean
  canManageMembers: boolean
  canManageOwners: boolean
}

export function projectKBActionPermissions(capabilities: readonly KBCapability[]): KBActionPermissions {
  const granted = new Set(capabilities)
  const has = (capability: KBCapability) => granted.has(capability)
  const canListDocuments = has('kb.documents.list')
  return {
    aiOnly: has('kb.ai.query') && !canListDocuments,
    canQueryAI: has('kb.ai.query'),
    canReadMetadata: has('kb.metadata.read'),
    canListDocuments,
    canPreviewDocument: has('kb.document.preview'),
    canPreviewChunks: has('kb.chunk.preview'),
    canDownloadDocument: has('kb.document.download'),
    canUploadDocument: has('kb.document.upload'),
    canEditDocument: has('kb.document.edit'),
    canDeleteDocument: has('kb.document.delete'),
    canReparseDocument: has('kb.document.reparse'),
    canEditChunks: has('kb.chunk.edit'),
    canDeleteChunks: has('kb.chunk.delete'),
    canManageFolders: has('kb.folder.manage'),
    canEditSettings: has('kb.settings.edit'),
    canDeleteKnowledgeBase: has('kb.delete'),
    canManageMembers: has('kb.members.manage'),
    canManageOwners: has('kb.owners.manage'),
  }
}

export function canMutateMemberRole(
  currentRole: string | undefined,
  nextRole: string | undefined,
  canManageOwners: boolean,
): boolean {
  if (currentRole === 'owner' || nextRole === 'owner') return canManageOwners
  return true
}
