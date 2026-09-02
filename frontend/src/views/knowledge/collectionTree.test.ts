import assert from 'node:assert/strict'
import test from 'node:test'
import {
  collectionKnowledgeBaseIDs,
  filterRowsByKnowledgeBaseIDs,
  flattenCollections,
  flattenVisibleCollections,
  unclassifiedKnowledgeBaseIDs,
} from './collectionTree'

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

test('visible tree hides descendants of collapsed collections', () => {
  assert.deepEqual(
    flattenVisibleCollections(collections, new Set()).map(item => [item.id, item.depth]),
    [['root', 0]],
  )
})

test('visible tree expands only the selected branches', () => {
  const rows = [
    ...collections,
    { id: 'sibling', tenant_id: 1, name: 'Sibling', sort_order: 1 },
    { id: 'sibling-child', tenant_id: 1, parent_id: 'sibling', name: 'Sibling Child', sort_order: 0 },
  ]
  assert.deepEqual(
    flattenVisibleCollections(rows, new Set(['root'])).map(item => [item.id, item.depth]),
    [['root', 0], ['child', 1], ['sibling', 0]],
  )
})

test('parent selection includes descendant knowledge bases', () => {
  assert.deepEqual(collectionKnowledgeBaseIDs('root', collections, bindings), ['kb-root', 'kb-child'])
})

test('unclassified excludes every bound knowledge base', () => {
  assert.deepEqual(unclassifiedKnowledgeBaseIDs(['kb-root', 'kb-free'], bindings), ['kb-free'])
})

test('null collection filter preserves every knowledge base', () => {
  const rows = [{ id: 'kb-root' }, { id: 'kb-child' }]
  assert.equal(filterRowsByKnowledgeBaseIDs(rows, null), rows)
})

test('collection filter keeps only selected knowledge bases', () => {
  const rows = [{ id: 'kb-root' }, { id: 'kb-child' }, { id: 'kb-free' }]
  assert.deepEqual(filterRowsByKnowledgeBaseIDs(rows, ['kb-child']), [{ id: 'kb-child' }])
})

test('empty collection filter returns an empty list', () => {
  assert.deepEqual(filterRowsByKnowledgeBaseIDs([{ id: 'kb-root' }], []), [])
})

test('collection filter supports nested organization rows', () => {
  const rows = [
    { knowledge_base: { id: 'kb-root' } },
    { knowledge_base: { id: 'kb-child' } },
  ]
  assert.deepEqual(
    filterRowsByKnowledgeBaseIDs(rows, ['kb-child'], row => row.knowledge_base.id),
    [{ knowledge_base: { id: 'kb-child' } }],
  )
})
