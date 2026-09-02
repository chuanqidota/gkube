# gkube 权限体系重构实施方案（P1 + P2-lite，两层作用域版）

**目标读者：** 实施 AI（Claude Code / 其他编码代理）
**性质：** 逐文件、逐函数的完整实施规范。所有代码块均为最终代码，除非标注"增量修改"。
**验证命令：** 后端每步完成后执行 `cd backend && go build ./...`；前端每步完成后执行 `cd frontend && npm run build`（vue-tsc 会捕获未使用变量/导入，报错必须修复后才能进入下一步）。
**实施顺序：** 严格按文档顺序。P1 可独立交付，P2-lite 依赖 P1 已完成。禁止跳过任何"验收清单"条目。

**与旧版方案的差异：** 本版**砍掉命名空间分组（Rancher Project 模式）**，作用域收敛为两层（集群级 > 命名空间级）。多 ns 授权 = 多行绑定（P1 已有的 `namespaces[]` 批量参数展开落库），不引入 ns_group / ns_group_item 两张表。保留 P1 全部内容 + 自定义角色 CRUD + 权限矩阵编辑器 + can-i 诊断 + 我的权限页。分组如将来需要，按第 11 节的恢复路径补做。**裁剪的完整论证与代价分析见 1.5 节。**

---

## 0. 给实施 AI 的总指令

1. 文档中的代码是**唯一事实源**。与现有代码冲突时以本文档为准。
2. "新增文件"给出完整内容；"修改文件"给出精确的锚点（原代码片段）与替换内容。用 Edit 工具做精确替换，禁止重写整个未标注的文件。
3. 文中所有 Go 代码必须原样落盘（可调整 import 排序满足 gofmt）。所有 i18n 键必须同时添加到 zh-CN 和 en 两个文件。
4. 每个阶段结束必须跑验收清单（第 8 / 9 节），全部通过才算完成。
5. 不实施本文档未提及的"顺手改进"。
6. **不要实施任何与命名空间分组（ns_group）相关的代码。** 旧版方案中的分组模型、分组缓存、三分支匹配、分组 UI 均已废弃。若在本仓库其他文档中看到分组相关描述，以本文档为准。

---

## 1. 背景与目标

### 1.1 现状问题

| # | 问题 | 现状代码位置 |
|---|------|------------|
| 1 | 成员列表以"绑定"为行，同一用户多条记录，观感混乱 | `frontend/src/views/cluster/ClusterMembersDialog.vue` |
| 2 | 添加成员多选 ns 时前端循环调 N 次 API，后端已有的 `namespaces[]` 批量参数未使用 | `frontend/src/views/cluster/AddBindingDialog.vue:145-157` |
| 3 | 无法通过 UI 创建集群级授权（对话框强制选 ns 且只显示 namespace 角色） | `AddBindingDialog.vue:60-62` |
| 4 | 集群列表显示 `memberCount` 但后端从未返回，恒为 0 | `internal/cluster/cluster.go` List |
| 5 | 角色只有 6 个预置写死的，不能自定义 | `internal/rbac/model/role.go` |
| 6 | `InvalidateAllPermissions()` 无调用方，角色变更无法失效缓存 | `pkg/auth/permcache.go:59` |
| 7 | 无法诊断"用户为什么有/没有某权限" | - |

### 1.2 目标

- **P1（展示与操作重构，无存储变更）**：成员列表改为"用户中心"卡片视图（一行 = 一个用户）；添加成员改三段式（用户 → 作用域 → 角色）；批量操作走集合级 API；集群列表返回真实 memberCount。
- **P2-lite（自定义角色 + 诊断）**：角色 CRUD + 权限矩阵编辑器；can-i 权限诊断；我的权限页。多 ns 授权沿用 P1 的多行绑定模式，**不引入命名空间分组**。

### 1.3 设计原则（不可违反）

```
1. 存储粒度 = 鉴权决策粒度：绑定表保持规范化，一行 = 一条授权事实。
2. 禁止把 namespace 集合存成 JSON 数组（违反 1NF，丢失 DB 唯一约束）。
   多 ns 授权 = 多行绑定，批量入口走 CreateBindingParams.Namespaces[]。
3. 权限模型纯叠加（additive）：任一绑定授权即放行，无 deny。
4. 作用域两层：集群级 > 命名空间级。集群级 = 全部命名空间（含未来新建）
   + 无命名空间资源（node/audit 等）+ ns='' 的全集群请求；命名空间级 = 仅该 ns。
5. 前端权限控制只是 UI 优化，安全边界永远在后端中间件。
```

### 1.4 鉴权匹配语义表（实施与验收的判定基准）

| 绑定类型 | 请求特征 | 是否匹配 |
|---------|---------|---------|
| 集群级（`namespace=''`） | 该集群任意请求（含无 ns 的请求） | ✅ |
| 集群级（`namespace=''`） | 请求 ns = 任意未来新建的 ns | ✅（自动覆盖） |
| 空间级（`namespace='dev'`） | 请求 ns = `dev` | ✅ |
| 空间级（`namespace='dev'`） | 请求 ns = `''`（全集群列表） | ❌ |
| 空间级（`namespace='dev'`） | 请求 ns = `test` | ❌ |

多行绑定相互独立：用户在 dev/test 各有一行空间级绑定时，两个 ns 均放行（叠加语义），两行之间无任何联动。

### 1.5 方案合理性论证（为什么这样做是对的）

本节回答"这种方案合理吗"。逐条给出论据；每条论据对应一个可被反驳的具体设计决策。若未来某人想推翻其中某条，应先在本节对应小节里找到原始论证并明确推翻它，而不是默默改代码。

#### 1.5.1 表结构：三实体 + 一张授权事实表

**决策：** `user` / `role` / `k8s_cluster` 为实体，`permission_binding` 为唯一的多对多授权表；`role.permissions` 用 JSON 文本存储；不建 ns_group 系表。

**为什么合理：**

1. **与鉴权引擎同构（最重要的一条）。** 存储模型 = `RequirePermission` 中间件的实际判定逻辑（`permission.go:81-96`）：遍历绑定 → 两分支 namespace 匹配 → 查角色 JSON。表结构不是"理想化设计"，而是把已经在跑的判定引擎原样固化。改表结构的方案（如分组）都要求同步改中间件三分支——本方案 P1/P2-lite **零中间件改动**，鉴权路径上没有新代码，也就没有新 bug 面。
2. **1NF 与唯一约束双收益。** `idx_user_cluster_ns (user_id, cluster_id, namespace)` 三列唯一索引在 DB 层面杜绝重复授权；"同一用户同一 ns 授两次同角色"在物理上不可能发生。若把 ns 集合存 JSON 数组（反例），唯一约束失效、`WHERE namespace = ?` 变成 `LIKE`/应用层过滤、索引报废——这正是拒绝 JSON 数组存储的原因（见 1.3 原则 2）。
3. **`role.permissions` 用 JSON 是刻意的非范式化。** 权限定义是"声明式配置"而非"关系型事实"：它只被整体读写（角色编辑器一次拉全量、鉴权时整体 Unmarshal 进内存缓存），从不需要 `WHERE verb='create'` 这类关系查询。给它建 role_permission/detail 表是教科书式过度设计——多两张表、多两次 join，换来永不使用的查询能力。GORM `PermissionsMap` 运行时缓存已消化了解析开销。
4. **RBAC 标准形态。** user—role—scope 三要素即经典 RBAC（NIST RBAC 的 role assignment + 已有 K8s 原生 RBAC 的 RoleBinding 形态），无自创概念。K8s 本身也是 "Role（规则表）+ RoleBinding（主体×角色×namespace）"，本模型是其多集群版，学习成本为零。
5. **空串哨兵（`namespace=''`）优于 NULL。** MySQL 唯一索引不去重 NULL，若用 NULL 表示集群级，同一 (user, cluster, NULL) 可插入多行，唯一约束失效。空串 + `NOT NULL DEFAULT ''` 让哨兵值参与唯一约束，且与 `DEFAULT` 配合天然幂等。

#### 1.5.2 两层作用域：为什么够用，砍分组砍对了没有

**决策：** 作用域只有 集群级（`namespace=''`）> 空间级（精确 ns）两层；不引入 ns_group。

**为什么合理：**

1. **覆盖律成立：任意分组语义都可由多行绑定精确模拟。** 分组 = "一次授权覆盖一组 ns"（如 Rancher Project）。若该 ns 集合是**静态**的，多行绑定逐条列出与分组语义完全等价——授权结果（并集）相同，仅录入方式不同（N 行 vs 1 行+分组定义）。P1 的 `namespaces[]` 批量参数已经把"录入繁琐"这个唯一痛点解决：管理员选 N 个 ns 仍是一次请求。因此分组在静态场景下是**纯冗余抽象**。
2. **分组唯一不可替代的场景是"动态 ns 集合"**（新 ns 自动落入某组、按标签选 ns）。该场景在本平台当前无真实需求（用户确认），且引入成本极高：2 张新表 + 四列唯一索引重建（锁表迁移）+ nsgroupcache 第二个缓存 + 中间件三分支匹配 + 全套分组管理 UI。为一个无需求场景支付六项成本，是本裁剪的核心论据。**该决策可逆**：恢复路径（第 11 节）已论证现有数据零迁移可加回——裁剪不烧桥。
3. **两层可解释、可审计。** 排查"用户为何能访问 X"时，两分支匹配读一眼就懂；三分支 + 分组展开（组内成员变化影响所有引用该组的绑定）的推理成本显著更高。权限系统的长期维护成本主要在"看懂"，不在"跑快"。
4. **集群级 ⊃ 全部命名空间，这是语义而非近似。** 集群级绑定自动覆盖未来新建 ns（管理员无需回来补授权）+ 覆盖 node/namespace/audit/cluster_mgmt 四个无 ns 资源组 + 覆盖 ns='' 的全集群列举请求。所以"想授权全部 ns"不需要任何分组机制——直接给集群级角色即可（这正是 1.4 语义表第 1、2 行）。
5. **与 K8s 生态同构。** K8s 原生 RBAC 也是两层：ClusterRole（含 cluster-wide 绑定）与 namespace 内 Role/RoleBinding，没有"namespace 组"概念。Rancher Project 是 Rancher 私有抽象，K8s 本身没有，用户也不可能在别处习得它。

#### 1.5.3 API 设计：集合级操作按"用户×角色"组而不是按行

**决策：** 新增 `PUT/DELETE bindings/batch`（按 userId+clusterId+roleId 组操作）与 `DELETE bindings/user/:userId`（移出集群），而不仅是现有逐行 CRUD。

**为什么合理：**

1. **操作粒度 = UI 心智模型。** 成员卡片视图中管理员的操作意图是"把 lisi 的空间编辑者权限整体撤掉"，而不是"删除 binding #41、#42、#43"。按组 API 让一次 UI 操作 = 一次网络请求 = 一条 SQL（`UPDATE ... WHERE user_id=? AND cluster_id=? AND role_id=?`），无部分失败状态。若只给逐行 API（反例），前端必须循环调用，中途失败会出现"删了一半"的中间态，需要前端写补偿逻辑——把复杂度从后端一条 SQL 转移到前端状态机，负收益。
2. **批量入口收敛在 CreateBinding 一处。** `namespaces[]` 展开逻辑（`handler.go:165-199`）已存在并经过验证，P1 前端改用即可。不为多 ns 新开 `/bindings/bulk` 端点是刻意的：同一能力两个入口 = 两份校验两份 bug。重复绑定靠唯一索引 + `isDuplicateErr` 跳过，天然幂等，重复提交安全。
3. **scopeType 一致性双重校验。** batch 换角色与单条 UpdateBinding 执行同一规则（`namespace=''` 绑定只能换 cluster 角色，反之亦然）。前端下拉过滤 + 后端拒绝，UI 层可被绕过但数据层不会脏。
4. **写操作全部挂 AuditLog + RequireAdmin，且 AuditLog 在前。** 403 拒绝也被审计——权限系统的变更操作本身就是最需要审计的资产。

#### 1.5.4 缓存一致性：TTL 兜底 + 主动失效双保险

**决策：** 保留 permcache（5 分钟 TTL，sync.Map 进程内缓存），绑定变更时逐用户失效，角色变更时全量失效。

**为什么合理：**

