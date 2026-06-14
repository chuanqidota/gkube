<template>
  <el-container class="app-layout">
    <el-aside :width="isCollapse ? '64px' : '220px'" class="app-aside">
      <Sidebar :is-collapse="isCollapse" />
    </el-aside>
    <el-container>
      <el-header class="app-header">
        <Header @toggle-collapse="isCollapse = !isCollapse" />
      </el-header>
      <el-main class="app-main">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import Sidebar from './Sidebar.vue'
import Header from './Header.vue'

const isCollapse = ref(false)
</script>

<style scoped>
.app-layout {
  height: 100vh;
}

.app-aside {
  background: #001529;
  transition: width 0.3s;
  overflow: hidden;
}

.app-header {
  padding: 0;
  height: 60px;
}

.app-main {
  background: #f0f2f5;
  padding: 0;
  overflow-y: auto;
  min-height: 0;
}

/* Responsive: auto-collapse sidebar on small screens */
@media (max-width: 768px) {
  .app-aside {
    position: fixed;
    z-index: 1000;
    height: 100vh;
  }
}
</style>
