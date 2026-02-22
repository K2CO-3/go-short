package auth

type RegisterRequest struct {
	Username string `json:"username" binding:"required,username"`
	Password string `json:"password" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
}
type LoginRequest struct {
	Username string `json:"username" binding:"required,username"`
	Password string `json:"password" binding:"required"`
}
