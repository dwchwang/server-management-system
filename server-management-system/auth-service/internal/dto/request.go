package dto

// RegisterRequest is the request body for user registration.
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=100"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	FullName string `json:"full_name" binding:"required"`
	RoleName string `json:"role_name" binding:"required,oneof=admin operator viewer"`
}

// LoginRequest is the request body for user login.
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// RefreshRequest is the request body for token refresh.
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}
