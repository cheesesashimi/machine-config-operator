package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	mcfgv1 "github.com/openshift/api/machineconfiguration/v1"
	ctrlcommon "github.com/openshift/machine-config-operator/pkg/controller/common"
	"k8s.io/klog/v2"
)

type ignitionApplicator struct{}

func (i *ignitionApplicator) ApplyIgnition(mc *mcfgv1.MachineConfig) error {
	ignFile, err := i.storeIgnitionInTempFile(mc)
	if err != nil {
		return fmt.Errorf("could not store Ignition config in temp file: %w", err)
	}

	defer func() {
		if err := os.RemoveAll(ignFile); err != nil {
			klog.Errorf("could not remove temp ignition config file %s: %s", ignFile, err)
		}
	}()

	ignCmd := strings.Join([]string{"exec", "-a", "ignition-apply", "/usr/lib/dracut/modules.d/30ignition/ignition", "--ignore-unsupported", ignFile}, " ")

	cmd := exec.Command("/bin/bash", "-c", ignCmd)

	klog.Infof("Running $ %s", cmd.String())

	cmd.Env = append(os.Environ(), "container=oci")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("could not apply ignition: %w", err)
	}

	return nil
}

func (i *ignitionApplicator) storeIgnitionInTempFile(mc *mcfgv1.MachineConfig) (string, error) {
	cfg, err := ctrlcommon.ParseAndConvertConfig(mc.Spec.Config.Raw)
	if err != nil {
		return "", fmt.Errorf("could not decode Ignition from MachineConfig %s: %w", mc.Name, err)
	}

	ignJSONBytes, err := json.Marshal(cfg)
	if err != nil {
		return "", fmt.Errorf("could not marshal Ignition to JSON: %w", err)
	}

	return i.writeToTempFile(ignJSONBytes, "ign-config")
}

func (i *ignitionApplicator) writeToTempFile(b []byte, prefix string) (string, error) {
	tmpFile, err := os.CreateTemp("", prefix)
	if err != nil {
		return "", fmt.Errorf("could not create temp file: %w", err)
	}

	var retErr error

	defer func() {
		if retErr != nil {
			if err := os.RemoveAll(tmpFile.Name()); err != nil {
				klog.Errorf("could not clean up temp file %s: %s", tmpFile.Name(), err)
			}
		}
	}()

	if _, err := tmpFile.Write(b); err != nil {
		retErr = err
		return "", fmt.Errorf("could not write to temp file %s: %w", tmpFile.Name(), err)
	}

	if err := tmpFile.Close(); err != nil {
		retErr = err
		return "", fmt.Errorf("could not close temp file %s: %w", tmpFile.Name(), err)
	}

	return tmpFile.Name(), nil
}
