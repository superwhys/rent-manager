<script setup lang="ts">
import type { Tenant } from '@/types'
import { ref, watch, computed } from 'vue'

// 定义组件属性
interface Props {
    tenant: Tenant | null
    mode: 'view' | 'edit'
}

// 定义事件
interface Emits {
    (e: 'save', tenant: Tenant): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

// 表单数据
const formData = ref<Tenant>({
    name: '',
    phone: '',
    idCard: '',
    room: '',
    deposit: 0,
    moveInDate: '',
    contractStatus: 'active'
})

// 合同状态选项
const contractStatusOptions = [
    { value: 'active', label: '正常' },
    { value: 'pending', label: '待租房' },
    { value: 'expired', label: '已到期' }
]

// 计算属性：是否为编辑模式
const isEditMode = computed(() => props.mode === 'edit')

// 监听租客数据变化，更新表单
watch(() => props.tenant, (newTenant) => {
    if (newTenant) {
        formData.value = { ...newTenant }
    } else {
        // 重置表单
        formData.value = {
            name: '',
            phone: '',
            idCard: '',
            room: '',
            deposit: 0,
            moveInDate: '',
            contractStatus: 'active'
        }
    }
}, { immediate: true })

// 保存租客信息
const handleSave = () => {
    // 基本验证
    if (!formData.value.name.trim()) {
        alert('请输入租客姓名')
        return
    }
    if (!formData.value.phone.trim()) {
        alert('请输入手机号')
        return
    }
    if (!formData.value.idCard.trim()) {
        alert('请输入身份证号')
        return
    }
    if (!formData.value.room.trim()) {
        alert('请输入房间号')
        return
    }
    if (formData.value.deposit <= 0) {
        alert('请输入有效的押金金额')
        return
    }
    if (!formData.value.moveInDate) {
        alert('请选择入住时间')
        return
    }

    emit('save', { ...formData.value })
}

// 获取合同状态显示文本
const getContractStatusText = (status: string) => {
    const statusMap = {
        'active': '正常',
        'pending': '待租房',
        'expired': '已到期'
    }
    return statusMap[status as keyof typeof statusMap] || status
}

// 格式化身份证号显示（隐藏中间部分）
const formatIdCard = (idCard: string) => {
    if (!idCard) return '无'
    if (idCard.length >= 10) {
        return idCard.slice(0, 6) + '****' + idCard.slice(-4)
    }
    return idCard
}
</script>

<template>
    <form @submit.prevent="handleSave">
        <div class="form-grid">
            <!-- 姓名 -->
            <div class="form-group">
                <label class="form-label">姓名</label>
                <el-input v-if="isEditMode" v-model="formData.name" class="form-input" placeholder="请输入租客姓名"
                    size="large" />
                <div v-else class="form-value">
                    <strong>{{ formData.name || '无' }}</strong>
                </div>
            </div>

            <!-- 手机号 -->
            <div class="form-group">
                <label class="form-label">手机号</label>
                <el-input v-if="isEditMode" v-model="formData.phone" type="tel" class="form-input" placeholder="请输入手机号"
                    size="large" />
                <div v-else class="form-value">
                    {{ formData.phone || '无' }}
                </div>
            </div>

            <!-- 身份证号 -->
            <div class="form-group full-width">
                <label class="form-label">身份证号</label>
                <el-input v-if="isEditMode" v-model="formData.idCard" class="form-input" placeholder="请输入身份证号"
                    size="large" />
                <div v-else class="form-value">
                    {{ formatIdCard(formData.idCard) }}
                </div>
            </div>

            <!-- 房间号 -->
            <div class="form-group">
                <label class="form-label">房间号</label>
                <el-input v-if="isEditMode" v-model="formData.room" class="form-input" placeholder="请输入房间号"
                    size="large" />
                <div v-else class="form-value">
                    <strong>{{ formData.room || '无' }}</strong>
                </div>
            </div>

            <!-- 押金 -->
            <div class="form-group">
                <label class="form-label">押金 (元)</label>
                <el-input-number v-if="isEditMode" v-model="formData.deposit" class="form-input" :min="0" :step="100"
                    controls-position="right" size="large" />
                <div v-else class="form-value">
                    ¥{{ formData.deposit.toLocaleString('zh-CN') }}
                </div>
            </div>

            <!-- 入住时间 -->
            <div class="form-group">
                <label class="form-label">入住时间</label>
                <el-date-picker v-if="isEditMode" v-model="formData.moveInDate" type="date" class="form-input"
                    placeholder="请选择入住时间" value-format="YYYY-MM-DD" teleported popper-class="high-z-index-popper"
                    size="large" />
                <div v-else class="form-value">
                    {{ formData.moveInDate || '无' }}
                </div>
            </div>

            <!-- 合同状态 -->
            <div class="form-group">
                <label class="form-label">合同状态</label>
                <el-select v-if="isEditMode" v-model="formData.contractStatus" class="form-select" placeholder="请选择合同状态"
                    teleported popper-class="high-z-index-popper" size="large">
                    <el-option v-for="option in contractStatusOptions" :key="option.value" :label="option.label"
                        :value="option.value" />
                </el-select>
                <div v-else class="form-value">
                    <span class="status-badge" :class="formData.contractStatus">
                        {{ getContractStatusText(formData.contractStatus) }}
                    </span>
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

.status-badge.active {
    background: #dcfce7;
    color: #166534;
}

.status-badge.pending {
    background: #fef3c7;
    color: #92400e;
}

.status-badge.expired {
    background: #fee2e2;
    color: #dc2626;
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