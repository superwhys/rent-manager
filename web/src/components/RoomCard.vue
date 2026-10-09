<script setup lang="ts">
import type { RoomInfo, RoomStatus } from '@/types'

interface Props {
    room: RoomInfo
    hasBill?: boolean // 是否已出账
}

// 定义事件
interface Emits {
    (e: 'click', room: RoomInfo): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

// 处理点击事件
const handleClick = () => {
    emit('click', props.room)
}

// 获取房间状态显示文本
const getStatusText = (status: RoomStatus) => {
    switch (status) {
        case 'occupied': return '已出租'
        case 'vacant': return '空置'
        case 'pending': return '待租房'
        default: return '未知'
    }
}

// 获取房间状态样式类
const getStatusClass = (status: RoomStatus) => {
    switch (status) {
        case 'occupied': return 'room-occupied'
        case 'vacant': return 'room-vacant'
        case 'pending': return 'room-pending'
        default: return ''
    }
}

// 获取出账状态文本
const getBillStatusText = (hasBill: boolean) => {
    return hasBill ? '已出账' : '未出账'
}

// 获取出账状态样式类
const getBillStatusClass = (hasBill: boolean) => {
    return hasBill ? 'bill-paid' : 'bill-unpaid'
}
</script>

<template>
    <div :class="['room-card', getStatusClass(room.status)]" @click="handleClick" role="button" tabindex="0"
        @keydown.enter="handleClick" @keydown.space="handleClick">
        <!-- 顶部区域：房间号 + 状态标签 -->
        <div class="card-header">
            <div class="room-number">{{ room.id }}</div>
            <div class="status-badges">
                <div class="status-badge">{{ getStatusText(room.status) }}</div>
                <div :class="['bill-badge', getBillStatusClass(hasBill || false)]">
                    {{ getBillStatusText(hasBill || false) }}
                </div>
            </div>
        </div>
        
        <!-- 分割线 -->
        <div class="divider"></div>
        
        <!-- 租客信息 -->
        <div class="tenant-section">
            <span class="tenant-icon">👤</span>
            <span class="tenant-name">{{ room.tenant }}</span>
        </div>
    </div>
</template>

<style scoped>
.room-card {
    padding: 16px;
    border-radius: 12px;
    box-shadow: 0 4px 15px rgba(0, 0, 0, 0.1);
    transition: transform 0.3s ease;
    cursor: pointer;
    user-select: none;
    color: white;
    position: relative;
    overflow: hidden;
    height: 120px;
    display: flex;
    flex-direction: column;
}

.room-card:hover {
    transform: translateY(-5px);
}

.room-card:focus {
    outline: 2px solid rgba(255, 255, 255, 0.5);
    outline-offset: 2px;
}

.room-card:active {
    transform: translateY(-2px);
}

.card-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    margin-bottom: 8px;
}

.room-number {
    font-size: 28px;
    font-weight: bold;
    line-height: 1;
}

.status-badges {
    display: flex;
    flex-direction: column;
    gap: 4px;
    align-items: flex-end;
}

.status-badge {
    background: rgba(255, 255, 255, 0.2);
    padding: 4px 8px;
    border-radius: 12px;
    font-size: 10px;
    font-weight: 600;
    backdrop-filter: blur(10px);
    border: 1px solid rgba(255, 255, 255, 0.3);
    white-space: nowrap;
}

.bill-badge {
    padding: 4px 8px;
    border-radius: 12px;
    font-size: 10px;
    font-weight: 600;
    backdrop-filter: blur(10px);
    border: 1px solid rgba(255, 255, 255, 0.3);
    white-space: nowrap;
}

.divider {
    width: 100%;
    height: 1px;
    background: rgba(255, 255, 255, 0.3);
    margin: 8px 0;
}

.tenant-section {
    display: flex;
    align-items: center;
    gap: 6px;
    opacity: 0.9;
    font-size: 13px;
    margin-top: auto;
}

.tenant-icon {
    font-size: 14px;
}

.tenant-name {
    font-weight: 500;
}

/* 房间状态样式 - 与StatCard类似的渐变 */
.room-occupied {
    background: linear-gradient(135deg, #11998e 0%, #38ef7d 100%);
}

.room-vacant {
    background: linear-gradient(135deg, #fc4a1a 0%, #f7b733 100%);
}

.room-pending {
    background: linear-gradient(135deg, #f7971e 0%, #ffd200 100%);
}

/* 账单状态样式 */
.bill-paid {
    background: rgba(34, 197, 94, 0.3);
    color: rgba(255, 255, 255, 0.95);
}

.bill-unpaid {
    background: rgba(239, 68, 68, 0.3);
    color: rgba(255, 255, 255, 0.95);
}
</style>