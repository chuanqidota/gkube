# YAML Editor Optimization — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Refactor YamlEditor component with unified toolbar, mode switching, saveable prop, save event, keyboard shortcuts. Simplify YAML dialog usage in DeploymentList and DeploymentDetail.

**Architecture:** The YamlEditor component gains internal mode-switching state (`isEditing`), a unified toolbar that merges Edit/Save/Cancel with Format/Copy, and emits a `save` event. Parent pages remove their inline button boilerplate and pass `saveable` + `@save` instead.

**Tech Stack:** Vue 3, Element Plus, Monaco Editor, TypeScript

## Global Constraints

- YamlEditor.vue is a shared component used across 15+ pages — changes must be backward-compatible
- Existing props (`modelValue`, `height`, `editable`, `readOnly`, `autoFormat`) retain their current behavior
- `saveable` defaults to `false` so read-only pages (StatefulSetList, etc.) need zero changes
- The `editable` prop controls whether the editor allows text input (always true for edit-capable resources)
- Keyboard shortcuts use `Ctrl+S` (Windows/Linux) and `Cmd+S` (macOS)

---

### Task 1: Refactor YamlEditor Component

**Files:**
- Modify: `frontend/src/components/YamlEditor.vue`

**Interfaces:**
- Produces: `YamlEditor` with new props `title?: string`, `saveable?: boolean` and new event `save(content: string)`
- Backward-compatible: all existing props and behavior unchanged when `saveable=false` (default)

- [ ] **Step 1: Rewrite YamlEditor.vue**

Replace the entire file with the following:

