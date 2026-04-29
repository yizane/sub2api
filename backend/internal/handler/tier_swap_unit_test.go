package handler

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestResolveTierStickyScopePrefersMappedModel(t *testing.T) {
	got := resolveTierStickyScope("gpt-4 alias", service.ChannelMappingResult{
		Mapped:      true,
		MappedModel: "gpt-4.1",
	})
	require.Equal(t, "gpt-4.1", got)
}

func TestResolveTierStickyScopeFallsBackToRequestedModel(t *testing.T) {
	got := resolveTierStickyScope(" claude-sonnet-4 ", service.ChannelMappingResult{})
	require.Equal(t, "claude-sonnet-4", got)
}

type tierTestGroupRepo struct {
	groups map[int64]*service.Group
}

func (r *tierTestGroupRepo) Create(context.Context, *service.Group) error {
	panic("unexpected Create call")
}
func (r *tierTestGroupRepo) Update(context.Context, *service.Group) error {
	panic("unexpected Update call")
}
func (r *tierTestGroupRepo) Delete(context.Context, int64) error { panic("unexpected Delete call") }
func (r *tierTestGroupRepo) DeleteCascade(context.Context, int64) ([]int64, error) {
	panic("unexpected DeleteCascade call")
}
func (r *tierTestGroupRepo) List(context.Context, pagination.PaginationParams) ([]service.Group, *pagination.PaginationResult, error) {
	panic("unexpected List call")
}
func (r *tierTestGroupRepo) ListWithFilters(context.Context, pagination.PaginationParams, string, string, string, *bool) ([]service.Group, *pagination.PaginationResult, error) {
	panic("unexpected ListWithFilters call")
}
func (r *tierTestGroupRepo) ListActive(context.Context) ([]service.Group, error) {
	panic("unexpected ListActive call")
}
func (r *tierTestGroupRepo) ListActiveByPlatform(context.Context, string) ([]service.Group, error) {
	panic("unexpected ListActiveByPlatform call")
}
func (r *tierTestGroupRepo) ExistsByName(context.Context, string) (bool, error) {
	panic("unexpected ExistsByName call")
}
func (r *tierTestGroupRepo) GetAccountCount(context.Context, int64) (int64, int64, error) {
	panic("unexpected GetAccountCount call")
}
func (r *tierTestGroupRepo) DeleteAccountGroupsByGroupID(context.Context, int64) (int64, error) {
	panic("unexpected DeleteAccountGroupsByGroupID call")
}
func (r *tierTestGroupRepo) GetAccountIDsByGroupIDs(context.Context, []int64) ([]int64, error) {
	panic("unexpected GetAccountIDsByGroupIDs call")
}
func (r *tierTestGroupRepo) BindAccountsToGroup(context.Context, int64, []int64) error {
	panic("unexpected BindAccountsToGroup call")
}
func (r *tierTestGroupRepo) UpdateSortOrders(context.Context, []service.GroupSortOrderUpdate) error {
	panic("unexpected UpdateSortOrders call")
}

func (r *tierTestGroupRepo) GetByID(_ context.Context, id int64) (*service.Group, error) {
	group, ok := r.groups[id]
	if !ok {
		return nil, service.ErrGroupNotFound
	}
	cloned := *group
	return &cloned, nil
}

func (r *tierTestGroupRepo) GetByIDLite(ctx context.Context, id int64) (*service.Group, error) {
	return r.GetByID(ctx, id)
}

type tierTestGatewayCache struct {
	tierSticky        map[string]int64
	deletedTierKeys   []string
	refreshedTierKeys []string
}

type tierTestSchedulerCache struct {
	buckets map[string][]*service.Account
}

type tierTestConcurrencyCache struct{}

