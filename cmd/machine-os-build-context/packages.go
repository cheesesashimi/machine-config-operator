package main

import (
	"fmt"
	"os"

	_ "embed"

	mcfgv1 "github.com/openshift/api/machineconfiguration/v1"
	ctrlcommon "github.com/openshift/machine-config-operator/pkg/controller/common"
	"github.com/openshift/machine-config-operator/pkg/helpers"
	"k8s.io/klog/v2"
)

//go:embed coreos-extensions.repo
var coreosExtensionsRepo []byte

const (
	coreosExtRepoPath string = "/etc/yum.repos.d/coreos-extensions.repo"
)

type packageApplicator struct{}

func (p *packageApplicator) InstallPackages(mc *mcfgv1.MachineConfig) error {
	kernelInstallNeeded, err := p.isKernelInstallNeeded(mc)
	if err != nil {
		return fmt.Errorf("could not determine if kernel installation is needed: %w", err)
	}

	if len(mc.Spec.Extensions) == 0 && !kernelInstallNeeded {
		klog.Infof("No extensions or kernel changes needed")
		return nil
	}

	klog.Infof("Writing extensions repo config to %s", coreosExtRepoPath)

	if err := os.WriteFile(coreosExtRepoPath, coreosExtensionsRepo, 0o644); err != nil {
		return fmt.Errorf("could not write coreos extensions repo: %w", err)
	}

	defer func() {
		if err := os.RemoveAll(coreosExtRepoPath); err != nil {
			klog.Errorf("could not remove coreos extensions repo path %s: %s", coreosExtRepoPath, err)
		} else {
			klog.Infof("Removed extensions repo config from %s", coreosExtRepoPath)
		}
	}()

	// If we know that we need both extensions and a kernel, we can install them in a single operation.
	if len(mc.Spec.Extensions) != 0 && kernelInstallNeeded {
		klog.Infof("Installing extensions and kernel")
		return p.installExtensionsAndKernel(mc)
	}

	if len(mc.Spec.Extensions) != 0 {
		klog.Infof("Installing packages for extensions %v", mc.Spec.Extensions)
		if err := p.installExtensions(mc); err != nil {
			return fmt.Errorf("could not install extensions: %w", err)
		}
	} else {
		klog.Infof("No extensions specified")
	}

	if kernelInstallNeeded {
		klog.Infof("Installing packages for kernel %s", mc.Spec.KernelType)
		if err := p.installKernel(mc); err != nil {
			return fmt.Errorf("could not install kernel: %w", err)
		}
	} else {
		klog.Infof("No kernel changes needed")
	}

	return nil

}

func (p *packageApplicator) installExtensionsAndKernel(mc *mcfgv1.MachineConfig) error {
	extPackages, err := ctrlcommon.GetPackagesForSupportedExtensions(mc.Spec.Extensions)
	if err != nil {
		return fmt.Errorf("could not get packages for supported extensions: %w", err)
	}

	kernelType, kernelPkgs, err := ctrlcommon.GetPackagesForSupportedKernelType(helpers.CanonicalizeKernelType(mc.Spec.KernelType))
	if err != nil {
		return fmt.Errorf("could not get kernel packages for install: %w", err)
	}

	rpmosc := &rpmostreeClient{}

	if err := rpmosc.OverrideKernelPackages(kernelPkgs[ctrlcommon.KernelTypeDefault], append(kernelPkgs[kernelType], extPackages...)); err != nil {
		return fmt.Errorf("could not install kernel packages: %w", err)
	}

	return nil
}

func (p *packageApplicator) installExtensions(mc *mcfgv1.MachineConfig) error {
	extPackages, err := ctrlcommon.GetPackagesForSupportedExtensions(mc.Spec.Extensions)
	if err != nil {
		return fmt.Errorf("could not get packages for supported extensions: %w", err)
	}

	rc := &rpmostreeClient{}
	return rc.InstallPackages(extPackages)
}

func (p *packageApplicator) installKernel(mc *mcfgv1.MachineConfig) error {
	kernelType, kernelPkgs, err := ctrlcommon.GetPackagesForSupportedKernelType(helpers.CanonicalizeKernelType(mc.Spec.KernelType))
	if err != nil {
		return fmt.Errorf("could not get kernel packages for install: %w", err)
	}

	rpmosc := &rpmostreeClient{}

	if err := rpmosc.OverrideKernelPackages(kernelPkgs[ctrlcommon.KernelTypeDefault], kernelPkgs[kernelType]); err != nil {
		return fmt.Errorf("could not install kernel packages: %w", err)
	}

	return nil
}

func (p *packageApplicator) isKernelInstallNeeded(mc *mcfgv1.MachineConfig) (bool, error) {
	kernelType, kernelPkgs, err := ctrlcommon.GetPackagesForSupportedKernelType(helpers.CanonicalizeKernelType(mc.Spec.KernelType))
	if err != nil {
		return false, fmt.Errorf("could not get kernel packages: %w", err)
	}

	rpmc := rpmClient{}

	_, allInstalled, err := rpmc.IsInstalled(kernelPkgs[kernelType]...)
	if err != nil {
		return false, fmt.Errorf("could not determine if %s was installed: %w", helpers.CanonicalizeKernelType(mc.Spec.KernelType), err)
	}

	return !allInstalled, nil
}
