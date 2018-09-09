package lightstep

import (
	"fmt"

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

	// HostPort is the address of the remote service that will receive reports.
	HostPort string
}

func newGRPCSatelliteClient(options ClientOptions) (*grpcSatelliteClient, error) {
	rec := &grpcSatelliteClient{options: options}

	conn, err := grpc.Dial(options.HostPort)
	if err != nil {
		return nil, fmt.Errorf("could not connect to satellite: %v", err)
	}
	rec.grpcClient = collectorpb.NewCollectorServiceClient(conn)
	return rec, nil
}

func (client *grpcSatelliteClient) Report(
	ctx context.Context,
	req *collectorpb.ReportRequest,
	opts ...grpc.CallOption,
) (*collectorpb.ReportResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("reportRequest cannot be nil")
	}
	return client.grpcClient.Report(ctx, req)
}
