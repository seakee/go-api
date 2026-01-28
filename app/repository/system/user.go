package system

import (
	"context"
	"fmt"
	"github.com/seakee/go-api/app/model/system"
	"github.com/seakee/go-api/app/model/system/role"
	"github.com/sk-pkg/logger"
	"github.com/sk-pkg/redis"
	"github.com/sk-pkg/util"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// UserRepo defines the user repository interface.
type UserRepo interface {
	Detail(ctx context.Context, user *system.User) (*system.User, error)
	Create(ctx context.Context, user *system.User) (*system.User, error)
	DetailByAccount(ctx context.Context, account string) (*system.User, error)
	CreateOrUpdate(ctx context.Context, user *system.User) (*system.User, error)
	DetailByID(ctx context.Context, id uint) (*system.User, error)
	DeleteByID(ctx context.Context, id uint) error
	Update(ctx context.Context, user *system.User) error
	UpdateTotpStatus(ctx context.Context, user *system.User) error
	Paginate(ctx context.Context, user *system.User, page, pageSize int) ([]system.User, error)
	GetOAuthUser(ctx context.Context, oauthType, id string) (*system.User, error)
	Count(ctx context.Context, user *system.User) (int64, error)
}

type userRepo struct {
	redis  *redis.Manager
	db     *gorm.DB
	logger *logger.Manager
}

func (u userRepo) Count(ctx context.Context, user *system.User) (int64, error) {
	return user.Count(ctx, u.db)
}

func (u userRepo) Paginate(ctx context.Context, user *system.User, page, pageSize int) ([]system.User, error) {
	if page == 0 {
		page = 1
	}

	switch {
	case pageSize > 100:
		pageSize = 100
	case pageSize <= 0:
		pageSize = 10
	}

	return user.FindWithPagination(ctx, u.db, page, pageSize)
}

func (u userRepo) Update(ctx context.Context, user *system.User) error {
	if user.ID <= 0 {
		return fmt.Errorf("user id is empty")
	}

	err := clearUserCache(ctx, u.redis)
	if err != nil {
		u.logger.Error(ctx, "clear user cache failed", zap.String("action", "update user"), zap.Error(err))
	}

	data := make(map[string]interface{})
	if user.Password != "" {
		salt := util.RandUpStr(32)
		data["salt"] = salt
		data["password"] = util.MD5(user.Password + salt)
	}

	if user.TotpEnabled {
		data["totp_enabled"] = true
	}

	if user.TotpKey != "" {
		data["totp_key"] = user.TotpKey
	}

	if user.Status != 0 {
		data["status"] = user.Status
	}

	if user.UserName != "" {
		data["user_name"] = user.UserName
	}

	if user.Avatar != "" {
		data["avatar"] = user.Avatar
	}

	if user.WechatId != "" {
		data["wechat_id"] = user.WechatId
	}

	if user.GithubId != "" {
		data["github_id"] = user.GithubId
	}

	if user.FeishuId != "" {
		data["feishu_id"] = user.FeishuId
	}

	if user.Account != "" {
		data["account"] = user.Account
	}

	updateUser := &system.User{Model: gorm.Model{ID: user.ID}}

	return updateUser.Updates(ctx, u.db, data)
}

func (u userRepo) UpdateTotpStatus(ctx context.Context, user *system.User) error {
	if user.ID <= 0 {
		return fmt.Errorf("user id is empty")
	}

	err := clearUserCache(ctx, u.redis)
	if err != nil {
		u.logger.Error(ctx, "clear user cache failed", zap.String("action", "update user"), zap.Error(err))
	}

	data := make(map[string]interface{})
	data["totp_enabled"] = user.TotpEnabled

	updateUser := &system.User{Model: gorm.Model{ID: user.ID}}

	return updateUser.Updates(ctx, u.db, data)
}

func (u userRepo) DeleteByID(ctx context.Context, id uint) error {
	err := clearUserCache(ctx, u.redis)
	if err != nil {
		u.logger.Error(ctx, "clear user cache failed", zap.String("action", "delete user"), zap.Error(err))
	}

	user := &system.User{Model: gorm.Model{ID: id}}
	return user.Delete(ctx, u.db)
}

func (u userRepo) Create(ctx context.Context, user *system.User) (*system.User, error) {
	err := clearUserCache(ctx, u.redis)
	if err != nil {
		u.logger.Error(ctx, "clear user cache failed", zap.String("action", "create user"), zap.Error(err))
	}

	// Associate the base role when creating a user
	return user, u.db.WithContext(ctx).Transaction(func(tx *gorm.DB) (err error) {
		if err = tx.Create(&user).Error; err != nil {
			return
		}
		// Check base role
		var baseRole system.Role
		if err = u.db.Where("name = 'base'").
			First(&baseRole).Error; err != nil {
			return
		}
		return tx.Create(&role.User{
			RoleId: baseRole.ID,
			UserId: user.ID,
		}).Error
	})
}

func (u userRepo) DetailByID(ctx context.Context, id uint) (*system.User, error) {
	user := &system.User{Model: gorm.Model{ID: id}}
	return user.First(ctx, u.db)
}

func (u userRepo) CreateOrUpdate(ctx context.Context, user *system.User) (*system.User, error) {
	existUser, err := user.First(ctx, u.db)
	if err != nil {
		return nil, err
	}

	err = clearUserCache(ctx, u.redis)
	if err != nil {
		u.logger.Error(ctx, "clear user cache failed", zap.String("action", "create or update user"), zap.Error(err))
	}

	if existUser != nil {
		user.ID = existUser.ID

		return user, u.Update(ctx, user)
	}

	user.Status = 1

	return u.Create(ctx, user)
}

// DetailByAccount retrieves user details by account.
func (u userRepo) DetailByAccount(ctx context.Context, account string) (*system.User, error) {
	user := &system.User{Account: account}

	return user.First(ctx, u.db)
}

func (u userRepo) Detail(ctx context.Context, user *system.User) (*system.User, error) {
	return user.First(ctx, u.db)
}

func (u userRepo) GetOAuthUser(ctx context.Context, oauthType, id string) (*system.User, error) {
	user := &system.User{}

	switch oauthType {
	case "github":
		user.GithubId = id
	case "feishu":
		user.FeishuId = id
	case "wechat":
		user.WechatId = id
	default:
		return nil, fmt.Errorf("%s type not support", oauthType)
	}

	return user.First(ctx, u.db)
}

// NewUserRepo creates a new UserRepo instance.
func NewUserRepo(db *gorm.DB, redis *redis.Manager, logger *logger.Manager) UserRepo {
	return &userRepo{redis: redis, db: db, logger: logger}
}
