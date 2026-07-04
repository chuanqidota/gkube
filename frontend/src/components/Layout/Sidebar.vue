<template>
  <div class="sidebar-container">
    <div class="sidebar-logo">
      <div class="logo-icon">
        <svg viewBox="0 0 32 32" width="28" height="28" fill="none">
          <rect width="32" height="32" rx="8" fill="#3b82f6"/>
          <path d="M8 16L14 22L24 10" stroke="white" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"/>
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
      router
      class="sidebar-menu"
    >
      <el-menu-item index="/dashboard">
        <el-icon><Odometer /></el-icon>
        <template #title>{{ t('sidebar.dashboard') }}</template>
      </el-menu-item>
      <el-menu-item index="/system/overview">
        <el-icon><Monitor /></el-icon>
        <template #title>{{ t('sidebar.systemOverview') }}</template>
      </el-menu-item>
      <el-menu-item index="/clusters">
        <el-icon><Connection /></el-icon>
        <template #title>{{ t('sidebar.clusters') }}</template>
      </el-menu-item>
      <el-sub-menu index="workloads">
        <template #title>
          <el-icon><Box /></el-icon>
          <span>{{ t('sidebar.workloads') }}</span>
        </template>
        <el-menu-item index="/workloads/pods">
          <el-icon><Coin /></el-icon>
          <template #title>{{ t('sidebar.pods') }}</template>
        </el-menu-item>
        <el-menu-item index="/workloads/deployments">
          <el-icon><DocumentCopy /></el-icon>
          <template #title>{{ t('sidebar.deployments') }}</template>
        </el-menu-item>
        <el-menu-item index="/workloads/statefulsets">
          <el-icon><List /></el-icon>
          <template #title>{{ t('sidebar.statefulsets') }}</template>
        </el-menu-item>
        <el-menu-item index="/workloads/daemonsets">
          <el-icon><SetUp /></el-icon>
          <template #title>{{ t('sidebar.daemonsets') }}</template>
        </el-menu-item>
        <el-menu-item index="/workloads/jobs">
          <el-icon><Finished /></el-icon>
          <template #title>{{ t('sidebar.jobs') }}</template>
        </el-menu-item>
        <el-menu-item index="/workloads/cronjobs">
          <el-icon><Timer /></el-icon>
          <template #title>{{ t('sidebar.cronjobs') }}</template>
        </el-menu-item>
        <el-menu-item index="/workloads/hpa">
          <el-icon><DataLine /></el-icon>
          <template #title>{{ t('sidebar.hpa') }}</template>
        </el-menu-item>
        <el-menu-item index="/workloads/pdb">
          <el-icon><Warning /></el-icon>
          <template #title>{{ t('sidebar.pdb') }}</template>
        </el-menu-item>
      </el-sub-menu>
      <el-sub-menu index="config">
        <template #title>
          <el-icon><Tickets /></el-icon>
          <span>{{ t('sidebar.config') }}</span>
        </template>
        <el-menu-item index="/config/configmaps">
          <el-icon><Tickets /></el-icon>
          <template #title>{{ t('sidebar.configmaps') }}</template>
        </el-menu-item>
        <el-menu-item index="/config/secrets">
          <el-icon><Key /></el-icon>
          <template #title>{{ t('sidebar.secrets') }}</template>
        </el-menu-item>
        <el-menu-item index="/config/resourcequotas">
          <el-icon><Coin /></el-icon>
          <template #title>{{ t('sidebar.resourcequotas') }}</template>
        </el-menu-item>
        <el-menu-item index="/config/limitranges">
          <el-icon><ScaleToOriginal /></el-icon>
          <template #title>{{ t('sidebar.limitranges') }}</template>
        </el-menu-item>
      </el-sub-menu>
      <el-sub-menu index="storage">
        <template #title>
          <el-icon><Coin /></el-icon>
          <span>{{ t('sidebar.storage') }}</span>
        </template>
        <el-menu-item index="/storage/pvs">
          <el-icon><Coin /></el-icon>
          <template #title>{{ t('sidebar.pvs') }}</template>
        </el-menu-item>
        <el-menu-item index="/storage/pvcs">
          <el-icon><Box /></el-icon>
          <template #title>{{ t('sidebar.pvcs') }}</template>
        </el-menu-item>
        <el-menu-item index="/storage/storageclasses">
          <el-icon><Files /></el-icon>
          <template #title>{{ t('sidebar.storageclasses') }}</template>
        </el-menu-item>
        <el-menu-item index="/storage/volumesnapshots">
          <el-icon><Camera /></el-icon>
          <template #title>{{ t('sidebar.volumeSnapshots') }}</template>
        </el-menu-item>
        <el-menu-item index="/storage/volumesnapshotclasses">
          <el-icon><CameraFilled /></el-icon>
          <template #title>{{ t('sidebar.volumeSnapshotClasses') }}</template>
        </el-menu-item>
      </el-sub-menu>
      <el-sub-menu index="network">
        <template #title>
          <el-icon><Share /></el-icon>
          <span>{{ t('sidebar.network') }}</span>
        </template>
        <el-menu-item index="/services">
          <el-icon><Connection /></el-icon>
          <template #title>{{ t('sidebar.services') }}</template>
        </el-menu-item>
        <el-menu-item index="/ingresses">
          <el-icon><Link /></el-icon>
          <template #title>{{ t('sidebar.ingresses') }}</template>
        </el-menu-item>
        <el-menu-item index="/network/networkpolicies">
          <el-icon><Lock /></el-icon>
          <template #title>{{ t('sidebar.networkpolicies') }}</template>
        </el-menu-item>
      </el-sub-menu>
      <el-menu-item index="/nodes">
        <el-icon><Cpu /></el-icon>
        <template #title>{{ t('sidebar.nodes') }}</template>
      </el-menu-item>
      <el-menu-item index="/namespaces">
        <el-icon><FolderOpened /></el-icon>
        <template #title>{{ t('sidebar.namespaces') }}</template>
      </el-menu-item>
      <el-menu-item index="/events">
        <el-icon><Bell /></el-icon>
        <template #title>{{ t('sidebar.events') }}</template>
      </el-menu-item>
      <el-menu-item index="/crd">
        <el-icon><Grid /></el-icon>
        <template #title>{{ t('sidebar.crd') }}</template>
      </el-menu-item>
      <el-sub-menu index="system">
        <template #title>
          <el-icon><Setting /></el-icon>
          <span>{{ t('sidebar.system') }}</span>
        </template>
        <el-menu-item index="/users">
          <el-icon><User /></el-icon>
          <template #title>{{ t('sidebar.users') }}</template>
        </el-menu-item>
        <el-menu-item index="/roles">
          <el-icon><UserFilled /></el-icon>
          <template #title>{{ t('sidebar.roles') }}</template>
        </el-menu-item>
        <el-menu-item index="/settings/auth">
          <el-icon><Setting /></el-icon>
          <template #title>{{ t('sidebar.authSettings') }}</template>
        </el-menu-item>
        <el-menu-item index="/audit">
          <el-icon><Document /></el-icon>
          <template #title>{{ t('sidebar.audit') }}</template>
        </el-menu-item>
      </el-sub-menu>
    </el-menu>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import {
  Odometer,
  Connection,
  Setting,
  User,
  UserFilled,
  Monitor,
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
  Warning,
  Grid,
  ScaleToOriginal,
  DocumentCopy,
  List,
  SetUp,
  Finished,
  Timer,
  Camera,
  CameraFilled,
} from '@element-plus/icons-vue'

defineProps<{
  isCollapse: boolean
}>()

const route = useRoute()
const { t } = useI18n()
const activeMenu = computed(() => route.path)
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