func (c *tierTestGatewayCache) GetSessionAccountID(context.Context, int64, string) (int64, error) {
	return 0, errors.New("not found")
}
func (c *tierTestGatewayCache) SetSessionAccountID(context.Context, int64, string, int64, time.Duration) error {
	return nil
}
func (c *tierTestGatewayCache) RefreshSessionTTL(context.Context, int64, string, time.Duration) error {
	return nil
}
func (c *tierTestGatewayCache) DeleteSessionAccountID(context.Context, int64, string) error {
	return nil
}
func (c *tierTestGatewayCache) GetTierStickyGroupID(_ context.Context, apiKeyID int64, scope string) (int64, error) {
	if c.tierSticky == nil {
		return 0, errors.New("not found")
	}
	id, ok := c.tierSticky[tierTestStickyKey(apiKeyID, scope)]
	if !ok {
		return 0, errors.New("not found")
	}
	return id, nil
}
func (c *tierTestGatewayCache) SetTierStickyGroupID(_ context.Context, apiKeyID int64, scope string, groupID int64, _ time.Duration) error {
	if c.tierSticky == nil {
		c.tierSticky = make(map[string]int64)
	}
	c.tierSticky[tierTestStickyKey(apiKeyID, scope)] = groupID
	return nil
}
func (c *tierTestGatewayCache) RefreshTierStickyTTL(_ context.Context, apiKeyID int64, scope string, _ time.Duration) error {
	c.refreshedTierKeys = append(c.refreshedTierKeys, tierTestStickyKey(apiKeyID, scope))
	return nil
}
func (c *tierTestGatewayCache) DeleteTierStickyGroupID(_ context.Context, apiKeyID int64, scope string) error {
	key := tierTestStickyKey(apiKeyID, scope)
	c.deletedTierKeys = append(c.deletedTierKeys, key)
	delete(c.tierSticky, key)
	return nil
}

func (c *tierTestSchedulerCache) GetSnapshot(_ context.Context, bucket service.SchedulerBucket) ([]*service.Account, bool, error) {
	if c.buckets == nil {
		return nil, true, nil
	}
	accounts := c.buckets[bucket.String()]
	return accounts, true, nil
}

func (c *tierTestSchedulerCache) SetSnapshot(context.Context, service.SchedulerBucket, []service.Account) error {
	return nil
}

func (c *tierTestSchedulerCache) GetAccount(_ context.Context, accountID int64) (*service.Account, error) {
	for _, accounts := range c.buckets {
		for _, account := range accounts {
			if account != nil && account.ID == accountID {
				cloned := *account
				return &cloned, nil
			}
		}
	}
	return nil, nil
}

func (c *tierTestSchedulerCache) SetAccount(context.Context, *service.Account) error { return nil }
func (c *tierTestSchedulerCache) DeleteAccount(context.Context, int64) error         { return nil }
func (c *tierTestSchedulerCache) UpdateLastUsed(context.Context, map[int64]time.Time) error {
	return nil
}
func (c *tierTestSchedulerCache) TryLockBucket(context.Context, service.SchedulerBucket, time.Duration) (bool, error) {
	return true, nil
}
func (c *tierTestSchedulerCache) ListBuckets(context.Context) ([]service.SchedulerBucket, error) {
	return nil, nil
}
func (c *tierTestSchedulerCache) GetOutboxWatermark(context.Context) (int64, error) { return 0, nil }
func (c *tierTestSchedulerCache) SetOutboxWatermark(context.Context, int64) error   { return nil }

