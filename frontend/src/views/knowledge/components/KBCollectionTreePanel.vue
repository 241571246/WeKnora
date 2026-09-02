<template>
  <section class="collection-panel" :class="{ 'collection-panel--collapsed': panelCollapsed }">
    <div class="collection-toolbar">
      <div v-if="!panelCollapsed" class="collection-heading">
        <strong>{{ t('kbAcl.collections.title') }}</strong>
        <span>{{ t('kbAcl.collections.description') }}</span>
      </div>
      <div class="collection-toolbar-actions">
        <t-tooltip v-if="canManage && !panelCollapsed" :content="t('kbAcl.collections.createRoot')" placement="bottom">
          <t-button size="small" variant="text" shape="square" @click="createCollection()">
            <template #icon><t-icon name="folder-add" /></template>
          </t-button>
        </t-tooltip>
        <t-tooltip :content="panelCollapsed ? t('knowledgeBase.folderTree.expand') : t('knowledgeBase.folderTree.collapse')"
          placement="bottom">
          <t-button size="small" variant="text" shape="square"
            :aria-label="panelCollapsed ? t('knowledgeBase.folderTree.expand') : t('knowledgeBase.folderTree.collapse')"
            @click="panelCollapsed = !panelCollapsed">
            <template #icon><t-icon :name="panelCollapsed ? 'chevron-right' : 'chevron-left'" /></template>
          </t-button>
        </t-tooltip>
      </div>
    </div>

    <button v-if="panelCollapsed" type="button" class="collection-collapsed-entry"
      :class="{ active: selectedCollectionId !== '__all' }" :title="selectedCollectionLabel"
      @click="panelCollapsed = false">
      <t-icon name="tree-square-dot" size="20px" />
      <small v-if="selectedCollectionId !== '__all'">{{ selectedKnowledgeBaseCount }}</small>
    </button>

    <template v-else>
      <t-alert v-if="errorMessage" theme="error" :message="errorMessage" close="false">
        <template #operation>
          <t-button size="small" variant="text" @click="load">{{ t('common.retry') }}</t-button>
        </template>
      </t-alert>

      <t-loading class="collection-loading" :loading="loading">
        <div class="collection-body">
          <button type="button" class="collection-node collection-node--root"
            :class="{ active: selectedCollectionId === '__all' }" @click="select('__all')">
            <span class="collection-node-main">
              <t-icon name="layers" />
              <span class="collection-node-label">{{ t('kbAcl.collections.all') }}</span>
            </span>
            <small>{{ knowledgeBases.length }}</small>
          </button>

          <div class="collection-tree" role="tree">
            <button type="button" class="collection-node collection-node--unclassified"
              :class="{ active: selectedCollectionId === '__unclassified' }" role="treeitem"
              @click="select('__unclassified')">
              <span class="collection-toggle-spacer" />
              <t-icon name="folder" class="collection-folder-icon" />
              <span class="collection-node-label">{{ t('kbAcl.collections.unclassified') }}</span>
              <small>{{ unclassifiedIds.length }}</small>
            </button>

            <div v-for="item in visibleCollections" :key="item.id" class="collection-tree-row"
              :class="{ active: selectedCollectionId === item.id }" :style="{ paddingLeft: `${8 + item.depth * 18}px` }"
              role="treeitem" :aria-level="item.depth + 1" :aria-selected="selectedCollectionId === item.id"
              :aria-expanded="collectionIDsWithChildren.has(item.id) ? expandedCollectionIds.has(item.id) : undefined"
              tabindex="0" @click="select(item.id)" @keydown.enter.prevent="select(item.id)"
              @keydown.space.prevent="select(item.id)">
              <button v-if="collectionIDsWithChildren.has(item.id)" type="button" class="collection-toggle"
                :aria-label="expandedCollectionIds.has(item.id) ? t('knowledgeBase.folderTree.collapseFolder') : t('knowledgeBase.folderTree.expandFolder')"
                @click.stop="toggleCollection(item.id)">
                <t-icon :name="expandedCollectionIds.has(item.id) ? 'chevron-down' : 'chevron-right'" size="14px" />
              </button>
              <span v-else class="collection-toggle-spacer" />
              <t-icon :name="expandedCollectionIds.has(item.id) ? 'folder-open' : 'folder'"
                class="collection-folder-icon" />
              <span class="collection-node-label" :title="item.name">{{ item.name }}</span>
              <small>{{ idsForCollection(item.id).length }}</small>

              <t-popup v-if="canManage" trigger="click" placement="bottom-left" destroy-on-close
                overlayClassName="collection-action-popup">
                <button type="button" class="collection-node-action"
                  :aria-label="`${item.name}: ${t('kbAcl.collections.rename')}/${t('kbAcl.collections.move')}/${t('kbAcl.collections.delete')}`"
                  @click.stop>
                  <t-icon name="ellipsis" size="16px" />
                </button>
                <template #content>
                  <div class="collection-action-menu" @click.stop>
                    <button type="button" @click="createCollection(item.id)">
                      <t-icon name="folder-add" />
                      <span>{{ t('kbAcl.collections.createChild') }}</span>
                    </button>
                    <button type="button" @click="renameCollection(item)">
                      <t-icon name="edit-1" />
                      <span>{{ t('kbAcl.collections.rename') }}</span>
                    </button>
                    <button type="button" @click="moveCollection(item)">
                      <t-icon name="move" />
                      <span>{{ t('kbAcl.collections.move') }}</span>
                    </button>
                    <button type="button" @click="sortCollection(item, -1)">
                      <t-icon name="arrow-up" />
                      <span>{{ t('kbAcl.collections.moveUp') }}</span>
                    </button>
                    <button type="button" @click="sortCollection(item, 1)">
                      <t-icon name="arrow-down" />
                      <span>{{ t('kbAcl.collections.moveDown') }}</span>
                    </button>
                    <button type="button" class="danger" @click="removeCollection(item)">
                      <t-icon name="delete" />
                      <span>{{ t('kbAcl.collections.delete') }}</span>
                    </button>
                  </div>
                </template>
              </t-popup>
            </div>
          </div>

          <t-empty v-if="!loading && !errorMessage && !collections.length && !knowledgeBases.length"
            :description="t('kbAcl.collections.empty')" />
        </div>
      </t-loading>

      <div v-if="canManage && selectedCollectionId && selectedCollectionId !== '__all'" class="collection-binding">
        <t-select v-model="bindingKBID" size="small" :placeholder="t('kbAcl.collections.bindPlaceholder')"
          filterable clearable>
          <t-option v-for="kb in knowledgeBases" :key="kb.id" :value="kb.id" :label="kb.name" />
        </t-select>
        <t-tooltip :content="t('kbAcl.collections.bind')" placement="top">
          <t-button size="small" shape="square" :disabled="!bindingKBID" @click="bindKnowledgeBase">
            <template #icon><t-icon name="check" /></template>
          </t-button>
        </t-tooltip>
      </div>
    </template>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { MessagePlugin } from 'tdesign-vue-next'
