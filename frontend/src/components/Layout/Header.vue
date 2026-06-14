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
      <el-dropdown @command="handleLangChange">
        <el-button size="small" text>
          <el-icon><Switch /></el-icon> {{ currentLang }}
        </el-button>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item command="zh-CN">中文</el-dropdown-item>
            <el-dropdown-item command="en">English</el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
      <el-select
        v-model="clusterStore.currentCluster"
        value-key="id"
        :placeholder="t('common.selectCluster')"
        :loading="clusterLoading"
        size="small"
        style="width: 200px"
        clearable
        @change="handleClusterChange"
      >
        <el-option
          v-for="c in clusterStore.clusterList"
          :key="c.id"
          :label="c.clusterName"
          :value="c"
        />
      </el-select>
      <el-dropdown @command="handleCommand">
        <div class="user-info">
          <el-avatar :size="32" style="background: #409eff">
            {{ (authStore.user?.username || '?')[0].toUpperCase() }}
          </el-avatar>
          <span class="username">{{ authStore.user?.displayName || authStore.user?.username || '-' }}</span>
          <el-icon><ArrowDown /></el-icon>
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

defineEmits(['toggleCollapse'])
const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const clusterStore = useClusterStore()
const { locale, t } = useI18n()
const clusterLoading = ref(false)

const currentLang = computed(() => locale.value === 'zh-CN' ? '中文' : 'English')

const breadcrumbs = computed(() => {
  const items: Array<{ title: string; path?: string; to?: { path: string } }> = []

  // If route has meta.parent, find the parent route and insert it first
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

  // Current page title
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
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 20px;
  background: #fff;
  border-bottom: 1px solid #e4e7ed;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.08);
}

.header-left {
  display: flex;
  align-items: center;
  gap: 16px;
}

.collapse-btn {
  font-size: 20px;
  cursor: pointer;
  color: #606266;
  transition: color 0.3s;
}

.collapse-btn:hover,
.collapse-btn:focus-visible {
  color: #409eff;
  outline: 2px solid #409eff;
  outline-offset: 2px;
  border-radius: 4px;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 20px;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  padding: 4px 8px;
  border-radius: 4px;
  transition: background-color 0.3s;
}

.user-info:hover {
  background: #f5f7fa;
}

.username {
  font-size: 14px;
  color: #606266;
}
</style>
