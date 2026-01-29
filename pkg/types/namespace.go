package types

// NamespaceProviderMirrorEnabled represents the provider mirror enabled setting for a namespace.
type NamespaceProviderMirrorEnabled struct {
	Inherited     bool
	NamespacePath string
	Value         bool
}
