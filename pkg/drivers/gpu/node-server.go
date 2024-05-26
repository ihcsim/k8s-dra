package gpu

import "context"

type NodeServer struct{}

func NewNodeServer(ctx context.Context) *NodeServer {
	return &NodeServer{}
}

func (n *NodeServer) Shutdown(ctx context.Context) error {
	return nil
}
