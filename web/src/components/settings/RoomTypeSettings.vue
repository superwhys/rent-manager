<script setup lang="ts">
import { onMounted, onUnmounted, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  createRoomType,
  deleteRoomType,
  listRoomTypes,
  updateRoomType,
} from '@/api/roomTypes'
import { errorMessage } from '@/utils/format'
import type { RoomType } from '@/types/roomTypes'
import AsyncState from '@/components/common/AsyncState.vue'
import AppIcon from '@/components/common/AppIcon.vue'

const roomTypes = ref<RoomType[]>([])
const loading = ref(false)
const error = ref('')
const search = ref('')
const appliedSearch = ref('')
const page = ref(0)
const pageSize = 12
const hasNext = ref(false)
const editorVisible = ref(false)
const editing = ref<RoomType>()
const name = ref('')
const saving = ref(false)
const formError = ref('')
const deleting = ref<number>()
let version = 0

async function load() {
  const current = ++version
  loading.value = true
  error.value = ''
  try {
    const result = await listRoomTypes({
      name: appliedSearch.value || undefined,
      limit: pageSize + 1,
      offset: page.value * pageSize,
    })
    if (current !== version) return
    roomTypes.value = result.slice(0, pageSize)
    hasNext.value = result.length > pageSize
  } catch (err) {
    if (current === version) error.value = errorMessage(err)
  } finally {
    if (current === version) loading.value = false
  }
}

function applySearch() {
  page.value = 0
  appliedSearch.value = search.value.trim()
  void load()
}

function changePage(direction: number) {
  if (loading.value || page.value + direction < 0) return
  page.value += direction
  void load()
}

function openEditor(type?: RoomType) {
  editing.value = type
  name.value = type?.name ?? ''
  formError.value = ''
  editorVisible.value = true
}

async function save() {
  if (saving.value) return
  const value = name.value.trim()
  if (!value || [...value].length > 12) {
    formError.value = '房型名称需为 1–12 个字'
    return
  }
  saving.value = true
  formError.value = ''
  try {
    if (editing.value) await updateRoomType(editing.value.id, value)
    else await createRoomType(value)
  } catch (err) {
    formError.value = errorMessage(err)
    return
  } finally {
    saving.value = false
  }
  editorVisible.value = false
  ElMessage.success('房型已保存')
  await load()
  if (error.value) ElMessage.warning('房型已保存，列表刷新失败，请重试')
}

async function remove(type: RoomType) {
  if (deleting.value !== undefined) return
  deleting.value = type.id
  try {
    await ElMessageBox.confirm(
      `确认删除“${type.name}”房型？删除后无法恢复。`,
      '删除房型',
      {
        confirmButtonText: '删除房型',
        cancelButtonText: '保留房型',
        type: 'warning',
      },
    )
  } catch {
    deleting.value = undefined
    return
  }
  try {
    await deleteRoomType(type.id)
  } catch (err) {
    ElMessage.error(errorMessage(err))
    return
  } finally {
    deleting.value = undefined
  }
  ElMessage.success('房型已删除')
  if (roomTypes.value.length === 1 && page.value > 0) page.value -= 1
  await load()
  if (error.value) ElMessage.warning('房型已删除，列表刷新失败，请重试')
}

onMounted(load)
onUnmounted(() => {
  version += 1
})
</script>

