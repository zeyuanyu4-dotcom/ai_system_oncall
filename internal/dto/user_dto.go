package dto

// RegisterRequest 用户注册请求
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=2,max=64"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6,max=32"`
	Phone    string `json:"phone"`
	Role     string `json:"role"`
}

// LoginRequest 用户登录请求
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token    string      `json:"token"`
	UserInfo *UserInfo   `json:"user_info"`
}

// UserInfo 用户信息
type UserInfo struct {
	ID       uint64 `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Role     string `json:"role"`
	Status   int8   `json:"status"`
}

// UpdateUserRequest 更新用户请求
type UpdateUserRequest struct {
	Username string `json:"username" binding:"omitempty,min=2,max=64"`
	Phone    string `json:"phone"`
}

// UpdateUserStatusRequest 更新用户状态请求
type UpdateUserStatusRequest struct {
	Status int8 `json:"status" binding:"oneof=0 1"`
}

// UserListRequest 用户列表请求
type UserListRequest struct {
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	Keyword  string `form:"keyword"`
	Role     string `form:"role"`
	Status   *int8  `form:"status"`
}

// UserListResponse 用户列表响应
type UserListResponse struct {
	Total int64      `json:"total"`
	List  []*UserInfo `json:"list"`
}

// ToUserInfo converts User model to UserInfo
func ToUserInfo(user *User) *UserInfo {
	if user == nil {
		return nil
	}
	return &UserInfo{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Phone:    user.Phone,
		Role:     user.Role,
		Status:   user.Status,
	}
}
