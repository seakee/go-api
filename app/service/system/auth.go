// Package system provides system-related functionality, including user authentication, token generation and other core system services.
package system

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/seakee/go-api/app/config"
	"github.com/seakee/go-api/app/model/system"
	"github.com/seakee/go-api/app/pkg/e"
	"github.com/seakee/go-api/app/pkg/totp"
	repo "github.com/seakee/go-api/app/repository/system"
	"github.com/sk-pkg/logger"
	"github.com/sk-pkg/notify"
	"github.com/sk-pkg/redis"
	"github.com/sk-pkg/util"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const (
	safeCodePrefix  = "admin:system:auth:safeCode:"
	DefaultPassword = "Qaz123$%^"
	oauthStateKey   = "admin:system:auth:oauth:%s"

	feishuUserAccessTokenAPI = "https://open.feishu.cn/open-apis/authen/v1/oidc/access_token"
	feishuUserInfoAPI        = "https://open.feishu.cn/open-apis/authen/v1/user_info"

	wechatAccessTokenAPI = "https://qyapi.weixin.qq.com/cgi-bin/gettoken"
	wechatUserInfoAPI    = "https://qyapi.weixin.qq.com/cgi-bin/auth/getuserinfo"
)

// AccessToken defines the structure of an access token.
type AccessToken struct {
	SafeCode string `json:"safe_code"`  // Safe code for secondary verification
	Token    string `json:"token"`      // JWT token
	ExpireIn int64  `json:"expires_in"` // Token expiration time in seconds
}

// AuthParam defines the structure of authentication parameters.
type AuthParam struct {
	Account     string `json:"account"`                        // User account
	GrantType   string `json:"grant_type" binding:"required"`  // Grant type (password: password login, feishu: Feishu login, wechat: WeChat Work login, github: GitHub login, totp: 2FA login)
	State       string `json:"state"`                          // State, used for some OAuth flows
	Credentials string `json:"credentials" binding:"required"` // Credentials (such as password, verification code, etc.)
}

// safeCode defines the structure of a safe code.
type safeCode struct {
	UserID uint   `json:"user_id"` // User ID
	Action string `json:"action"`  // Action type (e.g., "tfa" for two-factor authentication, "reset_password" for password reset)
}

// claims defines the structure of JWT claims.
type claims struct {
	UserName string `json:"user_name"` // Username
	UserID   uint   `json:"user_id"`   // User ID
	jwt.RegisteredClaims
}

// AuthService defines the authentication service interface.
type AuthService interface {
	Token(ctx context.Context, param *AuthParam) (token AccessToken, errCode int, err error)
	VerifyToken(tokenString string) (userName string, userID uint, err error)
	HasPermission(ctx context.Context, userID uint, permissionHash string) (bool, error)
	HasRole(ctx context.Context, userID uint, role string) (bool, error)
	Profile(ctx context.Context, userID uint) (user map[string]any, errCode int, err error)
	UpdateProfile(ctx context.Context, userID uint, userName, avatar string) (errCode int, err error)
	UserMenuList(ctx context.Context, userID uint) (list system.MenuList, err error)
	ResetPassword(ctx context.Context, sCode string, password string) (errCode int, err error)
	UpdatePassword(ctx context.Context, userID uint, totpCode string, password string) (errCode int, err error)
	UpdateAccount(ctx context.Context, userID uint, totpCode string, account string) (errCode int, err error)
	EnableTfa(ctx context.Context, userID uint, totpCode string, totpKey string) (errCode int, err error)
	DisableTfa(ctx context.Context, userID uint, totpCode string) (errCode int, err error)
	TotpKey(ctx context.Context, userID uint) (key, qrCode string, errCode int, err error)
	TfaStatus(ctx context.Context, userID uint) (enable bool, errCode int, err error)
	OauthUrl(ctx context.Context, oauthType, loginType string) (errCode int, url string)
}

