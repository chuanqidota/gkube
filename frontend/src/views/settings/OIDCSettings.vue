<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Save, Refresh } from '@element-plus/icons-vue'
import request from '@/api/request'

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
    ElMessage.success('Authentication settings saved')
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to save settings')
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
        <h3 style="margin: 0;">Authentication Settings</h3>
        <div style="display: flex; gap: 8px;">
          <el-button type="primary" :loading="saving" @click="handleSave"><el-icon><Save /></el-icon> Save</el-button>
          <el-button @click="fetchConfig"><el-icon><Refresh /></el-icon> Refresh</el-button>
        </div>
      </div>
    </el-card>

    <el-card shadow="never">
      <el-tabs v-model="activeTab">
        <!-- OIDC Configuration -->
        <el-tab-pane label="OIDC" name="oidc">
          <el-form label-width="160px" style="max-width: 700px;">
            <el-form-item label="Enable OIDC">
              <el-switch v-model="oidcConfig.enabled" />
            </el-form-item>

            <template v-if="oidcConfig.enabled">
              <el-divider>OIDC Provider Settings</el-divider>

              <el-form-item label="Issuer URL" required>
                <el-input v-model="oidcConfig.issuer" placeholder="https://accounts.google.com" />
                <div class="form-hint">The OIDC provider's issuer URL</div>
              </el-form-item>

              <el-form-item label="Client ID" required>
                <el-input v-model="oidcConfig.clientId" placeholder="your-client-id" />
              </el-form-item>

              <el-form-item label="Client Secret" required>
                <el-input v-model="oidcConfig.clientSecret" type="password" show-password placeholder="your-client-secret" />
              </el-form-item>

              <el-form-item label="Redirect URI">
                <el-input v-model="oidcConfig.redirectUri" placeholder="https://your-domain.com/callback" />
                <div class="form-hint">The callback URL registered with the OIDC provider</div>
              </el-form-item>

              <el-form-item label="Scopes">
                <el-input v-model="oidcConfig.scopes" placeholder="openid profile email" />
                <div class="form-hint">Space-separated list of scopes to request</div>
              </el-form-item>

              <el-divider>Claim Mapping</el-divider>

              <el-form-item label="Username Claim">
                <el-input v-model="oidcConfig.usernameClaim" placeholder="preferred_username" />
              </el-form-item>

              <el-form-item label="Email Claim">
                <el-input v-model="oidcConfig.emailClaim" placeholder="email" />
              </el-form-item>

              <el-form-item label="Groups Claim">
                <el-input v-model="oidcConfig.groupsClaim" placeholder="groups" />
                <div class="form-hint">Claim containing user groups for RBAC</div>
              </el-form-item>

              <el-alert title="After saving, restart the backend service for changes to take effect." type="warning" :closable="false" show-icon style="margin-top: 16px;" />
            </template>
          </el-form>
        </el-tab-pane>

        <!-- LDAP Configuration -->
        <el-tab-pane label="LDAP" name="ldap">
          <el-form label-width="180px" style="max-width: 700px;">
            <el-form-item label="Enable LDAP">
              <el-switch v-model="ldapConfig.enabled" />
            </el-form-item>

            <template v-if="ldapConfig.enabled">
              <el-divider>LDAP Server Settings</el-divider>

              <el-form-item label="Host" required>
                <el-input v-model="ldapConfig.host" placeholder="ldap.example.com" />
              </el-form-item>

              <el-form-item label="Port">
                <el-input-number v-model="ldapConfig.port" :min="1" :max="65535" />
              </el-form-item>

              <el-form-item label="Start TLS">
                <el-switch v-model="ldapConfig.startTLS" />
              </el-form-item>

              <el-form-item label="Skip TLS Verify">
                <el-switch v-model="ldapConfig.insecureSkipVerify" />
                <div class="form-hint">Skip TLS certificate verification (not recommended for production)</div>
              </el-form-item>

              <el-divider>Bind Settings</el-divider>

              <el-form-item label="Bind DN">
                <el-input v-model="ldapConfig.bindDN" placeholder="cn=admin,dc=example,dc=com" />
              </el-form-item>

              <el-form-item label="Bind Password">
                <el-input v-model="ldapConfig.bindPassword" type="password" show-password />
              </el-form-item>

              <el-divider>User Search</el-divider>

              <el-form-item label="User Search Base">
                <el-input v-model="ldapConfig.userSearchBase" placeholder="ou=users,dc=example,dc=com" />
              </el-form-item>

              <el-form-item label="User Search Filter">
                <el-input v-model="ldapConfig.userSearchFilter" placeholder="(uid=%s)" />
                <div class="form-hint">%s will be replaced with the username</div>
              </el-form-item>

              <el-divider>Group Search</el-divider>

              <el-form-item label="Group Search Base">
                <el-input v-model="ldapConfig.groupSearchBase" placeholder="ou=groups,dc=example,dc=com" />
              </el-form-item>

              <el-form-item label="Group Search Filter">
                <el-input v-model="ldapConfig.groupSearchFilter" placeholder="(member=%s)" />
                <div class="form-hint">%s will be replaced with the user's DN</div>
              </el-form-item>

              <el-alert title="After saving, restart the backend service for changes to take effect." type="warning" :closable="false" show-icon style="margin-top: 16px;" />
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
.form-hint { font-size: 12px; color: #909399; margin-top: 4px; }
</style>
