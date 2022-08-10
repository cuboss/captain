package apis

import "captain/apis/gaia/v1alpha1"

func init() {
	AddToSchemes = append(AddToSchemes, v1alpha1.SchemeBuilder.AddToScheme)
}