1. **失效路径与变更路径一一对应，无第三种写入口。** 绑定只能经 Create/Update/Delete/batch 五个 handler 写 → 每条路径都调 `InvalidateUserPermissions`；角色 permissions 只能经 UpdateRole 写 → 调 `InvalidateAllPermissions`。中间件读缓存，handler 写失效，职责闭环。即使某个失效调用被遗漏，TTL 5 分钟是有界收敛上限——权限收紧的最坏延迟被硬性封顶，这在权限系统里是可接受的工程折衷（K8s 原生 RBAC 的变更传播也有类似窗口）。
2. **P2-lite 之前 `InvalidateAllPermissions` 无调用方不是缺陷而是空位。** 现状角色是写死的（seed 永不变更），全量失效没有触发时机；引入 UpdateRole 的同时接上它，问题 #6 随功能落地自然消解，无需独立修补。
3. **进程内缓存而非 Redis 的合理性：** 单实例部署（当前 docker-compose 拓扑）、读写比极高（每个 K8s API 请求都查权限）、失效要求秒级——sync.Map 全部满足。引入分布式缓存是给不存在的横向扩展预付复杂度。

#### 1.5.5 交付切分：P1 / P2-lite 两阶段的边界依据

**决策：** P1 = 展示与操作重构（零存储变更）；P2-lite = 自定义角色 + 诊断（唯一存储变更为 `role.description` 一列）。

**为什么合理：**

1. **每阶段单一风险类别。** P1 全部是"读路径重排 + 新增路由"，回滚 = git revert，无数据面动作；P2-lite 触碰存储但仅 AutoMigrate 加一列（加列不锁全表、不动现有行、可 DROP 回滚）。两个阶段都不含"改鉴权语义"的步骤——鉴权中间件全程不动，这是整个方案风险面最小的根本原因。
2. **P2-lite 砍掉的（分组、P3 同步、deny）全是"可加性"特性。** additive-only 模型下，后续加分组或加原生 RBAC 同步都只是新增代码路径，不要求改写已交付的行；反过来若先做分组再砍掉，就要做数据反向迁移。先简后繁的顺序让每一步都可回退。
3. **验收清单按阶段独立闭合。** P1 验收（第 8 节）不依赖任何 P2 内容，P1 单独交付即可解决现状问题 #1-#4；P2-lite 解决 #5-#7。问题列表与阶段一一映射，无悬空项。

#### 1.5.6 已识别的代价与接受理由（诚实清单）

方案不是免费的。以下代价已识别并**接受**，列出以免未来误判为疏忽：

| # | 代价 | 量级 | 接受理由 |
|---|------|------|---------|
| 1 | 动态 ns 集合授权不支持（ns 加入/移出需手动调整绑定） | 中（仅影响"ns 集合频繁变动"的团队） | 当前无此需求；恢复路径成本已知（第 11 节）；多行绑定可先行兜底 |
| 2 | 授权行数随"用户×ns"线性增长（100 用户×20 ns = 2000 行） | 低 | `idx_user_cluster_ns` 前缀即 `user_id`，单用户权限查询恒为索引前缀扫描，行数增长不影响鉴权路径性能 |
| 3 | 角色权限变更最长 5 分钟生效窗口 | 低 | 有界（TTL 封顶）；变更角色是低频管理操作；收紧权限的紧急场景可重启进程立即清空（sync.Map 进程内） |
| 4 | MyPermissions 显示 clusterId 数字而非集群名 | 低（体验） | 避免 cluster store 全量拉取的耦合；后续在 my-permissions 响应加 clusterName 即可（第 10 节 #6） |
| 5 | can-i 无前端 UI（仅 API） | 低 | 诊断场景低频，curl / Go 客户端可用；UI 是纯增量工作，随时可补 |

#### 1.5.7 反例对照：被拒绝的三个替代方案

为防止未来"重新发明"，记录被明确否决的路线及否决理由：

| 替代方案 | 表面吸引力 | 致命缺陷 |
|---------|-----------|---------|
| **A. permission_binding.namespace 存 JSON 数组** | 一行表示多 ns，行数少 | 破坏 1NF；`idx_user_cluster_ns` 唯一约束失效（无法对 JSON 列建联合唯一索引）；鉴权从索引查退化全表扫 + 应用层解析；与现有中间件不兼容 |
| **B. 保留命名空间分组（旧版三作用域）** | 一次授权覆盖动态 ns 集合 | 当前无需求；2 新表 + 锁表索引重建 + 第二缓存 + 三分支匹配 + 全套 UI；排查权限需展开分组闭包；被裁剪（1.5.2） |
| **C. 角色权限拆 role_permission / role_permission_verb 关系表** | 教科书范式化 | 权限定义只整体读写，从无关系查询需求；多两张表两次 join 换零收益；JSON + 运行时 map 缓存已覆盖全部访问模式 |

---

## 2. P1 总览：文件变更清单

| # | 文件 | 操作 | 内容 |
|---|------|------|------|
| 1 | `backend/internal/rbac/handler.go` | 修改 | 新增 3 个集合级 handler；改造 ClusterMembers 返回结构 |
| 2 | `backend/internal/router/rbac.go` | 修改 | 注册新路由 |
| 3 | `backend/internal/cluster/cluster.go` | 修改 | List 返回 memberCount |
| 4 | `frontend/src/api/rbac.ts` | 修改 | 新增 3 个 API 函数 |
| 5 | `frontend/src/views/cluster/ClusterMembersDialog.vue` | 重写 | 用户中心卡片视图 |
| 6 | `frontend/src/views/cluster/AddBindingDialog.vue` | 重写 | 三段式对话框 |
| 7 | `frontend/src/locales/zh-CN.ts` / `en.ts` | 修改 | 新增 i18n 键 |

P1 不改任何 model、不改中间件、不改数据库。

---

## 3. P1 后端实施

### 3.1 `backend/internal/rbac/handler.go`：新增集合级操作

在文件末尾追加以下三个 handler（文件已有 import `fmt/net/http/regexp/strconv/strings/gin/authmodel/model/auth/database/logger/response`，无需新增 import）：

```go
// --- P1 集合级批量操作 ---

// UpdateBindingBatchParams 按组换角色参数
type UpdateBindingBatchParams struct {
	UserID    uint `json:"userId" binding:"required" label:"用户ID"`
	ClusterID uint `json:"clusterId" binding:"required" label:"集群ID"`
	FromRoleID uint `json:"fromRoleId" binding:"required" label:"原角色ID"`
	ToRoleID  uint `json:"toRoleId" binding:"required" label:"新角色ID"`
}

// UpdateBindingBatch 将某用户在某集群下指定角色的全部绑定，统一换成新角色。
// 前端"按组换角色"入口：一次调用覆盖 N 个 namespace 的 N 条绑定。
func (h *rbacHandler) UpdateBindingBatch(c *gin.Context) {
	var p UpdateBindingBatchParams
	if err := c.ShouldBindJSON(&p); err != nil {
		response.Fail(c, "参数校验失败")
		return
	}

	// 校验目标角色存在
	var toRole model.Role
	if err := database.DB.First(&toRole, p.ToRoleID).Error; err != nil {
		response.Fail(c, "角色不存在")
		return
	}

	// 查出待更新绑定（拿 scopeType 做校验）
	var bindings []model.PermissionBinding
	if err := database.DB.
		Where("user_id = ? AND cluster_id = ? AND role_id = ?", p.UserID, p.ClusterID, p.FromRoleID).
		Find(&bindings).Error; err != nil {
		logger.Error(fmt.Sprintf("批量查询绑定失败: %v", err))
		response.FailWithStatus(c, http.StatusInternalServerError, "批量更新绑定失败")
		return
	}
	if len(bindings) == 0 {
		response.FailWithStatus(c, http.StatusNotFound, "未找到匹配的绑定")
		return
	}

	// scopeType 一致性校验（与单条 UpdateBinding 规则相同）:
	// 集群级绑定只能换 cluster 角色；空间级绑定只能换 namespace 角色。
	for _, b := range bindings {
		if b.Namespace == "" && toRole.ScopeType != "cluster" {
			response.Fail(c, "角色类型与绑定作用域不匹配")
			return
		}
		if b.Namespace != "" && toRole.ScopeType != "namespace" {
			response.Fail(c, "角色类型与绑定作用域不匹配")
			return
		}
	}

	// 一条 SQL 批量更新
	res := database.DB.Model(&model.PermissionBinding{}).
		Where("user_id = ? AND cluster_id = ? AND role_id = ?", p.UserID, p.ClusterID, p.FromRoleID).
		Update("role_id", p.ToRoleID)
	if res.Error != nil {
		logger.Error(fmt.Sprintf("批量更新绑定失败: %v", res.Error))
		response.FailWithStatus(c, http.StatusInternalServerError, "批量更新绑定失败")
		return
	}

	// 失效缓存
	auth.InvalidateUserPermissions(p.UserID)

	response.Success(c, "批量更新成功", gin.H{"updated": res.RowsAffected})
}

// DeleteBindingBatchParams 按组删除参数
type DeleteBindingBatchParams struct {
	UserID    uint `json:"userId" binding:"required" label:"用户ID"`
	ClusterID uint `json:"clusterId" binding:"required" label:"集群ID"`
	RoleID    uint `json:"roleId" binding:"required" label:"角色ID"`
}

// DeleteBindingBatch 删除某用户在某集群下指定角色的全部绑定（按组删除）。
func (h *rbacHandler) DeleteBindingBatch(c *gin.Context) {
	var p DeleteBindingBatchParams
	if err := c.ShouldBindJSON(&p); err != nil {
		response.Fail(c, "参数校验失败")
		return
	}

	res := database.DB.Where("user_id = ? AND cluster_id = ? AND role_id = ?",
		p.UserID, p.ClusterID, p.RoleID).Delete(&model.PermissionBinding{})
	if res.Error != nil {
		logger.Error(fmt.Sprintf("批量删除绑定失败: %v", res.Error))
		response.FailWithStatus(c, http.StatusInternalServerError, "批量删除绑定失败")
		return
	}
	if res.RowsAffected == 0 {
		response.FailWithStatus(c, http.StatusNotFound, "未找到匹配的绑定")
		return
	}

	// 失效缓存
	auth.InvalidateUserPermissions(p.UserID)

	response.Success(c, "批量删除成功", gin.H{"deleted": res.RowsAffected})
}

// RemoveClusterMemberParams 移出集群参数
type RemoveClusterMemberParams struct {
	UserID    uint `json:"userId" binding:"required" label:"用户ID"`
	ClusterID uint `json:"clusterId" binding:"required" label:"集群ID"`
}

// RemoveClusterMember 移出集群：删除该用户在该集群下的所有绑定（不限角色）。
// 路由为 DELETE /rbac/bindings/user/:userId，clusterId 走 query。
func (h *rbacHandler) RemoveClusterMember(c *gin.Context) {
	userID, ok := parseID(c)
	if !ok {
		return
	}
	clusterIDStr := c.Query("clusterId")
	clusterID, err := strconv.ParseUint(clusterIDStr, 10, 64)
	if err != nil || clusterID == 0 {
		response.Fail(c, "无效的集群ID")
		return
	}

	res := database.DB.Where("user_id = ? AND cluster_id = ?", userID, clusterID).
		Delete(&model.PermissionBinding{})
	if res.Error != nil {
		logger.Error(fmt.Sprintf("移出集群失败: %v", err))
		response.FailWithStatus(c, http.StatusInternalServerError, "移出集群失败")
		return
	}

	// 失效缓存（即使 RowsAffected==0 也无害）
	auth.InvalidateUserPermissions(uint(userID))

	response.Success(c, "已移出集群", gin.H{"deleted": res.RowsAffected})
}
```

### 3.2 `backend/internal/rbac/handler.go`：改造 ClusterMembers

**替换整个 `ClusterMembers` 函数**（原函数体从 `// ClusterMembers 返回某集群的全部绑定` 注释开始到函数结尾）。新版本把绑定按用户分组，返回用户中心结构：

