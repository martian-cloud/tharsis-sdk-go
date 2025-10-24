package types

import (
	"testing"
)

func TestIsTRN(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"valid TRN", "trn:workspace:group1/my-workspace", true},
		{"valid TRN minimal", "trn:group:test", true},
		{"empty string", "", false},
		{"regular ID", "01234567-89ab-cdef-0123-456789abcdef", false},
		{"path only", "group1/workspace", false},
		{"partial TRN", "trn:", true},
		{"case sensitive", "TRN:workspace:test", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsTRN(tt.input)
			if result != tt.expected {
				t.Errorf("IsTRN(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseTRN(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expectedType string
		expectedPath string
		expectError  bool
	}{
		{
			name:         "valid workspace TRN",
			input:        "trn:workspace:group1/my-workspace",
			expectedType: "workspace",
			expectedPath: "group1/my-workspace",
			expectError:  false,
		},
		{
			name:         "valid group TRN",
			input:        "trn:group:parent/child",
			expectedType: "group",
			expectedPath: "parent/child",
			expectError:  false,
		},
		{
			name:         "valid module version TRN",
			input:        "trn:terraform_module_version:group1/module/1.0.0",
			expectedType: "terraform_module_version",
			expectedPath: "group1/module/1.0.0",
			expectError:  false,
		},
		{
			name:        "not a TRN",
			input:       "01234567-89ab-cdef-0123-456789abcdef",
			expectError: true,
		},
		{
			name:        "invalid format - no colon separator",
			input:       "trn:workspace",
			expectError: true,
		},
		{
			name:        "invalid format - too many parts",
			input:       "trn:workspace:group1:extra:part",
			expectError: true,
		},
		{
			name:        "empty resource path",
			input:       "trn:workspace:",
			expectError: true,
		},
		{
			name:        "resource path starts with slash",
			input:       "trn:workspace:/group1/workspace",
			expectError: true,
		},
		{
			name:        "resource path ends with slash",
			input:       "trn:workspace:group1/workspace/",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			modelType, resourcePath, err := ParseTRN(tt.input)
			
			if tt.expectError {
				if err == nil {
					t.Errorf("ParseTRN(%q) expected error, got none", tt.input)
				}
				return
			}
			
			if err != nil {
				t.Errorf("ParseTRN(%q) unexpected error: %v", tt.input, err)
				return
			}
			
			if modelType != tt.expectedType {
				t.Errorf("ParseTRN(%q) modelType = %q, want %q", tt.input, modelType, tt.expectedType)
			}
			
			if resourcePath != tt.expectedPath {
				t.Errorf("ParseTRN(%q) resourcePath = %q, want %q", tt.input, resourcePath, tt.expectedPath)
			}
		})
	}
}

func TestBuildTRN(t *testing.T) {
	tests := []struct {
		name         string
		modelType    string
		resourcePath string
		expected     string
	}{
		{
			name:         "workspace TRN",
			modelType:    "workspace",
			resourcePath: "group1/my-workspace",
			expected:     "trn:workspace:group1/my-workspace",
		},
		{
			name:         "group TRN",
			modelType:    "group",
			resourcePath: "parent/child",
			expected:     "trn:group:parent/child",
		},
		{
			name:         "module version TRN",
			modelType:    "terraform_module_version",
			resourcePath: "group1/module/1.0.0",
			expected:     "trn:terraform_module_version:group1/module/1.0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildTRN(tt.modelType, tt.resourcePath)
			if result != tt.expected {
				t.Errorf("BuildTRN(%q, %q) = %q, want %q", tt.modelType, tt.resourcePath, result, tt.expected)
			}
		})
	}
}

func TestValidateTRNModelType(t *testing.T) {
	tests := []struct {
		name              string
		trn               string
		expectedModelType string
		expectError       bool
	}{
		{
			name:              "valid workspace TRN",
			trn:               "trn:workspace:group1/my-workspace",
			expectedModelType: "workspace",
			expectError:       false,
		},
		{
			name:              "model type mismatch",
			trn:               "trn:workspace:group1/my-workspace",
			expectedModelType: "group",
			expectError:       true,
		},
		{
			name:              "invalid TRN format",
			trn:               "not-a-trn",
			expectedModelType: "workspace",
			expectError:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTRNModelType(tt.trn, tt.expectedModelType)
			
			if tt.expectError {
				if err == nil {
					t.Errorf("ValidateTRNModelType(%q, %q) expected error, got none", tt.trn, tt.expectedModelType)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateTRNModelType(%q, %q) unexpected error: %v", tt.trn, tt.expectedModelType, err)
				}
			}
		})
	}
}

