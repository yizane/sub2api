package service

import (
	"context"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/stretchr/testify/require"
)

type tierFallbackGroupRepoStub struct {
	groups map[int64]*Group
}

func (s *tierFallbackGroupRepoStub) Create(context.Context, *Group) error {
	panic("unexpected Create call")
}

func (s *tierFallbackGroupRepoStub) GetByID(_ context.Context, id int64) (*Group, error) {
	if group, ok := s.groups[id]; ok {
		clone := *group
		return &clone, nil
	}
	return nil, ErrGroupNotFound
}

func (s *tierFallbackGroupRepoStub) GetByIDLite(ctx context.Context, id int64) (*Group, error) {
	return s.GetByID(ctx, id)
}

func (s *tierFallbackGroupRepoStub) Update(context.Context, *Group) error {
	panic("unexpected Update call")
}
func (s *tierFallbackGroupRepoStub) Delete(context.Context, int64) error {
	panic("unexpected Delete call")
}

func (s *tierFallbackGroupRepoStub) DeleteCascade(context.Context, int64) ([]int64, error) {
	panic("unexpected DeleteCascade call")
}

func (s *tierFallbackGroupRepoStub) List(context.Context, pagination.PaginationParams) ([]Group, *pagination.PaginationResult, error) {
	panic("unexpected List call")
}

func (s *tierFallbackGroupRepoStub) ListWithFilters(context.Context, pagination.PaginationParams, string, string, string, *bool) ([]Group, *pagination.PaginationResult, error) {
	panic("unexpected ListWithFilters call")
}

func (s *tierFallbackGroupRepoStub) ListActive(context.Context) ([]Group, error) {
	panic("unexpected ListActive call")
}

func (s *tierFallbackGroupRepoStub) ListActiveByPlatform(context.Context, string) ([]Group, error) {
	panic("unexpected ListActiveByPlatform call")
}

func (s *tierFallbackGroupRepoStub) ExistsByName(context.Context, string) (bool, error) {
	panic("unexpected ExistsByName call")
}

func (s *tierFallbackGroupRepoStub) GetAccountCount(context.Context, int64) (int64, int64, error) {
	panic("unexpected GetAccountCount call")
}

func (s *tierFallbackGroupRepoStub) DeleteAccountGroupsByGroupID(context.Context, int64) (int64, error) {
	panic("unexpected DeleteAccountGroupsByGroupID call")
}

func (s *tierFallbackGroupRepoStub) GetAccountIDsByGroupIDs(context.Context, []int64) ([]int64, error) {
	panic("unexpected GetAccountIDsByGroupIDs call")
}

func (s *tierFallbackGroupRepoStub) BindAccountsToGroup(context.Context, int64, []int64) error {
	panic("unexpected BindAccountsToGroup call")
}

func (s *tierFallbackGroupRepoStub) UpdateSortOrders(context.Context, []GroupSortOrderUpdate) error {
	panic("unexpected UpdateSortOrders call")
}

type tierFallbackUserRepoStub struct {
	user *User
}

type userSubRepoStub struct {
	userSubRepoNoop
	getActiveByUserGroupFn func(context.Context, int64, int64) (*UserSubscription, error)
}

func (s *userSubRepoStub) GetActiveByUserIDAndGroupID(ctx context.Context, userID, groupID int64) (*UserSubscription, error) {
	if s.getActiveByUserGroupFn == nil {
		return nil, ErrSubscriptionNotFound
	}
	return s.getActiveByUserGroupFn(ctx, userID, groupID)
}

func (s *tierFallbackUserRepoStub) Create(context.Context, *User) error {
	panic("unexpected Create call")
}

func (s *tierFallbackUserRepoStub) GetByID(context.Context, int64) (*User, error) {
	clone := *s.user
	return &clone, nil
}

func (s *tierFallbackUserRepoStub) GetByEmail(context.Context, string) (*User, error) {
	panic("unexpected GetByEmail call")
}

func (s *tierFallbackUserRepoStub) GetFirstAdmin(context.Context) (*User, error) {
	panic("unexpected GetFirstAdmin call")
}

func (s *tierFallbackUserRepoStub) Update(context.Context, *User) error {
	panic("unexpected Update call")
}
func (s *tierFallbackUserRepoStub) Delete(context.Context, int64) error {
	panic("unexpected Delete call")
}

func (s *tierFallbackUserRepoStub) GetUserAvatar(context.Context, int64) (*UserAvatar, error) {
	panic("unexpected GetUserAvatar call")
}

