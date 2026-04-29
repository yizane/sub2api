package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// GetSystemDefaultTierGroupIDs 读取系统级 tier 降级链路默认值。
// 仅在 api_key.tier_group_ids 与 user.default_tier_group_ids 都为空时作为兜底。
// 该方法只在 gateway tier 解析路径上调用（FailoverExhausted 之后），不属于热路径，无需缓存。
func (s *SettingService) GetSystemDefaultTierGroupIDs(ctx context.Context) ([]int64, error) {
	if s == nil || s.settingRepo == nil {
		return nil, nil
	}
	if ctx == nil {
		ctx = context.Background()
	}

	raw, err := s.settingRepo.GetValue(ctx, SettingKeyDefaultTierGroupIDs)
	if err != nil {
		if errors.Is(err, ErrSettingNotFound) {
			return nil, nil
		}
		return nil, err
	}
	if raw == "" {
		return nil, nil
	}

	var ids []int64
	if err := json.Unmarshal([]byte(raw), &ids); err != nil {
		// 损坏的 JSON 不应阻塞 gateway，按"无系统默认"处理
		return nil, nil
	}
	return ids, nil
}

// validateAdminTierGroupIDs 校验 admin 写入的 tier 链路：去重 + 存在 + active。
// 不做用户级 ACL 校验（admin 视角下绕过用户允许范围是合理的——admin 在用户级默认链路里
// 配置一个用户尚未被允许的 group 视为待批准状态，由后续工单或 group 公开化解决）。
//
// admin_service.go 里 UpdateUser 与 group_handler.go 的 tier_fallback 校验都复用此函数。
func validateAdminTierGroupIDs(ctx context.Context, groupRepo GroupRepository, ids []int64) ([]int64, error) {
	if len(ids) == 0 {
		return []int64{}, nil
	}
	const maxLen = 16
	if len(ids) > maxLen {
		return nil, fmt.Errorf("tier group chain exceeds max length %d", maxLen)
	}
	seen := make(map[int64]struct{}, len(ids))
	out := make([]int64, 0, len(ids))
	for _, id := range ids {
		if id <= 0 {
			return nil, fmt.Errorf("invalid group id: %d", id)
		}
		if _, dup := seen[id]; dup {
			continue
		}
		seen[id] = struct{}{}
		group, err := groupRepo.GetByID(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("tier group %d: %w", id, err)
		}
		if group == nil || group.Status != StatusActive || strings.TrimSpace(group.Platform) != PlatformOpenAI {
			return nil, fmt.Errorf("tier group %d must be an active OpenAI group", id)
		}
		out = append(out, id)
	}
	return out, nil
}

// SetSystemDefaultTierGroupIDs 写入系统级 tier 降级链路默认值。
// 调用方无需预先校验；本方法通过 groupRepo 执行去重 + 存在 + active 检查。
func (s *SettingService) SetSystemDefaultTierGroupIDs(ctx context.Context, ids []int64) error {
	if s == nil || s.settingRepo == nil {
		return errors.New("setting repository not initialized")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if ids == nil {
		ids = []int64{}
	}
	// 使用 validateAdminTierGroupIDs 做去重 + 存在 + active 校验（需要 groupRepo）。
	var validated []int64
	if s.groupRepo != nil && len(ids) > 0 {
		var err error
		validated, err = validateAdminTierGroupIDs(ctx, s.groupRepo, ids)
		if err != nil {
			return err
		}
	} else {
		// groupRepo 未注入时降级为纯去重（测试场景）。
		seen := make(map[int64]struct{}, len(ids))
		validated = make([]int64, 0, len(ids))
		for _, id := range ids {
			if id <= 0 {
				return fmt.Errorf("invalid group id: %d", id)
			}
			if _, ok := seen[id]; ok {
				continue
			}
			seen[id] = struct{}{}
			validated = append(validated, id)
		}
	}
	raw, err := json.Marshal(validated)
	if err != nil {
		return err
	}
	if err := s.settingRepo.Set(ctx, SettingKeyDefaultTierGroupIDs, string(raw)); err != nil {
		return err
	}
	if s.onUpdate != nil {
		s.onUpdate()
	}
	return nil
}
