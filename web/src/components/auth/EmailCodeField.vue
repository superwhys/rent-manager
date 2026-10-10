<script setup lang="ts">
defineProps<{
  modelValue: string
  disabled: boolean
  sending: boolean
  remaining: number
  message: string
  sendError: string
  validationError: string
}>()
defineEmits<{
  'update:modelValue': [value: string]
  send: []
  blur: []
}>()
</script>

<template>
  <div class="field">
    <label for="email-code">邮箱验证码</label>
    <div class="code-controls">
      <input
        id="email-code"
        :value="modelValue"
        type="text"
        inputmode="numeric"
        autocomplete="one-time-code"
        placeholder="6 位验证码"
        maxlength="6"
        :disabled="disabled"
        :aria-invalid="Boolean(validationError)"
        aria-describedby="code-status"
        @input="
          $emit('update:modelValue', ($event.target as HTMLInputElement).value)
        "
        @blur="$emit('blur')"
      />
      <el-button
        native-type="button"
        :loading="sending"
        :disabled="disabled || sending || remaining > 0"
        @click="$emit('send')"
      >
        {{
          sending
            ? '发送中'
            : remaining > 0
              ? `${remaining} 秒后重发`
              : '获取验证码'
        }}
      </el-button>
    </div>
    <div id="code-status" class="code-status">
      <p v-if="validationError || sendError" class="auth-error" role="alert">
        {{ validationError || sendError }}
      </p>
      <small v-else-if="message" role="status">{{ message }}</small>
      <small v-else>验证码发送至上方邮箱，10 分钟内有效。</small>
    </div>
  </div>
</template>

<style scoped>
.code-controls {
  display: flex;
  align-items: stretch;
  gap: 0.55rem;
}
.code-controls input {
  flex: 1;
  width: 0;
}
.code-controls .el-button {
  flex-shrink: 0;
  min-height: 3.15rem;
  height: auto;
  padding-inline: 0.7rem;
}
.code-status {
  line-height: 1.6;
}
</style>