func (c *tierTestConcurrencyCache) AcquireAccountSlot(context.Context, int64, int, string) (bool, error) {
	return true, nil
}
func (c *tierTestConcurrencyCache) ReleaseAccountSlot(context.Context, int64, string) error {
	return nil
}
func (c *tierTestConcurrencyCache) GetAccountConcurrency(context.Context, int64) (int, error) {
	return 0, nil
}
func (c *tierTestConcurrencyCache) GetAccountConcurrencyBatch(_ context.Context, accountIDs []int64) (map[int64]int, error) {
	out := make(map[int64]int, len(accountIDs))
	for _, id := range accountIDs {
		out[id] = 0
	}
	return out, nil
}
func (c *tierTestConcurrencyCache) IncrementAccountWaitCount(context.Context, int64, int) (bool, error) {
	return true, nil
}
func (c *tierTestConcurrencyCache) DecrementAccountWaitCount(context.Context, int64) error {
	return nil
}
func (c *tierTestConcurrencyCache) GetAccountWaitingCount(context.Context, int64) (int, error) {
	return 0, nil
}
func (c *tierTestConcurrencyCache) AcquireUserSlot(context.Context, int64, int, string) (bool, error) {
	return true, nil
}
func (c *tierTestConcurrencyCache) ReleaseUserSlot(context.Context, int64, string) error { return nil }
func (c *tierTestConcurrencyCache) GetUserConcurrency(context.Context, int64) (int, error) {
	return 0, nil
}
func (c *tierTestConcurrencyCache) IncrementWaitCount(context.Context, int64, int) (bool, error) {
	return true, nil
}
func (c *tierTestConcurrencyCache) DecrementWaitCount(context.Context, int64) error { return nil }
func (c *tierTestConcurrencyCache) GetAccountsLoadBatch(context.Context, []service.AccountWithConcurrency) (map[int64]*service.AccountLoadInfo, error) {
	return map[int64]*service.AccountLoadInfo{}, nil
}
func (c *tierTestConcurrencyCache) GetUsersLoadBatch(context.Context, []service.UserWithConcurrency) (map[int64]*service.UserLoadInfo, error) {
	return map[int64]*service.UserLoadInfo{}, nil
}
func (c *tierTestConcurrencyCache) CleanupExpiredAccountSlots(context.Context, int64) error {
	return nil
}
func (c *tierTestConcurrencyCache) CleanupStaleProcessSlots(context.Context, string) error {
	return nil
}
func (c *tierTestConcurrencyCache) CleanupAllExpiredSlots(context.Context) error     { return nil }
func (c *tierTestConcurrencyCache) RegisterInstance(context.Context, string) error   { return nil }
func (c *tierTestConcurrencyCache) UnregisterInstance(context.Context, string) error { return nil }
func (c *tierTestConcurrencyCache) FindDeadInstancePrefixes(context.Context, int64) ([]string, error) {
	return nil, nil
}
func (c *tierTestConcurrencyCache) RemoveDeadInstances(context.Context, []string) error {
	return nil
}
func (c *tierTestConcurrencyCache) CleanupSlotsForPrefixes(context.Context, []string) error {
	return nil
}

func tierTestStickyKey(apiKeyID int64, scope string) string {
	return fmt.Sprintf("%d:%s", apiKeyID, scope)
}

func newTierTestContext(t *testing.T) *gin.Context {
	t.Helper()
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/messages", nil)
	return c
}

func newTierTestOpenAIHandler(t *testing.T, groups map[int64]*service.Group, cache service.GatewayCache) (*OpenAIGatewayHandler, func()) {
	t.Helper()
	cfg := &config.Config{RunMode: config.RunModeSimple}
	tierSvc := service.NewGatewayService(
		nil,
		&tierTestGroupRepo{groups: groups},
		nil,
		nil,
		nil,
		nil,
		nil,
		cache,
		cfg,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
	)
	billingSvc := service.NewBillingCacheService(nil, nil, nil, nil, nil, nil, cfg)
	return &OpenAIGatewayHandler{
		tierService:         tierSvc,
		billingCacheService: billingSvc,
	}, billingSvc.Stop
}

func newTierTestOpenAIWSRouteHandler(t *testing.T, groups map[int64]*service.Group, cache service.GatewayCache, schedulerCache service.SchedulerCache) (*OpenAIGatewayHandler, func()) {
	t.Helper()
	cfg := &config.Config{RunMode: config.RunModeSimple}
	cfg.Gateway.OpenAIWS.Enabled = true
	cfg.Gateway.OpenAIWS.OAuthEnabled = true
	cfg.Gateway.OpenAIWS.APIKeyEnabled = true
	cfg.Gateway.OpenAIWS.ResponsesWebsocketsV2 = true
	cfg.Gateway.OpenAIWS.StickyResponseIDTTLSeconds = 3600
	concurrencyCache := &tierTestConcurrencyCache{}
	concurrencySvc := service.NewConcurrencyService(concurrencyCache)
	schedulerSnapshot := service.NewSchedulerSnapshotService(schedulerCache, nil, nil, &tierTestGroupRepo{groups: groups}, nil)
	openAISvc := service.NewOpenAIGatewayService(
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		cache,
		cfg,
		schedulerSnapshot,
		concurrencySvc,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
	)
	tierSvc := service.NewGatewayService(
		nil,
		&tierTestGroupRepo{groups: groups},
		nil,
		nil,
		nil,
		nil,
		nil,
		cache,
		cfg,
		nil,
		concurrencySvc,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
	)
	billingSvc := service.NewBillingCacheService(nil, nil, nil, nil, nil, nil, cfg)
	return &OpenAIGatewayHandler{
		gatewayService:      openAISvc,
		tierService:         tierSvc,
		billingCacheService: billingSvc,
	}, billingSvc.Stop
}

