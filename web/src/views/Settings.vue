<script setup lang="ts">
import type { SystemSettings, UtilitySettings } from '@/types'
import { ref } from 'vue'
import UtilityCard from '@/components/UtilityCard.vue'

// 系统设置数据
const settings = ref<SystemSettings>({
    billGenerateDate: 1,
    overdueReminderDays: 3,
    autoBackup: 'daily'
})

// 水电设置数据
const utilitySettings = ref<UtilitySettings>({
    waterPrice: 3.5,
    waterMinimum: 15,
    electricPrice: 0.8,
    electricMinimum: 20,
    gasPrice: 2.8,
    gasMinimum: 10
})

// 最后备份时间
const lastBackupTime = ref('2025-01-08 10:30:00')

// 处理操作
const saveAllSettings = () => {
    console.log('保存所有设置:', settings.value)
}

const resetToDefault = () => {
    console.log('重置为默认设置')
}

const backupData = () => {
    console.log('立即备份数据')
    lastBackupTime.value = new Date().toLocaleString('zh-CN')
}

const restoreData = () => {
    console.log('恢复数据')
}

const cleanupData = () => {
    console.log('清理过期数据')
}

// 处理水电设置更新
const updateWaterSettings = (settings: { priceValue: number; minimumValue: number }) => {
    utilitySettings.value.waterPrice = settings.priceValue
    utilitySettings.value.waterMinimum = settings.minimumValue
    console.log('更新水费设置:', settings)
}

const updateElectricSettings = (settings: { priceValue: number; minimumValue: number }) => {
    utilitySettings.value.electricPrice = settings.priceValue
    utilitySettings.value.electricMinimum = settings.minimumValue
    console.log('更新电费设置:', settings)
}

const updateGasSettings = (settings: { priceValue: number; minimumValue: number }) => {
    utilitySettings.value.gasPrice = settings.priceValue
    utilitySettings.value.gasMinimum = settings.minimumValue
    console.log('更新燃气设置:', settings)
}


</script>

<template>
    <div class="settings">
        <h2 class="section-title">⚙️ 系统设置</h2>

        <!-- 水电设置 -->
        <div class="utilities-section">
            <div class="utilities-grid">
                <UtilityCard title="💧 水费设置" :settings="{
                    priceLabel: '水费单价 (元/方)',
                    priceValue: utilitySettings.waterPrice,
                    minimumLabel: '保底费用 (元)',
                    minimumValue: utilitySettings.waterMinimum
                }" @update-settings="updateWaterSettings" />

                <UtilityCard title="⚡ 电费设置" :settings="{
                    priceLabel: '电费单价 (元/度)',
                    priceValue: utilitySettings.electricPrice,
                    minimumLabel: '保底费用 (元)',
                    minimumValue: utilitySettings.electricMinimum
                }" @update-settings="updateElectricSettings" />

                <UtilityCard title="🔥 燃气设置" :settings="{
                    priceLabel: '燃气单价 (元/方)',
                    priceValue: utilitySettings.gasPrice,
                    minimumLabel: '保底费用 (元)',
                    minimumValue: utilitySettings.gasMinimum
                }" @update-settings="updateGasSettings" />
            </div>
        </div>

        <div class="settings-grid">
            <SettingsCard title="📅 系统配置">
                <div class="form-group">
                    <label>账单生成日期</label>
                    <el-select v-model="settings.billGenerateDate" class="form-select" placeholder="请选择账单生成日期"
                        teleported popper-class="high-z-index-popper" size="large">
                        <el-option value="1" label="每月1号" />
                        <el-option value="15" label="每月15号" />
                        <el-option value="30" label="每月30号" />
                    </el-select>
                </div>
                <div class="form-group">
                    <label>逾期提醒天数</label>
                    <el-input-number v-model="settings.overdueReminderDays" :min="1" :max="30" :step="1"
                        class="form-input" controls-position="right" size="large" />
                </div>
                <div class="form-group">
                    <label>自动备份</label>
                    <el-select v-model="settings.autoBackup" class="form-select" placeholder="请选择自动备份频率" teleported
                        popper-class="high-z-index-popper" size="large">
                        <el-option value="daily" label="每日备份" />
                        <el-option value="weekly" label="每周备份" />
                        <el-option value="monthly" label="每月备份" />
                    </el-select>
                </div>
            </SettingsCard>

            <SettingsCard title="💾 数据管理">
                <div class="backup-section">
                    <button class="btn btn-primary full-width" @click="backupData">
                        📥 立即备份数据
                    </button>
                    <button class="btn btn-success full-width" @click="restoreData">
                        📤 恢复数据
                    </button>
                    <button class="btn btn-warning full-width" @click="cleanupData">
                        🗑️ 清理过期数据
                    </button>
                </div>
                <div class="form-group">
                    <label>最后备份时间</label>
                    <el-input :model-value="lastBackupTime" readonly class="form-input" size="large" />
                </div>
            </SettingsCard>
        </div>

        <div class="save-section">
            <button class="btn btn-success btn-large" @click="saveAllSettings">
                💾 保存所有设置
            </button>
            <button class="btn btn-warning btn-large" @click="resetToDefault">
                🔄 重置为默认
            </button>
        </div>
    </div>
</template>

<style scoped>
.settings {
    padding: 20px;
}

.section-title {
    font-size: 24px;
    font-weight: 600;
    margin-bottom: 20px;
    color: #333;
}

.settings-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(350px, 1fr));
    gap: 25px;
    margin-bottom: 30px;
}

.form-group {
    display: flex;
    flex-direction: column;
    margin-bottom: 15px;
}

.form-group label {
    font-size: 14px;
    font-weight: 500;
    margin-bottom: 5px;
    color: #374151;
}


.form-input {
    width: 100%;
}

/* Element Plus 选择器样式 */
.form-select {
    width: 100%;
}

.backup-section {
    display: flex;
    flex-direction: column;
    gap: 10px;
    margin-bottom: 15px;
}



.save-section {
    text-align: center;
    background: white;
    padding: 25px;
    border-radius: 12px;
    box-shadow: 0 2px 10px rgba(0, 0, 0, 0.05);
}

.btn {
    padding: 12px 24px;
    border: none;
    border-radius: 8px;
    font-size: 14px;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.3s ease;
    margin: 5px;
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

.btn-warning {
    background: #f59e0b;
    color: white;
}

.btn-warning:hover {
    background: #d97706;
    transform: translateY(-2px);
}

.btn-large {
    padding: 15px 30px;
    font-size: 16px;
    margin: 0 10px;
}

.btn-small {
    padding: 8px 16px;
    font-size: 12px;
}

.full-width {
    width: 100%;
}

.utilities-section {
    margin-bottom: 30px;
}

.section-subtitle {
    font-size: 18px;
    font-weight: 600;
    margin-bottom: 15px;
    color: #333;
}

.utilities-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
    gap: 20px;
    margin-bottom: 20px;
}

/* 响应式设计 */
@media (max-width: 480px) {
    .settings {
        padding: 10px;
    }
    
    .section-title {
        font-size: 20px;
        margin-bottom: 15px;
    }
    
    .settings-grid {
        grid-template-columns: 1fr;
        gap: 15px;
        margin-bottom: 20px;
    }
    
    .save-section {
        padding: 15px;
        display: flex;
        flex-direction: column;
        gap: 10px;
    }
    
    .btn-large {
        padding: 12px 20px;
        font-size: 14px;
        margin: 0;
        width: 100%;
    }
    
    .form-group {
        margin-bottom: 12px;
    }
}
</style>