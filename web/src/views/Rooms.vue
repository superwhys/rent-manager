<script setup lang="ts">
import type { Room, StatusConfig } from '@/types'
import { ref, computed, nextTick } from 'vue'
import BaseModal from '@/components/BaseModal.vue'
import RoomForm from '@/components/RoomForm.vue'
import FilterSection from '@/components/FilterSection.vue'

// 房间管理数据
const roomsData = ref<Room[]>([
    { number: '101', floor: 1, area: 25, rent: 1200, tenant: '张三', status: 'occupied', moveInDate: '2024-01-15' },
    { number: '102', floor: 1, area: 25, rent: 1200, tenant: '李四', status: 'occupied', moveInDate: '2024-02-01' },
    { number: '103', floor: 1, area: 25, rent: 1200, tenant: null, status: 'vacant', moveInDate: null },
    { number: '104', floor: 1, area: 25, rent: 1200, tenant: '王五', status: 'occupied', moveInDate: '2024-03-10' },
    { number: '201', floor: 2, area: 25, rent: 1200, tenant: '赵六', status: 'pending', moveInDate: '2024-01-20' },
    { number: '202', floor: 2, area: 25, rent: 1200, tenant: '孙七', status: 'occupied', moveInDate: '2024-04-05' },
    { number: '203', floor: 2, area: 25, rent: 1200, tenant: '周八', status: 'occupied', moveInDate: '2024-02-15' },
    { number: '204', floor: 2, area: 25, rent: 1200, tenant: null, status: 'vacant', moveInDate: null },
    { number: '301', floor: 3, area: 28, rent: 1300, tenant: '吴九', status: 'occupied', moveInDate: '2024-01-05' },
    { number: '302', floor: 3, area: 28, rent: 1300, tenant: '郑十', status: 'occupied', moveInDate: '2024-03-20' },
    { number: '303', floor: 3, area: 28, rent: 1300, tenant: null, status: 'vacant', moveInDate: null },
    { number: '304', floor: 3, area: 28, rent: 1300, tenant: '钱一', status: 'occupied', moveInDate: '2024-02-28' }
])

// 表格列配置
const columns = [
    { key: 'number', label: '房间号' },
    { key: 'floor', label: '楼层' },
    { key: 'area', label: '面积(㎡)' },
    { key: 'rent', label: '租金(元/月)' },
    { key: 'status', label: '状态' },
    { key: 'tenant', label: '租客' },
    { key: 'moveInDate', label: '入住时间', width: '110px' }
]

// 状态配置
const statusConfig: StatusConfig = {
    'occupied': { text: '已出租', class: 'occupied' },
    'vacant': { text: '空置', class: 'vacant' },
    'pending': { text: '待缴费', class: 'pending' }
}

// 楼层选项
const floorOptions = computed(() => {
    const floors = [...new Set(roomsData.value.map(room => room.floor))].sort((a, b) => a - b)
    return floors.map(floor => ({ value: floor, label: `${floor}楼` }))
})

// 状态选项
const statusOptions = [
    { value: 'occupied', label: '已出租' },
    { value: 'vacant', label: '空置' },
    { value: 'pending', label: '待缴费' }
]

