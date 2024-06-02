package kubelet

import (
	"context"

	draclientset "github.com/ihcsim/k8s-dra/pkg/apis/clientset/versioned"
	gpuv1alpha1 "github.com/ihcsim/k8s-dra/pkg/apis/gpu/v1alpha1"
	"github.com/ihcsim/k8s-dra/pkg/drivers/gpu/kubelet/cdi"
	zlog "github.com/rs/zerolog"
	resourcev1alpha2 "k8s.io/api/resource/v1alpha2"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
	kubeletdrav1 "k8s.io/kubelet/pkg/apis/dra/v1alpha3"
)

const AvailableGPUsCount = 4

var _ kubeletdrav1.NodeServer = &NodeServer{}

// NodeServer provides the API implementation of the node server.
// see https://pkg.go.dev/k8s.io/kubelet/pkg/apis/dra/v1alpha3#NodeServer
type NodeServer struct {
	clientSets draclientset.Interface
	log        zlog.Logger
	namespace  string
	nodeName   string
}

// NewNodeServer returns a new instance of the NodeServer. It also initializes
// the associated	NodeDevices object, which defines the device specs of the
// corresponding node.
func NewNodeServer(
	ctx context.Context,
	clientSets draclientset.Interface,
	cdiRoot string,
	namespace string,
	nodeName string,
	log zlog.Logger) (*NodeServer, error) {
	logger := log.With().Str("namespace", namespace).Logger()
	logger.Info().Msg("initializing CDI registry and discovering CDI devices...")
	cdi.InitRegistryOnce(cdiRoot)
	gpus, err := cdi.DiscoverFromSpecs()
	if err != nil {
		return nil, err
	}
	logger.Info().Msgf("discovered %d CDI devices", len(gpus))

	gpuDevices := make([]*gpuv1alpha1.GPUDevice, len(gpus))
	for i, gpu := range gpus {
		gpuDevices[i] = &gpuv1alpha1.GPUDevice{
			UUID:        gpu.UUID,
			ProductName: gpu.ProductName,
			Vendor:      gpu.VendorName,
		}
	}

	if _, err := clientSets.GpuV1alpha1().NodeDevices(namespace).Get(ctx, nodeName, metav1.GetOptions{}); err != nil && apierrs.IsNotFound(err) {
		nodeDevices := &gpuv1alpha1.NodeDevices{
			ObjectMeta: metav1.ObjectMeta{
				Name:      nodeName,
				Namespace: namespace,
			},
			AllocatableGPUs: gpuDevices,
			Allocations:     map[string][]*gpuv1alpha1.DeviceAllocation{},
			NodeSuitability: map[string]gpuv1alpha1.NodeSuitability{},
		}
		logger.Info().Msgf("creating new NodeDevices %s...", nodeName)
		if _, err := clientSets.GpuV1alpha1().NodeDevices(namespace).Create(ctx, nodeDevices, metav1.CreateOptions{}); err != nil {
			return nil, err
		}
	}

	return &NodeServer{
		clientSets: clientSets,
		log:        logger,
		namespace:  namespace,
		nodeName:   nodeName,
	}, nil
}