func TestOpenAIGatewayHandlerTryTierSwapAdvancesState(t *testing.T) {
	primaryID := int64(10)
	groups := map[int64]*service.Group{
		10: {ID: 10, Platform: service.PlatformOpenAI, Status: service.StatusActive},
		11: {ID: 11, Platform: service.PlatformOpenAI, Status: service.StatusActive},
	}
	h, cleanup := newTierTestOpenAIHandler(t, groups, nil)
	defer cleanup()

	apiKey := &service.APIKey{
		ID:           7,
		UserID:       42,
		GroupID:      &primaryID,
		TierGroupIDs: []int64{11},
		User:         &service.User{ID: 42, Status: service.StatusActive},
	}
	state := &openAITierSwapState{originalAPIKey: apiKey}

	swapped, ok := h.tryOpenAITierSwap(newTierTestContext(t), state, false)
	require.True(t, ok)
	require.NotNil(t, swapped)
	require.NotNil(t, swapped.GroupID)
	require.Equal(t, int64(11), *swapped.GroupID)
	require.Equal(t, 1, state.depth)
	require.Equal(t, []int64{10, 11}, state.visited)
	require.True(t, state.activatedNow)
	require.NotNil(t, state.activeTierGroupID)
	require.Equal(t, int64(11), *state.activeTierGroupID)
}

func TestOpenAIGatewayHandlerTryTierSwapSkipsSubscriptionTierWithoutActiveSubscription(t *testing.T) {
	primaryID := int64(10)
	groups := map[int64]*service.Group{
		10: {ID: 10, Platform: service.PlatformOpenAI, Status: service.StatusActive},
		11: {ID: 11, Platform: service.PlatformOpenAI, Status: service.StatusActive, SubscriptionType: service.SubscriptionTypeSubscription},
		12: {ID: 12, Platform: service.PlatformOpenAI, Status: service.StatusActive},
	}
	h, cleanup := newTierTestOpenAIHandler(t, groups, nil)
	defer cleanup()

	apiKey := &service.APIKey{
		ID:           7,
		UserID:       42,
		GroupID:      &primaryID,
		TierGroupIDs: []int64{11, 12},
		User:         &service.User{ID: 42, Status: service.StatusActive, Balance: 100},
	}
	state := &openAITierSwapState{originalAPIKey: apiKey}

	swapped, ok := h.tryOpenAITierSwap(newTierTestContext(t), state, false)
	require.True(t, ok)
	require.NotNil(t, swapped)
	require.NotNil(t, swapped.GroupID)
	require.Equal(t, int64(12), *swapped.GroupID)
	require.Equal(t, 1, state.depth)
	require.Equal(t, []int64{10, 12}, state.visited)
	require.True(t, state.activatedNow)
	require.NotNil(t, state.activeTierGroupID)
	require.Equal(t, int64(12), *state.activeTierGroupID)
}

func TestOpenAIGatewayHandlerRestoreTierFromStickyRestoresEligibleGroup(t *testing.T) {
	primaryID := int64(10)
	cache := &tierTestGatewayCache{
		tierSticky: map[string]int64{
			tierTestStickyKey(7, "gpt-4"): 11,
		},
	}
	groups := map[int64]*service.Group{
		10: {ID: 10, Platform: service.PlatformOpenAI, Status: service.StatusActive},
		11: {ID: 11, Platform: service.PlatformOpenAI, Status: service.StatusActive},
	}
	h, cleanup := newTierTestOpenAIHandler(t, groups, cache)
	defer cleanup()

	apiKey := &service.APIKey{
		ID:           7,
		UserID:       42,
		GroupID:      &primaryID,
		TierGroupIDs: []int64{11},
		User:         &service.User{ID: 42, Status: service.StatusActive},
		MaxTierDepth: 2,
	}
	state := &openAITierSwapState{originalAPIKey: apiKey}

	restored, ok := h.restoreOpenAITierFromSticky(newTierTestContext(t), apiKey, "gpt-4", state, zap.NewNop())
	require.True(t, ok)
	require.NotNil(t, restored)
	require.NotNil(t, restored.GroupID)
	require.Equal(t, int64(11), *restored.GroupID)
	require.Equal(t, 1, state.depth)
	require.Equal(t, []int64{10, 11}, state.visited)
	require.True(t, state.restoredFromSticky)
	require.NotNil(t, state.activeTierGroupID)
	require.Equal(t, int64(11), *state.activeTierGroupID)
	require.Empty(t, cache.deletedTierKeys)
}