```go
// ClusterMembers 返回某集群的成员（按用户分组），所有已认证用户可访问。
// P1 返回结构：items[] = { userId, username, displayName, isSuperAdmin, bindings[] }
// bindings[] 每项 = { roleId, roleName, roleDisplayName, scopeType, namespace }
func (h *rbacHandler) ClusterMembers(c *gin.Context) {
	var q ClusterMembersQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		response.Fail(c, "参数校验失败")
		return
	}

	var bindings []model.PermissionBinding
	if err := database.DB.Preload("User").Preload("Role").
		Where("cluster_id = ?", q.ClusterID).
		Order("id ASC").Find(&bindings).Error; err != nil {
		logger.Error(fmt.Sprintf("查询集群成员失败: %v", err))
		response.FailWithStatus(c, http.StatusInternalServerError, "查询集群成员失败")
		return
	}

	// 按用户分组（保持首次出现顺序）
	type bindingInfo struct {
		BindingID       uint   `json:"bindingId"`
		RoleID          uint   `json:"roleId"`
		RoleName        string `json:"roleName"`
		RoleDisplayName string `json:"roleDisplayName"`
		ScopeType       string `json:"scopeType"`  // cluster|namespace
		Namespace       string `json:"namespace"`  // cluster 级为 ""
	}
	type memberInfo struct {
		UserID       uint          `json:"userId"`
		Username     string        `json:"username"`
		DisplayName  string        `json:"displayName"`
		IsSuperAdmin bool          `json:"isSuperAdmin"`
		Bindings     []bindingInfo `json:"bindings"`
	}

	order := make([]uint, 0, 8)
	members := make(map[uint]*memberInfo, 8)
	for _, b := range bindings {
		m, exists := members[b.UserID]
		if !exists {
			m = &memberInfo{
				UserID:       b.UserID,
				Username:     b.User.Username,
				DisplayName:  b.User.DisplayName,
				IsSuperAdmin: b.User.IsSuperAdmin,
				Bindings:     make([]bindingInfo, 0, 4),
			}
			members[b.UserID] = m
			order = append(order, b.UserID)
		}
		m.Bindings = append(m.Bindings, bindingInfo{
			BindingID:       b.ID,
			RoleID:          b.RoleID,
			RoleName:        b.Role.Name,
			RoleDisplayName: b.Role.DisplayName,
			ScopeType:       b.Role.ScopeType,
			Namespace:       b.Namespace,
		})
	}

	items := make([]memberInfo, 0, len(order))
	for _, uid := range order {
		items = append(items, *members[uid])
	}

	response.Success(c, "获取集群成员成功", gin.H{
		"items": items,
		"total": len(items),
	})
}
```

### 3.3 `backend/internal/router/rbac.go`：注册新路由

**替换整个文件**为：

```go
package router

import (
	"github.com/gin-gonic/gin"
	"gkube/internal/rbac"
	"gkube/pkg/middleware"
)

// registerRbacRoutes 注册 RBAC 权限管理路由
func registerRbacRoutes(rg *gin.RouterGroup) {
	// 所有已认证用户可访问（普通用户查自己权限、查看角色列表、查看集群成员）
	rg.GET("rbac/roles", rbac.RbacHandler.ListRoles)
	rg.GET("rbac/my-permissions", rbac.RbacHandler.MyPermissions)
	rg.GET("rbac/cluster-members", rbac.RbacHandler.ClusterMembers)

	// 仅管理员可访问（写操作 + 管理查询）
	// AuditLog 放在 RequireAdmin 之前，确保 403 失败尝试也被审计
	admin := rg.Group("rbac", middleware.AuditLog(), middleware.RequireAdmin())
	{
		admin.GET("bindings", rbac.RbacHandler.ListBindings)
		admin.POST("bindings", rbac.RbacHandler.CreateBinding)
		admin.PUT("bindings/batch", rbac.RbacHandler.UpdateBindingBatch)
		admin.DELETE("bindings/batch", rbac.RbacHandler.DeleteBindingBatch)
		admin.DELETE("bindings/user/:userId", rbac.RbacHandler.RemoveClusterMember)
		admin.PUT("bindings/:id", rbac.RbacHandler.UpdateBinding)
		admin.DELETE("bindings/:id", rbac.RbacHandler.DeleteBinding)
	}
}
```

**关键约束（gin 路由树冲突）：** `bindings/batch`、`bindings/user/:userId` 必须注册在 `bindings/:id` **之前**。gin 的 httprouter 会把 `PUT bindings/:id` 与 `PUT bindings/batch` 视为同一位置的静态/参数冲突，静态段优先匹配，注册顺序无关，但**删除** `bindings/:id` 与 `bindings/user/:userId` 时 `:id` 与 `:userId` 是不同参数名会 panic。因此**必须**：
- 保留 `PUT bindings/:id`（与 `PUT bindings/batch` 静态段共存没问题——gin v1.10 允许静态路由与参数路由共存于同一层级）；
- `DELETE bindings/user/:userId` 与 `DELETE bindings/:id` 参数名不同——**若 gin 报 panic，则把该路由改为 `DELETE bindings/user`（userID 改走 query 参数 `?userId=`）**。实施时先按上文写入，`go build` 通过后再用 curl 实测路由注册；若启动 panic，采用 query 参数变体并同步修改 `RemoveClusterMember`：删掉 `parseID(c)` 逻辑，改为 `userID, err := strconv.ParseUint(c.Query("userId"), 10, 64)`。前端对应调用也改为 `request.delete('/rbac/bindings/user', { params: { userId, clusterId } })`。

### 3.4 `backend/internal/cluster/cluster.go`：List 返回 memberCount

**定位锚点**（cluster.go List 函数末尾）：

```go
	response.Success(c, "获取集群列表成功", gin.H{
		"items": clusters,
		"total": total,
	})
}
```

**替换为**：

```go
	// 统计每个集群的成员数（绑定表按 user 去重）
	type memberRow struct {
		ClusterID uint
		Cnt       int64
	}
	var memberRows []memberRow
	if err := database.DB.Model(&rbacmodel.PermissionBinding{}).
		Select("cluster_id, COUNT(DISTINCT user_id) AS cnt").
		Group("cluster_id").Scan(&memberRows).Error; err != nil {
		logger.Error(fmt.Sprintf("统计集群成员数失败: %v", err))
		// 统计失败不阻塞列表返回，成员数显示为 0
		memberRows = nil
	}
	memberCount := make(map[uint]int64, len(memberRows))
	for _, r := range memberRows {
		memberCount[r.ClusterID] = r.Cnt
	}

	// 构造带成员数的返回项（不直接改 model，避免污染其他调用方）
	type clusterItem struct {
		model.K8SCluster
		MemberCount int64 `json:"memberCount"`
	}
	items := make([]clusterItem, 0, len(clusters))
	for _, cl := range clusters {
		items = append(items, clusterItem{K8SCluster: cl, MemberCount: memberCount[cl.ID]})
	}

	response.Success(c, "获取集群列表成功", gin.H{
		"items": items,
		"total": total,
	})
}
```

**import 增量：** 文件头部 import 块中新增 `"gkube/internal/rbac/model"`，别名 `rbacmodel`。锚点：

```go
import (
	// ...现有 import...
)
```

替换为（把 rbacmodel 行插入，保持字母序）：

```go
import (
	// ...现有 import 行保持不变，插入这一行（按字母序放置）...
	rbacmodel "gkube/internal/rbac/model"
)
```

注意：`internal/cluster` 导入 `internal/rbac/model` 不会产生循环依赖（rbac/model 只依赖 auth/model 与 cluster/model，不依赖 cluster 包本身）。

### 3.5 P1 后端验收

```bash
cd backend && go build ./...
```

编译通过后，手动冒烟（需要 MySQL 运行）：

```bash
go run main.go migrate && go run main.go seed
go run main.go &
# 用 admin 登录拿 token（密码来自 seed 输出）
TOKEN=<accessToken>
curl -s "http://localhost:8080/v1/rbac/cluster-members?clusterId=1" -H "Authorization: Bearer $TOKEN"
# 期望: items 数组元素含 userId/username/bindings 数组
curl -s -X PUT "http://localhost:8080/v1/rbac/bindings/batch" -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"userId":2,"clusterId":1,"fromRoleId":5,"toRoleId":6}'
# 期望: {"msg":"批量更新成功","code":200,"data":{"updated":N}}
curl -s "http://localhost:8080/v1/clusters?page=1&size=10" -H "Authorization: Bearer $TOKEN"
# 期望: items 元素含 memberCount 字段且与绑定表一致
```

---

## 4. P1 前端实施

### 4.1 `frontend/src/api/rbac.ts`：新增 API

**在文件末尾（`searchUsers` 之后）追加：**

```ts
// 按组批量换角色：把某用户在某集群下 fromRoleId 的所有绑定换成 toRoleId
export const updateBindingBatch = (data: {
  userId: number
  clusterId: number
  fromRoleId: number
  toRoleId: number
}) => request.put('/rbac/bindings/batch', data)

// 按组批量删除：删除某用户在某集群下指定角色的全部绑定
export const deleteBindingBatch = (data: {
  userId: number
  clusterId: number
  roleId: number
}) => request.delete('/rbac/bindings/batch', { data })

// 移出集群：删除该用户在该集群下所有绑定
export const removeClusterMember = (userId: number, clusterId: number) =>
  request.delete(`/rbac/bindings/user/${userId}`, { params: { clusterId } })
```

若 3.3 节采用了 query 参数变体，`removeClusterMember` 改为：

```ts
export const removeClusterMember = (userId: number, clusterId: number) =>
  request.delete('/rbac/bindings/user', { params: { userId, clusterId } })
```

### 4.2 重写 `frontend/src/views/cluster/ClusterMembersDialog.vue`

**整个文件替换为以下内容**（保持 props/emits 契约与 ClusterList.vue 现有调用一致：`v-model:visible` + `:cluster-id` + `:cluster-name`）：

