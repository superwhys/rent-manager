<template>
  <main class="not-found">
    <span>404</span>
    <h1>这个页面暂时找不到</h1>
    <p>地址可能已变更，回到工作台继续。</p>
    <RouterLink to="/dashboard" class="btn btn-primary">返回工作台</RouterLink>
  </main>
</template>

<style scoped>
.not-found {
  min-height: 100dvh;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 1.5rem;
  padding: 2rem;
  text-align: center;
}
.not-found > span {
  font-size: 5rem;
  color: #aac0b2;
  font-weight: 300;
}
.not-found h1 {
  font-size: 1.5rem;
}
.not-found p {
  color: var(--text-secondary);
}
</style>