import {
  createKBCollection,
  deleteKBCollection,
  getKBCollectionTree,
  setKBCollection,
  updateKBCollection,
  type KBCollection,
  type KBCollectionBinding,
} from '@/api/knowledge-base'
import { useAuthStore } from '@/stores/auth'
import { useI18n } from 'vue-i18n'
import {
  collectionKnowledgeBaseIDs,
  flattenVisibleCollections,
  unclassifiedKnowledgeBaseIDs,
} from '../collectionTree'

const props = defineProps<{ knowledgeBases: Array<{ id: string; name: string }> }>()
const emit = defineEmits<{ (e: 'filter', ids: string[] | null): void }>()
const authStore = useAuthStore()
const { t } = useI18n()
const collections = ref<KBCollection[]>([])
const bindings = ref<KBCollectionBinding[]>([])
const selectedCollectionId = ref('__all')
const bindingKBID = ref('')
const loading = ref(false)
const errorMessage = ref('')
const panelCollapsed = ref(false)
const expandedCollectionIds = ref<Set<string>>(new Set())
const expansionInitialized = ref(false)

const canManage = computed(() => authStore.hasRole('owner'))
const selectedCollection = computed(() => collections.value.find(item => item.id === selectedCollectionId.value))
const unclassifiedIds = computed(() => unclassifiedKnowledgeBaseIDs(props.knowledgeBases.map(item => item.id), bindings.value))
const collectionIDsWithChildren = computed(() => new Set(
  collections.value.map(item => item.parent_id).filter((id): id is string => Boolean(id)),
))
const visibleCollections = computed(() => flattenVisibleCollections(collections.value, expandedCollectionIds.value))
const selectedKnowledgeBaseCount = computed(() => {
  if (selectedCollectionId.value === '__all') return props.knowledgeBases.length
  if (selectedCollectionId.value === '__unclassified') return unclassifiedIds.value.length
  return idsForCollection(selectedCollectionId.value).length
})
const selectedCollectionLabel = computed(() => {
  if (selectedCollectionId.value === '__all') return t('kbAcl.collections.all')
  if (selectedCollectionId.value === '__unclassified') return t('kbAcl.collections.unclassified')
  return selectedCollection.value?.name || t('kbAcl.collections.title')
})

function idsForCollection(id: string) {
  return collectionKnowledgeBaseIDs(id, collections.value, bindings.value)
}

function select(id: string) {
  selectedCollectionId.value = id
  if (id === '__all') emit('filter', null)
  else if (id === '__unclassified') emit('filter', unclassifiedIds.value)
  else emit('filter', idsForCollection(id))
}

