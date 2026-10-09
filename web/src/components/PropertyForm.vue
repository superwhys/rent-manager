<script setup lang="ts">
import type { Property, PropertyStatus, CreatePropertyForm } from '@/types'
import { ref, watch, computed } from 'vue'

/**
 * 物业表单组件的属性接口
 */
interface Props {
    property: Property | null
    mode: 'view' | 'edit'
}

/**
 * 物业表单组件的事件接口
 */
interface Emits {
    (e: 'save', property: Property): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

// 表单数据
const formData = ref<CreatePropertyForm>({
    name: '',
    address: '',
    defaultRent: 1200,
    depositMultiple: 1,
    networkFee: 50,
    elevatorFee: 30
})

// 状态选项
const statusOptions = [
    { value: 'active', label: '正常运营' },
    { value: 'inactive', label: '暂停运营' },
    { value: 'maintenance', label: '维护中' }
]



// 计算属性：是否为编辑模式
const isEditMode = computed(() => props.mode === 'edit')

// 监听物业数据变化，更新表单
watch(() => props.property, (newProperty) => {
    if (newProperty) {
        formData.value = {
            name: newProperty.name,
            address: newProperty.address,
            defaultRent: newProperty.defaultRent,
            depositMultiple: newProperty.depositMultiple,
            networkFee: newProperty.networkFee,
            elevatorFee: newProperty.elevatorFee
        }
    } else {
        // 重置表单
        formData.value = {
            name: '',
            address: '',
            defaultRent: 1200,
            depositMultiple: 1,
            networkFee: 50,
            elevatorFee: 30
        }
    }
}, { immediate: true })

/**
 * 保存物业信息
 */
const handleSave = () => {
    // 基本验证
    if (!formData.value.name.trim()) {
        alert('请输入物业名称')
        return
    }
    if (!formData.value.address.trim()) {
        alert('请输入物业地址')
        return
    }

    // 构造完整的物业对象
    const propertyData: Property = {
        id: props.property?.id || Date.now(),
        name: formData.value.name,
        address: formData.value.address,
        totalRooms: props.property?.totalRooms || 0,
        occupiedRooms: props.property?.occupiedRooms || 0,
        monthlyIncome: props.property?.monthlyIncome || 0,
        status: props.property?.status || 'active',
        createDate: props.property?.createDate || new Date().toISOString().split('T')[0],
        defaultRent: formData.value.defaultRent,
        depositMultiple: formData.value.depositMultiple,
        networkFee: formData.value.networkFee,
        elevatorFee: formData.value.elevatorFee
    }

    emit('save', propertyData)
}

/**
 * 获取状态显示文本
 */
const getStatusText = (status: PropertyStatus) => {
    const statusMap = {
        'active': '正常运营',
        'inactive': '暂停运营',
        'maintenance': '维护中'
    }
    return statusMap[status]
}

/**
 * 格式化收入显示
 */
const formatIncome = (income: number) => {
    return `¥${income.toLocaleString('zh-CN')}`
}
</script>

<template>
    <form @submit.prevent="handleSave">
        <div class="form-grid">
            <!-- 物业名称 -->
            <div class="form-group">
                <label class="form-label">物业名称</label>
                <el-input v-if="isEditMode" v-model="formData.name" class="form-input" placeholder="请输入物业名称"
                    size="large" />
                <div v-else class="form-value">
                    <strong>{{ formData.name || '无' }}</strong>
                </div>
            </div>

            <!-- 状态 (仅查看模式显示) -->
            <div v-if="!isEditMode && property" class="form-group">
                <label class="form-label">运营状态</label>
                <div class="form-value">
                    <span class="status-badge" :class="property.status">
                        {{ getStatusText(property.status) }}
                    </span>
                </div>
            </div>

            <!-- 物业地址 -->
            <div class="form-group full-width">
                <label class="form-label">物业地址</label>
                <el-input v-if="isEditMode" v-model="formData.address" class="form-input" placeholder="请输入详细地址"
                    size="large" />
                <div v-else class="form-value">
                    {{ formData.address || '无' }}
                </div>
            </div>

            <!-- 房间总数 (仅查看模式) -->
            <div v-if="!isEditMode && property" class="form-group">
                <label class="form-label">房间总数</label>
                <div class="form-value">
                    {{ property.totalRooms }} 间
                </div>
            </div>

            <!-- 出租总数 (仅查看模式) -->
            <div v-if="!isEditMode && property" class="form-group">
                <label class="form-label">出租总数</label>
                <div class="form-value">
                    {{ property.occupiedRooms }} 间
                </div>
            </div>

            <!-- 月收入 (仅查看模式) -->
            <div v-if="!isEditMode && property" class="form-group">
                <label class="form-label">月收入</label>
                <div class="form-value">
                    {{ formatIncome(property.monthlyIncome) }}
                </div>
            </div>

            <!-- 创建时间 (仅查看模式) -->
            <div v-if="!isEditMode && property" class="form-group">
                <label class="form-label">创建时间</label>
                <div class="form-value">
                    {{ property.createDate }}
                </div>
            </div>

            <!-- 费用设置分割线 -->
            <div class="full-width">
                <div class="section-divider">费用设置</div>
            </div>

            <!-- 默认租金 -->
            <div class="form-group">
                <label class="form-label">默认租金 (元/月)</label>
                <el-input-number v-if="isEditMode" v-model="formData.defaultRent" :min="0" :step="50" 
                    class="form-input" controls-position="right" size="large" />
                <div v-else class="form-value">
                    ¥{{ formData.defaultRent }}/月
                </div>
            </div>

            <!-- 押金倍数 -->
            <div class="form-group">
                <label class="form-label">押金倍数</label>
                <el-input-number v-if="isEditMode" v-model="formData.depositMultiple" :min="0" :step="0.1" 
                    :precision="1" class="form-input" controls-position="right" size="large" />
                <div v-else class="form-value">
                    {{ formData.depositMultiple }}倍
                </div>
            </div>

            <!-- 网络费 -->
            <div class="form-group">
                <label class="form-label">网络费 (元/月)</label>
                <el-input-number v-if="isEditMode" v-model="formData.networkFee" :min="0" :step="10" 
                    class="form-input" controls-position="right" size="large" />
                <div v-else class="form-value">
                    ¥{{ formData.networkFee }}/月
                </div>
            </div>

            <!-- 电梯费 -->
            <div class="form-group">
                <label class="form-label">电梯费 (元/月)</label>
                <el-input-number v-if="isEditMode" v-model="formData.elevatorFee" :min="0" :step="10" 
                    class="form-input" controls-position="right" size="large" />
                <div v-else class="form-value">
                    ¥{{ formData.elevatorFee }}/月
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

.form-group.full-width,
.full-width {
    grid-column: 1 / -1;
}

.form-label {
    font-size: 14px;
    font-weight: 500;
    color: #374151;
    margin-bottom: 8px;
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

.status-badge.inactive {
    background: #fef2f2;
    color: #dc2626;
}

.status-badge.maintenance {
    background: #fef3c7;
    color: #92400e;
}

/* 分割线样式 */
.section-divider {
    padding: 15px 0 10px 0;
    font-size: 16px;
    font-weight: 600;
    color: #374151;
    border-bottom: 1px solid #e5e7eb;
    margin-bottom: 15px;
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