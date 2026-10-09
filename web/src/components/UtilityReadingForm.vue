<script setup lang="ts">
import type { UtilityReading } from '@/types'
import { ref, watch, computed } from 'vue'

// 定义组件属性
interface Props {
    reading: UtilityReading | null
    mode: 'view' | 'edit'
}

// 定义事件
interface Emits {
    (e: 'save', reading: UtilityReading): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

// 表单数据
const formData = ref<UtilityReading>({
    room: '',
    date: '',
    waterReading: 0,
    electricReading: 0,
    gasReading: 0,
    waterUsage: 0,
    electricUsage: 0,
    gasUsage: 0
})

// 计算属性：是否为编辑模式
const isEditMode = computed(() => props.mode === 'edit')

// 监听抄表记录数据变化，更新表单
watch(() => props.reading, (newReading) => {
    if (newReading) {
        formData.value = { ...newReading }
    } else {
        // 重置表单，设置默认日期为今天
        const today = new Date().toISOString().split('T')[0]

        formData.value = {
            room: '',
            date: today,
            waterReading: 0,
            electricReading: 0,
            gasReading: 0,
            waterUsage: 0,
            electricUsage: 0,
            gasUsage: 0
        }
    }
}, { immediate: true })

// 保存抄表记录信息
const handleSave = () => {
    // 基本验证
    if (!formData.value.room.trim()) {
        alert('请输入房间号')
        return
    }
    if (!formData.value.date) {
        alert('请选择抄表日期')
        return
    }
    if (formData.value.waterReading < 0) {
        alert('水表读数不能为负数')
        return
    }
    if (formData.value.electricReading < 0) {
        alert('电表读数不能为负数')
        return
    }
    if (formData.value.gasReading < 0) {
        alert('气表读数不能为负数')
        return
    }
    if (formData.value.waterUsage < 0) {
        alert('用水量不能为负数')
        return
    }
    if (formData.value.electricUsage < 0) {
        alert('用电量不能为负数')
        return
    }
    if (formData.value.gasUsage < 0) {
        alert('用气量不能为负数')
        return
    }

    emit('save', { ...formData.value })
}

// 格式化数值显示
const formatNumber = (value: number, unit: string) => {
    return `${value.toFixed(1)}${unit}`
}
</script>

<template>
    <form @submit.prevent="handleSave">
        <div class="form-grid">
            <!-- 房间号 -->
            <div class="form-group">
                <label class="form-label">房间号</label>
                <el-input v-if="isEditMode" v-model="formData.room" class="form-input" placeholder="请输入房间号"
                    size="large" />
                <div v-else class="form-value">
                    <strong>{{ formData.room || '无' }}</strong>
                </div>
            </div>

            <!-- 抄表日期 -->
            <div class="form-group">
                <label class="form-label">抄表日期</label>
                <el-date-picker v-if="isEditMode" v-model="formData.date" type="date" class="form-input"
                    placeholder="请选择抄表日期" value-format="YYYY-MM-DD" teleported popper-class="high-z-index-popper"
                    size="large" />
                <div v-else class="form-value">
                    {{ formData.date || '无' }}
                </div>
            </div>

            <!-- 水表读数 -->
            <div class="form-group">
                <label class="form-label">水表读数</label>
                <el-input-number v-if="isEditMode" v-model="formData.waterReading" class="form-input" :min="0"
                    :step="0.1" :precision="1" controls-position="right" size="large" />
                <div v-else class="form-value">
                    {{ formatNumber(formData.waterReading, '') }}
                </div>
            </div>

            <!-- 电表读数 -->
            <div class="form-group">
                <label class="form-label">电表读数</label>
                <el-input-number v-if="isEditMode" v-model="formData.electricReading" class="form-input" :min="0"
                    :step="0.1" :precision="1" controls-position="right" size="large" />
                <div v-else class="form-value">
                    {{ formatNumber(formData.electricReading, '') }}
                </div>
            </div>

            <!-- 气表读数 -->
            <div class="form-group">
                <label class="form-label">气表读数</label>
                <el-input-number v-if="isEditMode" v-model="formData.gasReading" class="form-input" :min="0" :step="0.1"
                    :precision="1" controls-position="right" size="large" />
                <div v-else class="form-value">
                    {{ formatNumber(formData.gasReading, '') }}
                </div>
            </div>

            <!-- 用水量 -->
            <div class="form-group">
                <label class="form-label">用水量 (方)</label>
                <el-input-number v-if="isEditMode" v-model="formData.waterUsage" class="form-input" :min="0" :step="0.1"
                    :precision="1" controls-position="right" size="large" />
                <div v-else class="form-value">
                    <span class="usage-text water">{{ formatNumber(formData.waterUsage, '方') }}</span>
                </div>
            </div>

            <!-- 用电量 -->
            <div class="form-group">
                <label class="form-label">用电量 (度)</label>
                <el-input-number v-if="isEditMode" v-model="formData.electricUsage" class="form-input" :min="0"
                    :step="0.1" :precision="1" controls-position="right" size="large" />
                <div v-else class="form-value">
                    <span class="usage-text electric">{{ formatNumber(formData.electricUsage, '度') }}</span>
                </div>
            </div>

            <!-- 用气量 -->
            <div class="form-group">
                <label class="form-label">用气量 (方)</label>
                <el-input-number v-if="isEditMode" v-model="formData.gasUsage" class="form-input" :min="0" :step="0.1"
                    :precision="1" controls-position="right" size="large" />
                <div v-else class="form-value">
                    <span class="usage-text gas">{{ formatNumber(formData.gasUsage, '方') }}</span>
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

.form-input {
    width: 100%;
}

/* 表单值显示样式 */
.form-value {
    padding: 10px 0;
    font-size: 14px;
    color: #374151;
    min-height: 20px;
}

.form-group :deep(.el-date-editor) {
    --el-date-editor-width: 100%;
}

/* 用量显示样式 */
.usage-text {
    padding: 4px 8px;
    border-radius: 4px;
    font-weight: 500;
    font-size: 13px;
}

.usage-text.water {
    background: #dbeafe;
    color: #1e40af;
}

.usage-text.electric {
    background: #fef3c7;
    color: #d97706;
}

.usage-text.gas {
    background: #fce7f3;
    color: #be185d;
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