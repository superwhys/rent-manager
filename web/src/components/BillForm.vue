<script setup lang="ts">
import type { Bill, BillStatus } from '@/types'
import { ref, watch, computed } from 'vue'

// 定义组件属性
interface Props {
    bill: Bill | null
    mode: 'view' | 'edit'
}

// 定义事件
interface Emits {
    (e: 'save', bill: Bill): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

// 表单数据
const formData = ref<Bill>({
    id: '',
    room: '',
    tenant: '',
    rent: 0,
    waterFee: 0,
    electricFee: 0,
    gasFee: 0,
    otherFee: 0,
    amount: 0,
    status: 'unpaid',
    createDate: ''
})

// 支付状态选项
const statusOptions = [
    { value: 'unpaid', label: '未支付' },
    { value: 'paid', label: '已支付' }
]

// 计算属性：是否为编辑模式
const isEditMode = computed(() => props.mode === 'edit')

// 计算属性：总金额
const totalAmount = computed(() => {
    return formData.value.rent + formData.value.waterFee + formData.value.electricFee + formData.value.gasFee + formData.value.otherFee
})

// 监听费用变化，自动更新总金额
watch(totalAmount, (newTotal) => {
    formData.value.amount = newTotal
})

// 监听账单数据变化，更新表单
watch(() => props.bill, (newBill) => {
    if (newBill) {
        formData.value = { ...newBill }
    } else {
        // 重置表单，生成新的账单ID
        const now = new Date()
        const billId = `B${now.getFullYear()}${String(now.getMonth() + 1).padStart(2, '0')}${String(now.getDate()).padStart(2, '0')}${String(Math.floor(Math.random() * 999) + 1).padStart(3, '0')}`
        const createDate = now.toISOString().split('T')[0]

        formData.value = {
            id: billId,
            room: '',
            tenant: '',
            rent: 0,
            waterFee: 0,
            electricFee: 0,
            gasFee: 0,
            otherFee: 0,
            amount: 0,
            status: 'unpaid',
            createDate: createDate
        }
    }
}, { immediate: true })

// 保存账单信息
const handleSave = () => {
    // 基本验证
    if (!formData.value.room.trim()) {
        alert('请输入房间号')
        return
    }
    if (!formData.value.tenant.trim()) {
        alert('请输入租客姓名')
        return
    }
    if (formData.value.rent < 0 || formData.value.waterFee < 0 || formData.value.electricFee < 0 || formData.value.gasFee < 0 || formData.value.otherFee < 0) {
        alert('费用不能为负数')
        return
    }
    if (formData.value.amount <= 0) {
        alert('请输入有效的金额')
        return
    }
    if (!formData.value.createDate) {
        alert('请选择生成时间')
        return
    }

    emit('save', { ...formData.value })
}

// 获取支付状态显示文本
const getStatusText = (status: BillStatus) => {
    const statusMap = {
        'paid': '已支付',
        'unpaid': '未支付'
    }
    return statusMap[status]
}

// 格式化金额显示
const formatAmount = (amount: number) => {
    return `¥${amount.toLocaleString('zh-CN')}`
}
</script>

<template>
    <form @submit.prevent="handleSave">
        <div class="form-grid">
            <!-- 账单编号 -->
            <div class="form-group">
                <label class="form-label">账单编号</label>
                <el-input v-if="isEditMode" v-model="formData.id" class="form-input" :readonly="!!props.bill"
                    placeholder="自动生成" size="large" />
                <div v-else class="form-value">
                    <strong>{{ formData.id || '无' }}</strong>
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

            <!-- 租客 -->
            <div class="form-group">
                <label class="form-label">租客</label>
                <el-input v-if="isEditMode" v-model="formData.tenant" class="form-input" placeholder="请输入租客姓名"
                    size="large" />
                <div v-else class="form-value">
                    {{ formData.tenant || '无' }}
                </div>
            </div>

