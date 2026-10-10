<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { getHomeSummary } from '@/api/home'
import type { HomeSummary } from '@/types/home'
import { useUnits } from '@/stores/units'
import AppIcon from '@/components/common/AppIcon.vue'
import PageHeader from '@/components/common/PageHeader.vue'
import AsyncState from '@/components/common/AsyncState.vue'
import { errorMessage } from '@/utils/format'

const data = ref<HomeSummary>()
const loading = ref(true)
const error = ref('')
const {
  units,
  loading: unitsLoading,
  error: unitsError,
  loadUnits,
  setSelectedUnit,
} = useUnits()
const today = new Date().toLocaleDateString('zh-CN', {
  month: 'long',
  day: 'numeric',
  weekday: 'long',
})
const occupancy = computed(() => {
  const summary = data.value?.building_room_summary
  return summary?.room_count
    ? Math.round((summary.rented_room_count / summary.room_count) * 100)
    : 0
})
const stats = computed(() => {
  if (!data.value) return []
  const summary = data.value.building_room_summary
  return [
    {
      label: '管理房源',
      value: summary.building_count,
      unit: '栋',
      icon: 'building',
      color: 'green',
      note: '自有及获授权房源',
    },
    {
      label: '房间总数',
      value: summary.room_count,
      unit: '间',
      icon: 'home',
      color: 'blue',
      note: '所有房源的房间',
    },
    {
      label: '已出租',
      value: summary.rented_room_count,
      unit: '间',
      icon: 'check',
      color: 'purple',
      note: `当前出租率 ${occupancy.value}%`,
    },
    {
      label: '空置房间',
      value: summary.vacant_room_count,
      unit: '间',
      icon: 'grid',
      color: 'orange',
      note: '可安排新的入住',
    },
  ]
})
async function load() {
  if (!loading.value) loading.value = true
  error.value = ''
  try {
    data.value = await getHomeSummary()
  } catch (cause) {
    error.value = errorMessage(cause)
  } finally {
    loading.value = false
  }
}
onMounted(load)
</script>

