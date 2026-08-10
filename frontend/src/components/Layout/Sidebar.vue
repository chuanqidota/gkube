<template>
  <div class="sidebar-container">
    <div class="sidebar-logo">
      <Logo :size="32" :show-text="!isCollapse" :text-size="20" tone="light" />
    </div>
    <el-menu
      :default-active="activeMenu"
      :collapse="isCollapse"
      :collapse-transition="false"
      class="sidebar-menu"
    >
      <!-- Overview -->
      <el-menu-item index="/dashboard" @click="navigateTo('/dashboard')">
        <el-icon><Odometer /></el-icon>
        <template #title>{{ t('sidebar.overview') }}</template>
      </el-menu-item>
      <!-- Cluster Management -->
      <el-menu-item index="/clusters" @click="navigateTo('/clusters')">
        <el-icon><Connection /></el-icon>
        <template #title>{{ t('sidebar.clusters') }}</template>
      </el-menu-item>
      <!-- Nodes -->
      <el-menu-item index="/nodes" @click="navigateTo('/nodes')">
        <el-icon><Cpu /></el-icon>
        <template #title>{{ t('sidebar.nodes') }}</template>
      </el-menu-item>
      <!-- Namespaces -->
      <el-menu-item index="/namespaces" @click="navigateTo('/namespaces')">
        <el-icon><FolderOpened /></el-icon>
        <template #title>{{ t('sidebar.namespaces') }}</template>
      </el-menu-item>
      <!-- Workloads -->
      <el-sub-menu index="workloads">
        <template #title>
          <el-icon><Box /></el-icon>
          <span>{{ t('sidebar.workloads') }}</span>
        </template>
        <el-menu-item index="/workloads/pods" @click="navigateTo('/workloads/pods')">
          <el-icon><Coin /></el-icon>
          <template #title>{{ t('sidebar.pods') }}</template>
        </el-menu-item>
        <el-menu-item index="/workloads/deployments" @click="navigateTo('/workloads/deployments')">
          <el-icon><DocumentCopy /></el-icon>
          <template #title>{{ t('sidebar.deployments') }}</template>
        </el-menu-item>
        <el-menu-item index="/workloads/statefulsets" @click="navigateTo('/workloads/statefulsets')">
          <el-icon><List /></el-icon>
          <template #title>{{ t('sidebar.statefulsets') }}</template>
        </el-menu-item>
        <el-menu-item index="/workloads/daemonsets" @click="navigateTo('/workloads/daemonsets')">
          <el-icon><SetUp /></el-icon>
          <template #title>{{ t('sidebar.daemonsets') }}</template>
        </el-menu-item>
        <el-menu-item index="/workloads/replicasets" @click="navigateTo('/workloads/replicasets')">
          <el-icon><CopyDocument /></el-icon>
          <template #title>{{ t('sidebar.replicasets') }}</template>
        </el-menu-item>
        <el-menu-item index="/workloads/jobs" @click="navigateTo('/workloads/jobs')">
          <el-icon><Finished /></el-icon>
          <template #title>{{ t('sidebar.jobs') }}</template>
        </el-menu-item>
        <el-menu-item index="/workloads/cronjobs" @click="navigateTo('/workloads/cronjobs')">
          <el-icon><Timer /></el-icon>
          <template #title>{{ t('sidebar.cronjobs') }}</template>
        </el-menu-item>
        <el-menu-item index="/autoscaling" @click="navigateTo('/autoscaling')">
          <el-icon><DataLine /></el-icon>
          <template #title>{{ t('sidebar.autoscaling') }}</template>
        </el-menu-item>
      </el-sub-menu>
      <!-- Network -->
      <el-sub-menu index="network">
        <template #title>
          <el-icon><Share /></el-icon>
          <span>{{ t('sidebar.network') }}</span>
        </template>
        <el-menu-item index="/network/services" @click="navigateTo('/network/services')">
          <el-icon><Connection /></el-icon>
          <template #title>{{ t('sidebar.services') }}</template>
        </el-menu-item>
        <el-menu-item index="/network/ingresses" @click="navigateTo('/network/ingresses')">
          <el-icon><Link /></el-icon>
          <template #title>{{ t('sidebar.ingresses') }}</template>
        </el-menu-item>
        <el-menu-item index="/network/networkpolicies" @click="navigateTo('/network/networkpolicies')">
          <el-icon><Lock /></el-icon>
          <template #title>{{ t('sidebar.networkpolicies') }}</template>
        </el-menu-item>
      </el-sub-menu>
      <!-- Storage -->
      <el-sub-menu index="storage">
        <template #title>
          <el-icon><Coin /></el-icon>
          <span>{{ t('sidebar.storage') }}</span>
        </template>
        <el-menu-item index="/storage/pvs" @click="navigateTo('/storage/pvs')">
          <el-icon><Coin /></el-icon>
          <template #title>{{ t('sidebar.pvs') }}</template>
        </el-menu-item>
        <el-menu-item index="/storage/pvcs" @click="navigateTo('/storage/pvcs')">
          <el-icon><Box /></el-icon>
          <template #title>{{ t('sidebar.pvcs') }}</template>
        </el-menu-item>
        <el-menu-item index="/storage/storageclasses" @click="navigateTo('/storage/storageclasses')">
          <el-icon><Files /></el-icon>
          <template #title>{{ t('sidebar.storageclasses') }}</template>
        </el-menu-item>
      </el-sub-menu>
      <!-- Configuration -->
      <el-sub-menu index="config">
        <template #title>
          <el-icon><Tickets /></el-icon>
          <span>{{ t('sidebar.config') }}</span>
        </template>
        <el-menu-item index="/config/configmaps" @click="navigateTo('/config/configmaps')">
          <el-icon><Tickets /></el-icon>
          <template #title>{{ t('sidebar.configmaps') }}</template>
        </el-menu-item>
        <el-menu-item index="/config/secrets" @click="navigateTo('/config/secrets')">
          <el-icon><Key /></el-icon>
          <template #title>{{ t('sidebar.secrets') }}</template>
        </el-menu-item>
      </el-sub-menu>
      <!-- Events -->
      <el-menu-item index="/events" @click="navigateTo('/events')">
        <el-icon><Bell /></el-icon>
        <template #title>{{ t('sidebar.events') }}</template>
      </el-menu-item>
      <!-- CRD -->
      <el-menu-item index="/crd" @click="navigateTo('/crd')">
        <el-icon><Grid /></el-icon>
        <template #title>{{ t('sidebar.crd') }}</template>
      </el-menu-item>
      <!-- System Management -->
      <el-sub-menu v-if="isAdmin" index="system">
        <template #title>
          <el-icon><Setting /></el-icon>
          <span>{{ t('sidebar.system') }}</span>
        </template>
        <el-menu-item index="/users" @click="navigateTo('/users')">
          <el-icon><User /></el-icon>
          <template #title>{{ t('sidebar.users') }}</template>
        </el-menu-item>
        <el-menu-item index="/audit" @click="navigateTo('/audit')">
          <el-icon><Document /></el-icon>
          <template #title>{{ t('sidebar.audit') }}</template>
        </el-menu-item>
      </el-sub-menu>
    </el-menu>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '@/stores/auth'
