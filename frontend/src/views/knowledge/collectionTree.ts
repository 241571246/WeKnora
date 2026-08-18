import type { KBCollection, KBCollectionBinding } from '@/api/knowledge-base'

export type FlatKBCollection = KBCollection & { depth: number }

export function flattenCollections(collections: KBCollection[]): FlatKBCollection[] {
  const children = new Map<string, KBCollection[]>()
  for (const item of collections) {
    const key = item.parent_id || ''
    children.set(key, [...(children.get(key) || []), item])
  }
  const result: FlatKBCollection[] = []
  const seen = new Set<string>()
  const visit = (parent: string, depth: number) => {
    for (const item of children.get(parent) || []) {
      if (seen.has(item.id)) continue
      seen.add(item.id)
      result.push({ ...item, depth })
      visit(item.id, depth + 1)
    }
  }
  visit('', 0)
  return result
}

export function collectionKnowledgeBaseIDs(
  collectionID: string,
  collections: KBCollection[],
  bindings: KBCollectionBinding[],
): string[] {
  const descendants = new Set<string>([collectionID])
  let changed = true
  while (changed) {
    changed = false
    for (const item of collections) {
      if (item.parent_id && descendants.has(item.parent_id) && !descendants.has(item.id)) {
        descendants.add(item.id)
        changed = true
      }
    }
  }
  return bindings.filter(item => descendants.has(item.collection_id)).map(item => item.kb_id)
}

export function unclassifiedKnowledgeBaseIDs(allIDs: string[], bindings: KBCollectionBinding[]): string[] {
  const bound = new Set(bindings.map(item => item.kb_id))
  return allIDs.filter(id => !bound.has(id))
}

export function filterRowsByKnowledgeBaseIDs<T>(
  rows: T[],
  knowledgeBaseIDs: string[] | null,
  getKnowledgeBaseID: (row: T) => string = row => String((row as { id?: unknown }).id ?? ''),
): T[] {
  if (knowledgeBaseIDs === null) return rows
  const allowed = new Set(knowledgeBaseIDs.map(String))
  return rows.filter(row => allowed.has(String(getKnowledgeBaseID(row))))
}