// authService implements the AuthService interface.
type authService struct {
	redis    *redis.Manager
	logger   *logger.Manager
	userRepo repo.UserRepo
	totp     totp.Generator
	authRepo repo.AuthRepo
	menuRepo repo.MenuRepo
	config   config.AdminConfig
	request  *resty.Client
	notify   *notify.Manager
}

// OauthUrl retrieves the OAuth URL.
//
// Parameters:
//   - ctx: Context
//   - oauthType: OAuth type (feishu, gitHub, wechat)
//   - loginType: Login type (qrcode) only for enterprise QR code scanning login
//
// Returns:
//   - errCode: Error code
//   - url: OAuth URL
func (a authService) OauthUrl(ctx context.Context, oauthType, loginType string) (errCode int, oauthUrl string) {
	state := util.RandUpStr(16)
	oauth := a.config.Oauth

	switch oauthType {
	case "feishu":
		oauthUrl = fmt.Sprintf(
			"%s?app_id=%s&state=%s&redirect_uri=%s",
			oauth.Feishu.OauthURL,
			oauth.Feishu.ClientID,
			state,
			url.QueryEscape(oauth.RedirectURL+"?type=feishu"),
		)
		errCode = e.SUCCESS
	case "wechat":
		// Build the OAuth URL with required parameters
		oauthUrl = fmt.Sprintf(
			"%s?appid=%s&redirect_uri=%s&response_type=code&scope=snsapi_privateinfo&agentid=%s&state=%s#wechat_redirect",
			oauth.Wechat.OauthURL,
			oauth.Wechat.CorpID,
			url.QueryEscape(oauth.RedirectURL+"?type=wechat"),
			oauth.Wechat.AgentID,
			state,
		)

		if loginType == "qrcode" {
			oauthUrl = fmt.Sprintf("https://login.work.weixin.qq.com/wwlogin/sso/login?login_type=CorpApp&appid=%s&agentid=%s&redirect_uri=%s&state=%s",
				oauth.Wechat.CorpID,
				oauth.Wechat.AgentID,
				url.QueryEscape(oauth.RedirectURL+"?type=wechat"),
				state,
			)
		}
		errCode = e.SUCCESS
	default:
		errCode = e.OauthTypeNotSupport
	}

	if errCode == e.SUCCESS {
		err := a.redis.SetStringWithContext(ctx, fmt.Sprintf(oauthStateKey, state), oauthType, 180)
		if err != nil {
			errCode = e.ERROR
			oauthUrl = ""
			a.logger.Error(ctx, "set oauth state failed", zap.Error(err))
		}
	}

	return
}

// UpdateProfile updates user information.
//
// Parameters:
//   - userID: User ID
//   - userName: Username
//   - avatar: User avatar
//
// Returns:
//   - errCode: Error code
//   - err: Error message
func (a authService) UpdateProfile(ctx context.Context, userID uint, userName, avatar string) (errCode int, err error) {
	var user *system.User

	if !isAvailableName(userName) {
		errCode = e.InvalidAccount
		return
	}

	// Get user information
	user, err = a.userRepo.Detail(ctx, &system.User{Model: gorm.Model{ID: userID}})
	if user == nil {
		errCode = e.AccountNotFound
		return
	}

	err = a.userRepo.Update(ctx, &system.User{Model: gorm.Model{ID: userID}, UserName: userName, Avatar: avatar})
	if err != nil {
		errCode = e.ERROR
	}

	return
}

// TfaStatus checks if user has TFA enabled.
//
// Parameters:
//   - userID: User ID
//
// Returns:
//   - enable: Whether TFA is enabled
//   - errCode: Error code
//   - err: Error message
func (a authService) TfaStatus(ctx context.Context, userID uint) (enable bool, errCode int, err error) {
	var user *system.User
	// Get user information
	user, err = a.userRepo.Detail(ctx, &system.User{Model: gorm.Model{ID: userID}})
	if user == nil {
		errCode = e.AccountNotFound
		return
	}

	enable = user.TotpEnabled

	return
}