// 筛选项配置
const filterItems = computed(() => [
    {
        type: 'select' as const,
        key: 'floor',
        placeholder: '选择楼层',
        options: floorOptions.value
    },
    {
        type: 'select' as const,
        key: 'status',
        placeholder: '选择状态',
        options: statusOptions
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

// 筛选后的房间数据
const filteredRooms = computed(() => {
    return roomsData.value.filter(room => {
        // 楼层筛选
        if (currentFilters.value.floor !== undefined && currentFilters.value.floor !== null && currentFilters.value.floor !== '' && room.floor !== currentFilters.value.floor) {
            return false
        }

        // 状态筛选
        if (currentFilters.value.status !== undefined && currentFilters.value.status !== null && currentFilters.value.status !== '' && room.status !== currentFilters.value.status) {
            return false
        }

        // 入住时间范围筛选
        if (currentFilters.value.moveInDateStart || currentFilters.value.moveInDateEnd) {
            if (!room.moveInDate) {
                return false
            }

            const moveInDate = new Date(room.moveInDate)

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
    nextTick(() => {
        if (filterSectionRef.value) {
            filterSectionRef.value.updateFilteredCount(filteredRooms.value.length)
        }
    })
}

// 处理重置筛选
const handleFilterReset = () => {
    currentFilters.value = {}
}

// 模态框状态管理
const showModal = ref(false)
const currentRoom = ref<Room | null>(null)
const modalMode = ref<'view' | 'edit'>('view')

// 计算属性：模态框标题
const modalTitle = computed(() => {
    if (modalMode.value === 'edit') {
        return currentRoom.value ? '编辑房间' : '新增房间'
    }
    return '房间详情'
})

// 处理编辑房间
const handleEdit = (room: Room) => {
    currentRoom.value = room
    modalMode.value = 'edit'
    showModal.value = true
}

// 处理查看房间详情
const handleView = (room: Room) => {
    currentRoom.value = room
    modalMode.value = 'view'
    showModal.value = true
}

// 处理删除房间
const handleDelete = (room: Room) => {
    if (confirm(`确定要删除房间 ${room.number} 吗？`)) {
        const index = roomsData.value.findIndex(r => r.number === room.number)
        if (index > -1) {
            roomsData.value.splice(index, 1)
        }
    }
}

// 显示添加房间模态框
const showAddRoomModal = () => {
    currentRoom.value = null
    modalMode.value = 'edit'
    showModal.value = true
}

// 关闭模态框
const handleCloseModal = () => {
    showModal.value = false
    currentRoom.value = null
}

// 保存房间信息
const handleSaveRoom = (room: Room) => {
    if (currentRoom.value) {
        // 编辑模式：更新现有房间
        const index = roomsData.value.findIndex(r => r.number === currentRoom.value!.number)
        if (index > -1) {
            roomsData.value[index] = { ...room }
        }
    } else {
        // 新增模式：检查房间号是否重复
        const exists = roomsData.value.some(r => r.number === room.number)
        if (exists) {
            alert('房间号已存在，请使用其他房间号')
            return
        }
        roomsData.value.push({ ...room })
    }

    handleCloseModal()
}
</script>

<template>
    <div class="rooms">
        <div class="section-header">
            <h2 class="section-title">🏠 房间管理</h2>
            <button class="btn btn-primary" @click="showAddRoomModal">
                ➕ 添加房间
            </button>
        </div>

        <!-- 筛选区域 -->
        <FilterSection ref="filterSectionRef" :filter-items="filterItems" :total-count="roomsData.length" item-name="房间"
            @filter-change="handleFilterChange" @reset="handleFilterReset" />

        <DataTable :columns="columns" :data="filteredRooms" :show-actions="true" @edit="handleEdit" @view="handleView"
            @delete="handleDelete">
            <template #floor="{ value }">
                {{ value }}楼
            </template>

            <template #rent="{ value }">
                ¥{{ value.toLocaleString('zh-CN') }}
            </template>

            <template #status="{ value }">
                <StatusBadge :status="value" :status-config="statusConfig" />
            </template>

            <template #tenant="{ value }">
                {{ value || '无' }}
            </template>

            <template #moveInDate="{ value }">
                {{ value || '空置' }}
            </template>
        </DataTable>

        <!-- 房间模态框 -->
        <BaseModal :show="showModal" :title="modalTitle" :show-footer="false" @close="handleCloseModal">
            <template #body>
                <RoomForm :room="currentRoom" :mode="modalMode" @save="handleSaveRoom" />
            </template>
        </BaseModal>
    </div>
</template>

<style scoped>
.rooms {
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