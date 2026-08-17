package bootc

import (
	"context"
	"testing"

	bootcv1alpha1 "github.com/bootc-dev/bootc-operator/api/v1alpha1"
	bootcfake "github.com/bootc-dev/bootc-operator/pkg/generated/clientset/versioned/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	mcfgv1 "github.com/openshift/api/machineconfiguration/v1"
)

func testPool(name string) *mcfgv1.MachineConfigPool {
	return &mcfgv1.MachineConfigPool{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: mcfgv1.MachineConfigPoolSpec{
			NodeSelector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"node-role.kubernetes.io/worker": ""},
			},
		},
	}
}

func TestReconcilePoolCreatesPausedPool(t *testing.T) {
	client := bootcfake.NewSimpleClientset()
	r := NewReconciler(client)

	if err := r.ReconcilePool(context.Background(), testPool("worker"), "quay.io/os@sha256:abc"); err != nil {
		t.Fatalf("ReconcilePool: %v", err)
	}

	np, err := client.NodeV1alpha1().Bootcnodepools().Get(context.Background(), "worker", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("expected BootcNodePool to be created: %v", err)
	}
	if np.Spec.Image.Ref != "quay.io/os@sha256:abc" {
		t.Errorf("unexpected image ref: %q", np.Spec.Image.Ref)
	}
	if np.Spec.Rollout == nil || !np.Spec.Rollout.Paused {
		t.Errorf("expected pool to be paused, got rollout=%+v", np.Spec.Rollout)
	}
	if np.Labels[PoolLabel] != "worker" {
		t.Errorf("expected pool label %q=worker, got %q", PoolLabel, np.Labels[PoolLabel])
	}
}

func TestReconcilePoolEmptyImageIsNoop(t *testing.T) {
	client := bootcfake.NewSimpleClientset()
	r := NewReconciler(client)

	if err := r.ReconcilePool(context.Background(), testPool("worker"), ""); err != nil {
		t.Fatalf("ReconcilePool: %v", err)
	}
	list, err := client.NodeV1alpha1().Bootcnodepools().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list.Items) != 0 {
		t.Errorf("expected no pool created for empty image, got %d", len(list.Items))
	}
}

func TestReconcilePoolKeepsExistingPaused(t *testing.T) {
	// A pre-existing pool that somehow got unpaused must be re-paused.
	existing := &bootcv1alpha1.BootcNodePool{
		ObjectMeta: metav1.ObjectMeta{Name: "worker"},
		Spec: bootcv1alpha1.BootcNodePoolSpec{
			Image:   bootcv1alpha1.ImageSpec{Ref: "quay.io/os@sha256:old"},
			Rollout: &bootcv1alpha1.RolloutSpec{Paused: false},
		},
	}
	client := bootcfake.NewSimpleClientset(existing)
	r := NewReconciler(client)

	if err := r.ReconcilePool(context.Background(), testPool("worker"), "quay.io/os@sha256:new"); err != nil {
		t.Fatalf("ReconcilePool: %v", err)
	}
	np, err := client.NodeV1alpha1().Bootcnodepools().Get(context.Background(), "worker", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if np.Spec.Image.Ref != "quay.io/os@sha256:new" {
		t.Errorf("expected image updated to new digest, got %q", np.Spec.Image.Ref)
	}
	if np.Spec.Rollout == nil || !np.Spec.Rollout.Paused {
		t.Errorf("expected pool re-paused")
	}
}

// stagedTestObjects builds a BootcNodePool with the given target digest and a
// BootcNode with the given staged/booted digests (empty string = nil).
func stagedTestObjects(poolMCName, nodeName, targetDigest, stagedDigest, bootedDigest string) []runtime.Object {
	np := &bootcv1alpha1.BootcNodePool{
		ObjectMeta: metav1.ObjectMeta{Name: poolName(poolMCName)},
		Status:     bootcv1alpha1.BootcNodePoolStatus{TargetDigest: targetDigest},
	}
	bn := &bootcv1alpha1.BootcNode{
		ObjectMeta: metav1.ObjectMeta{Name: nodeName},
	}
	if stagedDigest != "" {
		bn.Status.Staged = &bootcv1alpha1.ImageInfo{ImageDigest: stagedDigest}
	}
	if bootedDigest != "" {
		bn.Status.Booted = &bootcv1alpha1.ImageInfo{ImageDigest: bootedDigest}
	}
	return []runtime.Object{np, bn}
}

func TestNodeStagedOrBooted(t *testing.T) {
	const target = "sha256:target"
	tests := []struct {
		name    string
		objs    []runtime.Object
		want    bool
		wantErr bool
	}{
		{
			name: "staged with target digest",
			objs: stagedTestObjects("worker", "node-1", target, target, ""),
			want: true,
		},
		{
			name: "already booted into target digest",
			objs: stagedTestObjects("worker", "node-1", target, "", target),
			want: true,
		},
		{
			name: "staged with a different digest",
			objs: stagedTestObjects("worker", "node-1", target, "sha256:other", ""),
			want: false,
		},
		{
			name: "nothing staged or booted yet",
			objs: stagedTestObjects("worker", "node-1", target, "", ""),
			want: false,
		},
		{
			name: "pool has no target digest yet",
			objs: stagedTestObjects("worker", "node-1", "", target, ""),
			want: false,
		},
		{
			name: "missing pool is not ready (no error)",
			objs: []runtime.Object{&bootcv1alpha1.BootcNode{ObjectMeta: metav1.ObjectMeta{Name: "node-1"}}},
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			client := bootcfake.NewSimpleClientset(tc.objs...)
			r := NewReconciler(client)
			got, err := r.NodeStagedOrBooted(context.Background(), "worker", "node-1")
			if (err != nil) != tc.wantErr {
				t.Fatalf("NodeStagedOrBooted err = %v, wantErr %v", err, tc.wantErr)
			}
			if got != tc.want {
				t.Errorf("NodeStagedOrBooted = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestNodeStagedOrBootedMissingBootcNodeIsNotReady(t *testing.T) {
	// Pool exists with a target digest but the BootcNode has not been created yet.
	np := &bootcv1alpha1.BootcNodePool{
		ObjectMeta: metav1.ObjectMeta{Name: poolName("worker")},
		Status:     bootcv1alpha1.BootcNodePoolStatus{TargetDigest: "sha256:target"},
	}
	client := bootcfake.NewSimpleClientset(np)
	r := NewReconciler(client)
	got, err := r.NodeStagedOrBooted(context.Background(), "worker", "does-not-exist")
	if err != nil {
		t.Fatalf("expected no error for missing BootcNode, got %v", err)
	}
	if got {
		t.Errorf("expected not-ready for missing BootcNode, got ready")
	}
}
