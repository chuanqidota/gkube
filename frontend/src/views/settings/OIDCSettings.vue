<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { Check, Refresh } from '@element-plus/icons-vue'
import request from '@/api/request'

const { t } = useI18n()
const loading = ref(false)
const saving = ref(false)
const activeTab = ref('oidc')

const oidcConfig = ref({
  enabled: false,
  issuer: '',
  clientId: '',
  clientSecret: '',
  redirectUri: '',
  scopes: 'openid profile email',
  usernameClaim: 'preferred_username',
  emailClaim: 'email',
  groupsClaim: 'groups',
})

const ldapConfig = ref({
  enabled: false,
  host: '',
  port: 389,
  bindDN: '',
  bindPassword: '',
  userSearchBase: '',
  userSearchFilter: '(uid=%s)',
  groupSearchBase: '',
  groupSearchFilter: '(member=%s)',
  startTLS: false,
  insecureSkipVerify: false,
})

async function fetchConfig() {
  loading.value = true
  try {
    const res: any = await request.get('/settings/auth')
    if (res.data?.oidc) {
      Object.assign(oidcConfig.value, res.data.oidc)
    }
    if (res.data?.ldap) {
      Object.assign(ldapConfig.value, res.data.ldap)
    }
  } catch {
    // Use defaults
  } finally {
    loading.value = false
  }
}

async function handleSave() {
  saving.value = true
  try {
    await request.put('/settings/auth', {
      oidc: oidcConfig.value,
      ldap: ldapConfig.value,
    })
    ElMessage.success(t('settings.authSettingsSaved'))
  } catch (e: any) {
    ElMessage.error(e?.message || t('settings.saveFailed'))
  } finally {
    saving.value = false
  }
}

onMounted(fetchConfig)
</script>