<template>
  <PageHeader
    title="每一处日常，尽在掌握。"
    :description="`${today} · 欢迎回到你的租赁工作台`"
  >
    <template #actions>
      <el-button
        class="overview-refresh"
        :loading="loading"
        aria-label="刷新经营概览"
        @click="load"
      >
        <AppIcon name="refresh" :size="15" />
        <span>刷新</span>
      </el-button>
      <RouterLink class="btn btn-primary" to="/properties"
        ><AppIcon name="plus" :size="17" /> 管理房源</RouterLink
      >
    </template>
  </PageHeader>
  <AsyncState :loading="loading" :error="error" @retry="load">
    <template v-if="data">
      <div class="stat-grid">
        <article
          v-for="stat in stats"
          :key="stat.label"
          class="stat-card panel"
        >
          <div class="stat-top">
            <span>{{ stat.label }}</span
            ><span class="stat-icon" :class="stat.color"
              ><AppIcon :name="stat.icon" :size="19"
            /></span>
          </div>
          <div class="stat-value">
            {{ stat.value }}<small>{{ stat.unit }}</small>
          </div>
          <p>{{ stat.note }}</p>
        </article>
      </div>
      <div class="overview-middle">
        <section class="income-card">
          <div class="income-heading">
            <span><AppIcon name="wallet" :size="18" /> 本月已收款</span
            ><span class="income-month">{{ data.current_month }}</span>
          </div>
          <div class="income-amount">
            <span>¥</span>{{ data.income_amount
            }}<small>{{ data.income_decimal }}</small>
          </div>
          <p class="growth">
            较上月 <strong>{{ data.monthly_growth }}</strong
            ><span>以已收款账单为准</span>
          </p>
          <div class="income-footer">
            <span>把账目理清，给生活留白。</span
            ><RouterLink to="/bills"
              >查看账单 <AppIcon name="arrow-right" :size="17"
            /></RouterLink>
          </div>
        </section>
        <section class="panel occupancy-card">
          <div class="section-title">
            <h2>出租情况</h2>
            <span class="muted">当前</span>
          </div>
          <div class="occupancy-content">
            <div
              class="occupancy-ring"
              :style="{ '--occupancy': `${occupancy}%` }"
              role="img"
              :aria-label="`出租率 ${occupancy}%`"
            >
              <div>
                <strong>{{ occupancy }}<small>%</small></strong
                ><span>出租率</span>
              </div>
            </div>
            <div class="occupancy-legend">
              <div>
                <span class="legend-dot rented" />已出租<strong
                  >{{ data.building_room_summary.rented_room_count
                  }}<small>间</small></strong
                >
              </div>
              <div>
                <span class="legend-dot vacant" />空置中<strong
                  >{{ data.building_room_summary.vacant_room_count
                  }}<small>间</small></strong
                >
              </div>
              <RouterLink to="/rooms"
                >查看房间 <AppIcon name="chevron-right" :size="13"
              /></RouterLink>
            </div>
          </div>
        </section>
      </div>
      <div class="overview-bottom">
        <section class="panel properties-panel">
          <div class="section-title">
            <h2>
              我的房源 <span class="count-label">{{ units.length }}</span>
            </h2>
            <RouterLink to="/properties"
              >全部房源 <AppIcon name="chevron-right" :size="14"
            /></RouterLink>
          </div>
          <el-scrollbar
            class="overview-scrollbar"
            always
            :tabindex="0"
            role="region"
            aria-label="我的房源列表"
          >
            <AsyncState
              :loading="unitsLoading"
              :error="unitsError"
              :empty="units.length === 0"
              empty-title="从你的第一处房源开始"
              empty-description="添加楼栋后，就可以管理房间、记录抄表和生成账单。"
              @retry="loadUnits"
            >
              <div class="property-list">
                <RouterLink
                  v-for="unit in units"
                  :key="unit.id"
                  class="property-row"
                  to="/rooms"
                  @click="setSelectedUnit(unit.id)"
                >
                  <span class="property-icon"
                    ><AppIcon name="building" :size="25"
                  /></span>
                  <div class="property-description">
                    <strong>{{ unit.name }}</strong
                    ><span v-if="unit.address !== undefined"
                      >{{ unit.province }} {{ unit.city }} ·
                      {{ unit.address }}</span
                    >
                    <span v-else>详细资料未授权</span>
                  </div>
                  <div class="property-numbers">
                    <strong
                      >{{ unit.rented_count ?? '—'
                      }}<span> / {{ unit.room_count ?? '—' }}</span></strong
                    ><small>已出租 / 总房间</small>
                  </div>
                  <AppIcon name="chevron-right" :size="15" />
                </RouterLink>
              </div>
            </AsyncState>
          </el-scrollbar>
        </section>
        <section class="panel todo-panel">
          <div class="section-title">
            <h2>待办提醒</h2>
            <span class="count-label">{{ data.todos.length }}</span>
          </div>
          <el-scrollbar
            class="overview-scrollbar"
            always
            :tabindex="0"
            role="region"
            aria-label="待办提醒列表"
          >
            <div v-if="!data.todos.length" class="todo-empty">
              <AppIcon name="check" :size="28" />
              <h3>今天，一切就绪</h3>
              <p>暂时没有待处理事项。</p>
            </div>
            <div
              v-for="(todo, index) in data.todos"
              :key="`${todo.title}-${index}`"
              class="todo-item"
            >
              <span class="todo-dot" />
              <div>
                <div class="todo-title">
                  <h3>{{ todo.title }}</h3>
                  <span>{{ todo.tag }}</span>
                </div>
                <p>{{ todo.description }}</p>
              </div>
            </div>
          </el-scrollbar>
          <RouterLink to="/bills" class="todo-action"
            >前往账单管理 <AppIcon name="arrow-right" :size="16"
          /></RouterLink>
        </section>
      </div>
      <section class="quick-actions">
        <span>快捷操作</span
        ><RouterLink to="/rooms"
          ><AppIcon name="home" :size="17" />管理房间</RouterLink
        ><RouterLink to="/utilities"
          ><AppIcon name="droplet" :size="17" />录入抄表</RouterLink
        ><RouterLink to="/bills"
          ><AppIcon name="wallet" :size="17" />生成账单</RouterLink
        ><RouterLink to="/users"
          ><AppIcon name="shield" :size="17" />管理授权</RouterLink
        >
      </section>
    </template>
  </AsyncState>
