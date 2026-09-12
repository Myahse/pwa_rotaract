-- +goose Down
DROP TABLE IF EXISTS club_due_payments;

ALTER TABLE chat_groups DROP CONSTRAINT IF EXISTS chat_groups_check;
ALTER TABLE chat_groups ADD CONSTRAINT chat_groups_check CHECK (
    (group_type = 'club' AND commission_id IS NULL) OR
    (group_type = 'commission' AND commission_id IS NOT NULL)
);
