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
        <template #title>仪表盘</template>
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
          <template #title>容器组</template>
        </el-menu-item>
        <el-menu-item index="/workloads/deployments">
          <el-icon><Files /></el-icon>
          <template #title>无状态负载</template>
        </el-menu-item>
        <el-menu-item index="/workloads/statefulsets">
          <el-icon><Files /></el-icon>
          <template #title>有状态负载</template>
        </el-menu-item>
        <el-menu-item index="/workloads/daemonsets">
          <el-icon><Files /></el-icon>
          <template #title>守护进程集</template>
        </el-menu-item>
        <el-menu-item index="/workloads/jobs">
          <el-icon><Files /></el-icon>
          <template #title>任务</template>
        </el-menu-item>
        <el-menu-item index="/workloads/cronjobs">
          <el-icon><Files /></el-icon>
          <template #title>定时任务</template>
        </el-menu-item>
        <el-menu-item index="/workloads/hpa">
          <el-icon><DataLine /></el-icon>
          <template #title>弹性伸缩</template>
        </el-menu-item>
        <el-menu-item index="/workloads/pdb">
          <el-icon><Warning /></el-icon>
          <template #title>中断预算</template>
        </el-menu-item>
      </el-sub-menu>
      <el-sub-menu index="config">
        <template #title>
          <el-icon><Tickets /></el-icon>
          <span>配置</span>
        </template>
        <el-menu-item index="/config/configmaps">
          <el-icon><Tickets /></el-icon>
          <template #title>配置字典</template>
        </el-menu-item>
        <el-menu-item index="/config/secrets">
          <el-icon><Key /></el-icon>
          <template #title>保密字典</template>
        </el-menu-item>
      </el-sub-menu>
      <el-sub-menu index="storage">
        <template #title>
          <el-icon><Coin /></el-icon>
          <span>存储</span>
        </template>
        <el-menu-item index="/storage/pvs">
          <el-icon><Coin /></el-icon>
          <template #title>持久卷</template>
        </el-menu-item>
        <el-menu-item index="/storage/pvcs">
          <el-icon><Box /></el-icon>
          <template #title>持久卷声明</template>
        </el-menu-item>
        <el-menu-item index="/storage/storageclasses">
          <el-icon><Files /></el-icon>
          <template #title>存储类</template>
        </el-menu-item>
      </el-sub-menu>
      <el-sub-menu index="network">
        <template #title>
          <el-icon><Share /></el-icon>
          <span>服务发现</span>
        </template>
        <el-menu-item index="/services">
          <el-icon><Connection /></el-icon>
          <template #title>服务</template>
        </el-menu-item>
        <el-menu-item index="/ingresses">
          <el-icon><Link /></el-icon>
          <template #title>路由</template>
        </el-menu-item>
        <el-menu-item index="/network/networkpolicies">
          <el-icon><Lock /></el-icon>
          <template #title>网络策略</template>
        </el-menu-item>
      </el-sub-menu>
      <el-menu-item index="/nodes">
        <el-icon><Cpu /></el-icon>
        <template #title>节点管理</template>
      </el-menu-item>
      <el-menu-item index="/namespaces">
        <el-icon><FolderOpened /></el-icon>
        <template #title>命名空间</template>
      </el-menu-item>
      <el-menu-item index="/events">
        <el-icon><Bell /></el-icon>
        <template #title>事件</template>
      </el-menu-item>
      <el-menu-item index="/rbac">
        <el-icon><UserFilled /></el-icon>
        <template #title>RBAC</template>
      </el-menu-item>
      <el-menu-item index="/crd">
        <el-icon><Grid /></el-icon>
        <template #title>CRD</template>
      </el-menu-item>
      <el-sub-menu index="tools">
        <template #title>
          <el-icon><Monitor /></el-icon>
          <span>工具</span>
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
        <el-menu-item index="/users">
          <el-icon><User /></el-icon>
          <template #title>用户管理</template>
        </el-menu-item>
        <el-menu-item index="/roles">
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