```vue
<script setup lang="ts">
import { ref, watch, onMounted, onUnmounted, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, User } from '@element-plus/icons-vue'
import { getRoles, getClusterMembers, deleteBindingBatch, removeClusterMember } from '@/api/rbac'
import { useAuthStore } from '@/stores/auth'
import { useUIStore } from '@/stores/ui'
import AddBindingDialog from './AddBindingDialog.vue'
import MemberEditDialog from './MemberEditDialog.vue'

interface MemberBinding {
  bindingId: number
  roleId: number
  roleName: string
  roleDisplayName: string
  scopeType: 'cluster' | 'namespace'
  namespace: string
}
interface Member {
  userId: number
  username: string
  displayName: string
  isSuperAdmin: boolean
  bindings: MemberBinding[]
}

const props = defineProps<{
  visible: boolean
  clusterId: number
  clusterName: string
}>()

const emit = defineEmits<{
  'update:visible': [val: boolean]
}>()

const { t } = useI18n()
const authStore = useAuthStore()
const uiStore = useUIStore()
const loading = ref(false)
const members = ref<Member[]>([])
const roles = ref<any[]>([])
const searchQuery = ref('')
const addDialogVisible = ref(false)
const editDialogVisible = ref(false)
const editingMember = ref<Member | null>(null)

// 响应式抽屉宽度：移动端全屏，桌面端撑到左侧菜单栏
const isMobile = ref(window.matchMedia('(max-width: 768px)').matches)
let mobileHandler: ((e: MediaQueryListEvent) => void) | undefined
const drawerSize = computed(() => {
  if (isMobile.value) return '100%'
  return uiStore.sidebarCollapsed
    ? 'calc(100vw - var(--gk-sidebar-collapsed-width))'
    : 'calc(100vw - var(--gk-sidebar-width))'
})

onMounted(() => {
  mobileHandler = (e: MediaQueryListEvent) => { isMobile.value = e.matches }
  window.matchMedia('(max-width: 768px)').addEventListener('change', mobileHandler)
})
onUnmounted(() => {
  if (mobileHandler) {
    window.matchMedia('(max-width: 768px)').removeEventListener('change', mobileHandler)
  }
})

const isAdmin = computed(() => authStore.user?.isAdmin || authStore.user?.isSuperAdmin || false)

const filteredMembers = computed(() => {
  if (!searchQuery.value.trim()) return members.value
  const query = searchQuery.value.trim().toLowerCase()
  return members.value.filter(m =>
    (m.username || '').toLowerCase().includes(query) ||
    (m.displayName || '').toLowerCase().includes(query) ||
    m.bindings.some(b =>
      (b.roleDisplayName || '').toLowerCase().includes(query) ||
      (b.namespace || '').toLowerCase().includes(query)
    )
  )
})

watch(() => props.visible, (val) => {
  if (val && props.clusterId) fetchData()
})

async function fetchData() {
  loading.value = true
  try {
    const [membersRes, rolesRes] = await Promise.all([
      getClusterMembers(props.clusterId),
      getRoles(),
    ])
    const mData: any = membersRes?.data ?? membersRes
    members.value = (mData?.items || []) as Member[]
    const rData: any = rolesRes?.data ?? rolesRes
    roles.value = Array.isArray(rData) ? rData : (rData?.items || [])
  } catch (e: any) {
    ElMessage.error(e?.message || t('rbac.loadFailed'))
  } finally {
    loading.value = false
  }
}

function handleAdd() {
  addDialogVisible.value = true
}

function handleAddSuccess() {
  fetchData()
  authStore.fetchPermissions()
}

function handleEditMember(m: Member) {
  editingMember.value = m
  editDialogVisible.value = true
}

function handleEditSuccess() {
  fetchData()
  authStore.fetchPermissions()
}

// 按组删除：删除该成员在指定角色下的全部绑定
async function handleRemoveRole(m: Member, b: MemberBinding) {
  const label = b.scopeType === 'cluster'
    ? b.roleDisplayName
    : `${b.roleDisplayName} (${b.namespace})`
  try {
    await ElMessageBox.confirm(
      t('rbac.removeRoleConfirm', { name: escapeHtml(m.username), role: escapeHtml(label) }),
      t('common.confirm'),
      { type: 'warning', dangerouslyUseHTMLString: true }
    )
  } catch { return }
  try {
    await deleteBindingBatch({ userId: m.userId, clusterId: props.clusterId, roleId: b.roleId })
    ElMessage.success(t('rbac.removeRoleSuccess'))
    fetchData()
    authStore.fetchPermissions()
  } catch (e: any) {
    ElMessage.error(e?.message || t('common.deleteFailed'))
  }
}

// 移出集群：删除该成员全部绑定
async function handleRemoveMember(m: Member) {
  try {
    await ElMessageBox.confirm(
      t('rbac.removeMemberConfirm', { name: escapeHtml(m.username) }),
      t('common.confirm'),
      { type: 'warning', dangerouslyUseHTMLString: true }
    )
  } catch { return }
  try {
    await removeClusterMember(m.userId, props.clusterId)
    ElMessage.success(t('rbac.removeMemberSuccess'))
    fetchData()
    authStore.fetchPermissions()
  } catch (e: any) {
    ElMessage.error(e?.message || t('common.deleteFailed'))
  }
}

function roleTagType(roleName: string): string {
  if (!roleName) return 'info'
  if (roleName.includes('admin')) return 'danger'
  if (roleName.includes('editor')) return 'primary'
  if (roleName.includes('viewer')) return 'success'
  return 'info'
}

function escapeHtml(str: string): string {
  const map: Record<string, string> = {
    '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;',
  }
  return str.replace(/[&<>"']/g, c => map[c])
}
</script>

<template>
  <el-drawer
    :model-value="visible"
    @update:model-value="emit('update:visible', $event)"
    :title="t('rbac.clusterMembers', { cluster: clusterName })"
    :size="drawerSize"
    destroy-on-close
  >
    <template #header>
      <div class="drawer-header">
        <div class="header-title">
          <el-icon :size="20" style="color: var(--gk-color-primary); margin-right: 8px;"><User /></el-icon>
          <span>{{ t('rbac.clusterMembers', { cluster: clusterName }) }}</span>
        </div>
        <div class="header-stats">
          <el-tag size="small" type="info">{{ t('rbac.totalMembers', { count: members.length }) }}</el-tag>
        </div>
      </div>
    </template>

    <div class="drawer-body">
      <div class="toolbar">
        <el-input
          v-model="searchQuery"
          :placeholder="t('rbac.searchMembers')"
          clearable
          style="width: 220px;"
          size="default"
        />
        <el-button v-if="isAdmin" type="primary" @click="handleAdd">
          <el-icon><Plus /></el-icon> {{ t('rbac.addMember') }}
        </el-button>
      </div>

      <div v-loading="loading" class="member-cards">
        <el-empty v-if="!loading && filteredMembers.length === 0" :description="t('rbac.noMembers')" />

        <el-card
          v-for="m in filteredMembers"
          :key="m.userId"
          shadow="never"
          class="member-card"
        >
          <div class="member-head">
            <div class="member-info">
              <el-avatar :size="36" class="member-avatar">
                {{ (m.username || '?')[0].toUpperCase() }}
              </el-avatar>
              <div class="member-names">
                <span class="username">{{ m.username }}</span>
                <span v-if="m.displayName" class="display-name">{{ m.displayName }}</span>
              </div>
            </div>
            <div v-if="isAdmin" class="member-actions">
              <el-button size="small" type="primary" plain @click="handleEditMember(m)">
                {{ t('common.edit') }}
              </el-button>
              <el-button size="small" type="danger" plain @click="handleRemoveMember(m)">
                {{ t('rbac.removeMember') }}
              </el-button>
            </div>
          </div>

          <div class="member-bindings">
            <div v-for="b in m.bindings" :key="b.bindingId" class="binding-row">
              <el-tag :type="roleTagType(b.roleName)" size="small">{{ b.roleDisplayName }}</el-tag>
              <el-tag v-if="b.scopeType === 'cluster'" size="small" type="info">
                {{ t('rbac.clusterScope') }}
              </el-tag>
              <el-tag v-else type="warning" size="small">{{ b.namespace }}</el-tag>
              <el-button
                v-if="isAdmin"
                size="small"
                type="danger"
                text
                @click="handleRemoveRole(m, b)"
              >
                {{ t('rbac.removeBinding') }}
              </el-button>
            </div>
            <span v-if="m.bindings.length === 0" class="no-role">{{ t('rbac.noRole') }}</span>
            <span v-if="m.isSuperAdmin" class="super-admin-badge">Super Admin</span>
          </div>
        </el-card>
      </div>
    </div>
  </el-drawer>

  <AddBindingDialog
    v-model:visible="addDialogVisible"
    :cluster-id="clusterId"
    :cluster-name="clusterName"
    :roles="roles"
    @success="handleAddSuccess"
  />
  <MemberEditDialog
    v-if="editingMember"
    v-model:visible="editDialogVisible"
    :cluster-id="clusterId"
    :member="editingMember"
    :roles="roles"
    @success="handleEditSuccess"
  />
</template>

<style scoped>
.drawer-header { display: flex; flex-direction: column; gap: 8px; }
.header-title { display: flex; align-items: center; font-size: 16px; font-weight: 600; }
.header-stats { display: flex; gap: 8px; flex-wrap: wrap; }
.drawer-body { display: flex; flex-direction: column; gap: 16px; }
.toolbar { display: flex; justify-content: space-between; align-items: center; flex-wrap: wrap; gap: 12px; }
.member-cards { display: flex; flex-direction: column; gap: 12px; min-height: 120px; }
.member-card { border: 1px solid var(--gk-color-border); }
.member-head { display: flex; justify-content: space-between; align-items: center; flex-wrap: wrap; gap: 8px; }
.member-info { display: flex; align-items: center; gap: 12px; }
.member-avatar { background: var(--gk-color-primary); color: #fff; flex-shrink: 0; }
.member-names { display: flex; flex-direction: column; }
.username { font-weight: 600; }
.display-name { font-size: 12px; color: var(--gk-color-text-secondary); }
.member-actions { display: flex; gap: 8px; }
.member-bindings { margin-top: 12px; display: flex; flex-direction: column; gap: 6px; }
.binding-row { display: flex; align-items: center; gap: 8px; flex-wrap: wrap; }
.no-role { color: var(--gk-color-text-disabled); font-size: 12px; }
.super-admin-badge {
  background: linear-gradient(135deg, #f56c6c 0%, #e6a23c 100%);
  color: white; padding: 2px 8px; border-radius: 4px;
  font-size: 12px; font-weight: 600; align-self: flex-start;
}
</style>
```

### 4.3 新增 `frontend/src/views/cluster/MemberEditDialog.vue`

**新建文件**，内容如下。编辑现有成员：每条授权行内换角色（走 batch API）+ 移除单条组授权：

```vue
<script setup lang="ts">
import { ref, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import { updateBindingBatch, deleteBindingBatch } from '@/api/rbac'

interface MemberBinding {
  bindingId: number
  roleId: number
  roleName: string
  roleDisplayName: string
  scopeType: 'cluster' | 'namespace'
  namespace: string
}
interface Member {
  userId: number
  username: string
  displayName: string
  isSuperAdmin: boolean
  bindings: MemberBinding[]
}

const props = defineProps<{
  visible: boolean
  clusterId: number
  member: Member
  roles: any[]
}>()

const emit = defineEmits<{
  'update:visible': [val: boolean]
  'success': []
}>()

const { t } = useI18n()
const saving = ref(false)

const dialogVisible = computed({
  get: () => props.visible,
  set: (val: boolean) => emit('update:visible', val),
})

// 每条授权可选的角色：与其 scopeType 一致的角色
function availableRoles(scopeType: string) {
  return props.roles.filter(r => r.scopeType === scopeType)
}

async function handleChangeRole(b: MemberBinding, newRoleId: number) {
  if (newRoleId === b.roleId) return
  saving.value = true
  try {
    await updateBindingBatch({
      userId: props.member.userId,
      clusterId: props.clusterId,
      fromRoleId: b.roleId,
      toRoleId: newRoleId,
    })
    ElMessage.success(t('rbac.updateSuccess'))
    emit('success')
  } catch (e: any) {
    ElMessage.error(e?.message || t('common.failed'))
  } finally {
    saving.value = false
  }
}

async function handleRemoveBinding(b: MemberBinding) {
  try {
    await ElMessageBox.confirm(
      t('rbac.removeRoleConfirm', {
        name: props.member.username,
        role: b.scopeType === 'cluster' ? b.roleDisplayName : `${b.roleDisplayName} (${b.namespace})`,
      }),
      t('common.confirm'),
      { type: 'warning' }
    )
  } catch { return }
  saving.value = true
  try {
    await deleteBindingBatch({
      userId: props.member.userId,
      clusterId: props.clusterId,
      roleId: b.roleId,
    })
    ElMessage.success(t('rbac.removeRoleSuccess'))
    emit('success')
  } catch (e: any) {
    ElMessage.error(e?.message || t('common.deleteFailed'))
  } finally {
    saving.value = false
  }
}

</script>

<template>
  <el-dialog v-model="dialogVisible" :title="t('rbac.editMember', { name: member.username })" width="640px">
    <div v-loading="saving" class="binding-list">
      <div v-for="b in member.bindings" :key="b.bindingId" class="binding-row">
        <div class="scope-cell">
          <el-tag v-if="b.scopeType === 'cluster'" size="small" type="info">{{ t('rbac.clusterScope') }}</el-tag>
          <el-tag v-else size="small" type="warning">{{ b.namespace }}</el-tag>
        </div>
        <el-select
          :model-value="b.roleId"
          size="default"
          style="width: 220px;"
          :disabled="saving"
          @change="(v: any) => handleChangeRole(b, Number(v))"
        >
          <el-option
            v-for="r in availableRoles(b.scopeType)"
            :key="r.id"
            :label="r.displayName"
            :value="r.id"
          />
        </el-select>
        <el-button size="small" type="danger" plain @click="handleRemoveBinding(b)">
          {{ t('rbac.removeBinding') }}
        </el-button>
      </div>
      <el-empty v-if="member.bindings.length === 0" :description="t('rbac.noRole')" />
    </div>
    <el-alert type="info" :closable="false" style="margin-top: 12px;">
      {{ t('rbac.scopeChangeWarning') }}
    </el-alert>
    <template #footer>
      <el-button @click="dialogVisible = false">{{ t('common.close') }}</el-button>
    </template>
  </el-dialog>
</template>

<style scoped>
.binding-list { display: flex; flex-direction: column; gap: 12px; }
.binding-row { display: flex; align-items: center; gap: 12px; flex-wrap: wrap; }
.scope-cell { width: 140px; }
</style>
```

注意：`common.close` 键若在 i18n 中不存在，必须在 4.5 节一并添加。

### 4.4 重写 `frontend/src/views/cluster/AddBindingDialog.vue`

**整个文件替换**。三段式：用户 → 作用域（集群级 / 指定命名空间多选）→ 角色（按作用域过滤）。多选 ns 时**一次**调用 `createBinding` 传 `namespaces` 数组：

