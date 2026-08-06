-- +goose Down
DROP TABLE IF EXISTS social_group_messages;
DROP TABLE IF EXISTS social_group_join_requests;
DROP TABLE IF EXISTS social_group_members;
DROP TABLE IF EXISTS social_groups;
