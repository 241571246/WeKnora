CREATE TABLE IF NOT EXISTS external_auth_identities (
    id                    TEXT PRIMARY KEY,
    provider              TEXT NOT NULL,
    corp_id               TEXT NOT NULL,
    external_user_id      TEXT NOT NULL,
    union_id              TEXT,
    user_id               TEXT REFERENCES users(id) ON DELETE SET NULL,
    display_name          TEXT NOT NULL,
    email                 TEXT,
    mobile                TEXT,
    status                TEXT NOT NULL,
    binding_method        TEXT NOT NULL,
    raw_profile           TEXT NOT NULL DEFAULT '{}',
    last_authenticated_at DATETIME,
    created_at            DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (provider, corp_id, external_user_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_external_auth_identity_union
    ON external_auth_identities (provider, corp_id, union_id)
    WHERE union_id IS NOT NULL AND union_id <> '';
CREATE INDEX IF NOT EXISTS idx_external_auth_identity_user
    ON external_auth_identities (user_id);
