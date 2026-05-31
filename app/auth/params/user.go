package params

type CreateUserParams struct {
	Username    string `json:"username" binding:"required" label:"用户名"`
	Password    string `json:"password" binding:"required" label:"密码"`
	Email       string `json:"email" label:"邮箱"`
	DisplayName string `json:"displayName" label:"显示名称"`
	RoleIDs     []uint `json:"roleIds" label:"角色ID列表"`
}

type UpdateUserParams struct {
	ID          uint   `json:"id" binding:"required" label:"用户ID"`
	Email       string `json:"email" label:"邮箱"`
	DisplayName string `json:"displayName" label:"显示名称"`
	Status      *int   `json:"status" label:"状态"`
	RoleIDs     []uint `json:"roleIds" label:"角色ID列表"`
}

type UserQueryParams struct {
	Page     int    `form:"page" json:"page" label:"页码"`
	Size     int    `form:"size" json:"size" label:"每页数量"`
	Username string `form:"username" json:"username" label:"用户名"`
	Status   *int   `form:"status" json:"status" label:"状态"`
}

type ChangePasswordParams struct {
	OldPassword string `json:"oldPassword" binding:"required" label:"旧密码"`
	NewPassword string `json:"newPassword" binding:"required" label:"新密码"`
}
