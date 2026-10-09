<script setup lang="ts">
import type { Bill, BillStats, StatusConfig } from '@/types'
import { ref, computed, nextTick } from 'vue'
import BaseModal from '@/components/BaseModal.vue'
import BillForm from '@/components/BillForm.vue'
import FilterSection from '@/components/FilterSection.vue'

// 账单统计数据
const billStats = ref<BillStats>({
    totalAmount: { name: '本月应收', value: 52800 },
    paidAmount: { name: '已收金额', value: 48600 },
    unpaidAmount: { name: '未收金额', value: 4200 },
    paymentRate: { name: '收费率', value: 92 }
})

// 账单管理数据
const billsData = ref<Bill[]>([
    { id: 'B202507001', room: '101', tenant: '张三', rent: 1200, waterFee: 45, electricFee: 40, gasFee: 0, otherFee: 0, amount: 1285, status: 'paid', createDate: '2025-07-01' },
    { id: 'B202507002', room: '102', tenant: '李四', rent: 1200, waterFee: 52, electricFee: 40, gasFee: 0, otherFee: 0, amount: 1292, status: 'paid', createDate: '2025-07-01' },
    { id: 'B202507003', room: '104', tenant: '王五', rent: 1200, waterFee: 38, electricFee: 40, gasFee: 0, otherFee: 0, amount: 1278, status: 'unpaid', createDate: '2025-07-01' },
    { id: 'B202507004', room: '201', tenant: '赵六', rent: 1200, waterFee: 55, electricFee: 40, gasFee: 0, otherFee: 0, amount: 1295, status: 'unpaid', createDate: '2025-07-01' },
    { id: 'B202507005', room: '202', tenant: '孙七', rent: 1200, waterFee: 48, electricFee: 40, gasFee: 0, otherFee: 0, amount: 1288, status: 'paid', createDate: '2025-07-01' },
    { id: 'B202507006', room: '203', tenant: '周八', rent: 1200, waterFee: 42, electricFee: 40, gasFee: 0, otherFee: 0, amount: 1282, status: 'paid', createDate: '2025-07-01' },
    { id: 'B202507007', room: '301', tenant: '吴九', rent: 1300, waterFee: 65, electricFee: 40, gasFee: 0, otherFee: 0, amount: 1405, status: 'paid', createDate: '2025-07-01' },
    { id: 'B202507008', room: '302', tenant: '郑十', rent: 1300, waterFee: 58, electricFee: 40, gasFee: 0, otherFee: 0, amount: 1398, status: 'unpaid', createDate: '2025-07-01' },
    { id: 'B202507009', room: '304', tenant: '钱一', rent: 1300, waterFee: 72, electricFee: 40, gasFee: 0, otherFee: 0, amount: 1412, status: 'paid', createDate: '2025-07-01' }
])

// 表格列配置
const columns = [
    { key: 'id', label: '账单编号', width: '140px', ellipsis: true, showTooltip: true },
    { key: 'room', label: '房间号', width: '80px' },
    { key: 'tenant', label: '租客', width: '80px' },
    { key: 'rent', label: '租金', width: '100px' },
    { key: 'waterFee', label: '水费', width: '100px' },
    { key: 'electricFee', label: '电费', width: '100px' },
    { key: 'gasFee', label: '燃气费', width: '100px' },
    { key: 'otherFee', label: '其他费用', width: '100px' },
    { key: 'amount', label: '总金额'},
    { key: 'status', label: '支付状态', width: '100px' },
    { key: 'createDate', label: '生成时间', width: '110px' }
]

// 状态配置
const statusConfig: StatusConfig = {
    'paid': { text: '已支付', class: 'paid' },
    'unpaid': { text: '未支付', class: 'unpaid' }
}

// 房间号选项
const roomOptions = computed(() => {
    const rooms = [...new Set(billsData.value.map(bill => bill.room))].sort()
    return rooms.map(room => ({ value: room, label: `${room}房间` }))
})

