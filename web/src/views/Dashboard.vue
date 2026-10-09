<script setup lang="ts">
import type { DashboardData, RoomInfo, FeeCalculationResult } from '@/types'
import { ref, computed, onMounted } from 'vue'
import BaseModal from '@/components/BaseModal.vue'
import RoomDetailsForm from '@/components/RoomDetailsForm.vue'
import FilterSection from '@/components/FilterSection.vue'

// 仪表板统计数据
const dashboardData = ref<DashboardData>({
    totalRooms: {
        name: '总房间数',
        value: 43
    },
    totalTenants: {
        name: '已出租房间',
        value: 40
    },
    totalBills: {
        name: '本月总收入',
        value: 52800
    },
    totalUtilities: {
        name: '本月总支出',
        value: 3200
    },
    totalReports: {
        name: '报表总数',
        value: 0
    },
    totalSettings: {
        name: '系统设置总数',
        value: 0
    }
})

// 综合房间数据（包含抄表记录和账单数据）
const enhancedRoomsData = ref<RoomInfo[]>([
    {
        id: '101',
        status: 'occupied',
        rent: 1200,
        tenant: '张三',
        feeConfig: {
            baseRent: 1200,
            networkFee: 50,
            elevatorFee: 30,
            otherFee: 0,
            waterPrice: 3.5,
            waterMinimum: 15,
            electricPrice: 0.8,
            electricMinimum: 20,
            gasPrice: 2.8,
            gasMinimum: 10
        },
        utilityReading: {
            date: '2025-01-01',
            lastWaterReading: 1233.3,
            lastElectricReading: 3275.4,
            lastGasReading: 884.3,
            waterReading: 1245.6,
            electricReading: 3421.2,
            gasReading: 892.5,
            waterUsage: 12.3,
            electricUsage: 145.8,
            gasUsage: 8.2
        },
        bill: {
            id: 'B202501001',
            rent: 1200,
            waterFee: 45,
            electricFee: 40,
            gasFee: 0,
            otherFee: 0,
            amount: 1285,
            status: 'paid',
            createDate: '2025-01-01'
        }
    },
    {
        id: '102',
        status: 'occupied',
        rent: 1200,
        tenant: '李四',
        utilityReading: {
            date: '2025-01-01',
            lastWaterReading: 1183.2,
            lastElectricReading: 3198.4,
            lastGasReading: 847.2,
            waterReading: 1198.4,
            electricReading: 3356.8,
            gasReading: 856.3,
            waterUsage: 15.2,
            electricUsage: 158.4,
            gasUsage: 9.1
        },
        bill: {
            id: 'B202501002',
            rent: 1200,
            waterFee: 52,
            electricFee: 40,
            gasFee: 0,
            otherFee: 0,
            amount: 1292,
            status: 'unpaid',
            createDate: '2025-01-01'
        }
    },
    { id: '103', status: 'vacant', rent: 1200, tenant: '暂无租客' },
    {
        id: '104',
        status: 'occupied',
        rent: 1200,
        tenant: '王五',
        utilityReading: {
            date: '2025-01-01',
            lastWaterReading: 1145.0,
            lastElectricReading: 3156.9,
            lastGasReading: 826.7,
            waterReading: 1156.8,
            electricReading: 3289.5,
            gasReading: 834.2,
            waterUsage: 11.8,
            electricUsage: 132.6,
            gasUsage: 7.5
        }
    },
    {
        id: '201',
        status: 'pending',
        rent: 1200,
        tenant: '赵六',
        utilityReading: {
            date: '2025-01-01',
            lastWaterReading: 1272.7,
            lastElectricReading: 3287.9,
            lastGasReading: 902.2,
            waterReading: 1289.2,
            electricReading: 3456.8,
            gasReading: 912.4,
            waterUsage: 16.5,
            electricUsage: 168.9,
            gasUsage: 10.2
        },
        bill: {
            id: 'B202501003',
            rent: 1200,
            waterFee: 55,
            electricFee: 40,
            gasFee: 0,
            otherFee: 0,
            amount: 1295,
            status: 'paid',
            createDate: '2025-01-01'
        }
    },
    {
        id: '202',
        status: 'occupied',
        rent: 1200,
        tenant: '孙七',
        utilityReading: {
            date: '2025-01-01',
            lastWaterReading: 1220.7,
            lastElectricReading: 3249.0,
            lastGasReading: 870.9,
            waterReading: 1234.5,
            electricReading: 3398.2,
            gasReading: 879.6,
            waterUsage: 13.8,
            electricUsage: 149.2,
            gasUsage: 8.7
        }
    },
    { id: '203', status: 'occupied', rent: 1200, tenant: '周八' },
    { id: '204', status: 'vacant', rent: 1200, tenant: '暂无租客' },
    { id: '301', status: 'occupied', rent: 1300, tenant: '吴九' },
    { id: '302', status: 'occupied', rent: 1300, tenant: '郑十' },
    { id: '303', status: 'vacant', rent: 1300, tenant: '暂无租客' },
    { id: '304', status: 'occupied', rent: 1300, tenant: '钱一' },
    { id: '401', status: 'occupied', rent: 1200, tenant: '孙七' },
    { id: '402', status: 'occupied', rent: 1200, tenant: '周八' },
    { id: '403', status: 'vacant', rent: 1200, tenant: '暂无租客' },
    { id: '404', status: 'occupied', rent: 1300, tenant: '吴九' },
    { id: '501', status: 'occupied', rent: 1300, tenant: '郑十' },
    { id: '502', status: 'vacant', rent: 1300, tenant: '暂无租客' },
    { id: '503', status: 'occupied', rent: 1300, tenant: '钱一' }
])

