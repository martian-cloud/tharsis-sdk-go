package types

// TerraformProviderVersionMirrorSortableField represents fields
// that a TerraformProviderVersionMirror can be sorted by.
type TerraformProviderVersionMirrorSortableField string

// TerraformProviderVersionMirrorSortableField constants
const (
	TerraformProviderVersionMirrorSortableFieldCreatedAtAsc  TerraformProviderVersionMirrorSortableField = "CREATED_AT_ASC"
	TerraformProviderVersionMirrorSortableFieldCreatedAtDesc TerraformProviderVersionMirrorSortableField = "CREATED_AT_DESC"
)

// GetTerraformProviderVersionMirrorsInput is the input for listing TerraformProviderVersionMirrors.
type GetTerraformProviderVersionMirrorsInput struct {
	// Include inherited returns version mirrors that are inherited
	IncludeInherited *bool
	// Sort specifies the field to sort on and direction
	Sort *TerraformProviderVersionMirrorSortableField
	// PaginationOptions supports cursor based pagination
	PaginationOptions *PaginationOptions
	// GroupPath is the path of the group that contains the version mirror.
	GroupPath string
}

// GetTerraformProviderVersionMirrorsOutput is the output when listing TerraformProviderVersionMirrors.
type GetTerraformProviderVersionMirrorsOutput struct {
	PageInfo       *PageInfo
	VersionMirrors []TerraformProviderVersionMirror
}

// GetPageInfo allows GetTerraformProviderVersionMirrorsOutput to implement the PaginatedResponse interface.
func (o *GetTerraformProviderVersionMirrorsOutput) GetPageInfo() *PageInfo {
	return o.PageInfo
}

// TerraformProviderVersionMirror represents a Tharsis provider version mirror.
type TerraformProviderVersionMirror struct {
	Metadata          ResourceMetadata
	SemanticVersion   string
	RegistryHostname  string
	RegistryNamespace string
	Type              string
}

// GetTerraformProviderVersionMirrorInput is the input to specify a single provider version mirror to fetch.
type GetTerraformProviderVersionMirrorInput struct {
	ID string `json:"id"`
}

// GetTerraformProviderVersionMirrorByAddressInput is the input for retrieving a single
// provider version mirror by address.
type GetTerraformProviderVersionMirrorByAddressInput struct {
	RegistryHostname  string `json:"registryHostname"`
	RegistryNamespace string `json:"registryNamespace"`
	Type              string `json:"type"`
	Version           string `json:"version"`
	GroupPath         string `json:"groupPath"`
}

// GetAvailableProviderVersionsInput is the input for retrieving all cached versions for a provider.
type GetAvailableProviderVersionsInput struct {
	GroupPath         string
	RegistryHostname  string
	RegistryNamespace string
	Type              string
}

// CreateTerraformProviderVersionMirrorInput is the input for creating a new provider version mirror.
type CreateTerraformProviderVersionMirrorInput struct {
	RegistryToken     *string `json:"registryToken,omitempty"`
	GroupPath         string  `json:"groupPath"`
	Type              string  `json:"type"`
	RegistryNamespace string  `json:"registryNamespace"`
	RegistryHostname  string  `json:"registryHostname"`
	SemanticVersion   string  `json:"semanticVersion"`
}

// DeleteTerraformProviderVersionMirrorInput is the input for deleting a provider version mirror.
type DeleteTerraformProviderVersionMirrorInput struct {
	ID    string `json:"id"`
	Force bool   `json:"force"`
}
