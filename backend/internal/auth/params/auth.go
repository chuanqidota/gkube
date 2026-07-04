package params

type LoginParams struct {
	Username string `json:"username" binding:"required" label:"用户名"`
	Password string `json:"password" binding:"required" label:"密码"`
}

type RefreshParams struct {
	RefreshToken string `json:"refreshToken" binding:"required" label:"Refresh Token"`
}
