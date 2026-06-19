# YAML Editor Optimization Design

**Date:** 2026-06-19
**Status:** Approved
**Scope:** Redesign YamlEditor component and simplify YAML dialog usage across all pages

## Problem

The current YAML editing experience has two issues:
1. **Fragmented toolbars** — Save/Cancel/Edit buttons are in the dialog, Format/Copy are in the editor toolbar, creating a messy two-layer layout
2. **Inconsistent patterns** — DeploymentList has Save/Cancel always visible, DeploymentDetail has Edit→Save/Cancel mode switching, StatefulSetList is read-only with no edit capability
3. **Plain visual design** — minimal toolbar styling, no visual polish

## Design

### Unified Toolbar in YamlEditor Component

The YamlEditor component will own the entire toolbar, combining action buttons and editor tools into one clean row:

```
┌──────────────────────────────────────────────────────────────────┐
│  [Edit] [Save✓] [Cancel]   │   [Format] [Copy]   │   Read-only  │
└──────────────────────────────────────────────────────────────────┘
│                                                                  │
│  Monaco Editor                                                   │
│                                                                  │
└──────────────────────────────────────────────────────────────────┘
```

**Left section:** Edit/Save/Cancel buttons (conditional on mode and `saveable` prop)
**Center section:** Format/Copy buttons (always visible in edit mode)
**Right section:** Mode indicator badge (Read-only / Editing)

### New Props

| Prop | Type | Default | Description |
|------|------|---------|-------------|
| `title` | `string` | `''` | Optional title shown in toolbar (resource name) |
| `saveable` | `boolean` | `false` | Show Edit/Save/Cancel buttons |

Existing props (`modelValue`, `height`, `editable`, `readOnly`, `autoFormat`) remain unchanged.

### New Events

| Event | Payload | Description |
|-------|---------|-------------|
| `save` | `content: string` | Emitted when user clicks Save. Parent handles API call. |

### Mode Switching Logic

- **Initial state:** Determined by `readOnly` prop
  - `readOnly=true` → Read-only mode, shows Edit button
  - `readOnly=false` → Edit mode, shows Save/Cancel
  - `saveable=false` → No Edit/Save/Cancel buttons at all (view-only, like StatefulSetList)
- **Click Edit** → Enters edit mode (internal `isEditing` ref), editor becomes writable
- **Click Save** → Emits `save` event with current content. Parent calls API, then can set `readOnly=true` to return to read-only mode
- **Click Cancel** → Restores original content (captured when entering edit mode), returns to read-only mode

### Keyboard Shortcuts

- `Ctrl+S` / `Cmd+S` → Save (when in edit mode)
- `Escape` → Cancel (when in edit mode)

### Dialog Simplification

After the component change, each page's YAML dialog becomes much simpler:

**Before (DeploymentList):**
```vue
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

**After (DeploymentList):**
```vue
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

**Before (DeploymentDetail):**
```vue
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

**After (DeploymentDetail):**
```vue
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

**Before (StatefulSetList — read-only):**
```vue
<el-dialog v-model="yamlDialogVisible" title="StatefulSet YAML" width="70%" top="5vh" destroy-on-close>
  <div v-loading="yamlLoading"><YamlEditor v-model="yamlContent" height="500px" read-only auto-format /></div>
</el-dialog>
```

**After (StatefulSetList — no change needed, saveable=false is default):**
```vue
<el-dialog v-model="yamlDialogVisible" title="StatefulSet YAML" width="70%" top="5vh" destroy-on-close>
  <div v-loading="yamlLoading"><YamlEditor v-model="yamlContent" height="500px" read-only auto-format /></div>
</el-dialog>
```

## Files to Modify

1. **`frontend/src/components/YamlEditor.vue`** — Major refactor: new toolbar, mode switching, save event, keyboard shortcuts
2. **`frontend/src/views/workload/DeploymentList.vue`** — Simplify YAML dialog (remove inline buttons, use saveable prop)
3. **`frontend/src/views/workload/DeploymentDetail.vue`** — Simplify YAML dialog (remove inline buttons, use saveable + readOnly props)
4. **Other list/detail pages** — Minimal or no changes (read-only pages like StatefulSetList don't need changes since saveable defaults to false)

## Out of Scope

- Changing the YAML dialog width or layout (stays at 70%)
- Adding diff view or version history
- Changing the Monaco Editor configuration
