// request/request.go
package admin

import (
	"time"
)

// PageInfo 分页请求
type PageInfo struct {
	Page     int `json:"page" form:"page" binding:"required,min=1"`                 // 页码
	PageSize int `json:"pageSize" form:"pageSize" binding:"required,min=1,max=100"` // 每页数量
}

// IDRequest ID请求
type IDRequest struct {
	ID uint `json:"id" form:"id" uri:"id" binding:"required,min=1"`
}

// TimeRangeRequest 时间范围请求
type TimeRangeRequest struct {
	StartTime time.Time `json:"startTime" form:"startTime" time_format:"2006-01-02 15:04:05"`
	EndTime   time.Time `json:"endTime" form:"endTime" time_format:"2006-01-02 15:04:05"`
}

// CreateAdminRequest 创建管理员请求
type CreateUserRequest struct {
	Username string `json:"username" binding:"required,username"`
	Password string `json:"password" binding:"required,password"`
	RealName string `json:"realName"` //暂不使用
	Email    string `json:"email" binding:"required,email"`
	Phone    string `json:"phone"` //暂不使用
	Role     string `json:"role" binding:"required,role"`
	Status   string `json:"status" binding:"required"`
	Remark   string `json:"remark" binding:"max=500"` //暂不使用
}

// UpdateAdminRequest 更新管理员请求
type UpdateAdminRequest struct {
	IDRequest
	RealName string `json:"realName" binding:"required,min=2,max=20"`
	Email    string `json:"email" binding:"required,email"`
	Phone    string `json:"phone" binding:"required,phone"`
	RoleID   uint   `json:"roleId" binding:"required,min=1"`
	Status   int    `json:"status" binding:"required,oneof=0 1"`
	Remark   string `json:"remark" binding:"max=500"`
}

// UpdateAdminPasswordRequest 修改密码请求
type UpdateAdminPasswordRequest struct {
	IDRequest
	OldPassword     string `json:"oldPassword" binding:"required,min=6,max=50"`
	NewPassword     string `json:"newPassword" binding:"required,min=8,max=50,password"`
	ConfirmPassword string `json:"confirmPassword" binding:"required,eqfield=NewPassword"`
}

// AdminListRequest 管理员列表请求
type AdminListRequest struct {
	PageInfo
	Username string `json:"username" form:"username"`
	RealName string `json:"realName" form:"realName"`
	Phone    string `json:"phone" form:"phone"`
	Status   *int   `json:"status" form:"status"` // 使用指针，可以区分0值和未传值
	RoleID   *uint  `json:"roleId" form:"roleId"`
	TimeRangeRequest
}

// ChangeAdminStatusRequest 修改管理员状态
type ChangeAdminStatusRequest struct {
	IDRequest
	Status int `json:"status" binding:"required,oneof=0 1"`
}
