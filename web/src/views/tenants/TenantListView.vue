<script setup lang="ts">
import PageHeader from '@/components/common/PageHeader.vue'
import UnavailableFeature from '@/components/common/UnavailableFeature.vue'
</script>

<template>
  <PageHeader title="租客管理" description="关于居住的人，留一份妥帖的记录。" />
  <UnavailableFeature
    title="租客与合同管理暂未开放"
    icon="users"
    description="当前尚未开放租客档案和租约管理。你可以先在房间管理中维护房间信息与出租状态。"
  />
</template>