```vue
<template>
  <div class="yaml-editor">
    <div class="yaml-editor-toolbar">
      <!-- Left: Edit/Save/Cancel -->
      <div class="toolbar-left">
        <template v-if="saveable">
          <el-button v-if="!isEditing" size="small" type="primary" @click="enterEdit">
            <el-icon><Edit /></el-icon> Edit
          </el-button>
          <template v-else>
            <el-button size="small" type="success" :loading="saving" @click="handleSave">
              <el-icon><Check /></el-icon> Save
            </el-button>
            <el-button size="small" @click="handleCancel">Cancel</el-button>
          </template>
        </template>
        <span v-if="title" class="toolbar-title">{{ title }}</span>
      </div>

      <!-- Center: Format/Copy (edit mode only) -->
      <div class="toolbar-center" v-if="isEditing">
        <el-button-group>
          <el-button size="small" @click="handleFormat">Format</el-button>
          <el-button size="small" @click="handleCopy">Copy</el-button>
        </el-button-group>
      </div>

      <!-- Right: Mode indicator -->
      <div class="toolbar-right">
        <el-tag v-if="isEditing" type="success" size="small" effect="plain">Editing</el-tag>
        <el-tag v-else-if="readOnly" type="info" size="small" effect="plain">Read-only</el-tag>
      </div>
    </div>

    <MonacoEditor
      :value="displayValue"
      :options="editorOptions"
      language="yaml"
      :style="{ height: height }"
      @update:value="handleChange"
      @mount="handleEditorMount"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch, nextTick, onMounted, onUnmounted } from 'vue'
import { Editor as MonacoEditor } from '@guolao/vue-monaco-editor'
import { ElMessage } from 'element-plus'
import { Edit, Check } from '@element-plus/icons-vue'
import yaml from 'js-yaml'

const props = withDefaults(defineProps<{
  modelValue: string
  height?: string
  editable?: boolean
  readOnly?: boolean
  autoFormat?: boolean
  title?: string
  saveable?: boolean
}>(), {
  height: '400px',
  editable: false,
  readOnly: false,
  autoFormat: false,
  title: '',
  saveable: false,
})

const emit = defineEmits(['update:modelValue', 'save'])

// Internal editing state
const isEditing = ref(false)
const originalContent = ref('')
const saving = ref(false)
const displayValue = ref('')

// Track a local formatted version for display
watch(() => props.modelValue, (val) => {
  if (props.autoFormat && val) {
    try {
      const parsed = yaml.load(val)
      const formatted = yaml.dump(parsed, {
        indent: 2,
        lineWidth: 120,
        noRefs: true,
        sortKeys: false,
      })
      displayValue.value = formatted
      if (formatted !== val) {
        emit('update:modelValue', formatted)
      }
    } catch {
      displayValue.value = val
    }
  } else {
    displayValue.value = val || ''
  }
}, { immediate: true })

// Sync isEditing when readOnly prop changes (e.g., parent resets after save)
watch(() => props.readOnly, (val) => {
  if (val) {
    isEditing.value = false
  }
})

const editorOptions = computed(() => ({
  minimap: { enabled: false },
  fontSize: 13,
  lineNumbers: 'on',
  scrollBeyondLastLine: false,
  wordWrap: 'on',
  readOnly: !isEditing.value && (props.readOnly || !props.editable),
  automaticLayout: true,
  tabSize: 2,
}))

function enterEdit() {
  originalContent.value = props.modelValue
  isEditing.value = true
}

function handleSave() {
  emit('save', props.modelValue)
}

function handleCancel() {
  isEditing.value = false
  if (originalContent.value !== props.modelValue) {
    emit('update:modelValue', originalContent.value)
  }
}

// Force Monaco to re-layout after dialog open animation
function handleEditorMount() {
  nextTick(() => {
    setTimeout(() => {
      window.dispatchEvent(new Event('resize'))
    }, 300)
  })
}

function handleChange(value: string) {
  emit('update:modelValue', value)
}

function handleFormat() {
  try {
    const parsed = yaml.load(props.modelValue)
    const formatted = yaml.dump(parsed, {
      indent: 2,
      lineWidth: 120,
      noRefs: true,
      sortKeys: false,
    })
    emit('update:modelValue', formatted)
    ElMessage.success('Formatted')
  } catch (e: any) {
    ElMessage.error('Invalid YAML: ' + (e.message || 'Format failed'))
  }
}

function handleCopy() {
  navigator.clipboard.writeText(props.modelValue)
  ElMessage.success('Copied to clipboard')
}

// Keyboard shortcuts
function handleKeydown(e: KeyboardEvent) {
  if (!isEditing.value) return
  if ((e.ctrlKey || e.metaKey) && e.key === 's') {
    e.preventDefault()
    handleSave()
  }
  if (e.key === 'Escape') {
    e.preventDefault()
    handleCancel()
  }
}

onMounted(() => {
  document.addEventListener('keydown', handleKeydown)
})

onUnmounted(() => {
  document.removeEventListener('keydown', handleKeydown)
})

// Expose saving state for parent to control
defineExpose({ saving })
</script>

<style scoped>
.yaml-editor {
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  overflow: hidden;
}
.yaml-editor-toolbar {
  padding: 6px 12px;
  background: #f5f7fa;
  border-bottom: 1px solid #dcdfe6;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}
.toolbar-left {
  display: flex;
  align-items: center;
  gap: 8px;
}
.toolbar-center {
  display: flex;
  align-items: center;
}
.toolbar-right {
  display: flex;
  align-items: center;
}
.toolbar-title {
  font-size: 13px;
  color: #909399;
  margin-left: 4px;
}
</style>
```

- [ ] **Step 2: Verify the build compiles**

Run: `cd frontend && npm run build`
Expected: Build succeeds with no errors

- [ ] **Step 3: Commit**

```bash
git add frontend/src/components/YamlEditor.vue
git commit -m "feat: refactor YamlEditor with unified toolbar, mode switching, and save event"
```

---

### Task 2: Simplify DeploymentList YAML Dialog

**Files:**
- Modify: `frontend/src/views/workload/DeploymentList.vue`

