package perm_service

import (
	"admin/internal/common/auth"
	"admin/internal/model"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"
	"wallet/common-lib/zapx"

	"wallet/common-lib/dbs"
	"wallet/common-lib/rdb"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

const cacheKeyPrefix = "admin.perms:"
const cacheExpireSeconds = 3600 * time.Second

func CheckPerms(ctx context.Context, uid int64, code auth.PermCode) bool {
	perms, err := UserPerms(ctx, uid)
	if err != nil {
		zapx.ErrorCtx(ctx, "userPerms error", zap.Error(err))
		return false
	}
	for _, v := range perms {
		if v == string(code) {
			return true
		}
	}
	return false
}

func UserPerms(ctx context.Context, uid int64) ([]string, error) {
	cacheKey := fmt.Sprintf("%s%d", cacheKeyPrefix, uid)

	cached, err := rdb.Client.SMembers(ctx, cacheKey).Result()
	if err == nil && len(cached) > 0 {
		return cached, nil
	}

	var userPerm model.AdminPerm
	err = userPerm.GetByUID(ctx, dbs.Admin, uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return []string{}, nil
		}
		return nil, err
	}

	var perms []string
	if err := json.Unmarshal(userPerm.Perms, &perms); err != nil {
		return nil, err
	}

	for _, perm := range perms {
		if err = rdb.Client.SAdd(ctx, cacheKey, perm).Err(); err != nil {
			zapx.ErrorCtx(ctx, "sAdd perm cache error", zap.Error(err))
		}
	}
	rdb.Client.Expire(ctx, cacheKey, cacheExpireSeconds)

	return perms, nil
}

func InvalidateCache(ctx context.Context, uid int64) error {
	cacheKey := fmt.Sprintf("%s%d", cacheKeyPrefix, uid)
	return rdb.Client.Del(ctx, cacheKey).Err()
}

func UpdateUserPermissions(ctx context.Context, currentUID, targetUID int64, reqPerms []string) error {
	if len(reqPerms) == 0 {
		return errors.New("permissions cannot be empty")
	}

	for _, perm := range reqPerms {
		if !auth.IsValidPerm(auth.PermCode(perm)) {
			return fmt.Errorf("invalid permission: %s", perm)
		}
	}

	parentPerms, err := UserPerms(ctx, currentUID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("failed to check current user permissions: %w", err)
	}

	permMap := make(map[string]bool)
	for _, p := range parentPerms {
		permMap[p] = true
	}

	for _, reqPerm := range reqPerms {
		if !permMap[reqPerm] {
			return fmt.Errorf("insufficient permission to grant: %s", reqPerm)
		}
	}

	permsJSON, err := json.Marshal(reqPerms)
	if err != nil {
		return fmt.Errorf("failed to marshal permissions: %w", err)
	}

	userPerm := &model.AdminPerm{}
	if err := dbs.Admin.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return userPerm.CreateOrUpdate(ctx, tx, targetUID, permsJSON)
	}); err != nil {
		return fmt.Errorf("failed to update permissions: %w", err)
	}

	if err := InvalidateCache(ctx, targetUID); err != nil {
		zapx.ErrorCtx(ctx, "failed to invalidate cache", zap.Error(err))
	}

	return nil
}
