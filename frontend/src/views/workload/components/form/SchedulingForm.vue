<script setup lang="ts">
import { Plus, Delete } from '@element-plus/icons-vue'

interface Label {
  key: string
  value: string
}
interface Tolerance {
  key: string
  operator: string
  value: string
  effect: string
  tolerationSeconds: number | null
}
interface AffinityRule {
  weight: number
  topologyKey: string
  namespaces: string
  labelKey: string
  labelValue: string
}
interface TopologySpreadConstraint {
  maxSkew: number
  topologyKey: string
  whenUnsatisfiable: string
  labelKey: string
  labelValue: string
}

const nodeSelector = defineModel<Label[]>('nodeSelector', { required: true })
const tolerations = defineModel<Tolerance[]>('tolerations', { required: true })
const podAffinityRules = defineModel<AffinityRule[]>('podAffinityRules', { required: true })
const podAntiAffinityRules = defineModel<AffinityRule[]>('podAntiAffinityRules', { required: true })
const topologySpreadConstraints = defineModel<TopologySpreadConstraint[]>(
  'topologySpreadConstraints',
  { required: true },
)

function addNodeSelector() {
  nodeSelector.value.push({ key: '', value: '' })
}

function removeNodeSelector(i: number) {
  nodeSelector.value.splice(i, 1)
}

function addToleration() {
  tolerations.value.push({
    key: '',
    operator: 'Equal',
    value: '',
    effect: 'NoSchedule',
    tolerationSeconds: null,
  })
}

function removeToleration(i: number) {
  tolerations.value.splice(i, 1)
}

function addAffinityRule(type: 'podAffinity' | 'podAntiAffinity') {
  const list = type === 'podAffinity' ? podAffinityRules : podAntiAffinityRules
  list.value.push({
    weight: 1,
    topologyKey: 'kubernetes.io/hostname',
    namespaces: '',
    labelKey: '',
    labelValue: '',
  })
}

function removeAffinityRule(type: 'podAffinity' | 'podAntiAffinity', i: number) {
  const list = type === 'podAffinity' ? podAffinityRules : podAntiAffinityRules
  list.value.splice(i, 1)
}

function addTopologySpreadConstraint() {
  topologySpreadConstraints.value.push({
    maxSkew: 1,
    topologyKey: 'kubernetes.io/hostname',
    whenUnsatisfiable: 'DoNotSchedule',
    labelKey: '',
    labelValue: '',
  })
}

function removeTopologySpreadConstraint(i: number) {
  topologySpreadConstraints.value.splice(i, 1)
}
</script>