<template>
  <section class="room-types panel">
    <div class="section-heading">
      <div>
        <h2>房型配置</h2>
        <p class="muted">用清晰的房型名称，整理不同空间。</p>
      </div>
      <el-button type="primary" plain @click="openEditor()"
        ><AppIcon name="plus" />新增房型</el-button
      >
    </div>
    <form class="toolbar type-search" @submit.prevent="applySearch">
      <el-input
        v-model="search"
        clearable
        placeholder="搜索房型名称"
        aria-label="搜索房型名称"
        @clear="applySearch"
      />
      <el-button native-type="submit" :loading="loading">搜索</el-button>
    </form>
    <AsyncState
      :loading="loading"
      :error="error"
      :empty="roomTypes.length === 0"
      empty-title="暂无房型"
      empty-description="添加单间、一居室等房型，创建房间时即可选择。"
      @retry="load"
    >
      <div class="type-list">
        <div v-for="type in roomTypes" :key="type.id" class="type-row">
          <div class="type-name">
            <AppIcon name="grid" /><strong>{{ type.name }}</strong
            ><span class="muted">#{{ type.id }}</span>
          </div>
          <div class="type-actions">
            <el-button
              text
              :aria-label="`编辑${type.name}`"
              @click="openEditor(type)"
              ><AppIcon name="edit"
            /></el-button>
            <el-button
              text
              type="danger"
              :aria-label="`删除${type.name}`"
              :loading="deleting === type.id"
              :disabled="deleting !== undefined"
              @click="remove(type)"
              ><AppIcon name="trash"
            /></el-button>
          </div>
        </div>
      </div>
    </AsyncState>
    <div v-if="page > 0 || hasNext" class="type-pagination">
      <span class="muted">第 {{ page + 1 }} 页</span>
      <el-button :disabled="page === 0 || loading" @click="changePage(-1)"
        >上一页</el-button
      >
      <el-button :disabled="!hasNext || loading" @click="changePage(1)"
        >下一页</el-button
      >
    </div>
    <el-dialog
      v-model="editorVisible"
      :title="editing ? '编辑房型' : '新增房型'"
      width="420px"
      append-to-body
      :close-on-click-modal="false"
      :close-on-press-escape="!saving"
      :show-close="!saving"
    >
      <form @submit.prevent="save">
        <label class="type-label" for="room-type-name">房型名称</label>
        <el-input
          id="room-type-name"
          v-model="name"
          maxlength="12"
          show-word-limit
          placeholder="例如：阳光一居室"
          autofocus
        />
        <p v-if="formError" class="form-error" role="alert">{{ formError }}</p>
        <div class="editor-actions">
          <el-button :disabled="saving" @click="editorVisible = false"
            >取消</el-button
          >
          <el-button type="primary" native-type="submit" :loading="saving"
            >保存房型</el-button
          >
        </div>
      </form>
    </el-dialog>
  </section>
</template>

<style scoped>
.room-types {
  padding: 1.5rem;
}
.section-heading {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 1rem;
}
.section-heading h2 {
  font-size: 1.05rem;
  margin: 0;
  font-weight: 600;
}
.section-heading p {
  font-size: 0.8rem;
  margin: 0.45rem 0 0;
}
.type-search {
  margin: 1.3rem 0 0.75rem;
}
.type-search .el-input {
  flex: 1;
}
.type-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 1rem;
  padding: 0.8rem 0;
  border-bottom: 1px solid var(--border, #eff1ee);
}
.type-row:last-child {
  border-bottom: 0;
}
.type-name {
  display: flex;
  gap: 0.65rem;
  align-items: center;
  min-width: 0;
}
.type-name :deep(svg) {
  color: #779187;
  flex-shrink: 0;
}
.type-name strong {
  font-size: 0.86rem;
  font-weight: 500;
  overflow-wrap: anywhere;
}
.type-name span {
  font-size: 0.7rem;
}
.type-actions {
  display: flex;
  flex-shrink: 0;
}
.type-actions :deep(.el-button + .el-button) {
  margin-left: 0;
}
.type-pagination {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  margin-top: 1rem;
}
.type-pagination > span {
  margin-right: auto;
  font-size: 0.75rem;
}
.type-label {
  display: block;
  margin-bottom: 0.65rem;
  font-size: 0.85rem;
}
.form-error {
  color: var(--el-color-danger);
  font-size: 0.8rem;
}
.editor-actions {
  display: flex;
  justify-content: flex-end;
  margin-top: 1.5rem;
  gap: 0.4rem;
}
@media (max-width: 480px) {
  .room-types {
    padding: 1.1rem;
  }
  .section-heading {
    flex-wrap: wrap;
  }
}
</style>
