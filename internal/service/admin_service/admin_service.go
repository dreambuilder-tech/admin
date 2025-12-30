package admin_service

import (
	"admin/internal/common/auth"
	"admin/internal/model"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
	"wallet/common-lib/config"
	"wallet/common-lib/consts/rds_keys"
	"wallet/common-lib/dbs"
	"wallet/common-lib/kms"
	"wallet/common-lib/rdb"
	"wallet/common-lib/rpcx/kms_rpcx"
	"wallet/common-lib/utils/authx"
	"wallet/common-lib/utils/bcryptx"
	"wallet/common-lib/zapx"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// GetRolesResp 获取角色列表响应
type GetRolesResp struct {
	Roles []*model.Role `json:"roles"`
}

// GetRoles 获取角色列表
func GetRoles(ctx context.Context) (*GetRolesResp, error) {
	roleModel := new(model.Role)
	roles, err := roleModel.GetAll(ctx, dbs.Admin)
	if err != nil {
		zapx.ErrorCtx(ctx, "get roles failed", zap.Error(err))
		return nil, err
	}
	return &GetRolesResp{Roles: roles}, nil
}

// CreateAdminReq 创建管理员请求
type CreateAdminReq struct {
	Account  string `json:"account" binding:"required"`
	Password string `json:"password" binding:"required"`
	RoleID   int    `json:"role_id" binding:"required"`
}

// CreateAdmin 创建管理员
func CreateAdmin(ctx context.Context, req *CreateAdminReq) error {
	// 验证角色是否存在
	roleModel := new(model.Role)
	exists, err := roleModel.Exists(ctx, dbs.Admin, req.RoleID)
	if err != nil {
		zapx.ErrorCtx(ctx, "check role exists failed", zap.Error(err))
		return err
	}
	if !exists {
		return errors.New("角色不存在")
	}

	// 检查账号是否已存在
	userModel := new(model.Admin)
	accountExists, err := userModel.AccountExists(ctx, dbs.Admin, req.Account)
	if err != nil {
		zapx.ErrorCtx(ctx, "check account exists failed", zap.Error(err))
		return err
	}
	if accountExists {
		return errors.New("账号已存在")
	}

	// todo 之后看具体需求对密码明文长度或做强密码校验
	// 加密密码
	hashedPassword, err := bcryptx.Hash(req.Password)
	if err != nil {
		zapx.ErrorCtx(ctx, "hash password failed", zap.Error(err))
		return err
	}

	// 创建用户
	user := &model.Admin{
		Account:  req.Account,
		Password: hashedPassword,
		RoleID:   req.RoleID,
		Status:   1, // 默认启用
	}

	if err = user.Create(ctx, dbs.Admin); err != nil {
		zapx.ErrorCtx(ctx, "create admin user failed", zap.Error(err))
		return err
	}

	return nil
}

// LoginReq 登录请求
type LoginReq struct {
	Account  string `json:"account" binding:"required"`
	Password string `json:"password" binding:"required"`
	TotpCode string `json:"totp_code"` // 谷歌验证器动态码
}

// LoginResp 登录响应
type LoginResp struct {
	SessionID string `json:"session_id"`
	User      *User  `json:"user"`
}

type User struct {
	ID           int64  `json:"id"`
	Account      string `json:"account"`
	RoleID       int    `json:"role_id"`
	IsSuperAdmin bool   `json:"is_super_admin"`
	MfaEnabled   bool   `json:"mfa_enabled"`
}

// Login 管理员登录
func Login(ctx context.Context, req *LoginReq, loginIP string, svrConf *config.ServiceConfig) (*LoginResp, error) {
	// 查询用户
	userModel := new(model.Admin)
	if err := userModel.GetByAccount(ctx, dbs.Admin, req.Account); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("账号或密码错误")
		}
		zapx.ErrorCtx(ctx, "get user by account failed", zap.Error(err))
		return nil, err
	}

	// 检查用户状态
	if userModel.Status == 0 {
		return nil, errors.New("账号已被禁用")
	}

	// 验证密码
	if !bcryptx.Check(userModel.Password, req.Password) {
		return nil, errors.New("账号或密码错误")
	}
	// 检查IP白名单
	ipAllowed, err := ValidateLoginIP(ctx, loginIP)
	if err != nil {
		zapx.ErrorCtx(ctx, "validate login ip failed", zap.Error(err))
		return nil, errors.New("登录验证失败")
	}
	if !ipAllowed {
		zapx.WarnCtx(ctx, "login ip not in whitelist", zap.String("ip", loginIP), zap.String("account", req.Account))
		return nil, errors.New("登录IP不在白名单内")
	}

	// 如果启用了谷歌验证器，验证动态码
	if len(userModel.MfaSecret) > 0 && svrConf.Service.Auth.Login.Totp {
		if req.TotpCode == "" {
			return nil, errors.New("请输入谷歌验证器动态码")
		}

		// 解密MFA密钥
		secret, err := kms_rpcx.Decrypt(ctx, userModel.MfaSecret, kms.PurposeUserTotpSecret, userModel.ID)
		if err != nil {
			zapx.ErrorCtx(ctx, "decrypt err", zap.Error(err))
			return nil, fmt.Errorf("decrypt err: %v", err)
		}
		// 验证TOTP码
		if !authx.ValidateTOTP(secret, req.TotpCode) {
			return nil, errors.New("谷歌验证器动态码错误")
		}
	}

	// 生成session
	sessionID, err := auth.NewSessionID()
	if err != nil {
		zapx.ErrorCtx(ctx, "generate session id failed", zap.Error(err))
		return nil, err
	}

	// 保存session到Redis
	sessionUser := &auth.User{
		ID:      userModel.ID,
		Account: userModel.Account,
		Role:    userModel.RoleID,
		LoginIP: loginIP,
		LoginAt: time.Now().Unix(),
	}

	sessionData, err := json.Marshal(sessionUser)
	if err != nil {
		zapx.ErrorCtx(ctx, "marshal session data failed", zap.Error(err))
		return nil, err
	}

	if err = rdb.Client.Set(ctx, sessionID, sessionData, auth.SessionExpireTime).Err(); err != nil {
		zapx.ErrorCtx(ctx, "save session to redis failed", zap.Error(err))
		return nil, err
	}

	// 更新登录信息
	if err = userModel.UpdateLoginInfo(ctx, dbs.Admin, userModel.ID, loginIP); err != nil {
		zapx.ErrorCtx(ctx, "update login info failed", zap.Error(err))
		// 不影响登录流程，只记录日志
	}

	return &LoginResp{
		SessionID: sessionID,
		User: &User{
			ID:           userModel.ID,
			Account:      userModel.Account,
			RoleID:       userModel.RoleID,
			IsSuperAdmin: userModel.RoleID == auth.SuperAdminRoleID,
			MfaEnabled:   len(userModel.MfaSecret) > 0,
		},
	}, nil
}

