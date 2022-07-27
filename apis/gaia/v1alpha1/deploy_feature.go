package v1alpha1

//DeployFeature container deployment feature
type DeployFeature struct {
	PersistentStorage bool `json:"persistentStorage,omitempty"`
	DifferentHost     bool `json:"differentHost,omitempty"`
}
