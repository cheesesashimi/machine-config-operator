# bootc-operator integration (proof of concept)

This directory documents the proof-of-concept integration between the
machine-config-operator (MCO) and the upstream
[bootc-operator](https://github.com/bootc-dev/bootc-operator).

When enabled, the MCO no longer drives the node OS image rebase directly.
Instead it delegates the OS switch + reboot to the bootc-operator via the
`node.bootc.dev` API group (`BootcNodePool` / `BootcNode`), while retaining
ownership of rollout pacing and node draining.

## Enabling the integration

The integration is gated (PoC only) by a CLI flag / environment variable on both
the machine-config-controller and the machine-config-daemon. It is **off by
default**.

Enable it by either:

- setting the flag `--enable-bootc-node-management=true`, or
- setting the environment variable `MCO_BOOTC_NODE_MANAGEMENT=true`

on the `machine-config-controller` Deployment and the `machine-config-daemon`
DaemonSet (see the commented lines in
`manifests/machineconfigcontroller/deployment.yaml` and
`manifests/machineconfigdaemon/daemonset.yaml`).

The bootc-operator itself is deployed separately, in its own container image and
its own `openshift-bootc-operator` namespace, using the manifests shipped in the
bootc-operator repository.

> NOTE: For productization this flag must be replaced with a real
> `FeatureGateBootcNodeManagement` FeatureGate defined in `openshift/api` and
> consumed via `fgHandler.Enabled(...)`.

## Rollout flow

1. A MachineConfigPool changes (new resolved `OSImageURL`).
2. The MCO node controller ensures a **paused** `BootcNodePool` exists for the
   pool, mirroring its node selector and target OS image.
3. The node controller selects candidate nodes using its existing
   `maxUnavailable` logic and sets the usual `desiredConfig`/`desiredImage`
   annotations for the non-OS parts of the update.
4. The MCD applies non-OS changes, **skips the rpm-ostree/bootc rebase**, and
   requests a drain via its normal annotation handshake.
5. The drain controller drains the node and records completion.
6. The node controller observes the completed drain and patches the node's
   `BootcNode` `spec.desiredImageState` to `Booted`.
7. The bootc-operator daemon applies the staged image and reboots the node.
8. Once `BootcNode.status.booted` reflects the pool's `targetDigest`, the node
   controller advances the pool and proceeds to the next node.

## Bootstrap / firstboot

During bootstrap and firstboot the bootc-operator is not yet running, so the MCD
**always** performs the OS pivot itself (`checkStateOnFirstRun` and
`RunFirstbootCompleteMachineconfig`). The delegation only applies to the
steady-state, cluster-connected update path.

## Files

- `worker.bootcnodepool.yaml` — reference example of the (auto-created) paused
  BootcNodePool for the `worker` pool.
