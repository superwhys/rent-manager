<script setup lang="ts">
interface Props {
    to: string
    icon: string
    text: string
}

defineProps<Props>()

// 定义事件发射器
const emit = defineEmits<{
    click: []
}>()

// 处理导航项点击
const handleClick = () => {
    emit('click')
}
</script>

<template>
    <li class="nav-item">
        <router-link :to="to" class="nav-link" @click="handleClick">
            <span class="nav-icon">{{ icon }}</span>
            {{ text }}
        </router-link>
    </li>
</template>

<style scoped>
.nav-item {
    margin-bottom: 8px;
}

.nav-link {
    display: flex;
    align-items: center;
    padding: 12px 15px;
    text-decoration: none;
    color: #64748b;
    border-radius: 8px;
    transition: all 0.3s ease;
    font-weight: 500;
}

.nav-link:hover,
.nav-link.router-link-active {
    background: #4f46e5;
    color: white;
    transform: translateX(5px);
}

.nav-icon {
    margin-right: 10px;
    font-size: 16px;
}
</style>