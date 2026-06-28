<template>
  <div class="header">
    <div class="header-left">
      <el-icon
        class="collapse-btn"
        role="button"
        tabindex="0"
        :aria-label="t('common.toggleSidebar')"
        @click="$emit('toggleCollapse')"
        @keyup.enter="$emit('toggleCollapse')"
      >
        <Fold />
      </el-icon>
      <el-breadcrumb separator="/">
        <el-breadcrumb-item :to="{ path: '/dashboard' }">{{ t('common.home') }}</el-breadcrumb-item>
        <el-breadcrumb-item
          v-for="item in breadcrumbs"
          :key="item.path || item.title"
          :to="item.to"
        >
          {{ item.title }}
        </el-breadcrumb-item>
      </el-breadcrumb>
    </div>
    <div class="header-right">
      <!-- Cluster Selector -->
      <el-select
        v-model="clusterStore.currentCluster"
        value-key="id"
        :placeholder="t('common.selectCluster')"
        :loading="clusterLoading"
        size="small"
        class="cluster-select"
        clearable
        @change="handleClusterChange"
      >
        <template #prefix>
          <el-icon><Connection /></el-icon>
        </template>
        <el-option
          v-for="c in clusterStore.clusterList"
          :key="c.id"
          :label="c.clusterName"
          :value="c"
        />
      </el-select>

      <!-- Language Switcher -->
      <el-dropdown @command="handleLangChange">
        <el-button size="small" text class="header-action-btn">
          <el-icon><Switch /></el-icon>
        </el-button>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item command="zh-CN">中文</el-dropdown-item>
            <el-dropdown-item command="en">English</el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>

      <!-- Theme Toggle -->
      <el-tooltip :content="isDark ? t('common.lightMode') : t('common.darkMode')" placement="bottom">
        <el-button size="small" text class="header-action-btn" @click="toggle()">
          <el-icon :size="18">
            <Sunny v-if="isDark" />
            <Moon v-else />
          </el-icon>
        </el-button>
      </el-tooltip>

      <!-- User Menu -->
      <el-dropdown @command="handleCommand">
        <div class="user-info">
          <el-avatar :size="32" class="user-avatar">
            {{ (authStore.user?.username || '?')[0].toUpperCase() }}
          </el-avatar>
          <span class="username">{{ authStore.user?.displayName || authStore.user?.username || '-' }}</span>
          <el-icon class="user-arrow"><ArrowDown /></el-icon>
        </div>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item disabled>
              <el-icon><User /></el-icon>
              {{ authStore.user?.username }}
            </el-dropdown-item>
            <el-dropdown-item divided command="logout">
              <el-icon><SwitchButton /></el-icon>
              {{ t('common.logout') }}
            </el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '@/stores/auth'
import { useClusterStore } from '@/stores/cluster'
import { useTheme } from '@/styles/theme-switcher'
import {
  Fold,
  Switch,
  ArrowDown,
  User,
  SwitchButton,
  Connection,
  Sunny,
  Moon,
} from '@element-plus/icons-vue'

defineEmits(['toggleCollapse'])
const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const clusterStore = useClusterStore()
const { locale, t } = useI18n()
const clusterLoading = ref(false)
const { isDark, toggle } = useTheme()

const breadcrumbs = computed(() => {
  const items: Array<{ title: string; path?: string; to?: { path: string } }> = []

  if (route.meta?.parent) {
    const parentRoute = router.getRoutes().find(r => r.name === route.meta.parent)
    if (parentRoute?.meta?.title) {
      items.push({
        title: parentRoute.meta.title as string,
        path: parentRoute.path,
        to: { path: parentRoute.path },
      })
    }
  }

  if (route.meta?.title) {
    items.push({ title: route.meta.title as string })
  }

  return items
})

onMounted(async () => {
  clusterLoading.value = true
  try {
    await clusterStore.fetchClusters()
  } finally {
    clusterLoading.value = false
  }
})

function handleLangChange(lang: string) {
  locale.value = lang
  localStorage.setItem('gkube_locale', lang)
}

function handleClusterChange(val: any) {
  clusterStore.setCurrentCluster(val || null)
}

function handleCommand(command: string) {
  if (command === 'logout') {
    authStore.logout()
    router.push('/login')
  }
}
</script>

<style scoped>
.header {
  height: var(--gk-header-height);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 var(--gk-space-5);
  background: var(--gk-color-bg-header);
  border-bottom: 1px solid var(--gk-color-border);
  box-shadow: var(--gk-shadow-sm);
}

.header-left {
  display: flex;
  align-items: center;
  gap: var(--gk-space-4);
}

.collapse-btn {
  font-size: var(--gk-font-size-2xl);
  cursor: pointer;
  color: var(--gk-color-text-secondary);
  transition: color var(--gk-transition-fast);
  padding: var(--gk-space-1);
  border-radius: var(--gk-radius-sm);
}

.collapse-btn:hover,
.collapse-btn:focus-visible {
  color: var(--gk-color-primary);
  background: var(--gk-color-primary-bg);
  outline: none;
}

.header-right {
  display: flex;
  align-items: center;
  gap: var(--gk-space-2);
}

.cluster-select {
  width: 200px;
}

.cluster-select :deep(.el-input__wrapper) {
  border-radius: var(--gk-radius-md);
}

.header-action-btn {
  color: var(--gk-color-text-secondary);
  border-radius: var(--gk-radius-md);
  padding: var(--gk-space-2);
}

.header-action-btn:hover {
  color: var(--gk-color-primary);
  background: var(--gk-color-primary-bg);
}

.user-info {
  display: flex;
  align-items: center;
  gap: var(--gk-space-2);
  cursor: pointer;
  padding: var(--gk-space-1) var(--gk-space-2);
  border-radius: var(--gk-radius-md);
  transition: background-color var(--gk-transition-fast);
}

.user-info:hover {
  background: var(--gk-neutral-100);
}

.user-avatar {
  background: var(--gk-color-primary);
  color: var(--gk-white);
  font-weight: 600;
  font-size: var(--gk-font-size-sm);
}

.username {
  font-size: var(--gk-font-size-base);
  color: var(--gk-color-text-primary);
  font-weight: 500;
  max-width: 120px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.user-arrow {
  color: var(--gk-color-text-secondary);
  font-size: var(--gk-font-size-xs);
}
</style>
