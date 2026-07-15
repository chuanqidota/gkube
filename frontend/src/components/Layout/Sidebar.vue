<template>
  <div class="sidebar-container">
    <div class="sidebar-logo">
      <div class="logo-icon">
        <svg viewBox="0 0 48 48" width="32" height="32">
          <defs>
            <linearGradient id="sidebar-hex" x1="0" y1="0" x2="1" y2="1">
              <stop offset="0%" stop-color="#6366f1"/>
              <stop offset="50%" stop-color="#818cf8"/>
              <stop offset="100%" stop-color="#3b82f6"/>
            </linearGradient>
          </defs>
          <path d="M24 2 L44 14 L44 34 L24 46 L4 34 L4 14 Z" fill="url(#sidebar-hex)"/>
          <path d="M30 16 C26.5 13.5 21.5 13.5 18 16 C14.5 18.5 14 23 14 24 C14 28 16 31 20 33 C23 34.5 27 34.5 30 33 L30 26 L24 26 L24 23 L30 23 Z" fill="white" opacity="0.95"/>
          <circle cx="36" cy="12" r="3" fill="#a78bfa" opacity="0.8"/>
        </svg>
      </div>
      <transition name="fade">
        <span v-show="!isCollapse" class="logo-text">GKube</span>
      </transition>
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
        <el-menu-item index="/workloads/hpa" @click="navigateTo('/workloads/hpa')">
          <el-icon><DataLine /></el-icon>
          <template #title>{{ t('sidebar.hpa') }}</template>
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
      <!-- Access Control (RBAC) -->
      <el-sub-menu index="access-control">
        <template #title>
          <el-icon><Lock /></el-icon>
          <span>{{ t('sidebar.accessControl') }}</span>
        </template>
        <!-- Namespace-scoped -->
        <el-sub-menu index="rbac-namespace-scoped">
          <template #title>
            <el-icon><FolderOpened /></el-icon>
            <span>{{ t('sidebar.rbacNamespaceScoped') }}</span>
          </template>
          <el-menu-item index="/rbac/roles" @click="navigateTo('/rbac/roles')">
            <el-icon><UserFilled /></el-icon>
            <template #title>{{ t('sidebar.roles') }}</template>
          </el-menu-item>
          <el-menu-item index="/rbac/rolebindings" @click="navigateTo('/rbac/rolebindings')">
            <el-icon><Link /></el-icon>
            <template #title>{{ t('sidebar.rolebindings') }}</template>
          </el-menu-item>
          <el-menu-item index="/rbac/serviceaccounts" @click="navigateTo('/rbac/serviceaccounts')">
            <el-icon><User /></el-icon>
            <template #title>{{ t('sidebar.serviceaccounts') }}</template>
          </el-menu-item>
        </el-sub-menu>
        <!-- Cluster-scoped -->
        <el-sub-menu index="rbac-cluster-scoped">
          <template #title>
            <el-icon><Connection /></el-icon>
            <span>{{ t('sidebar.rbacClusterScoped') }}</span>
          </template>
          <el-menu-item index="/rbac/clusterroles" @click="navigateTo('/rbac/clusterroles')">
            <el-icon><Stamp /></el-icon>
            <template #title>{{ t('sidebar.clusterroles') }}</template>
          </el-menu-item>
          <el-menu-item index="/rbac/clusterrolebindings" @click="navigateTo('/rbac/clusterrolebindings')">
            <el-icon><CircleCheck /></el-icon>
            <template #title>{{ t('sidebar.clusterrolebindings') }}</template>
          </el-menu-item>
        </el-sub-menu>
        <!-- Tools -->
        <el-sub-menu index="rbac-tools">
          <template #title>
            <el-icon><SetUp /></el-icon>
            <span>{{ t('sidebar.rbacTools') }}</span>
          </template>
          <el-menu-item index="/rbac/permission-check" @click="navigateTo('/rbac/permission-check')">
            <el-icon><Search /></el-icon>
            <template #title>{{ t('sidebar.permissionCheck') }}</template>
          </el-menu-item>
          <el-menu-item index="/rbac/templates" @click="navigateTo('/rbac/templates')">
            <el-icon><DocumentCopy /></el-icon>
            <template #title>{{ t('sidebar.rbacTemplates') }}</template>
          </el-menu-item>
        </el-sub-menu>
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
      <el-sub-menu index="system">
        <template #title>
          <el-icon><Setting /></el-icon>
          <span>{{ t('sidebar.system') }}</span>
        </template>
        <el-menu-item index="/users" @click="navigateTo('/users')">
          <el-icon><User /></el-icon>
          <template #title>{{ t('sidebar.users') }}</template>
        </el-menu-item>
        <el-menu-item index="/settings/auth" @click="navigateTo('/settings/auth')">
          <el-icon><Setting /></el-icon>
          <template #title>{{ t('sidebar.authSettings') }}</template>
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
import {
  Odometer,
  Connection,
  Setting,
  User,
  UserFilled,
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
  Stamp,
  CircleCheck,
  Search,
} from '@element-plus/icons-vue'

defineProps<{
  isCollapse: boolean
}>()

const route = useRoute()
const router = useRouter()
const { t } = useI18n()

const activeMenu = computed(() => {
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

.logo-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.logo-text {
  color: var(--gk-sidebar-text-active);
  font-size: var(--gk-font-size-xl);
  font-weight: 700;
  white-space: nowrap;
  letter-spacing: -0.02em;
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

/* Fade transition for logo text */
.fade-enter-active,
.fade-leave-active {
  transition: opacity var(--gk-transition-fast);
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