func TestOpenAIGatewayHandlerRestoreTierFromStickyClearsSubscriptionTierWithoutActiveSubscription(t *testing.T) {
	primaryID := int64(10)
	cache := &tierTestGatewayCache{
		tierSticky: map[string]int64{
			tierTestStickyKey(7, "gpt-4"): 11,
		},
	}
	groups := map[int64]*service.Group{
		10: {ID: 10, Platform: service.PlatformOpenAI, Status: service.StatusActive},
		11: {ID: 11, Platform: service.PlatformOpenAI, Status: service.StatusActive, SubscriptionType: service.SubscriptionTypeSubscription},
	}
	h, cleanup := newTierTestOpenAIHandler(t, groups, cache)
	defer cleanup()

	apiKey := &service.APIKey{
		ID:           7,
		UserID:       42,
		GroupID:      &primaryID,
		TierGroupIDs: []int64{11},
		User:         &service.User{ID: 42, Status: service.StatusActive, Balance: 100},
		MaxTierDepth: 2,
	}
	state := &openAITierSwapState{originalAPIKey: apiKey}

	restored, ok := h.restoreOpenAITierFromSticky(newTierTestContext(t), apiKey, "gpt-4", state, zap.NewNop())
	require.False(t, ok)
	require.Nil(t, restored)
	require.Equal(t, []string{tierTestStickyKey(7, "gpt-4")}, cache.deletedTierKeys)
}

func TestOpenAIGatewayHandlerRestoreTierFromStickyRechecksClaudeCodeRestriction(t *testing.T) {
	primaryID := int64(10)
	claudeCodeOnlyID := int64(11)
	fallbackID := int64(12)
	cache := &tierTestGatewayCache{
		tierSticky: map[string]int64{
			tierTestStickyKey(7, "gpt-4"): claudeCodeOnlyID,
		},
	}
	groups := map[int64]*service.Group{
		10: {ID: 10, Platform: service.PlatformOpenAI, Status: service.StatusActive},
		11: {ID: 11, Platform: service.PlatformOpenAI, Status: service.StatusActive, ClaudeCodeOnly: true, FallbackGroupID: &fallbackID},
		12: {ID: 12, Platform: service.PlatformOpenAI, Status: service.StatusActive},
	}
	h, cleanup := newTierTestOpenAIHandler(t, groups, cache)
	defer cleanup()

	apiKey := &service.APIKey{
		ID:           7,
		UserID:       42,
		GroupID:      &primaryID,
		TierGroupIDs: []int64{claudeCodeOnlyID},
		User:         &service.User{ID: 42, Status: service.StatusActive},
		MaxTierDepth: 2,
	}
	state := &openAITierSwapState{originalAPIKey: apiKey}
	c := newTierTestContext(t)
	c.Request = c.Request.WithContext(service.SetClaudeCodeClient(c.Request.Context(), false))

	restored, ok := h.restoreOpenAITierFromSticky(c, apiKey, "gpt-4", state, zap.NewNop())
	require.True(t, ok)
	require.NotNil(t, restored)
	require.NotNil(t, restored.GroupID)
	require.Equal(t, fallbackID, *restored.GroupID)
	require.Equal(t, 1, state.depth)
	require.Equal(t, []int64{10, claudeCodeOnlyID, fallbackID}, state.visited)
	require.True(t, state.stickyNeedsRebind)
	require.Equal(t, []string{tierTestStickyKey(7, "gpt-4")}, cache.deletedTierKeys)
}