// GenerateMFASecretReq 生成MFA密钥请求
type GenerateMFASecretReq struct {
	UserID int64 `json:"user_id"`
}

// GenerateMFASecretResp 生成MFA密钥响应
type GenerateMFASecretResp struct {
	QRCode string `json:"qr_code"` // base64编码的二维码图片
}

// GenerateMFASecret 生成谷歌验证器密钥并返回二维码
func GenerateMFASecret(ctx context.Context, userID int64, issuer string) (*GenerateMFASecretResp, error) {
	if rdb.Client == nil {
		return nil, errors.New("MFA功能需要Redis支持，当前Redis不可用")
	}

	// 查询用户
	userModel := new(model.Admin)
	if err := userModel.GetByID(ctx, dbs.Admin, userID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户不存在")
		}
		zapx.ErrorCtx(ctx, "get user by id failed", zap.Error(err))
		return nil, err
	}

	// 检查是否已绑定
	if len(userModel.MfaSecret) > 0 {
		return nil, errors.New("已绑定谷歌验证器，请先解绑")
	}

	// 生成TOTP密钥和二维码
	secret, qrCode, err := authx.GenerateTOTPSecret(userModel.Account, issuer)
	if err != nil {
		zapx.ErrorCtx(ctx, "generate totp secret failed", zap.Error(err))
		return nil, err
	}

	// 临时保存到Redis（5分钟有效期）
	redisKey := rds_keys.AdminMFATempSecret(userID)
	if err = rdb.Client.Set(ctx, redisKey, secret, 5*time.Minute).Err(); err != nil {
		zapx.ErrorCtx(ctx, "save temp mfa secret to redis failed", zap.Error(err))
		return nil, err
	}

	return &GenerateMFASecretResp{
		QRCode: qrCode,
	}, nil
}

