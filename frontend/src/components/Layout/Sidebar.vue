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
        <template #title>{{ t('sidebar.dashboard') }}</template>
      </el-menu-item>
      <el-menu-item index="/search">
        <el-icon><Search /></el-icon>
        <template #title>{{ t('sidebar.search') }}</template>
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
          <el-icon><Files /></el-icon>
          <template #title>{{ t('sidebar.deployments') }}</template>
        </el-menu-item>
        <el-menu-item index="/workloads/statefulsets">
          <el-icon><Files /></el-icon>
          <template #title>{{ t('sidebar.statefulsets') }}</template>
        </el-menu-item>
        <el-menu-item index="/workloads/daemonsets">
          <el-icon><Files /></el-icon>
          <template #title>{{ t('sidebar.daemonsets') }}</template>
        </el-menu-item>
        <el-menu-item index="/workloads/jobs">
          <el-icon><Files /></el-icon>
          <template #title>{{ t('sidebar.jobs') }}</template>
        </el-menu-item>
        <el-menu-item index="/workloads/cronjobs">
          <el-icon><Files /></el-icon>
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
      <el-menu-item index="/rbac">
        <el-icon><UserFilled /></el-icon>
        <template #title>{{ t('sidebar.rbac') }}</template>
      </el-menu-item>
      <el-menu-item index="/crd">
        <el-icon><Grid /></el-icon>
        <template #title>{{ t('sidebar.crd') }}</template>
      </el-menu-item>
      <el-menu-item index="/catalog">
        <el-icon><Grid /></el-icon>
        <template #title>{{ t('sidebar.catalog') }}</template>
      </el-menu-item>
      <el-menu-item index="/gitops">
        <el-icon><Connection /></el-icon>
        <template #title>{{ t('sidebar.gitops') }}</template>
      </el-menu-item>
      <el-menu-item index="/tenancy">
        <el-icon><UserFilled /></el-icon>
        <template #title>{{ t('sidebar.tenancy') }}</template>
      </el-menu-item>
      <el-menu-item index="/approvals">
        <el-icon><CircleCheck /></el-icon>
        <template #title>{{ t('sidebar.approvals') }}</template>
      </el-menu-item>
      <el-sub-menu index="tools">
        <template #title>
          <el-icon><Monitor /></el-icon>
          <span>{{ t('sidebar.tools') }}</span>
        </template>
        <el-menu-item index="/terminal">
          <el-icon><Promotion /></el-icon>
          <template #title>{{ t('sidebar.terminal') }}</template>
        </el-menu-item>
        <el-menu-item index="/logs">
          <el-icon><Document /></el-icon>
          <template #title>{{ t('sidebar.logs') }}</template>
        </el-menu-item>
        <el-menu-item index="/monitoring">
          <el-icon><TrendCharts /></el-icon>
          <template #title>{{ t('sidebar.monitoring') }}</template>
        </el-menu-item>
        <el-menu-item index="/topology">
          <el-icon><Share /></el-icon>
          <template #title>{{ t('sidebar.topology') }}</template>
        </el-menu-item>
        <el-menu-item index="/tools/diff">
          <el-icon><Document /></el-icon>
          <template #title>{{ t('sidebar.diff') }}</template>
        </el-menu-item>
        <el-menu-item index="/tools/yaml-editor">
          <el-icon><Document /></el-icon>
          <template #title>{{ t('sidebar.yamlEditor') }}</template>
        </el-menu-item>
      </el-sub-menu>
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
  Promotion,
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
  TrendCharts,
  ScaleToOriginal,
  CircleCheck,
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
