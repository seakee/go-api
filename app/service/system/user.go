package system

import (
	"context"
	"errors"
	"github.com/seakee/go-api/app/model/system"
	"github.com/seakee/go-api/app/pkg/e"
	repo "github.com/seakee/go-api/app/repository/system"
	"github.com/sk-pkg/logger"
	"github.com/sk-pkg/redis"
	"github.com/sk-pkg/util"
	"gorm.io/gorm"
)

// UserService defines the user service interface.
type UserService interface {
	Create(ctx context.Context, user *system.User) (errCode int, err error)
	Delete(ctx context.Context, id uint) error
	Detail(ctx context.Context, id uint) (user *system.User, errCode int, err error)
	Paginate(ctx context.Context, user *system.User, page, size int) (list []system.User, total int64, err error)
	Update(ctx context.Context, user *system.User) (errCode int, err error)

	Roles(ctx context.Context, userID uint) ([]uint, error)
	UpdateRole(ctx context.Context, userID uint, roles []uint) error
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

func (s userService) UpdateRole(ctx context.Context, userID uint, roles []uint) error {
	hasRole, _ := s.authRepo.HasRole(ctx, userID, "super_admin")
	if hasRole {
		return errors.New("this user is a super admin, cannot change permissions")
	}

	return s.roleUserRepo.UpdateUserRole(ctx, userID, roles)
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
	errCode = e.AccountCantBeNull
	if user.Account != "" {
		var a *system.User

		errCode = e.AccountExists
		a, err = s.userRepo.Detail(ctx, &system.User{Account: user.Account})
		if a == nil {
			errCode = e.ERROR

			user.Status = 1
			user.Salt = util.RandUpStr(32)
			if user.Password == "" || user.Password == "d41d8cd98f00b204e9800998ecf8427e" {
				user.Password = util.MD5(DefaultPassword)
			}

			user.Password = util.MD5(user.Password + user.Salt)

			if !isAvailableName(user.Account) {
				errCode = e.InvalidAccount
				return
			}

			if !isAvailableName(user.UserName) {
				errCode = e.InvalidUserName
				return
			}

			_, err = s.userRepo.Create(ctx, user)
			if err == nil {
				errCode = e.SUCCESS
			}
		}
	}

	return
}

func (s userService) Delete(ctx context.Context, id uint) error {
	hasRole, err := s.authRepo.HasRole(ctx, id, "super_admin")
	if hasRole {
		return errors.New("this user is a super admin, cannot be deleted")
	}

	err = s.userRepo.DeleteByID(ctx, id)
	if err == nil {
		err = s.roleUserRepo.DeleteByRoleID(ctx, id)
	}

	return err
}

func (s userService) Detail(ctx context.Context, id uint) (user *system.User, errCode int, err error) {
	user, err = s.userRepo.Detail(ctx, &system.User{Model: gorm.Model{ID: id}})
	if err != nil {
		return nil, e.AccountNotFound, err
	}

	return user, e.SUCCESS, nil
}

func (s userService) Update(ctx context.Context, user *system.User) (errCode int, err error) {
	hasRole, _ := s.authRepo.HasRole(ctx, user.ID, "super_admin")
	if hasRole {
		return e.ERROR, errors.New("this user is a super admin, cannot be edited")
	}

	errCode = e.AccountNotFound

	u, err := s.userRepo.Detail(ctx, &system.User{Model: gorm.Model{ID: user.ID}})
	if u != nil {
		if !isAvailableName(user.Account) {
			errCode = e.InvalidAccount
			return
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
