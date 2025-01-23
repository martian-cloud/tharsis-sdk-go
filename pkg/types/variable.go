package types

// VariableCategory represents the categories of variables, Terraform or environment.
type VariableCategory string

// Variable Category types
const (
	TerraformVariableCategory   VariableCategory = "terraform"
	EnvironmentVariableCategory VariableCategory = "environment"
)

// NamespaceVariable models a namespace variable.
type NamespaceVariable struct {
	Value         *string
	Metadata      ResourceMetadata
	ID            string
	NamespacePath string
	Category      VariableCategory
	Key           string
}

// CreateNamespaceVariableInput is the input for creating a namespace variable.
type CreateNamespaceVariableInput struct {
	NamespacePath string           `json:"namespacePath"`
	Category      VariableCategory `json:"category"`
	Key           string           `json:"key"`
	Value         string           `json:"value"` // The value is required, not optional.
}

// SetNamespaceVariablesVariable is the input for setting ALL variables in a namespace.
type SetNamespaceVariablesVariable struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// UpdateNamespaceVariableInput is the input for updating a namespace variable.
type UpdateNamespaceVariableInput struct {
	ID    string `json:"id"`
	Key   string `json:"key"`
	Value string `json:"value"`
}

// GetNamespaceVariableInput is the input for retrieving a namespace variable.
type GetNamespaceVariableInput struct {
	ID string `json:"id"`
}

// DeleteNamespaceVariableInput is the input for deleting a namespace variable.
type DeleteNamespaceVariableInput struct {
	ID string `json:"id"`
}

// SetNamespaceVariablesInput is the input for setting a namespace variable.
type SetNamespaceVariablesInput struct {
	NamespacePath string                          `json:"namespacePath"`
	Category      VariableCategory                `json:"category"`
	Variables     []SetNamespaceVariablesVariable `json:"variables"`
}