</template>

<style scoped>
.overview-refresh :deep(> span) {
  gap: 0.5rem;
}
.stat-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 1.2rem;
  margin-bottom: 1.5rem;
}
.stat-card {
  margin: 0;
  padding: 1.3rem 1.4rem;
}
.stat-top {
  display: flex;
  justify-content: space-between;
  align-items: center;
  color: var(--text-secondary);
  font-size: 0.83rem;
}
.stat-icon {
  height: 34px;
  width: 34px;
  display: grid;
  place-items: center;
  border-radius: 10px;
}
.green {
  background: #edf6f0;
  color: #508d6b;
}
.blue {
  background: #edf3fa;
  color: #6a90b6;
}
.purple {
  background: #f1eff8;
  color: #9985bf;
}
.orange {
  background: #fdf3e8;
  color: #c6a06a;
}
.stat-value {
  font-size: 2.35rem;
  font-weight: 600;
  letter-spacing: -0.04em;
  margin: 0.5rem 0 0.3rem;
  font-variant-numeric: tabular-nums;
}
.stat-value small {
  font-size: 0.8rem;
  color: #8d979a;
  margin-left: 0.6rem;
  font-weight: 400;
}
.stat-card p {
  font-size: 0.74rem;
  color: #909b9f;
}
.overview-middle {
  display: grid;
  grid-template-columns: 1.25fr 1fr;
  gap: 1.5rem;
  margin-bottom: 1.5rem;
}
.income-card {
  padding: 1.65rem 1.8rem 0;
  background: #eaf3ef;
  border: 1px solid #e0ece5;
  border-radius: 16px;
  color: #2d5944;
  position: relative;
  overflow: hidden;
}
.income-heading {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  font-size: 0.86rem;
}
.income-heading > span:first-child {
  display: flex;
  align-items: center;
  gap: 0.7rem;
}
.income-month {
  border: 1px solid #d6e4dc;
  border-radius: 5px;
  padding: 0.15rem 0.5rem;
  font-size: 0.7rem;
  color: #789583;
}
.income-amount {
  font-size: clamp(2.5rem, 4vw, 3.3rem);
  line-height: 1.25;
  letter-spacing: -0.035em;
  font-weight: 550;
  margin: 1.2rem 0 0.6rem;
  font-variant-numeric: tabular-nums;
}
.income-amount > span {
  font-size: 1.6rem;
  margin-right: 0.6rem;
  font-weight: 400;
  color: #719c83;
}
.income-amount small {
  font-size: 1.8rem;
  color: #6d947e;
  font-weight: 400;
}
.growth {
  color: #799284;
  font-size: 0.74rem;
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  align-items: center;
}
.growth strong {
  color: #487459;
  font-weight: 500;
}
.growth > span {
  margin-left: 0.5rem;
  color: #92a79b;
}
.income-footer {
  margin-top: 1.75rem;
  padding: 1rem 0;
  border-top: 1px solid #dbe7df;
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 0.72rem;
  gap: 1rem;
  color: #8a9f91;
}
.income-footer a {
  color: #4a755b;
  display: flex;
  gap: 0.5rem;
  align-items: center;
}
.occupancy-card {
  margin: 0;
  padding: 1.6rem;
}
.section-title {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
}
.section-title h2 {
  font-size: 1rem;
}
.section-title > a {
  display: flex;
  align-items: center;
  color: #8b9790;
  gap: 0.25rem;
  font-size: 0.73rem;
}
.section-title > .muted {
  font-size: 0.72rem;
}
.occupancy-content {
  display: flex;
  align-items: center;
  justify-content: space-evenly;
  gap: 1.5rem;
  padding-top: 1.9rem;
}
.occupancy-ring {
  flex-shrink: 0;
  width: 133px;
  height: 133px;
  border-radius: 50%;
  background: conic-gradient(
    #6f9f86 0 var(--occupancy),
    #edf1ee var(--occupancy) 100%
  );
  display: grid;
  place-items: center;
  transform: rotate(-90deg);
}
.occupancy-ring > div {
  width: 110px;
  height: 110px;
  border-radius: 50%;
  background: white;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  transform: rotate(90deg);
}
.occupancy-ring strong {
  font-size: 2rem;
  font-weight: 550;
  line-height: 1.3;
}
.occupancy-ring small {
  font-size: 0.9rem;
  margin-left: 0.15rem;
}
.occupancy-ring span {
  font-size: 0.72rem;
  color: #98a29a;
}
.occupancy-legend {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  min-width: 115px;
}
.occupancy-legend > div {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.76rem;
  color: #849087;
}
.legend-dot {
  width: 7px;
  height: 7px;
  border-radius: 2px;
}
.rented {
  background: #6f9f86;
}
.vacant {
  background: #e4ebe6;
}
.occupancy-legend strong {
  margin-left: auto;
  font-weight: 550;
  color: #59645e;
}
.occupancy-legend small {
  font-size: 0.65rem;
  font-weight: 400;
  margin-left: 0.3rem;
}
.occupancy-legend a {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  color: #7c9283;
  font-size: 0.73rem;
}
.overview-bottom {
  display: grid;
  grid-template-columns: minmax(0, 1.5fr) minmax(0, 1fr);
  gap: 1.5rem;
}
.overview-bottom .panel {
  margin: 0;
  min-width: 0;
  min-height: 420px;
  max-height: 480px;
  display: flex;
  flex-direction: column;
}
.overview-bottom .section-title,
.todo-action {
  flex-shrink: 0;
}
.overview-scrollbar {
  flex: 1;
  min-height: 0;
}
.overview-scrollbar :deep(.el-scrollbar__view) {
  padding-right: 0.75rem;
}
.count-label {
  font-size: 0.7rem;
  color: #90a094;
  background: #f2f5f2;
  padding: 0.15rem 0.4rem;
  border-radius: 4px;
  margin-left: 0.4rem;
  font-weight: 400;
}
.property-list {
  margin-top: 0.85rem;
}
.property-row {
  display: flex;
  gap: 0.9rem;
  align-items: center;
  padding: 1.15rem 0;
  border-bottom: 1px solid #f0f3f1;
}
.property-row:last-child {
  border-bottom: 0;
  padding-bottom: 0.25rem;
}
.property-row > svg {
  color: #c0c9c4;
}
.property-icon {
  width: 43px;
  height: 48px;
  border-radius: 9px;
  background: #f0f3ef;
  color: #91a389;
  display: grid;
  place-items: center;
  flex-shrink: 0;
}
.property-description {
  flex: 1;
  min-width: 0;
}
.property-description strong {
  display: block;
  font-size: 0.85rem;
  font-weight: 500;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.property-description span {
  display: block;
  margin-top: 0.35rem;
  font-size: 0.68rem;
  color: #919b96;
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
}
.property-numbers {
  text-align: right;
  white-space: nowrap;
}
.property-numbers strong {
  font-size: 0.85rem;
  font-weight: 550;
  color: #718278;
}
.property-numbers span {
  color: #9eaaa1;
  font-weight: 400;
}
.property-numbers small {
  display: block;
  margin-top: 0.3rem;
  font-size: 0.63rem;
  color: #a1aaa5;
}
.todo-item {
  display: flex;
  align-items: flex-start;
  gap: 0.65rem;
  padding-top: 1.6rem;
}
.todo-dot {
  width: 6px;
  height: 6px;
  background: #bd9b6a;
  border-radius: 50%;
  margin-top: 0.45rem;
  flex-shrink: 0;
}
.todo-item > div {
  flex: 1;
  min-width: 0;
  overflow-wrap: anywhere;
}
.todo-title {
  display: flex;
  flex-wrap: wrap;
  gap: 0.6rem;
  align-items: center;
}
.todo-title h3 {
  font-size: 0.82rem;
  font-weight: 500;
}
.todo-title span {
  font-size: 0.62rem;
  color: #b89b6e;
  border-radius: 4px;
  background: #faf4eb;
  padding: 0.1rem 0.35rem;
}
.todo-item p {
  font-size: 0.72rem;
  color: #939c98;
  margin-top: 0.5rem;
  line-height: 1.7;
}
.todo-action {
  border-top: 1px solid #edf1ee;
  padding-top: 1rem;
  margin-top: 1.5rem;
  display: flex;
  justify-content: space-between;
  font-size: 0.75rem;
  color: #7f9485;
}
.todo-empty {
  text-align: center;
  padding: 1.5rem 0 0.5rem;
  color: #8da996;
}
.todo-empty h3 {
  font-size: 0.85rem;
  margin-top: 0.5rem;
}
.todo-empty p {
  font-size: 0.75rem;
  margin-top: 0.4rem;
  color: #9aa69e;
}
.quick-actions {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  padding: 1.1rem 1.5rem;
  margin-top: 1.5rem;
  border: 1px dashed #dce5df;
  border-radius: 12px;
}
.quick-actions > span {
  color: #95a197;
  font-size: 0.75rem;
}
.quick-actions a {
  display: flex;
  align-items: center;
  gap: 0.65rem;
  font-size: 0.77rem;
  color: #728579;
}
@media (max-width: 1200px) {
  .overview-middle {
    grid-template-columns: 1.1fr 1fr;
  }
  .occupancy-content {
    gap: 0.8rem;
  }
  .occupancy-ring {
    width: 108px;
    height: 108px;
  }
  .occupancy-ring > div {
    width: 88px;
    height: 88px;
  }
  .income-card {
    padding-left: 1.4rem;
    padding-right: 1.4rem;
  }
  .growth > span {
    display: none;
  }
}
@media (max-width: 1000px) {
  .stat-grid {
    gap: 0.75rem;
  }
  .stat-card {
    padding: 1.1rem;
  }
  .overview-middle,
  .overview-bottom {
    gap: 1rem;
  }
  .occupancy-content {
    flex-direction: column;
    padding-top: 1rem;
  }
  .occupancy-legend {
    flex-direction: row;
    flex-wrap: wrap;
    justify-content: center;
    gap: 0.75rem;
  }
  .occupancy-legend a {
    display: none;
  }
  .quick-actions {
    flex-wrap: wrap;
  }
}
@media (max-width: 600px) {
  .stat-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
  .overview-middle,
  .overview-bottom {
    grid-template-columns: 1fr;
  }
  .occupancy-content {
    flex-direction: row;
    padding: 1.5rem 0 0.5rem;
  }
  .occupancy-legend {
    flex-direction: column;
  }
  .income-amount {
    font-size: 3rem;
  }
  .property-numbers {
    display: none;
  }
  .quick-actions {
    display: grid;
    grid-template-columns: 1fr 1fr;
    row-gap: 1.4rem;
  }
  .quick-actions > span {
    grid-column: 1 / -1;
  }
}
</style>
