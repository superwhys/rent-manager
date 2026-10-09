<script setup lang="ts">
import type { Tenant, StatusConfig } from '@/types'
import { ref, computed, nextTick } from 'vue'
import BaseModal from '@/components/BaseModal.vue'
import TenantForm from '@/components/TenantForm.vue'
import FilterSection from '@/components/FilterSection.vue'

// 租客管理数据
const tenantsData = ref<Tenant[]>([
    { name: '张三', phone: '138-0001-0001', idCard: '440123199001010001', room: '101', deposit: 1200, moveInDate: '2024-01-15', contractStatus: 'active' },
    { name: '李四', phone: '138-0002-0002', idCard: '440123199002020002', room: '102', deposit: 1200, moveInDate: '2024-02-01', contractStatus: 'active' },
    { name: '王五', phone: '138-0003-0003', idCard: '440123199003030003', room: '104', deposit: 1200, moveInDate: '2024-03-10', contractStatus: 'active' },
    { name: '赵六', phone: '138-0004-0004', idCard: '440123199004040004', room: '201', deposit: 1200, moveInDate: '2024-01-20', contractStatus: 'pending' },
    { name: '孙七', phone: '138-0005-0005', idCard: '440123199005050005', room: '202', deposit: 1200, moveInDate: '2024-04-05', contractStatus: 'active' },
    { name: '周八', phone: '138-0006-0006', idCard: '440123199006060006', room: '203', deposit: 1200, moveInDate: '2024-02-15', contractStatus: 'active' },
    { name: '吴九', phone: '138-0007-0007', idCard: '440123199007070007', room: '301', deposit: 1300, moveInDate: '2024-01-05', contractStatus: 'active' },
    { name: '郑十', phone: '138-0008-0008', idCard: '440123199008080008', room: '302', deposit: 1300, moveInDate: '2024-03-20', contractStatus: 'active' },
    { name: '钱一', phone: '138-0009-0009', idCard: '440123199009090009', room: '304', deposit: 1300, moveInDate: '2024-02-28', contractStatus: 'active' }
])

// 表格列配置
const columns = [
    { key: 'name', label: '姓名', width: '100px' },
    { key: 'phone', label: '手机号', width: '200px', ellipsis: true, showTooltip: true },
    { key: 'idCard', label: '身份证号', ellipsis: true, showTooltip: true },
    { key: 'room', label: '房间号', width: '100px' },
    { key: 'deposit', label: '押金(元)', width: '100px' },
    { key: 'moveInDate', label: '入住时间', width: '160px' },
    { key: 'contractStatus', label: '合同状态', width: '100px' }
]

// 状态配置
const statusConfig: StatusConfig = {
    'active': { text: '正常', class: 'active' },
    'pending': { text: '待租房', class: 'pending' },
    'expired': { text: '已到期', class: 'expired' }
}

// 房间号选项
const roomOptions = computed(() => {
    const rooms = [...new Set(tenantsData.value.map(tenant => tenant.room))].sort()
    return rooms.map(room => ({ value: room, label: `${room}房间` }))
})

// 合同状态选项
const contractStatusOptions = [
    { value: 'active', label: '正常' },
    { value: 'pending', label: '待租房' },
    { value: 'expired', label: '已到期' }
]

// 筛选项配置
const filterItems = computed(() => [
    {
        type: 'select' as const,
        key: 'room',
        placeholder: '选择房间号',
        options: roomOptions.value
    },
    {
        type: 'select' as const,
        key: 'contractStatus',
        placeholder: '合同状态',
        options: contractStatusOptions
    },
    {
        type: 'date-range' as const,
        key: 'moveInDate',
        startPlaceholder: '入住开始日期',
        endPlaceholder: '入住结束日期'
    }
])

// 当前筛选条件
const currentFilters = ref<Record<string, any>>({})

// 筛选后的租客数据
const filteredTenants = computed(() => {
    return tenantsData.value.filter(tenant => {
        // 房间号筛选
        if (currentFilters.value.room && tenant.room !== currentFilters.value.room) {
            return false
        }

        // 合同状态筛选
        if (currentFilters.value.contractStatus && tenant.contractStatus !== currentFilters.value.contractStatus) {
            return false
        }

        // 入住时间范围筛选
        if (currentFilters.value.moveInDateStart || currentFilters.value.moveInDateEnd) {
            if (!tenant.moveInDate) {
                return false
            }

            const moveInDate = new Date(tenant.moveInDate)

            if (currentFilters.value.moveInDateStart) {
                const startDate = new Date(currentFilters.value.moveInDateStart)
                if (moveInDate < startDate) return false
            }

            if (currentFilters.value.moveInDateEnd) {
                const endDate = new Date(currentFilters.value.moveInDateEnd)
                if (moveInDate > endDate) return false
            }
        }

        return true
    })
})

