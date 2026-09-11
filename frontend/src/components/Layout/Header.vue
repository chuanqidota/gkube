<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '@/stores/auth'
import { useTheme } from '@/styles/theme-switcher'
import ClusterSelector from './ClusterSelector.vue'
import {
  Fold,
  Switch,
  ArrowDown,
  User,
  SwitchButton,
  Sunny,
  Moon,
  Avatar,
} from '@element-plus/icons-vue'

defineEmits(['toggleCollapse'])
const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const { locale, t } = useI18n()
const { isDark, toggle } = useTheme()

function resolveTitle(meta: Record<string, unknown> | undefined): string {
  if (!meta) return ''
  if (meta.titleKey) return t(meta.titleKey as string)
  return (meta.title as string) || ''
}

const breadcrumbs = computed(() => {
  const items: Array<{ title: string; path?: string; to?: { path: string } }> = []

  if (route.meta?.parent) {
    const parentRoute = router.getRoutes().find((r) => r.name === route.meta.parent)
    if (parentRoute?.meta) {
      const parentTitle = resolveTitle(parentRoute.meta as Record<string, unknown>)
      if (parentTitle) {
        items.push({
          title: parentTitle,
          path: parentRoute.path,
          to: { path: parentRoute.path },
        })
      }
    }
  }

  const currentTitle = resolveTitle(route.meta as Record<string, unknown>)
  if (currentTitle) {
    items.push({ title: currentTitle })
  }

  return items
})

function handleLangChange(lang: string) {
  locale.value = lang
  localStorage.setItem('gkube_locale', lang)
}

async function handleCommand(command: string) {
  if (command === 'logout') {
    await authStore.logout()
    router.push('/login')
  } else if (command === 'myPermissions') {
    router.push('/my-permissions')
  }
}
</script>

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
      <ClusterSelector />

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
      <el-tooltip
        :content="isDark ? t('common.lightMode') : t('common.darkMode')"
        placement="bottom"
      >
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
          <span class="username">{{
            authStore.user?.display_name || authStore.user?.username || '-'
          }}</span>
          <el-icon class="user-arrow"><ArrowDown /></el-icon>
        </div>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item disabled>
              <el-icon><User /></el-icon>
              {{ authStore.user?.username }}
            </el-dropdown-item>
            <el-dropdown-item command="myPermissions">
              <el-icon><Avatar /></el-icon>
              {{ t('rbac.myPermissions') }}
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
  gap: var(--gk-space-3);
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
  background: var(--gk-color-primary-bg);
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
