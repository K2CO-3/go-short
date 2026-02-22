package user

// UpdateUserRequest 更新用户信息请求
type UpdateUserRequest struct {
	Email *string `json:"email" binding:"omitempty,email"`
}

// UpdatePasswordRequest 修改密码请求
type UpdatePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,new_password"`
}
