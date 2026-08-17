package common

import (
	"os"
	"strconv"
)

// BootcNodeManagementEnvVar is the environment variable used, in addition to the
// --enable-bootc-node-management CLI flag, to enable delegation of OS image
// rollout to the bootc-operator. This is a PoC gating mechanism; productization
// should replace it with a real openshift/api FeatureGate.
const BootcNodeManagementEnvVar = "MCO_BOOTC_NODE_MANAGEMENT"

// BootcNodeManagementEnabledFromEnv reports whether the bootc node management
// integration is enabled via the environment variable. It is used as the default
// value for the --enable-bootc-node-management flag so the behavior can be
// toggled from the component Deployment/DaemonSet without argument plumbing.
func BootcNodeManagementEnabledFromEnv() bool {
	v, ok := os.LookupEnv(BootcNodeManagementEnvVar)
	if !ok {
		return false
	}
	enabled, err := strconv.ParseBool(v)
	if err != nil {
		return false
	}
	return enabled
}
