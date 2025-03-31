// Code generated by applyconfiguration-gen. DO NOT EDIT.

package v1

import (
	machineconfigurationv1 "github.com/openshift/api/machineconfiguration/v1"
)

// PinnedImageRefApplyConfiguration represents a declarative configuration of the PinnedImageRef type for use
// with apply.
type PinnedImageRefApplyConfiguration struct {
	Name *machineconfigurationv1.ImageDigestFormat `json:"name,omitempty"`
}

// PinnedImageRefApplyConfiguration constructs a declarative configuration of the PinnedImageRef type for use with
// apply.
func PinnedImageRef() *PinnedImageRefApplyConfiguration {
	return &PinnedImageRefApplyConfiguration{}
}

// WithName sets the Name field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Name field is set to the value of the last call.
func (b *PinnedImageRefApplyConfiguration) WithName(value machineconfigurationv1.ImageDigestFormat) *PinnedImageRefApplyConfiguration {
	b.Name = &value
	return b
}
