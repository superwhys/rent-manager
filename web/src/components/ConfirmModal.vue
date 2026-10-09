<script setup lang="ts">
import { computed } from 'vue'

/**
 * 自定义确认框组件
 * 用于替代Element Plus MessageBox，提供更好的样式控制
 */

interface Props {
  // 是否显示模态框
  visible?: boolean
  // 标题
  title?: string
  // 内容
  message?: string
  // 确认按钮文字
  confirmText?: string
  // 取消按钮文字
  cancelText?: string
  // 图标类型
  type?: 'warning' | 'info' | 'success' | 'error'
  // 是否显示关闭按钮
  showClose?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  visible: false,
  title: '确认',
  message: '',
  confirmText: '确定',
  cancelText: '取消',
  type: 'warning',
  showClose: true
})

const emit = defineEmits<{
  confirm: []
  cancel: []
  close: []
}>()

// 图标映射
const iconMap = {
  warning: '⚠️',
  info: 'ℹ️',
  success: '✅',
  error: '❌'
}

// 计算图标
const icon = computed(() => iconMap[props.type])

// 处理确认
const handleConfirm = () => {
  emit('confirm')
}

// 处理取消
const handleCancel = () => {
  emit('cancel')
}

// 处理关闭
const handleClose = () => {
  emit('close')
}

// 处理遮罩层点击
const handleMaskClick = () => {
  emit('cancel')
}
</script>

<template>
  <!-- 遮罩层 -->
  <Transition name="mask">
    <div 
      v-if="visible" 
      class="confirm-modal-mask"
      @click="handleMaskClick"
    >
      <!-- 模态框容器 -->
      <Transition name="modal">
        <div 
          v-if="visible"
          class="confirm-modal"
          @click.stop
        >
          <!-- 关闭按钮 -->
          <button 
            v-if="showClose"
            class="close-btn"
            @click="handleClose"
          >
            ✕
          </button>

          <!-- 头部 -->
          <div class="modal-header">
            <div class="icon-wrapper">
              {{ icon }}
            </div>
            <h3 class="modal-title">{{ title }}</h3>
          </div>

          <!-- 内容 -->
          <div class="modal-content">
            <p class="modal-message">{{ message }}</p>
          </div>

          <!-- 底部按钮 -->
          <div class="modal-footer">
            <button 
              class="btn btn-cancel"
              @click="handleCancel"
            >
              {{ cancelText }}
            </button>
            <button 
              class="btn btn-confirm"
              @click="handleConfirm"
            >
              {{ confirmText }}
            </button>
          </div>
        </div>
      </Transition>
    </div>
  </Transition>
</template>

<style scoped>
/* 遮罩层 */
.confirm-modal-mask {
  position: fixed;
  top: 0;
  left: 0;
  width: 100vw;
  height: 100vh;
  background: rgba(0, 0, 0, 0.5);
  backdrop-filter: blur(4px);
  z-index: 4000;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 20px;
}

/* 模态框 */
.confirm-modal {
  background: rgba(255, 255, 255, 0.98);
  border-radius: 20px;
  box-shadow: 0 25px 50px rgba(0, 0, 0, 0.2);
  backdrop-filter: blur(10px);
  border: none;
  max-width: 420px;
  min-width: 360px;
  width: 100%;
  position: relative;
  overflow: hidden;
}

/* 关闭按钮 */
.close-btn {
  position: absolute;
  top: 16px;
  right: 16px;
  width: 24px;
  height: 24px;
  border: none;
  background: rgba(107, 114, 128, 0.1);
  border-radius: 50%;
  color: #6b7280;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
  transition: all 0.2s ease;
  z-index: 10;
}

.close-btn:hover {
  background: rgba(107, 114, 128, 0.2);
  color: #374151;
  transform: scale(1.1);
}

/* 头部 */
.modal-header {
  padding: 24px 24px 16px;
  border-bottom: 1px solid rgba(0, 0, 0, 0.05);
  text-align: center;
}

.icon-wrapper {
  font-size: 32px;
  margin-bottom: 12px;
  display: flex;
  justify-content: center;
}

.modal-title {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  color: #333;
  line-height: 1.4;
}

/* 内容 */
.modal-content {
  padding: 20px 24px;
}

.modal-message {
  margin: 0;
  color: #666;
  line-height: 1.5;
  font-size: 14px;
  text-align: center;
}

/* 底部 */
.modal-footer {
  padding: 16px 24px 24px;
  display: flex;
  gap: 12px;
  justify-content: center;
}

/* 按钮样式 */
.btn {
  min-width: 100px;
  height: 44px;
  border-radius: 12px;
  font-weight: 600;
  font-size: 14px;
  cursor: pointer;
  transition: all 0.3s ease;
  border: none;
  outline: none;
}

.btn-cancel {
  background: rgba(107, 114, 128, 0.1);
  border: 1px solid rgba(107, 114, 128, 0.2);
  color: #6b7280;
}

.btn-cancel:hover {
  background: rgba(107, 114, 128, 0.15);
  border-color: rgba(107, 114, 128, 0.3);
  transform: translateY(-1px);
}

.btn-confirm {
  background: linear-gradient(135deg, #ef4444 0%, #dc2626 100%);
  border: none;
  color: white;
  box-shadow: 0 4px 12px rgba(239, 68, 68, 0.3);
}

.btn-confirm:hover {
  background: linear-gradient(135deg, #dc2626 0%, #b91c1c 100%);
  transform: translateY(-1px);
  box-shadow: 0 6px 16px rgba(239, 68, 68, 0.4);
}

/* 动画效果 */
.mask-enter-active {
  transition: all 0.3s ease;
}

.mask-leave-active {
  transition: all 0.2s ease;
}

.mask-enter-from,
.mask-leave-to {
  opacity: 0;
}

.modal-enter-active {
  transition: all 0.3s ease;
}

.modal-leave-active {
  transition: all 0.2s ease;
}

.modal-enter-from {
  opacity: 0;
  transform: scale(0.8) translateY(-20px);
}

.modal-leave-to {
  opacity: 0;
  transform: scale(0.9) translateY(10px);
}

/* 响应式 */
@media (max-width: 480px) {
  .confirm-modal {
    min-width: 300px;
    margin: 10px;
  }
  
  .modal-header {
    padding: 20px 20px 16px;
  }
  
  .modal-content {
    padding: 16px 20px;
  }
  
  .modal-footer {
    padding: 16px 20px 20px;
    flex-direction: column;
  }
  
  .btn {
    width: 100%;
    margin: 0;
  }
}
</style>