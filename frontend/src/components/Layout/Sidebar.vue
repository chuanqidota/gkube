<template>
  <div class="sidebar-container">
    <div class="sidebar-logo">
      <img src="@/assets/vue.svg" alt="logo" class="logo-img" />
      <span v-show="!isCollapse" class="logo-text">GKube</span>
    </div>
    <el-menu
      :default-active="activeMenu"
      :collapse="isCollapse"
      :collapse-transition="false"
      router
      background-color="#001529"
      text-color="#ffffffb3"
      active-text-color="#409eff"
      class="sidebar-menu"
    >
      <el-menu-item index="/dashboard">
        <el-icon><Odometer /></el-icon>
        <template #title>Dashboard</template>
      </el-menu-item>
      <el-menu-item index="/clusters">
        <el-icon><Connection /></el-icon>
        <template #title>集群管理</template>
      </el-menu-item>
      <el-sub-menu index="workloads">
        <template #title>
          <el-icon><Box /></el-icon>
          <span>工作负载</span>
        </template>
        <el-menu-item index="/workloads/pods">
          <el-icon><Coin /></el-icon>
          <template #title>Pods</template>
        </el-menu-item>
        <el-menu-item index="/workloads/deployments">
          <el-icon><Files /></el-icon>
          <template #title>Deployments</template>
        </el-menu-item>
      </el-sub-menu>
      <el-sub-menu index="tools">
        <template #title>
          <el-icon><Monitor /></el-icon>
          <span>运维工具</span>
        </template>
        <el-menu-item index="/terminal">
          <el-icon><Promotion /></el-icon>
          <template #title>Web 终端</template>
        </el-menu-item>
        <el-menu-item index="/logs">
          <el-icon><Document /></el-icon>
          <template #title>日志查看</template>
        </el-menu-item>
      </el-sub-menu>
      <el-sub-menu index="system">
        <template #title>
          <el-icon><Setting /></el-icon>
          <span>系统管理</span>
        </template>
        <el-menu-item index="/system/users">
          <el-icon><User /></el-icon>
          <template #title>用户管理</template>
        </el-menu-item>
        <el-menu-item index="/system/roles">
          <el-icon><UserFilled /></el-icon>
          <template #title>角色管理</template>
        </el-menu-item>
      </el-sub-menu>
    </el-menu>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import {
  Odometer,
  Connection,
  Setting,
  User,
  UserFilled,
  Monitor,
  Promotion,
  Document,
  Box,
  Coin,
  Files,
} from '@element-plus/icons-vue'

defineProps<{
  isCollapse: boolean
}>()

const route = useRoute()
const activeMenu = computed(() => route.path)
</script>

<style scoped>
.sidebar-container {
  height: 100%;
  display: flex;
  flex-direction: column;
  background-color: #001529;
}

.sidebar-logo {
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0 16px;
  gap: 8px;
  border-bottom: 1px solid #ffffff1a;
}

.logo-img {
  width: 32px;
  height: 32px;
}

.logo-text {
  color: #fff;
  font-size: 18px;
  font-weight: 600;
  white-space: nowrap;
}

.sidebar-menu {
  flex: 1;
  border-right: none;
  overflow-y: auto;
}

.sidebar-menu:not(.el-menu--collapse) {
  width: 220px;
}
</style>
