-- +goose Up
CREATE TABLE session_photos (
    id              BIGSERIAL PRIMARY KEY,
    session_id      BIGINT NOT NULL REFERENCES sessions(id),
    module_id       BIGINT REFERENCES session_modules(id),
    photo_type      VARCHAR(20) NOT NULL CHECK (photo_type IN ('before', 'label')),
    file_path       VARCHAR(500) NOT NULL,
    original_name   VARCHAR(255) NOT NULL,
    content_type    VARCHAR(100) NOT NULL,
    size_bytes      BIGINT NOT NULL,
    version         INTEGER NOT NULL DEFAULT 1,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by      BIGINT NOT NULL,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_by      BIGINT NOT NULL,
    CONSTRAINT chk_label_requires_module CHECK (photo_type != 'label' OR module_id IS NOT NULL)
);

CREATE INDEX idx_session_photos_session_id ON session_photos(session_id);
CREATE INDEX idx_session_photos_module_id ON session_photos(module_id) WHERE module_id IS NOT NULL;

-- +goose Down
DROP TABLE IF EXISTS session_photos;