```vue
<script setup lang="ts">
import { ref, computed, watch, nextTick } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import { createBinding, getNamespaceList, searchUsers } from '@/api/rbac'

const { t } = useI18n()

const props = defineProps<{
  visible: boolean
  clusterId: number
  clusterName: string
  binding?: any | null   // 兼容旧调用（不再使用编辑模式，编辑走 MemberEditDialog）
  roles: any[]
}>()

const emit = defineEmits<{
  (e: 'update:visible', val: boolean): void
  (e: 'success'): void
}>()

const dialogVisible = computed({
  get: () => props.visible,
  set: (val: boolean) => emit('update:visible', val),
})

const formRef = ref<FormInstance>()
const saving = ref(false)

// scope: 'cluster' | 'namespace'
const form = ref({
  userId: '' as string | number,
  scope: 'namespace' as 'cluster' | 'namespace',
  roleId: '' as string | number,
  namespaces: [] as string[],
})

const userOptions = ref<any[]>([])
const userLoading = ref(false)

const nsList = ref<string[]>([])
const nsLoading = ref(false)
const nsLoadError = ref(false)

const rules = computed<FormRules>(() => ({
  userId: [{ required: true, message: t('rbac.selectUser'), trigger: 'change' }],
  roleId: [{ required: true, message: t('rbac.selectRole'), trigger: 'change' }],
  namespaces: [{
    validator: (_rule: any, _value: any, callback: (err?: Error) => void) => {
      if (form.value.scope === 'namespace' && form.value.namespaces.length === 0) {
        callback(new Error(t('rbac.selectNamespace')))
      } else {
        callback()
      }
    },
    trigger: 'change',
  }],
}))

// 角色按作用域过滤：集群级 -> scopeType=cluster；空间级 -> scopeType=namespace
const filteredRoles = computed(() => {
  return props.roles.filter(r => r.scopeType === form.value.scope)
})

const selectedRole = computed(() => {
  return filteredRoles.value.find(r => r.id === form.value.roleId)
})

// 作用域切换时重置角色选择
watch(() => form.value.scope, () => {
  form.value.roleId = ''
  nextTick(() => formRef.value?.clearValidate('roleId'))
})

async function loadAllUsers() {
  if (userOptions.value.length > 0) return
  userLoading.value = true
  try {
    const res: any = await searchUsers({ page: 1, size: 100 })
    userOptions.value = res?.data?.items || []
  } catch {
    userOptions.value = []
  } finally {
    userLoading.value = false
  }
}

watch(() => props.visible, (val) => {
  if (val) {
    form.value.userId = ''
    form.value.scope = 'namespace'
    form.value.roleId = ''
    form.value.namespaces = []
    nextTick(() => formRef.value?.clearValidate())
    loadAllUsers()
  }
})

watch(() => props.visible, async (val) => {
  if (!val || !props.clusterName) return
  nsLoading.value = true
  nsLoadError.value = false
  try {
    const res: any = await getNamespaceList(props.clusterName)
    const nsData = res?.data ?? res
    const nsArr: any[] = Array.isArray(nsData) ? nsData : (nsData?.items || [])
    nsList.value = nsArr.map((ns: any) => ns.metadata?.name || ns.name || ns)
  } catch {
    nsLoadError.value = true
  } finally {
    nsLoading.value = false
  }
})

async function handleSubmit() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return

  saving.value = true
  try {
    await createBinding({
      userId: Number(form.value.userId),
      roleId: Number(form.value.roleId),
      clusterId: props.clusterId,
      // 集群级不传 namespaces（后端归一化为 [""]）；空间级传数组走批量插入
      ...(form.value.scope === 'namespace' ? { namespaces: form.value.namespaces } : {}),
    })
    ElMessage.success(t('rbac.createSuccess'))
    emit('success')
    dialogVisible.value = false
  } catch (e: any) {
    ElMessage.error(e?.message || t('common.failed'))
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <el-dialog v-model="dialogVisible" :title="t('rbac.addUser')" width="560px" :close-on-click-modal="false">
    <el-form ref="formRef" :model="form" :rules="rules" label-width="80px" label-position="right">
      <el-form-item :label="t('rbac.username')" prop="userId">
        <el-select
          v-model="form.userId"
          filterable
          :loading="userLoading"
          :placeholder="t('rbac.selectUser')"
          style="width: 100%;"
        >
          <el-option
            v-for="u in userOptions"
            :key="u.id"
            :label="`${u.username} (${u.display_name || '-'})`"
            :value="Number(u.id)"
          />
        </el-select>
      </el-form-item>

      <el-form-item :label="t('rbac.scope')">
        <el-radio-group v-model="form.scope">
          <el-radio value="namespace">{{ t('rbac.namespaceScope') }}</el-radio>
          <el-radio value="cluster">{{ t('rbac.clusterScope') }}</el-radio>
        </el-radio-group>
      </el-form-item>

      <el-form-item v-if="form.scope === 'namespace'" :label="t('rbac.namespace')" prop="namespaces">
        <el-select
          v-if="!nsLoadError"
          v-model="form.namespaces"
          filterable
          multiple
          collapse-tags
          collapse-tags-tooltip
          :loading="nsLoading"
          :placeholder="t('rbac.selectNamespace')"
          style="width: 100%;"
        >
          <el-option v-for="ns in nsList" :key="ns" :label="ns" :value="ns" />
        </el-select>
        <template v-else>
          <el-alert type="warning" :closable="false">
            {{ t('rbac.clusterOfflineNsHint') }}
          </el-alert>
        </template>
      </el-form-item>

      <el-form-item :label="t('rbac.role')" prop="roleId">
        <el-select v-model="form.roleId" :placeholder="t('rbac.selectRole')" style="width: 100%;">
          <el-option
            v-for="r in filteredRoles"
            :key="r.id"
            :label="`${r.displayName} (${r.name})`"
            :value="r.id"
          />
        </el-select>
      </el-form-item>

      <el-alert v-if="selectedRole" type="info" :closable="false" style="margin-top: 4px;">
        {{ selectedRole.displayName }}
      </el-alert>
    </el-form>

    <template #footer>
      <el-button @click="dialogVisible = false">{{ t('common.cancel') }}</el-button>
      <el-button type="primary" :loading="saving" @click="handleSubmit">
        {{ t('common.confirm') }}
      </el-button>
    </template>
  </el-dialog>
</template>
```

**删除项：** 原文件的 `isEditMode`、`binding` 编辑分支、`handleSelectAllNs`、`__all__` 选项逻辑全部移除（编辑功能已由 MemberEditDialog 承接）。`binding` prop 保留声明以兼容 ClusterList 旧调用残留，但不再读取（若 vue-tsc 报 unused prop 不会，props 声明不算 unused）。

**后端 CreateBinding 兼容性确认（已核对 `handler.go:144-227`）：** 集群级角色 + 空 `namespaces` + 空 `namespace` 时，后端 `nsList = [""]` 归一化逻辑已存在（`handler.go:187-189`），无需后端改动。

### 4.5 i18n 增量

**`frontend/src/locales/zh-CN.ts`** 的 `rbac: {` 对象内追加（放在 `members: '成员',` 之后）：

```ts
    editMember: '编辑成员: {name}',
    removeMember: '移出集群',
    removeMemberConfirm: '确定将 {name} 移出该集群？其在该集群下的全部授权将被删除。',
    removeMemberSuccess: '已移出集群',
    removeBinding: '移除',
    removeRoleConfirm: '确定移除 {name} 的「{role}」授权？',
    removeRoleSuccess: '已移除授权',
    close: '关闭',
```

**`frontend/src/locales/en.ts`** 的 `rbac: {` 对象内同样位置追加：

```ts
    editMember: 'Edit Member: {name}',
    removeMember: 'Remove from Cluster',
    removeMemberConfirm: 'Remove {name} from this cluster? All their permissions in this cluster will be deleted.',
    removeMemberSuccess: 'Removed from cluster',
    removeBinding: 'Remove',
    removeRoleConfirm: 'Remove the "{role}" permission of {name}?',
    removeRoleSuccess: 'Permission removed',
    close: 'Close',
```

同时检查 `common.close` 是否存在于两个文件的 `common:` 段；不存在则在 zh-CN `common: {` 内加 `close: '关闭',`、en 加 `close: 'Close',`。

### 4.6 P1 前端验收

```bash
cd frontend && npm run build
```

手动验收（`npm run dev` + 后端运行）：

1. 集群列表 -> 点"成员" -> 抽屉中每个用户一张卡片，卡片内列出其全部授权（角色 tag + ns tag）。
2. 管理员可见"添加成员/编辑/移出集群/移除"按钮；非管理员不可见。
3. 添加成员：选空间级 + 多选 3 个 ns + 选 ns-editor -> 提交后**网络面板只有一次** `POST /v1/rbac/bindings`，body 含 `namespaces: [...]`。
4. 添加成员：切换到集群级 -> ns 选择框消失 -> 角色下拉只剩 cluster-admin/editor/viewer。
5. 编辑成员：改某条授权的角色 -> 一次 `PUT /v1/rbac/bindings/batch`；移除 -> 一次 `DELETE /v1/rbac/bindings/batch`。
6. 移出集群 -> 一次 `DELETE /v1/rbac/bindings/user/:userId?clusterId=N`。
7. 集群列表 memberCount 显示真实数字。

---

## 5. P2-lite 后端实施（自定义角色 + 诊断，无分组）

本阶段**不新建任何表**。唯一存储变更是 `role` 表加 `description` 列（AutoMigrate 直接处理，无手工 DDL）。绑定表 `permission_binding` **保持三列唯一索引不变**，无迁移。

### 5.1 修改 `backend/internal/rbac/model/role.go`

增量修改（保留全部现有代码，只改两处）：

**改动 1：Role 结构体加 Description 字段。** 锚点：

```go
	IsSystem      bool              `gorm:"not null;default:true;comment:系统预置不可删除" json:"isSystem"`
	Permissions   string            `gorm:"type:text;not null;comment:JSON权限定义" json:"permissions"`
```

替换为：

```go
	IsSystem      bool              `gorm:"not null;default:true;comment:系统预置不可删除" json:"isSystem"`
	Permissions   string            `gorm:"type:text;not null;comment:JSON权限定义" json:"permissions"`
	Description   string            `gorm:"type:varchar(255);default:'';comment:角色描述" json:"description"`
```

**改动 2：文件末尾追加资源字典。**

```go
// ResourceVerbDict 资源组与动词字典（角色矩阵编辑器的唯一事实源）。
// key = resourceGroup（与 RequirePermission 的 resolveResourceGroup 输出一致）；
// value = 该资源组允许配置的动词列表。
// clusterOnly = true 的资源组仅集群级角色可配置。
var ResourceVerbDict = []ResourceGroupDef{
	{Group: "workload", Verbs: []string{"read", "create", "update", "delete", "terminal"}, ClusterOnly: false},
	{Group: "network", Verbs: []string{"read", "create", "update", "delete"}, ClusterOnly: false},
	{Group: "storage", Verbs: []string{"read", "create", "update", "delete"}, ClusterOnly: false},
	{Group: "config", Verbs: []string{"read", "create", "update", "delete"}, ClusterOnly: false},
	{Group: "event", Verbs: []string{"read"}, ClusterOnly: false},
	{Group: "crd", Verbs: []string{"read", "create", "update", "delete"}, ClusterOnly: false},
	{Group: "terminal", Verbs: []string{"terminal"}, ClusterOnly: false},
	{Group: "node", Verbs: []string{"read", "cordon", "taint", "drain", "delete"}, ClusterOnly: true},
	{Group: "namespace", Verbs: []string{"read", "create", "update", "delete"}, ClusterOnly: true},
	{Group: "audit", Verbs: []string{"read"}, ClusterOnly: true},
	{Group: "cluster_mgmt", Verbs: []string{"read"}, ClusterOnly: true},
}

// ResourceGroupDef 资源组定义
type ResourceGroupDef struct {
	Group       string   `json:"group"`
	Verbs       []string `json:"verbs"`
	ClusterOnly bool     `json:"clusterOnly"`
}
```

**迁移说明：** 在 `cmd/migrate.go` 的 AutoMigrate 参数中 `role` 模型已存在（若无则补上）。`Description` 列由 AutoMigrate 自动新增，默认 `''`，预置角色描述为空。**无需手工 DDL，无需索引重建。**

### 5.2 `backend/pkg/auth/permcache.go`：确认无需改动

现有实现（`CachedPermissions{IsSuperAdmin, Bindings, ExpireAt}`，TTL 5 分钟）已满足本方案。`InvalidateAllPermissions()` 将在 5.3 节 `UpdateRole` 中获得调用方（现状问题 #6 就此解决）。**本步骤无代码改动**，仅作为依赖确认。

### 5.3 新增角色 CRUD（`backend/internal/rbac/handler.go` 追加）

文件末尾追加：

