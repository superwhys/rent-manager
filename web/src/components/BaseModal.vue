<script setup lang="ts">
import { watch } from 'vue'

// 定义组件属性
interface Props {
    show: boolean
    title: string
    showFooter?: boolean
    size?: 'small' | 'medium' | 'large'
}

// 定义事件
interface Emits {
    (e: 'close'): void
}

const props = withDefaults(defineProps<Props>(), {
    showFooter: true,
    size: 'medium'
})

const emit = defineEmits<Emits>()

const getSizeClass = () => {
    const sizeMap = {
        small: 'modal-small',
        medium: 'modal-medium',
        large: 'modal-large'
    }
    return sizeMap[props.size]
}

watch(() => props.show, (newShow) => {
    if (newShow) {
        document.body.style.overflow = 'hidden'
    } else {
        document.body.style.overflow = ''
    }
})

const handleClose = () => {
    emit('close')
}
</script>

<template>
    <teleport to="body">
        <transition name="modal" appear>
            <div v-if="show" class="modal-overlay" @click.self="handleClose">
                <transition name="modal-content" appear>
                    <div v-if="show" class="modal-container" :class="getSizeClass()">
                        <!-- 模态框头部 -->
                        <div class="modal-header">
                            <h3 class="modal-title">{{ title }}</h3>
                            <button class="close-btn" @click="handleClose">✕</button>
                        </div>

                        <!-- 模态框主体内容 -->
                        <div class="modal-body">
                            <slot name="body"></slot>
                        </div>

                        <!-- 模态框底部 -->
                        <div v-if="showFooter" class="modal-footer">
                            <slot name="footer">
                                <button class="btn btn-secondary" @click="handleClose">
                                    取消
                                </button>
                            </slot>
                        </div>
                    </div>
                </transition>
            </div>
        </transition>
    </teleport>
</template>

<style scoped>
/* 模态框动画效果 */
.modal-enter-active,
.modal-leave-active {
    transition: opacity 0.3s ease;
}

.modal-enter-from,
.modal-leave-to {
    opacity: 0;
}

.modal-content-enter-active {
    transition: all 0.3s ease;
}

.modal-content-leave-active {
    transition: all 0.2s ease;
}

.modal-content-enter-from {
    opacity: 0;
    transform: scale(0.8) translateY(-20px);
}

.modal-content-leave-to {
    opacity: 0;
    transform: scale(0.95) translateY(10px);
}

/* 模态框覆盖层样式 */
.modal-overlay {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: rgba(0, 0, 0, 0.6);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 9999;
    backdrop-filter: blur(8px);
    -webkit-backdrop-filter: blur(8px);
}

/* 模态框容器样式 */
.modal-container {
    background: white;
    border-radius: 12px;
    width: 90%;
    max-height: 90vh;
    overflow: hidden;
    box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 10px 10px -5px rgba(0, 0, 0, 0.04);
}

/* 模态框尺寸 */
.modal-small {
    max-width: 400px;
}

.modal-medium {
    max-width: 600px;
}

.modal-large {
    max-width: 1200px;
    width: 95%;
}

/* 模态框头部样式 */
.modal-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 20px 24px;
    border-bottom: 1px solid #e5e7eb;
    background: #f9fafb;
}

.modal-title {
    font-size: 18px;
    font-weight: 600;
    color: #111827;
    margin: 0;
}

.close-btn {
    background: none;
    border: none;
    font-size: 20px;
    color: #6b7280;
    cursor: pointer;
    padding: 4px;
    border-radius: 4px;
    transition: all 0.2s;
}

.close-btn:hover {
    background: #e5e7eb;
    color: #374151;
}

/* 模态框主体样式 */
.modal-body {
    padding: 24px;
    max-height: 70vh;
    overflow-y: auto;
}

/* 自定义滚动条样式 */
.modal-body::-webkit-scrollbar {
    width: 8px;
}

.modal-body::-webkit-scrollbar-track {
    background: #f1f1f1;
    border-radius: 4px;
}

.modal-body::-webkit-scrollbar-thumb {
    background: #c1c1c1;
    border-radius: 4px;
}

.modal-body::-webkit-scrollbar-thumb:hover {
    background: #a8a8a8;
}

/* 模态框底部样式 */
.modal-footer {
    display: flex;
    justify-content: flex-end;
    gap: 12px;
    padding: 16px 24px;
    border-top: 1px solid #e5e7eb;
    background: #f9fafb;
}

/* 按钮样式 */
.btn {
    padding: 10px 20px;
    border: none;
    border-radius: 6px;
    font-size: 14px;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.2s;
}

.btn-secondary {
    background: #f3f4f6;
    color: #374151;
}

.btn-secondary:hover {
    background: #e5e7eb;
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
    .modal-container {
        width: 95%;
        margin: 20px;
    }

    .modal-header,
    .modal-body,
    .modal-footer {
        padding: 16px;
    }
}
</style>