function toggleCollection(id: string) {
  const next = new Set(expandedCollectionIds.value)
  if (next.has(id)) next.delete(id)
  else next.add(id)
  expandedCollectionIds.value = next
}

async function load() {
  loading.value = true
  errorMessage.value = ''
  try {
    const response: any = await getKBCollectionTree()
    collections.value = response?.data?.collections || []
    bindings.value = response?.data?.bindings || []
    const validIds = new Set(collections.value.map(item => item.id))
    expandedCollectionIds.value = expansionInitialized.value
      ? new Set([...expandedCollectionIds.value].filter(id => validIds.has(id)))
      : validIds
    expansionInitialized.value = true
    if (!selectedCollectionId.value.startsWith('__') && !validIds.has(selectedCollectionId.value)) {
      selectedCollectionId.value = '__all'
    }
    select(selectedCollectionId.value)
  } catch (error: any) {
    errorMessage.value = error?.message || t('kbAcl.collections.loadFailed')
    MessagePlugin.error(errorMessage.value)
  } finally {
    loading.value = false
  }
}

async function createCollection(parentId?: string) {
  const name = window.prompt(parentId ? t('kbAcl.collections.childNamePrompt') : t('kbAcl.collections.namePrompt'))?.trim()
  if (!name) return
  try {
    await createKBCollection({ parent_id: parentId, name })
    if (parentId) expandedCollectionIds.value = new Set([...expandedCollectionIds.value, parentId])
    await load()
    MessagePlugin.success(t('kbAcl.collections.created'))
  } catch (error: any) {
    MessagePlugin.error(error?.message || t('kbAcl.collections.createFailed'))
  }
}

async function removeCollection(collection: KBCollection) {
  if (!window.confirm(t('kbAcl.collections.deleteConfirm'))) return
  try {
    await deleteKBCollection(collection.id)
    if (selectedCollectionId.value === collection.id) selectedCollectionId.value = '__all'
    await load()
    MessagePlugin.success(t('kbAcl.collections.deleted'))
  } catch (error: any) {
    MessagePlugin.error(error?.message || t('kbAcl.collections.deleteFailed'))
  }
}

async function renameCollection(collection: KBCollection) {
  const name = window.prompt(t('kbAcl.collections.renamePrompt'), collection.name)?.trim()
  if (!name) return
  try {
    await updateKBCollection(collection.id, {
      parent_id: collection.parent_id,
      name,
      sort_order: collection.sort_order,
    })
    await load()
    MessagePlugin.success(t('kbAcl.collections.renamed'))
  } catch (error: any) {
    MessagePlugin.error(error?.message || t('kbAcl.collections.renameFailed'))
  }
}

async function moveCollection(collection: KBCollection) {
  const choices = collections.value
    .filter(item => item.id !== collection.id)
    .map(item => `${item.id} = ${item.name}`)
    .join('\n')
  const parentId = window.prompt(`${t('kbAcl.collections.movePrompt')}\n${choices}`, collection.parent_id || '')
  if (parentId === null) return
  try {
    await updateKBCollection(collection.id, {
      parent_id: parentId.trim() || undefined,
      name: collection.name,
      sort_order: collection.sort_order,
    })
    await load()
    MessagePlugin.success(t('kbAcl.collections.moved'))
  } catch (error: any) {
    MessagePlugin.error(error?.message || t('kbAcl.collections.moveFailed'))
  }
}

async function sortCollection(collection: KBCollection, offset: number) {
  try {
    await updateKBCollection(collection.id, {
      parent_id: collection.parent_id,
      name: collection.name,
      sort_order: collection.sort_order + offset,
    })
    await load()
  } catch (error: any) {
    MessagePlugin.error(error?.message || '排序失败')
  }
}

async function bindKnowledgeBase() {
  const targetCollectionID = selectedCollectionId.value === '__unclassified' ? '' : selectedCollectionId.value
  try {
    await setKBCollection(bindingKBID.value, targetCollectionID)
    bindingKBID.value = ''
    await load()
    MessagePlugin.success(t('kbAcl.collections.bound'))
  } catch (error: any) {
    MessagePlugin.error(error?.message || t('kbAcl.collections.bindFailed'))
  }
}

onMounted(load)
defineExpose({ reload: load })
</script>

<style scoped>
.collection-panel {
  flex: 0 0 260px;
  width: 260px;
  min-width: 0;
  height: 100%;
  min-height: 0;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  border: 1px solid var(--td-component-border);
  border-radius: 10px;
  background: var(--td-bg-color-container);
  transition: flex-basis 0.2s ease, width 0.2s ease;
}

.collection-panel--collapsed {
  flex-basis: 48px;
  width: 48px;
}