func (s *tierFallbackUserRepoStub) UpsertUserAvatar(context.Context, int64, UpsertUserAvatarInput) (*UserAvatar, error) {
	panic("unexpected UpsertUserAvatar call")
}

func (s *tierFallbackUserRepoStub) DeleteUserAvatar(context.Context, int64) error {
	panic("unexpected DeleteUserAvatar call")
}

func (s *tierFallbackUserRepoStub) List(context.Context, pagination.PaginationParams) ([]User, *pagination.PaginationResult, error) {
	panic("unexpected List call")
}

func (s *tierFallbackUserRepoStub) ListWithFilters(context.Context, pagination.PaginationParams, UserListFilters) ([]User, *pagination.PaginationResult, error) {
	panic("unexpected ListWithFilters call")
}

func (s *tierFallbackUserRepoStub) GetLatestUsedAtByUserIDs(context.Context, []int64) (map[int64]*time.Time, error) {
	panic("unexpected GetLatestUsedAtByUserIDs call")
}

func (s *tierFallbackUserRepoStub) GetLatestUsedAtByUserID(context.Context, int64) (*time.Time, error) {
	panic("unexpected GetLatestUsedAtByUserID call")
}

func (s *tierFallbackUserRepoStub) UpdateUserLastActiveAt(context.Context, int64, time.Time) error {
	panic("unexpected UpdateUserLastActiveAt call")
}

func (s *tierFallbackUserRepoStub) UpdateBalance(context.Context, int64, float64) error {
	panic("unexpected UpdateBalance call")
}

func (s *tierFallbackUserRepoStub) DeductBalance(context.Context, int64, float64) error {
	panic("unexpected DeductBalance call")
}

func (s *tierFallbackUserRepoStub) UpdateConcurrency(context.Context, int64, int) error {
	panic("unexpected UpdateConcurrency call")
}

func (s *tierFallbackUserRepoStub) ExistsByEmail(context.Context, string) (bool, error) {
	panic("unexpected ExistsByEmail call")
}

func (s *tierFallbackUserRepoStub) RemoveGroupFromAllowedGroups(context.Context, int64) (int64, error) {
	panic("unexpected RemoveGroupFromAllowedGroups call")
}

func (s *tierFallbackUserRepoStub) AddGroupToAllowedGroups(context.Context, int64, int64) error {
	panic("unexpected AddGroupToAllowedGroups call")
}

func (s *tierFallbackUserRepoStub) RemoveGroupFromUserAllowedGroups(context.Context, int64, int64) error {
	panic("unexpected RemoveGroupFromUserAllowedGroups call")
}

func (s *tierFallbackUserRepoStub) ListUserAuthIdentities(context.Context, int64) ([]UserAuthIdentityRecord, error) {
	panic("unexpected ListUserAuthIdentities call")
}

func (s *tierFallbackUserRepoStub) UnbindUserAuthProvider(context.Context, int64, string) error {
	panic("unexpected UnbindUserAuthProvider call")
}

func (s *tierFallbackUserRepoStub) UpdateTotpSecret(context.Context, int64, *string) error {
	panic("unexpected UpdateTotpSecret call")
}

func (s *tierFallbackUserRepoStub) EnableTotp(context.Context, int64) error {
	panic("unexpected EnableTotp call")
}
func (s *tierFallbackUserRepoStub) DisableTotp(context.Context, int64) error {
	panic("unexpected DisableTotp call")
}

func tierFallbackInt64Ptr(v int64) *int64 { return &v }

type tierFallbackSettingRepoStub struct {
	value string
}

func (s *tierFallbackSettingRepoStub) Get(context.Context, string) (*Setting, error) {
	panic("unexpected Get call")
}

func (s *tierFallbackSettingRepoStub) GetValue(context.Context, string) (string, error) {
	return s.value, nil
}

func (s *tierFallbackSettingRepoStub) Set(context.Context, string, string) error {
	panic("unexpected Set call")
}

func (s *tierFallbackSettingRepoStub) GetMultiple(context.Context, []string) (map[string]string, error) {
	panic("unexpected GetMultiple call")
}

func (s *tierFallbackSettingRepoStub) SetMultiple(context.Context, map[string]string) error {
	panic("unexpected SetMultiple call")
}

func (s *tierFallbackSettingRepoStub) GetAll(context.Context) (map[string]string, error) {
	panic("unexpected GetAll call")
}

