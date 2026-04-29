package service

import (
	"context"
	"errors"
	"strings"
)

type usableTierCandidate struct {
	candidateID int64
	resolvedID  int64
	group       *Group
}

// ErrTierExhausted tier 链路无可用候选时返回。
// 调用方应保留 FailoverExhausted 的原有处理（502 / handleFailoverExhausted）。
var ErrTierExhausted = errors.New("tier fallback chain exhausted")

// resolveTierCandidates 按优先级解析 tier 候选列表（不含主 group，不含已 visited）。
// 语义上分两层：
//  1. 显式 tier 链路来源三选一：apiKey.TierGroupIDs > apiKey.User.DefaultTierGroupIDs > system default
//  2. group.tier_fallback_group_id 单指针递归作为最终兜底
//
// 这样可以同时满足两点：
//   - 保留 key/user/system 三层之间的覆盖关系，避免显式 per-key 链路意外“继承”更低优先级默认链路
//   - 当当前显式来源在运行时全部不可用（平台不匹配、权限撤销、CCO 解析失败等）时，
//     仍然可以继续落到 group 级单指针链路，而不会过早判定为 exhausted。
//
// 注：候选列表只构建一次（按 apiKey 缓存层），后续 tier 推进通过 visited 集合裁剪。
// ACL 校验（专属分组权限）在 ResolveNextTier 中对每个候选的解析后 group 执行。
func (s *GatewayService) resolveTierCandidates(ctx context.Context, apiKey *APIKey) []int64 {
	if apiKey == nil {
		return nil
	}

	out := make([]int64, 0, 8)
	seen := make(map[int64]struct{}, 8)
	appendUnique := func(ids []int64) {
		for _, id := range ids {
			if id <= 0 {
				continue
			}
			if _, ok := seen[id]; ok {
				continue
			}
			seen[id] = struct{}{}
			out = append(out, id)
		}
	}

	// 1. 显式 tier 链路来源三选一
	if len(apiKey.TierGroupIDs) > 0 {
		appendUnique(apiKey.TierGroupIDs)
	} else if apiKey.User != nil && len(apiKey.User.DefaultTierGroupIDs) > 0 {
		appendUnique(apiKey.User.DefaultTierGroupIDs)
	} else if s.settingService != nil {
		if ids, _ := s.settingService.GetSystemDefaultTierGroupIDs(ctx); len(ids) > 0 {
			appendUnique(ids)
		}
	}
	// 2. group 链路：从主 group 起沿 tier_fallback_group_id 递归，作为最终兜底
	if apiKey.GroupID != nil && *apiKey.GroupID > 0 {
		appendUnique(s.walkGroupTierChain(ctx, *apiKey.GroupID))
	}
	return out
}

// walkGroupTierChain 沿 group.tier_fallback_group_id 单指针递归，返回有序链路（不含起点）。
// 带环检测（visited map），避免链路成环导致死循环。
func (s *GatewayService) walkGroupTierChain(ctx context.Context, startGroupID int64) []int64 {
	visited := map[int64]struct{}{startGroupID: {}}
	chain := make([]int64, 0, 4)

	currentID := startGroupID
	for i := 0; i < 16; i++ { // 硬上限 16 跳，防止配置错误导致超长链路
		grp, err := s.resolveGroupByID(ctx, currentID)
		if err != nil || grp == nil {
			break
		}
		if grp.TierFallbackGroupID == nil || *grp.TierFallbackGroupID <= 0 {
			break
		}
		next := *grp.TierFallbackGroupID
		if _, seen := visited[next]; seen {
			break
		}
		visited[next] = struct{}{}
		chain = append(chain, next)
		currentID = next
	}
	return chain
}

