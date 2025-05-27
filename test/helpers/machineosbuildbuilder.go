package helpers

import (
	"fmt"

	mcfgv1 "github.com/openshift/api/machineconfiguration/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type MachineOSBuildBuilder struct {
	mosb *mcfgv1.MachineOSBuild
}

func NewMachineOSBuildBuilder(name string) *MachineOSBuildBuilder {
	return &MachineOSBuildBuilder{
		mosb: &mcfgv1.MachineOSBuild{
			TypeMeta: metav1.TypeMeta{
				Kind:       "MachineOSBuild",
				APIVersion: "machineconfiguration.openshift.io/v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:        name,
				Labels:      map[string]string{},
				Annotations: map[string]string{},
			},
			Spec: mcfgv1.MachineOSBuildSpec{
				MachineConfig:   mcfgv1.MachineConfigReference{},
				MachineOSConfig: mcfgv1.MachineOSConfigReference{},
			},
		},
	}
}

func NewMachineOSBuildBuilderFromMachineConfigPool(mcp *mcfgv1.MachineConfigPool) *MachineOSBuildBuilder {
	m := NewMachineOSBuildBuilder(fmt.Sprintf("%s-%s-builder", mcp.Name, mcp.Spec.Configuration.Name))
	m.mosb.Spec.MachineConfig.Name = mcp.Spec.Configuration.Name
	return m
}

func (m *MachineOSBuildBuilder) WithName(name string) *MachineOSBuildBuilder {
	m.mosb.Name = name
	return m
}

func (m *MachineOSBuildBuilder) WithRenderedImagePushspec(pushspec string) *MachineOSBuildBuilder {
	m.mosb.Spec.RenderedImagePushSpec = mcfgv1.ImageTagFormat(pushspec)
	return m
}

func (m *MachineOSBuildBuilder) WithMachineOSConfig(name string) *MachineOSBuildBuilder {
	m.mosb.Spec.MachineOSConfig.Name = name
	return m
}

func (m *MachineOSBuildBuilder) WithDesiredConfig(name string) *MachineOSBuildBuilder {
	m.mosb.Spec.MachineConfig.Name = name
	return m
}

func (m *MachineOSBuildBuilder) WithAnnotations(annos map[string]string) *MachineOSBuildBuilder {
	for k, v := range annos {
		m.mosb.Annotations[k] = v
	}

	return m
}

func (m *MachineOSBuildBuilder) WithLabels(labels map[string]string) *MachineOSBuildBuilder {
	for k, v := range labels {
		m.mosb.Labels[k] = v
	}

	return m
}

func (m *MachineOSBuildBuilder) MachineOSBuild() *mcfgv1.MachineOSBuild {
	return m.mosb.DeepCopy()
}
