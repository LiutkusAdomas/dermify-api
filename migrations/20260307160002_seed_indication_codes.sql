-- +goose Up
-- +goose StatementBegin

-- IPL indication codes.
INSERT INTO indication_codes (code, name, module_type) VALUES
    ('IPL_PIGMENTATION', 'Pigmented lesion treatment', 'ipl'),
    ('IPL_VASCULAR', 'Vascular lesion treatment', 'ipl'),
    ('IPL_REJUVENATION', 'Skin rejuvenation', 'ipl'),
    ('IPL_ROSACEA', 'Rosacea management', 'ipl'),
    ('IPL_HAIR_REMOVAL', 'Hair removal', 'ipl');

-- Nd:YAG indication codes.
INSERT INTO indication_codes (code, name, module_type) VALUES
    ('NDYAG_HAIR_REMOVAL', 'Laser hair removal', 'ndyag'),
    ('NDYAG_VASCULAR', 'Vascular lesion treatment', 'ndyag'),
    ('NDYAG_TATTOO', 'Tattoo removal', 'ndyag'),
    ('NDYAG_SKIN_TIGHTENING', 'Skin tightening', 'ndyag'),
    ('NDYAG_PIGMENTATION', 'Pigmented lesion treatment', 'ndyag');

-- CO2 indication codes.
INSERT INTO indication_codes (code, name, module_type) VALUES
    ('CO2_RESURFACING', 'Skin resurfacing', 'co2'),
    ('CO2_SCAR', 'Scar treatment', 'co2'),
    ('CO2_WRINKLE', 'Wrinkle reduction', 'co2'),
    ('CO2_LESION', 'Lesion removal', 'co2'),
    ('CO2_REJUVENATION', 'Skin rejuvenation', 'co2');

-- RF indication codes.
INSERT INTO indication_codes (code, name, module_type) VALUES
    ('RF_TIGHTENING', 'Skin tightening', 'rf'),
    ('RF_WRINKLE', 'Wrinkle reduction', 'rf'),
    ('RF_ACNE_SCAR', 'Acne scar treatment', 'rf'),
    ('RF_CONTOURING', 'Body contouring', 'rf'),
    ('RF_REJUVENATION', 'Skin rejuvenation', 'rf');

-- Filler indication codes.
INSERT INTO indication_codes (code, name, module_type) VALUES
    ('FILLER_VOLUME', 'Volume restoration', 'filler'),
    ('FILLER_CONTOUR', 'Facial contouring', 'filler'),
    ('FILLER_WRINKLE', 'Wrinkle filling', 'filler'),
    ('FILLER_LIP', 'Lip augmentation', 'filler'),
    ('FILLER_TEAR_TROUGH', 'Tear trough correction', 'filler');

-- Botulinum toxin indication codes.
INSERT INTO indication_codes (code, name, module_type) VALUES
    ('BTX_GLABELLAR', 'Glabellar lines treatment', 'botulinum_toxin'),
    ('BTX_FOREHEAD', 'Forehead lines treatment', 'botulinum_toxin'),
    ('BTX_CROWS_FEET', 'Crow''s feet treatment', 'botulinum_toxin'),
    ('BTX_MASSETER', 'Masseter reduction', 'botulinum_toxin'),
    ('BTX_HYPERHIDROSIS', 'Hyperhidrosis treatment', 'botulinum_toxin');

-- IPL clinical endpoints.
INSERT INTO clinical_endpoints (code, name, module_type) VALUES
    ('IPL_LESION_CLEARANCE', 'Lesion clearance percentage', 'ipl'),
    ('IPL_ERYTHEMA_REDUCTION', 'Erythema reduction score', 'ipl'),
    ('IPL_PIGMENT_REDUCTION', 'Pigment reduction grade', 'ipl'),
    ('IPL_PATIENT_SATISFACTION', 'Patient satisfaction score', 'ipl'),
    ('IPL_HAIR_REDUCTION', 'Hair reduction percentage', 'ipl');

-- Nd:YAG clinical endpoints.
INSERT INTO clinical_endpoints (code, name, module_type) VALUES
    ('NDYAG_HAIR_CLEARANCE', 'Hair clearance rate', 'ndyag'),
    ('NDYAG_VESSEL_CLEARANCE', 'Vessel clearance score', 'ndyag'),
    ('NDYAG_INK_FADING', 'Tattoo ink fading grade', 'ndyag'),
    ('NDYAG_SKIN_LAXITY', 'Skin laxity improvement', 'ndyag'),
    ('NDYAG_PATIENT_SATISFACTION', 'Patient satisfaction score', 'ndyag');

-- CO2 clinical endpoints.
INSERT INTO clinical_endpoints (code, name, module_type) VALUES
    ('CO2_TEXTURE_IMPROVEMENT', 'Skin texture improvement score', 'co2'),
    ('CO2_SCAR_REDUCTION', 'Scar reduction grade', 'co2'),
    ('CO2_WRINKLE_REDUCTION', 'Wrinkle depth reduction', 'co2'),
    ('CO2_DOWNTIME', 'Recovery downtime in days', 'co2'),
    ('CO2_PATIENT_SATISFACTION', 'Patient satisfaction score', 'co2');

-- RF clinical endpoints.
INSERT INTO clinical_endpoints (code, name, module_type) VALUES
    ('RF_LAXITY_IMPROVEMENT', 'Skin laxity improvement grade', 'rf'),
    ('RF_WRINKLE_REDUCTION', 'Wrinkle reduction score', 'rf'),
    ('RF_SCAR_IMPROVEMENT', 'Scar improvement grade', 'rf'),
    ('RF_COLLAGEN_RESPONSE', 'Collagen remodeling response', 'rf'),
    ('RF_PATIENT_SATISFACTION', 'Patient satisfaction score', 'rf');

-- Filler clinical endpoints.
INSERT INTO clinical_endpoints (code, name, module_type) VALUES
    ('FILLER_VOLUME_SCORE', 'Volume correction score', 'filler'),
    ('FILLER_SYMMETRY', 'Symmetry assessment grade', 'filler'),
    ('FILLER_WRINKLE_SCORE', 'Wrinkle severity scale change', 'filler'),
    ('FILLER_DURATION', 'Product longevity in months', 'filler'),
    ('FILLER_PATIENT_SATISFACTION', 'Patient satisfaction score', 'filler');

-- Botulinum toxin clinical endpoints.
INSERT INTO clinical_endpoints (code, name, module_type) VALUES
    ('BTX_LINE_REDUCTION', 'Line severity reduction score', 'botulinum_toxin'),
    ('BTX_ONSET_DAYS', 'Onset time in days', 'botulinum_toxin'),
    ('BTX_DURATION_WEEKS', 'Effect duration in weeks', 'botulinum_toxin'),
    ('BTX_SYMMETRY', 'Treatment symmetry assessment', 'botulinum_toxin'),
    ('BTX_PATIENT_SATISFACTION', 'Patient satisfaction score', 'botulinum_toxin');

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM clinical_endpoints;
DELETE FROM indication_codes;
-- +goose StatementEnd
