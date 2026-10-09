<script setup lang="ts">
import { computed } from 'vue'
import { useAuth } from '../composables/useAuth'
import type { PermissionCheckOptions } from '../types'

/**
 * 权限控制组件
 * 根据用户权限决定是否渲染子内容
 */

interface Props {
    // 需要的权限（字符串或字符串数组）
    permission?: string | string[]
    // 权限检查选项
    options?: PermissionCheckOptions
    // 无权限时是否显示警告信息
    showWarning?: boolean
    // 自定义无权限时的提示文本
    noPermissionText?: string
}

const props = withDefaults(defineProps<Props>(), {
    permission: '',
    options: () => ({}),
    showWarning: false,
    noPermissionText: '您没有权限查看此内容'
})

const { hasPermission, hasPermissions } = useAuth()

// 检查权限
const hasAccess = computed(() => {
    if (!props.permission) return true

    if (typeof props.permission === 'string') {
        return hasPermission(props.permission)
    } else if (Array.isArray(props.permission)) {
        return hasPermissions(props.permission, props.options)
    }

    return false
})
</script>

<template>
    <!-- 有权限时渲染子内容 -->
    <template v-if="hasAccess">
        <slot />
    </template>

    <!-- 无权限时的备用内容 -->
    <template v-else>
        <!-- 如果有fallback插槽，显示自定义无权限内容 -->
        <slot v-if="$slots.fallback" name="fallback" />

        <!-- 如果没有fallback插槽但showWarning为true，显示警告信息 -->
        <div v-else-if="showWarning" class="permission-warning">
            <el-alert :title="noPermissionText" type="warning" show-icon :closable="false" />
        </div>

        <!-- 如果传递了fallback选项，渲染它 -->
        <component v-else-if="options.fallback" :is="options.fallback" />
    </template>
</template>

<style scoped>
.permission-warning {
    margin: 8px 0;
}
</style>