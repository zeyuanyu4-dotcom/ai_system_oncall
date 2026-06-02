package service

import (
	"errors"

	"ai_system_oncall/internal/constant"
	"ai_system_oncall/internal/dto"
	"ai_system_oncall/internal/model"
	"ai_system_oncall/internal/repository"
	"ai_system_oncall/pkg/jwt"
	"ai_system_oncall/pkg/password"

	"gorm.io/gorm"
)

type AuthService struct {
	userRepo *repository.UserRepository
}

func NewAuthService(userRepo *repository.UserRepository) *AuthService {
	return &AuthService{userRepo: userRepo}
}

// Register registers a new user
func (s *AuthService) Register(req *dto.RegisterRequest) (*model.User, error) {
	// Check if email exists
	exists, err := s.userRepo.ExistsByEmail(req.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("邮箱已被注册")
	}

	// Check if username exists
	exists, err = s.userRepo.ExistsByUsername(req.Username)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("用户名已被使用")
	}

	// Hash password
	hashedPassword, err := password.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	// Set default role
	role := req.Role
	if role == "" {
		role = constant.RoleNormalUser
	}

	// Create user
	user := &model.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hashedPassword,
		Phone:        req.Phone,
		Role:         role,
		Status:       constant.UserStatusEnabled,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}

// Login authenticates a user and returns a token
func (s *AuthService) Login(req *dto.LoginRequest) (*dto.LoginResponse, error) {
	// Find user by email
	user, err := s.userRepo.FindByEmail(req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("邮箱或密码错误")
		}
		return nil, err
	}

	// Check password
	if !password.CheckPassword(req.Password, user.PasswordHash) {
		return nil, errors.New("邮箱或密码错误")
	}

	// Check user status
	if !user.IsEnabled() {
		return nil, errors.New("用户已被禁用")
	}

	// Generate token
	token, err := jwt.GenerateToken(user.ID, user.Username, user.Email, user.Role)
	if err != nil {
		return nil, err
	}

	return &dto.LoginResponse{
		Token:    token,
		UserInfo: dto.ToUserInfo(user),
	}, nil
}

// GetCurrentUser gets current user info
func (s *AuthService) GetCurrentUser(userID uint64) (*dto.UserInfo, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}
	return dto.ToUserInfo(user), nil
}
