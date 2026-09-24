-- Additive external identity mapping for provider-specific login adapters.
CREATE TABLE IF NOT EXISTS external_auth_identities (
    id                    VARCHAR(36) PRIMARY KEY,
    provider              VARCHAR(32) NOT NULL,
    corp_id               VARCHAR(128) NOT NULL,
    external_user_id      VARCHAR(128) NOT NULL,
    union_id              VARCHAR(128),
    user_id               VARCHAR(36) REFERENCES users(id) ON DELETE SET NULL,
    display_name          VARCHAR(255) NOT NULL,
    email                 VARCHAR(255),
    mobile                VARCHAR(64),
    status                VARCHAR(32) NOT NULL,
    binding_method        VARCHAR(32) NOT NULL,
    raw_profile           JSONB NOT NULL DEFAULT '{}'::jsonb,
    last_authenticated_at TIMESTAMPTZ,
    created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_external_auth_identity UNIQUE (provider, corp_id, external_user_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_external_auth_identity_union
    ON external_auth_identities (provider, corp_id, union_id)
    WHERE union_id IS NOT NULL AND union_id <> '';
CREATE INDEX IF NOT EXISTS idx_external_auth_identity_user
    ON external_auth_identities (user_id);
