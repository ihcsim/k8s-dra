package kubelet

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	draclientset "github.com/ihcsim/k8s-dra/pkg/apis/clientset/versioned"
	gpuv1alpha1 "github.com/ihcsim/k8s-dra/pkg/apis/gpu/v1alpha1"
	cdi "github.com/ihcsim/k8s-dra/pkg/drivers/gpu/kubelet/cdi"
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
	clientSets    draclientset.Interface
	log           zlog.Logger
	namespace     string
	nodeName      string
	availableGPUs int
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
	availableGPUs int,
	log zlog.Logger) (*NodeServer, error) {
	gpus := discoverGPUs(availableGPUs)
	var gpuDevices []*gpuv1alpha1.GPUDevice
	for _, gpu := range gpus {
		gpuDevices = append(gpuDevices, gpu.device)
	}

	if _, err := clientSets.GpuV1alpha1().NodeDevices(namespace).Get(ctx, nodeName, metav1.GetOptions{}); err != nil && apierrs.IsNotFound(err) {
		nodeDevices := &gpuv1alpha1.NodeDevices{
			ObjectMeta: metav1.ObjectMeta{
				Name:      nodeName,
				Namespace: namespace,
			},
			Spec: gpuv1alpha1.NodeDevicesSpec{
				AllocatableGPUs: gpuDevices,
			},
			Status: gpuv1alpha1.NodeDevicesStatus{
				State: gpuv1alpha1.Ready,
			},
		}
		if _, err := clientSets.GpuV1alpha1().NodeDevices(namespace).Create(ctx, nodeDevices, metav1.CreateOptions{}); err != nil {
			return nil, err
		}
	}

	cdi.InitRegistryOnce(cdiRoot)

	logger := log.With().Str("namespace", namespace).Logger()
	return &NodeServer{
		clientSets:    clientSets,
		log:           logger,
		namespace:     namespace,
		nodeName:      nodeName,
		availableGPUs: availableGPUs,
	}, nil
}

// NodePrepareResources prepares several ResourceClaims for use on the node.
// see https://pkg.go.dev/k8s.io/kubelet/pkg/apis/dra/v1alpha3#NodeServer
func (n *NodeServer) NodePrepareResources(ctx context.Context, req *kubeletdrav1.NodePrepareResourcesRequest) (*kubeletdrav1.NodePrepareResourcesResponse, error) {
	res := &kubeletdrav1.NodePrepareResourcesResponse{
		Claims: map[string]*kubeletdrav1.NodePrepareResourceResponse{},
	}

	for _, claim := range req.Claims {
		claimUID := claim.GetUid()
		res.Claims[claimUID] = n.nodePrepareResource(ctx, claimUID)
	}

	return res, nil
}

func (n *NodeServer) nodePrepareResource(ctx context.Context, claimUID string) *kubeletdrav1.NodePrepareResourceResponse {
	var (
		preparedGPUs = []*gpuv1alpha1.GPUDevice{}
		res          = &kubeletdrav1.NodePrepareResourceResponse{}
	)

	if err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		nodeDevices, err := n.clientSets.GpuV1alpha1().NodeDevices(n.namespace).Get(ctx, n.nodeName, metav1.GetOptions{})
		if err != nil {
			return err
		}

		var errs error
		for _, allocatedGPU := range nodeDevices.Status.AllocatedGPUs[claimUID] {
			// if the allocated GPU is already prepared, add its CDI qualified name to
			// the response
			if _, exists := nodeDevices.Status.PreparedGPUs[claimUID]; exists {
				res.CDIDevices = append(res.CDIDevices, cdi.DeviceQualifiedName(allocatedGPU))
				continue
			}

			// otherwise, if the allocated GPU is still allocatable, mark it as prepared
			var found bool
			for _, allocatableGPU := range nodeDevices.Spec.AllocatableGPUs {
				if allocatedGPU.UUID == allocatableGPU.UUID {
					found = true
					nodeDevices.Status.PreparedGPUs[claimUID] = append(nodeDevices.Status.PreparedGPUs[claimUID], allocatedGPU)
					preparedGPUs = append(preparedGPUs, allocatedGPU)
					res.CDIDevices = append(res.CDIDevices, cdi.DeviceQualifiedName(allocatedGPU))
				}
			}

			if !found {
				errs = errors.Join(errs, fmt.Errorf("allocated GPU %s is no longer allocatable", allocatedGPU.UUID))
			}
		}
		if errs != nil {
			return errs
		}

		if _, err := n.clientSets.GpuV1alpha1().NodeDevices(n.namespace).Update(ctx, nodeDevices, metav1.UpdateOptions{}); err != nil {
			return err
		}

		if len(preparedGPUs) > 0 {
			if err := cdi.CreateClaimSpecFile(claimUID, preparedGPUs); err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		res.Error = err.Error()
		return res
	}

	return res
}

// NodeUnprepareResources is the opposite of NodePrepareResources.
// see https://pkg.go.dev/k8s.io/kubelet/pkg/apis/dra/v1alpha3#NodeServer
func (n *NodeServer) NodeUnprepareResources(ctx context.Context, req *kubeletdrav1.NodeUnprepareResourcesRequest) (*kubeletdrav1.NodeUnprepareResourcesResponse, error) {
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
	if err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		nodeDevices, err := n.clientSets.GpuV1alpha1().NodeDevices(n.namespace).Get(ctx, n.nodeName, metav1.GetOptions{})
		if err != nil {
			return err
		}

		// nothing to unprepare
		if _, exists := nodeDevices.Status.PreparedGPUs[claimUID]; !exists {
			return nil
		}

		if err := cdi.DeleteClaimSpecFile(claimUID); err != nil {
			return err
		}

		delete(nodeDevices.Status.PreparedGPUs, claimUID)
		if _, err := n.clientSets.GpuV1alpha1().NodeDevices(n.namespace).Update(ctx, nodeDevices, metav1.UpdateOptions{}); err != nil {
			return err
		}
		return nil
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

type gpu struct {
	device *gpuv1alpha1.GPUDevice
}

func discoverGPUs(maxAvailableGPU int) []*gpu {
	gpus := make([]*gpu, maxAvailableGPU)
	for i := 0; i < maxAvailableGPU; i++ {
		gpus = append(gpus, &gpu{
			device: &gpuv1alpha1.GPUDevice{
				UUID:        uuid.NewString(),
				ProductName: "NVIDIA Tesla V100",
			},
		})
	}

	return gpus
}