// TotpKey generates a TOTP key.
//
// Parameters:
//   - userID: User ID
//
// Returns:
//   - key: TOTP key
//   - qrCode: QR code
//   - errCode: Error code
//   - err: Error message
func (a authService) TotpKey(ctx context.Context, userID uint) (key, qrCode string, errCode int, err error) {
	var user *system.User
	// Get user information
	user, err = a.userRepo.Detail(ctx, &system.User{Model: gorm.Model{ID: userID}})
	if user == nil {
		errCode = e.AccountNotFound
		return
	}

	key = util.RandBase32Str(32)

	qrCode, err = a.totp.GenerateQRCodeBase64(user.Account, key)

	return
}

// DisableTfa disables TFA.
//
// Parameters:
//   - userID: User ID
//   - totpCode: TOTP code
//
// Returns:
//   - errCode: Error code
//   - err: Error message
func (a authService) DisableTfa(ctx context.Context, userID uint, totpCode string) (errCode int, err error) {
	// Check if safe code is empty
	if totpCode == "" {
		errCode = e.TotpCodeCanNotBeNull
		return
	}

	var user *system.User
	// Get user information
	user, err = a.userRepo.Detail(ctx, &system.User{Model: gorm.Model{ID: userID}})
	if user == nil {
		errCode = e.AccountNotFound
		return
	}

	// Verify TOTP code
	if !a.totp.VerifyTOTPCode(totpCode, user.TotpKey, 1) {
		errCode = e.InvalidTotpCode
		return
	}

	err = a.userRepo.UpdateTotpStatus(ctx, &system.User{Model: gorm.Model{ID: userID}, TotpEnabled: false})
	if err != nil {
		errCode = e.ERROR
	}

	return
}

// EnableTfa enables TFA.
//
// Parameters:
//   - userID: User ID
//   - totpCode: TOTP code
//
// Returns:
//   - errCode: Error code
//   - err: Error message
func (a authService) EnableTfa(ctx context.Context, userID uint, totpCode string, totpKey string) (errCode int, err error) {
	// Check if totpKey is empty
	if totpKey == "" {
		errCode = e.TotpKeyCanNotBeNull
		return
	}

	// Check if safe code is empty
	if totpCode == "" {
		errCode = e.TotpCodeCanNotBeNull
		return
	}

	// Verify TOTP code
	if !a.totp.VerifyTOTPCode(totpCode, totpKey, 1) {
		errCode = e.InvalidTotpCode
		return
	}

	err = a.userRepo.Update(ctx, &system.User{Model: gorm.Model{ID: userID}, TotpEnabled: true, TotpKey: totpKey})
	if err != nil {
		errCode = e.ERROR
	}

	return
}

// UpdateAccount updates the user account.
//
// Parameters:
//   - userID: User ID
//   - totpCode: TOTP code
//   - account: New account
//
// Returns:
//   - errCode: Error code
//   - err: Error message
func (a authService) UpdateAccount(ctx context.Context, userID uint, totpCode string, account string) (errCode int, err error) {
	if !isAvailableName(account) {
		errCode = e.InvalidAccount
		return
	}

	// Check if account is empty
	if account == "" {
		errCode = e.AccountCantBeNull
		return
	}

	// Check if safe code is empty
	if totpCode == "" {
		errCode = e.TotpCodeCanNotBeNull
		return
	}

	var user *system.User
	// Get user information
	user, err = a.userRepo.Detail(ctx, &system.User{Model: gorm.Model{ID: userID}})
	if user == nil {
		errCode = e.AccountNotFound
		return
	}

	// Verify TOTP code
	if !a.totp.VerifyTOTPCode(totpCode, user.TotpKey, 1) {
		errCode = e.InvalidTotpCode
		return
	}

	var u *system.User
	u, err = a.userRepo.Detail(ctx, &system.User{Account: account})
	if u != nil && u.ID != user.ID {
		errCode = e.AccountExists
		return
	}

	err = a.userRepo.Update(ctx, &system.User{Model: gorm.Model{ID: userID}, Account: account})
	if err != nil {
		errCode = e.ERROR
	}

	return
}