// 筛选区域引用
const filterSectionRef = ref()

// 处理筛选条件变化
const handleFilterChange = (filters: Record<string, any>) => {
    currentFilters.value = filters
    // 使用 nextTick 确保计算属性更新后再更新筛选结果数量
    nextTick(() => {
        if (filterSectionRef.value) {
            filterSectionRef.value.updateFilteredCount(filteredTenants.value.length)
        }
    })
}

// 处理重置筛选
const handleFilterReset = () => {
    currentFilters.value = {}
}

// 模态框状态管理
const showModal = ref(false)
const currentTenant = ref<Tenant | null>(null)
const modalMode = ref<'view' | 'edit'>('view')

// 计算属性：模态框标题
const modalTitle = computed(() => {
    if (modalMode.value === 'edit') {
        return currentTenant.value ? '编辑租客' : '新增租客'
    }
    return '租客详情'
})

// 处理编辑租客
const handleEdit = (tenant: Tenant) => {
    currentTenant.value = tenant
    modalMode.value = 'edit'
    showModal.value = true
}

// 处理查看租客详情
const handleView = (tenant: Tenant) => {
    currentTenant.value = tenant
    modalMode.value = 'view'
    showModal.value = true
}

// 处理删除租客
const handleDelete = (tenant: Tenant) => {
    if (confirm(`确定要删除租客 ${tenant.name} 吗？`)) {
        const index = tenantsData.value.findIndex(t => t.name === tenant.name && t.room === tenant.room)
        if (index > -1) {
            tenantsData.value.splice(index, 1)
        }
    }
}

// 显示添加租客模态框
const showAddTenantModal = () => {
    currentTenant.value = null
    modalMode.value = 'edit'
    showModal.value = true
}

// 关闭模态框
const handleCloseModal = () => {
    showModal.value = false
    currentTenant.value = null
}

// 保存租客信息
const handleSaveTenant = (tenant: Tenant) => {
    if (currentTenant.value) {
        // 编辑模式：更新现有租客
        const index = tenantsData.value.findIndex(t => t.name === currentTenant.value!.name && t.room === currentTenant.value!.room)
        if (index > -1) {
            tenantsData.value[index] = { ...tenant }
        }
    } else {
        // 新增模式：检查租客是否重复
        const exists = tenantsData.value.some(t => t.name === tenant.name && t.room === tenant.room)
        if (exists) {
            alert('租客信息已存在')
            return
        }
        tenantsData.value.push({ ...tenant })
    }

    handleCloseModal()
}
</script>

<template>
    <div class="tenants">
        <div class="section-header">
            <h2 class="section-title">👥 租客管理</h2>
            <button class="btn btn-primary" @click="showAddTenantModal">
                ➕ 添加租客
            </button>
        </div>

        <!-- 筛选区域 -->
        <FilterSection ref="filterSectionRef" :filter-items="filterItems" :total-count="tenantsData.length"
            item-name="租客" @filter-change="handleFilterChange" @reset="handleFilterReset" />

        <DataTable :columns="columns" :data="filteredTenants" :show-actions="true" @edit="handleEdit" @view="handleView"
            @delete="handleDelete">
            <template #name="{ value }">
                <strong>{{ value }}</strong>
            </template>

            <template #deposit="{ value }">
                ¥{{ value.toLocaleString('zh-CN') }}
            </template>

            <template #contractStatus="{ value }">
                <StatusBadge :status="value" :status-config="statusConfig" />
            </template>
        </DataTable>

        <!-- 租客模态框 -->
        <BaseModal :show="showModal" :title="modalTitle" :show-footer="false" @close="handleCloseModal">
            <template #body>
                <TenantForm :tenant="currentTenant" :mode="modalMode" @save="handleSaveTenant" />
            </template>
        </BaseModal>
    </div>
</template>

<style scoped>
.tenants {
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