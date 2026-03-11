package system

import (
	"context"
	"errors"
	"github.com/seakee/go-api/app/model/system"
	"github.com/seakee/go-api/app/pkg/e"
	pwd "github.com/seakee/go-api/app/pkg/password"
	repo "github.com/seakee/go-api/app/repository/system"
	"github.com/sk-pkg/logger"
	"github.com/sk-pkg/redis"
	"gorm.io/gorm"
)

// UserService defines the user service interface.
type UserService interface {
	Create(ctx context.Context, user *system.User) (errCode int, err error)
	Delete(ctx context.Context, id uint) (errCode int, err error)
	Detail(ctx context.Context, id uint) (user *system.User, errCode int, err error)
	Paginate(ctx context.Context, user *system.User, page, size int) (list []system.User, total int64, err error)
	Update(ctx context.Context, user *system.User) (errCode int, err error)

	Roles(ctx context.Context, userID uint) ([]uint, error)
	UpdateRole(ctx context.Context, userID uint, roles []uint) (errCode int, err error)
	ResetPassword(ctx context.Context, userID uint, password string) (errCode int, err error)
	DisableTfa(ctx context.Context, userID uint) (errCode int, err error)
}

type userService struct {
	redis        *redis.Manager
	logger       *logger.Manager
	userRepo     repo.UserRepo
	roleUserRepo repo.RoleUserRepo
	authRepo     repo.AuthRepo
}

func (s userService) Roles(ctx context.Context, userID uint) ([]uint, error) {
	return s.roleUserRepo.ListByUserID(ctx, userID)
}

func (s userService) UpdateRole(ctx context.Context, userID uint, roles []uint) (errCode int, err error) {
	hasRole, _ := s.authRepo.HasRole(ctx, userID, "super_admin")
	if hasRole {
		return e.UserCanNotBeOperated, errors.New("this user is a super admin, cannot change permissions")
	}

	err = s.roleUserRepo.UpdateUserRole(ctx, userID, roles)
	if err != nil {
		return e.ERROR, err
	}

	return e.SUCCESS, nil
}

func (s userService) ResetPassword(ctx context.Context, userID uint, password string) (errCode int, err error) {
	if password == "" {
		return e.PasswordCanNotBeNull, nil
	}

	hasRole, _ := s.authRepo.HasRole(ctx, userID, "super_admin")
	if hasRole {
		return e.UserCanNotBeOperated, errors.New("this user is a super admin, cannot be operated")
	}

	var user *system.User
	user, err = s.userRepo.Detail(ctx, &system.User{Model: gorm.Model{ID: userID}})
	if user == nil {
		return e.UserNotFound, err
	}

	err = s.userRepo.Update(ctx, &system.User{Model: gorm.Model{ID: userID}, Password: password})
	if err != nil {
		return e.ERROR, err
	}

	return e.SUCCESS, nil
}

func (s userService) DisableTfa(ctx context.Context, userID uint) (errCode int, err error) {
	hasRole, _ := s.authRepo.HasRole(ctx, userID, "super_admin")
	if hasRole {
		return e.UserCanNotBeOperated, errors.New("this user is a super admin, cannot be operated")
	}

	var user *system.User
	user, err = s.userRepo.Detail(ctx, &system.User{Model: gorm.Model{ID: userID}})
	if user == nil {
		return e.UserNotFound, err
	}

	err = s.userRepo.UpdateTotpStatus(ctx, &system.User{Model: gorm.Model{ID: userID}, TotpEnabled: false})
	if err != nil {
		return e.ERROR, err
	}

	return e.SUCCESS, nil
}

func (s userService) Paginate(ctx context.Context, user *system.User, page, size int) (list []system.User, total int64, err error) {
	total, err = s.userRepo.Count(ctx, user)
	if err != nil {
		return nil, 0, err
	}

	list, err = s.userRepo.Paginate(ctx, user, page, size)

	return
}

