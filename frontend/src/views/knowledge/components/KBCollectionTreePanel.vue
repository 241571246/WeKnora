<template>
  <section class="collection-panel">
    <div class="collection-toolbar">
      <div><strong>{{ t('kbAcl.collections.title') }}</strong><span>{{ t('kbAcl.collections.description') }}</span></div>
      <div v-if="canManage" class="collection-actions">
        <t-button size="small" variant="outline" @click="createCollection()">{{ t('kbAcl.collections.createRoot') }}</t-button>
        <t-button size="small" variant="text" :disabled="!selectedCollectionId" @click="createCollection(selectedCollectionId)">{{ t('kbAcl.collections.createChild') }}</t-button>
        <t-button size="small" variant="text" :disabled="!selectedCollection" @click="renameCollection">{{ t('kbAcl.collections.rename') }}</t-button>
        <t-button size="small" variant="text" :disabled="!selectedCollection" @click="moveCollection">{{ t('kbAcl.collections.move') }}</t-button>
        <t-button size="small" variant="text" :disabled="!selectedCollection" @click="sortCollection(-1)">{{ t('kbAcl.collections.moveUp') }}</t-button>
        <t-button size="small" variant="text" :disabled="!selectedCollection" @click="sortCollection(1)">{{ t('kbAcl.collections.moveDown') }}</t-button>
        <t-button size="small" theme="danger" variant="text" :disabled="!selectedCollectionId" @click="removeCollection">{{ t('kbAcl.collections.delete') }}</t-button>
      </div>
    </div>
    <t-alert v-if="errorMessage" theme="error" :message="errorMessage" close="false">
      <template #operation><t-button size="small" variant="text" @click="load">{{ t('common.retry') }}</t-button></template>
    </t-alert>
    <t-loading :loading="loading">
    <div class="collection-body">
      <button class="collection-node" :class="{ active: selectedCollectionId === '__all' }" @click="select('__all')">
        <span>{{ t('kbAcl.collections.all') }}</span><small>{{ knowledgeBases.length }}</small>
      </button>
      <button class="collection-node" :class="{ active: selectedCollectionId === '__unclassified' }" @click="select('__unclassified')">
        <span>{{ t('kbAcl.collections.unclassified') }}</span><small>{{ unclassifiedIds.length }}</small>
      </button>
      <button v-for="item in flatCollections" :key="item.id" class="collection-node"
        :class="{ active: selectedCollectionId === item.id }" :style="{ paddingLeft: `${16 + item.depth * 20}px` }"
        @click="select(item.id)">
        <span><t-icon name="folder" /> {{ item.name }}</span><small>{{ idsForCollection(item.id).length }}</small>
      </button>
      <t-empty v-if="!loading && !errorMessage && !flatCollections.length && !knowledgeBases.length"
        :description="t('kbAcl.collections.empty')" />
    </div>
    </t-loading>
    <div v-if="canManage && selectedCollectionId && selectedCollectionId !== '__all'" class="collection-binding">
      <t-select v-model="bindingKBID" :placeholder="t('kbAcl.collections.bindPlaceholder')" filterable clearable>
        <t-option v-for="kb in knowledgeBases" :key="kb.id" :value="kb.id" :label="kb.name" />
      </t-select>
      <t-button size="small" :disabled="!bindingKBID" @click="bindKnowledgeBase">{{ t('kbAcl.collections.bind') }}</t-button>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { MessagePlugin } from 'tdesign-vue-next'