```go
// --- P2 角色管理 ---

type CreateRoleParams struct {
	Name        string              `json:"name" binding:"required,max=50" label:"角色标识"`
	DisplayName string              `json:"displayName" binding:"required,max=100" label:"角色名称"`
	ScopeType   string              `json:"scopeType" binding:"required,oneof=cluster namespace" label:"作用域类型"`
	Description string              `json:"description" binding:"max=255" label:"描述"`
	Permissions map[string][]string `json:"permissions" binding:"required" label:"权限定义"`
}

type UpdateRoleParams struct {
	DisplayName string              `json:"displayName" binding:"required,max=100" label:"角色名称"`
	Description string              `json:"description" binding:"max=255" label:"描述"`
	Permissions map[string][]string `json:"permissions" binding:"required" label:"权限定义"`
}

// validateRolePermissions 校验权限矩阵合法性：资源组与动词必须在 ResourceVerbDict 内。
// clusterOnly 资源组只能出现在 cluster 作用域角色中。
func validateRolePermissions(scopeType string, perms map[string][]string) string {
	dict := model.ResourceVerbDict
	groupDef := make(map[string]model.ResourceGroupDef, len(dict))
	for _, d := range dict {
		groupDef[d.Group] = d
	}
	for group, verbs := range perms {
		d, ok := groupDef[group]
		if !ok {
			return fmt.Sprintf("未知的资源组 %q", group)
		}
		if d.ClusterOnly && scopeType != "cluster" {
			return fmt.Sprintf("资源组 %q 仅集群级角色可配置", group)
		}
		for _, v := range verbs {
			found := false
			for _, allowed := range d.Verbs {
				if v == allowed {
					found = true
					break
				}
			}
			if !found {
				return fmt.Sprintf("资源组 %q 不支持动词 %q", group, v)
			}
		}
	}
	return ""
}

// CreateRole 创建自定义角色
func (h *rbacHandler) CreateRole(c *gin.Context) {
	var p CreateRoleParams
	if err := c.ShouldBindJSON(&p); err != nil {
		response.Fail(c, "参数校验失败")
		return
	}

	if msg := validateRolePermissions(p.ScopeType, p.Permissions); msg != "" {
		response.Fail(c, msg)
		return
	}

	// 角色名不能与预置角色重名（预置名占用 name 唯一索引，DB 兜底）
	var cnt int64
	database.DB.Model(&model.Role{}).Where("name = ?", p.Name).Count(&cnt)
	if cnt > 0 {
		response.Fail(c, "角色标识已存在")
		return
	}

	permsJSON, err := json.Marshal(p.Permissions)
	if err != nil {
		response.Fail(c, "权限定义序列化失败")
		return
	}

	role := model.Role{
		Name: p.Name, DisplayName: p.DisplayName,
		ScopeType: p.ScopeType, Description: p.Description,
		IsSystem: false, Permissions: string(permsJSON),
	}
	if err := database.DB.Create(&role).Error; err != nil {
		if isDuplicateErr(err) {
			response.Fail(c, "角色标识已存在")
			return
		}
		logger.Error(fmt.Sprintf("创建角色失败: %v", err))
		response.FailWithStatus(c, http.StatusInternalServerError, "创建角色失败")
		return
	}

	response.Success(c, "创建成功", role)
}

// UpdateRole 更新自定义角色（预置角色只读）
func (h *rbacHandler) UpdateRole(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	var role model.Role
	if err := database.DB.First(&role, id).Error; err != nil {
		response.FailWithStatus(c, http.StatusNotFound, "角色不存在")
		return
	}
	if role.IsSystem {
		response.Fail(c, "系统预置角色不可修改")
		return
	}

	var p UpdateRoleParams
	if err := c.ShouldBindJSON(&p); err != nil {
		response.Fail(c, "参数校验失败")
		return
	}

	if msg := validateRolePermissions(role.ScopeType, p.Permissions); msg != "" {
		response.Fail(c, msg)
		return
	}

	permsJSON, err := json.Marshal(p.Permissions)
	if err != nil {
		response.Fail(c, "权限定义序列化失败")
		return
	}

	if err := database.DB.Model(&role).Updates(map[string]any{
		"DisplayName": p.DisplayName,
		"Description": p.Description,
		"Permissions": string(permsJSON),
	}).Error; err != nil {
		logger.Error(fmt.Sprintf("更新角色失败: %v", err))
		response.FailWithStatus(c, http.StatusInternalServerError, "更新角色失败")
		return
	}

	// 角色权限变更：失效全部用户权限缓存
	auth.InvalidateAllPermissions()

	database.DB.First(&role, role.ID)
	response.Success(c, "更新成功", role)
}

// DeleteRole 删除自定义角色（预置角色只读；存在绑定时拒绝）
func (h *rbacHandler) DeleteRole(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	var role model.Role
	if err := database.DB.First(&role, id).Error; err != nil {
		response.FailWithStatus(c, http.StatusNotFound, "角色不存在")
		return
	}
	if role.IsSystem {
		response.Fail(c, "系统预置角色不可删除")
		return
	}

	var cnt int64
	database.DB.Model(&model.PermissionBinding{}).Where("role_id = ?", role.ID).Count(&cnt)
	if cnt > 0 {
		response.Fail(c, fmt.Sprintf("该角色仍有 %d 条绑定引用，请先移除相关授权", cnt))
		return
	}

	if err := database.DB.Delete(&role).Error; err != nil {
		logger.Error(fmt.Sprintf("删除角色失败: %v", err))
		response.FailWithStatus(c, http.StatusInternalServerError, "删除角色失败")
		return
	}

	response.Success(c, "删除成功", nil)
}

// ListResourceDict 返回资源组×动词字典（矩阵编辑器数据源）
func (h *rbacHandler) ListResourceDict(c *gin.Context) {
	response.Success(c, "获取资源字典成功", model.ResourceVerbDict)
}
```

**改造现有 ListRoles（handler.go，容易遗漏）：** 现有 `ListRoles` 内部的 `roleResp` 结构体（锚点 `type roleResp struct`）不含 `Description`，RoleManagement 表格描述列会恒为空。在 `roleResp` 中补一行：

```go
		ScopeType   string            `json:"scopeType"`
		IsSystem    bool              `json:"isSystem"`
		Description string            `json:"description"`
		Permissions map[string][]string `json:"permissions"`
```

（构造循环中的 `items = append(items, roleResp{...})` 同步加 `Description: r.Description,`。）

**import 增量（handler.go）：** 头部 import 块需新增 `"encoding/json"`。锚点：

```go
import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
```

替换为：

```go
import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
```

**isDuplicateErr 辅助函数确认：** 若 handler.go（或 rbac 包内）尚无该函数，则在文件末尾追加：

```go
// isDuplicateErr 判断是否为唯一索引冲突（MySQL 1062）
func isDuplicateErr(err error) bool {
	return err != nil && strings.Contains(err.Error(), "Duplicate entry")
}
```

### 5.4 新增 can-i 诊断（`backend/internal/rbac/handler.go` 追加）

```go
// --- P2 权限诊断 ---

type CanIParams struct {
	UserID        uint   `json:"userId" binding:"required" label:"用户ID"`
	ClusterID     uint   `json:"clusterId" binding:"required" label:"集群ID"`
	Namespace     string `json:"namespace"`
	ResourceGroup string `json:"resourceGroup" binding:"required" label:"资源组"`
	Verb          string `json:"verb" binding:"required" label:"动词"`
}

// CanI 诊断：模拟中间件鉴权逻辑，返回允许与否 + 命中的绑定。
// 与 middleware.RequirePermission 的匹配语义保持一致（超管放行、两层作用域匹配）。
func (h *rbacHandler) CanI(c *gin.Context) {
	var p CanIParams
	if err := c.ShouldBindJSON(&p); err != nil {
		response.Fail(c, "参数校验失败")
		return
	}

	// 资源组与动词合法性（用字典校验，避免任意字符串）
	validGroup := false
	var validVerbs []string
	clusterOnly := false
	for _, d := range model.ResourceVerbDict {
		if d.Group == p.ResourceGroup {
			validGroup = true
			validVerbs = d.Verbs
			clusterOnly = d.ClusterOnly
			break
		}
	}
	if !validGroup {
		response.Fail(c, "未知的资源组")
		return
	}
	validVerb := false
	for _, v := range validVerbs {
		if v == p.Verb {
			validVerb = true
			break
		}
	}
	if !validVerb {
		response.Fail(c, "该资源组不支持此动词")
		return
	}

	// 超管
	var user authmodel.User
	if err := database.DB.First(&user, p.UserID).Error; err != nil {
		response.Fail(c, "用户不存在")
		return
	}
	if auth.IsAdmin(user.Username) {
		response.Success(c, "诊断完成", gin.H{"allow": true, "reason": "super_admin_config", "matchedBinding": nil})
		return
	}

	cached := auth.GetUserPermissions(p.UserID)
	if cached.IsSuperAdmin {
		response.Success(c, "诊断完成", gin.H{"allow": true, "reason": "super_admin_db", "matchedBinding": nil})
		return
	}

	// 两层作用域匹配（与 middleware.RequirePermission 现有逻辑一致，无分组分支）
	type matchInfo struct {
		BindingID uint   `json:"bindingId"`
		ScopeType string `json:"scopeType"` // cluster|namespace
		Namespace string `json:"namespace"`
		RoleName  string `json:"roleName"`
	}

	for _, b := range cached.Bindings {
		if b.ClusterID != p.ClusterID {
			continue
		}
		scope := ""
		matched := false
		switch {
		case b.Namespace == "":
			// 集群级绑定：覆盖全集群
			scope, matched = "cluster", true
		case b.Namespace == p.Namespace:
			// 空间级绑定：精确 ns 匹配
			scope, matched = "namespace", true
		}
		if !matched {
			continue
		}
		perms := b.Role.GetPermissionsMap()
		if verbs, ok := perms[p.ResourceGroup]; ok {
			for _, v := range verbs {
				if v == p.Verb {
					response.Success(c, "诊断完成", gin.H{
						"allow": true,
						"reason": "role_permission",
						"matchedBinding": matchInfo{
							BindingID: b.ID, ScopeType: scope,
							Namespace: b.Namespace,
							RoleName:  b.Role.Name,
						},
					})
					return
				}
			}
		}
	}

	_ = clusterOnly // 预留：诊断提示用
	response.Success(c, "诊断完成", gin.H{"allow": false, "reason": "no_matching_binding", "matchedBinding": nil})
}
```

**匹配语义对照（必须与 middleware 一致）：** 空间级绑定 `b.Namespace == p.Namespace` 在 `p.Namespace == ""` 时不匹配——因为该分支要求 `b.Namespace != ""` 且等于请求 ns，请求 ns 为空串时永远不等于非空的绑定 ns。与 1.4 节语义表第 4 行一致。

### 5.5 路由注册（P2-lite，`backend/internal/router/rbac.go`）

在 3.3 节最终文件基础上**替换整个文件**：

```go
package router

import (
	"github.com/gin-gonic/gin"
	"gkube/internal/rbac"
	"gkube/pkg/middleware"
)

// registerRbacRoutes 注册 RBAC 权限管理路由
func registerRbacRoutes(rg *gin.RouterGroup) {
	// 所有已认证用户可访问（查自己权限、角色列表、集群成员、资源字典）
	rg.GET("rbac/roles", rbac.RbacHandler.ListRoles)
	rg.GET("rbac/resources", rbac.RbacHandler.ListResourceDict)
	rg.GET("rbac/my-permissions", rbac.RbacHandler.MyPermissions)
	rg.GET("rbac/cluster-members", rbac.RbacHandler.ClusterMembers)

	// 仅管理员可访问（写操作 + 管理查询 + 诊断）
	// AuditLog 放在 RequireAdmin 之前，确保 403 失败尝试也被审计
	admin := rg.Group("rbac", middleware.AuditLog(), middleware.RequireAdmin())
	{
		admin.GET("bindings", rbac.RbacHandler.ListBindings)
		admin.POST("bindings", rbac.RbacHandler.CreateBinding)
		admin.PUT("bindings/batch", rbac.RbacHandler.UpdateBindingBatch)
		admin.DELETE("bindings/batch", rbac.RbacHandler.DeleteBindingBatch)
		admin.DELETE("bindings/user/:userId", rbac.RbacHandler.RemoveClusterMember)
		admin.PUT("bindings/:id", rbac.RbacHandler.UpdateBinding)
		admin.DELETE("bindings/:id", rbac.RbacHandler.DeleteBinding)
		admin.POST("can-i", rbac.RbacHandler.CanI)

		admin.POST("roles", rbac.RbacHandler.CreateRole)
		admin.PUT("roles/:id", rbac.RbacHandler.UpdateRole)
		admin.DELETE("roles/:id", rbac.RbacHandler.DeleteRole)
	}
}
```

**gin 冲突注意：** `GET rbac/roles`（公开）与 `POST/PUT/DELETE rbac/roles[...]`（admin 组）分属不同 method+path，无冲突。

### 5.6 审计中间件兼容（确认，无需改动）

已核对 `pkg/middleware/audit.go`：`parseK8sPath` 只解析 `/v1/k8s/` 与 `/v1/rbac/` 前缀。新路由 `roles/:id`、`can-i` 落在 `/v1/rbac/` 下会被记录（GET 自动跳过；POST/PUT/DELETE 记录 action 为路径段如 "roles"，数字段不匹配 strconv.Atoi 时保留原文）。**无需改动**，但需在验收时确认 audit log 无 panic。

## 6. P2-lite 前端实施

### 6.1 `frontend/src/api/rbac.ts`：追加 P2 API

在文件末尾追加：

