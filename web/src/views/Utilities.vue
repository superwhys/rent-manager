<script setup lang="ts">
import type { UtilityReading } from '@/types'
import { ref, computed, nextTick } from 'vue'
import BaseModal from '@/components/BaseModal.vue'
import UtilityReadingForm from '@/components/UtilityReadingForm.vue'
import FilterSection from '@/components/FilterSection.vue'

// 抄表记录数据
const readingsData = ref<UtilityReading[]>([
    { room: '101', date: '2025-07-01', waterReading: 1245.6, electricReading: 3421.2, gasReading: 892.5, waterUsage: 12.3, electricUsage: 145.8, gasUsage: 8.2 },
    { room: '102', date: '2025-07-01', waterReading: 1198.4, electricReading: 3356.8, gasReading: 856.3, waterUsage: 15.2, electricUsage: 158.4, gasUsage: 9.1 },
    { room: '104', date: '2025-07-01', waterReading: 1156.8, electricReading: 3289.5, gasReading: 834.2, waterUsage: 11.8, electricUsage: 132.6, gasUsage: 7.5 },
    { room: '201', date: '2025-07-01', waterReading: 1289.2, electricReading: 3456.8, gasReading: 912.4, waterUsage: 16.5, electricUsage: 168.9, gasUsage: 10.2 },
    { room: '202', date: '2025-07-01', waterReading: 1234.5, electricReading: 3398.2, gasReading: 879.6, waterUsage: 13.8, electricUsage: 149.2, gasUsage: 8.7 },
    { room: '203', date: '2025-07-01', waterReading: 1167.9, electricReading: 3312.5, gasReading: 845.8, waterUsage: 12.9, electricUsage: 138.4, gasUsage: 7.9 }
])

// 表格列配置
const columns = [
    { key: 'room', label: '房间号', width: '100px' },
    { key: 'date', label: '抄表日期', width: '120px' },
    { key: 'waterReading', label: '水表读数' },
    { key: 'electricReading', label: '电表读数' },
    { key: 'gasReading', label: '气表读数' },
    { key: 'waterUsage', label: '用水量' },
    { key: 'electricUsage', label: '用电量' },
    { key: 'gasUsage', label: '用气量' }
]

// 房间号选项
const roomOptions = computed(() => {
    const rooms = [...new Set(readingsData.value.map(reading => reading.room))].sort()
    return rooms.map(room => ({ value: room, label: `${room}房间` }))
})

// 筛选项配置
const filterItems = computed(() => [
    {
        type: 'select' as const,
        key: 'room',
        placeholder: '选择房间号',
        options: roomOptions.value
    },
    {
        type: 'month' as const,
        key: 'date',
        placeholder: '选择年月'
    }
])

// 当前筛选条件
const currentFilters = ref<Record<string, any>>({})

// 筛选后的抄表记录数据
const filteredReadings = computed(() => {
    return readingsData.value.filter(reading => {
        // 房间号筛选
        if (currentFilters.value.room && reading.room !== currentFilters.value.room) {
            return false
        }

        // 抄表月份筛选
        if (currentFilters.value.date) {
            if (!reading.date) {
                return false
            }

            // 提取抄表记录的年月 (YYYY-MM 格式)
            const readingMonth = reading.date.substring(0, 7)

            // 筛选选中的年月
            if (readingMonth !== currentFilters.value.date) {
                return false
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
            filterSectionRef.value.updateFilteredCount(filteredReadings.value.length)
        }
    })
}

// 处理重置筛选
const handleFilterReset = () => {
    currentFilters.value = {}
}

// 模态框状态管理
const showModal = ref(false)
const currentReading = ref<UtilityReading | null>(null)
const modalMode = ref<'view' | 'edit'>('view')

// 计算属性：模态框标题
const modalTitle = computed(() => {
    if (modalMode.value === 'edit') {
        return currentReading.value ? '编辑抄表记录' : '新增抄表记录'
    }
    return '抄表记录详情'
})

// 处理编辑抄表记录
const handleEdit = (reading: UtilityReading) => {
    currentReading.value = reading
    modalMode.value = 'edit'
    showModal.value = true
}

// 处理查看抄表详情
const handleView = (reading: UtilityReading) => {
    currentReading.value = reading
    modalMode.value = 'view'
    showModal.value = true
}

// 处理删除抄表记录
const handleDelete = (reading: UtilityReading) => {
    if (confirm(`确定要删除房间 ${reading.room} 在 ${reading.date} 的抄表记录吗？`)) {
        const index = readingsData.value.findIndex(r => r.room === reading.room && r.date === reading.date)
        if (index > -1) {
            readingsData.value.splice(index, 1)
        }
    }
}

// 显示录入抄表模态框
const showRecordReadingModal = () => {
    currentReading.value = null
    modalMode.value = 'edit'
    showModal.value = true
}

// 关闭模态框
const handleCloseModal = () => {
    showModal.value = false
    currentReading.value = null
}

// 保存抄表记录
const handleSaveReading = (reading: UtilityReading) => {
    if (currentReading.value) {
        // 编辑模式：更新现有记录
        const index = readingsData.value.findIndex(r => r.room === currentReading.value!.room && r.date === currentReading.value!.date)
        if (index > -1) {
            readingsData.value[index] = { ...reading }
        }
    } else {
        // 新增模式：检查记录是否重复
        const exists = readingsData.value.some(r => r.room === reading.room && r.date === reading.date)
        if (exists) {
            alert('该房间在此日期的抄表记录已存在')
            return
        }
        readingsData.value.push({ ...reading })
    }

    handleCloseModal()
}
</script>

<template>
    <div class="utilities">
        <div class="section-header">
            <h2 class="section-title">📊 抄表管理</h2>
            <button class="btn btn-primary" @click="showRecordReadingModal">
                📊 录入抄表
            </button>
        </div>

        <!-- 抄表记录 -->
        <div class="readings-section">
            <h3 class="section-subtitle">抄表记录</h3>

            <!-- 筛选区域 -->
            <FilterSection ref="filterSectionRef" :filter-items="filterItems" :total-count="readingsData.length"
                item-name="抄表记录" @filter-change="handleFilterChange" @reset="handleFilterReset" />

            <DataTable :columns="columns" :data="filteredReadings" :show-actions="true" @edit="handleEdit"
                @view="handleView" @delete="handleDelete">
                <template #room="{ value }">
                    <strong>{{ value }}</strong>
                </template>

                <template #waterUsage="{ value }">
                    {{ value }}方
                </template>

                <template #electricUsage="{ value }">
                    {{ value }}度
                </template>

                <template #gasUsage="{ value }">
                    {{ value }}方
                </template>
            </DataTable>
        </div>

        <!-- 抄表记录模态框 -->
        <BaseModal :show="showModal" :title="modalTitle" :show-footer="false" @close="handleCloseModal">
            <template #body>
                <UtilityReadingForm :reading="currentReading" :mode="modalMode" @save="handleSaveReading" />
            </template>
        </BaseModal>
    </div>
</template>

<style scoped>
.utilities {
    padding: 20px;
}

.section-title {
    font-size: 24px;
    font-weight: 600;
    margin-bottom: 20px;
    color: #333;
}

.section-subtitle {
    font-size: 18px;
    font-weight: 600;
    margin-bottom: 15px;
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

.readings-section {
    margin-top: 30px;
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