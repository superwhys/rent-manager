<script setup lang="ts">
import type { StatusConfig } from '@/types'

interface Props {
    status: string
    statusConfig: StatusConfig
}

const props = defineProps<Props>()

// 获取状态配置
const getStatusInfo = (status: string) => {
    return props.statusConfig[status] || { text: status, class: 'default' }
}
</script>

<template>
    <span 
        :class="['status-badge', getStatusInfo(status).class]"
    >
        {{ getStatusInfo(status).text }}
    </span>
</template>

<style scoped>
.status-badge {
    padding: 4px 8px;
    border-radius: 20px;
    font-size: 12px;
    font-weight: 500;
    display: inline-block;
}

.status-badge.occupied,
.status-badge.active,
.status-badge.paid {
    background: #dcfce7;
    color: #166534;
}

.status-badge.vacant,
.status-badge.expired,
.status-badge.unpaid {
    background: #fee2e2;
    color: #991b1b;
}

.status-badge.pending {
    background: #fef3c7;
    color: #92400e;
}

.status-badge.default {
    background: #f3f4f6;
    color: #374151;
}
</style> 