// BindMFAReq 绑定MFA请求
type BindMFAReq struct {
	UserID   int64  `json:"user_id"`
	Password string `json:"password" binding:"required"`
	TotpCode string `json:"totp_code" binding:"required"`
}

// BindMFA 绑定谷歌验证器
func BindMFA(ctx context.Context, req *BindMFAReq) error {
	// 查询用户
	userModel := new(model.Admin)
	if err := userModel.GetByID(ctx, dbs.Admin, req.UserID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("用户不存在")
		}
		zapx.ErrorCtx(ctx, "get user by id failed", zap.Error(err))
		return err
	}

	// 验证密码
	if !bcryptx.Check(userModel.Password, req.Password) {
		return errors.New("密码错误")
	}

	// 从Redis获取临时密钥
	redisKey := rds_keys.AdminMFATempSecret(req.UserID)
	secret, err := rdb.Client.Get(ctx, redisKey).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return errors.New("密钥已过期，请重新获取二维码")
		}
		zapx.ErrorCtx(ctx, "get temp mfa secret from redis failed", zap.Error(err))
		return err
	}

	// 验证TOTP码
	if !authx.ValidateTOTP(secret, req.TotpCode) {
		return errors.New("谷歌验证器动态码错误")
	}

	// 加密密钥
	blob, _, err := kms_rpcx.Encrypt(ctx, secret, kms.PurposeUserTotpSecret, userModel.ID)
	if err != nil {
		zapx.ErrorCtx(ctx, "encrypt error", zap.Error(err))
		return err
	}

	// 保存到数据库
	if err = userModel.UpdateMfaSecret(ctx, dbs.Admin, req.UserID, blob); err != nil {
		zapx.ErrorCtx(ctx, "update mfa secret failed", zap.Error(err))
		return err
	}

	// 删除临时密钥
	_ = rdb.Client.Del(ctx, redisKey).Err()

	return nil
}

// UnbindMFAReq 解绑MFA请求
type UnbindMFAReq struct {
	OperatorID   int64 `json:"operator_id"`    // 操作者ID
	TargetUserID int64 `json:"target_user_id"` // 目标用户ID
}

// UnbindMFA 解绑谷歌验证器（仅超级管理员可操作）
func UnbindMFA(ctx context.Context, req *UnbindMFAReq) error {
	// 检查操作者是否为超级管理员
	operatorModel := new(model.Admin)
	if err := operatorModel.GetByID(ctx, dbs.Admin, req.OperatorID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("操作者不存在")
		}
		zapx.ErrorCtx(ctx, "get operator by id failed", zap.Error(err))
		return err
	}

	if operatorModel.RoleID != auth.SuperAdminRoleID {
		return errors.New("仅超级管理员可以解绑谷歌验证器")
	}

	// 查询目标用户
	targetUserModel := new(model.Admin)
	if err := targetUserModel.GetByID(ctx, dbs.Admin, req.TargetUserID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("目标用户不存在")
		}
		zapx.ErrorCtx(ctx, "get target user by id failed", zap.Error(err))
		return err
	}

	// 清空MFA密钥
	if err := targetUserModel.ClearMfaSecret(ctx, dbs.Admin, req.TargetUserID); err != nil {
		zapx.ErrorCtx(ctx, "clear mfa secret failed", zap.Error(err))
		return err
	}

	// 记录日志
	zapx.InfoCtx(ctx, "unbind mfa success",
		zap.Int64("operator_id", req.OperatorID),
		zap.String("operator_account", operatorModel.Account),
		zap.Int64("target_user_id", req.TargetUserID),
		zap.String("target_account", targetUserModel.Account))

	return nil
}
func ValidateLoginIP(ctx context.Context, loginIP string) (bool, error) {
	systemConf := new(model.SystemConf)
	err := systemConf.GetByKey(ctx, dbs.System, "ip_whitelist")
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return true, nil
		}
		zapx.ErrorCtx(ctx, "get ip whitelist config failed", zap.Error(err))
		return false, err
	}

	if systemConf.Val == "" {
		return true, nil
	}

	ips := strings.Split(systemConf.Val, ",")
	for _, ip := range ips {
		ip = strings.TrimSpace(ip)
		if ip == "" {
			continue
		}

		if isIPMatch(ip, loginIP) {
			return true, nil
		}
	}

	return false, nil
}

func isIPMatch(pattern, ip string) bool {
	if pattern == ip {
		return true
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return false
	}

	return re.MatchString(ip)
}