// ResolveNextTier 给定当前已访问的 group 列表（visited），返回下一档可用 group。
// 返回值 (*Group, *int64, error)：
//   - 成功：解析后的 group + resolved group ID（可能因 CCO 链路与原始 ID 不同）+ nil
//   - 链路耗尽：nil, nil, ErrTierExhausted
//   - 其他错误：nil, nil, err
//
// depth 是已经发生的 tier 切换次数（首次 swap 时传 0）。当 apiKey.MaxTierDepth > 0
// 且 depth >= MaxTierDepth 时返回 ErrTierExhausted（硬阀门）。
func (s *GatewayService) ResolveNextTier(
	ctx context.Context,
	apiKey *APIKey,
	visited []int64,
	depth int,
) (*Group, *int64, error) {
	if apiKey == nil {
		return nil, nil, ErrTierExhausted
	}
	if apiKey.MaxTierDepth > 0 && depth >= apiKey.MaxTierDepth {
		return nil, nil, ErrTierExhausted
	}

	candidates := s.resolveUsableTierCandidates(ctx, apiKey)
	if len(candidates) == 0 {
		return nil, nil, ErrTierExhausted
	}

	primaryID := int64(0)
	if apiKey.GroupID != nil {
		primaryID = *apiKey.GroupID
	}

	visitedSet := make(map[int64]struct{}, len(visited)+1)
	for _, id := range visited {
		visitedSet[id] = struct{}{}
	}
	if primaryID > 0 {
		visitedSet[primaryID] = struct{}{}
	}

	for _, candidate := range candidates {
		if _, seen := visitedSet[candidate.candidateID]; seen {
			continue
		}
		// 解析后的 group 也要避免与已 visited 重合（CCO 链可能解析回已用过的 group）
		if candidate.resolvedID > 0 {
			if _, seen := visitedSet[candidate.resolvedID]; seen {
				continue
			}
		}
		resolvedID := candidate.resolvedID
		return candidate.group, &resolvedID, nil
	}
	return nil, nil, ErrTierExhausted
}

func (s *GatewayService) resolveUsableTierCandidates(ctx context.Context, apiKey *APIKey) []usableTierCandidate {
	if apiKey == nil {
		return nil
	}

	rawCandidates := s.resolveTierCandidates(ctx, apiKey)
	if len(rawCandidates) == 0 {
		return nil
	}

	primaryPlatform := s.apiKeyPrimaryPlatform(ctx, apiKey)
	out := make([]usableTierCandidate, 0, len(rawCandidates))
	seenResolved := make(map[int64]struct{}, len(rawCandidates))
	for _, candidateID := range rawCandidates {
		candidate, ok := s.resolveUsableTierCandidate(ctx, apiKey, primaryPlatform, candidateID)
		if !ok {
			continue
		}
		if _, dup := seenResolved[candidate.resolvedID]; dup {
			continue
		}
		seenResolved[candidate.resolvedID] = struct{}{}
		out = append(out, candidate)
	}
	return out
}

func (s *GatewayService) resolveUsableTierCandidate(ctx context.Context, apiKey *APIKey, primaryPlatform string, candidateID int64) (usableTierCandidate, bool) {
	if apiKey == nil || apiKey.User == nil || candidateID <= 0 {
		return usableTierCandidate{}, false
	}

	gid := candidateID
	// 复用现有 resolveGatewayGroup：自动叠加 CCO 链路（claude_code_only + fallback_group_id）。
	grp, resolvedID, err := s.resolveGatewayGroup(ctx, &gid)
	if err != nil || grp == nil || !grp.IsActive() {
		return usableTierCandidate{}, false
	}
	if primaryPlatform != "" && grp.Platform != primaryPlatform {
		return usableTierCandidate{}, false
	}

	// 运行时准入校验需要与 API key 绑定规则保持一致：
	// - subscription group：必须存在有效订阅
	// - exclusive standard group：AllowedGroups 必须包含该 group
	if grp.IsSubscriptionType() {
		if s.userSubRepo == nil {
			return usableTierCandidate{}, false
		}
		if _, err := s.userSubRepo.GetActiveByUserIDAndGroupID(ctx, apiKey.User.ID, grp.ID); err != nil {
			return usableTierCandidate{}, false
		}
	} else if grp.IsExclusive && !apiKey.User.CanBindGroup(grp.ID, true) {
		return usableTierCandidate{}, false
	}

	actualResolvedID := grp.ID
	if resolvedID != nil && *resolvedID > 0 {
		actualResolvedID = *resolvedID
	}
	return usableTierCandidate{
		candidateID: candidateID,
		resolvedID:  actualResolvedID,
		group:       grp,
	}, true
}

func (s *GatewayService) apiKeyPrimaryPlatform(ctx context.Context, apiKey *APIKey) string {
	if apiKey == nil {
		return ""
	}
	if apiKey.Group != nil {
		return strings.TrimSpace(apiKey.Group.Platform)
	}
	if apiKey.GroupID == nil || *apiKey.GroupID <= 0 {
		return ""
	}
	group, err := s.resolveGroupByID(ctx, *apiKey.GroupID)
	if err != nil || group == nil {
		return ""
	}
	return strings.TrimSpace(group.Platform)
}

func normalizeTierStickyScope(scope string) string {
	scope = strings.TrimSpace(strings.ToLower(scope))
	if scope == "" {
		return ""
	}
	return scope
}
