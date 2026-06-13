<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { Bell, Refresh, Delete, Setting, Check } from '@element-plus/icons-vue'
import request from '@/api/request'

const { t } = useI18n()
const loading = ref(false)
const notifications = ref<any[]>([])
const selectedType = ref('')
const showUnreadOnly = ref(false)
const showSettings = ref(false)

// Notification settings
const settings = ref({
  email: true,
  slack: false,
  webhook: false,
  sound: true,
  desktop: true,
})

const filteredNotifications = computed(() => {
  let result = notifications.value
  if (selectedType.value) {
    result = result.filter(n => n.type === selectedType.value)
  }
  if (showUnreadOnly.value) {
    result = result.filter(n => !n.read)
  }
  return result
})

const unreadCount = computed(() => {
  return notifications.value.filter(n => !n.read).length
})

function notificationIcon(type: string) {
  switch (type) {
    case 'warning': return '⚠️'
    case 'error': return '❌'
    case 'success': return '✅'
    case 'info': return 'ℹ️'
    default: return '🔔'
  }
}

function notificationColor(type: string) {
  switch (type) {
    case 'warning': return 'warning'
    case 'error': return 'danger'
    case 'success': return 'success'
    case 'info': return 'info'
    default: return 'info'
  }
}

function formatTime(time: string) {
  if (!time) return '-'
  const date = new Date(time)
  const now = new Date()
  const diff = now.getTime() - date.getTime()
  const minutes = Math.floor(diff / 60000)
  const hours = Math.floor(diff / 3600000)
  const days = Math.floor(diff / 86400000)

  if (minutes < 1) return t('notification.justNow')
  if (minutes < 60) return t('notification.minutesAgo', { n: minutes })
  if (hours < 24) return t('notification.hoursAgo', { n: hours })
  return t('notification.daysAgo', { n: days })
}

async function fetchNotifications() {
  loading.value = true
  try {
    // Simulate notifications from events
    const res: any = await request.get('/k8s/event/list')
    const events = res.data || []

    notifications.value = events.slice(0, 50).map((e: any, i: number) => ({
      id: i + 1,
      type: e.type === 'Warning' ? 'warning' : 'info',
      title: e.reason || 'Event',
      message: e.message || '',
      resource: e.involvedObject?.name || '',
      namespace: e.namespace || '',
      timestamp: e.lastTimestamp || e.eventTime,
      read: i > 10,
    }))
  } catch {
    notifications.value = []
  } finally {
    loading.value = false
  }
}

function markAsRead(id: number) {
  const notification = notifications.value.find(n => n.id === id)
  if (notification) {
    notification.read = true
  }
}

function markAllAsRead() {
  notifications.value.forEach(n => n.read = true)
  ElMessage.success(t('notification.markedAllRead'))
}

function deleteNotification(id: number) {
  notifications.value = notifications.value.filter(n => n.id !== id)
}

function clearAll() {
  notifications.value = []
  ElMessage.success(t('notification.clearedAll'))
}

onMounted(fetchNotifications)
</script>

