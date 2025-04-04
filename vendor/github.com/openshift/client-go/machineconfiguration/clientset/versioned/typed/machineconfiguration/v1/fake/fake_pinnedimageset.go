// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	v1 "github.com/openshift/api/machineconfiguration/v1"
	machineconfigurationv1 "github.com/openshift/client-go/machineconfiguration/applyconfigurations/machineconfiguration/v1"
	typedmachineconfigurationv1 "github.com/openshift/client-go/machineconfiguration/clientset/versioned/typed/machineconfiguration/v1"
	gentype "k8s.io/client-go/gentype"
)

// fakePinnedImageSets implements PinnedImageSetInterface
type fakePinnedImageSets struct {
	*gentype.FakeClientWithListAndApply[*v1.PinnedImageSet, *v1.PinnedImageSetList, *machineconfigurationv1.PinnedImageSetApplyConfiguration]
	Fake *FakeMachineconfigurationV1
}

func newFakePinnedImageSets(fake *FakeMachineconfigurationV1) typedmachineconfigurationv1.PinnedImageSetInterface {
	return &fakePinnedImageSets{
		gentype.NewFakeClientWithListAndApply[*v1.PinnedImageSet, *v1.PinnedImageSetList, *machineconfigurationv1.PinnedImageSetApplyConfiguration](
			fake.Fake,
			"",
			v1.SchemeGroupVersion.WithResource("pinnedimagesets"),
			v1.SchemeGroupVersion.WithKind("PinnedImageSet"),
			func() *v1.PinnedImageSet { return &v1.PinnedImageSet{} },
			func() *v1.PinnedImageSetList { return &v1.PinnedImageSetList{} },
			func(dst, src *v1.PinnedImageSetList) { dst.ListMeta = src.ListMeta },
			func(list *v1.PinnedImageSetList) []*v1.PinnedImageSet { return gentype.ToPointerSlice(list.Items) },
			func(list *v1.PinnedImageSetList, items []*v1.PinnedImageSet) {
				list.Items = gentype.FromPointerSlice(items)
			},
		),
		fake,
	}
}