<template>
  <div class="page-container">
    <el-card shadow="never" class="filter-card">
      <div class="filter-bar">
        <h3 style="margin: 0;">{{ t('settings.title') }}</h3>
        <div style="display: flex; gap: 8px;">
          <el-button type="primary" :loading="saving" @click="handleSave"><el-icon><Check /></el-icon> {{ t('common.save') }}</el-button>
          <el-button @click="fetchConfig"><el-icon><Refresh /></el-icon> {{ t('common.refresh') }}</el-button>
        </div>
      </div>
    </el-card>

    <el-card shadow="never">
      <el-tabs v-model="activeTab">
        <!-- OIDC Configuration -->
        <el-tab-pane :label="t('settings.oidc')" name="oidc">
          <el-form label-width="160px" style="max-width: 700px;">
            <el-form-item :label="t('settings.enableOidc')">
              <el-switch v-model="oidcConfig.enabled" />
            </el-form-item>

            <template v-if="oidcConfig.enabled">
              <el-divider>{{ t('settings.oidcProviderSettings') }}</el-divider>

              <el-form-item :label="t('settings.issuerUrl')" required>
                <el-input v-model="oidcConfig.issuer" placeholder="https://accounts.google.com" />
                <div class="form-hint">{{ t('settings.issuerUrlHint') }}</div>
              </el-form-item>

              <el-form-item :label="t('settings.clientId')" required>
                <el-input v-model="oidcConfig.clientId" placeholder="your-client-id" />
              </el-form-item>

              <el-form-item :label="t('settings.clientSecret')" required>
                <el-input v-model="oidcConfig.clientSecret" type="password" show-password placeholder="your-client-secret" />
              </el-form-item>

              <el-form-item :label="t('settings.redirectUri')">
                <el-input v-model="oidcConfig.redirectUri" placeholder="https://your-domain.com/callback" />
                <div class="form-hint">{{ t('settings.redirectUriHint') }}</div>
              </el-form-item>

              <el-form-item :label="t('settings.scopes')">
                <el-input v-model="oidcConfig.scopes" placeholder="openid profile email" />
                <div class="form-hint">{{ t('settings.scopesHint') }}</div>
              </el-form-item>

              <el-divider>{{ t('settings.claimMapping') }}</el-divider>

              <el-form-item :label="t('settings.usernameClaim')">
                <el-input v-model="oidcConfig.usernameClaim" placeholder="preferred_username" />
              </el-form-item>

              <el-form-item :label="t('settings.emailClaim')">
                <el-input v-model="oidcConfig.emailClaim" placeholder="email" />
              </el-form-item>

              <el-form-item :label="t('settings.groupsClaim')">
                <el-input v-model="oidcConfig.groupsClaim" placeholder="groups" />
                <div class="form-hint">{{ t('settings.groupsClaimHint') }}</div>
              </el-form-item>

              <el-alert :title="t('settings.restartHint')" type="warning" :closable="false" show-icon style="margin-top: 16px;" />
            </template>
          </el-form>
        </el-tab-pane>

        <!-- LDAP Configuration -->
        <el-tab-pane :label="t('settings.ldap')" name="ldap">
          <el-form label-width="180px" style="max-width: 700px;">
            <el-form-item :label="t('settings.enableLdap')">
              <el-switch v-model="ldapConfig.enabled" />
            </el-form-item>

            <template v-if="ldapConfig.enabled">
              <el-divider>{{ t('settings.ldapServerSettings') }}</el-divider>

              <el-form-item :label="t('settings.host')" required>
                <el-input v-model="ldapConfig.host" placeholder="ldap.example.com" />
              </el-form-item>

              <el-form-item :label="t('settings.port')">
                <el-input-number v-model="ldapConfig.port" :min="1" :max="65535" />
              </el-form-item>

              <el-form-item :label="t('settings.startTls')">
                <el-switch v-model="ldapConfig.startTLS" />
              </el-form-item>

              <el-form-item :label="t('settings.skipTlsVerify')">
                <el-switch v-model="ldapConfig.insecureSkipVerify" />
                <div class="form-hint">{{ t('settings.skipTlsVerifyHint') }}</div>
              </el-form-item>

              <el-divider>{{ t('settings.bindSettings') }}</el-divider>

              <el-form-item :label="t('settings.bindDn')">
                <el-input v-model="ldapConfig.bindDN" placeholder="cn=admin,dc=example,dc=com" />
              </el-form-item>

              <el-form-item :label="t('settings.bindPassword')">
                <el-input v-model="ldapConfig.bindPassword" type="password" show-password />
              </el-form-item>

              <el-divider>{{ t('settings.userSearch') }}</el-divider>

              <el-form-item :label="t('settings.userSearchBase')">
                <el-input v-model="ldapConfig.userSearchBase" placeholder="ou=users,dc=example,dc=com" />
              </el-form-item>

              <el-form-item :label="t('settings.userSearchFilter')">
                <el-input v-model="ldapConfig.userSearchFilter" placeholder="(uid=%s)" />
                <div class="form-hint">{{ t('settings.userSearchFilterHint') }}</div>
              </el-form-item>

              <el-divider>{{ t('settings.groupSearch') }}</el-divider>

              <el-form-item :label="t('settings.groupSearchBase')">
                <el-input v-model="ldapConfig.groupSearchBase" placeholder="ou=groups,dc=example,dc=com" />
              </el-form-item>

              <el-form-item :label="t('settings.groupSearchFilter')">
                <el-input v-model="ldapConfig.groupSearchFilter" placeholder="(member=%s)" />
                <div class="form-hint">{{ t('settings.groupSearchFilterHint') }}</div>
              </el-form-item>

              <el-alert :title="t('settings.restartHint')" type="warning" :closable="false" show-icon style="margin-top: 16px;" />
            </template>
          </el-form>
        </el-tab-pane>
      </el-tabs>
    </el-card>
  </div>
</template>

<style scoped>
.page-container { padding: 20px; }
.filter-card { margin-bottom: 16px; }
.filter-bar { display: flex; justify-content: space-between; align-items: center; }
.form-hint { font-size: 12px; color: var(--gk-color-text-secondary); margin-top: 4px; }
</style>
