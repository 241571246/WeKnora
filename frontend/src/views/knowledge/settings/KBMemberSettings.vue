<template>
  <div class="kb-member-settings">
    <div class="section-header">
      <h3 class="section-title">{{ t('kbAcl.members.title') }}</h3>
      <p class="section-desc">{{ t('kbAcl.members.description') }}</p>
    </div>

    <div class="member-add-row">
      <t-select v-model="candidateUserId" :placeholder="t('kbAcl.members.selectEmployee')" filterable>
        <t-option v-for="member in availableTenantMembers" :key="member.user_id" :value="member.user_id"
          :label="member.username || member.email || member.user_id" />
      </t-select>
      <t-select v-model="candidateRole" class="role-select">
        <t-option v-for="option in candidateRoleOptions" :key="option.value" :value="option.value" :label="option.label" />
      </t-select>
      <t-button theme="primary" :loading="saving" :disabled="!candidateUserId" @click="addMember">{{ t('common.add') }}</t-button>
    </div>
    <div v-if="candidateRole === 'custom'" class="capability-matrix">
      <strong>{{ t('kbAcl.members.customPermissions') }}</strong>
      <t-checkbox-group v-model="candidateCapabilities" :options="capabilityOptions" />
    </div>

    <t-loading :loading="loading">
      <div v-if="rows.length" class="member-list">
        <div v-for="row in rows" :key="row.user_id" class="member-row">
          <div class="member-identity">
            <strong>{{ tenantMemberMap.get(row.user_id)?.username || tenantMemberMap.get(row.user_id)?.email || row.user_id }}</strong>
            <span>{{ tenantMemberMap.get(row.user_id)?.email }}</span>
          </div>
          <t-select :value="row.role" class="role-select" :disabled="row.role === 'owner' && !canManageOwners"
            @change="(value: string | number) => changeRole(row, value as KBMemberRole)">
            <t-option v-for="option in rowRoleOptions(row)" :key="option.value" :value="option.value" :label="option.label" />
          </t-select>
          <t-popconfirm :content="row.role === 'owner' ? t('kbAcl.members.removeOwnerConfirm') : t('kbAcl.members.removeConfirm')"
            @confirm="removeMember(row)">
            <t-button theme="danger" variant="text" :disabled="row.role === 'owner' && !canManageOwners">{{ t('common.remove') }}</t-button>
          </t-popconfirm>
          <div v-if="row.role === 'custom'" class="capability-matrix member-capabilities">
            <t-checkbox-group :value="rowCapabilityValues(row)" :options="capabilityOptions"
              @change="(value: Array<string | number>) => updateCustomCapabilities(row, value as KBCapability[])" />
          </div>
        </div>
      </div>
      <t-empty v-else :description="t('kbAcl.members.empty')" />
    </t-loading>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { MessagePlugin } from 'tdesign-vue-next'
import { addKBMember, deleteKBMember, listKBMembers, updateKBMember, type KBCapability, type KBMembership, type KBMemberRole } from '@/api/knowledge-base'
import { fetchAllTenantMembers, type TenantMember } from '@/api/tenant/members'
import { useAuthStore } from '@/stores/auth'
import { useI18n } from 'vue-i18n'
import { canMutateMemberRole } from '../kbAclPresentation'

const props = withDefaults(defineProps<{ kbId: string; canManageOwners?: boolean }>(), { canManageOwners: false })
const authStore = useAuthStore()
const { t } = useI18n()
const loading = ref(false)
const saving = ref(false)
const rows = ref<KBMembership[]>([])
const tenantMembers = ref<TenantMember[]>([])
const candidateUserId = ref('')
const candidateRole = ref<KBMemberRole>('ai_user')
const candidateCapabilities = ref<KBCapability[]>(['kb.ai.query', 'kb.metadata.read'])
const roleOptions = computed<Array<{ value: KBMemberRole; label: string }>>(() => [
  { value: 'owner', label: t('kbAcl.roles.owner') }, { value: 'editor', label: t('kbAcl.roles.editor') },
  { value: 'document_viewer', label: t('kbAcl.roles.documentViewer') }, { value: 'ai_user', label: t('kbAcl.roles.aiUser') },
  { value: 'custom', label: t('kbAcl.roles.custom') },
])
const candidateRoleOptions = computed(() => roleOptions.value.filter(option => option.value !== 'owner' || props.canManageOwners))
const capabilityOptions = computed<Array<{ value: KBCapability; label: string }>>(() => [
  ...(['aiQuery','metadataRead','documentsList','documentPreview','chunkPreview','documentDownload','documentUpload','documentEdit','documentDelete','documentReparse','chunkEdit','chunkDelete','folderManage','settingsEdit','kbDelete','membersManage','ownersManage'] as const)
    .map((key, index) => ({ value: capabilityCodes[index], label: t(`kbAcl.capabilities.${key}`) })),
])
const capabilityCodes: KBCapability[] = ['kb.ai.query','kb.metadata.read','kb.documents.list','kb.document.preview','kb.chunk.preview','kb.document.download','kb.document.upload','kb.document.edit','kb.document.delete','kb.document.reparse','kb.chunk.edit','kb.chunk.delete','kb.folder.manage','kb.settings.edit','kb.delete','kb.members.manage','kb.owners.manage']
const tenantMemberMap = computed(() => new Map(tenantMembers.value.map(item => [item.user_id, item])))
const availableTenantMembers = computed(() => {
  const assigned = new Set(rows.value.map(item => item.user_id))
  return tenantMembers.value.filter(item => !assigned.has(item.user_id))
})

