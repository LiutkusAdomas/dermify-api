-- +goose Up
-- +goose StatementBegin
CREATE TABLE indication_codes (
    id          BIGSERIAL PRIMARY KEY,
    code        VARCHAR(50) NOT NULL UNIQUE,
    name        VARCHAR(200) NOT NULL,
    module_type VARCHAR(50) NOT NULL,
    active      BOOLEAN NOT NULL DEFAULT true
);

CREATE TABLE clinical_endpoints (
    id          BIGSERIAL PRIMARY KEY,
    code        VARCHAR(50) NOT NULL UNIQUE,
    name        VARCHAR(200) NOT NULL,
    module_type VARCHAR(50) NOT NULL,
    active      BOOLEAN NOT NULL DEFAULT true
);

CREATE INDEX idx_indication_codes_module ON indication_codes(module_type);
CREATE INDEX idx_clinical_endpoints_module ON clinical_endpoints(module_type);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS clinical_endpoints;
DROP TABLE IF EXISTS indication_codes;
-- +goose StatementEnd