import { createKBCollection, deleteKBCollection, getKBCollectionTree, setKBCollection, updateKBCollection, type KBCollection, type KBCollectionBinding } from '@/api/knowledge-base'
import { useAuthStore } from '@/stores/auth'
import { useI18n } from 'vue-i18n'
import { collectionKnowledgeBaseIDs, flattenCollections, unclassifiedKnowledgeBaseIDs } from '../collectionTree'

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
const canManage = computed(() => authStore.hasRole('owner'))
const selectedCollection = computed(() => collections.value.find(item => item.id === selectedCollectionId.value))
const unclassifiedIds = computed(() => unclassifiedKnowledgeBaseIDs(props.knowledgeBases.map(item => item.id), bindings.value))
const flatCollections = computed(() => flattenCollections(collections.value))
function idsForCollection(id: string) {
  return collectionKnowledgeBaseIDs(id, collections.value, bindings.value)
}
function select(id: string) {
  selectedCollectionId.value = id
  if (id === '__all') emit('filter', null)
  else if (id === '__unclassified') emit('filter', unclassifiedIds.value)
  else emit('filter', idsForCollection(id))
}
async function load() {
  loading.value = true
  errorMessage.value = ''
  try {
    const response: any = await getKBCollectionTree()
    collections.value = response?.data?.collections || []
    bindings.value = response?.data?.bindings || []
    select(selectedCollectionId.value)
  } catch (error: any) {
    errorMessage.value = error?.message || t('kbAcl.collections.loadFailed')
    MessagePlugin.error(errorMessage.value)
  } finally { loading.value = false }
}
async function createCollection(parentId?: string) {
  const name = window.prompt(parentId ? t('kbAcl.collections.childNamePrompt') : t('kbAcl.collections.namePrompt'))?.trim()
  if (!name) return
  try { await createKBCollection({ parent_id: parentId, name }); await load(); MessagePlugin.success(t('kbAcl.collections.created')) }
  catch (error: any) { MessagePlugin.error(error?.message || t('kbAcl.collections.createFailed')) }
}
async function removeCollection() {
  if (!selectedCollectionId.value || !window.confirm(t('kbAcl.collections.deleteConfirm'))) return
  try { await deleteKBCollection(selectedCollectionId.value); selectedCollectionId.value = '__all'; await load(); MessagePlugin.success(t('kbAcl.collections.deleted')) }
  catch (error: any) { MessagePlugin.error(error?.message || t('kbAcl.collections.deleteFailed')) }
}
async function renameCollection() {
  const current = selectedCollection.value
  if (!current) return
  const name = window.prompt(t('kbAcl.collections.renamePrompt'), current.name)?.trim()
  if (!name) return
  try { await updateKBCollection(current.id, { parent_id: current.parent_id, name, sort_order: current.sort_order }); await load(); MessagePlugin.success(t('kbAcl.collections.renamed')) }
  catch (error: any) { MessagePlugin.error(error?.message || t('kbAcl.collections.renameFailed')) }
}
async function moveCollection() {
  const current = selectedCollection.value
  if (!current) return
  const choices = collections.value.filter(item => item.id !== current.id).map(item => `${item.id} = ${item.name}`).join('\n')
  const parentId = window.prompt(`${t('kbAcl.collections.movePrompt')}\n${choices}`, current.parent_id || '')
  if (parentId === null) return
  try { await updateKBCollection(current.id, { parent_id: parentId.trim() || undefined, name: current.name, sort_order: current.sort_order }); await load(); MessagePlugin.success(t('kbAcl.collections.moved')) }
  catch (error: any) { MessagePlugin.error(error?.message || t('kbAcl.collections.moveFailed')) }
}
async function sortCollection(offset: number) {
  const current = selectedCollection.value
  if (!current) return
  try { await updateKBCollection(current.id, { parent_id: current.parent_id, name: current.name, sort_order: current.sort_order + offset }); await load() }
  catch (error: any) { MessagePlugin.error(error?.message || '排序失败') }
}
async function bindKnowledgeBase() {
  const targetCollectionID = selectedCollectionId.value === '__unclassified' ? '' : selectedCollectionId.value
  try { await setKBCollection(bindingKBID.value, targetCollectionID); bindingKBID.value = ''; await load(); MessagePlugin.success(t('kbAcl.collections.bound')) }
  catch (error: any) { MessagePlugin.error(error?.message || t('kbAcl.collections.bindFailed')) }
}
onMounted(load)
defineExpose({ reload: load })
</script>

<style scoped>
.collection-panel{margin-bottom:18px;border:1px solid var(--td-component-border);border-radius:10px;background:var(--td-bg-color-container)}.collection-toolbar{display:flex;justify-content:space-between;gap:16px;padding:14px 16px;border-bottom:1px solid var(--td-component-stroke)}.collection-toolbar>div:first-child{display:flex;align-items:baseline;gap:10px}.collection-toolbar span{font-size:12px;color:var(--td-text-color-secondary)}.collection-actions{display:flex;gap:4px}.collection-body{display:flex;flex-wrap:wrap;gap:6px;padding:12px}.collection-node{display:flex;justify-content:space-between;gap:12px;min-width:150px;padding:8px 12px;border:0;border-radius:7px;background:transparent;color:var(--td-text-color-primary);cursor:pointer}.collection-node:hover,.collection-node.active{background:var(--td-brand-color-light)}.collection-node small{color:var(--td-text-color-placeholder)}.collection-binding{display:flex;gap:8px;padding:0 16px 14px}.collection-binding .t-select{max-width:360px}
</style>
