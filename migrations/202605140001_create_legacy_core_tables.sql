BEGIN;

CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    username VARCHAR(50) NOT NULL,
    display_name VARCHAR(100),
    email VARCHAR(255) NOT NULL,
    password VARCHAR(255) NOT NULL,
    avatar VARCHAR(500),
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    deleted_at TIMESTAMPTZ NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_users_username
    ON users (username);

CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email
    ON users (email);

CREATE INDEX IF NOT EXISTS idx_users_deleted_at
    ON users (deleted_at);

CREATE TABLE IF NOT EXISTS postcards (
    id BIGSERIAL PRIMARY KEY,
    title VARCHAR(200) NOT NULL,
    content JSONB NOT NULL,
    visibility VARCHAR(20) NOT NULL DEFAULT 'private',
    author_id BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    deleted_at TIMESTAMPTZ NULL
);

CREATE INDEX IF NOT EXISTS idx_postcards_author_id
    ON postcards (author_id);

CREATE INDEX IF NOT EXISTS idx_postcards_deleted_at
    ON postcards (deleted_at);

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'postcards_author_id_fkey'
          AND conrelid = 'postcards'::regclass
    ) THEN
        ALTER TABLE postcards
            ADD CONSTRAINT postcards_author_id_fkey
            FOREIGN KEY (author_id)
            REFERENCES users(id)
            ON UPDATE CASCADE
            ON DELETE RESTRICT;
    END IF;

    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'postcards_title_not_blank'
          AND conrelid = 'postcards'::regclass
    ) THEN
        ALTER TABLE postcards
            ADD CONSTRAINT postcards_title_not_blank
            CHECK (title <> '');
    END IF;

    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'postcards_visibility_valid'
          AND conrelid = 'postcards'::regclass
    ) THEN
        ALTER TABLE postcards
            ADD CONSTRAINT postcards_visibility_valid
            CHECK (visibility IN ('public', 'private'));
    END IF;
END $$;

CREATE TABLE IF NOT EXISTS postcard_media (
    id BIGSERIAL PRIMARY KEY,
    postcard_id BIGINT NOT NULL,
    media_type VARCHAR(20) NOT NULL,
    oss_key VARCHAR(500) NOT NULL,
    url VARCHAR(500) NOT NULL,
    thumbnail_url VARCHAR(500),
    thumbnail_oss_key VARCHAR(500),
    original_width INTEGER,
    original_height INTEGER,
    duration INTEGER,
    file_size BIGINT NOT NULL,
    position INTEGER NOT NULL DEFAULT 1,
    media_group VARCHAR(50) NOT NULL DEFAULT 'gallery',
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    deleted_at TIMESTAMPTZ NULL
);

CREATE INDEX IF NOT EXISTS idx_postcard_media_postcard_id
    ON postcard_media (postcard_id);

CREATE UNIQUE INDEX IF NOT EXISTS idx_postcard_media_position
    ON postcard_media (postcard_id, position);

CREATE INDEX IF NOT EXISTS idx_postcard_media_deleted_at
    ON postcard_media (deleted_at);

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'postcard_media_postcard_id_fkey'
          AND conrelid = 'postcard_media'::regclass
    ) THEN
        ALTER TABLE postcard_media
            ADD CONSTRAINT postcard_media_postcard_id_fkey
            FOREIGN KEY (postcard_id)
            REFERENCES postcards(id)
            ON UPDATE CASCADE
            ON DELETE CASCADE;
    END IF;

    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'postcard_media_type_valid'
          AND conrelid = 'postcard_media'::regclass
    ) THEN
        ALTER TABLE postcard_media
            ADD CONSTRAINT postcard_media_type_valid
            CHECK (media_type IN ('image', 'video', 'audio'));
    END IF;

    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'postcard_media_file_size_positive'
          AND conrelid = 'postcard_media'::regclass
    ) THEN
        ALTER TABLE postcard_media
            ADD CONSTRAINT postcard_media_file_size_positive
            CHECK (file_size > 0);
    END IF;

    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'postcard_media_position_positive'
          AND conrelid = 'postcard_media'::regclass
    ) THEN
        ALTER TABLE postcard_media
            ADD CONSTRAINT postcard_media_position_positive
            CHECK (position > 0);
    END IF;

    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'postcard_media_group_valid'
          AND conrelid = 'postcard_media'::regclass
    ) THEN
        ALTER TABLE postcard_media
            ADD CONSTRAINT postcard_media_group_valid
            CHECK (media_group IN ('header', 'gallery', 'bgm'));
    END IF;
END $$;

COMMIT;
