// Package bootc contains the machine-config-operator node controller's
// integration with the bootc-operator (node.bootc.dev API group).
//
// When bootc node management is enabled, the MCO no longer performs the OS image
// rebase itself. Instead, for each MachineConfigPool the controller maintains a
// corresponding, always-paused BootcNodePool. The bootc-operator daemon fetches
// and stages the target OS image on each node while the node is still online (no
// drain, no reboot), reporting progress via the BootcNode's status.staged field.
//
// The MCO retains full ownership of rollout pacing, draining, and rebooting: the
// node controller only selects a node for update once its target image is staged
// (see Reconciler.NodeStagedOrBooted). The MCD then drains the node, applies the
// non-OS config, and reboots into the already-staged deployment as part of its
// normal update flow. Keeping the reboot MCD-owned avoids racing with a
// bootc-driven reboot. The MCO therefore never sets desiredImageState to
// "Booted"; the pool stays paused and BootcNodes stay at "Staged".
package bootc

import (
	"context"
	"fmt"

	bootcv1alpha1 "github.com/bootc-dev/bootc-operator/api/v1alpha1"
	bootcclientset "github.com/bootc-dev/bootc-operator/pkg/generated/clientset/versioned"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"

	mcfgv1 "github.com/openshift/api/machineconfiguration/v1"
)

const (
	// PoolLabel is applied to BootcNodePool objects created by the MCO to
	// associate them back to the owning MachineConfigPool.
	PoolLabel = "machineconfiguration.openshift.io/machine-config-pool"

	// managedByValue marks resources managed by the MCO node controller.
	managedByLabel = "app.kubernetes.io/managed-by"
	managedByValue = "machine-config-operator"
)

// Reconciler encapsulates the BootcNodePool/BootcNode reconciliation performed
// by the node controller when bootc node management is enabled.
type Reconciler struct {
	client bootcclientset.Interface
}

// NewReconciler returns a Reconciler backed by the given bootc clientset.
func NewReconciler(client bootcclientset.Interface) *Reconciler {
	return &Reconciler{client: client}
}

// ReconcilePool ensures a BootcNodePool exists for the given MachineConfigPool,
// mirroring its node selector and target OS image. The pool is always created
// and kept paused: the MCO drives rollout pacing itself, so the bootc-operator
// must never start its own reboot slots. targetImage is the resolved OS image
// pullspec for the pool (e.g. ControllerConfig.BaseOSContainerImage or the
// rendered MachineConfig's OSImageURL).
func (r *Reconciler) ReconcilePool(ctx context.Context, pool *mcfgv1.MachineConfigPool, targetImage string) error {
	if targetImage == "" {
		klog.V(4).Infof("bootc: pool %s has no resolved OS image yet; skipping BootcNodePool reconcile", pool.Name)
		return nil
	}

	name := poolName(pool.Name)
	desired := r.desiredPool(pool, name, targetImage)

	existing, err := r.client.NodeV1alpha1().Bootcnodepools().Get(ctx, name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		klog.Infof("bootc: creating paused BootcNodePool %q for MachineConfigPool %q (image=%s)", name, pool.Name, targetImage)
		_, err = r.client.NodeV1alpha1().Bootcnodepools().Create(ctx, desired, metav1.CreateOptions{})
		return err
	}
	if err != nil {
		return fmt.Errorf("getting BootcNodePool %q: %w", name, err)
	}

	// Reconcile the fields the MCO owns. Keep it paused and in sync with the
	// pool's selector and target image.
	updated := existing.DeepCopy()
	updated.Spec.NodeSelector = pool.Spec.NodeSelector.DeepCopy()
	updated.Spec.Image = bootcv1alpha1.ImageSpec{Ref: targetImage}
	ensurePaused(updated)

	if poolSpecEqual(existing, updated) {
		return nil
	}

	klog.Infof("bootc: updating BootcNodePool %q (image=%s, paused=true)", name, targetImage)
	_, err = r.client.NodeV1alpha1().Bootcnodepools().Update(ctx, updated, metav1.UpdateOptions{})
	return err
}