func TestOpenAIGatewayHandlerRestoreTierFromStickyMarksEarlierTiersVisited(t *testing.T) {
	primaryID := int64(10)
	tierOneID := int64(11)
	tierTwoID := int64(12)
	cache := &tierTestGatewayCache{
		tierSticky: map[string]int64{
			tierTestStickyKey(7, "gpt-4"): tierTwoID,
		},
	}
	groups := map[int64]*service.Group{
		10: {ID: 10, Platform: service.PlatformOpenAI, Status: service.StatusActive},
		11: {ID: 11, Platform: service.PlatformOpenAI, Status: service.StatusActive},
		12: {ID: 12, Platform: service.PlatformOpenAI, Status: service.StatusActive},
	}
	h, cleanup := newTierTestOpenAIHandler(t, groups, cache)
	defer cleanup()

	apiKey := &service.APIKey{
		ID:           7,
		UserID:       42,
		GroupID:      &primaryID,
		TierGroupIDs: []int64{tierOneID, tierTwoID},
		User:         &service.User{ID: 42, Status: service.StatusActive},
		MaxTierDepth: 3,
	}
	state := &openAITierSwapState{originalAPIKey: apiKey}

	restored, ok := h.restoreOpenAITierFromSticky(newTierTestContext(t), apiKey, "gpt-4", state, zap.NewNop())
	require.True(t, ok)
	require.NotNil(t, restored)
	require.Equal(t, 2, state.depth)
	require.Equal(t, []int64{10, 11, 12}, state.visited)
}

func TestOpenAIGatewayHandlerPersistTierStickyOnSuccessRefreshesRestoredSticky(t *testing.T) {
	cache := &tierTestGatewayCache{}
	h, cleanup := newTierTestOpenAIHandler(t, map[int64]*service.Group{}, cache)
	defer cleanup()

	activeGroupID := int64(11)
	state := &openAITierSwapState{
		originalAPIKey:     &service.APIKey{ID: 7},
		restoredFromSticky: true,
		activeTierGroupID:  &activeGroupID,
	}

	h.persistOpenAITierStickyOnSuccess(newTierTestContext(t), state, "gpt-4", zap.NewNop(), "openai")

	require.Equal(t, []string{tierTestStickyKey(7, "gpt-4")}, cache.refreshedTierKeys)
	require.Empty(t, cache.deletedTierKeys)
}

func TestOpenAIGatewayHandlerPersistTierStickyOnSuccessBindsActivatedTier(t *testing.T) {
	cache := &tierTestGatewayCache{}
	h, cleanup := newTierTestOpenAIHandler(t, map[int64]*service.Group{}, cache)
	defer cleanup()

	activeGroupID := int64(11)
	state := &openAITierSwapState{
		originalAPIKey:     &service.APIKey{ID: 7},
		restoredFromSticky: true,
		activatedNow:       true,
		activeTierGroupID:  &activeGroupID,
	}

	h.persistOpenAITierStickyOnSuccess(newTierTestContext(t), state, "gpt-4", zap.NewNop(), "openai")

	require.Equal(t, int64(11), cache.tierSticky[tierTestStickyKey(7, "gpt-4")])
	require.Empty(t, cache.refreshedTierKeys)
}

func TestOpenAIGatewayHandlerPersistTierStickyOnSuccessRebindsResolvedStickyTarget(t *testing.T) {
	cache := &tierTestGatewayCache{}
	h, cleanup := newTierTestOpenAIHandler(t, map[int64]*service.Group{}, cache)
	defer cleanup()

	activeGroupID := int64(12)
	state := &openAITierSwapState{
		originalAPIKey:     &service.APIKey{ID: 7},
		restoredFromSticky: true,
		stickyNeedsRebind:  true,
		activeTierGroupID:  &activeGroupID,
	}

	h.persistOpenAITierStickyOnSuccess(newTierTestContext(t), state, "gpt-4", zap.NewNop(), "openai")

	require.Equal(t, int64(12), cache.tierSticky[tierTestStickyKey(7, "gpt-4")])
	require.Empty(t, cache.refreshedTierKeys)
}

