<template>
  <el-container class="app-layout">
    <el-aside
      :width="isCollapse ? 'var(--gk-sidebar-collapsed-width)' : 'var(--gk-sidebar-width)'"
      class="app-aside"
    >
      <Sidebar :is-collapse="isCollapse" />
    </el-aside>
    <el-container class="app-content-wrapper">
      <el-header class="app-header">
        <Header @toggle-collapse="isCollapse = !isCollapse" />
      </el-header>
      <el-main class="app-main">
        <router-view :key="reloadKey" />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useClusterStore } from '@/stores/cluster'
import Sidebar from './Sidebar.vue'
import Header from './Header.vue'

const clusterStore = useClusterStore()
const isCollapse = ref(false)

// 仅在切换集群时强制重挂当前路由页面，触发各页面 refetch 新集群数据。
// 不再对每次路由 fullPath 变化都重挂（原 viewKey 行为已移除），避免无谓的重渲染。
const reloadKey = ref(0)
watch(() => clusterStore.currentCluster, () => {
  reloadKey.value++
})
</script>

<style scoped>
.app-layout {
  height: 100vh;
  overflow: hidden;
}

.app-aside {
  background: var(--gk-color-bg-sidebar);
  transition: width var(--gk-transition-slow);
  overflow-x: hidden;
  overflow-y: auto;
  border-right: 1px solid var(--gk-color-border);
}

.app-header {
  padding: 0;
  height: var(--gk-header-height);
  border-bottom: 1px solid var(--gk-color-border);
  box-shadow: var(--gk-shadow-sm);
  background: var(--gk-color-bg-header);
}

.app-main {
  background: var(--gk-color-bg-page);
  padding: 0;
  overflow-y: auto;
  min-height: 0;
}

.app-content-wrapper {
  overflow: hidden;
}

/* Responsive: auto-collapse sidebar on small screens */
@media (max-width: 768px) {
  .app-aside {
    position: fixed;
    z-index: 1000;
    height: 100vh;
    box-shadow: var(--gk-shadow-lg);
  }
}
</style>
