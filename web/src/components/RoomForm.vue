<script setup lang="ts">
import type { Room, RoomStatus } from '@/types'
import { ref, watch, computed } from 'vue'

// 定义组件属性
interface Props {
    room: Room | null
    mode: 'view' | 'edit'
}

// 定义事件
interface Emits {
    (e: 'save', room: Room): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

// 表单数据
const formData = ref<Room>({
    number: '',
    floor: 1,
    area: 0,
    rent: 0,
    tenant: null,
    status: 'vacant',
    moveInDate: null
})

// 状态选项
const statusOptions = [
    { value: 'vacant', label: '空置' },
    { value: 'occupied', label: '已出租' },
    { value: 'pending', label: '待缴费' }
]

// 计算属性：是否为编辑模式
const isEditMode = computed(() => props.mode === 'edit')

// 监听房间数据变化，更新表单
watch(() => props.room, (newRoom) => {
    if (newRoom) {
        formData.value = { ...newRoom }
    } else {
        // 重置表单
        formData.value = {
            number: '',
            floor: 1,
            area: 0,
            rent: 0,
            tenant: null,
            status: 'vacant',
            moveInDate: null
        }
    }
}, { immediate: true })

// 保存房间信息
const handleSave = () => {
    // 基本验证
    if (!formData.value.number.trim()) {
        alert('请输入房间号')
        return
    }
    if (formData.value.area <= 0) {
        alert('请输入有效的房间面积')
        return
    }
    if (formData.value.rent <= 0) {
        alert('请输入有效的租金')
        return
    }

    emit('save', { ...formData.value })
}

// 获取状态显示文本
const getStatusText = (status: RoomStatus) => {
    const statusMap = {
        'vacant': '空置',
        'occupied': '已出租',
        'pending': '待缴费'
    }
    return statusMap[status]
}

// 计算属性：处理日期类型兼容性
const moveInDateValue = computed({
    get: () => formData.value.moveInDate || undefined,
    set: (value: string | undefined) => {
        formData.value.moveInDate = value || null
    }
})

// 格式化日期
const formatDate = (date: string | null) => {
    return date || '无'
}
</script>

<template>
    <form @submit.prevent="handleSave">
        <div class="form-grid">
            <!-- 房间号 -->
            <div class="form-group">
                <label class="form-label">房间号</label>
                <el-input v-if="isEditMode" v-model="formData.number" class="form-input" placeholder="请输入房间号"
                    size="large" />
                <div v-else class="form-value">
                    <strong>{{ formData.number || '无' }}</strong>
                </div>
            </div>

            <!-- 楼层 -->
            <div class="form-group">
                <label class="form-label">楼层</label>
                <el-input-number v-if="isEditMode" v-model="formData.floor" class="form-input" :min="1" :step="1"
                    controls-position="right" size="large" />
                <div v-else class="form-value">
                    {{ formData.floor }}楼
                </div>
            </div>

            <!-- 面积 -->
            <div class="form-group">
                <label class="form-label">面积 (㎡)</label>
                <el-input-number v-if="isEditMode" v-model="formData.area" class="form-input" :min="0" :step="0.1"
                    :precision="1" controls-position="right" size="large" />
                <div v-else class="form-value">
                    {{ formData.area }}㎡
                </div>
            </div>

            <!-- 租金 -->
            <div class="form-group">
                <label class="form-label">租金 (元/月)</label>
                <el-input-number v-if="isEditMode" v-model="formData.rent" class="form-input" :min="0" :step="50"
                    controls-position="right" size="large" />
                <div v-else class="form-value">
                    ¥{{ formData.rent.toLocaleString('zh-CN') }}
                </div>
            </div>

            <!-- 状态 -->
            <div class="form-group">
                <label class="form-label">状态</label>
                <el-select v-if="isEditMode" v-model="formData.status" class="form-select" placeholder="请选择状态"
                    teleported popper-class="high-z-index-popper" size="large">
                    <el-option v-for="option in statusOptions" :key="option.value" :label="option.label"
                        :value="option.value" />
                </el-select>
                <div v-else class="form-value">
                    <span class="status-badge" :class="formData.status">
                        {{ getStatusText(formData.status) }}
                    </span>
                </div>
            </div>

            <!-- 租客 -->
            <div class="form-group">
                <label class="form-label">租客</label>
                <el-input v-if="isEditMode" v-model="formData.tenant" class="form-input" placeholder="请输入租客姓名"
                    size="large" />
                <div v-else class="form-value">
                    {{ formData.tenant || '无' }}
                </div>
            </div>

            <!-- 入住时间 -->
            <div class="form-group full-width">
                <label class="form-label">入住时间</label>
                <el-date-picker v-if="isEditMode" v-model="moveInDateValue" type="date" class="form-input"
                    placeholder="请选择入住时间" value-format="YYYY-MM-DD" teleported popper-class="high-z-index-popper"
                    clearable size="large" />
                <div v-else class="form-value">
                    {{ formatDate(formData.moveInDate) }}
                </div>
            </div>
        </div>

        <!-- 编辑模式下的保存按钮 -->
        <div v-if="isEditMode" class="form-actions">
            <button type="submit" class="btn btn-primary">
                保存
            </button>
        </div>
    </form>
</template>

<style scoped>
/* 表单网格布局 */
.form-grid {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 20px;
}

.form-group {
    display: flex;
    flex-direction: column;
}

.form-group.full-width {
    grid-column: 1 / -1;
}

.form-input {
    width: 100%;
}

.form-select {
    width: 100%;
}

.form-group :deep(.el-date-editor) {
    --el-date-editor-width: 100%;
}

/* 表单值显示样式 */
.form-value {
    padding: 10px 0;
    font-size: 14px;
    color: #374151;
    min-height: 20px;
}

/* 状态标签样式 */
.status-badge {
    padding: 4px 8px;
    border-radius: 4px;
    font-size: 12px;
    font-weight: 500;
}

.status-badge.occupied {
    background: #dcfce7;
    color: #166534;
}

.status-badge.vacant {
    background: #f3f4f6;
    color: #374151;
}

.status-badge.pending {
    background: #fef3c7;
    color: #92400e;
}

/* 表单操作按钮 */
.form-actions {
    margin-top: 20px;
    display: flex;
    justify-content: flex-end;
}

.btn {
    padding: 10px 20px;
    border: none;
    border-radius: 6px;
    font-size: 14px;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.2s;
}

.btn-primary {
    background: #4f46e5;
    color: white;
}

.btn-primary:hover {
    background: #4338ca;
}

/* 响应式设计 */
@media (max-width: 640px) {
    .form-grid {
        grid-template-columns: 1fr;
    }
}
</style>