package system

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

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
	GetOAuthUser(ctx context.Context, oauthType, tenant, id string) (*system.User, error)
	BindOAuthIdentity(ctx context.Context, input OAuthBindInput) error
	Count(ctx context.Context, user *system.User) (int64, error)
}

var (
	ErrOAuthIdentityConflict = errors.New("oauth identity already bound")
	ErrOAuthBindUserNotFound = errors.New("oauth bind target user not found")
)

type OAuthBindInput struct {
	UserID          uint
	Provider        string
	ProviderTenant  string
	ProviderSubject string
	DisplayName     string
	AvatarURL       string
	RawProfile      interface{}
	SyncUserName    string
	SyncAvatar      string
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

func (u userRepo) GetOAuthUser(ctx context.Context, oauthType, tenant, id string) (*system.User, error) {
	normalizedProvider := strings.TrimSpace(oauthType)
	normalizedTenant := strings.TrimSpace(tenant)
	normalizedID := strings.TrimSpace(id)
	if normalizedProvider == "" || normalizedTenant == "" || normalizedID == "" {
		return nil, nil
	}

	var user system.User
	err := u.db.WithContext(ctx).
		Table(user.TableName()+" AS u").
		Joins("JOIN sys_user_identity ui ON ui.user_id = u.id").
		Where("ui.provider = ? AND ui.provider_tenant = ? AND ui.provider_subject = ?", normalizedProvider, normalizedTenant, normalizedID).
		First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("find oauth user failed: %w", err)
	}

	return &user, nil
}

func (u userRepo) BindOAuthIdentity(ctx context.Context, input OAuthBindInput) error {
	if input.UserID == 0 {
		return fmt.Errorf("user id is empty")
	}

	normalizedProvider := strings.TrimSpace(input.Provider)
	normalizedTenant := strings.TrimSpace(input.ProviderTenant)
	normalizedSubject := strings.TrimSpace(input.ProviderSubject)
	if normalizedProvider == "" || normalizedTenant == "" || normalizedSubject == "" {
		return fmt.Errorf("oauth identity is incomplete")
	}

	if err := clearUserCache(ctx, u.redis); err != nil && u.logger != nil {
		u.logger.Error(ctx, "clear user cache failed", zap.String("action", "bind oauth identity"), zap.Error(err))
	}

	return u.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		identityRepo := userIdentityRepo{db: tx, redis: u.redis, logger: u.logger}

		boundIdentity, err := identityRepo.DetailByProvider(ctx, normalizedProvider, normalizedTenant, normalizedSubject)
		if err != nil {
			return err
		}
		if boundIdentity != nil && boundIdentity.UserID != input.UserID {
			return ErrOAuthIdentityConflict
		}

		boundUser, err := (&system.User{Model: gorm.Model{ID: input.UserID}}).First(ctx, tx)
		if err != nil {
			return err
		}
		if boundUser == nil || boundUser.Status != 1 {
			return ErrOAuthBindUserNotFound
		}

		if boundIdentity == nil {
			now := time.Now()
			userIdentity := &system.UserIdentity{
				UserID:          input.UserID,
				Provider:        normalizedProvider,
				ProviderTenant:  normalizedTenant,
				ProviderSubject: normalizedSubject,
				DisplayName:     strings.TrimSpace(input.DisplayName),
				AvatarURL:       strings.TrimSpace(input.AvatarURL),
				BoundAt:         &now,
				LastLoginAt:     &now,
			}

			if input.RawProfile != nil {
				rawProfileJSON, err := json.Marshal(input.RawProfile)
				if err != nil {
					return err
				}
				userIdentity.RawProfileJSON = string(rawProfileJSON)
			}

			if _, err := userIdentity.Create(ctx, tx); err != nil {
				return fmt.Errorf("create user identity failed: %w", err)
			}
		}

		updates := make(map[string]interface{})
		if userName := strings.TrimSpace(input.SyncUserName); userName != "" {
			updates["user_name"] = userName
		}
		if avatar := strings.TrimSpace(input.SyncAvatar); avatar != "" {
			updates["avatar"] = avatar
		}
		if len(updates) == 0 {
			return nil
		}

		updateUser := &system.User{Model: gorm.Model{ID: input.UserID}}
		return updateUser.Updates(ctx, tx, updates)
	})
}

// NewUserRepo creates a new UserRepo instance.
func NewUserRepo(db *gorm.DB, redis *redis.Manager, logger *logger.Manager) UserRepo {
	return &userRepo{redis: redis, db: db, logger: logger}
}
