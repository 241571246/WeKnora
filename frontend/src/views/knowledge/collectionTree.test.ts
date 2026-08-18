import assert from 'node:assert/strict'
import test from 'node:test'
import { collectionKnowledgeBaseIDs, flattenCollections, unclassifiedKnowledgeBaseIDs } from './collectionTree'

const collections = [
  { id: 'root', tenant_id: 1, name: 'Root', sort_order: 0 },
  { id: 'child', tenant_id: 1, parent_id: 'root', name: 'Child', sort_order: 0 },
]
const bindings = [
  { kb_id: 'kb-root', tenant_id: 1, collection_id: 'root' },
  { kb_id: 'kb-child', tenant_id: 1, collection_id: 'child' },
]

test('flattens nested collections with depth', () => {
  assert.deepEqual(flattenCollections(collections).map(item => [item.id, item.depth]), [['root', 0], ['child', 1]])
})

test('parent selection includes descendant knowledge bases', () => {
  assert.deepEqual(collectionKnowledgeBaseIDs('root', collections, bindings), ['kb-root', 'kb-child'])
})

test('unclassified excludes every bound knowledge base', () => {
  assert.deepEqual(unclassifiedKnowledgeBaseIDs(['kb-root', 'kb-free'], bindings), ['kb-free'])
})