// UpdatePassword updates the password.
//
// Parameters:
//   - userID: User ID
//   - totpCode: TOTP code
//   - password: New password
//
// Returns:
//   - errCode: Error code
//   - err: Error message
func (a authService) UpdatePassword(ctx context.Context, userID uint, totpCode string, password string) (errCode int, err error) {
	// Check if password is empty
	if password == "" {
		errCode = e.PasswordCanNotBeNull
		return
	}

	// Check if safe code is empty
	if totpCode == "" {
		errCode = e.TotpCodeCanNotBeNull
		return
	}

	var user *system.User
	// Get user information
	user, err = a.userRepo.Detail(ctx, &system.User{Model: gorm.Model{ID: userID}})
	if user == nil {
		errCode = e.AccountNotFound
		return
	}

	// Verify TOTP code
	if !a.totp.VerifyTOTPCode(totpCode, user.TotpKey, 1) {
		errCode = e.InvalidTotpCode
		return
	}

	err = a.userRepo.Update(ctx, &system.User{Model: gorm.Model{ID: userID}, Password: password})
	if err != nil {
		errCode = e.ERROR
	}

	return
}

// ResetPassword resets the password.
//
// Parameters:
//   - sCode: Safe code
//   - password: New password
//
// Returns:
//   - errCode: Error code
//   - err: Error message
func (a authService) ResetPassword(ctx context.Context, sCode string, password string) (errCode int, err error) {
	// Check if password is empty
	if password == "" {
		errCode = e.PasswordCanNotBeNull
		return
	}

	// Check if safe code is empty
	if sCode == "" {
		errCode = e.SafeCodeCanNotBeNull
		return
	}

	// Parse safe code
	var sc *safeCode
	sc, err = a.parseSafeCode(ctx, sCode)
	if err != nil {
		errCode = e.ERROR
		return
	}

	if sc == nil {
		errCode = e.InvalidSafeCode
		return
	}

	// Check safe code action type
	if sc.Action != "reset_password" {
		errCode = e.InvalidSafeCode
		return
	}

	var user *system.User
	// Get user information
	user, err = a.userRepo.Detail(ctx, &system.User{Model: gorm.Model{ID: sc.UserID}})
	if user == nil {
		errCode = e.AccountNotFound
		return
	}

	err = a.userRepo.Update(ctx, &system.User{Model: gorm.Model{ID: sc.UserID}, Password: password})
	if err != nil {
		errCode = e.ERROR
	}

	return
}

func (a authService) UserMenuList(ctx context.Context, userID uint) (list system.MenuList, err error) {
	var isSuperAdmin bool
	isSuperAdmin, err = a.authRepo.HasRole(ctx, userID, "super_admin")

	return a.menuRepo.UserMenuList(ctx, userID, isSuperAdmin)
}

func (a authService) Profile(ctx context.Context, userID uint) (user map[string]any, errCode int, err error) {
	var u *system.User
	u, err = a.userRepo.DetailByID(ctx, userID)
	if u == nil {
		errCode = e.AccountNotFound
		return
	}

	user = map[string]any{
		"id":        u.ID,
		"user_name": u.UserName,
		"avatar":    u.Avatar,
	}

	return
}

// HasPermission checks if user has the specified permission.
//
// Parameters:
//   - ctx: Context
//   - userID: User ID
//   - permissionHash: Permission hash value
//
// Returns:
//   - bool: Whether user has the specified permission
//   - error: Error message
func (a authService) HasPermission(ctx context.Context, userID uint, permissionHash string) (bool, error) {
	return a.authRepo.HasPermission(ctx, userID, permissionHash)
}