**Interfaces:**
- Consumes: `YamlEditor` with `saveable` prop and `@save` event (from Task 1)
- Produces: `handleSaveYaml(content: string)` that accepts content from the save event

- [ ] **Step 1: Remove unused YAML state refs**

In the `<script setup>` section, remove these refs that are no longer needed:
```ts
const yamlEditing = ref(false)
const yamlSaving = ref(false)
```

Keep `yamlDialogVisible`, `yamlContent`, `yamlLoading`, and `yamlTarget` — they're still needed.

- [ ] **Step 2: Simplify handleSaveYaml to accept content parameter**

Replace the current `handleSaveYaml` function:

```ts
async function handleSaveYaml() {
  if (!yamlTarget.value) return
  yamlSaving.value = true
  try {
    await updateDeploymentYaml({
      namespace: yamlTarget.value.namespace,
      name: yamlTarget.value.name,
      yaml: yamlContent.value,
    })
    ElMessage.success('YAML saved successfully')
    yamlEditing.value = false
    yamlDialogVisible.value = false
    fetchDeployments()
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to save YAML')
  } finally {
    yamlSaving.value = false
  }
}
```

With:

```ts
async function handleSaveYaml(content: string) {
  if (!yamlTarget.value) return
  try {
    await updateDeploymentYaml({
      namespace: yamlTarget.value.namespace,
      name: yamlTarget.value.name,
      yaml: content,
    })
    ElMessage.success('YAML saved successfully')
    yamlDialogVisible.value = false
    fetchDeployments()
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to save YAML')
  }
}
```

- [ ] **Step 3: Simplify handleViewYaml**

Replace the current `handleViewYaml`:

```ts
async function handleViewYaml(row: any) {
  yamlTarget.value = row
  yamlEditing.value = true
  yamlDialogVisible.value = true
  yamlLoading.value = true
  yamlContent.value = ''
  try {
    const res: any = await getDeploymentYaml({
      namespace: row.namespace,
      name: row.name,
    })
    yamlContent.value = res.data?.yaml || res.data || ''
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to load YAML')
    yamlDialogVisible.value = false
  } finally {
    yamlLoading.value = false
  }
}
```

With:

```ts
async function handleViewYaml(row: any) {
  yamlTarget.value = row
  yamlDialogVisible.value = true
  yamlLoading.value = true
  yamlContent.value = ''
  try {
    const res: any = await getDeploymentYaml({
      namespace: row.namespace,
      name: row.name,
    })
    yamlContent.value = res.data?.yaml || res.data || ''
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to load YAML')
    yamlDialogVisible.value = false
  } finally {
    yamlLoading.value = false
  }
}
```

(Removed `yamlEditing.value = true` — no longer needed)

- [ ] **Step 4: Replace YAML dialog template**

Replace the current YAML dialog:

```vue
<!-- YAML Dialog -->
<el-dialog v-model="yamlDialogVisible" title="Deployment YAML" width="70%" top="5vh" destroy-on-close>
  <div style="margin-bottom: 12px; display: flex; gap: 8px;">
    <el-button type="success" :loading="yamlSaving" @click="handleSaveYaml">Save</el-button>
    <el-button @click="handleViewYaml(yamlTarget)">Cancel</el-button>
  </div>
  <div v-loading="yamlLoading">
    <YamlEditor v-model="yamlContent" height="500px" editable :read-only="!yamlEditing" auto-format />
  </div>
</el-dialog>
```

With:

```vue
<!-- YAML Dialog -->
<el-dialog v-model="yamlDialogVisible" title="Deployment YAML" width="70%" top="5vh" destroy-on-close>
  <div v-loading="yamlLoading">
    <YamlEditor
      v-model="yamlContent"
      height="600px"
      :read-only="false"
      :saveable="true"
      auto-format
      @save="handleSaveYaml"
    />
  </div>
</el-dialog>
```

- [ ] **Step 5: Remove unused icon imports**

Check if `Refresh` is still used in the template (it is — for the refresh button). No import changes needed.