func (s userService) Create(ctx context.Context, user *system.User) (errCode int, err error) {
	user.Email = normalizeEmail(user.Email)
	user.Phone = normalizePhone(user.Phone)

	errCode = e.IdentifierCantBeNull
	if user.Email == "" && user.Phone == "" {
		return
	}

	if user.Email != "" && !isValidEmail(user.Email) {
		return e.InvalidIdentifier, nil
	}

	if user.Phone != "" && !isValidPhone(user.Phone) {
		return e.InvalidIdentifier, nil
	}

	if user.Email != "" {
		var emailUser *system.User
		emailUser, err = s.userRepo.DetailByEmail(ctx, user.Email)
		if err != nil {
			return e.ERROR, err
		}
		if emailUser != nil {
			return e.IdentifierExists, nil
		}
	}

	if user.Phone != "" {
		var phoneUser *system.User
		phoneUser, err = s.userRepo.DetailByPhone(ctx, user.Phone)
		if err != nil {
			return e.ERROR, err
		}
		if phoneUser != nil {
			return e.IdentifierExists, nil
		}
	}

	errCode = e.ERROR

	user.Status = 1
	if user.Password == "" {
		return e.PasswordCanNotBeNull, nil
	}

	user.Password, err = pwd.HashCredential(user.Password)
	if err != nil {
		return e.ERROR, err
	}

	if !isAvailableName(user.UserName) {
		errCode = e.InvalidUserName
		return
	}

	_, err = s.userRepo.Create(ctx, user)
	if err == nil {
		errCode = e.SUCCESS
	}

	return
}

func (s userService) Delete(ctx context.Context, id uint) (errCode int, err error) {
	hasRole, _ := s.authRepo.HasRole(ctx, id, "super_admin")
	if hasRole {
		return e.UserCanNotBeOperated, errors.New("this user is a super admin, cannot be deleted")
	}

	errCode = e.ERROR
	err = s.userRepo.DeleteByID(ctx, id)
	if err == nil {
		err = s.roleUserRepo.DeleteByUserID(ctx, id)
	}

	if err == nil {
		errCode = e.SUCCESS
	}

	return
}

func (s userService) Detail(ctx context.Context, id uint) (user *system.User, errCode int, err error) {
	user, err = s.userRepo.Detail(ctx, &system.User{Model: gorm.Model{ID: id}})
	if err != nil {
		return nil, e.UserNotFound, err
	}

	return user, e.SUCCESS, nil
}

func (s userService) Update(ctx context.Context, user *system.User) (errCode int, err error) {
	hasRole, _ := s.authRepo.HasRole(ctx, user.ID, "super_admin")
	if hasRole {
		return e.UserCanNotBeOperated, errors.New("this user is a super admin, cannot be edited")
	}

	errCode = e.UserNotFound

	u, err := s.userRepo.Detail(ctx, &system.User{Model: gorm.Model{ID: user.ID}})
	if u != nil {
		user.Email = normalizeEmail(user.Email)
		user.Phone = normalizePhone(user.Phone)

		if user.Email != "" && !isValidEmail(user.Email) {
			return e.InvalidIdentifier, nil
		}

		if user.Phone != "" && !isValidPhone(user.Phone) {
			return e.InvalidIdentifier, nil
		}

		if user.Email == "" && user.Phone == "" && u.Email == "" && u.Phone == "" {
			return e.IdentifierCantBeNull, nil
		}

		if user.Email != "" {
			emailUser, detailErr := s.userRepo.DetailByEmail(ctx, user.Email)
			if detailErr != nil {
				return e.ERROR, detailErr
			}
			if emailUser != nil && emailUser.ID != user.ID {
				return e.IdentifierExists, nil
			}
		}

		if user.Phone != "" {
			phoneUser, detailErr := s.userRepo.DetailByPhone(ctx, user.Phone)
			if detailErr != nil {
				return e.ERROR, detailErr
			}
			if phoneUser != nil && phoneUser.ID != user.ID {
				return e.IdentifierExists, nil
			}
		}

		if !isAvailableName(user.UserName) {
			errCode = e.InvalidUserName
			return
		}

		errCode = e.ERROR
		err = s.userRepo.Update(ctx, user)
		if err == nil {
			errCode = e.SUCCESS
		}
	}

	return
}

// NewUserService creates a new UserService instance.
func NewUserService(redis *redis.Manager, logger *logger.Manager, db *gorm.DB) UserService {
	return &userService{
		redis:        redis,
		logger:       logger,
		userRepo:     repo.NewUserRepo(db, redis, logger),
		roleUserRepo: repo.NewRoleUserRepo(db, redis, logger),
		authRepo:     repo.NewAuthRepo(db, redis, logger),
	}
}
