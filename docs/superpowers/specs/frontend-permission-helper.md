# 前端权限控制基础设施方案

> 日期: 2026-09-02 | 状态: 待实施

---

## 目标

在 auth store 中新增 `canDo(clusterId, resourceGroup, verb, namespace?)` 工具函数，为后续所有页面的按钮级权限控制提供统一入口。**本次只建基础设施，不改造现有页面。**

## 背景

### 现状

| 层 | 状态 |
|---|---|
| 后端 `RequirePermission()` 中间件 | ✅ 202 条 K8s 路由全覆盖 |
| 前端 `fetchPermissions()` 加载用户绑定 | ✅ 登录后自动拉取 |
| 前端 `canAccess()` / `hasRole()` | ✅ 已有，但只判断"有无绑定"，不判断具体操作权限 |
| 前端按钮按权限显隐 | ❌ 全前端仅 6 处，其余按钮对所有用户可见 |

### 问题

- 用户点击无权限的按钮 → 后端 403 → 弹"权限不足"，体验差
- 后端权限判断基于 **resourceGroup + verb**（如 `workload` + `create`），前端缺少对应的判断函数
- 每次加新功能都要从头想权限怎么判断，没有统一工具

### 设计依据

后端角色的 permissions JSON 结构：

```json
{
  "workload":  ["read","create","update","delete","terminal"],
  "network":   ["read","create","update","delete"],
  "storage":   ["read","create","update","delete"],
  "config":    ["read","create","update","delete"],
  "node":      ["read","cordon","taint","drain","delete"],
  "namespace": ["read","create","update","delete"],
  "event":     ["read"],
  "audit":     ["read"],
  "crd":       ["read","create","update","delete"],
  "terminal":  ["terminal"],
  "cluster_mgmt": ["read"]
}
```

前端 `canDo` 使用同一份数据源（`GET /rbac/roles`），判断逻辑与后端中间件完全一致。

---

## 实施内容

### 1. 修改 `frontend/src/stores/auth.ts`

在 `fetchPermissions` 函数之后、`canAccess` 之前新增：

```typescript
import { getRoles } from '@/api/rbac'

// 角色权限定义缓存（登录后加载一次，进程生命周期内有效）
const rolePerms = ref<Map<string, Record<string, string[]>>>(new Map())

/**
 * 加载角色权限定义。登录后调用一次，将 GET /rbac/roles 的结果
 * 缓存到 rolePerms 中，供 canDo() 判断使用。
 */
async function loadRoles(): Promise<void> {
  try {
    const res: any = await getRoles()
    const data = res?.data ?? res
    const roles = Array.isArray(data) ? data : (data?.items || [])
    const map = new Map<string, Record<string, string[]>>()
    for (const r of roles) {
      map.set(r.name, r.permissions || {})
    }
    rolePerms.value = map
  } catch {
    // 加载失败不阻塞页面，canDo 会返回 false（安全降级）
  }
}

/**
 * 判断当前用户是否可以对指定资源执行指定操作。
 * 判断逻辑与后端 RequirePermission 中间件一致：
 *   1. 超级管理员直接放行
 *   2. 集群级绑定（namespace=""）覆盖该集群所有命名空间
 *   3. 命名空间级绑定精确匹配
 *   4. 查角色 permissions JSON 中 resourceGroup 是否包含 verb
 *
 * @param clusterId   集群 ID
 * @param resourceGroup 资源组（workload/network/storage/config/node/namespace/event/audit/crd/terminal/cluster_mgmt）
 * @param verb        动词（read/create/update/delete/terminal/cordon/taint/drain）
 * @param namespace   命名空间（可选，不传则只匹配集群级绑定）
 */
function canDo(clusterId: number, resourceGroup: string, verb: string, namespace?: string): boolean {
  if (!user.value) return false
  if (user.value.isSuperAdmin) return true
  if (!user.value.permissions) return false

  return user.value.permissions.some(p => {
    if (p.clusterId !== clusterId) return false
    // 集群级绑定覆盖所有 namespace
    if (p.namespace === '' || (namespace && p.namespace === namespace)) {
      const perms = rolePerms.value.get(p.roleName)
      return perms?.[resourceGroup]?.includes(verb) ?? false
    }
    return false
  })
}
```

在 return 语句中导出新函数：

```typescript
return { user, token, isLoggedIn, login, logout, setUser, fetchPermissions, loadRoles, canAccess, hasRole, canDo }
```

