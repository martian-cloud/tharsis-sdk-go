package types

// Supporting structs for the TerraformModuleAttestation paginator:

// TerraformModuleAttestationSortableField represents the fields that a TerraformModuleAttestation can be sorted by
type TerraformModuleAttestationSortableField string

// TerraformModuleSortableField constants
const (
	TerraformModuleAttestationSortableFieldPredicateAsc  TerraformModuleAttestationSortableField = "PREDICATE_ASC"
	TerraformModuleAttestationSortableFieldPredicateDesc TerraformModuleAttestationSortableField = "PREDICATE_DESC"
	TerraformModuleAttestationSortableFieldCreatedAtAsc  TerraformModuleAttestationSortableField = "CREATED_AT_ASC"
	TerraformModuleAttestationSortableFieldCreatedAtDesc TerraformModuleAttestationSortableField = "CREATED_AT_DESC"
)

// TerraformModuleAttestationFilter contains the supported fields for filtering TerraformModuleAttestation resources
type TerraformModuleAttestationFilter struct {
	Digest                   *string
	TerraformModuleID        *string
	TerraformModuleVersionID *string
}

// GetTerraformModuleAttestationsInput is the input for listing TerraformModuleAttestations
type GetTerraformModuleAttestationsInput struct {
	// Sort specifies the field to sort on and direction
	Sort *TerraformModuleAttestationSortableField
	// PaginationOptions supports cursor based pagination
	PaginationOptions *PaginationOptions
	// Filter is used to filter the results
	Filter *TerraformModuleAttestationFilter
}

// GetTerraformModuleAttestationsOutput is the output when listing TerraformModuleAttestations
type GetTerraformModuleAttestationsOutput struct {
	PageInfo           *PageInfo
	ModuleAttestations []TerraformModuleAttestation
}

// GetPageInfo allows GetTerraformModulesOutput to implement the PaginatedResponse interface.
func (gta *GetTerraformModuleAttestationsOutput) GetPageInfo() *PageInfo {
	return gta.PageInfo
}

// TerraformModuleAttestation represents a terraform module attestation
type TerraformModuleAttestation struct {
	ModuleID      string
	Description   string
	SchemaType    string
	PredicateType string
	Data          string
	Metadata      ResourceMetadata
	Digests       []string
}

// CreateTerraformModuleAttestationInput is the input for creating a terraform module attestation.
type CreateTerraformModuleAttestationInput struct {
	ModulePath      string `json:"modulePath"`
	Description     string `json:"description"`
	AttestationData string `json:"attestationData"`
}

// UpdateTerraformModuleAttestationInput is the input for updating a terraform module attestation.
type UpdateTerraformModuleAttestationInput struct {
	ID          string `json:"id"`
	Description string `json:"description"`
}

// DeleteTerraformModuleAttestationInput is the input for deleting a terraform module attestation.
type DeleteTerraformModuleAttestationInput struct {
	ID string `json:"id"`
}
