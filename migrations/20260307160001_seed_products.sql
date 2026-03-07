-- +goose Up
-- +goose StatementBegin

-- Filler products.
INSERT INTO products (name, manufacturer, product_type, concentration) VALUES
    ('Juvederm Ultra XC', 'Allergan', 'filler', '24 mg/mL'),
    ('Restylane Lyft', 'Galderma', 'filler', '20 mg/mL'),
    ('Belotero Balance', 'Merz', 'filler', '22.5 mg/mL');

-- Botulinum toxin products.
INSERT INTO products (name, manufacturer, product_type, concentration) VALUES
    ('Botox', 'Allergan', 'botulinum_toxin', '100 units/vial'),
    ('Dysport', 'Galderma', 'botulinum_toxin', '300 units/vial'),
    ('Xeomin', 'Merz', 'botulinum_toxin', '100 units/vial');

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM products;
-- +goose StatementEnd