### 2. 修改登录流程调用 `loadRoles`

`frontend/src/stores/auth.ts` 的 `login` 函数末尾（`localStorage.setItem` 之后）追加：

```typescript
// 登录后加载角色权限定义（不阻塞登录流程）
loadRoles()
```

### 3. 修改路由守卫恢复加载

`frontend/src/router/index.ts` 的 beforeEach 守卫中，在 `fetchPermissions` 成功后追加 `loadRoles`：

```typescript
// 现有代码（约 591-603 行）
if (authStore.user && authStore.user.permissions === undefined) {
  try {
    await authStore.fetchPermissions()
    await authStore.loadRoles()   // ← 新增：刷新页面后也加载角色定义
  } catch {
    // ...
  }
}
```

---

## 涉及文件

| 文件 | 操作 | 改动量 |
|---|---|---|
| `frontend/src/stores/auth.ts` | 修改 | +40 行（loadRoles + canDo + import） |
| `frontend/src/router/index.ts` | 修改 | +1 行（loadRoles 调用） |

**总计：2 个文件，约 40 行改动。后端零改动。**

---

## 后续使用方式

基础设施建好后，新功能/改造旧页面时只需一行：

```vue
<!-- 创建 Deployment 按钮 -->
<el-button v-if="canDo(clusterId, 'workload', 'create', namespace)" type="success">
  创建
</el-button>

<!-- 删除按钮 -->
<el-button v-if="canDo(clusterId, 'workload', 'delete', namespace)" type="danger">
  删除
</el-button>

<!-- 终端按钮 -->
<el-button v-if="canDo(clusterId, 'terminal', 'terminal', namespace)">
  终端
</el-button>

<!-- Node cordon -->
<el-button v-if="canDo(clusterId, 'node', 'cordon')">
  Cordon
</el-button>

<!-- 菜单可见性 -->
<el-menu-item v-if="canDo(clusterId, 'audit', 'read')" index="/audit">
  审计日志
</el-menu-item>
```

不需要引入额外组件、不需要装饰器、不需要路由 meta——一个函数搞定。

---

## 资源组 × 动词速查表

供写 `v-if` 时参考，与后端 `RequirePermission` 的 `resolveResourceGroup` + `specialVerbOverrides` 一致：

| 页面/操作 | resourceGroup | verb |
|---|---|---|
| Deployment/StatefulSet/DaemonSet/Job/CronJob/Pod 列表查看 | workload | read |
| 创建 Deployment 等 | workload | create |
| 编辑 YAML / 扩缩容 / 重启 / 回滚 | workload | update |
| 删除 Pod/Deployment 等 | workload | delete |
| 进入终端 (exec) | terminal | terminal |
| 查看日志 | terminal | terminal |
| Service/Ingress 列表查看 | network | read |
| 创建/编辑 Service/Ingress | network | create / update |
| 删除 Service/Ingress | network | delete |
| PV/PVC/StorageClass 查看 | storage | read |
| 创建/编辑/删除 PV/PVC | storage | create / update / delete |
| ConfigMap/Secret 查看 | config | read |
| 创建/编辑/删除 ConfigMap/Secret | config | create / update / delete |
| Node 列表/详情查看 | node | read |
| Cordon/Uncordon 节点 | node | cordon |
| 修改节点标签/污点 | node | taint |
| Drain 节点 | node | drain |
| 删除节点 | node | delete |
| Namespace 列表/详情查看 | namespace | read |
| 创建/编辑/删除 Namespace | namespace | create / update / delete |
| 事件查看 | event | read |
| 审计日志查看 | audit | read |
| CRD/自定义资源查看 | crd | read |
| 创建/编辑/删除 CRD 资源 | crd | create / update / delete |
| 集群管理（成员/版本信息） | cluster_mgmt | read |

---

## 验收

- [ ] `npm run build` 零错误
- [ ] 登录后 `localStorage` 中 `rolePerms` 缓存有数据（可在 devtools 中检查）
- [ ] 在任意页面 console 中调 `useAuthStore().canDo(clusterId, 'workload', 'read')` 返回正确布尔值
- [ ] 超管用户 `canDo` 任何参数均返回 `true`
- [ ] 无绑定用户 `canDo` 任何参数均返回 `false`
- [ ] 现有页面行为无变化（本次不改动任何页面模板）