.collection-toolbar {
  min-height: 52px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  padding: 8px 10px 8px 14px;
  border-bottom: 1px solid var(--td-component-stroke);
  flex-shrink: 0;
}

.collection-panel--collapsed .collection-toolbar {
  justify-content: center;
  padding: 8px 6px;
}

.collection-heading {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 1px;
}

.collection-heading strong {
  color: var(--td-text-color-primary);
  font-size: 15px;
  line-height: 22px;
}

.collection-heading span {
  overflow: hidden;
  color: var(--td-text-color-secondary);
  font-size: 11px;
  line-height: 16px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.collection-toolbar-actions {
  display: flex;
  align-items: center;
  flex-shrink: 0;
}

.collection-collapsed-entry {
  position: relative;
  width: 36px;
  height: 40px;
  margin: 8px auto;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 0;
  border-radius: 8px;
  background: transparent;
  color: var(--td-text-color-secondary);
  cursor: pointer;
}

.collection-collapsed-entry:hover,
.collection-collapsed-entry.active {
  background: var(--td-brand-color-light);
  color: var(--td-brand-color);
}

.collection-collapsed-entry small {
  position: absolute;
  top: 2px;
  right: 1px;
  min-width: 14px;
  height: 14px;
  padding: 0 3px;
  border-radius: 7px;
  background: var(--td-brand-color);
  color: var(--td-text-color-anti);
  font-size: 9px;
  line-height: 14px;
}

.collection-loading {
  flex: 1;
  min-height: 0;
}

.collection-loading :deep(.t-loading__parent) {
  height: 100%;
}

.collection-body {
  height: 100%;
  overflow-y: auto;
  padding: 10px 8px;
}

.collection-tree {
  margin-top: 4px;
}

.collection-node,
.collection-tree-row {
  width: 100%;
  min-height: 36px;
  box-sizing: border-box;
  display: flex;
  align-items: center;
  gap: 6px;
  border: 0;
  border-radius: 7px;
  background: transparent;
  color: var(--td-text-color-primary);
  font: inherit;
  text-align: left;
  cursor: pointer;
}

.collection-node {
  padding: 7px 9px;
}

.collection-tree-row {
  padding: 5px 6px 5px 8px;
}

.collection-node:hover,
.collection-tree-row:hover,
.collection-node.active,
.collection-tree-row.active {
  background: var(--td-brand-color-light);
}

.collection-node--root {
  font-weight: 600;
}

.collection-node--unclassified {
  padding-left: 8px;
}

.collection-node-main {
  min-width: 0;
  display: flex;
  align-items: center;
  gap: 7px;
  flex: 1;
}

.collection-toggle,
.collection-node-action {
  width: 22px;
  height: 22px;
  padding: 0;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  border: 0;
  border-radius: 5px;
  background: transparent;
  color: var(--td-text-color-placeholder);
  cursor: pointer;
}

.collection-toggle:hover,
.collection-node-action:hover {
  background: var(--td-bg-color-container-hover);
  color: var(--td-text-color-primary);
}

.collection-node-action {
  opacity: 0;
}

.collection-tree-row:hover .collection-node-action,
.collection-tree-row:focus-within .collection-node-action {
  opacity: 1;
}

.collection-toggle-spacer {
  width: 22px;
  flex: 0 0 22px;
}

.collection-folder-icon {
  flex-shrink: 0;
  color: var(--td-brand-color);
}

.collection-node-label {
  min-width: 0;
  flex: 1;
  overflow: hidden;
  font-size: 13px;
  line-height: 20px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.collection-node small,
.collection-tree-row small {
  min-width: 20px;
  flex-shrink: 0;
  color: var(--td-text-color-placeholder);
  font-size: 11px;
  text-align: right;
}

.collection-binding {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 10px;
  border-top: 1px solid var(--td-component-stroke);
  flex-shrink: 0;
}

.collection-binding .t-select {
  min-width: 0;
  flex: 1;
}

.collection-action-menu {
  min-width: 148px;
  padding: 4px;
}

.collection-action-menu button {
  width: 100%;
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 7px 10px;
  border: 0;
  border-radius: 5px;
  background: transparent;
  color: var(--td-text-color-primary);
  font-size: 13px;
  text-align: left;
  cursor: pointer;
}

.collection-action-menu button:hover {
  background: var(--td-bg-color-container-hover);
}

.collection-action-menu button.danger {
  color: var(--td-error-color);
}

@media (max-width: 1100px) {
  .collection-panel {
    flex-basis: 220px;
    width: 220px;
  }

  .collection-panel--collapsed {
    flex-basis: 48px;
    width: 48px;
  }

  .collection-heading span {
    display: none;
  }
}
</style>

<style>
.collection-action-popup .t-popup__content {
  padding: 0;
}
</style>
