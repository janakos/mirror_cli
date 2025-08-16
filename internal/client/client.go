package client

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/alexcode/mirror_cli/internal/config"
	pb "github.com/alexcode/mirror_cli/proto/gen"
)

// Client wraps the gRPC client with convenience methods
type Client struct {
	conn       *grpc.ClientConn
	flowClient pb.FlowServiceClient
	config     *config.Config
}

// NewClient creates a new PeerDB gRPC client
func NewClient(cfg *config.Config) (*Client, error) {
	var opts []grpc.DialOption

	// Set up credentials
	if cfg.TLS {
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	// Add timeout
	opts = append(opts, grpc.WithTimeout(30*time.Second))

	// Connect to PeerDB
	conn, err := grpc.Dial(cfg.Address(), opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PeerDB at %s: %w", cfg.Address(), err)
	}

	return &Client{
		conn:       conn,
		flowClient: pb.NewFlowServiceClient(conn),
		config:     cfg,
	}, nil
}

// Close closes the gRPC connection
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// CreateCDCMirror creates a new CDC mirror
func (c *Client) CreateCDCMirror(ctx context.Context, req *pb.CreateCDCFlowRequest) (*pb.CreateCDCFlowResponse, error) {
	return c.flowClient.CreateCDCFlow(ctx, req)
}

// ListMirrors lists all mirrors
func (c *Client) ListMirrors(ctx context.Context) (*pb.ListMirrorsResponse, error) {
	return c.flowClient.ListMirrors(ctx, &pb.ListMirrorsRequest{})
}

// ListMirrorNames lists all mirror names
func (c *Client) ListMirrorNames(ctx context.Context) (*pb.ListMirrorNamesResponse, error) {
	return c.flowClient.ListMirrorNames(ctx, &pb.ListMirrorNamesRequest{})
}

// GetMirrorStatus gets the status of a specific mirror
func (c *Client) GetMirrorStatus(ctx context.Context, mirrorName string) (*pb.MirrorStatusResponse, error) {
	req := &pb.MirrorStatusRequest{
		FlowJobName:     mirrorName,
		IncludeFlowInfo: true,
		ExcludeBatches:  false,
	}
	return c.flowClient.MirrorStatus(ctx, req)
}

// PauseMirror pauses a mirror
func (c *Client) PauseMirror(ctx context.Context, mirrorName string) error {
	req := &pb.FlowStateChangeRequest{
		FlowJobName:        mirrorName,
		RequestedFlowState: pb.FlowStatus_STATUS_PAUSED,
	}
	_, err := c.flowClient.FlowStateChange(ctx, req)
	return err
}

// ResumeMirror resumes a mirror
func (c *Client) ResumeMirror(ctx context.Context, mirrorName string) error {
	req := &pb.FlowStateChangeRequest{
		FlowJobName:        mirrorName,
		RequestedFlowState: pb.FlowStatus_STATUS_RUNNING,
	}
	_, err := c.flowClient.FlowStateChange(ctx, req)
	return err
}

// DropMirror terminates and drops a mirror
func (c *Client) DropMirror(ctx context.Context, mirrorName string, skipDestinationDrop bool) error {
	req := &pb.FlowStateChangeRequest{
		FlowJobName:         mirrorName,
		RequestedFlowState:  pb.FlowStatus_STATUS_TERMINATED,
		DropMirrorStats:     true,
		SkipDestinationDrop: skipDestinationDrop,
	}
	_, err := c.flowClient.FlowStateChange(ctx, req)
	return err
}

// UpdateMirror updates mirror configuration
func (c *Client) UpdateMirror(ctx context.Context, mirrorName string, update *pb.FlowConfigUpdate) error {
	// First pause the mirror
	if err := c.PauseMirror(ctx, mirrorName); err != nil {
		return fmt.Errorf("failed to pause mirror: %w", err)
	}

	// Apply the update
	req := &pb.FlowStateChangeRequest{
		FlowJobName:        mirrorName,
		RequestedFlowState: pb.FlowStatus_STATUS_PAUSED,
		FlowConfigUpdate:   update,
	}

	if _, err := c.flowClient.FlowStateChange(ctx, req); err != nil {
		return fmt.Errorf("failed to update mirror configuration: %w", err)
	}

	// Resume the mirror
	if err := c.ResumeMirror(ctx, mirrorName); err != nil {
		return fmt.Errorf("failed to resume mirror after update: %w", err)
	}

	return nil
}

// ListPeers lists all peers
func (c *Client) ListPeers(ctx context.Context) (*pb.ListPeersResponse, error) {
	return c.flowClient.ListPeers(ctx, &pb.ListPeersRequest{})
}

// CreatePeer creates a new peer
func (c *Client) CreatePeer(ctx context.Context, peer *pb.Peer, allowUpdate bool) (*pb.CreatePeerResponse, error) {
	req := &pb.CreatePeerRequest{
		Peer:        peer,
		AllowUpdate: allowUpdate,
	}
	return c.flowClient.CreatePeer(ctx, req)
}

// DropPeer drops a peer
func (c *Client) DropPeer(ctx context.Context, peerName string) error {
	req := &pb.DropPeerRequest{
		PeerName: peerName,
	}
	_, err := c.flowClient.DropPeer(ctx, req)
	return err
}

// ValidatePeer validates a peer configuration
func (c *Client) ValidatePeer(ctx context.Context, peer *pb.Peer) (*pb.ValidatePeerResponse, error) {
	req := &pb.ValidatePeerRequest{
		Peer: peer,
	}
	return c.flowClient.ValidatePeer(ctx, req)
}
