-- +goose Up
-- +goose StatementBegin
CREATE TABLE contraindication_screenings (
    id                  BIGSERIAL PRIMARY KEY,
    session_id          BIGINT NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    pregnant            BOOLEAN NOT NULL DEFAULT false,
    breastfeeding       BOOLEAN NOT NULL DEFAULT false,
    active_infection    BOOLEAN NOT NULL DEFAULT false,
    active_cold_sores   BOOLEAN NOT NULL DEFAULT false,
    isotretinoin        BOOLEAN NOT NULL DEFAULT false,
    photosensitivity    BOOLEAN NOT NULL DEFAULT false,
    autoimmune_disorder BOOLEAN NOT NULL DEFAULT false,
    keloid_history      BOOLEAN NOT NULL DEFAULT false,
    anticoagulants      BOOLEAN NOT NULL DEFAULT false,
    recent_tan          BOOLEAN NOT NULL DEFAULT false,
    has_flags           BOOLEAN NOT NULL DEFAULT false,
    mitigation_notes    TEXT,
    notes               TEXT,
    version             INTEGER NOT NULL DEFAULT 1,
    created_at          TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by          BIGINT NOT NULL REFERENCES users(id),
    updated_at          TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_by          BIGINT NOT NULL REFERENCES users(id),
    UNIQUE (session_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS contraindication_screenings;
-- +goose StatementEnd
