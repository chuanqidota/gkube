package params

type CreateRoleParams struct {
	Name          string `json:"name" binding:"required" label:"角色名"`
	Description   string `json:"description" label:"描述"`
	PermissionIDs []uint `json:"permissionIds" label:"权限ID列表"`
}

type UpdateRoleParams struct {
	ID            uint   `json:"id" binding:"required" label:"角色ID"`
	Description   string `json:"description" label:"描述"`
	PermissionIDs []uint `json:"permissionIds" label:"权限ID列表"`
}

type RoleQueryParams struct {
	Page int    `form:"page" json:"page" label:"页码"`
	Size int    `form:"size" json:"size" label:"每页数量"`
	Name string `form:"name" json:"name" label:"角色名"`
}
