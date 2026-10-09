<script setup lang="ts">
import type { Property } from '@/types'
import { ref, computed } from 'vue'
import BaseModal from '@/components/BaseModal.vue'
import PropertyForm from '@/components/PropertyForm.vue'


/**
 * 物业管理页面
 * 提供物业的增删改查功能
 */
const propertiesData = ref<Property[]>([
    {
        id: 1,
        name: '深圳光明区福华东八巷三号',
        address: '广东省深圳市光明区马田街道福华东八巷三号',
        totalRooms: 12,
        occupiedRooms: 10,
        monthlyIncome: 14400,
        status: 'active',
        createDate: '2024-01-15',
        defaultRent: 1200,
        depositMultiple: 1.5,
        networkFee: 50,
        elevatorFee: 30
    },
    {
        id: 2,
        name: '深圳南山区科技园南区',
        address: '广东省深圳市南山区科技园南区深南大道',
        totalRooms: 24,
        occupiedRooms: 20,
        monthlyIncome: 31200,
        status: 'active',
        createDate: '2024-02-01',
        defaultRent: 1300,
        depositMultiple: 2,
        networkFee: 60,
        elevatorFee: 40
    },
    {
        id: 3,
        name: '深圳福田区华强北赛格广场',
        address: '广东省深圳市福田区华强北路赛格广场',
        totalRooms: 18,
        occupiedRooms: 15,
        monthlyIncome: 22500,
        status: 'maintenance',
        createDate: '2024-03-10',
        defaultRent: 1250,
        depositMultiple: 1,
        networkFee: 55,
        elevatorFee: 35
    }
])

// 表格列配置
const columns = [
    { key: 'name', label: '物业名称', showTooltip: true, ellipsis: true, width: '300px' },
    { key: 'address', label: '物业地址', showTooltip: true, ellipsis: true },
    { key: 'totalRooms', label: '房间总数', width: '100px' },
    { key: 'occupiedRooms', label: '出租总数', width: '100px' },
    { key: 'defaultRent', label: '默认租金', width: '100px' },
    { key: 'depositMultiple', label: '押金倍数', width: '100px' },
    { key: 'networkFee', label: '网络费', width: '100px' },
    { key: 'elevatorFee', label: '电梯费', width: '100px' },
    { key: 'monthlyIncome', label: '月收入', width: '100px' }
]



// 模态框状态管理
const showModal = ref(false)
const currentProperty = ref<Property | null>(null)
const modalMode = ref<'view' | 'edit'>('view')

// 计算属性：模态框标题
const modalTitle = computed(() => {
    if (modalMode.value === 'edit') {
        return currentProperty.value ? '编辑物业' : '新增物业'
    }
    return '物业详情'
})

/**
 * 处理编辑物业
 */
const handleEdit = (property: Property) => {
    currentProperty.value = property
    modalMode.value = 'edit'
    showModal.value = true
}

/**
 * 处理查看物业详情
 */
const handleView = (property: Property) => {
    currentProperty.value = property
    modalMode.value = 'view'
    showModal.value = true
}

/**
 * 处理删除物业
 */
const handleDelete = (property: Property) => {
    if (confirm(`确定要删除物业 "${property.name}" 吗？`)) {
        const index = propertiesData.value.findIndex(p => p.id === property.id)
        if (index > -1) {
            propertiesData.value.splice(index, 1)
        }
    }
}

/**
 * 显示添加物业模态框
 */
const showAddPropertyModal = () => {
    currentProperty.value = null
    modalMode.value = 'edit'
    showModal.value = true
}

/**
 * 关闭模态框
 */
const handleCloseModal = () => {
    showModal.value = false
    currentProperty.value = null
}

/**
 * 保存物业信息
 */
const handleSaveProperty = (property: Property) => {
    if (currentProperty.value) {
        // 编辑模式：更新现有物业
        const index = propertiesData.value.findIndex(p => p.id === currentProperty.value!.id)
        if (index > -1) {
            propertiesData.value[index] = { ...property }
        }
    } else {
        // 新增模式：检查物业名称是否重复
        const exists = propertiesData.value.some(p => p.name === property.name)
        if (exists) {
            alert('物业名称已存在，请使用其他名称')
            return
        }
        propertiesData.value.push({ ...property })
    }

    handleCloseModal()
}

/**
 * 格式化收入显示
 */
const formatIncome = (income: number) => {
    return `¥${income.toLocaleString('zh-CN')}`
}

/**
 * 截断长文本
 */
const truncateText = (text: string, maxLength: number = 30) => {
    if (text.length <= maxLength) return text
    return text.substring(0, maxLength) + '...'
}
</script>

<template>
    <div class="properties">
        <div class="section-header">
            <h2 class="section-title">🏢 物业管理</h2>
            <PermissionGuard permission="properties.create">
                <button class="btn btn-primary" @click="showAddPropertyModal">
                    ➕ 添加物业
                </button>
            </PermissionGuard>
        </div>

        <DataTable :columns="columns" :data="propertiesData" :show-actions="true" @edit="handleEdit" @view="handleView"
            @delete="handleDelete" edit-permission="properties.edit" delete-permission="properties.delete">
            <template #name="{ value }">
                <span :title="value">{{ truncateText(value, 25) }}</span>
            </template>

            <template #address="{ value }">
                <span :title="value">{{ truncateText(value, 40) }}</span>
            </template>

            <template #totalRooms="{ value }">
                {{ value }} 间
            </template>

            <template #occupiedRooms="{ value }">
                {{ value }} 间
            </template>

            <template #defaultRent="{ value }">
                ¥{{ value }}/月
            </template>

            <template #depositMultiple="{ value }">
                {{ value }}倍
            </template>

            <template #networkFee="{ value }">
                ¥{{ value }}/月
            </template>

            <template #elevatorFee="{ value }">
                ¥{{ value }}/月
            </template>

            <template #monthlyIncome="{ value }">
                {{ formatIncome(value) }}
            </template>
        </DataTable>

        <!-- 物业模态框 -->
        <BaseModal :show="showModal" :title="modalTitle" :show-footer="false" @close="handleCloseModal">
            <template #body>
                <PropertyForm :property="currentProperty" :mode="modalMode" @save="handleSaveProperty" />
            </template>
        </BaseModal>
    </div>
</template>

<style scoped>
.properties {
    padding: 20px;
}

.section-title {
    font-size: 24px;
    font-weight: 600;
    margin-bottom: 20px;
    color: #333;
}

.section-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 30px;
    flex-wrap: wrap;
    gap: 15px;
}

.btn {
    padding: 12px 24px;
    border: none;
    border-radius: 8px;
    font-size: 14px;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.3s ease;
}

.btn-primary {
    background: #4f46e5;
    color: white;
}

.btn-primary:hover {
    background: #4338ca;
    transform: translateY(-2px);
}
</style>