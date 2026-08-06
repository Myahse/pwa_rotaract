-- +goose Up

-- Threaded replies (one level: parent is a top-level comment)
ALTER TABLE social_comments
    ADD COLUMN parent_id UUID REFERENCES social_comments(id) ON DELETE CASCADE;

CREATE INDEX idx_social_comments_parent ON social_comments (parent_id)
    WHERE parent_id IS NOT NULL;

-- Friendships (distinct from follows)
CREATE TABLE social_friendships (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    requester_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    addressee_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status TEXT NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'accepted', 'declined')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (requester_id <> addressee_id)
);

CREATE UNIQUE INDEX idx_social_friendships_pair
    ON social_friendships (LEAST(requester_id, addressee_id), GREATEST(requester_id, addressee_id));

CREATE INDEX idx_social_friendships_addressee_pending
    ON social_friendships (addressee_id, created_at DESC)
    WHERE status = 'pending';

CREATE INDEX idx_social_friendships_user_accepted
    ON social_friendships (requester_id, addressee_id)
    WHERE status = 'accepted';

-- Direct messages between friends
CREATE TABLE social_conversations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE social_conversation_members (
    conversation_id UUID NOT NULL REFERENCES social_conversations(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    PRIMARY KEY (conversation_id, user_id)
);

CREATE INDEX idx_social_conversation_members_user
    ON social_conversation_members (user_id);

-- Enforce one DM thread per unordered pair
CREATE TABLE social_dm_pairs (
    conversation_id UUID PRIMARY KEY REFERENCES social_conversations(id) ON DELETE CASCADE,
    user_low UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    user_high UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    CHECK (user_low < user_high),
    UNIQUE (user_low, user_high)
);

CREATE TABLE social_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    conversation_id UUID NOT NULL REFERENCES social_conversations(id) ON DELETE CASCADE,
    sender_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    body TEXT NOT NULL DEFAULT '',
    shared_comment_id UUID REFERENCES social_comments(id) ON DELETE SET NULL,
    shared_post_id UUID REFERENCES social_posts(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (
        char_length(trim(body)) > 0
        OR shared_comment_id IS NOT NULL
        OR shared_post_id IS NOT NULL
    ),
    CHECK (char_length(body) <= 4000)
);

CREATE INDEX idx_social_messages_conversation
    ON social_messages (conversation_id, created_at DESC);