import Logo from '@/components/Logo.vue'
import {
  Odometer,
  Connection,
  Setting,
  User,
  Document,
  Box,
  Coin,
  Files,
  Share,
  Link,
  Cpu,
  FolderOpened,
  Tickets,
  Key,
  Bell,
  DataLine,
  Lock,
  Grid,
  DocumentCopy,
  List,
  SetUp,
  Finished,
  Timer,
  CopyDocument,
} from '@element-plus/icons-vue'

defineProps<{
  isCollapse: boolean
}>()

const route = useRoute()
const router = useRouter()
const { t } = useI18n()
const authStore = useAuthStore()

// 仅管理员可见系统管理（用户/审计）入口；未加载到用户信息时默认可见，避免误隐藏。
const isAdmin = computed(() => authStore.user?.isAdmin !== false)

const activeMenu = computed(() => {
  if (route.path.startsWith('/autoscaling')) return '/autoscaling'
  // 详情页高亮父级菜单项
  if (route.meta.parent) {
    const parentRoute = router.resolve({ name: route.meta.parent as string })
    if (parentRoute.matched.length) return parentRoute.path
  }
  return route.path
})

function navigateTo(path: string) {
  if (route.path !== path) {
    router.push(path).catch(() => {})
  }
}
</script>

<style scoped>
.sidebar-container {
  height: 100%;
  display: flex;
  flex-direction: column;
  background-color: var(--gk-color-bg-sidebar);
}

.sidebar-logo {
  height: var(--gk-header-height);
  display: flex;
  align-items: center;
  padding: 0 var(--gk-space-4);
  gap: var(--gk-space-3);
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);
  flex-shrink: 0;
}

.sidebar-menu {
  flex: 1;
  border-right: none;
  overflow-y: auto;
  overflow-x: hidden;
}

.sidebar-menu:not(.el-menu--collapse) {
  width: var(--gk-sidebar-width);
}

/* Override Element Plus menu item styles */
.sidebar-menu .el-menu-item,
.sidebar-menu :deep(.el-sub-menu__title) {
  height: 44px;
  line-height: 44px;
  margin: 2px var(--gk-space-2);
  border-radius: var(--gk-radius-md);
  transition: all var(--gk-transition-fast);
}

.sidebar-menu .el-menu-item:hover,
.sidebar-menu :deep(.el-sub-menu__title:hover) {
  background-color: var(--gk-sidebar-hover-bg);
}

.sidebar-menu .el-menu-item.is-active {
  background-color: var(--gk-sidebar-active-bg);
  color: var(--gk-sidebar-text-active);
  position: relative;
}

.sidebar-menu .el-menu-item.is-active::before {
  content: '';
  position: absolute;
  left: 0;
  top: 8px;
  bottom: 8px;
  width: 3px;
  border-radius: 0 2px 2px 0;
  background-color: var(--gk-sidebar-active-indicator);
}

/* Sub-menu items - more indentation */
.sidebar-menu .el-sub-menu .el-menu-item {
  padding-left: 52px !important;
}

/* Scrollbar styling for sidebar */
.sidebar-menu::-webkit-scrollbar {
  width: 4px;
}

.sidebar-menu::-webkit-scrollbar-thumb {
  background: rgba(255, 255, 255, 0.15);
  border-radius: 2px;
}

.sidebar-menu::-webkit-scrollbar-thumb:hover {
  background: rgba(255, 255, 255, 0.25);
}

.sidebar-menu::-webkit-scrollbar-track {
  background: transparent;
}
</style>
