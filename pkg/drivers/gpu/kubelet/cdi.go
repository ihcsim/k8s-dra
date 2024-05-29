package kubelet

import (
	"fmt"

	cdiapi "github.com/container-orchestrated-devices/container-device-interface/pkg/cdi"
	cdispec "github.com/container-orchestrated-devices/container-device-interface/specs-go"
	gpuv1alpha1 "github.com/ihcsim/k8s-dra/pkg/apis/gpu/v1alpha1"
)

const (
	cdiVendor = "resources.ihcsim"
	cdiClass  = "gpu"
)

var cdiKind = cdiVendor + "/" + cdiClass

func initCDIRegistry(cdiRoot string) cdiapi.Registry {
	return cdiapi.GetRegistry(cdiapi.WithSpecDirs(cdiRoot))
}

func cdiQualifiedName(gpu *gpuv1alpha1.GPUDevice) string {
	return cdiapi.QualifiedName(cdiVendor, cdiClass, gpu.UUID)
}

func createClaimSpecFile(r cdiapi.Registry, claimUID string, gpus []*gpuv1alpha1.GPUDevice) error {
	specName := cdiapi.GenerateTransientSpecName(cdiVendor, cdiClass, claimUID)
	spec := &cdispec.Spec{
		Kind:    cdiKind,
		Devices: []cdispec.Device{},
	}

	for _, gpu := range gpus {
		cdiDevice := cdispec.Device{
			Name: gpu.UUID,
			ContainerEdits: cdispec.ContainerEdits{
				Env: []string{
					fmt.Sprintf("GPU_DEVICE_UUID=%s", gpu.UUID),
					fmt.Sprintf("GPU_DEVICE_PRODUCT_NAME=%s", gpu.ProductName),
				},
			},
		}
		spec.Devices = append(spec.Devices, cdiDevice)
	}

	minVersion, err := cdiapi.MinimumRequiredVersion(spec)
	if err != nil {
		return fmt.Errorf("failed to get minimum required CDI spec version: %v", err)
	}
	spec.Version = minVersion

	return r.SpecDB().WriteSpec(spec, specName)
}