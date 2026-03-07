-- +goose Up
-- +goose StatementBegin

-- IPL devices.
INSERT INTO devices (name, manufacturer, model, device_type) VALUES
    ('Lumenis M22', 'Lumenis', 'M22', 'ipl'),
    ('Candela Nordlys', 'Candela', 'Nordlys', 'ipl');

-- IPL handpieces.
INSERT INTO handpieces (device_id, name) VALUES
    ((SELECT id FROM devices WHERE model = 'M22'), 'IPL 515nm Filter'),
    ((SELECT id FROM devices WHERE model = 'M22'), 'IPL 560nm Filter'),
    ((SELECT id FROM devices WHERE model = 'Nordlys'), 'Ellipse IPL'),
    ((SELECT id FROM devices WHERE model = 'Nordlys'), 'Frax 1550 Non-Ablative');

-- Nd:YAG devices.
INSERT INTO devices (name, manufacturer, model, device_type) VALUES
    ('Candela GentleMax Pro', 'Candela', 'GentleMax Pro', 'ndyag'),
    ('Cutera Excel V+', 'Cutera', 'Excel V+', 'ndyag');

-- Nd:YAG handpieces.
INSERT INTO handpieces (device_id, name) VALUES
    ((SELECT id FROM devices WHERE model = 'GentleMax Pro'), '1064nm Nd:YAG'),
    ((SELECT id FROM devices WHERE model = 'GentleMax Pro'), '755nm Alexandrite'),
    ((SELECT id FROM devices WHERE model = 'Excel V+'), '532nm KTP'),
    ((SELECT id FROM devices WHERE model = 'Excel V+'), '1064nm Nd:YAG Long Pulse');

-- CO2 devices.
INSERT INTO devices (name, manufacturer, model, device_type) VALUES
    ('Lumenis UltraPulse', 'Lumenis', 'UltraPulse', 'co2'),
    ('DEKA SmartXide Touch', 'DEKA', 'SmartXide Touch', 'co2');

-- CO2 handpieces.
INSERT INTO handpieces (device_id, name) VALUES
    ((SELECT id FROM devices WHERE model = 'UltraPulse'), 'ActiveFX'),
    ((SELECT id FROM devices WHERE model = 'UltraPulse'), 'DeepFX'),
    ((SELECT id FROM devices WHERE model = 'SmartXide Touch'), 'DOT Scanner'),
    ((SELECT id FROM devices WHERE model = 'SmartXide Touch'), 'CW Scanner');

-- RF devices.
INSERT INTO devices (name, manufacturer, model, device_type) VALUES
    ('Lutronic Genius', 'Lutronic', 'Genius', 'rf'),
    ('EndyMed Intensif', 'EndyMed', 'Intensif', 'rf');

-- RF handpieces.
INSERT INTO handpieces (device_id, name) VALUES
    ((SELECT id FROM devices WHERE model = 'Genius'), '49-Pin Microneedle'),
    ((SELECT id FROM devices WHERE model = 'Genius'), '16-Pin Microneedle'),
    ((SELECT id FROM devices WHERE model = 'Intensif'), 'Intensif 25-Pin'),
    ((SELECT id FROM devices WHERE model = 'Intensif'), 'Intensif Gold 25-Pin');

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM handpieces WHERE device_id IN (SELECT id FROM devices);
DELETE FROM devices;
-- +goose StatementEnd
