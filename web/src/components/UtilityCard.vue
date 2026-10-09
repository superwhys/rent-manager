<script setup lang="ts">
// 水电设置卡片组件
interface Props {
    title: string
    settings: {
        priceLabel: string
        priceValue: number
        minimumLabel: string
        minimumValue: number
    }
}

const props = defineProps<Props>()
const emit = defineEmits<{
    updateSettings: [settings: { priceValue: number; minimumValue: number }]
}>()

// 处理设置更新
const updatePrice = (value: number | undefined) => {
    emit('updateSettings', {
        priceValue: value || 0,
        minimumValue: props.settings.minimumValue
    })
}

const updateMinimum = (value: number | undefined) => {
    emit('updateSettings', {
        priceValue: props.settings.priceValue,
        minimumValue: value || 0
    })
}
</script>

<template>
    <div class="utility-card">
        <h3>{{ title }}</h3>
        <div class="form-group">
            <label>{{ settings.priceLabel }}</label>
            <el-input-number :model-value="settings.priceValue" :min="0" :step="0.01" :precision="2" class="form-input"
                controls-position="right" size="large" @change="updatePrice" />
        </div>
        <div class="form-group">
            <label>{{ settings.minimumLabel }}</label>
            <el-input-number :model-value="settings.minimumValue" :min="0" :step="1" class="form-input"
                controls-position="right" size="large" @change="updateMinimum" />
        </div>
    </div>
</template>

<style scoped>
.utility-card {
    background: white;
    padding: 20px;
    border-radius: 10px;
    box-shadow: 0 2px 10px rgba(0, 0, 0, 0.05);
}

.utility-card h3 {
    margin-bottom: 15px;
    color: #1f2937;
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
</style>