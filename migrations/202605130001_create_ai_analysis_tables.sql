CREATE TABLE IF NOT EXISTS ai_analysis_jobs (
    id BIGSERIAL PRIMARY KEY,
    postcard_id BIGINT NOT NULL,
    postcard_version TEXT NOT NULL,
    status TEXT NOT NULL,
    attempts INTEGER NOT NULL DEFAULT 0,
    next_run_at TIMESTAMPTZ NULL,
    locked_at TIMESTAMPTZ NULL,
    last_error_code TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_ai_analysis_jobs_due
    ON ai_analysis_jobs (status, next_run_at, created_at);

CREATE INDEX IF NOT EXISTS idx_ai_analysis_jobs_postcard_version
    ON ai_analysis_jobs (postcard_id, postcard_version);

CREATE TABLE IF NOT EXISTS media_ai_analysis (
    id BIGSERIAL PRIMARY KEY,
    media_id BIGINT NOT NULL,
    media_version TEXT NOT NULL,
    prompt_version TEXT NOT NULL,
    schema_version TEXT NOT NULL,
    model_version TEXT NOT NULL,
    status TEXT NOT NULL,
    result_json JSONB NULL,
    confidence DOUBLE PRECISION NOT NULL DEFAULT 0,
    uncertainty TEXT NOT NULL DEFAULT '',
    error_code TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_media_ai_analysis_version UNIQUE (media_id, media_version, prompt_version, schema_version, model_version)
);

CREATE TABLE IF NOT EXISTS postcard_ai_analysis (
    id BIGSERIAL PRIMARY KEY,
    postcard_id BIGINT NOT NULL,
    postcard_version TEXT NOT NULL,
    prompt_version TEXT NOT NULL,
    schema_version TEXT NOT NULL,
    model_version TEXT NOT NULL,
    status TEXT NOT NULL,
    result_json JSONB NULL,
    confidence DOUBLE PRECISION NOT NULL DEFAULT 0,
    uncertainty TEXT NOT NULL DEFAULT '',
    error_code TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_postcard_ai_analysis_version UNIQUE (postcard_id, postcard_version, prompt_version, schema_version, model_version)
);