// 模态框状态管理
const showModal = ref(false)
const currentRoomId = ref<string>('')
const currentRoomData = ref<RoomInfo | null>(null)
const modalMode = ref<'view' | 'edit'>('edit')
const currentRoomHasReading = ref(false)
const currentRoomHasBill = ref(false)

// FilterSection 组件引用
const filterSectionRef = ref<InstanceType<typeof FilterSection> | null>(null)



// 筛选器配置
const filterItems = ref([
    {
        type: 'month' as const,
        key: 'selectedMonth',
        placeholder: '选择年月',
        width: '160px'
    }
])

// 当前选择的月份（默认为当前月份）
const currentDate = new Date()
const defaultMonth = `${currentDate.getFullYear()}-${String(currentDate.getMonth() + 1).padStart(2, '0')}`
const selectedMonth = ref(defaultMonth)

// 检查房间当月是否已抄表
const checkRoomReadingStatus = (roomId: string, month: string): boolean => {
    const room = enhancedRoomsData.value.find(room => room.id === roomId)
    if (room?.utilityReading) {
        const readingMonth = room.utilityReading.date.substring(0, 7)
        return readingMonth === month
    }
    return false
}

// 检查房间当月是否已有账单
const checkRoomBillStatus = (roomId: string, month: string): boolean => {
    const room = enhancedRoomsData.value.find(room => room.id === roomId)
    if (room?.bill) {
        const billMonth = room.bill.createDate.substring(0, 7)
        return billMonth === month
    }
    return false
}

const modalTitle = computed(() => {
    return `房间 ${currentRoomId.value} 费用详情`
})

const displayMonth = computed(() => {
    if (selectedMonth.value) {
        const [year, month] = selectedMonth.value.split('-')
        return `${year}年${month}月`
    }
    return '当月'
})

// 处理筛选器变化
const handleFilterChange = (filters: Record<string, any>) => {
    const newMonth = filters.selectedMonth || defaultMonth
    selectedMonth.value = newMonth
    // 根据选择的月份重新加载数据
    loadDashboardData(newMonth)
}

// 重置筛选器
const handleFilterReset = () => {
    selectedMonth.value = defaultMonth
    loadDashboardData(defaultMonth)
}

// 加载仪表板数据（根据选择的月份）
const loadDashboardData = (month: string) => {
    // 这里可以根据月份参数调用API获取对应月份的数据
    // 目前使用模拟数据，实际项目中应该调用API
    console.log(`加载 ${month} 月份的仪表板数据`)

    // 模拟数据更新（实际项目中应该从API获取）
    // dashboardData.value = await fetchDashboardData(month)
    // enhancedRoomsData.value = await fetchEnhancedRoomsData(month)
}

// 处理房间卡片点击
const handleRoomClick = (room: RoomInfo) => {
    currentRoomId.value = room.id
    currentRoomData.value = room

    // 检查当前房间当月是否已抄表
    currentRoomHasReading.value = checkRoomReadingStatus(room.id, selectedMonth.value)

    // 检查当前房间当月是否已有账单
    currentRoomHasBill.value = checkRoomBillStatus(room.id, selectedMonth.value)

    modalMode.value = 'edit'
    showModal.value = true
}

