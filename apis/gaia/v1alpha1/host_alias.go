package v1alpha1

//HostAlias host aliases record
type HostAlias struct {
	// Hosts host names
	Hosts []string `json:"hosts"`
	// IP host ip
	IP string `json:"ip"`
}
