package main

import (
	"encoding/json"
	"fmt"
	"os"

	_ "embed"

	mcfgv1 "github.com/openshift/api/machineconfiguration/v1"
	"k8s.io/klog/v2"
)

//go:embed usbguard_ipsec.conf
var usbguardIpsecCfg []byte

func doUSBGuard() error {
	usbguardFile := "/etc/lib/tmpfiles.d/usbguard.conf"
	_, err := os.Stat(usbguardFile)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("could not determine if %s exists", usbguardFile)
	}

	if err == nil {
		if err := os.RemoveAll(usbguardFile); err != nil {
			return fmt.Errorf("could not remove usbguard file %s: %w", usbguardFile, err)
		} else {
			klog.Infof("Removed %s", usbguardFile)
		}
	}

	usbguardIpsecCfgFile := "/usr/lib/tmpfiles.d/usbguard_ipsec.conf"

	if err := os.WriteFile(usbguardIpsecCfgFile, usbguardIpsecCfg, 0o644); err != nil {
		return fmt.Errorf("could not write %s: %w", usbguardIpsecCfgFile, err)
	}

	klog.Infof("Wrote %s", usbguardIpsecCfgFile)

	return nil
}

func doUpdateCATrust() error { return nil }

func loadMachineConfig(path string) (*mcfgv1.MachineConfig, error) {
	inBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not read MachineConfig from %s: %w", path, err)
	}

	mc := &mcfgv1.MachineConfig{}
	if err := json.Unmarshal(inBytes, mc); err != nil {
		return nil, fmt.Errorf("could not decode MachineConfig from %s: %w", path, err)
	}

	klog.Infof("Loaded MachineConfig %s from %s", mc.Name, path)

	return mc, nil
}

func entrypoint() error {
	machineconfigPath := "/etc/machine-config-daemon/currentconfig"

	mc, err := loadMachineConfig(machineconfigPath)
	if err != nil {
		return fmt.Errorf("could not load MachineConfig: %w", err)
	}

	ia := &ignitionApplicator{}
	if err := ia.ApplyIgnition(mc); err != nil {
		return fmt.Errorf("could not apply ignition: %w", err)
	}

	pa := &packageApplicator{}

	// mc.Spec.KernelType = ctrlcommon.KernelTypeRealtime
	if err := pa.InstallPackages(mc); err != nil {
		return fmt.Errorf("could not do package installation: %w", err)
	}

	if err := doUSBGuard(); err != nil {
		return fmt.Errorf("could not perform usbguard checks: %w", err)
	}

	if err := doUpdateCATrust(); err != nil {
		return fmt.Errorf("could not update CA trust: %w", err)
	}

	return nil
}

func main() {
	if err := entrypoint(); err != nil {
		panic(err)
	}
}