// HasRole checks if user has the specified role.
//
// Parameters:
//   - ctx: Context
//   - userID: User ID
//   - role: Role name
//
// Returns:
//   - bool: Whether user has the specified role
//   - error: Error message
func (a authService) HasRole(ctx context.Context, userID uint, role string) (bool, error) {
	return a.authRepo.HasRole(ctx, userID, role)
}

// Token generates an access token.
//
// Parameters:
//   - ctx: Context
//   - param: Authentication parameters
//
// Returns:
//   - token: Access token
//   - errCode: Error code
//   - err: Error message
//
// Example:
//
//	authParam := &AuthParam{
//	    Account: "user@example.com",
//	    GrantType: "password",
//	    Credentials: "password123",
//	}
//	token, errCode, err := authService.Token(context.Background(), authParam)
//	if err != nil {
//	    // Handle error
//	}
//	// Use generated token
func (a authService) Token(ctx context.Context, param *AuthParam) (token AccessToken, errCode int, err error) {
	var user *system.User

	// Verify based on different grant types
	switch param.GrantType {
	case "password":
		user, errCode, err = a.verifyByPassword(ctx, param.Account, param.Credentials)
	case "totp":
		user, errCode, err = a.verifyByTotp(ctx, param.Credentials, param.Account)
	case "feishu":
		user, errCode, err = a.verifyByFeishu(ctx, param.Credentials, param.State)
	case "wechat":
		user, errCode, err = a.verifyByWechat(ctx, param.Credentials, param.State)
	default:
		errCode = e.InvalidGrantType
	}

	// Handle different cases based on verification result
	switch {
	case errCode == e.SUCCESS && user != nil:
		// Verification successful, generate JWT token
		var jwtToken string
		jwtToken, err = a.generateToken(user.UserName, user.ID)
		if err != nil {
			errCode = e.ERROR
			return token, errCode, err
		}

		token.Token = jwtToken
		token.ExpireIn = a.config.TokenExpireIn
	case errCode == e.NeedTfa && user != nil:
		// Need two-factor authentication
		sc := safeCode{
			UserID: user.ID,
			Action: "tfa",
		}
		token.SafeCode, err = a.generateSafeCode(ctx, sc)
		if err != nil {
			errCode = e.ERROR
		}
	case errCode == e.NeedResetPWD && user != nil:
		// Need password reset
		sc := safeCode{
			UserID: user.ID,
			Action: "reset_password",
		}
		token.SafeCode, err = a.generateSafeCode(ctx, sc)
		if err != nil {
			errCode = e.ERROR
		}
	}

	return
}

// verifyByFeishu verifies user identity using Feishu authorization code.
//
// Parameters:
//   - ctx: Context
//   - code: Feishu authorization code
//   - state: State parameter
//
// Returns:
//   - user: User information
//   - errCode: Error code
//   - err: Error message
func (a authService) verifyByFeishu(ctx context.Context, code, state string) (user *system.User, errCode int, err error) {
	oauthType, err := a.redis.GetStringWithContext(ctx, fmt.Sprintf(oauthStateKey, state))
	if oauthType != "feishu" {
		errCode = e.InvalidOauthState
		return
	}

	_, err = a.redis.DelWithContext(ctx, fmt.Sprintf(oauthStateKey, state))
	if err != nil {
		a.logger.Error(ctx, "failed to delete Feishu authorization state", zap.Error(err))
	}

	accessToken, err := a.getFeishuUserAccessToken(ctx, code)
	if err != nil {
		errCode = e.ERROR
		return
	}

	userID, err := a.getFeishuUserID(ctx, accessToken)
	if err != nil {
		errCode = e.ERROR
		return
	}

	user, err = a.userRepo.Detail(ctx, &system.User{FeishuId: userID})

	return
}

