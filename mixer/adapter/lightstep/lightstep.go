package lightstep

import (
	"context"
	"fmt"
	"net"

	"istio.io/istio/mixer/template/tracespan"

	"google.golang.org/grpc"

	"istio.io/api/mixer/adapter/model/v1beta1"
)

// Server is basic server interface
type Server interface {
	Addr() string
	Close() error
	Run(shutdown chan error)
}

// Adapter supports metric template.
type Adapter struct {
	listener net.Listener
	server   *grpc.Server
}

var _ tracespan.HandleTraceSpanServiceServer = &Adapter{}

// HandleTraceSpan records TraceSpan entries
func (s *Adapter) HandleTraceSpan(
	ctx context.Context,
	request *tracespan.HandleTraceSpanRequest,
) (*v1beta1.ReportResult, error) {
	return nil, nil
}

// Addr returns the listening address of the server
func (s *Adapter) Addr() string {
	return s.listener.Addr().String()
}

// Run starts the server run
func (s *Adapter) Run(shutdown chan error) {
	shutdown <- s.server.Serve(s.listener)
}

// Close gracefully shuts down the server; used for testing
func (s *Adapter) Close() error {
	if s.server != nil {
		s.server.GracefulStop()
	}

	if s.listener != nil {
		_ = s.listener.Close()
	}

	return nil
}

// NewLightStepAdapter creates a new GRPC adapter that listens at provided port.
func NewLightStepAdapter(addr string) (Server, error) {
	if addr == "" {
		addr = "0"
	}
	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", addr))
	if err != nil {
		return nil, fmt.Errorf("unable to listen on socket: %v", err)
	}
	s := &Adapter{
		listener: listener,
	}
	fmt.Printf("listening on \"%v\"\n", s.Addr())
	s.server = grpc.NewServer()
	tracespan.RegisterHandleTraceSpanServiceServer(s.server, s)
	return s, nil
}
