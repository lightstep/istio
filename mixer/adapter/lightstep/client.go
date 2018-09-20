package lightstep

import (
	"fmt"
	"istio.io/api/mixer/adapter/model/v1beta1"
	"istio.io/istio/mixer/template/tracespan"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/lightstep/lightstep-tracer-go/collectorpb"
)

var _ collectorpb.CollectorServiceClient = &grpcSatelliteClient{}

// grpcSatelliteClient specifies how to send reports back to a LightStep
// Satellite via grpc.
type grpcSatelliteClient struct {
	options    ClientOptions
	grpcClient collectorpb.CollectorServiceClient
}

type ClientOptions struct {
	// ReporterID the ID of a particular Adapter client
	ReporterID uint64

	// AccessToken is the access token used for explicit trace collection requests.
	AccessToken string

	// SocketAddress is the address of the remote service that will receive reports.
	SocketAddress string
}

func newGRPCSatelliteClient(options ClientOptions) (*grpcSatelliteClient, error) {
	rec := &grpcSatelliteClient{options: options}

	conn, err := grpc.Dial(options.SocketAddress, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return nil, fmt.Errorf("could not connect to satellite: %v", err)
	}
	rec.grpcClient = collectorpb.NewCollectorServiceClient(conn)
	return rec, nil
}

// HandleTraceSpan records TraceSpan entries
func (c *grpcSatelliteClient) HandleTraceSpan(
	ctx context.Context,
	request *tracespan.HandleTraceSpanRequest,
) (*v1beta1.ReportResult, error) {
	satelliteRequest, err := convertRequest(request, c.options.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to send report: %v", err)
	}
	// TODO: do something with AdapterConfig and/or DedupId?
	// TODO: do something with response?
	_, err = c.Report(ctx, satelliteRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to send report: %v", err)
	}

	return &v1beta1.ReportResult{}, nil
}

func (c *grpcSatelliteClient) Report(
	ctx context.Context,
	req *collectorpb.ReportRequest,
	opts ...grpc.CallOption,
) (*collectorpb.ReportResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("reportRequest cannot be nil")
	}
	return c.grpcClient.Report(ctx, req)
}