// NodeStagedOrBooted reports whether the bootc-operator has finished preparing
// the node's target OS image, i.e. the image is either already staged for the
// next boot or already booted. This is the gate the MCO node controller uses
// before selecting a node for update: once the target image is staged, the MCD
// can drain the node, apply its config, and reboot into the staged deployment
// itself. This keeps the reboot owned by the MCD (as in the non-bootc flow) and
// avoids racing with a bootc-driven reboot.
//
// The booted case is included so a node that is already on the target image
// (e.g. it was staged+rebooted previously, or joined already up to date) is not
// blocked from proceeding through the rest of its (non-OS) config update.
//
// Returns false (not an error) when the BootcNodePool or BootcNode does not yet
// exist or has not reported a target/staged digest, so the caller simply waits
// and requeues.
func (r *Reconciler) NodeStagedOrBooted(ctx context.Context, poolMCName, nodeName string) (bool, error) {
	np, err := r.client.NodeV1alpha1().Bootcnodepools().Get(ctx, poolName(poolMCName), metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if np.Status.TargetDigest == "" {
		return false, nil
	}

	bn, err := r.client.NodeV1alpha1().Bootcnodes().Get(ctx, nodeName, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	if bn.Status.Staged != nil && bn.Status.Staged.ImageDigest == np.Status.TargetDigest {
		return true, nil
	}
	if bn.Status.Booted != nil && bn.Status.Booted.ImageDigest == np.Status.TargetDigest {
		return true, nil
	}
	return false, nil
}

func (r *Reconciler) desiredPool(pool *mcfgv1.MachineConfigPool, name, targetImage string) *bootcv1alpha1.BootcNodePool {
	np := &bootcv1alpha1.BootcNodePool{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				PoolLabel:      pool.Name,
				managedByLabel: managedByValue,
			},
		},
		Spec: bootcv1alpha1.BootcNodePoolSpec{
			NodeSelector: pool.Spec.NodeSelector.DeepCopy(),
			Image:        bootcv1alpha1.ImageSpec{Ref: targetImage},
		},
	}
	ensurePaused(np)
	return np
}

// ensurePaused sets spec.rollout.paused=true, allocating the rollout struct if
// necessary. The MCO always keeps the pool paused so the bootc-operator never
// initiates its own reboot slots/drains.
func ensurePaused(np *bootcv1alpha1.BootcNodePool) {
	if np.Spec.Rollout == nil {
		np.Spec.Rollout = &bootcv1alpha1.RolloutSpec{}
	}
	np.Spec.Rollout.Paused = true
}

func poolSpecEqual(a, b *bootcv1alpha1.BootcNodePool) bool {
	if a.Spec.Image.Ref != b.Spec.Image.Ref {
		return false
	}
	if rolloutPaused(a) != rolloutPaused(b) {
		return false
	}
	return selectorEqual(a.Spec.NodeSelector, b.Spec.NodeSelector)
}

func rolloutPaused(np *bootcv1alpha1.BootcNodePool) bool {
	return np.Spec.Rollout != nil && np.Spec.Rollout.Paused
}

func selectorEqual(a, b *metav1.LabelSelector) bool {
	if a == nil || b == nil {
		return a == b
	}
	if len(a.MatchLabels) != len(b.MatchLabels) {
		return false
	}
	for k, v := range a.MatchLabels {
		if b.MatchLabels[k] != v {
			return false
		}
	}
	return len(a.MatchExpressions) == len(b.MatchExpressions)
}

// poolName derives the BootcNodePool name for a given MachineConfigPool name.
func poolName(mcpName string) string {
	return mcpName
}

// Ensure corev1 stays imported for potential future node-based helpers.
var _ = corev1.Node{}
