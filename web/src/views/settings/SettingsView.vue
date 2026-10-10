<script setup lang="ts">
import { onMounted, onUnmounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import PageHeader from '@/components/common/PageHeader.vue'
import SessionDetails from '@/components/settings/SessionDetails.vue'
import MeterItemsSettings from '@/components/settings/MeterItemsSettings.vue'
import AccountSettingsPanel from '@/components/settings/AccountSettingsPanel.vue'
import RoomTypeSettings from '@/components/settings/RoomTypeSettings.vue'
import { getSettings, updateSetting } from '@/api/settings'
import {
  createMeterItem,
  listMeterItems,
  updateMeterItem,
} from '@/api/meterItems'
import { errorMessage } from '@/utils/format'
import { settingError } from '@/utils/settingsValidation'
import type { AccountSettings, SettingInput } from '@/types/settings'
import type { MeterItem, MeterItemInput } from '@/types/meterItems'

const settings = ref<AccountSettings>({})
const items = ref<MeterItem[]>([])
const settingsLoading = ref(false)
const itemsLoading = ref(false)
const settingsError = ref('')
const itemsError = ref('')
const settingsBusy = ref(false)
const itemsBusy = ref(false)
const settingsActionError = ref('')
const itemsActionError = ref('')
const settingsVersion = ref(0)
const itemsVersion = ref(0)
const itemsOffset = ref(0)
const itemsHasNext = ref(false)
let settingsRequest = 0
let itemsRequest = 0
let disposed = false

async function loadSettings() {
  const sequence = ++settingsRequest
  settingsLoading.value = true
  settingsError.value = ''
  try {
    const result = await getSettings()
    if (sequence === settingsRequest && !disposed) settings.value = result
  } catch (error) {
    if (sequence === settingsRequest && !disposed)
      settingsError.value = errorMessage(error)
  } finally {
    if (sequence === settingsRequest && !disposed) settingsLoading.value = false
  }
}
async function loadItems() {
  const sequence = ++itemsRequest
  itemsLoading.value = true
  itemsError.value = ''
  try {
    const result = await listMeterItems({
      limit: 21,
      offset: itemsOffset.value,
    })
    if (sequence === itemsRequest && !disposed) {
      items.value = result.slice(0, 20)
      itemsHasNext.value = result.length > 20
    }
  } catch (error) {
    if (sequence === itemsRequest && !disposed)
      itemsError.value = errorMessage(error)
  } finally {
    if (sequence === itemsRequest && !disposed) itemsLoading.value = false
  }
}
async function saveSetting(value: SettingInput) {
  if (settingsBusy.value || settingError(value)) return
  settingsBusy.value = true
  settingsActionError.value = ''
  try {
    const result = await updateSetting(value)
    settingsRequest++
    settingsLoading.value = false
    settingsError.value = ''
    settings.value = result
    settingsVersion.value++
    ElMessage.success('账号配置已保存')
  } catch (error) {
    settingsActionError.value = errorMessage(error)
  } finally {
    settingsBusy.value = false
  }
}
async function saveItem(id: number | null, value: MeterItemInput) {
  if (
    itemsBusy.value ||
    !value.code.trim() ||
    !value.unit.trim() ||
    !Number.isFinite(value.unit_price) ||
    value.unit_price < 0
  )
    return
  itemsBusy.value = true
  itemsActionError.value = ''
  try {
    if (id) await updateMeterItem(id, value)
    else await createMeterItem(value)
    itemsVersion.value++
    ElMessage.success('计费项已保存')
    await loadItems()
  } catch (error) {
    itemsActionError.value = errorMessage(error)
  } finally {
    itemsBusy.value = false
  }
}
function changeItemsPage(offset: number) {
  itemsOffset.value = offset
  void loadItems()
}
onMounted(() => {
  void loadSettings()
  void loadItems()
})
onUnmounted(() => {
  disposed = true
})
</script>

<template>
  <div class="settings-view">
    <PageHeader title="设置" description="整理常用配置，让日常管理更顺手。" />
    <SessionDetails />
    <MeterItemsSettings
      :items="items"
      :loading="itemsLoading"
      :error="itemsError"
      :busy="itemsBusy"
      :action-error="itemsActionError"
      :saved-version="itemsVersion"
      :offset="itemsOffset"
      :has-next="itemsHasNext"
      @retry="loadItems"
      @save="saveItem"
      @page="changeItemsPage"
      @clear-error="itemsActionError = ''"
    />
    <RoomTypeSettings />
    <AccountSettingsPanel
      :settings="settings"
      :loading="settingsLoading"
      :error="settingsError"
      :busy="settingsBusy"
      :action-error="settingsActionError"
      :saved-version="settingsVersion"
      @retry="loadSettings"
      @save="saveSetting"
      @clear-error="settingsActionError = ''"
    />
    <section class="settings-unavailable">
      <div>
        <h2>数据备份与恢复</h2>
        <p class="muted">备份、恢复和自动清理暂未开放。</p>
      </div>
      <span class="badge">暂未开放</span>
    </section>
  </div>
</template>

<style scoped>
.settings-view {
  display: grid;
  gap: 1.5rem;
}
.settings-unavailable {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 1rem;
  border: 1px dashed var(--border, #d9dce2);
  border-radius: 1.25rem;
  padding: 1.5rem;
}
.settings-unavailable h2 {
  margin: 0 0 0.4rem;
  font-size: 1rem;
}
.settings-unavailable p {
  margin: 0;
  font-size: 0.86rem;
}
</style>