func (s *tierFallbackSettingRepoStub) Delete(context.Context, string) error {
	panic("unexpected Delete call")
}

type tierFallbackAPIKeyRepoStub struct {
	apiKey    *APIKey
	updated   *APIKey
	updateErr error
}

func (s *tierFallbackAPIKeyRepoStub) Create(context.Context, *APIKey) error {
	panic("unexpected Create call")
}

func (s *tierFallbackAPIKeyRepoStub) GetByID(context.Context, int64) (*APIKey, error) {
	clone := *s.apiKey
	return &clone, nil
}

func (s *tierFallbackAPIKeyRepoStub) GetKeyAndOwnerID(context.Context, int64) (string, int64, error) {
	panic("unexpected GetKeyAndOwnerID call")
}

func (s *tierFallbackAPIKeyRepoStub) GetByKey(context.Context, string) (*APIKey, error) {
	panic("unexpected GetByKey call")
}

func (s *tierFallbackAPIKeyRepoStub) GetByKeyForAuth(context.Context, string) (*APIKey, error) {
	panic("unexpected GetByKeyForAuth call")
}

func (s *tierFallbackAPIKeyRepoStub) Update(_ context.Context, key *APIKey) error {
	if s.updateErr != nil {
		return s.updateErr
	}
	clone := *key
	s.updated = &clone
	return nil
}

func (s *tierFallbackAPIKeyRepoStub) Delete(context.Context, int64) error {
	panic("unexpected Delete call")
}

func (s *tierFallbackAPIKeyRepoStub) ListByUserID(context.Context, int64, pagination.PaginationParams, APIKeyListFilters) ([]APIKey, *pagination.PaginationResult, error) {
	panic("unexpected ListByUserID call")
}

func (s *tierFallbackAPIKeyRepoStub) VerifyOwnership(context.Context, int64, []int64) ([]int64, error) {
	panic("unexpected VerifyOwnership call")
}

func (s *tierFallbackAPIKeyRepoStub) CountByUserID(context.Context, int64) (int64, error) {
	panic("unexpected CountByUserID call")
}

func (s *tierFallbackAPIKeyRepoStub) ExistsByKey(context.Context, string) (bool, error) {
	panic("unexpected ExistsByKey call")
}

func (s *tierFallbackAPIKeyRepoStub) ListByGroupID(context.Context, int64, pagination.PaginationParams) ([]APIKey, *pagination.PaginationResult, error) {
	panic("unexpected ListByGroupID call")
}

func (s *tierFallbackAPIKeyRepoStub) SearchAPIKeys(context.Context, int64, string, int) ([]APIKey, error) {
	panic("unexpected SearchAPIKeys call")
}

func (s *tierFallbackAPIKeyRepoStub) ClearGroupIDByGroupID(context.Context, int64) (int64, error) {
	panic("unexpected ClearGroupIDByGroupID call")
}

func (s *tierFallbackAPIKeyRepoStub) UpdateGroupIDByUserAndGroup(context.Context, int64, int64, int64) (int64, error) {
	panic("unexpected UpdateGroupIDByUserAndGroup call")
}

func (s *tierFallbackAPIKeyRepoStub) CountByGroupID(context.Context, int64) (int64, error) {
	panic("unexpected CountByGroupID call")
}

func (s *tierFallbackAPIKeyRepoStub) ListKeysByUserID(context.Context, int64) ([]string, error) {
	panic("unexpected ListKeysByUserID call")
}

func (s *tierFallbackAPIKeyRepoStub) ListKeysByGroupID(context.Context, int64) ([]string, error) {
	panic("unexpected ListKeysByGroupID call")
}

func (s *tierFallbackAPIKeyRepoStub) IncrementQuotaUsed(context.Context, int64, float64) (float64, error) {
	panic("unexpected IncrementQuotaUsed call")
}

func (s *tierFallbackAPIKeyRepoStub) UpdateLastUsed(context.Context, int64, time.Time) error {
	panic("unexpected UpdateLastUsed call")
}

func (s *tierFallbackAPIKeyRepoStub) IncrementRateLimitUsage(context.Context, int64, float64) error {
	panic("unexpected IncrementRateLimitUsage call")
}

func (s *tierFallbackAPIKeyRepoStub) ResetRateLimitWindows(context.Context, int64) error {
	panic("unexpected ResetRateLimitWindows call")
}