func TestOpenAIGatewayHandlerSelectOpenAIWSInitialRouteFallsBackToTierGroup(t *testing.T) {
	primaryID := int64(10)
	fallbackID := int64(11)
	accountID := int64(1001)
	groups := map[int64]*service.Group{
		primaryID:  {ID: primaryID, Platform: service.PlatformOpenAI, Status: service.StatusActive},
		fallbackID: {ID: fallbackID, Platform: service.PlatformOpenAI, Status: service.StatusActive},
	}

	schedulerCache := &tierTestSchedulerCache{
		buckets: map[string][]*service.Account{},
	}
	primaryBucket := service.SchedulerBucket{GroupID: primaryID, Platform: service.PlatformOpenAI, Mode: service.SchedulerModeSingle}
	fallbackBucket := service.SchedulerBucket{GroupID: fallbackID, Platform: service.PlatformOpenAI, Mode: service.SchedulerModeSingle}
	primaryMixedBucket := service.SchedulerBucket{GroupID: primaryID, Platform: service.PlatformOpenAI, Mode: service.SchedulerModeMixed}
	fallbackMixedBucket := service.SchedulerBucket{GroupID: fallbackID, Platform: service.PlatformOpenAI, Mode: service.SchedulerModeMixed}
	schedulerCache.buckets[primaryBucket.String()] = nil
	schedulerCache.buckets[primaryMixedBucket.String()] = nil
	schedulerCache.buckets[fallbackBucket.String()] = []*service.Account{
		{
			ID:          accountID,
			Name:        "fallback-openai",
			Platform:    service.PlatformOpenAI,
			Type:        service.AccountTypeOAuth,
			Credentials: map[string]any{"access_token": "tok_xxx"},
			Extra: map[string]any{
				"openai_oauth_responses_websockets_v2_enabled": true,
			},
			Concurrency: 1,
			Priority:    1,
			Status:      service.StatusActive,
			Schedulable: true,
			AccountGroups: []service.AccountGroup{
				{AccountID: accountID, GroupID: fallbackID},
			},
		},
	}
	schedulerCache.buckets[fallbackMixedBucket.String()] = schedulerCache.buckets[fallbackBucket.String()]

	h, cleanup := newTierTestOpenAIWSRouteHandler(t, groups, nil, schedulerCache)
	defer cleanup()

	c := newTierTestContext(t)
	apiKey := &service.APIKey{
		ID:           7,
		UserID:       42,
		GroupID:      &primaryID,
		Status:       service.StatusActive,
		TierGroupIDs: []int64{fallbackID},
		User: &service.User{
			ID:          42,
			Status:      service.StatusActive,
			Concurrency: 10,
			Balance:     100,
		},
		Group: groups[primaryID],
	}
	firstMessage := []byte(`{"type":"response.create","model":"gpt-5.1","stream":false}`)

	currentAPIKey, currentSubscription, channelMapping, tierState, sessionHash, selection, _, billingErr, err := h.selectOpenAIWSInitialRoute(
		c,
		apiKey,
		nil,
		"gpt-5.1",
		"",
		firstMessage,
		zap.NewNop(),
	)

	require.NoError(t, billingErr)
	require.NoError(t, err)
	require.NotNil(t, currentAPIKey)
	require.NotNil(t, currentAPIKey.GroupID)
	require.Equal(t, fallbackID, *currentAPIKey.GroupID)
	require.Nil(t, currentSubscription)
	require.False(t, channelMapping.Mapped)
	require.NotNil(t, selection)
	require.NotNil(t, selection.Account)
	require.Equal(t, accountID, selection.Account.ID)
	require.NotEmpty(t, sessionHash)
	require.NotNil(t, tierState)
	require.Equal(t, 1, tierState.depth)
	require.Equal(t, []int64{primaryID, fallbackID}, tierState.visited)
}

func TestCloneAPIKeyWithGroupClearsUserGroupRPMOverride(t *testing.T) {
	primaryID := int64(10)
	override := 77
	apiKey := &service.APIKey{
		ID:      1,
		UserID:  42,
		GroupID: &primaryID,
		User: &service.User{
			ID:                   42,
			Status:               service.StatusActive,
			UserGroupRPMOverride: &override,
		},
		Group: &service.Group{ID: primaryID, Platform: service.PlatformOpenAI, Status: service.StatusActive},
	}
	newGroup := &service.Group{ID: 12, Platform: service.PlatformOpenAI, Status: service.StatusActive}

	cloned := cloneAPIKeyWithGroup(apiKey, newGroup)
	require.NotNil(t, cloned)
	require.NotNil(t, cloned.User)
	require.Nil(t, cloned.User.UserGroupRPMOverride)
	require.NotNil(t, apiKey.User.UserGroupRPMOverride)
	require.Equal(t, 77, *apiKey.User.UserGroupRPMOverride)
	require.NotSame(t, apiKey.User, cloned.User)
	require.NotNil(t, cloned.GroupID)
	require.Equal(t, int64(12), *cloned.GroupID)
}
