// nolint:lll
// Generates the LightStep adapter's resource yaml.
//go:generate $GOPATH/src/istio.io/istio/bin/mixer_codegen.sh -a mixer/adapter/lightstep/config/config.proto -x "-s=false -n lightstep -t tracespan"

package lightstep

import (
	"context"
	"fmt"
	"net"

	"istio.io/api/mixer/adapter/model/v1beta1"
	"istio.io/istio/mixer/template/tracespan"

	"google.golang.org/grpc"

)

type AdapterOptions struct {
	Server ServerOptions
	Client ClientOptions
}

// Adapter supports tracespan template.
type Adapter struct {
	listener net.Listener
	server   *grpc.Server
	client   *grpcSatelliteClient
}

var _ tracespan.HandleTraceSpanServiceServer = &Adapter{}

// HandleTraceSpan records TraceSpan entries
func (s *Adapter) HandleTraceSpan(
	ctx context.Context,
	request *tracespan.HandleTraceSpanRequest,
) (*v1beta1.ReportResult, error) {
	return s.client.HandleTraceSpan(ctx, request)
}

// NewLightStepAdapter creates a new GRPC adapter that listens at provided port.
func NewLightStepAdapter(opts AdapterOptions) (*Adapter, error) {
	serverAddress := "0"
	if opts.Server.Address != "" {
		serverAddress = opts.Server.Address
	}
	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", serverAddress))
	if err != nil {
		return nil, fmt.Errorf("unable to listen on socket: %v", err)
	}
	adapter := &Adapter{
		listener: listener,
	}
	fmt.Printf("listening on \"%v\"\n", adapter.Addr())
	adapter.server = grpc.NewServer()
	tracespan.RegisterHandleTraceSpanServiceServer(adapter.server, adapter)

	satelliteClient, err := newGRPCSatelliteClient(opts.Client)
	if err != nil {
		return nil, err
	}
	adapter.client = satelliteClient
	return adapter, nil
}
