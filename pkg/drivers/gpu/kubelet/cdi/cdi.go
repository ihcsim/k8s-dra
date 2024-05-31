package cdi

import (
	"errors"
	"fmt"
	"sync"

	cdiapi "github.com/container-orchestrated-devices/container-device-interface/pkg/cdi"
	cdispec "github.com/container-orchestrated-devices/container-device-interface/specs-go"
	gpuv1alpha1 "github.com/ihcsim/k8s-dra/pkg/apis/gpu/v1alpha1"
)

const (
	cdiVendor = "resources.ihcsim"
	cdiClass  = "gpu"
)

var (
	cdiKind = cdiVendor + "/" + cdiClass

	registry cdiapi.Registry
	once     sync.Once
)

// GPUDevice is used to encapsulate the CDI information of a GPU device.
type GPUDevice struct {
	UUID        string
	ProductName string
	VendorName  string
}

func InitRegistryOnce(cdiRoot string) {
	once.Do(func() {
		registry = cdiapi.GetRegistry(cdiapi.WithSpecDirs(cdiRoot))
	})
}

func DiscoverFromSpecs() ([]*GPUDevice, error) {
	specs, err := Specs()
	if err != nil {
		return nil, err
	}

	var gpuDevices []*GPUDevice
	for _, spec := range specs {
		for _, device := range spec.Devices {
			gpuDevices = append(gpuDevices, &GPUDevice{
				UUID:        device.Name,
				ProductName: device.ContainerEdits.Env[1],
				VendorName:  device.ContainerEdits.Env[2],
			})
		}
	}

	return gpuDevices, nil
}

func DeviceQualifiedName(gpu *gpuv1alpha1.GPUDevice) string {
	return cdiapi.QualifiedName(cdiVendor, cdiClass, gpu.UUID)
}

func CreateCDISpec(claimUID string, gpus []*gpuv1alpha1.GPUDevice) error {
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

	return registry.SpecDB().WriteSpec(spec, specName)
}

func DeleteCDISpec(claimUID string) error {
	specName := cdiapi.GenerateTransientSpecName(cdiVendor, cdiClass, claimUID)
	return registry.SpecDB().RemoveSpec(specName)
}

func Specs() ([]*cdiapi.Spec, error) {
	specs := registry.SpecDB().GetVendorSpecs(cdiVendor)

	var errs error
	for _, spec := range specs {
		specErrs := registry.SpecDB().GetSpecErrors(spec)
		specErrs = append(specErrs, errs)
		errs = errors.Join(specErrs...)
	}
	return specs, errs
}
