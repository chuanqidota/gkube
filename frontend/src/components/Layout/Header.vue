<template>
  <div class="header-container">
    <div class="header-left">
      <el-icon class="collapse-btn" @click="$emit('toggleCollapse')">
        <Fold v-if="!isMobile" />
        <Expand v-else />
      </el-icon>
      <el-breadcrumb separator="/">
        <el-breadcrumb-item
          v-for="item in breadcrumbs"
          :key="item.path"
          :to="item.path ? { path: item.path } : undefined"
        >
          {{ item.title }}
        </el-breadcrumb-item>
      </el-breadcrumb>
    </div>
    <div class="header-right">
      <el-dropdown trigger="click" @command="handleCommand">
        <span class="user-dropdown-link">
          <el-icon><UserFilled /></el-icon>
          <span class="username">{{ username }}</span>
          <el-icon class="el-icon--right"><ArrowDown /></el-icon>
        </span>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item command="logout">
              <el-icon><SwitchButton /></el-icon>
              退出登录
            </el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { removeToken } from '@/utils/auth'
import {
  Fold,
  Expand,
  UserFilled,
  ArrowDown,
  SwitchButton,
} from '@element-plus/icons-vue'
import { ElMessageBox } from 'element-plus'

defineEmits<{
  toggleCollapse: []
}>()

const route = useRoute()
const router = useRouter()
const username = ref('Admin')
const isMobile = ref(false)

interface BreadcrumbItem {
  title: string
  path?: string
}

const breadcrumbs = computed<BreadcrumbItem[]>(() => {
  const matched = route.matched
  const items: BreadcrumbItem[] = []

  for (const record of matched) {
    if (record.meta?.parent) {
      items.push({ title: record.meta.parent as string })
    }
    if (record.meta?.title) {
      items.push({
        title: record.meta.title as string,
        path: record.path || undefined,
      })
    }
  }

  return items
})

function handleCommand(command: string) {
  if (command === 'logout') {
    ElMessageBox.confirm('确定要退出登录吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning',
    }).then(() => {
      removeToken()
      router.push('/login')
    }).catch(() => {
      // cancelled
    })
  }
}
</script>

<style scoped>
.header-container {
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 20px;
  background: #fff;
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
}

.collapse-btn:hover {
  color: #409eff;
}

.header-right {
  display: flex;
  align-items: center;
}

.user-dropdown-link {
  display: flex;
  align-items: center;
  gap: 4px;
  cursor: pointer;
  color: #606266;
  font-size: 14px;
}

.user-dropdown-link:hover {
  color: #409eff;
}

.username {
  max-width: 120px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>