// getFeishuUserAccessToken retrieves Feishu user access token.
//
// Parameters:
//   - ctx: Context
//   - code: Feishu user authorization code
//
// Returns:
//   - token: Feishu user access token
//   - err: Error message
func (a authService) getFeishuUserAccessToken(ctx context.Context, code string) (token string, err error) {
	type tokenResult struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}

	appToken, err := a.notify.Lark.GetToken("go-api")
	if err != nil {
		return
	}

	resp, err := a.request.R().SetContext(ctx).
		SetAuthToken(appToken).
		SetBody(map[string]string{
			"grant_type": "authorization_code",
			"code":       code,
		}).Post(feishuUserAccessTokenAPI)
	if err != nil {
		return
	}

	var result tokenResult
	err = json.Unmarshal(resp.Body(), &result)
	if err != nil {
		return
	}

	if result.Code != 0 {
		err = fmt.Errorf("feishu get user access token error: %s", result.Msg)
		return
	}

	return result.Data.AccessToken, nil
}

// getFeishuUserID retrieves Feishu user ID.
//
// Parameters:
//   - ctx: Context
//   - token: Feishu user access token
//
// Returns:
//   - userID: Feishu user ID
//   - err: Error message
func (a authService) getFeishuUserID(ctx context.Context, token string) (userID string, err error) {
	type userInfoResult struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			UserId string `json:"user_id"`
		} `json:"data"`
	}
	resp, err := a.request.R().SetContext(ctx).
		SetAuthToken(token).
		Get(feishuUserInfoAPI)
	if err != nil {
		return
	}

	var result userInfoResult
	err = json.Unmarshal(resp.Body(), &result)
	if err != nil {
		return
	}

	if result.Code != 0 {
		err = fmt.Errorf("feishu get user info error: %s", result.Msg)
		return
	}

	return result.Data.UserId, nil
}

// verifyByWechat verifies user identity using WeChat Work authorization code.
//
// Parameters:
//   - ctx: Context
//   - code: WeChat Work authorization code
//   - state: State parameter
//
// Returns:
//   - user: User information
//   - errCode: Error code
//   - err: Error message
func (a authService) verifyByWechat(ctx context.Context, code, state string) (user *system.User, errCode int, err error) {
	oauthType, err := a.redis.GetStringWithContext(ctx, fmt.Sprintf(oauthStateKey, state))
	if oauthType != "wechat" {
		errCode = e.InvalidOauthState
		return
	}

	_, err = a.redis.DelWithContext(ctx, fmt.Sprintf(oauthStateKey, state))
	if err != nil {
		a.logger.Error(ctx, "failed to delete WeChat Work authorization state", zap.Error(err))
	}

	accessToken, err := a.getWechatAccessToken(ctx)
	if err != nil {
		errCode = e.ERROR
		return
	}

	userID, err := a.getWechatUserID(ctx, accessToken, code)
	if err != nil {
		errCode = e.ERROR
		return
	}

	user, err = a.userRepo.Detail(ctx, &system.User{WechatId: userID})

	return
}

// getWechatAccessToken retrieves WeChat Work access_token.
//
// Parameters:
//   - ctx: Context
//
// Returns:
//   - token: WeChat Work access_token
//   - err: Error message
func (a authService) getWechatAccessToken(ctx context.Context) (token string, err error) {
	const wechatTokenCacheKey = "admin:system:auth:wechat:token"

	// Try to get token from Redis cache first
	cachedToken, err := a.redis.GetStringWithContext(ctx, wechatTokenCacheKey)
	if err == nil && cachedToken != "" {
		return cachedToken, nil
	}

	type tokenResult struct {
		ErrCode     int    `json:"errcode"`
		ErrMsg      string `json:"errmsg"`
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}

	tokenURL := fmt.Sprintf("%s?corpid=%s&corpsecret=%s", wechatAccessTokenAPI,
		a.config.Oauth.Wechat.CorpID, a.config.Oauth.Wechat.CorpSecret)

	request := a.request
	if a.config.Oauth.Wechat.ProxyURL != "" {
		request = request.SetProxy(a.config.Oauth.Wechat.ProxyURL)
	}

	resp, err := request.R().SetContext(ctx).Get(tokenURL)
	if err != nil {
		return
	}

	var result tokenResult
	err = json.Unmarshal(resp.Body(), &result)
	if err != nil {
		return
	}

	if result.ErrCode != 0 {
		err = fmt.Errorf("wechat work get access token error: %s", result.ErrMsg)
		return
	}

	// Store token in Redis cache, set expiration to expires_in-60 seconds (refresh 60 seconds early)
	err = a.redis.SetStringWithContext(ctx, wechatTokenCacheKey, result.AccessToken, result.ExpiresIn-60)
	if err != nil {
		// If cache set fails, only log the error, don't affect main flow
		a.logger.Error(ctx, "cache wechat access token failed", zap.Error(err))
	}

	return result.AccessToken, nil
}

