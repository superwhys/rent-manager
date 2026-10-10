<script setup lang="ts">
import PageHeader from '@/components/common/PageHeader.vue'
import UnavailableFeature from '@/components/common/UnavailableFeature.vue'
</script>

<template>
  <PageHeader title="经营报表" description="看见经营的每一步变化。" />
  <UnavailableFeature
    title="经营报表暂未开放"
    icon="chart"
    description="当前尚未开放历史趋势与报表导出。经营概览中已提供本月收入、出租情况和待办提醒。"
  />
</template>
