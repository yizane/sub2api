-- API Key 多级降级链（Tier Fallback Chain）
-- 1) api_keys.tier_group_ids: 每 key 独立的有序降级 group 链路（JSONB int64 数组）
-- 2) api_keys.max_tier_depth: 链路最大深度，0 = 不限制
-- 3) users.default_tier_group_ids: 用户级默认链路（api_keys.tier_group_ids 为空时兜底）
-- 4) groups.tier_fallback_group_id: 基础设施级单指针链路（最末级兜底）
--    独立于 fallback_group_id（CCO 触发）和 fallback_group_id_on_invalid_request（无效请求触发）

ALTER TABLE api_keys
    ADD COLUMN IF NOT EXISTS tier_group_ids JSONB NOT NULL DEFAULT '[]'::jsonb,
    ADD COLUMN IF NOT EXISTS max_tier_depth INTEGER NOT NULL DEFAULT 0;

ALTER TABLE users
    ADD COLUMN IF NOT EXISTS default_tier_group_ids JSONB NOT NULL DEFAULT '[]'::jsonb;

ALTER TABLE groups
    ADD COLUMN IF NOT EXISTS tier_fallback_group_id BIGINT NULL
        REFERENCES groups(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_groups_tier_fallback_group_id
    ON groups (tier_fallback_group_id)
    WHERE tier_fallback_group_id IS NOT NULL;

COMMENT ON COLUMN api_keys.tier_group_ids IS '按顺序的降级 group 链路；当主分组所有账号都失败时按列表顺序切换';
COMMENT ON COLUMN api_keys.max_tier_depth IS 'Tier 链路最大深度，0 = 不限制';
COMMENT ON COLUMN users.default_tier_group_ids IS '用户级默认 tier 链路；api_key.tier_group_ids 为空时兜底';
COMMENT ON COLUMN groups.tier_fallback_group_id IS 'Tier 降级链单指针；与 fallback_group_id (CCO) 和 fallback_group_id_on_invalid_request 语义独立';