async function load() {
  loading.value = true
  try {
    const [memberRes, workspaceMembers] = await Promise.all([
      listKBMembers(props.kbId) as any,
      fetchAllTenantMembers(Number(authStore.currentTenantId)),
    ])
    rows.value = memberRes?.data?.members || []
    tenantMembers.value = workspaceMembers
  } catch (error: any) {
    MessagePlugin.error(error?.message || '成员加载失败')
  } finally { loading.value = false }
}

async function addMember() {
  if (!canMutateMemberRole(undefined, candidateRole.value, props.canManageOwners)) {
    MessagePlugin.error(t('kbAcl.members.ownerPermissionRequired'))
    return
  }
  if (candidateRole.value === 'owner' && !window.confirm(t('kbAcl.members.addOwnerConfirm'))) return
  saving.value = true
  try {
    await addKBMember(props.kbId, {
      user_id: candidateUserId.value,
      role: candidateRole.value,
      capabilities: candidateRole.value === 'custom' ? candidateCapabilities.value : undefined,
    })
    candidateUserId.value = ''
    MessagePlugin.success(t('kbAcl.members.added'))
    await load()
  } catch (error: any) { MessagePlugin.error(error?.message || t('kbAcl.members.addFailed')) }
  finally { saving.value = false }
}

async function changeRole(row: KBMembership, role: KBMemberRole) {
  if (!canMutateMemberRole(row.role, role, props.canManageOwners)) {
    MessagePlugin.error(t('kbAcl.members.ownerPermissionRequired'))
    return
  }
  if ((row.role === 'owner' || role === 'owner') && !window.confirm(t('kbAcl.members.changeOwnerConfirm'))) return
  try {
    const current = rowCapabilityValues(row)
    const capabilities = role === 'custom'
      ? (current.length ? current : ['kb.ai.query', 'kb.metadata.read'] as KBCapability[])
      : undefined
    await updateKBMember(props.kbId, row.user_id, { role, capabilities })
    MessagePlugin.success(t('kbAcl.members.updated'))
    await load()
  } catch (error: any) { MessagePlugin.error(error?.message || t('kbAcl.members.updateFailed')) }
}

function rowRoleOptions(row: KBMembership) {
  return roleOptions.value.filter(option => option.value !== 'owner' || props.canManageOwners || row.role === 'owner')
}

function rowCapabilityValues(row: KBMembership): KBCapability[] {
  return (row.capabilities || []).map(item => item.capability)
}

async function updateCustomCapabilities(row: KBMembership, capabilities: KBCapability[]) {
  try {
    await updateKBMember(props.kbId, row.user_id, { role: 'custom', capabilities })
    row.capabilities = capabilities.map(capability => ({ membership_id: row.id, capability }))
    MessagePlugin.success(t('kbAcl.members.updated'))
  } catch (error: any) { MessagePlugin.error(error?.message || t('kbAcl.members.updateFailed')) }
}

async function removeMember(row: KBMembership) {
  if (!canMutateMemberRole(row.role, undefined, props.canManageOwners)) {
    MessagePlugin.error(t('kbAcl.members.ownerPermissionRequired'))
    return
  }
  try {
    await deleteKBMember(props.kbId, row.user_id)
    MessagePlugin.success(t('kbAcl.members.removed'))
    await load()
  } catch (error: any) { MessagePlugin.error(error?.message || t('kbAcl.members.removeFailed')) }
}

onMounted(load)
</script>

<style scoped>
.member-add-row,.member-row{display:grid;grid-template-columns:minmax(220px,1fr) 160px auto;gap:12px;align-items:center}.member-add-row{margin:20px 0}.member-list{border:1px solid var(--td-component-border);border-radius:8px}.member-row{padding:14px 16px;border-bottom:1px solid var(--td-component-stroke)}.member-row:last-child{border-bottom:0}.member-identity{display:flex;flex-direction:column;gap:4px;min-width:0}.member-identity span{font-size:12px;color:var(--td-text-color-secondary)}.role-select{width:160px}.capability-matrix{padding:12px 14px;margin:-8px 0 16px;border:1px solid var(--td-component-stroke);border-radius:8px;background:var(--td-bg-color-secondarycontainer)}.capability-matrix strong{display:block;margin-bottom:10px}.member-capabilities{grid-column:1/-1;margin:0}.capability-matrix :deep(.t-checkbox){min-width:150px;margin-bottom:8px}
</style>