func (s *tierFallbackAPIKeyRepoStub) GetRateLimitData(context.Context, int64) (*APIKeyRateLimitData, error) {
	panic("unexpected GetRateLimitData call")
}

func TestGatewayServiceResolveNextTierFallsBackToGroupChainAfterUnusableSystemDefaults(t *testing.T) {
	systemRepo := &tierFallbackSettingRepoStub{value: `[99]`}
	groupRepo := &tierFallbackGroupRepoStub{
		groups: map[int64]*Group{
			1:  {ID: 1, Platform: PlatformAnthropic, Status: StatusActive, TierFallbackGroupID: tierFallbackInt64Ptr(2)},
			2:  {ID: 2, Platform: PlatformAnthropic, Status: StatusActive},
			99: {ID: 99, Platform: PlatformOpenAI, Status: StatusActive},
		},
	}
	svc := &GatewayService{
		groupRepo:      groupRepo,
		settingService: NewSettingService(systemRepo, &config.Config{}),
	}

	apiKey := &APIKey{
		GroupID: tierFallbackInt64Ptr(1),
		Group:   &Group{ID: 1, Platform: PlatformAnthropic, Status: StatusActive},
		User:    &User{ID: 7, Status: StatusActive},
	}

	group, resolvedID, err := svc.ResolveNextTier(context.Background(), apiKey, nil, 0)
	require.NoError(t, err)
	require.NotNil(t, group)
	require.Equal(t, int64(2), group.ID)
	require.NotNil(t, resolvedID)
	require.Equal(t, int64(2), *resolvedID)
}

func TestGatewayServiceResolveTierCandidatesDoesNotInheritLowerPriorityDefaults(t *testing.T) {
	systemRepo := &tierFallbackSettingRepoStub{value: `[99]`}
	svc := &GatewayService{
		settingService: NewSettingService(systemRepo, &config.Config{}),
	}

	apiKey := &APIKey{
		TierGroupIDs: []int64{11},
		User: &User{
			DefaultTierGroupIDs: []int64{22},
		},
	}

	got := svc.resolveTierCandidates(context.Background(), apiKey)
	require.Equal(t, []int64{11}, got)
}

func TestGatewayServiceResolveNextTierSkipsSubscriptionGroupWithoutActiveSubscription(t *testing.T) {
	groupRepo := &tierFallbackGroupRepoStub{
		groups: map[int64]*Group{
			1: {ID: 1, Platform: PlatformAnthropic, Status: StatusActive},
			2: {ID: 2, Platform: PlatformAnthropic, Status: StatusActive, SubscriptionType: SubscriptionTypeSubscription},
			3: {ID: 3, Platform: PlatformAnthropic, Status: StatusActive},
		},
	}
	subRepo := &userSubRepoStub{
		getActiveByUserGroupFn: func(context.Context, int64, int64) (*UserSubscription, error) {
			return nil, ErrSubscriptionNotFound
		},
	}
	svc := &GatewayService{
		groupRepo:      groupRepo,
		userSubRepo:    subRepo,
		settingService: NewSettingService(&tierFallbackSettingRepoStub{}, &config.Config{}),
	}

	apiKey := &APIKey{
		GroupID:      tierFallbackInt64Ptr(1),
		TierGroupIDs: []int64{2, 3},
		Group:        &Group{ID: 1, Platform: PlatformAnthropic, Status: StatusActive},
		User:         &User{ID: 7, Status: StatusActive},
	}

	group, resolvedID, err := svc.ResolveNextTier(context.Background(), apiKey, nil, 0)
	require.NoError(t, err)
	require.NotNil(t, group)
	require.Equal(t, int64(3), group.ID)
	require.NotNil(t, resolvedID)
	require.Equal(t, int64(3), *resolvedID)
}

func TestGatewayServiceGetTierGroupDepthIgnoresUnusableCandidates(t *testing.T) {
	systemRepo := &tierFallbackSettingRepoStub{value: `[99,2]`}
	groupRepo := &tierFallbackGroupRepoStub{
		groups: map[int64]*Group{
			1:  {ID: 1, Platform: PlatformAnthropic, Status: StatusActive},
			2:  {ID: 2, Platform: PlatformAnthropic, Status: StatusActive},
			99: {ID: 99, Platform: PlatformOpenAI, Status: StatusActive},
		},
	}
	svc := &GatewayService{
		groupRepo:      groupRepo,
		settingService: NewSettingService(systemRepo, &config.Config{}),
	}

	apiKey := &APIKey{
		GroupID: tierFallbackInt64Ptr(1),
		Group:   &Group{ID: 1, Platform: PlatformAnthropic, Status: StatusActive},
		User:    &User{ID: 7, Status: StatusActive},
	}

	depth := svc.GetTierGroupDepth(context.Background(), apiKey, 2)
	require.Equal(t, 1, depth)
}

