<script setup lang="ts">
// 数据表格组件
import { ref, computed, onMounted, onUnmounted } from 'vue'
import PermissionGuard from './PermissionGuard.vue'

interface TableColumn {
    key: string
    label: string
    width?: string
    ellipsis?: boolean // 是否启用文本省略
    showTooltip?: boolean // 是否显示悬浮提示
}

interface Props<T = Record<string, any>> {
    columns: TableColumn[]
    data: T[]
    showActions?: boolean
    // 权限控制
    editPermission?: string | string[] // 编辑操作需要的权限
    deletePermission?: string | string[] // 删除操作需要的权限
}

const props = defineProps<Props>()
const emit = defineEmits<{
    edit: [item: any]
    delete: [item: any]
    view: [item: any]
}>()

const screenWidth = ref(window.innerWidth)

const actionColumnWidth = computed(() => {
    return screenWidth.value <= 480 ? '120' : '200'
})

const updateScreenWidth = () => {
    screenWidth.value = window.innerWidth
}

onMounted(() => {
    window.addEventListener('resize', updateScreenWidth)
})

onUnmounted(() => {
    window.removeEventListener('resize', updateScreenWidth)
})

// 处理操作按钮点击
const handleEdit = (item: any) => {
    emit('edit', item)
}

const handleDelete = (item: any) => {
    emit('delete', item)
}

const handleView = (item: any) => {
    emit('view', item)
}
</script>

<template>
    <div class="data-table-container">
        <el-table :data="props.data" style="width: 100%" stripe border :header-cell-style="{
            background: '#f8fafc',
            color: '#374151',
            fontWeight: '600'
        }">
            <el-table-column v-for="column in props.columns" :key="column.key" :prop="column.key" :label="column.label"
                :width="column.width">
                <template #default="{ row }">
                    <el-tooltip v-if="column.ellipsis || column.showTooltip" :content="String(row[column.key] || '')"
                        placement="top" :disabled="!row[column.key] || String(row[column.key]).length <= 10"
                        effect="dark" :show-after="100">
                        <div class="text-ellipsis">
                            <slot :name="column.key" :item="row" :value="row[column.key]">
                                {{ row[column.key] }}
                            </slot>
                        </div>
                    </el-tooltip>

                    <div v-else>
                        <slot :name="column.key" :item="row" :value="row[column.key]">
                            {{ row[column.key] }}
                        </slot>
                    </div>
                </template>
            </el-table-column>

            <el-table-column v-if="props.showActions" label="操作" :width="actionColumnWidth" fixed="right">
                <template #default="{ row }">
                    <div class="action-buttons">
                        <!-- 编辑按钮 - 权限控制 -->
                        <PermissionGuard v-if="props.editPermission" :permission="props.editPermission">
                            <el-button type="primary" size="small" @click="handleEdit(row)">
                                编辑
                            </el-button>
                        </PermissionGuard>
                        <el-button v-else type="primary" size="small" @click="handleEdit(row)">
                            编辑
                        </el-button>

                        <el-button type="info" size="small" @click="handleView(row)">
                            详情
                        </el-button>

                        <!-- 删除按钮 - 权限控制 -->
                        <PermissionGuard v-if="props.deletePermission" :permission="props.deletePermission">
                            <el-button type="danger" size="small" @click="handleDelete(row)">
                                删除
                            </el-button>
                        </PermissionGuard>
                        <el-button v-else type="danger" size="small" @click="handleDelete(row)">
                            删除
                        </el-button>
                    </div>
                </template>
            </el-table-column>
        </el-table>
    </div>
</template>

<style scoped>
.data-table-container {
    background: white;
    border-radius: 10px;
    overflow: hidden;
    box-shadow: 0 2px 10px rgba(0, 0, 0, 0.05);
}

.action-buttons {
    display: flex;
    gap: 8px;
    justify-content: center;
    align-items: center;
    padding: 0 8px;
}

.text-ellipsis {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    max-width: 100%;
}

@media (max-width: 480px) {
    .action-buttons {
        flex-direction: column;
        gap: 6px;
        padding: 6px 4px;
        align-items: center;
        justify-content: flex-start;
        min-height: auto;
    }
    
    .action-buttons .el-button {
        width: 80px;
        min-width: auto;
        margin: 0;
    }
    
    /* 确保表格单元格垂直对齐 */
    :deep(.el-table__cell) {
        vertical-align: top;
    }
}
</style>