<template>
  <div class="page-container">
    <el-card shadow="never" class="filter-card">
      <div class="filter-bar">
        <div class="filter-left">
          <h3 style="margin: 0;"><el-icon><Bell /></el-icon> {{ t('notification.title') }}</h3>
          <el-badge :value="unreadCount" :max="99" v-if="unreadCount > 0">
            <el-tag type="danger" size="small">{{ unreadCount }} {{ t('notification.unread') }}</el-tag>
          </el-badge>
        </div>
        <div class="filter-right">
          <el-select v-model="selectedType" :placeholder="t('notification.allTypes')" clearable style="width: 120px;">
            <el-option :label="t('notification.info')" value="info" />
            <el-option :label="t('notification.warning')" value="warning" />
            <el-option :label="t('notification.error')" value="error" />
            <el-option :label="t('notification.success')" value="success" />
          </el-select>
          <el-checkbox v-model="showUnreadOnly">{{ t('notification.unreadOnly') }}</el-checkbox>
          <el-button @click="markAllAsRead"><el-icon><Check /></el-icon> {{ t('notification.markAllRead') }}</el-button>
          <el-button type="danger" @click="clearAll"><el-icon><Delete /></el-icon> {{ t('notification.clearAll') }}</el-button>
          <el-button @click="showSettings = true"><el-icon><Setting /></el-icon></el-button>
          <el-button type="primary" @click="fetchNotifications"><el-icon><Refresh /></el-icon> {{ t('common.refresh') }}</el-button>
        </div>
      </div>
    </el-card>

    <el-card shadow="never">
      <div class="notification-list">
        <div
          v-for="notification in filteredNotifications"
          :key="notification.id"
          class="notification-item"
          :class="{ unread: !notification.read }"
          @click="markAsRead(notification.id)"
        >
          <div class="notification-icon">{{ notificationIcon(notification.type) }}</div>
          <div class="notification-content">
            <div class="notification-header">
              <span class="notification-title">{{ notification.title }}</span>
              <span class="notification-time">{{ formatTime(notification.timestamp) }}</span>
            </div>
            <div class="notification-message">{{ notification.message }}</div>
            <div class="notification-meta">
              <el-tag :type="notificationColor(notification.type)" size="small">{{ notification.type }}</el-tag>
              <span v-if="notification.resource" class="notification-resource">{{ notification.resource }}</span>
              <span v-if="notification.namespace" class="notification-namespace">{{ notification.namespace }}</span>
            </div>
          </div>
          <div class="notification-actions">
            <el-button size="small" @click.stop="deleteNotification(notification.id)">
              <el-icon><Delete /></el-icon>
            </el-button>
          </div>
        </div>
      </div>

      <el-empty v-if="!loading && filteredNotifications.length === 0" :description="t('notification.noNotifications')" />
    </el-card>

    <!-- Settings Dialog -->
    <el-dialog v-model="showSettings" :title="t('notification.settings')" width="400px">
      <el-form :model="settings" label-width="100px">
        <el-form-item :label="t('notification.emailNotification')">
          <el-switch v-model="settings.email" />
        </el-form-item>
        <el-form-item :label="t('notification.slackNotification')">
          <el-switch v-model="settings.slack" />
        </el-form-item>
        <el-form-item :label="t('notification.webhook')">
          <el-switch v-model="settings.webhook" />
        </el-form-item>
        <el-form-item :label="t('notification.soundAlert')">
          <el-switch v-model="settings.sound" />
        </el-form-item>
        <el-form-item :label="t('notification.desktopNotification')">
          <el-switch v-model="settings.desktop" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showSettings = false">{{ t('common.cancel') }}</el-button>
        <el-button type="primary" @click="showSettings = false; ElMessage.success(t('notification.settingsSaved'))">{{ t('common.save') }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.page-container { padding: 20px; }
.filter-card { margin-bottom: 16px; }
.filter-bar { display: flex; justify-content: space-between; align-items: center; }
.filter-left { display: flex; align-items: center; gap: 12px; }
.filter-right { display: flex; align-items: center; gap: 8px; }
.notification-list { max-height: 600px; overflow-y: auto; }
.notification-item { display: flex; align-items: flex-start; padding: 16px; border-bottom: 1px solid #ebeef5; cursor: pointer; transition: background-color 0.2s; }
.notification-item:hover { background-color: #f5f7fa; }
.notification-item.unread { background-color: #ecf5ff; }
.notification-icon { font-size: 24px; margin-right: 12px; flex-shrink: 0; }
.notification-content { flex: 1; }
.notification-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 4px; }
.notification-title { font-weight: 500; color: #303133; }
.notification-time { font-size: 12px; color: #909399; }
.notification-message { color: #606266; font-size: 14px; margin-bottom: 8px; }
.notification-meta { display: flex; align-items: center; gap: 8px; }
.notification-resource { font-size: 12px; color: #909399; }
.notification-namespace { font-size: 12px; color: #909399; }
.notification-actions { margin-left: 12px; }
</style>
