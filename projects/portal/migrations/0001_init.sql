-- +goose Up
-- Portal DB: новости и система предложений (feedback)

CREATE TABLE news (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title       TEXT        NOT NULL,
    body        TEXT        NOT NULL,
    author_id   TEXT        NOT NULL, -- UUID из auth DB
    published   BOOLEAN     NOT NULL DEFAULT false,
    pinned      BOOLEAN     NOT NULL DEFAULT false,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_news_published ON news (published, created_at DESC);

-- Предложения/фидбек игроков
CREATE TABLE feedback_posts (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    author_id   TEXT        NOT NULL, -- UUID из auth DB
    author_name TEXT        NOT NULL,
    title       TEXT        NOT NULL CHECK (length(title) BETWEEN 5 AND 200),
    body        TEXT        NOT NULL CHECK (length(body) >= 20),
    status      TEXT        NOT NULL DEFAULT 'pending'
                    CHECK (status IN ('pending','approved','rejected','implemented')),
    vote_count  BIGINT      NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_feedback_status_votes ON feedback_posts (status, vote_count DESC, created_at DESC);
CREATE INDEX idx_feedback_author ON feedback_posts (author_id, created_at DESC);

-- Голоса за предложения (100 кредитов за голос, неограниченное кол-во)
CREATE TABLE feedback_votes (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    post_id     UUID        NOT NULL REFERENCES feedback_posts (id) ON DELETE CASCADE,
    user_id     TEXT        NOT NULL, -- UUID из auth DB
    credits_spent BIGINT   NOT NULL DEFAULT 100,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
    -- Один пользователь может голосовать несколько раз (богатый игрок = больше влияния)
);

CREATE INDEX idx_feedback_votes_post ON feedback_votes (post_id);
CREATE INDEX idx_feedback_votes_user ON feedback_votes (user_id, post_id);

-- Вложения к предложениям (скриншоты)
CREATE TABLE feedback_attachments (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    post_id     UUID        NOT NULL REFERENCES feedback_posts (id) ON DELETE CASCADE,
    url         TEXT        NOT NULL,
    filename    TEXT        NOT NULL,
    size_bytes  BIGINT      NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_feedback_attachments_post ON feedback_attachments (post_id);

-- Комментарии к предложениям (с поддержкой треда)
CREATE TABLE feedback_comments (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    post_id     UUID        NOT NULL REFERENCES feedback_posts (id) ON DELETE CASCADE,
    parent_id   UUID        REFERENCES feedback_comments (id) ON DELETE CASCADE,
    author_id   TEXT        NOT NULL, -- UUID из auth DB
    author_name TEXT        NOT NULL,
    body        TEXT        NOT NULL CHECK (length(body) >= 1),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    edited_at   TIMESTAMPTZ,
    deleted_at  TIMESTAMPTZ
);

CREATE INDEX idx_feedback_comments_post ON feedback_comments (post_id, created_at ASC);
CREATE INDEX idx_feedback_comments_parent ON feedback_comments (parent_id) WHERE parent_id IS NOT NULL;

-- +goose Down
DROP TABLE IF EXISTS feedback_comments;
DROP TABLE IF EXISTS feedback_attachments;
DROP TABLE IF EXISTS feedback_votes;
DROP TABLE IF EXISTS feedback_posts;
DROP TABLE IF EXISTS news;