// 租客选项
const tenantOptions = computed(() => {
    const tenants = [...new Set(billsData.value.map(bill => bill.tenant))].sort()
    return tenants.map(tenant => ({ value: tenant, label: tenant }))
})

// 支付状态选项
const paymentStatusOptions = [
    { value: 'paid', label: '已支付' },
    { value: 'unpaid', label: '未支付' }
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
        key: 'tenant',
        placeholder: '选择租客',
        options: tenantOptions.value
    },
    {
        type: 'select' as const,
        key: 'status',
        placeholder: '支付状态',
        options: paymentStatusOptions
    },
    {
        type: 'date-range' as const,
        key: 'createDate',
        startPlaceholder: '生成开始日期',
        endPlaceholder: '生成结束日期'
    }
])

// 当前筛选条件
const currentFilters = ref<Record<string, any>>({})

// 筛选后的账单数据
const filteredBills = computed(() => {
    return billsData.value.filter(bill => {
        // 房间号筛选
        if (currentFilters.value.room && bill.room !== currentFilters.value.room) {
            return false
        }

        // 租客筛选
        if (currentFilters.value.tenant && bill.tenant !== currentFilters.value.tenant) {
            return false
        }

        // 支付状态筛选
        if (currentFilters.value.status && bill.status !== currentFilters.value.status) {
            return false
        }

        // 生成时间范围筛选
        if (currentFilters.value.createDateStart || currentFilters.value.createDateEnd) {
            if (!bill.createDate) {
                return false
            }

            const createDate = new Date(bill.createDate)

            if (currentFilters.value.createDateStart) {
                const startDate = new Date(currentFilters.value.createDateStart)
                if (createDate < startDate) return false
            }

            if (currentFilters.value.createDateEnd) {
                const endDate = new Date(currentFilters.value.createDateEnd)
                if (createDate > endDate) return false
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
            filterSectionRef.value.updateFilteredCount(filteredBills.value.length)
        }
    })
}

// 处理重置筛选
const handleFilterReset = () => {
    currentFilters.value = {}
}

// 模态框状态管理
const showModal = ref(false)
const currentBill = ref<Bill | null>(null)
const modalMode = ref<'view' | 'edit'>('view')

// 计算属性：模态框标题
const modalTitle = computed(() => {
    if (modalMode.value === 'edit') {
        return currentBill.value ? '编辑账单' : '新增账单'
    }
    return '账单详情'
})

// 处理编辑账单
const handleEdit = (bill: Bill) => {
    currentBill.value = bill
    modalMode.value = 'edit'
    showModal.value = true
}

// 处理查看账单详情
const handleView = (bill: Bill) => {
    currentBill.value = bill
    modalMode.value = 'view'
    showModal.value = true
}

// 处理删除账单
const handleDelete = (bill: Bill) => {
    if (confirm(`确定要删除账单 ${bill.id} 吗？`)) {
        const index = billsData.value.findIndex(b => b.id === bill.id)
        if (index > -1) {
            billsData.value.splice(index, 1)
            // 更新统计数据
            updateBillStats()
        }
    }
}

// 关闭模态框
const handleCloseModal = () => {
    showModal.value = false
    currentBill.value = null
}

// 保存账单信息
const handleSaveBill = (bill: Bill) => {
    if (currentBill.value) {
        // 编辑模式：更新现有账单
        const index = billsData.value.findIndex(b => b.id === currentBill.value!.id)
        if (index > -1) {
            billsData.value[index] = { ...bill }
        }
    } else {
        // 新增模式：检查账单ID是否重复
        const exists = billsData.value.some(b => b.id === bill.id)
        if (exists) {
            alert('账单编号已存在，请使用其他编号')
            return
        }
        billsData.value.push({ ...bill })
    }

    // 更新统计数据
    updateBillStats()
    handleCloseModal()
}

// 更新账单统计数据
const updateBillStats = () => {
    const totalAmount = billsData.value.reduce((sum, bill) => sum + bill.amount, 0)
    const paidAmount = billsData.value.filter(bill => bill.status === 'paid').reduce((sum, bill) => sum + bill.amount, 0)
    const unpaidAmount = totalAmount - paidAmount
    const paymentRate = totalAmount > 0 ? Math.round((paidAmount / totalAmount) * 100) : 0

    billStats.value = {
        totalAmount: { name: '本月应收', value: totalAmount },
        paidAmount: { name: '已收金额', value: paidAmount },
        unpaidAmount: { name: '未收金额', value: unpaidAmount },
        paymentRate: { name: '收费率', value: paymentRate }
    }
}

const generateBills = () => {
    // 显示新增账单模态框
    currentBill.value = null
    modalMode.value = 'edit'
    showModal.value = true
}

const exportBills = () => {
    console.log('导出账单')
}
</script>

<template>
    <div class="bills">
        <div class="section-header">
            <h2 class="section-title">💰 账单管理</h2>
            <div class="header-controls">
                <button class="btn btn-primary" @click="generateBills">
                    📝 生成本月账单
                </button>
                <button class="btn btn-success" @click="exportBills">
                    📤 导出账单
                </button>
            </div>
        </div>

        <!-- 账单统计 -->
        <div class="bills-stats">
            <StatCard :name="billStats.totalAmount.name"
                :value="`¥${billStats.totalAmount.value.toLocaleString('zh-CN')}`"
                background-color="linear-gradient(135deg, #11998e 0%, #38ef7d 100%)" />
            <StatCard :name="billStats.paidAmount.name"
                :value="`¥${billStats.paidAmount.value.toLocaleString('zh-CN')}`"
                background-color="linear-gradient(135deg, #667eea 0%, #764ba2 100%)" />
            <StatCard :name="billStats.unpaidAmount.name"
                :value="`¥${billStats.unpaidAmount.value.toLocaleString('zh-CN')}`"
                background-color="linear-gradient(135deg, #fc4a1a 0%, #f7b733 100%)" />
            <StatCard :name="billStats.paymentRate.name" :value="`${billStats.paymentRate.value}%`"
                background-color="linear-gradient(135deg, #4facfe 0%, #00f2fe 100%)" />
        </div>

        <!-- 筛选区域 -->
        <FilterSection ref="filterSectionRef" :filter-items="filterItems" :total-count="billsData.length" item-name="账单"
            @filter-change="handleFilterChange" @reset="handleFilterReset" />

        <DataTable :columns="columns" :data="filteredBills" :show-actions="true" @edit="handleEdit" @view="handleView"
            @delete="handleDelete">
            <template #id="{ value }">
                <strong>{{ value }}</strong>
            </template>

            <template #rent="{ value }">
                ¥{{ value.toLocaleString('zh-CN') }}
            </template>

            <template #waterFee="{ value }">
                ¥{{ value.toLocaleString('zh-CN') }}
            </template>

            <template #electricFee="{ value }">
                ¥{{ value.toLocaleString('zh-CN') }}
            </template>

            <template #gasFee="{ value }">
                ¥{{ value.toLocaleString('zh-CN') }}
            </template>

            <template #otherFee="{ value }">
                ¥{{ value.toLocaleString('zh-CN') }}
            </template>

            <template #amount="{ value }">
                <strong>¥{{ value.toLocaleString('zh-CN') }}</strong>
            </template>

            <template #status="{ value }">
                <StatusBadge :status="value" :status-config="statusConfig" />
            </template>
        </DataTable>

        <!-- 账单模态框 -->
        <BaseModal :show="showModal" :title="modalTitle" :show-footer="false" @close="handleCloseModal">
            <template #body>
                <BillForm :bill="currentBill" :mode="modalMode" @save="handleSaveBill" />
            </template>
        </BaseModal>
    </div>
</template>

<style scoped>
.bills {
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

.header-controls {
    display: flex;
    gap: 10px;
    flex-wrap: wrap;
}

.bills-stats {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
    gap: 20px;
    margin-bottom: 30px;
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

.btn-success {
    background: #10b981;
    color: white;
}

.btn-success:hover {
    background: #059669;
    transform: translateY(-2px);
}
</style>