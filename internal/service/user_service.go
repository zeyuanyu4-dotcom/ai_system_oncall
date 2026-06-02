package service

import (
	"errors"

	"ai_system_oncall/internal/constant"
	"ai_system_oncall/internal/dto"
	"ai_system_oncall/internal/repository"

	"gorm.io/gorm"
)

type UserService struct {
	userRepo *repository.UserRepository
}

func NewUserService(userRepo *repository.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

// GetUserByID gets a user by ID
func (s *UserService) GetUserByID(id uint64) (*dto.UserInfo, error) {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户不存在")
		}
		return nil, err
	}
	return dto.ToUserInfo(user), nil
}

// ListUsers lists users with pagination
func (s *UserService) ListUsers(req *dto.UserListRequest) (*dto.UserListResponse, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	users, total, err := s.userRepo.List(req.Page, req.PageSize, req.Keyword, req.Role, req.Status)
	if err != nil {
		return nil, err
	}

	list := make([]*dto.UserInfo, 0, len(users))
	for _, user := range users {
		list = append(list, dto.ToUserInfo(&user))
	}

	return &dto.UserListResponse{
		Total: total,
		List:  list,
	}, nil
}

// UpdateUser updates a user
func (s *UserService) UpdateUser(id uint64, req *dto.UpdateUserRequest) (*dto.UserInfo, error) {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户不存在")
		}
		return nil, err
	}

	if req.Username != "" {
		// Check if username is taken by another user
		existingUser, _ := s.userRepo.FindByUsername(req.Username)
		if existingUser != nil && existingUser.ID != id {
			return nil, errors.New("用户名已被使用")
		}
		user.Username = req.Username
	}

	if req.Phone != "" {
		user.Phone = req.Phone
	}

	if err := s.userRepo.Update(user); err != nil {
		return nil, err
	}

	return dto.ToUserInfo(user), nil
}

// UpdateUserStatus updates a user's status
func (s *UserService) UpdateUserStatus(operatorRole string, userID uint64, status int8) error {
	// Only system admin can update user status
	if operatorRole != constant.RoleSystemAdmin {
		return errors.New("无权限操作")
	}

	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("用户不存在")
		}
		return err
	}

	// Cannot disable system admin
	if user.Role == constant.RoleSystemAdmin {
		return errors.New("不能禁用系统管理员")
	}

	return s.userRepo.UpdateStatus(userID, status)
}

// DeleteUser deletes a user (soft delete)
func (s *UserService) DeleteUser(operatorRole string, userID uint64) error {
	// Only system admin can delete user
	if operatorRole != constant.RoleSystemAdmin {
		return errors.New("无权限操作")
	}

	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("用户不存在")
		}
		return err
	}

	// Cannot delete system admin
	if user.Role == constant.RoleSystemAdmin {
		return errors.New("不能删除系统管理员")
	}

	return s.userRepo.Delete(userID)
}
