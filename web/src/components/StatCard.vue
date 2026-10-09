<script setup lang="ts">
// 统计卡片组件
interface Props {
    name: string
    value: string | number
    backgroundColor?: string
    textColor?: string
}

const props = defineProps<Props>()

// 格式化数字显示
const formatValue = (value: string | number): string => {
    if (typeof value === 'number') {
        return value.toLocaleString('zh-CN')
    }
    return value
}
</script>

<template>
    <div 
        class="stat-card" 
        :style="{
            background: backgroundColor || 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
            color: textColor || 'white'
        }"
    >
        <div class="stat-value">{{ formatValue(value) }}</div>
        <div class="stat-label">{{ name }}</div>
    </div>
</template>

<style scoped>
.stat-card {
    padding: 25px;
    border-radius: 12px;
    text-align: center;
    box-shadow: 0 4px 15px rgba(0, 0, 0, 0.1);
    transition: transform 0.3s ease;
}

.stat-card:hover {
    transform: translateY(-5px);
}

.stat-value {
    font-size: 32px;
    font-weight: bold;
    margin-bottom: 5px;
}

.stat-label {
    font-size: 14px;
    opacity: 0.9;
}
</style> 