- [ ] **Step 6: Verify the build compiles**

Run: `cd frontend && npm run build`
Expected: Build succeeds with no errors

- [ ] **Step 7: Commit**

```bash
git add frontend/src/views/workload/DeploymentList.vue
git commit -m "fix: simplify DeploymentList YAML dialog to use saveable prop"
```

---

### Task 3: Simplify DeploymentDetail YAML Dialog

**Files:**
- Modify: `frontend/src/views/workload/DeploymentDetail.vue`

**Interfaces:**
- Consumes: `YamlEditor` with `saveable` prop and `@save` event (from Task 1)
- Produces: `handleSaveYaml(content: string)` that accepts content from the save event

- [ ] **Step 1: Remove unused YAML state refs**

In the `<script setup>` section, remove these refs:
```ts
const yamlEditing = ref(false)
const yamlSaving = ref(false)
```

Keep `yamlContent`, `yamlLoading`, and `yamlDialogVisible`.

- [ ] **Step 2: Simplify handleSaveYaml to accept content parameter**

Replace the current `handleSaveYaml`:

```ts
async function handleSaveYaml() {
  yamlSaving.value = true
  try {
    await updateDeploymentYaml({ namespace, name, yaml: yamlContent.value })
    ElMessage.success('YAML saved successfully')
    yamlEditing.value = false
    yamlDialogVisible.value = false
    fetchDetail()
    fetchReplicaSets()
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to save YAML')
  } finally {
    yamlSaving.value = false
  }
}
```

With:

```ts
async function handleSaveYaml(content: string) {
  try {
    await updateDeploymentYaml({ namespace, name, yaml: content })
    ElMessage.success('YAML saved successfully')
    yamlDialogVisible.value = false
    fetchDetail()
    fetchReplicaSets()
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to save YAML')
  }
}
```

- [ ] **Step 3: Simplify handleOpenYaml**

Replace the current `handleOpenYaml`:

```ts
function handleOpenYaml() {
  yamlEditing.value = false
  fetchYaml()
  yamlDialogVisible.value = true
}
```

With:

```ts
function handleOpenYaml() {
  fetchYaml()
  yamlDialogVisible.value = true
}
```

(Removed `yamlEditing.value = false` — no longer needed)

- [ ] **Step 4: Replace YAML dialog template**

Replace the current YAML dialog:

```vue
<!-- YAML Dialog -->
<el-dialog v-model="yamlDialogVisible" title="YAML Editor" width="70%" top="5vh" destroy-on-close>
  <div style="margin-bottom: 12px; display: flex; gap: 8px;">
    <el-button v-if="!yamlEditing" type="primary" @click="yamlEditing = true">Edit</el-button>
    <template v-if="yamlEditing">
      <el-button type="success" :loading="yamlSaving" @click="handleSaveYaml">Save</el-button>
      <el-button @click="yamlEditing = false; fetchYaml()">Cancel</el-button>
    </template>
  </div>
  <div v-loading="yamlLoading">
    <YamlEditor v-model="yamlContent" height="600px" :read-only="!yamlEditing" />
  </div>
</el-dialog>
```

With:

```vue
<!-- YAML Dialog -->
<el-dialog v-model="yamlDialogVisible" title="YAML Editor" width="70%" top="5vh" destroy-on-close>
  <div v-loading="yamlLoading">
    <YamlEditor
      v-model="yamlContent"
      height="600px"
      :read-only="true"
      :saveable="true"
      @save="handleSaveYaml"
    />
  </div>
</el-dialog>
```

- [ ] **Step 5: Verify the build compiles**

Run: `cd frontend && npm run build`
Expected: Build succeeds with no errors

- [ ] **Step 6: Commit**

```bash
git add frontend/src/views/workload/DeploymentDetail.vue
git commit -m "fix: simplify DeploymentDetail YAML dialog to use saveable prop"
```
