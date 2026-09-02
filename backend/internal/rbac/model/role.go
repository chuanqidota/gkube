package model

import (
	"encoding/json"
	"fmt"
	"time"

	"gkube/pkg/logger"
)

// Role 角色表
type Role struct {
	ID            uint              `gorm:"primaryKey;autoIncrement" json:"id"`
	Name          string            `gorm:"type:varchar(50);uniqueIndex;not null" json:"name"`
	DisplayName   string            `gorm:"type:varchar(100);not null" json:"displayName"`
	ScopeType     string            `gorm:"type:varchar(20);not null;comment:cluster|namespace" json:"scopeType"`
	IsSystem      bool              `gorm:"not null;default:true;comment:系统预置不可删除" json:"isSystem"`
	Permissions   string            `gorm:"type:text;not null;comment:JSON权限定义" json:"permissions"`
	Description   string            `gorm:"type:varchar(255);default:'';comment:角色描述" json:"description"`
	CreatedAt     time.Time         `json:"createdAt"`
	UpdatedAt     time.Time         `json:"updatedAt"`
	PermissionsMap map[string][]string `gorm:"-" json:"-"` // 运行时缓存，不入库
}

func (Role) TableName() string { return "role" }

// GetPermissionsMap 解析 Permissions JSON 到 map，解析失败返回空 map 并记录日志。
func (r *Role) GetPermissionsMap() map[string][]string {
	if r.PermissionsMap != nil {
		return r.PermissionsMap
	}
	m := make(map[string][]string)
	if err := json.Unmarshal([]byte(r.Permissions), &m); err != nil {
		logger.Warn(fmt.Sprintf("角色 %s 的权限JSON解析失败: %v", r.Name, err))
		r.PermissionsMap = m
		return m
	}
	r.PermissionsMap = m
	return m
}

// clusterPerms 集群级角色权限 JSON
var clusterAdminPerms = `{"workload":["read","create","update","delete","terminal"],"network":["read","create","update","delete"],"storage":["read","create","update","delete"],"config":["read","create","update","delete"],"node":["read","cordon","taint","drain","delete"],"namespace":["read","create","update","delete"],"event":["read"],"audit":["read"],"crd":["read","create","update","delete"],"terminal":["terminal"],"cluster_mgmt":["read"]}`

var clusterEditorPerms = `{"workload":["read","create","update","terminal"],"network":["read","create","update"],"storage":["read","create","update"],"config":["read","create","update"],"node":["read"],"namespace":["read"],"event":["read"],"audit":["read"],"crd":["read","create","update"],"terminal":["terminal"]}`

var clusterViewerPerms = `{"workload":["read","terminal"],"network":["read"],"storage":["read"],"config":["read"],"node":["read"],"namespace":["read"],"event":["read"],"audit":["read"],"crd":["read"],"terminal":["terminal"]}`

// nsPerms 命名空间级角色权限 JSON
var nsAdminPerms = `{"workload":["read","create","update","delete","terminal"],"network":["read","create","update","delete"],"storage":["read","create","update","delete"],"config":["read","create","update","delete"],"event":["read"],"crd":["read","create","update","delete"],"terminal":["terminal"]}`

var nsEditorPerms = `{"workload":["read","create","update","terminal"],"network":["read","create","update"],"storage":["read","create","update"],"config":["read","create","update"],"event":["read"],"crd":["read","create","update"],"terminal":["terminal"]}`

var nsViewerPerms = `{"workload":["read"],"network":["read"],"storage":["read"],"config":["read"],"event":["read"],"crd":["read"]}`

// PresetRoles 6个预置角色定义（seed 时写入数据库）
var PresetRoles = []Role{
	{Name: "cluster-admin", DisplayName: "集群管理员", ScopeType: "cluster", IsSystem: true, Permissions: clusterAdminPerms},
	{Name: "cluster-editor", DisplayName: "集群编辑者", ScopeType: "cluster", IsSystem: true, Permissions: clusterEditorPerms},
	{Name: "cluster-viewer", DisplayName: "集群观察者", ScopeType: "cluster", IsSystem: true, Permissions: clusterViewerPerms},
	{Name: "ns-admin", DisplayName: "空间管理员", ScopeType: "namespace", IsSystem: true, Permissions: nsAdminPerms},
	{Name: "ns-editor", DisplayName: "空间编辑者", ScopeType: "namespace", IsSystem: true, Permissions: nsEditorPerms},
	{Name: "ns-viewer", DisplayName: "空间观察者", ScopeType: "namespace", IsSystem: true, Permissions: nsViewerPerms},
}

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
