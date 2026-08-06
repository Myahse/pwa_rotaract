DELETE FROM permissions WHERE key = 'club.requests.review';

DROP TABLE IF EXISTS access_requests;
DROP TYPE IF EXISTS access_request_status;

DROP INDEX IF EXISTS idx_clubs_invite_code;
ALTER TABLE clubs DROP COLUMN IF EXISTS invite_code;

ALTER TABLE users
    DROP COLUMN IF EXISTS birth_date,
    DROP COLUMN IF EXISTS profession,
    DROP COLUMN IF EXISTS member_since;
