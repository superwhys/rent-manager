<script setup lang="ts">
import type { RoomInfo, FeeCalculationResult } from '@/types'
import { ref, watch, computed } from 'vue'

// 定义组件属性
interface Props {
    roomData: RoomInfo | null
    mode: 'view' | 'edit'
    hasReading?: boolean // 是否已抄表
    hasBill?: boolean // 是否已有账单
}

// 定义事件
interface Emits {
    (e: 'generateBill', roomData: RoomInfo, result: FeeCalculationResult): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

// 费用配置的默认值
const defaultFeeConfig = {
    baseRent: 0,
    networkFee: 0,
    elevatorFee: 0,
    otherFee: 0,
    waterPrice: 0,
    waterMinimum: 0,
    electricPrice: 0,
    electricMinimum: 0,
    gasPrice: 0,
    gasMinimum: 0
}

// 本地编辑副本
const localFeeConfig = ref({ ...defaultFeeConfig })

// 本地读数状态
const localReadings = ref({
    waterReading: 0,
    electricReading: 0,
    gasReading: 0
})

// 计算结果
const calculationResult = ref<FeeCalculationResult>({
    waterTotal: 0,
    electricTotal: 0,
    gasTotal: 0,
    grandTotal: 0
})

const isEditMode = computed(() => props.mode === 'edit')

const buttonText = computed(() => {
    return props.hasBill ? '保存账单' : '生成账单'
})

// 计算属性：获取抄表读数
const readingData = computed(() => {
    return props.roomData?.utilityReading || {
        date: new Date().toISOString().split('T')[0],
        lastWaterReading: 0,
        lastElectricReading: 0,
        lastGasReading: 0,
        waterReading: 0,
        electricReading: 0,
        gasReading: 0,
        waterUsage: 0,
        electricUsage: 0,
        gasUsage: 0
    }
})

// 计算总费用
const calculateTotal = () => {
    const originalReading = readingData.value
    const config = localFeeConfig.value
    const currentReadings = localReadings.value

    // 计算水费
    const waterUsage = Math.max(0, currentReadings.waterReading - originalReading.lastWaterReading)
    const waterCost = waterUsage * config.waterPrice
    const waterTotal = Math.max(waterCost, config.waterMinimum)

    // 计算电费
    const electricUsage = Math.max(0, currentReadings.electricReading - originalReading.lastElectricReading)
    const electricCost = electricUsage * config.electricPrice
    const electricTotal = Math.max(electricCost, config.electricMinimum)

    // 计算气费
    const gasUsage = Math.max(0, currentReadings.gasReading - originalReading.lastGasReading)
    const gasCost = gasUsage * config.gasPrice
    const gasTotal = Math.max(gasCost, config.gasMinimum)

    // 计算总计
    const grandTotal = config.baseRent + config.networkFee +
        config.elevatorFee + config.otherFee +
        waterTotal + electricTotal + gasTotal

    calculationResult.value = {
        waterTotal,
        electricTotal,
        gasTotal,
        grandTotal
    }
}

// 当房间数据变化时，更新本地编辑副本
watch(() => props.roomData?.feeConfig, (newFeeConfig) => {
    if (newFeeConfig) {
        localFeeConfig.value = { ...newFeeConfig }
    } else {
        localFeeConfig.value = { ...defaultFeeConfig }
    }
}, { immediate: true })

// 当房间读数数据变化时，更新本地读数
watch(() => props.roomData?.utilityReading, (newReading) => {
    if (newReading) {
        localReadings.value = {
            waterReading: newReading.waterReading,
            electricReading: newReading.electricReading,
            gasReading: newReading.gasReading
        }
    } else {
        localReadings.value = {
            waterReading: 0,
            electricReading: 0,
            gasReading: 0
        }
    }
}, { immediate: true })

// 监听本地配置和读数变化，重新计算费用
watch(() => [localFeeConfig.value, localReadings.value, props.roomData?.utilityReading], () => {
    calculateTotal()
}, { deep: true, immediate: true })

// 生成账单
const handleGenerateBill = () => {
    calculateTotal()

    if (!props.roomData) return

    const originalReading = readingData.value
    const waterUsage = Math.max(0, localReadings.value.waterReading - originalReading.lastWaterReading)
    const electricUsage = Math.max(0, localReadings.value.electricReading - originalReading.lastElectricReading)
    const gasUsage = Math.max(0, localReadings.value.gasReading - originalReading.lastGasReading)

    // 将本地编辑的费用配置和读数合并到房间数据中
    const updatedRoomData: RoomInfo = {
        ...props.roomData,
        feeConfig: { ...localFeeConfig.value },
        utilityReading: {
            date: originalReading.date || new Date().toISOString().split('T')[0],
            lastWaterReading: originalReading.lastWaterReading,
            lastElectricReading: originalReading.lastElectricReading,
            lastGasReading: originalReading.lastGasReading,
            waterReading: localReadings.value.waterReading,
            electricReading: localReadings.value.electricReading,
            gasReading: localReadings.value.gasReading,
            waterUsage,
            electricUsage,
            gasUsage
        }
    }

    emit('generateBill', updatedRoomData, { ...calculationResult.value })
}

// 格式化货币显示
const formatCurrency = (amount: number) => {
    return amount.toFixed(2)
}
</script>

<template>
    <div class="room-details">
        <!-- 抄表状态提示 -->
        <div class="reading-status">
            <div v-if="props.hasReading" class="status-item status-success">
                ✅ 当月已抄表，数据已更新
            </div>
            <div v-else class="status-item status-warning">
                ⚠️ 当月尚未抄表，保存账单后将自动同步抄表记录
            </div>
        </div>

