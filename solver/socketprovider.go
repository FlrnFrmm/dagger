package solver

import (
	"context"
	"fmt"
	"strings"

	"github.com/moby/buildkit/session"
	"github.com/moby/buildkit/session/sshforward"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const (
	unixPrefix  = "unix="
	npipePrefix = "npipe="
)

type SocketProvider struct {
}

func NewDockerSocketProvider() session.Attachable {
	return &SocketProvider{}
}

func (sp *SocketProvider) Register(server *grpc.Server) {
	sshforward.RegisterSSHServer(server, sp)
}

func (sp *SocketProvider) CheckAgent(ctx context.Context, req *sshforward.CheckAgentRequest) (*sshforward.CheckAgentResponse, error) {
	id := sshforward.DefaultID
	if req.ID != "" {
		id = req.ID
	}
	if !strings.HasPrefix(id, unixPrefix) && !strings.HasPrefix(id, npipePrefix) {
		return &sshforward.CheckAgentResponse{}, fmt.Errorf("invalid socket forward key %s", id)
	}
	return &sshforward.CheckAgentResponse{}, nil
}

func (sp *SocketProvider) ForwardAgent(stream sshforward.SSH_ForwardAgentServer) error {
	id := sshforward.DefaultID

	opts, _ := metadata.FromIncomingContext(stream.Context()) // if no metadata continue with empty object

	if v, ok := opts[sshforward.KeySSHID]; ok && len(v) > 0 && v[0] != "" {
		id = v[0]
	}

	conn, err := dialStream(id)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", id, err)
	}
	defer conn.Close()

	return sshforward.Copy(context.TODO(), conn, stream, nil)
}