// getWechatUserID retrieves WeChat Work user ID.
//
// Parameters:
//   - ctx: Context
//   - accessToken: WeChat Work access_token
//   - code: Authorization code
//
// Returns:
//   - userID: WeChat Work user ID
//   - err: Error message
func (a authService) getWechatUserID(ctx context.Context, accessToken, code string) (userID string, err error) {
	type userInfoResult struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
		UserID  string `json:"userid"`
		OpenID  string `json:"openid"`
	}

	userURL := fmt.Sprintf("%s?access_token=%s&code=%s", wechatUserInfoAPI,
		accessToken, code)

	request := a.request
	if a.config.Oauth.Wechat.ProxyURL != "" {
		request = request.SetProxy(a.config.Oauth.Wechat.ProxyURL)
	}

	resp, err := request.R().SetContext(ctx).Get(userURL)
	if err != nil {
		return
	}

	var result userInfoResult
	err = json.Unmarshal(resp.Body(), &result)
	if err != nil {
		return
	}

	if result.ErrCode != 0 {
		err = fmt.Errorf("wechat work get user info error: %s", result.ErrMsg)
		return
	}

	// Prefer UserID, use OpenID if not available
	if result.UserID != "" {
		return result.UserID, nil
	}

	return result.OpenID, nil
}

// verifyByPassword verifies user by password.
//
// Parameters:
//   - ctx: Context
//   - account: User account
//   - password: User password
//
// Returns:
//   - user: User information upon successful verification
//   - errCode: Error code
//   - err: Error message
func (a authService) verifyByPassword(ctx context.Context, account, password string) (user *system.User, errCode int, err error) {
	// Check if account and password are empty
	if account == "" || password == "" {
		errCode = e.AccountOrPasswordCanNotBeNull
		return
	}

	// Query user information
	user, err = a.userRepo.Detail(ctx, &system.User{Account: account, Status: 1})
	if user == nil {
		errCode = e.AccountNotFound
		return
	}

	// Verify password
	md5Password := util.MD5(password + user.Salt)

	if user.Password != md5Password {
		errCode = e.AccountOrPasswordFail
		return
	}

	// Check if it's the initial password
	if md5Password == util.MD5(util.MD5(DefaultPassword)+user.Salt) {
		errCode = e.NeedResetPWD
		return
	}

	// Check if two-factor authentication is enabled
	if user.TotpEnabled {
		errCode = e.NeedTfa
		return
	}

	return
}

// verifyByTotp verifies user by TOTP.
//
// Parameters:
//   - ctx: Context
//   - totpCode: TOTP verification code
//   - sCode: Safe code
//
// Returns:
//   - user: User information upon successful verification
//   - errCode: Error code
//   - err: Error message
func (a authService) verifyByTotp(ctx context.Context, totpCode, sCode string) (user *system.User, errCode int, err error) {
	// Check if TOTP verification code and safe code are empty
	if totpCode == "" || sCode == "" {
		errCode = e.InvalidTotpCode
		return
	}

	// Parse safe code
	var sc *safeCode
	sc, err = a.parseSafeCode(ctx, sCode)
	if err != nil {
		errCode = e.ERROR
		return
	}

	if sc == nil {
		errCode = e.InvalidSafeCode
		return
	}

	// Check safe code action type
	if sc.Action != "tfa" {
		errCode = e.InvalidSafeCode
		return
	}

	// Get user information
	user, err = a.userRepo.Detail(ctx, &system.User{Model: gorm.Model{ID: sc.UserID}})
	// Verify TOTP code
	if !a.totp.VerifyTOTPCode(totpCode, user.TotpKey, 1) {
		errCode = e.InvalidTotpCode
		return
	}

	return
}