        <div class="details-columns">
            <div class="details-section basic-fees">
                <h3 class="details-title">💰 基础费用</h3>

                <div class="form-group">
                    <label>基础租金 (元)</label>
                    <el-input-number v-if="isEditMode" v-model="localFeeConfig.baseRent" class="form-input" :min="0"
                        :step="50" controls-position="right" size="default" />
                    <div v-else class="form-value">¥{{ formatCurrency(localFeeConfig.baseRent) }}</div>
                </div>

                <div class="form-group">
                    <label>网络费 (元)</label>
                    <el-input-number v-if="isEditMode" v-model="localFeeConfig.networkFee" class="form-input" :min="0"
                        :step="10" controls-position="right" size="default" />
                    <div v-else class="form-value">¥{{ formatCurrency(localFeeConfig.networkFee) }}</div>
                </div>

                <div class="form-group">
                    <label>电梯及楼道灯费 (元)</label>
                    <el-input-number v-if="isEditMode" v-model="localFeeConfig.elevatorFee" class="form-input" :min="0"
                        :step="10" controls-position="right" size="default" />
                    <div v-else class="form-value">¥{{ formatCurrency(localFeeConfig.elevatorFee) }}</div>
                </div>

                <div class="form-group">
                    <label>其他费用 (元)</label>
                    <el-input-number v-if="isEditMode" v-model="localFeeConfig.otherFee" class="form-input" :min="0"
                        :step="10" controls-position="right" size="default" />
                    <div v-else class="form-value">¥{{ formatCurrency(localFeeConfig.otherFee) }}</div>
                </div>
            </div>

            <div class="details-section water-fees">
                <h3 class="details-title">💧 水费</h3>

                <div class="form-group">
                    <label>上月用水 (方)</label>
                    <div class="form-value readonly">{{ formatCurrency(readingData.lastWaterReading) }}方</div>
                </div>

                <div class="form-group">
                    <label>本月用水 (方)</label>
                    <el-input-number v-if="isEditMode" v-model="localReadings.waterReading" class="form-input" :min="0"
                        :step="0.01" :precision="2" controls-position="right" size="default" />
                    <div v-else class="form-value">{{ formatCurrency(localReadings.waterReading) }}方</div>
                </div>

                <div class="form-group">
                    <label>水费单价 (元/方)</label>
                    <el-input-number v-if="isEditMode" v-model="localFeeConfig.waterPrice" class="form-input" :min="0"
                        :step="0.01" :precision="2" controls-position="right" size="default" />
                    <div v-else class="form-value">¥{{ formatCurrency(localFeeConfig.waterPrice) }}/方</div>
                </div>

