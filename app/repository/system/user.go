package system

import (
	"context"
	"fmt"
	"strings"

	"github.com/seakee/go-api/app/model/system"
	"github.com/seakee/go-api/app/model/system/role"
	pwd "github.com/seakee/go-api/app/pkg/password"
	"github.com/sk-pkg/logger"
	"github.com/sk-pkg/redis"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// UserRepo defines the user repository interface.
type UserRepo interface {
	Detail(ctx context.Context, user *system.User) (*system.User, error)
	Create(ctx context.Context, user *system.User) (*system.User, error)
	DetailByEmail(ctx context.Context, email string) (*system.User, error)
	DetailByPhone(ctx context.Context, phone string) (*system.User, error)
	DetailByIdentifier(ctx context.Context, identifier string) (*system.User, error)
	CreateOrUpdate(ctx context.Context, user *system.User) (*system.User, error)
	DetailByID(ctx context.Context, id uint) (*system.User, error)
	ListByIDs(ctx context.Context, ids []uint) ([]system.User, error)
	DeleteByID(ctx context.Context, id uint) error
	Update(ctx context.Context, user *system.User) error
	UpdateIdentifier(ctx context.Context, user *system.User) error
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
		hashedPassword := user.Password
		if !pwd.IsBcryptHash(user.Password) {
			hashedPassword, err = pwd.HashCredential(user.Password)
			if err != nil {
				return err
			}
		}
		data["password"] = hashedPassword
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

	if user.Email != "" {
		data["email"] = user.Email
	}

	if user.Phone != "" {
		data["phone"] = user.Phone
	}

	updateUser := &system.User{Model: gorm.Model{ID: user.ID}}

	return updateUser.Updates(ctx, u.db, data)
}

func (u userRepo) UpdateIdentifier(ctx context.Context, user *system.User) error {
	if user.ID <= 0 {
		return fmt.Errorf("user id is empty")
	}

	err := clearUserCache(ctx, u.redis)
	if err != nil {
		u.logger.Error(ctx, "clear user cache failed", zap.String("action", "update user identifier"), zap.Error(err))
	}

	updateUser := &system.User{Model: gorm.Model{ID: user.ID}}
	data := map[string]interface{}{
		"email": user.Email,
		"phone": user.Phone,
	}

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
		baseRoleModel := &system.Role{}
		baseRole, err := baseRoleModel.Where("name = ?", "base").First(ctx, tx)
		if err != nil {
			return
		}
		if baseRole == nil {
			return fmt.Errorf("base role not found")
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

func (u userRepo) ListByIDs(ctx context.Context, ids []uint) ([]system.User, error) {
	if len(ids) == 0 {
		return []system.User{}, nil
	}

	users := make([]system.User, 0, len(ids))
	if err := u.db.WithContext(ctx).
		Select("id", "user_name").
		Where("id IN ?", ids).
		Find(&users).Error; err != nil {
		return nil, err
	}

	return users, nil
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

// DetailByEmail retrieves user details by email.
func (u userRepo) DetailByEmail(ctx context.Context, email string) (*system.User, error) {
	user := &system.User{Email: email}

	return user.First(ctx, u.db)
}

// DetailByPhone retrieves user details by phone.
func (u userRepo) DetailByPhone(ctx context.Context, phone string) (*system.User, error) {
	user := &system.User{Phone: phone}

	return user.First(ctx, u.db)
}

// DetailByIdentifier retrieves user details by email or phone.
func (u userRepo) DetailByIdentifier(ctx context.Context, identifier string) (*system.User, error) {
	normalizedEmail := strings.ToLower(strings.TrimSpace(identifier))
	normalizedPhone := strings.TrimSpace(identifier)
	user := (&system.User{}).Where("email = ? OR phone = ?", normalizedEmail, normalizedPhone)

	return user.First(ctx, u.db)
}

func (u userRepo) Detail(ctx context.Context, user *system.User) (*system.User, error) {
	return user.First(ctx, u.db)
}

func (u userRepo) GetOAuthUser(ctx context.Context, oauthType, id string) (*system.User, error) {
	normalizedID := strings.TrimSpace(id)
	if normalizedID == "" {
		return nil, nil
	}

	switch oauthType {
	case "github":
		return (&system.User{}).Where("github_id = ?", normalizedID).First(ctx, u.db)
	case "feishu":
		return (&system.User{}).Where("feishu_id = ?", normalizedID).First(ctx, u.db)
	case "wechat":
		return (&system.User{}).Where("wechat_id = ?", normalizedID).First(ctx, u.db)
	default:
		return nil, fmt.Errorf("%s type not support", oauthType)
	}
}

// NewUserRepo creates a new UserRepo instance.
func NewUserRepo(db *gorm.DB, redis *redis.Manager, logger *logger.Manager) UserRepo {
	return &userRepo{redis: redis, db: db, logger: logger}
}