```ts
// --- P2 ---

// 获取资源组×动词字典（角色矩阵编辑器数据源）
export const getResourceDict = () => request.get('/rbac/resources')

// 创建自定义角色
export const createRole = (data: {
  name: string
  displayName: string
  scopeType: 'cluster' | 'namespace'
  description?: string
  permissions: Record<string, string[]>
}) => request.post('/rbac/roles', data)

// 更新自定义角色（预置角色不可改）
export const updateRole = (id: number, data: {
  displayName: string
  description?: string
  permissions: Record<string, string[]>
}) => request.put(`/rbac/roles/${id}`, data)

// 删除自定义角色
export const deleteRole = (id: number) => request.delete(`/rbac/roles/${id}`)

// 权限诊断
export const canI = (data: {
  userId: number
  clusterId: number
  namespace?: string
  resourceGroup: string
  verb: string
}) => request.post('/rbac/can-i', data)
```

### 6.2 新增 `frontend/src/views/rbac/RoleManagement.vue`

**新建文件**（系统管理区角色列表 + 矩阵编辑对话框）：

```vue
<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { getRoles, createRole, updateRole, deleteRole, getResourceDict } from '@/api/rbac'
import ResourceListToolbar from '@/components/ResourceListToolbar.vue'

interface RoleItem {
  id: number
  name: string
  displayName: string
  scopeType: 'cluster' | 'namespace'
  isSystem: boolean
  description?: string
  permissions: Record<string, string[]>
}
interface ResourceGroupDef {
  group: string
  verbs: string[]
  clusterOnly: boolean
}

const { t } = useI18n()
const loading = ref(false)
const roles = ref<RoleItem[]>([])
const dict = ref<ResourceGroupDef[]>([])
const searchName = ref('')

// 矩阵编辑对话框
const editVisible = ref(false)
const editRole = ref<RoleItem | null>(null)
const editForm = ref({
  name: '',
  displayName: '',
  scopeType: 'namespace' as 'cluster' | 'namespace',
  description: '',
  permissions: {} as Record<string, string[]>,
})
const saving = ref(false)

const filteredRoles = computed(() => {
  if (!searchName.value.trim()) return roles.value
  const q = searchName.value.trim().toLowerCase()
  return roles.value.filter(r =>
    r.name.toLowerCase().includes(q) || r.displayName.toLowerCase().includes(q)
  )
})

// 当前编辑作用域下可配置的资源组（clusterOnly 组仅 cluster 作用域）
const availableGroups = computed(() => {
  if (editForm.value.scopeType === 'cluster') return dict.value
  return dict.value.filter(d => !d.clusterOnly)
})

// 矩阵表头：字典中出现过的全部动词（去重）
const allVerbs = computed(() => {
  const s = new Set<string>()
  for (const d of dict.value) d.verbs.forEach(v => s.add(v))
  return [...s].sort()
})

async function fetchData() {
  loading.value = true
  try {
    const [rolesRes, dictRes] = await Promise.all([
      getRoles(),
      getResourceDict(),
    ])
    const rData: any = rolesRes?.data ?? rolesRes
    roles.value = (Array.isArray(rData) ? rData : rData?.items || []) as RoleItem[]
    const dData: any = dictRes?.data ?? dictRes
    dict.value = (Array.isArray(dData) ? dData : dData?.items || []) as ResourceGroupDef[]
  } catch (e: any) {
    ElMessage.error(e?.message || t('rbac.loadFailed'))
  } finally {
    loading.value = false
  }
}

function openCreate() {
  editRole.value = null
  editForm.value = { name: '', displayName: '', scopeType: 'namespace', description: '', permissions: {} }
  editVisible.value = true
}

function openEdit(role: RoleItem) {
  if (role.isSystem) {
    // 预置角色：复制为新角色起点
    editRole.value = null
    editForm.value = {
      name: role.name + '-copy',
      displayName: role.displayName + '(副本)',
      scopeType: role.scopeType,
      description: '',
      permissions: JSON.parse(JSON.stringify(role.permissions || {})),
    }
  } else {
    editRole.value = role
    editForm.value = {
      name: role.name,
      displayName: role.displayName,
      scopeType: role.scopeType,
      description: role.description || '',
      permissions: JSON.parse(JSON.stringify(role.permissions || {})),
    }
  }
  editVisible.value = true
}

async function handleDelete(role: RoleItem) {
  try {
    await ElMessageBox.confirm(
      t('rbac.deleteRoleConfirm', { name: role.displayName }),
      t('common.confirm'),
      { type: 'warning' }
    )
  } catch { return }
  try {
    await deleteRole(role.id)
    ElMessage.success(t('common.deleted'))
    fetchData()
  } catch (e: any) {
    ElMessage.error(e?.message || t('common.deleteFailed'))
  }
}

// 矩阵勾选联动
function toggleVerb(group: string, verb: string, checked: boolean) {
  const perms = { ...editForm.value.permissions }
  const cur = perms[group] || []
  if (checked) {
    if (!cur.includes(verb)) cur.push(verb)
  } else {
    const i = cur.indexOf(verb)
    if (i >= 0) cur.splice(i, 1)
  }
  if (cur.length === 0) delete perms[group]
  else perms[group] = cur
  editForm.value.permissions = perms
}

function isVerbOn(group: string, verb: string): boolean {
  return (editForm.value.permissions[group] || []).includes(verb)
}

// 切换作用域时清除 clusterOnly 组的配置
function handleScopeChange() {
  if (editForm.value.scopeType === 'namespace') {
    const clusterOnlyGroups = new Set(dict.value.filter(d => d.clusterOnly).map(d => d.group))
    const perms: Record<string, string[]> = {}
    for (const [g, verbs] of Object.entries(editForm.value.permissions)) {
      if (!clusterOnlyGroups.has(g)) perms[g] = verbs
    }
    editForm.value.permissions = perms
  }
}

async function handleSave() {
  if (!editForm.value.name || !editForm.value.displayName) {
    ElMessage.warning(t('rbac.roleFormRequired'))
    return
  }
  saving.value = true
  try {
    if (editRole.value) {
      await updateRole(editRole.value.id, {
        displayName: editForm.value.displayName,
        description: editForm.value.description,
        permissions: editForm.value.permissions,
      })
    } else {
      await createRole({
        name: editForm.value.name,
        displayName: editForm.value.displayName,
        scopeType: editForm.value.scopeType,
        description: editForm.value.description,
        permissions: editForm.value.permissions,
      })
    }
    ElMessage.success(t('rbac.updateSuccess'))
    editVisible.value = false
    fetchData()
  } catch (e: any) {
    ElMessage.error(e?.message || t('common.failed'))
  } finally {
    saving.value = false
  }
}

onMounted(fetchData)
</script>

<template>
  <div class="page-container">
    <ResourceListToolbar
      :search-value="searchName"
      :total-count="roles.length"
      :show-namespace="false"
      :search-placeholder="t('rbac.searchRoles')"
      @search-input="searchName = $event"
    >
      <template #actions>
        <el-button type="success" @click="openCreate">
          <el-icon><Plus /></el-icon> {{ t('rbac.createRole') }}
        </el-button>
      </template>
    </ResourceListToolbar>

    <el-card shadow="never" class="table-card">
      <el-table :data="filteredRoles" v-loading="loading" stripe>
        <el-table-column prop="displayName" label="角色" min-width="140" />
        <el-table-column prop="name" label="标识" min-width="140" />
        <el-table-column label="类型" width="120" align="center">
          <template #default="{ row }">
            <el-tag size="small" :type="row.scopeType === 'cluster' ? 'danger' : 'warning'">
              {{ row.scopeType === 'cluster' ? t('rbac.clusterScope') : t('rbac.namespaceScope') }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="来源" width="100" align="center">
          <template #default="{ row }">
            <el-tag size="small" :type="row.isSystem ? 'info' : 'success'">
              {{ row.isSystem ? t('rbac.presetRole') : t('rbac.customRole') }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="description" label="描述" min-width="200" show-overflow-tooltip>
          <template #default="{ row }">{{ row.description || '-' }}</template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right" align="center">
          <template #default="{ row }">
            <el-button size="small" type="primary" plain @click="openEdit(row)">
              {{ row.isSystem ? t('rbac.copyRole') : t('common.edit') }}
            </el-button>
            <el-button v-if="!row.isSystem" size="small" type="danger" plain @click="handleDelete(row)">
              {{ t('common.delete') }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 矩阵编辑对话框 -->
    <el-dialog v-model="editVisible" :title="editRole ? t('rbac.editRole') : t('rbac.createRole')" width="720px">
      <el-form label-width="80px">
        <el-form-item :label="t('rbac.roleName')">
          <el-input v-model="editForm.displayName" :disabled="!!editRole" style="width: 240px;" />
        </el-form-item>
        <el-form-item v-if="!editRole" :label="t('rbac.roleIdent')">
          <el-input v-model="editForm.name" style="width: 240px;" placeholder="如 dba-readonly" />
        </el-form-item>
        <el-form-item :label="t('rbac.scope')">
          <el-radio-group v-model="editForm.scopeType" :disabled="!!editRole" @change="handleScopeChange">
            <el-radio value="namespace">{{ t('rbac.namespaceScope') }}</el-radio>
            <el-radio value="cluster">{{ t('rbac.clusterScope') }}</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item :label="t('rbac.roleDescLabel')">
          <el-input v-model="editForm.description" type="textarea" :rows="2" />
        </el-form-item>
      </el-form>

      <el-divider content-position="left">{{ t('rbac.permissionMatrix') }}</el-divider>
      <el-table :data="availableGroups" size="small" border max-height="360">
        <el-table-column prop="group" label="资源组" width="140" />
        <el-table-column v-for="v in allVerbs" :key="v" :label="v" width="90" align="center">
          <template #default="{ row }">
            <el-checkbox
              v-if="row.verbs.includes(v)"
              :model-value="isVerbOn(row.group, v)"
              @change="(val: any) => toggleVerb(row.group, v, !!val)"
            />
            <span v-else class="na">-</span>
          </template>
        </el-table-column>
      </el-table>

      <el-alert type="warning" :closable="false" style="margin-top: 12px;">
        {{ t('rbac.roleCacheWarning') }}
      </el-alert>

      <template #footer>
        <el-button @click="editVisible = false">{{ t('common.cancel') }}</el-button>
        <el-button type="primary" :loading="saving" @click="handleSave">{{ t('common.save') }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.na { color: var(--gk-color-text-disabled); }
</style>
```

注意：现有 `ListRoles` 已把 `permissions` 解析为 map 返回（`roleResp.Permissions` 类型是 `map[string][]string`），前端拿到的**通常已是对象**。但防御性归一仍保留——若未来 ListRoles 改为直接返回 model（permissions 变回 JSON 字符串），组件不会静默坏掉。`openEdit` 的两处 `permissions: ...` 改为：

```ts
      permissions: parsePerms(role.permissions),
```

并在 setup 顶部（`fetchData` 之前）加：

```ts
// 类型归一：后端正常返回对象；若返回 JSON 字符串也能解析
function parsePerms(p: any): Record<string, string[]> {
  if (!p) return {}
  if (typeof p === 'string') {
    try { return JSON.parse(p) } catch { return {} }
  }
  return p
}
```

### 6.3 新增 `frontend/src/views/rbac/MyPermissions.vue`

**新建文件**（顶栏用户菜单进入）：

```vue
<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { getMyPermissions } from '@/api/rbac'
import { useAuthStore } from '@/stores/auth'

const { t } = useI18n()
const authStore = useAuthStore()
const loading = ref(false)
const isSuperAdmin = ref(false)
const bindings = ref<any[]>([])

interface ClusterEntry {
  clusterId: number
  items: { roleName: string; namespace: string }[]
}

const byCluster = ref<ClusterEntry[]>([])

async function fetchData() {
  loading.value = true
  try {
    const res: any = await getMyPermissions()
    const data = res?.data ?? res
    isSuperAdmin.value = !!data?.isSuperAdmin
    bindings.value = data?.bindings || []
    const map = new Map<number, ClusterEntry>()
    for (const b of bindings.value) {
      let entry = map.get(b.clusterId)
      if (!entry) {
        entry = { clusterId: b.clusterId, items: [] }
        map.set(b.clusterId, entry)
      }
      entry.items.push({
        roleName: b.roleName,
        namespace: b.namespace,
      })
    }
    byCluster.value = [...map.values()]
  } finally {
    loading.value = false
  }
}

onMounted(fetchData)
</script>

<template>
  <div class="page-container" v-loading="loading">
    <el-card shadow="never" class="table-card">
      <template #header>
        <span>{{ t('rbac.myPermissions') }} — {{ authStore.user?.username }}</span>
      </template>

      <el-alert v-if="isSuperAdmin" type="success" :closable="false" style="margin-bottom: 16px;">
        {{ t('rbac.superAdminBypass') }}
      </el-alert>

      <div v-for="c in byCluster" :key="c.clusterId" class="cluster-block">
        <div class="cluster-title">{{ t('rbac.clusterLabel') }} #{{ c.clusterId }}</div>
        <div v-for="(item, i) in c.items" :key="i" class="perm-row">
          <el-tag size="small">{{ item.roleName }}</el-tag>
          <el-tag v-if="item.namespace" size="small" type="warning">{{ item.namespace }}</el-tag>
          <el-tag v-else size="small" type="info">{{ t('rbac.clusterScope') }}</el-tag>
        </div>
      </div>
      <el-empty v-if="!isSuperAdmin && byCluster.length === 0" :description="t('rbac.noPermissions')" />
    </el-card>
  </div>
</template>

<style scoped>
.cluster-block { margin-bottom: 16px; }
.cluster-title { font-weight: 600; margin-bottom: 8px; }
.perm-row { display: flex; gap: 8px; margin-bottom: 6px; }
</style>
```

