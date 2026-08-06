-- +goose Down
DROP TABLE IF EXISTS social_messages;
DROP TABLE IF EXISTS social_dm_pairs;
DROP TABLE IF EXISTS social_conversation_members;
DROP TABLE IF EXISTS social_conversations;
DROP TABLE IF EXISTS social_friendships;
DROP INDEX IF EXISTS idx_social_comments_parent;
ALTER TABLE social_comments DROP COLUMN IF EXISTS parent_id;
