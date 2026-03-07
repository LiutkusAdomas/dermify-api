package domain

// IndicationCode represents a clinical indication code used to categorize treatments.
type IndicationCode struct {
	ID         int64  `json:"id"`
	Code       string `json:"code"`
	Name       string `json:"name"`
	ModuleType string `json:"module_type"`
	Active     bool   `json:"active"`
}

// ClinicalEndpoint represents a measurable clinical outcome for a treatment module.
type ClinicalEndpoint struct {
	ID         int64  `json:"id"`
	Code       string `json:"code"`
	Name       string `json:"name"`
	ModuleType string `json:"module_type"`
	Active     bool   `json:"active"`
}