func TestGatewayServiceResolveTierVisitedPrefixUsesUsableDepth(t *testing.T) {
	systemRepo := &tierFallbackSettingRepoStub{value: `[99,2,3]`}
	groupRepo := &tierFallbackGroupRepoStub{
		groups: map[int64]*Group{
			1:  {ID: 1, Platform: PlatformAnthropic, Status: StatusActive},
			2:  {ID: 2, Platform: PlatformAnthropic, Status: StatusActive},
			3:  {ID: 3, Platform: PlatformAnthropic, Status: StatusActive},
			99: {ID: 99, Platform: PlatformOpenAI, Status: StatusActive},
		},
	}
	svc := &GatewayService{
		groupRepo:      groupRepo,
		settingService: NewSettingService(systemRepo, &config.Config{}),
	}

	apiKey := &APIKey{
		GroupID: tierFallbackInt64Ptr(1),
		Group:   &Group{ID: 1, Platform: PlatformAnthropic, Status: StatusActive},
		User:    &User{ID: 7, Status: StatusActive},
	}

	visited := svc.ResolveTierVisitedPrefix(context.Background(), apiKey, 1)
	require.Equal(t, []int64{1, 2}, visited)
}

func TestAPIKeyServiceUpdateRejectsStaleTierChainWhenPrimaryGroupChanges(t *testing.T) {
	oldPrimary := int64(10)
	newPrimary := int64(20)

	apiKeyRepo := &tierFallbackAPIKeyRepoStub{
		apiKey: &APIKey{
			ID:           1,
			UserID:       42,
			Key:          "sk-test",
			GroupID:      &oldPrimary,
			Status:       StatusActive,
			TierGroupIDs: []int64{11},
		},
	}
	userRepo := &tierFallbackUserRepoStub{
		user: &User{ID: 42, Status: StatusActive},
	}
	groupRepo := &tierFallbackGroupRepoStub{
		groups: map[int64]*Group{
			20: {ID: 20, Platform: PlatformOpenAI, Status: StatusActive},
			11: {ID: 11, Platform: PlatformAnthropic, Status: StatusActive},
		},
	}

	svc := NewAPIKeyService(apiKeyRepo, userRepo, groupRepo, nil, nil, nil, &config.Config{})
	updated, err := svc.Update(context.Background(), 1, 42, UpdateAPIKeyRequest{
		GroupID: &newPrimary,
	})

	require.Error(t, err)
	require.Nil(t, updated)
	require.Contains(t, err.Error(), "existing tier_group_ids are invalid for the new primary group")
	require.Nil(t, apiKeyRepo.updated)
}

func TestAPIKeyServiceValidateTierGroupIDsRejectsNonOpenAIPrimary(t *testing.T) {
	primaryID := int64(10)
	groupRepo := &tierFallbackGroupRepoStub{
		groups: map[int64]*Group{
			10: {ID: 10, Platform: PlatformAnthropic, Status: StatusActive},
			11: {ID: 11, Platform: PlatformOpenAI, Status: StatusActive},
		},
	}

	svc := &APIKeyService{groupRepo: groupRepo}
	validated, err := svc.validateTierGroupIDs(context.Background(), &User{ID: 42, Status: StatusActive}, []int64{11}, &primaryID)

	require.Error(t, err)
	require.Nil(t, validated)
	require.Contains(t, err.Error(), "tier_group_ids requires an OpenAI primary group")
}

func TestSettingServiceSetSystemDefaultTierGroupIDsRejectsNonOpenAIGroup(t *testing.T) {
	settingSvc := NewSettingService(&tierFallbackSettingRepoStub{}, &config.Config{})
	settingSvc.SetGroupRepository(&tierFallbackGroupRepoStub{
		groups: map[int64]*Group{
			99: {ID: 99, Platform: PlatformAnthropic, Status: StatusActive},
		},
	})

	err := settingSvc.SetSystemDefaultTierGroupIDs(context.Background(), []int64{99})
	require.Error(t, err)
	require.Contains(t, err.Error(), "must be an active OpenAI group")
}