// generateToken generates a JWT token.
//
// Parameters:
//   - userName: Username
//   - userID: User ID
//
// Returns:
//   - token: Generated JWT token
//   - err: Error message
func (a authService) generateToken(userName string, userID uint) (token string, err error) {
	// Create JWT claims
	c := claims{
		UserName: userName,
		UserID:   userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Second * time.Duration(a.config.TokenExpireIn))),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	// Create JWT token
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	token, err = t.SignedString([]byte(a.config.JwtSecret))
	return
}

// VerifyToken verifies a JWT token.
//
// Parameters:
//   - tokenString: JWT token string
//
// Returns:
//   - userName: Username
//   - userID: User ID
//   - err: Error message
func (a authService) VerifyToken(tokenString string) (userName string, userID uint, err error) {
	// Parse JWT token
	token, err := jwt.ParseWithClaims(tokenString, &claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(a.config.JwtSecret), nil
	})

	if err != nil {
		return "", 0, err
	}

	// Verify token and extract claims
	if c, ok := token.Claims.(*claims); ok && token.Valid {
		return c.UserName, c.UserID, nil
	}

	return "", 0, errors.New("invalid token")
}

// generateSafeCode generates a safe code.
//
// Parameters:
//   - ctx: Context
//   - safeCode: Safe code structure
//
// Returns:
//   - string: Generated safe code
//   - error: Error message
func (a authService) generateSafeCode(ctx context.Context, safeCode safeCode) (string, error) {
	// Generate random string as safe code
	code := util.RandUpStr(32)

	// Store safe code in Redis
	key := util.SpliceStr(safeCodePrefix, code)
	err := a.redis.SetJSONWithContext(ctx, key, safeCode, a.config.SafeCodeExpireIn)

	return code, err
}

// parseSafeCode parses a safe code.
//
// Parameters:
//   - ctx: Context
//   - code: Safe code
//
// Returns:
//   - *safeCode: Parsed safe code structure
//   - error: Error message
func (a authService) parseSafeCode(ctx context.Context, code string) (*safeCode, error) {
	var sc *safeCode

	// Get safe code information from Redis
	key := util.SpliceStr(safeCodePrefix, code)
	err := a.redis.GetJSONWithContext(ctx, key, &sc)
	if err != nil {
		return nil, err
	}

	// Delete safe code from Redis
	_, err = a.redis.Del(key)
	if err != nil {
		a.logger.Error(ctx, "redis delete safe code failed", zap.Error(err))
	}

	return sc, nil
}

// NewAuthService creates a new AuthService instance.
//
// Parameters:
//   - redis: Redis manager
//   - logger: Logger manager
//   - db: Database connection
//   - notify: Notification manager
//
// Returns:
//   - AuthService: Authentication service interface
func NewAuthService(redis *redis.Manager, logger *logger.Manager, db *gorm.DB, notify *notify.Manager) AuthService {
	return &authService{
		redis:    redis,
		logger:   logger,
		userRepo: repo.NewUserRepo(db, redis, logger),
		totp:     totp.NewGenerator("go-api-admin"),
		authRepo: repo.NewAuthRepo(db, redis, logger),
		menuRepo: repo.NewMenuRepo(db, redis, logger),
		config:   config.Get().System.Admin,
		request:  resty.New(),
		notify:   notify,
	}
}