<template>
  <el-form-item label="节点选择器">
    <div style="width: 100%">
      <div v-for="(ns, i) in nodeSelector" :key="i" class="kv-row">
        <el-input v-model="ns.key" placeholder="Key" />
        <el-input v-model="ns.value" placeholder="Value" />
        <el-button type="danger" text circle @click="removeNodeSelector(i)">
          <el-icon><Delete /></el-icon>
        </el-button>
      </div>
      <el-button text type="primary" size="small" @click="addNodeSelector">
        <el-icon><Plus /></el-icon> 添加节点选择器
      </el-button>
    </div>
  </el-form-item>

  <el-form-item label="容忍规则">
    <div style="width: 100%">
      <div v-for="(tol, i) in tolerations" :key="i" class="toleration-row">
        <el-input v-model="tol.key" placeholder="Key" />
        <el-select v-model="tol.operator" style="width: 100px">
          <el-option label="Equal" value="Equal" /><el-option label="Exists" value="Exists" />
        </el-select>
        <el-input v-model="tol.value" placeholder="Value" />
        <el-select v-model="tol.effect" style="width: 150px">
          <el-option label="NoSchedule" value="NoSchedule" />
          <el-option label="PreferNoSchedule" value="PreferNoSchedule" />
          <el-option label="NoExecute" value="NoExecute" />
        </el-select>
        <el-button type="danger" text circle @click="removeToleration(i)">
          <el-icon><Delete /></el-icon>
        </el-button>
      </div>
      <el-button text type="primary" size="small" @click="addToleration">
        <el-icon><Plus /></el-icon> 添加容忍规则
      </el-button>
    </div>
  </el-form-item>

  <!-- Pod Affinity -->
  <el-divider />
  <el-form-item label="Pod 亲和性">
    <div style="width: 100%">
      <div class="affinity-section">
        <div class="affinity-section-title">亲和规则（Pod Affinity）</div>
        <div v-for="(rule, i) in podAffinityRules" :key="i" class="affinity-row">
          <el-input v-model="rule.topologyKey" placeholder="topologyKey" style="flex: 1" />
          <el-input v-model="rule.labelKey" placeholder="标签 Key" style="width: 140px" />
          <el-input v-model="rule.labelValue" placeholder="标签 Value" style="width: 140px" />
          <el-input-number
            v-model="rule.weight"
            :min="0"
            :max="100"
            placeholder="权重"
            style="width: 100px"
          />
          <el-button type="danger" text circle @click="removeAffinityRule('podAffinity', i)">
            <el-icon><Delete /></el-icon>
          </el-button>
        </div>
        <el-button text type="primary" size="small" @click="addAffinityRule('podAffinity')">
          <el-icon><Plus /></el-icon> 添加亲和规则
        </el-button>
      </div>
      <div class="affinity-section" style="margin-top: var(--gk-space-4)">
        <div class="affinity-section-title">反亲和规则（Pod Anti-Affinity）</div>
        <div v-for="(rule, i) in podAntiAffinityRules" :key="i" class="affinity-row">
          <el-input v-model="rule.topologyKey" placeholder="topologyKey" style="flex: 1" />
          <el-input v-model="rule.labelKey" placeholder="标签 Key" style="width: 140px" />
          <el-input v-model="rule.labelValue" placeholder="标签 Value" style="width: 140px" />
          <el-input-number
            v-model="rule.weight"
            :min="0"
            :max="100"
            placeholder="权重"
            style="width: 100px"
          />
          <el-button type="danger" text circle @click="removeAffinityRule('podAntiAffinity', i)">
            <el-icon><Delete /></el-icon>
          </el-button>
        </div>
        <el-button text type="primary" size="small" @click="addAffinityRule('podAntiAffinity')">
          <el-icon><Plus /></el-icon> 添加反亲和规则
        </el-button>
      </div>
      <div style="font-size: 12px; color: var(--el-text-color-secondary); margin-top: 8px">
        权重 0 = required（必须满足），权重 1-100 = preferred（优先满足）。topologyKey 常用:
        kubernetes.io/hostname, topology.kubernetes.io/zone
      </div>
    </div>
  </el-form-item>

  <!-- Topology Spread Constraints -->
  <el-divider />
  <el-form-item label="拓扑分布约束">
    <div style="width: 100%">
      <div v-for="(tc, i) in topologySpreadConstraints" :key="i" class="topology-row">
        <el-input v-model="tc.topologyKey" placeholder="topologyKey" style="flex: 1" />
        <el-input v-model="tc.labelKey" placeholder="标签 Key" style="width: 140px" />
        <el-input v-model="tc.labelValue" placeholder="标签 Value" style="width: 140px" />
        <el-input-number v-model="tc.maxSkew" :min="1" placeholder="maxSkew" style="width: 100px" />
        <el-select v-model="tc.whenUnsatisfiable" style="width: 150px">
          <el-option label="DoNotSchedule" value="DoNotSchedule" />
          <el-option label="ScheduleAnyway" value="ScheduleAnyway" />
        </el-select>
        <el-button type="danger" text circle @click="removeTopologySpreadConstraint(i)">
          <el-icon><Delete /></el-icon>
        </el-button>
      </div>
      <el-button text type="primary" size="small" @click="addTopologySpreadConstraint">
        <el-icon><Plus /></el-icon> 添加拓扑约束
      </el-button>
      <div style="font-size: 12px; color: var(--el-text-color-secondary); margin-top: 8px">
        控制 Pod 在不同拓扑域（节点、可用区等）间的分布。maxSkew 表示最大偏差，topologyKey 常用:
        kubernetes.io/hostname, topology.kubernetes.io/zone
      </div>
    </div>
  </el-form-item>
</template>

<style scoped>
.kv-row {
  display: flex;
  gap: 8px;
  margin-bottom: 8px;
  align-items: center;
}

.toleration-row {
  display: flex;
  gap: 8px;
  margin-bottom: 8px;
  align-items: center;
  flex-wrap: wrap;
}

.affinity-section {
  width: 100%;
}

.affinity-section-title {
  font-size: 13px;
  font-weight: 500;
  color: var(--el-text-color-primary);
  margin-bottom: 8px;
}

.affinity-row {
  display: flex;
  gap: 8px;
  margin-bottom: 8px;
  align-items: center;
  flex-wrap: wrap;
}

.topology-row {
  display: flex;
  gap: 8px;
  margin-bottom: 8px;
  align-items: center;
  flex-wrap: wrap;
}
</style>