### 6.4 路由与导航注册

**`frontend/src/router/index.ts`** 在 users 路由（锚点）：

```ts
        {
          path: 'users',
          name: 'UserList',
          component: () => import('@/views/system/UserList.vue'),
          meta: { title: '用户管理', icon: 'User', requireAdmin: true },
        },
```

之后插入：

```ts
        {
          path: 'roles',
          name: 'RoleManagement',
          component: () => import('@/views/rbac/RoleManagement.vue'),
          meta: { title: '角色管理', icon: 'Avatar', requireAdmin: true },
        },
        {
          path: 'my-permissions',
          name: 'MyPermissions',
          component: () => import('@/views/rbac/MyPermissions.vue'),
          meta: { title: '我的权限', icon: 'Avatar' },
        },
```

**`frontend/src/components/Layout/Sidebar.vue`** 系统管理子菜单（锚点）：

```html
        <el-menu-item index="/audit" @click="navigateTo('/audit')">
          <el-icon><Document /></el-icon>
          <template #title>{{ t('sidebar.audit') }}</template>
        </el-menu-item>
```

之后插入：

```html
        <el-menu-item index="/roles" @click="navigateTo('/roles')">
          <el-icon><Avatar /></el-icon>
          <template #title>{{ t('sidebar.roles') }}</template>
        </el-menu-item>
```

注意 Sidebar.vue 需确认 import 了 `Avatar` 图标（`Key` 已被 secrets 菜单占用，不要复用）。若 import 列表没有 `Avatar` 则从 `@element-plus/icons-vue` 追加。

**`frontend/src/components/Layout/Header.vue`** 用户菜单（锚点）：

```html
            <el-dropdown-item divided command="logout">
```

之前插入：

```html
            <el-dropdown-item command="myPermissions">
              <el-icon><Avatar /></el-icon>
              {{ t('rbac.myPermissions') }}
            </el-dropdown-item>
```

`handleCommand` 函数锚点：

```ts
async function handleCommand(command: string) {
  if (command === 'logout') {
    await authStore.logout()
    router.push('/login')
  }
}
```

替换为：

```ts
async function handleCommand(command: string) {
  if (command === 'logout') {
    await authStore.logout()
    router.push('/login')
  } else if (command === 'myPermissions') {
    router.push('/my-permissions')
  }
}
```

Header.vue import 列表追加 `Avatar`（`@element-plus/icons-vue`）。

### 6.5 P2-lite i18n 增量

**zh-CN.ts** `rbac:` 段追加（P1 键之后）：

```ts
    searchRoles: '搜索角色',
    createRole: '创建角色',
    editRole: '编辑角色',
    copyRole: '复制',
    presetRole: '预置',
    customRole: '自定义',
    clusterLabel: '集群',
    roleIdent: '角色标识',
    roleDescLabel: '描述',
    permissionMatrix: '权限矩阵',
    roleCacheWarning: '保存后立即失效服务端权限缓存；当前已登录用户的会话界面可能需要刷新或重新登录后才能看到变化。',
    deleteRoleConfirm: '确定删除角色「{name}」？',
    roleFormRequired: '请填写角色名称与标识',
    myPermissions: '我的权限',
    superAdminBypass: '平台管理员：跳过所有权限检查。',
    noPermissions: '暂无任何集群权限',
```

**注意：`rbac.roleName` 已存在**（`zh-CN.ts` rbac 段 `roleName: '角色'`），**不要再添加**——同对象重复属性会导致 vue-tsc 编译失败。角色列表的"角色名称"表头直接复用现有键。

**zh-CN.ts** `sidebar:` 段追加：

```ts
    roles: '角色管理',
```

**en.ts** `rbac:` 段追加：

```ts
    searchRoles: 'Search roles',
    createRole: 'Create Role',
    editRole: 'Edit Role',
    copyRole: 'Copy',
    presetRole: 'Preset',
    customRole: 'Custom',
    clusterLabel: 'Cluster',
    roleIdent: 'Role Identifier',
    roleDescLabel: 'Description',
    permissionMatrix: 'Permission Matrix',
    roleCacheWarning: 'Server-side permission cache is invalidated immediately; logged-in users may need to refresh or re-login to see the change.',
    deleteRoleConfirm: 'Delete role "{name}"?',
    roleFormRequired: 'Role name and identifier are required',
    myPermissions: 'My Permissions',
    superAdminBypass: 'Platform admin: bypasses all permission checks.',
    noPermissions: 'No cluster permissions yet',
```

同样**不要添加 `roleName`**（en.ts rbac 段已存在 `roleName: 'Role'`）。

**en.ts** `sidebar:` 段追加：

```ts
    roles: 'Roles',
```

---

## 7. 实施顺序总表（执行核对单）

按顺序打勾执行，每项完成后立即跑对应验证命令：

| 步骤 | 内容 | 验证 |
|------|------|------|
| P1-1 | 3.1 handler.go 追加 3 个批量 handler | `go build ./...` |
| P1-2 | 3.2 替换 ClusterMembers | `go build ./...` |
| P1-3 | 3.3 替换 router/rbac.go，实测路由无 panic | `go build ./...` + 启动冒烟 |
| P1-4 | 3.4 cluster.go List 加 memberCount | `go build ./...` |
| P1-5 | 3.5 后端验收（curl 冒烟） | 3.5 节命令全过 |
| P1-6 | 4.1 api/rbac.ts 追加 | `npm run build` |
| P1-7 | 4.2 重写 ClusterMembersDialog.vue | `npm run build` |
| P1-8 | 4.3 新建 MemberEditDialog.vue | `npm run build` |
| P1-9 | 4.4 重写 AddBindingDialog.vue | `npm run build` |
| P1-10 | 4.5 i18n（zh + en + common.close 检查） | `npm run build` |
| P1-11 | 4.6 前端手动验收 7 项 | 浏览器逐项过 |
| P2-1 | 5.1 role.go 加 Description + ResourceVerbDict | `go build ./...` |
| P2-2 | 5.2 permcache.go 确认无改动 | `go build ./...` |
| P2-3 | 5.3 角色 CRUD + json import + isDuplicateErr 确认 | `go build ./...` |
| P2-4 | 5.4 CanI handler | `go build ./...` |
| P2-5 | 5.5 替换 router/rbac.go（P2-lite 版） | `go build` + 启动冒烟 |
| P2-6 | 5.6 审计兼容确认 | 启动后写操作看 audit 无 panic |
| P2-7 | 6.1 api/rbac.ts 追加 | `npm run build` |
| P2-8 | 6.2 新建 RoleManagement.vue（含 parsePerms 归一 + ListRoles 补 description 字段） | `npm run build` |
| P2-9 | 6.3 新建 MyPermissions.vue | `npm run build` |
| P2-10 | 6.4 路由 + Sidebar + Header | `npm run build` |
| P2-11 | 6.5 i18n 全量（zh + en） | `npm run build` |

---

## 8. P1 验收清单（全部必须通过）

后端：

- [ ] `go build ./...` 零错误
- [ ] `GET /v1/rbac/cluster-members?clusterId=N` 返回按用户分组的 items（含 bindings 数组）
- [ ] `PUT /v1/rbac/bindings/batch` 换角色后，绑定表对应用户同 fromRoleId 的行 role_id 全部变更
- [ ] `DELETE /v1/rbac/bindings/batch` 后对应行删除
- [ ] `DELETE /v1/rbac/bindings/user/:userId?clusterId=N` 后该用户该集群全部绑定删除
- [ ] `GET /v1/clusters` items 元素含 memberCount 且等于 COUNT(DISTINCT user_id)
- [ ] 三个批量端点均要求管理员（普通用户 403 + 审计记录）

前端：

- [ ] `npm run build` 零错误（vue-tsc 通过）
- [ ] 成员抽屉一行一用户，卡片内列全部授权
- [ ] 多选 ns 添加成员只发一次 POST，body 含 namespaces 数组
- [ ] 集群级授权可通过 UI 创建（scope 切 cluster）
- [ ] 非管理员看不到操作按钮
- [ ] 移出集群 / 按组移除 / 按组换角色均工作
- [ ] 集群列表 memberCount 真实

## 9. P2-lite 验收清单（全部必须通过）

后端：

- [ ] `go run main.go migrate` 幂等（连续两次无异常），`role` 表多出 `description` 列，**无其他表结构变化**（无 ns_group 等新表）
- [ ] 角色自定义 CRUD：创建含 clusterOnly 组的 namespace 角色被 400 拒绝；更新角色后其他用户权限缓存失效（5 分钟内行为变化）
- [ ] 预置角色 Update/Delete 均被拒绝；有绑定引用的自定义角色 Delete 被拒绝（提示 N 条引用）
- [ ] `GET /v1/rbac/roles` 返回项含 `description` 字段（roleResp 已补）
- [ ] can-i：超管返回 allow=true reason=super_admin_*；普通用户按 1.4 语义表返回（集群级绑定 ns='' 请求也放行；空间级绑定 ns='' 请求拒绝）
- [ ] 审计：roles / can-i 写操作产生审计记录，无 panic

前端：

- [ ] `npm run build` 零错误
- [ ] 角色管理页矩阵渲染自资源字典；勾选/取消保存后重新加载正确
- [ ] 预置角色只有"复制"，自定义角色有编辑/删除
- [ ] 顶栏"我的权限"页显示自己的全部绑定
- [ ] i18n 中英文切换无缺失键警告（浏览器 console）

---

## 10. 已知边界与不实施项

1. **不实施** deny 语义、角色层级继承、workspace 多级作用域。
2. **不实施** P3（平台权限同步为集群内原生 RBAC）。留待后续。
3. **不实施命名空间分组**（本版明确裁剪，理由见文档头部"与旧版方案的差异"）。多 ns 授权 = 多行绑定（P1 `namespaces[]` 批量参数）；"一组 ns 会动态增减"的场景暂不支持，需要管理员逐条调整绑定。恢复路径见第 11 节。
4. `ListBindings`（管理员原始绑定列表）在 P1/P2-lite 均不做 UI，仅保留 API 供排查。
5. memberCount 统计失败时静默显示 0（不阻塞集群列表）。
6. MyPermissions 中 clusterId 显示数字 ID 而非集群名——避免引入 cluster store 全量拉取；如需要后续在 my-permissions 响应中加 clusterName 字段（不在本次范围）。
7. can-i 前端 UI 不在本版范围（API 已交付，可用 curl / 后续补 UI）。

## 11. 回滚方案

- **P1**：纯展示层与新增路由回滚 = git revert 对应 commit；无数据风险。
- **P2-lite**：存储变更仅 `role.description` 一列（AutoMigrate 新增，可 `ALTER TABLE role DROP COLUMN description` 或保留无害）。代码回滚 = git revert。**无索引重建、无新表、无数据迁移，回滚零数据风险。**
- **将来补回命名空间分组**（若需要）：按旧版方案 5.2 的迁移路径执行——`permission_binding` 加 `ns_group_id` 列（默认 0）+ 唯一索引从三列重建为四列 + 新建 `ns_group` / `ns_group_item` 表 + `pkg/auth/nsgroupcache.go` + 中间件三分支匹配。现有数据零迁移（旧行 ns_group_id=0 语义不变）。

## 12. 完成定义（DoD）

1. 第 7 节 22 步全部打勾。
2. 第 8、9 节验收清单全部通过。
3. `go build ./...` 与 `npm run build` 干净。
4. 无超出本文档范围的文件被修改（`git status` 核对）。
5. 仓库中不存在任何 ns_group / NsGroup / nsgroupcache 相关代码（分组裁剪自检）。