// 关闭模态框
const handleCloseModal = () => {
    showModal.value = false
    currentRoomId.value = ''
    currentRoomData.value = null
}

// 生成账单
const handleGenerateBill = (roomData: RoomInfo, result: FeeCalculationResult) => {
    // 更新房间数据中的费用配置
    const roomIndex = enhancedRoomsData.value.findIndex(room => room.id === roomData.id)
    if (roomIndex !== -1) {
        enhancedRoomsData.value[roomIndex] = { ...roomData }
    }

    // 这里可以跳转到账单页面或者显示成功提示
    console.log('生成账单:', {
        roomId: roomData.id,
        roomData: roomData,
        calculation: result
    })
    // 调用接口生成账单
    alert(`房间 ${roomData.id} 的账单已生成！\n总计：¥${result.grandTotal.toFixed(2)}`)
    handleCloseModal()
}

// 组件挂载时初始化
onMounted(() => {
    // 初始化加载当前月份的数据
    loadDashboardData(selectedMonth.value)
})
</script>

<template>
    <div class="dashboard">
        <h2 class="section-title">📊 {{ displayMonth }}概览</h2>

        <!-- 筛选器 -->
        <FilterSection ref="filterSectionRef" :filter-items="filterItems" :total-count="0" item-name="数据"
            @filter-change="handleFilterChange" @reset="handleFilterReset" />

        <!-- 统计卡片 -->
        <div class="dashboard-stats">
            <StatCard backgroundColor="linear-gradient(135deg, #4facfe 0%, #00f2fe 100%)"
                :name="dashboardData.totalRooms.name" :value="dashboardData.totalRooms.value" />
            <StatCard backgroundColor="linear-gradient(135deg, #11998e 0%, #38ef7d 100%)"
                :name="dashboardData.totalTenants.name" :value="dashboardData.totalTenants.value" />
            <StatCard backgroundColor="linear-gradient(135deg, #f7971e 0%, #ffd200 100%)"
                :name="dashboardData.totalBills.name"
                :value="`¥${dashboardData.totalBills.value.toLocaleString('zh-CN')}`" />
            <StatCard backgroundColor="linear-gradient(135deg, #fc4a1a 0%, #f7b733 100%)"
                :name="dashboardData.totalUtilities.name"
                :value="`¥${dashboardData.totalUtilities.value.toLocaleString('zh-CN')}`" />
            <StatCard backgroundColor="linear-gradient(135deg, #667eea 0%, #764ba2 100%)"
                :name="dashboardData.totalReports.name" :value="dashboardData.totalReports.value" />
            <StatCard backgroundColor="linear-gradient(135deg, #36d1dc 0%, #5b86e5 100%)"
                :name="dashboardData.totalSettings.name" :value="dashboardData.totalSettings.value" />
        </div>

        <div class="rooms-section">
            <h3 class="section-title">🏠 房间状态总览</h3>
            <div class="rooms-grid">
                <RoomCard v-for="room in enhancedRoomsData" :key="room.id" :room="room"
                    :has-bill="checkRoomBillStatus(room.id, selectedMonth)" @click="handleRoomClick(room)"
                    class="clickable-room-card" />
            </div>
        </div>

        <BaseModal :show="showModal" :title="modalTitle" :show-footer="false" size="large" @close="handleCloseModal">
            <template #body>
                <RoomDetailsForm :room-data="currentRoomData" :mode="modalMode" :has-reading="currentRoomHasReading"
                    :has-bill="currentRoomHasBill" @generate-bill="handleGenerateBill" />
            </template>
        </BaseModal>
    </div>
</template>

<style scoped>
.dashboard {
    padding: 20px;
}

.section-title {
    font-size: 24px;
    font-weight: 600;
    margin-bottom: 20px;
    color: #333;
}

.dashboard-stats {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
    gap: 20px;
    margin-bottom: 40px;
}

.rooms-section {
    margin-top: 30px;
}

.rooms-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
    gap: 15px;
}

.clickable-room-card {
    cursor: pointer;
    transition: all 0.3s ease;
}

.clickable-room-card:hover {
    transform: translateY(-2px);
    box-shadow: 0 8px 25px rgba(0, 0, 0, 0.15);
}
</style>