package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/redis/go-redis/v9"
)

const (
	stickySessionPrefix = "sticky_session:"
	tierStickyPrefix    = "tier_sticky:"
)

type gatewayCache struct {
	rdb *redis.Client
}

func NewGatewayCache(rdb *redis.Client) service.GatewayCache {
	return &gatewayCache{rdb: rdb}
}

// buildSessionKey 构建 session key，包含 groupID 实现分组隔离
// 格式: sticky_session:{groupID}:{sessionHash}
func buildSessionKey(groupID int64, sessionHash string) string {
	return fmt.Sprintf("%s%d:%s", stickySessionPrefix, groupID, sessionHash)
}

func (c *gatewayCache) GetSessionAccountID(ctx context.Context, groupID int64, sessionHash string) (int64, error) {
	key := buildSessionKey(groupID, sessionHash)
	return c.rdb.Get(ctx, key).Int64()
}

func (c *gatewayCache) SetSessionAccountID(ctx context.Context, groupID int64, sessionHash string, accountID int64, ttl time.Duration) error {
	key := buildSessionKey(groupID, sessionHash)
	return c.rdb.Set(ctx, key, accountID, ttl).Err()
}

func (c *gatewayCache) RefreshSessionTTL(ctx context.Context, groupID int64, sessionHash string, ttl time.Duration) error {
	key := buildSessionKey(groupID, sessionHash)
	return c.rdb.Expire(ctx, key, ttl).Err()
}

// DeleteSessionAccountID 删除粘性会话与账号的绑定关系。
// 当检测到绑定的账号不可用（如状态错误、禁用、不可调度等）时调用，
// 以便下次请求能够重新选择可用账号。
//
// DeleteSessionAccountID removes the sticky session binding for the given session.
// Called when the bound account becomes unavailable (e.g., error status, disabled,
// or unschedulable), allowing subsequent requests to select a new available account.
func (c *gatewayCache) DeleteSessionAccountID(ctx context.Context, groupID int64, sessionHash string) error {
	key := buildSessionKey(groupID, sessionHash)
	return c.rdb.Del(ctx, key).Err()
}

// buildTierStickyKey 构建 tier 粘性 key：tier_sticky:{apiKeyID}:{scope}
// scope 当前按请求模型隔离，避免一个模型的降级影响其他模型。
func buildTierStickyKey(apiKeyID int64, scope string) string {
	return fmt.Sprintf("%s%d:%s", tierStickyPrefix, apiKeyID, scope)
}

// GetTierStickyGroupID 获取 (api_key, scope) 当前激活的 tier group ID。
// 未命中时 redis 返回 redis.Nil err，调用方应据此判断需要从主 group 起跳。
func (c *gatewayCache) GetTierStickyGroupID(ctx context.Context, apiKeyID int64, scope string) (int64, error) {
	return c.rdb.Get(ctx, buildTierStickyKey(apiKeyID, scope)).Int64()
}

// SetTierStickyGroupID 写入或覆盖 tier 粘性。使用普通 SET（非 SETNX）以允许后续切换到更深 tier。
func (c *gatewayCache) SetTierStickyGroupID(ctx context.Context, apiKeyID int64, scope string, groupID int64, ttl time.Duration) error {
	if ttl <= 0 {
		return nil
	}
	return c.rdb.Set(ctx, buildTierStickyKey(apiKeyID, scope), groupID, ttl).Err()
}

// RefreshTierStickyTTL 刷新 tier 粘性的过期时间，每次成功命中已粘性的 tier 时调用。
func (c *gatewayCache) RefreshTierStickyTTL(ctx context.Context, apiKeyID int64, scope string, ttl time.Duration) error {
	if ttl <= 0 {
		return nil
	}
	return c.rdb.Expire(ctx, buildTierStickyKey(apiKeyID, scope), ttl).Err()
}

// DeleteTierStickyGroupID 删除 tier 粘性。当粘性指向的 group 已失效（删除/禁用/不在链路）时调用。
func (c *gatewayCache) DeleteTierStickyGroupID(ctx context.Context, apiKeyID int64, scope string) error {
	return c.rdb.Del(ctx, buildTierStickyKey(apiKeyID, scope)).Err()
}
