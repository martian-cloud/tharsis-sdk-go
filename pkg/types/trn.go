package types

import (
	"fmt"
	"strings"
)

const (
	// TRNPrefix is the prefix for all TRNs.
	TRNPrefix = "trn:"
)

// IsTRN indicates if the given string contains "trn:" prefix.
func IsTRN(value string) bool {
	return strings.HasPrefix(value, TRNPrefix)
}

// ParseTRN parses a TRN and returns the model type and resource path.
func ParseTRN(trn string) (modelType string, resourcePath string, err error) {
	if !IsTRN(trn) {
		return "", "", fmt.Errorf("not a TRN: %s", trn)
	}

	parts := strings.Split(trn[len(TRNPrefix):], ":")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid TRN format: %s", trn)
	}

	modelType = parts[0]
	resourcePath = parts[1]

	if resourcePath == "" || strings.HasPrefix(resourcePath, "/") || strings.HasSuffix(resourcePath, "/") {
		return "", "", fmt.Errorf("invalid TRN resource path: %s", resourcePath)
	}

	return modelType, resourcePath, nil
}

// BuildTRN builds a TRN from a model type and resource path.
func BuildTRN(modelType, resourcePath string) string {
	return fmt.Sprintf("%s%s:%s", TRNPrefix, modelType, resourcePath)
}

// ValidateTRNModelType validates that the TRN model type matches the expected type.
func ValidateTRNModelType(trn, expectedModelType string) error {
	modelType, _, err := ParseTRN(trn)
	if err != nil {
		return err
	}
	
	if modelType != expectedModelType {
		return fmt.Errorf("TRN model type mismatch: expected %s, got %s", expectedModelType, modelType)
	}
	
	return nil
}

// ValidateInputIdentifiers ensures only one identifier type is provided and returns the resolved path.
// Returns (path, id, error) where path is extracted from TRN if provided.
func ValidateInputIdentifiers(id, path, trn *string, expectedModelType string) (*string, *string, error) {
	identifierCount := 0
	if id != nil && *id != "" {
		identifierCount++
	}
	if path != nil && *path != "" {
		identifierCount++
	}
	if trn != nil && *trn != "" {
		identifierCount++
	}

	if identifierCount == 0 {
		return nil, nil, fmt.Errorf("must specify one of: ID, Path, or TRN")
	}
	if identifierCount > 1 {
		return nil, nil, fmt.Errorf("must specify only one of: ID, Path, or TRN")
	}

	// If TRN is provided, validate and extract path
	if trn != nil && *trn != "" {
		if err := ValidateTRNModelType(*trn, expectedModelType); err != nil {
			return nil, nil, err
		}
		_, resourcePath, err := ParseTRN(*trn)
		if err != nil {
			return nil, nil, err
		}
		return &resourcePath, nil, nil
	}

	// Return the provided path or id
	return path, id, nil
}

// ValidateIDOrTRN validates input for resources that only support ID or TRN (no path).
// Returns the resolved ID (either original ID or TRN for node query).
func ValidateIDOrTRN(id string, trn *string, expectedModelType string) (string, error) {
	if id == "" && (trn == nil || *trn == "") {
		return "", fmt.Errorf("must specify either ID or TRN")
	}
	if id != "" && trn != nil && *trn != "" {
		return "", fmt.Errorf("must specify only one of: ID or TRN")
	}

	// If TRN is provided, validate it and return for node query
	if trn != nil && *trn != "" {
		if err := ValidateTRNModelType(*trn, expectedModelType); err != nil {
			return "", err
		}
		return *trn, nil
	}

	// Return the original ID
	return id, nil
}
