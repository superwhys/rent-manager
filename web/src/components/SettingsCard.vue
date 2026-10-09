<script setup lang="ts">
// 设置卡片组件
interface Props {
    title: string
}

const props = defineProps<Props>()
</script>

<template>
    <div class="settings-card">
        <h3>{{ title }}</h3>
        <slot />
    </div>
</template>

<style scoped>
.settings-card {
    background: white;
    padding: 25px;
    border-radius: 12px;
    box-shadow: 0 2px 10px rgba(0, 0, 0, 0.05);
}

.settings-card h3 {
    margin-bottom: 20px;
    color: #1f2937;
    border-bottom: 2px solid #4f46e5;
    padding-bottom: 10px;
}
</style> 