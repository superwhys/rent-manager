<script setup lang="ts">
import { ref, computed, watch } from 'vue'

/**
 * 通用筛选组件
 * 支持多种筛选器类型和自定义配置
 */

// 筛选项类型定义
interface FilterItem {
    type: 'select' | 'date-range' | 'month-range' | 'month' | 'input'
    key: string
    placeholder?: string
    startPlaceholder?: string // 日期范围开始占位符
    endPlaceholder?: string   // 日期范围结束占位符
    options?: Array<{ value: any, label: string }>
    width?: string
}

interface Props {
    filterItems: FilterItem[]
    totalCount: number
    itemName?: string
}

interface Emits {
    (e: 'filter-change', filters: Record<string, any>): void
    (e: 'reset'): void
}

const props = withDefaults(defineProps<Props>(), {
    itemName: '条目'
})

const emit = defineEmits<Emits>()

// 筛选条件
const filterValues = ref<Record<string, any>>({})

// 初始化筛选条件
const initFilters = () => {
    const initialFilters: Record<string, any> = {}
    props.filterItems.forEach(item => {
        if (item.type === 'date-range' || item.type === 'month-range') {
            initialFilters[`${item.key}Start`] = ''
            initialFilters[`${item.key}End`] = ''
        } else {
            initialFilters[item.key] = ''
        }
    })
    filterValues.value = initialFilters
}

// 筛选后的数量（通过父组件传入）
const filteredCount = ref(props.totalCount)

// 计算是否有筛选条件
const hasFilters = computed(() => {
    return Object.values(filterValues.value).some(value => value)
})

// 重置筛选条件
const resetFilters = () => {
    isInitialized.value = false  // 暂时禁用事件触发
    initFilters()
    isInitialized.value = true   // 重新启用事件触发
    emit('reset')
}

// 是否已初始化标记
const isInitialized = ref(false)

// 监听筛选条件变化 - 避免初始化时触发
watch(filterValues, (newValues) => {
    if (isInitialized.value) {
        emit('filter-change', newValues)
    }
}, { deep: true })

// 初始化
initFilters()
// 标记初始化完成
isInitialized.value = true

// 更新筛选结果数量
const updateFilteredCount = (count: number) => {
    filteredCount.value = count
}

// 暴露方法给父组件
defineExpose({
    updateFilteredCount
})
</script>

<template>
    <div class="search-section">
        <div class="search-form">
            <!-- 渲染不同类型的筛选器 -->
            <template v-for="item in filterItems" :key="item.key">
                <!-- 选择器类型 -->
                <el-select v-if="item.type === 'select'" v-model="filterValues[item.key]"
                    :placeholder="item.placeholder" clearable :style="{ width: item.width || '120px' }">
                    <el-option v-for="option in item.options" :key="option.value" :label="option.label"
                        :value="option.value" />
                </el-select>

                <!-- 日期范围类型 -->
                <template v-else-if="item.type === 'date-range'">
                    <el-date-picker v-model="filterValues[`${item.key}Start`]" type="date"
                        :placeholder="item.startPlaceholder" format="YYYY-MM-DD" value-format="YYYY-MM-DD"
                        :style="{ width: item.width || '140px' }" />
                    <el-date-picker v-model="filterValues[`${item.key}End`]" type="date"
                        :placeholder="item.endPlaceholder" format="YYYY-MM-DD" value-format="YYYY-MM-DD"
                        :style="{ width: item.width || '140px' }" />
                </template>

                <!-- 月份范围类型 -->
                <template v-else-if="item.type === 'month-range'">
                    <el-date-picker v-model="filterValues[`${item.key}Start`]" type="month"
                        :placeholder="item.startPlaceholder" format="YYYY-MM" value-format="YYYY-MM"
                        :style="{ width: item.width || '140px' }" />
                    <el-date-picker v-model="filterValues[`${item.key}End`]" type="month"
                        :placeholder="item.endPlaceholder" format="YYYY-MM" value-format="YYYY-MM"
                        :style="{ width: item.width || '140px' }" />
                </template>

                <!-- 月份选择类型 -->
                <template v-else-if="item.type === 'month'">
                    <el-date-picker v-model="filterValues[item.key]" type="month"
                        :placeholder="item.placeholder || '选择年月'" format="YYYY-MM" value-format="YYYY-MM"
                        :style="{ width: item.width || '140px' }" />
                </template>

                <!-- 输入框类型 -->
                <el-input v-else-if="item.type === 'input'" v-model="filterValues[item.key]"
                    :placeholder="item.placeholder" clearable :style="{ width: item.width || '200px' }" />
            </template>

            <button class="btn btn-secondary" @click="resetFilters">重置</button>
        </div>

        <!-- 筛选结果统计 -->
        <div class="filter-summary">
            <span class="summary-text">
                共找到 <strong>{{ filteredCount }}</strong> 个{{ itemName }}
                <template v-if="hasFilters && filteredCount !== totalCount">
                    （从 {{ totalCount }} 个{{ itemName }}中筛选）
                </template>
            </span>
        </div>
    </div>
</template>

<style scoped>
.search-section {
    margin-bottom: 30px;
}

.search-form {
    display: flex;
    gap: 15px;
    align-items: center;
    flex-wrap: wrap;
    padding: 20px;
    background: rgba(255, 255, 255, 0.95);
    backdrop-filter: blur(10px);
    border-radius: 12px;
    box-shadow: 0 4px 15px rgba(0, 0, 0, 0.1);
}

.filter-summary {
    padding: 15px 20px 0;
    border-top: 1px solid rgba(0, 0, 0, 0.1);
    margin-top: 15px;
}

.summary-text {
    color: #6b7280;
    font-size: 14px;
}

.summary-text strong {
    color: #4f46e5;
    font-weight: 600;
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

.btn-secondary {
    background: #6b7280;
    color: white;
}

.btn-secondary:hover {
    background: #4b5563;
    transform: translateY(-2px);
}

/* 响应式设计 */
@media (max-width: 768px) {
    .search-form {
        flex-direction: column;
        align-items: stretch;
    }

    .search-form>* {
        width: 100% !important;
    }
}
</style>