                <div class="form-group">
                    <label>水费保底 (元)</label>
                    <el-input-number v-if="isEditMode" v-model="localFeeConfig.waterMinimum" class="form-input" :min="0"
                        :step="1" controls-position="right" size="default" />
                    <div v-else class="form-value">¥{{ formatCurrency(localFeeConfig.waterMinimum) }}</div>
                </div>
            </div>

            <div class="details-section electric-fees">
                <h3 class="details-title">⚡ 电费</h3>

                <div class="form-group">
                    <label>上月用电 (度)</label>
                    <div class="form-value readonly">{{ formatCurrency(readingData.lastElectricReading) }}度</div>
                </div>

                <div class="form-group">
                    <label>本月用电 (度)</label>
                    <el-input-number v-if="isEditMode" v-model="localReadings.electricReading" class="form-input"
                        :min="0" :step="0.01" :precision="2" controls-position="right" size="default" />
                    <div v-else class="form-value">{{ formatCurrency(localReadings.electricReading) }}度</div>
                </div>

                <div class="form-group">
                    <label>电费单价 (元/度)</label>
                    <el-input-number v-if="isEditMode" v-model="localFeeConfig.electricPrice" class="form-input"
                        :min="0" :step="0.01" :precision="2" controls-position="right" size="default" />
                    <div v-else class="form-value">¥{{ formatCurrency(localFeeConfig.electricPrice) }}/度</div>
                </div>

                <div class="form-group">
                    <label>电费保底 (元)</label>
                    <el-input-number v-if="isEditMode" v-model="localFeeConfig.electricMinimum" class="form-input"
                        :min="0" :step="1" controls-position="right" size="default" />
                    <div v-else class="form-value">¥{{ formatCurrency(localFeeConfig.electricMinimum) }}</div>
                </div>
            </div>

            <div class="details-section gas-fees">
                <h3 class="details-title">🔥 燃气费</h3>

                <div class="form-group">
                    <label>上月用气 (方)</label>
                    <div class="form-value readonly">{{ formatCurrency(readingData.lastGasReading) }}方</div>
                </div>

                <div class="form-group">
                    <label>本月用气 (方)</label>
                    <el-input-number v-if="isEditMode" v-model="localReadings.gasReading" class="form-input" :min="0"
                        :step="0.01" :precision="2" controls-position="right" size="default" />
                    <div v-else class="form-value">{{ formatCurrency(localReadings.gasReading) }}方</div>
                </div>

                <div class="form-group">
                    <label>燃气单价 (元/方)</label>
                    <el-input-number v-if="isEditMode" v-model="localFeeConfig.gasPrice" class="form-input" :min="0"
                        :step="0.01" :precision="2" controls-position="right" size="default" />
                    <div v-else class="form-value">¥{{ formatCurrency(localFeeConfig.gasPrice) }}/方</div>
                </div>

                <div class="form-group">
                    <label>燃气费保底 (元)</label>
                    <el-input-number v-if="isEditMode" v-model="localFeeConfig.gasMinimum" class="form-input" :min="0"
                        :step="1" controls-position="right" size="default" />
                    <div v-else class="form-value">¥{{ formatCurrency(localFeeConfig.gasMinimum) }}</div>
                </div>
            </div>
        </div>

        <!-- 计算结果和操作按钮 -->
        <div class="calculation-result">
            <div class="result-summary">
                <div class="cost-breakdown">
                    <div class="cost-item">💧 水费: ¥{{ formatCurrency(calculationResult.waterTotal) }}</div>
                    <div class="cost-item">⚡ 电费: ¥{{ formatCurrency(calculationResult.electricTotal) }}</div>
                    <div class="cost-item">🔥 气费: ¥{{ formatCurrency(calculationResult.gasTotal) }}</div>
                    <div class="total-amount">💰 总计: ¥{{ formatCurrency(calculationResult.grandTotal) }}</div>
                </div>

