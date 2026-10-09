<script setup lang="ts">
// 报表卡片组件
interface Props {
    title: string
    hasChart?: boolean
}

const props = defineProps<Props>()
</script>

<template>
    <div class="report-card">
        <h3>{{ title }}</h3>
        <div v-if="hasChart" class="chart-container">
            <div class="chart-placeholder">
                📊 图表数据可视化区域<br>
                <small>即将上线，敬请期待</small>
            </div>
        </div>
        <slot />
    </div>
</template>

<style scoped>
.report-card {
    background: white;
    padding: 25px;
    border-radius: 12px;
    box-shadow: 0 4px 15px rgba(0, 0, 0, 0.05);
}

.report-card h3 {
    margin-bottom: 20px;
    color: #1f2937;
}

.chart-container {
    height: 200px;
    display: flex;
    align-items: center;
    justify-content: center;
    background: #f8fafc;
    border-radius: 8px;
    border: 2px dashed #d1d5db;
    margin-bottom: 15px;
}

.chart-placeholder {
    color: #64748b;
    font-size: 14px;
    text-align: center;
}
</style> 