            <!-- 租金 -->
            <div class="form-group">
                <label class="form-label">租金 (元)</label>
                <el-input-number v-if="isEditMode" v-model="formData.rent" class="form-input" :min="0" :step="0.01"
                    :precision="2" controls-position="right" size="large" />
                <div v-else class="form-value">
                    <strong class="amount-text">{{ formatAmount(formData.rent) }}</strong>
                </div>
            </div>

            <!-- 水费 -->
            <div class="form-group">
                <label class="form-label">水费 (元)</label>
                <el-input-number v-if="isEditMode" v-model="formData.waterFee" class="form-input" :min="0" :step="0.01"
                    :precision="2" controls-position="right" size="large" />
                <div v-else class="form-value">
                    <strong class="amount-text">{{ formatAmount(formData.waterFee) }}</strong>
                </div>
            </div>

            <!-- 电费 -->
            <div class="form-group">
                <label class="form-label">电费 (元)</label>
                <el-input-number v-if="isEditMode" v-model="formData.electricFee" class="form-input" :min="0" :step="0.01"
                    :precision="2" controls-position="right" size="large" />
                <div v-else class="form-value">
                    <strong class="amount-text">{{ formatAmount(formData.electricFee) }}</strong>
                </div>
            </div>

            <!-- 燃气费 -->
            <div class="form-group">
                <label class="form-label">燃气费 (元)</label>
                <el-input-number v-if="isEditMode" v-model="formData.gasFee" class="form-input" :min="0" :step="0.01"
                    :precision="2" controls-position="right" size="large" />
                <div v-else class="form-value">
                    <strong class="amount-text">{{ formatAmount(formData.gasFee) }}</strong>
                </div>
            </div>

            <!-- 其他费用 -->
            <div class="form-group">
                <label class="form-label">其他费用 (元)</label>
                <el-input-number v-if="isEditMode" v-model="formData.otherFee" class="form-input" :min="0" :step="0.01"
                    :precision="2" controls-position="right" size="large" />
                <div v-else class="form-value">
                    <strong class="amount-text">{{ formatAmount(formData.otherFee) }}</strong>
                </div>
            </div>

            <!-- 总金额 -->
            <div class="form-group">
                <label class="form-label">总金额 (元)</label>
                <div class="form-value">
                    <strong class="amount-text total-amount">{{ formatAmount(totalAmount) }}</strong>
                </div>
            </div>

            <!-- 支付状态 -->
            <div class="form-group">
                <label class="form-label">支付状态</label>
                <el-select v-if="isEditMode" v-model="formData.status" class="form-select" placeholder="请选择支付状态"
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

            <!-- 生成时间 -->
            <div class="form-group">
                <label class="form-label">生成时间</label>
                <el-date-picker v-if="isEditMode" v-model="formData.createDate" type="date" class="form-input"
                    placeholder="请选择生成时间" value-format="YYYY-MM-DD" teleported popper-class="high-z-index-popper"
                    size="large" />
                <div v-else class="form-value">
                    {{ formData.createDate || '无' }}
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
    grid-template-columns: repeat(3, 1fr);
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

.form-input:read-only {
    background: #f9fafb;
    color: #6b7280;
}

/* 表单值显示样式 */
.form-value {
    padding: 10px 0;
    font-size: 14px;
    color: #374151;
    min-height: 20px;
}

.amount-text {
    color: #059669;
    font-size: 16px;
}

.total-amount {
    color: #dc2626;
    font-size: 18px;
    font-weight: 700;
}

/* 状态标签样式 */
.status-badge {
    padding: 4px 8px;
    border-radius: 4px;
    font-size: 12px;
    font-weight: 500;
}

.status-badge.paid {
    background: #dcfce7;
    color: #166534;
}

.status-badge.unpaid {
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
@media (max-width: 1024px) {
    .form-grid {
        grid-template-columns: repeat(2, 1fr);
    }
}

@media (max-width: 640px) {
    .form-grid {
        grid-template-columns: 1fr;
    }
}
</style>