// NodePrepareResources prepares several ResourceClaims for use on the node.
// see https://pkg.go.dev/k8s.io/kubelet/pkg/apis/dra/v1alpha3#NodeServer
func (n *NodeServer) NodePrepareResources(ctx context.Context, req *kubeletdrav1.NodePrepareResourcesRequest) (*kubeletdrav1.NodePrepareResourcesResponse, error) {
	n.log.Info().Msg("preparing resources...")
	res := &kubeletdrav1.NodePrepareResourcesResponse{
		Claims: map[string]*kubeletdrav1.NodePrepareResourceResponse{},
	}

	nodeDevices, err := n.clientSets.GpuV1alpha1().NodeDevices(n.namespace).Get(ctx, n.nodeName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	for _, claim := range req.Claims {
		claimUID := claim.GetUid()
		res.Claims[claimUID] = n.nodePrepareResource(ctx, nodeDevices, claimUID)
	}

	return res, nil
}

func (n *NodeServer) nodePrepareResource(ctx context.Context, nodeDevices *gpuv1alpha1.NodeDevices, claimUID string) *kubeletdrav1.NodePrepareResourceResponse {
	var (
		cdiDevices = []*cdi.GPUDevice{}
		res        = &kubeletdrav1.NodePrepareResourceResponse{}
		log        = n.log.With().Str("claim", claimUID).Logger()
	)

	claimAllocations, exists := nodeDevices.Allocations[claimUID]
	if !exists {
		log.Info().Msg("no device allocation found")
		return &kubeletdrav1.NodePrepareResourceResponse{}
	}

	for i, claimAllocation := range claimAllocations {
		if claimAllocation.State != gpuv1alpha1.DeviceAllocationStateAllocated && claimAllocation.State != gpuv1alpha1.DeviceAllocationStatePrepared {
			log.Info().Msg("device allocation is not in either allocated or protected state")
			return &kubeletdrav1.NodePrepareResourceResponse{}
		}

		var (
			device    = claimAllocation.Device
			cdiDevice = &cdi.GPUDevice{
				UUID:        device.UUID,
				ProductName: device.ProductName,
				VendorName:  device.Vendor,
			}
			qualifiedName = cdi.DeviceQualifiedName(cdiDevice)
		)

		log.Info().
			Str("deviceUUID", device.UUID).
			Str("deviceProductName", device.ProductName).
			Str("deviceVendor", device.Vendor).
			Str("deviceState", string(claimAllocation.State)).
			Str("qualifiedName", qualifiedName).
			Msg("preparing CDI device...")
		res.CDIDevices = append(res.CDIDevices, qualifiedName)
		cdiDevices = append(cdiDevices, cdiDevice)

		if claimAllocation.State == gpuv1alpha1.DeviceAllocationStateAllocated {
			claimAllocationClone := claimAllocation.DeepCopy()
			claimAllocationClone.State = gpuv1alpha1.DeviceAllocationStatePrepared

			nodeDevicesClone := nodeDevices.DeepCopy()
			nodeDevicesClone.Allocations[claimUID][i] = claimAllocationClone

			if err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
				_, err := n.clientSets.GpuV1alpha1().NodeDevices(n.namespace).Update(ctx, nodeDevicesClone, metav1.UpdateOptions{})
				return err
			}); err != nil {
				res.Error = err.Error()
				return res
			}
		}
	}

	if len(cdiDevices) > 0 {
		if err := cdi.CreateCDISpec(claimUID, cdiDevices); err != nil {
			res.Error = err.Error()
		}
	}

	return res
}

// NodeUnprepareResources is the opposite of NodePrepareResources.
// see https://pkg.go.dev/k8s.io/kubelet/pkg/apis/dra/v1alpha3#NodeServer
func (n *NodeServer) NodeUnprepareResources(ctx context.Context, req *kubeletdrav1.NodeUnprepareResourcesRequest) (*kubeletdrav1.NodeUnprepareResourcesResponse, error) {
	n.log.Info().Msg("unpreparing resources...")
	res := &kubeletdrav1.NodeUnprepareResourcesResponse{
		Claims: map[string]*kubeletdrav1.NodeUnprepareResourceResponse{},
	}

	for _, claim := range req.Claims {
		claimUID := claim.GetUid()
		res.Claims[claimUID] = n.nodeUnprepareResource(ctx, claimUID)
	}

	return res, nil
}

func (n *NodeServer) nodeUnprepareResource(ctx context.Context, claimUID string) *kubeletdrav1.NodeUnprepareResourceResponse {
	nodeDevices, err := n.clientSets.GpuV1alpha1().NodeDevices(n.namespace).Get(ctx, n.nodeName, metav1.GetOptions{})
	if err != nil {
		return &kubeletdrav1.NodeUnprepareResourceResponse{
			Error: err.Error(),
		}
	}

	if _, exists := nodeDevices.Allocations[claimUID]; !exists {
		n.log.Info().Msg("no device allocation found, skipping resource unpreparation...")
		return &kubeletdrav1.NodeUnprepareResourceResponse{}
	}

	n.log.Info().Str("claimUID", claimUID).Msg("unpreparing claim allocations...")
	delete(nodeDevices.Allocations, claimUID)
	if err := cdi.DeleteCDISpec(claimUID); err != nil {
		return &kubeletdrav1.NodeUnprepareResourceResponse{
			Error: err.Error(),
		}
	}

	if err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		_, err := n.clientSets.GpuV1alpha1().NodeDevices(n.namespace).Update(ctx, nodeDevices, metav1.UpdateOptions{})
		return err
	}); err != nil {
		return &kubeletdrav1.NodeUnprepareResourceResponse{
			Error: err.Error(),
		}
	}

	return &kubeletdrav1.NodeUnprepareResourceResponse{}
}

// NodeListAndWatchResources returns a stream of NodeResourcesResponse objects.
// see https://pkg.go.dev/k8s.io/kubelet/pkg/apis/dra/v1alpha3#NodeServer
func (n *NodeServer) NodeListAndWatchResources(req *kubeletdrav1.NodeListAndWatchResourcesRequest, s kubeletdrav1.Node_NodeListAndWatchResourcesServer) error {
	namedResources := &resourcev1alpha2.NamedResourcesResources{
		Instances: []resourcev1alpha2.NamedResourcesInstance{
			{Name: "test-named-resource-instance"},
		},
	}
	resources := []*resourcev1alpha2.ResourceModel{
		{NamedResources: namedResources},
	}
	res := &kubeletdrav1.NodeListAndWatchResourcesResponse{
		Resources: resources,
	}
	return s.Send(res)
}
