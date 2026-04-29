package handler

import (
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

func cloneAPIKeyWithGroup(apiKey *service.APIKey, group *service.Group) *service.APIKey {
	if apiKey == nil || group == nil {
		return apiKey
	}
	cloned := *apiKey
	if apiKey.User != nil {
		clonedUser := *apiKey.User
		// Per-group RPM override belongs to the original group only.
		clonedUser.UserGroupRPMOverride = nil
		cloned.User = &clonedUser
	}
	groupID := group.ID
	cloned.GroupID = &groupID
	cloned.Group = group
	return &cloned
}

func appendTierVisited(visited []int64, groupID int64, resolvedID *int64) []int64 {
	if groupID > 0 {
		visited = append(visited, groupID)
	}
	if resolvedID != nil && *resolvedID > 0 && *resolvedID != groupID {
		visited = append(visited, *resolvedID)
	}
	return visited
}

func resolveTierStickyScope(requestedModel string, channelMapping service.ChannelMappingResult) string {
	if mapped := strings.TrimSpace(channelMapping.MappedModel); mapped != "" {
		return mapped
	}
	return strings.TrimSpace(requestedModel)
}