func TestValidateInputIdentifiers(t *testing.T) {
	id := "01234567-89ab-cdef-0123-456789abcdef"
	path := "group1/my-workspace"
	trn := "trn:workspace:group1/my-workspace"
	
	tests := []struct {
		name              string
		id                *string
		path              *string
		trn               *string
		expectedModelType string
		expectedPath      *string
		expectedID        *string
		expectError       bool
	}{
		{
			name:              "ID provided",
			id:                &id,
			expectedModelType: "workspace",
			expectedPath:      nil,
			expectedID:        &id,
			expectError:       false,
		},
		{
			name:              "Path provided",
			path:              &path,
			expectedModelType: "workspace",
			expectedPath:      &path,
			expectedID:        nil,
			expectError:       false,
		},
		{
			name:              "TRN provided",
			trn:               &trn,
			expectedModelType: "workspace",
			expectedPath:      &path,
			expectedID:        nil,
			expectError:       false,
		},
		{
			name:              "no identifiers",
			expectedModelType: "workspace",
			expectError:       true,
		},
		{
			name:              "ID and Path provided",
			id:                &id,
			path:              &path,
			expectedModelType: "workspace",
			expectError:       true,
		},
		{
			name:              "ID and TRN provided",
			id:                &id,
			trn:               &trn,
			expectedModelType: "workspace",
			expectError:       true,
		},
		{
			name:              "Path and TRN provided",
			path:              &path,
			trn:               &trn,
			expectedModelType: "workspace",
			expectError:       true,
		},
		{
			name:              "TRN model type mismatch",
			trn:               &trn,
			expectedModelType: "group",
			expectError:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resultPath, resultID, err := ValidateInputIdentifiers(tt.id, tt.path, tt.trn, tt.expectedModelType)
			
			if tt.expectError {
				if err == nil {
					t.Errorf("ValidateInputIdentifiers() expected error, got none")
				}
				return
			}
			
			if err != nil {
				t.Errorf("ValidateInputIdentifiers() unexpected error: %v", err)
				return
			}
			
			if (resultPath == nil) != (tt.expectedPath == nil) {
				t.Errorf("ValidateInputIdentifiers() path result mismatch: got %v, want %v", resultPath, tt.expectedPath)
			} else if resultPath != nil && tt.expectedPath != nil && *resultPath != *tt.expectedPath {
				t.Errorf("ValidateInputIdentifiers() path = %q, want %q", *resultPath, *tt.expectedPath)
			}
			
			if (resultID == nil) != (tt.expectedID == nil) {
				t.Errorf("ValidateInputIdentifiers() ID result mismatch: got %v, want %v", resultID, tt.expectedID)
			} else if resultID != nil && tt.expectedID != nil && *resultID != *tt.expectedID {
				t.Errorf("ValidateInputIdentifiers() ID = %q, want %q", *resultID, *tt.expectedID)
			}
		})
	}
}

func TestValidateIDOrTRN(t *testing.T) {
	id := "01234567-89ab-cdef-0123-456789abcdef"
	trn := "trn:service_account:group1/my-service-account"
	
	tests := []struct {
		name              string
		id                string
		trn               *string
		expectedModelType string
		expectedResult    string
		expectError       bool
	}{
		{
			name:              "ID provided",
			id:                id,
			expectedModelType: "service_account",
			expectedResult:    id,
			expectError:       false,
		},
		{
			name:              "TRN provided",
			trn:               &trn,
			expectedModelType: "service_account",
			expectedResult:    trn,
			expectError:       false,
		},
		{
			name:              "no identifiers",
			expectedModelType: "service_account",
			expectError:       true,
		},
		{
			name:              "both ID and TRN provided",
			id:                id,
			trn:               &trn,
			expectedModelType: "service_account",
			expectError:       true,
		},
		{
			name:              "TRN model type mismatch",
			trn:               &trn,
			expectedModelType: "workspace",
			expectError:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ValidateIDOrTRN(tt.id, tt.trn, tt.expectedModelType)
			
			if tt.expectError {
				if err == nil {
					t.Errorf("ValidateIDOrTRN() expected error, got none")
				}
				return
			}
			
			if err != nil {
				t.Errorf("ValidateIDOrTRN() unexpected error: %v", err)
				return
			}
			
			if result != tt.expectedResult {
				t.Errorf("ValidateIDOrTRN() = %q, want %q", result, tt.expectedResult)
			}
		})
	}
}