                <div v-if="isEditMode" class="action-buttons">
                    <button class="btn btn-success" @click="handleGenerateBill">{{ buttonText }}</button>
                </div>
            </div>
        </div>
    </div>
</template>

<style scoped>
.reading-status {
    margin-bottom: 20px;
    padding: 12px;
    border-radius: 8px;
    border: 2px solid;
}

.status-item {
    font-size: 14px;
    font-weight: 500;
    text-align: center;
}

.status-success {
    background: #f0fdf4;
    border-color: #22c55e;
    color: #166534;
}

.status-warning {
    background: #fffbeb;
    border-color: #f59e0b;
    color: #92400e;
}

.details-columns {
    display: grid;
    grid-template-columns: repeat(4, 1fr);
    gap: 15px;
    margin-bottom: 15px;
}

.details-section {
    padding: 12px;
    background: #f8fafc;
    border-radius: 8px;
    border: 2px solid #e2e8f0;
    min-height: 250px;
}

.basic-fees {
    border-color: #4f46e5;
}

.water-fees {
    border-color: #06b6d4;
}

.electric-fees {
    border-color: #f59e0b;
}

.gas-fees {
    border-color: #ef4444;
}

.details-title {
    font-size: 14px;
    font-weight: 600;
    color: #1f2937;
    margin-bottom: 10px;
    text-align: center;
}

.form-group {
    display: flex;
    flex-direction: column;
    margin-bottom: 8px;
}

.form-group label {
    font-size: 12px;
    font-weight: 500;
    color: #374151;
    margin-bottom: 3px;
}

.form-input {
    width: 100%;
}

.form-value {
    padding: 6px 0;
    font-size: 13px;
    color: #374151;
    min-height: 18px;
}

.form-value.readonly {
    background: #f9fafb;
    padding: 6px 8px;
    border-radius: 4px;
    border: 1px solid #e5e7eb;
    color: #6b7280;
    font-weight: 500;
}

.calculation-result {
    background: white;
    border-radius: 8px;
    border: 2px solid #059669;
    margin-top: 15px;
}

.result-summary {
    padding: 15px;
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 20px;
}

.cost-breakdown {
    display: flex;
    gap: 20px;
    align-items: center;
}

.cost-item {
    font-size: 14px;
    color: #4b5563;
    white-space: nowrap;
}

.total-amount {
    font-size: 16px;
    font-weight: 600;
    color: #059669;
    margin-left: 20px;
    padding-left: 20px;
    border-left: 2px solid #d1d5db;
}

.action-buttons {
    display: flex;
    gap: 8px;
    flex-shrink: 0;
}

.btn {
    padding: 6px 12px;
    border: none;
    border-radius: 6px;
    font-size: 13px;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.2s;
    white-space: nowrap;
}

.btn-primary {
    background: #4f46e5;
    color: white;
}

.btn-primary:hover {
    background: #4338ca;
}

.btn-success {
    background: #10b981;
    color: white;
}

.btn-success:hover {
    background: #059669;
}

.btn-secondary {
    background: #6b7280;
    color: white;
}

.btn-secondary:hover {
    background: #4b5563;
}

/* 响应式设计 */
@media (max-width: 1200px) {
    .details-columns {
        grid-template-columns: repeat(2, 1fr);
        gap: 12px;
    }

    .details-section {
        min-height: auto;
        padding: 10px;
    }
}

@media (max-width: 960px) {
    .result-summary {
        flex-direction: column;
        gap: 15px;
        align-items: stretch;
    }

    .cost-breakdown {
        flex-direction: column;
        gap: 8px;
    }

    .total-amount {
        margin-left: 0;
        padding-left: 0;
        border-left: none;
        border-top: 2px solid #d1d5db;
        padding-top: 10px;
        text-align: center;
    }

    .action-buttons {
        justify-content: center;
    }
}

@media (max-width: 768px) {
    .details-columns {
        grid-template-columns: 1fr;
        gap: 10px;
    }

    .result-summary {
        flex-direction: column;
        gap: 15px;
        align-items: stretch;
    }

    .cost-breakdown {
        flex-direction: column;
        gap: 8px;
    }

    .total-amount {
        margin-left: 0;
        padding-left: 0;
        border-left: none;
        border-top: 2px solid #d1d5db;
        padding-top: 10px;
        text-align: center;
    }

    .action-buttons {
        justify-content: center;
    }

    .details-section {
        padding: 12px;
        min-height: auto;
    }
}
</style>