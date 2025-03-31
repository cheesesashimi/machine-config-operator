// Code generated by applyconfiguration-gen. DO NOT EDIT.

package internal

import (
	fmt "fmt"
	sync "sync"

	typed "sigs.k8s.io/structured-merge-diff/v4/typed"
)

func Parser() *typed.Parser {
	parserOnce.Do(func() {
		var err error
		parser, err = typed.NewParser(schemaYAML)
		if err != nil {
			panic(fmt.Sprintf("Failed to parse schema: %v", err))
		}
	})
	return parser
}

var parserOnce sync.Once
var parser *typed.Parser
var schemaYAML = typed.YAMLObject(`types:
- name: com.github.openshift.api.machineconfiguration.v1.ContainerRuntimeConfig
  scalar: untyped
  list:
    elementType:
      namedType: __untyped_atomic_
    elementRelationship: atomic
  map:
    elementType:
      namedType: __untyped_deduced_
    elementRelationship: separable
- name: com.github.openshift.api.machineconfiguration.v1.ControllerConfig
  scalar: untyped
  list:
    elementType:
      namedType: __untyped_atomic_
    elementRelationship: atomic
  map:
    elementType:
      namedType: __untyped_deduced_
    elementRelationship: separable
- name: com.github.openshift.api.machineconfiguration.v1.KubeletConfig
  scalar: untyped
  list:
    elementType:
      namedType: __untyped_atomic_
    elementRelationship: atomic
  map:
    elementType:
      namedType: __untyped_deduced_
    elementRelationship: separable
- name: com.github.openshift.api.machineconfiguration.v1.MachineConfig
  scalar: untyped
  list:
    elementType:
      namedType: __untyped_atomic_
    elementRelationship: atomic
  map:
    elementType:
      namedType: __untyped_deduced_
    elementRelationship: separable
- name: com.github.openshift.api.machineconfiguration.v1.MachineConfigPool
  scalar: untyped
  list:
    elementType:
      namedType: __untyped_atomic_
    elementRelationship: atomic
  map:
    elementType:
      namedType: __untyped_deduced_
    elementRelationship: separable
- name: com.github.openshift.api.machineconfiguration.v1.MachineOSBuild
  scalar: untyped
  list:
    elementType:
      namedType: __untyped_atomic_
    elementRelationship: atomic
  map:
    elementType:
      namedType: __untyped_deduced_
    elementRelationship: separable
- name: com.github.openshift.api.machineconfiguration.v1.MachineOSConfig
  scalar: untyped
  list:
    elementType:
      namedType: __untyped_atomic_
    elementRelationship: atomic
  map:
    elementType:
      namedType: __untyped_deduced_
    elementRelationship: separable
- name: com.github.openshift.api.machineconfiguration.v1.PinnedImageSet
  scalar: untyped
  list:
    elementType:
      namedType: __untyped_atomic_
    elementRelationship: atomic
  map:
    elementType:
      namedType: __untyped_deduced_
    elementRelationship: separable
- name: com.github.openshift.api.machineconfiguration.v1alpha1.BuildInputs
  map:
    fields:
    - name: baseImagePullSecret
      type:
        namedType: com.github.openshift.api.machineconfiguration.v1alpha1.ImageSecretObjectReference
      default: {}
    - name: baseOSExtensionsImagePullspec
      type:
        scalar: string
    - name: baseOSImagePullspec
      type:
        scalar: string
    - name: containerFile
      type:
        list:
          elementType:
            namedType: com.github.openshift.api.machineconfiguration.v1alpha1.MachineOSContainerfile
          elementRelationship: associative
          keys:
          - containerfileArch
    - name: imageBuilder
      type:
        namedType: com.github.openshift.api.machineconfiguration.v1alpha1.MachineOSImageBuilder
    - name: releaseVersion
      type:
        scalar: string
    - name: renderedImagePushSecret
      type:
        namedType: com.github.openshift.api.machineconfiguration.v1alpha1.ImageSecretObjectReference
      default: {}
    - name: renderedImagePushspec
      type:
        scalar: string
      default: ""
- name: com.github.openshift.api.machineconfiguration.v1alpha1.BuildOutputs
  map:
    fields:
    - name: currentImagePullSecret
      type:
        namedType: com.github.openshift.api.machineconfiguration.v1alpha1.ImageSecretObjectReference
      default: {}
    unions:
    - fields:
      - fieldName: currentImagePullSecret
        discriminatorValue: CurrentImagePullSecret
- name: com.github.openshift.api.machineconfiguration.v1alpha1.ImageSecretObjectReference
  map:
    fields:
    - name: name
      type:
        scalar: string
      default: ""
- name: com.github.openshift.api.machineconfiguration.v1alpha1.MCOObjectReference
  map:
    fields:
    - name: name
      type:
        scalar: string
      default: ""
- name: com.github.openshift.api.machineconfiguration.v1alpha1.MachineConfigNode
  map:
    fields:
    - name: apiVersion
      type:
        scalar: string
    - name: kind
      type:
        scalar: string
    - name: metadata
      type:
        namedType: io.k8s.apimachinery.pkg.apis.meta.v1.ObjectMeta
      default: {}
    - name: spec
      type:
        namedType: com.github.openshift.api.machineconfiguration.v1alpha1.MachineConfigNodeSpec
      default: {}
    - name: status
      type:
        namedType: com.github.openshift.api.machineconfiguration.v1alpha1.MachineConfigNodeStatus
      default: {}
- name: com.github.openshift.api.machineconfiguration.v1alpha1.MachineConfigNodeSpec
  map:
    fields:
    - name: configVersion
      type:
        namedType: com.github.openshift.api.machineconfiguration.v1alpha1.MachineConfigNodeSpecMachineConfigVersion
      default: {}
    - name: node
      type:
        namedType: com.github.openshift.api.machineconfiguration.v1alpha1.MCOObjectReference
      default: {}
    - name: pinnedImageSets
      type:
        list:
          elementType:
            namedType: com.github.openshift.api.machineconfiguration.v1alpha1.MachineConfigNodeSpecPinnedImageSet
          elementRelationship: associative
          keys:
          - name
    - name: pool
      type:
        namedType: com.github.openshift.api.machineconfiguration.v1alpha1.MCOObjectReference
      default: {}
- name: com.github.openshift.api.machineconfiguration.v1alpha1.MachineConfigNodeSpecMachineConfigVersion
  map:
    fields:
    - name: desired
      type:
        scalar: string
      default: ""
- name: com.github.openshift.api.machineconfiguration.v1alpha1.MachineConfigNodeSpecPinnedImageSet
  map:
    fields:
    - name: name
      type:
        scalar: string
      default: ""
- name: com.github.openshift.api.machineconfiguration.v1alpha1.MachineConfigNodeStatus
  map:
    fields:
    - name: conditions
      type:
        list:
          elementType:
            namedType: io.k8s.apimachinery.pkg.apis.meta.v1.Condition
          elementRelationship: associative
          keys:
          - type
    - name: configVersion
      type:
        namedType: com.github.openshift.api.machineconfiguration.v1alpha1.MachineConfigNodeStatusMachineConfigVersion
      default: {}
    - name: observedGeneration
      type:
        scalar: numeric
    - name: pinnedImageSets
      type:
        list:
          elementType:
            namedType: com.github.openshift.api.machineconfiguration.v1alpha1.MachineConfigNodeStatusPinnedImageSet
          elementRelationship: associative
          keys:
          - name
- name: com.github.openshift.api.machineconfiguration.v1alpha1.MachineConfigNodeStatusMachineConfigVersion
  map:
    fields:
    - name: current
      type:
        scalar: string
      default: ""
    - name: desired
      type:
        scalar: string
      default: ""
- name: com.github.openshift.api.machineconfiguration.v1alpha1.MachineConfigNodeStatusPinnedImageSet
  map:
    fields:
    - name: currentGeneration
      type:
        scalar: numeric
    - name: desiredGeneration
      type:
        scalar: numeric
    - name: lastFailedGeneration
      type:
        scalar: numeric
    - name: lastFailedGenerationErrors
      type:
        list:
          elementType:
            scalar: string
          elementRelationship: atomic
    - name: name
      type:
        scalar: string
      default: ""
- name: com.github.openshift.api.machineconfiguration.v1alpha1.MachineConfigPoolReference
  map:
    fields:
    - name: name
      type:
        scalar: string
      default: ""
- name: com.github.openshift.api.machineconfiguration.v1alpha1.MachineOSBuild
  map:
    fields:
    - name: apiVersion
      type:
        scalar: string
    - name: kind
      type:
        scalar: string
    - name: metadata
      type:
        namedType: io.k8s.apimachinery.pkg.apis.meta.v1.ObjectMeta
      default: {}
    - name: spec
      type:
        namedType: com.github.openshift.api.machineconfiguration.v1alpha1.MachineOSBuildSpec
      default: {}
    - name: status
      type:
        namedType: com.github.openshift.api.machineconfiguration.v1alpha1.MachineOSBuildStatus
      default: {}
- name: com.github.openshift.api.machineconfiguration.v1alpha1.MachineOSBuildSpec
  map:
    fields:
    - name: configGeneration
      type:
        scalar: numeric
      default: 0
    - name: desiredConfig
      type:
        namedType: com.github.openshift.api.machineconfiguration.v1alpha1.RenderedMachineConfigReference
      default: {}
    - name: machineOSConfig
      type:
        namedType: com.github.openshift.api.machineconfiguration.v1alpha1.MachineOSConfigReference
      default: {}
    - name: renderedImagePushspec
      type:
        scalar: string
      default: ""
    - name: version
      type:
        scalar: numeric
      default: 0
- name: com.github.openshift.api.machineconfiguration.v1alpha1.MachineOSBuildStatus
  map:
    fields:
    - name: buildEnd
      type:
        namedType: io.k8s.apimachinery.pkg.apis.meta.v1.Time
    - name: buildStart
      type:
        namedType: io.k8s.apimachinery.pkg.apis.meta.v1.Time
    - name: builderReference
      type:
        namedType: com.github.openshift.api.machineconfiguration.v1alpha1.MachineOSBuilderReference
    - name: conditions
      type:
        list:
          elementType:
            namedType: io.k8s.apimachinery.pkg.apis.meta.v1.Condition
          elementRelationship: associative
          keys:
          - type
    - name: finalImagePullspec
      type:
        scalar: string
    - name: relatedObjects
      type:
        list:
          elementType:
            namedType: com.github.openshift.api.machineconfiguration.v1alpha1.ObjectReference
          elementRelationship: atomic
- name: com.github.openshift.api.machineconfiguration.v1alpha1.MachineOSBuilderReference
  map:
    fields:
    - name: buildPod
      type:
        namedType: com.github.openshift.api.machineconfiguration.v1alpha1.ObjectReference
    - name: imageBuilderType
      type:
        scalar: string
      default: ""
    unions:
    - discriminator: imageBuilderType
      fields:
      - fieldName: buildPod
        discriminatorValue: PodImageBuilder
- name: com.github.openshift.api.machineconfiguration.v1alpha1.MachineOSConfig
  map:
    fields:
    - name: apiVersion
      type:
        scalar: string
    - name: kind
      type:
        scalar: string
    - name: metadata
      type:
        namedType: io.k8s.apimachinery.pkg.apis.meta.v1.ObjectMeta
      default: {}
    - name: spec
      type:
        namedType: com.github.openshift.api.machineconfiguration.v1alpha1.MachineOSConfigSpec
      default: {}
    - name: status
      type:
        namedType: com.github.openshift.api.machineconfiguration.v1alpha1.MachineOSConfigStatus
      default: {}
- name: com.github.openshift.api.machineconfiguration.v1alpha1.MachineOSConfigReference
  map:
    fields:
    - name: name
      type:
        scalar: string
      default: ""
- name: com.github.openshift.api.machineconfiguration.v1alpha1.MachineOSConfigSpec
  map:
    fields:
    - name: buildInputs
      type:
        namedType: com.github.openshift.api.machineconfiguration.v1alpha1.BuildInputs
      default: {}
    - name: buildOutputs
      type:
        namedType: com.github.openshift.api.machineconfiguration.v1alpha1.BuildOutputs
      default: {}
    - name: machineConfigPool
      type:
        namedType: com.github.openshift.api.machineconfiguration.v1alpha1.MachineConfigPoolReference
      default: {}
- name: com.github.openshift.api.machineconfiguration.v1alpha1.MachineOSConfigStatus
  map:
    fields:
    - name: conditions
      type:
        list:
          elementType:
            namedType: io.k8s.apimachinery.pkg.apis.meta.v1.Condition
          elementRelationship: associative
          keys:
          - type
    - name: currentImagePullspec
      type:
        scalar: string
    - name: observedGeneration
      type:
        scalar: numeric
- name: com.github.openshift.api.machineconfiguration.v1alpha1.MachineOSContainerfile
  map:
    fields:
    - name: containerfileArch
      type:
        scalar: string
      default: ""
    - name: content
      type:
        scalar: string
      default: ""
- name: com.github.openshift.api.machineconfiguration.v1alpha1.MachineOSImageBuilder
  map:
    fields:
    - name: imageBuilderType
      type:
        scalar: string
      default: ""
- name: com.github.openshift.api.machineconfiguration.v1alpha1.ObjectReference
  map:
    fields:
    - name: group
      type:
        scalar: string
      default: ""
    - name: name
      type:
        scalar: string
      default: ""
    - name: namespace
      type:
        scalar: string
    - name: resource
      type:
        scalar: string
      default: ""
- name: com.github.openshift.api.machineconfiguration.v1alpha1.PinnedImageRef
  map:
    fields:
    - name: name
      type:
        scalar: string
      default: ""
- name: com.github.openshift.api.machineconfiguration.v1alpha1.PinnedImageSet
  map:
    fields:
    - name: apiVersion
      type:
        scalar: string
    - name: kind
      type:
        scalar: string
    - name: metadata
      type:
        namedType: io.k8s.apimachinery.pkg.apis.meta.v1.ObjectMeta
      default: {}
    - name: spec
      type:
        namedType: com.github.openshift.api.machineconfiguration.v1alpha1.PinnedImageSetSpec
      default: {}
    - name: status
      type:
        namedType: com.github.openshift.api.machineconfiguration.v1alpha1.PinnedImageSetStatus
      default: {}
- name: com.github.openshift.api.machineconfiguration.v1alpha1.PinnedImageSetSpec
  map:
    fields:
    - name: pinnedImages
      type:
        list:
          elementType:
            namedType: com.github.openshift.api.machineconfiguration.v1alpha1.PinnedImageRef
          elementRelationship: associative
          keys:
          - name
- name: com.github.openshift.api.machineconfiguration.v1alpha1.PinnedImageSetStatus
  map:
    fields:
    - name: conditions
      type:
        list:
          elementType:
            namedType: io.k8s.apimachinery.pkg.apis.meta.v1.Condition
          elementRelationship: associative
          keys:
          - type
- name: com.github.openshift.api.machineconfiguration.v1alpha1.RenderedMachineConfigReference
  map:
    fields:
    - name: name
      type:
        scalar: string
      default: ""
- name: io.k8s.apimachinery.pkg.apis.meta.v1.Condition
  map:
    fields:
    - name: lastTransitionTime
      type:
        namedType: io.k8s.apimachinery.pkg.apis.meta.v1.Time
    - name: message
      type:
        scalar: string
      default: ""
    - name: observedGeneration
      type:
        scalar: numeric
    - name: reason
      type:
        scalar: string
      default: ""
    - name: status
      type:
        scalar: string
      default: ""
    - name: type
      type:
        scalar: string
      default: ""
- name: io.k8s.apimachinery.pkg.apis.meta.v1.FieldsV1
  map:
    elementType:
      scalar: untyped
      list:
        elementType:
          namedType: __untyped_atomic_
        elementRelationship: atomic
      map:
        elementType:
          namedType: __untyped_deduced_
        elementRelationship: separable
- name: io.k8s.apimachinery.pkg.apis.meta.v1.ManagedFieldsEntry
  map:
    fields:
    - name: apiVersion
      type:
        scalar: string
    - name: fieldsType
      type:
        scalar: string
    - name: fieldsV1
      type:
        namedType: io.k8s.apimachinery.pkg.apis.meta.v1.FieldsV1
    - name: manager
      type:
        scalar: string
    - name: operation
      type:
        scalar: string
    - name: subresource
      type:
        scalar: string
    - name: time
      type:
        namedType: io.k8s.apimachinery.pkg.apis.meta.v1.Time
- name: io.k8s.apimachinery.pkg.apis.meta.v1.ObjectMeta
  map:
    fields:
    - name: annotations
      type:
        map:
          elementType:
            scalar: string
    - name: creationTimestamp
      type:
        namedType: io.k8s.apimachinery.pkg.apis.meta.v1.Time
    - name: deletionGracePeriodSeconds
      type:
        scalar: numeric
    - name: deletionTimestamp
      type:
        namedType: io.k8s.apimachinery.pkg.apis.meta.v1.Time
    - name: finalizers
      type:
        list:
          elementType:
            scalar: string
          elementRelationship: associative
    - name: generateName
      type:
        scalar: string
    - name: generation
      type:
        scalar: numeric
    - name: labels
      type:
        map:
          elementType:
            scalar: string
    - name: managedFields
      type:
        list:
          elementType:
            namedType: io.k8s.apimachinery.pkg.apis.meta.v1.ManagedFieldsEntry
          elementRelationship: atomic
    - name: name
      type:
        scalar: string
    - name: namespace
      type:
        scalar: string
    - name: ownerReferences
      type:
        list:
          elementType:
            namedType: io.k8s.apimachinery.pkg.apis.meta.v1.OwnerReference
          elementRelationship: associative
          keys:
          - uid
    - name: resourceVersion
      type:
        scalar: string
    - name: selfLink
      type:
        scalar: string
    - name: uid
      type:
        scalar: string
- name: io.k8s.apimachinery.pkg.apis.meta.v1.OwnerReference
  map:
    fields:
    - name: apiVersion
      type:
        scalar: string
      default: ""
    - name: blockOwnerDeletion
      type:
        scalar: boolean
    - name: controller
      type:
        scalar: boolean
    - name: kind
      type:
        scalar: string
      default: ""
    - name: name
      type:
        scalar: string
      default: ""
    - name: uid
      type:
        scalar: string
      default: ""
    elementRelationship: atomic
- name: io.k8s.apimachinery.pkg.apis.meta.v1.Time
  scalar: untyped
- name: __untyped_atomic_
  scalar: untyped
  list:
    elementType:
      namedType: __untyped_atomic_
    elementRelationship: atomic
  map:
    elementType:
      namedType: __untyped_atomic_
    elementRelationship: atomic
- name: __untyped_deduced_
  scalar: untyped
  list:
    elementType:
      namedType: __untyped_atomic_
    elementRelationship: atomic
  map:
    elementType:
      namedType: __untyped_deduced_
    elementRelationship: separable
`)
