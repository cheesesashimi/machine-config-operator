package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"k8s.io/klog/v2"
)

type rpmClient struct{}

func (r *rpmClient) IsInstalled(pkgs ...string) ([]string, bool, error) {
	args := append([]string{"-q"}, pkgs...)
	cmd := exec.Command("rpm", args...)

	klog.Infof("Running $ %s", cmd)

	output, err := cmd.CombinedOutput()
	if err == nil {
		return nil, true, nil
	}

	uninstalled := []string{}
	for _, pkg := range pkgs {
		if strings.Contains(string(output), fmt.Sprintf("package %s is not installed", pkg)) {
			uninstalled = append(uninstalled, pkg)
		}
	}

	if len(uninstalled) == 0 && err != nil {
		return nil, false, fmt.Errorf("could not determine if packages were installed: %w", err)
	}

	return uninstalled, false, nil
}

type rpmostreeClient struct{}

func (r *rpmostreeClient) InstallPackages(pkgs []string) error {
	args := append([]string{"install"}, pkgs...)

	if err := r.runRpmOstree(args); err != nil {
		return fmt.Errorf("could not install packages %v: %w", pkgs, err)
	}

	return nil
}

func (r *rpmostreeClient) RemovePackages(pkgs []string) error {
	args := append([]string{"uninstall"}, pkgs...)

	if err := r.runRpmOstree(args); err != nil {
		return fmt.Errorf("could not install packages %v: %w", pkgs, err)
	}

	return nil
}

func (r *rpmostreeClient) OverrideKernelPackages(originalKernelPkgs, desiredKernelPkgs []string) error {
	args := []string{"override", "remove"}
	args = append(args, originalKernelPkgs...)

	for _, pkg := range desiredKernelPkgs {
		args = append(args, "--install", pkg)
	}

	if err := r.runRpmOstree(args); err != nil {
		return fmt.Errorf("could not override kernel packages: %w", err)
	}

	return nil
}

func (r *rpmostreeClient) runRpmOstree(args []string) error {
	cmd := exec.Command("rpm-ostree", args...)
	klog.Infof("Running $ %s", cmd.String())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
