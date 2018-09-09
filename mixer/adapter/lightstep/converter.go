package lightstep

import (
	"fmt"
	"net"
	"strconv"
	"time"

	"istio.io/api/policy/v1beta1"
	"istio.io/istio/mixer/template/tracespan"

	"github.com/gogo/protobuf/types"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"

	"github.com/lightstep/lightstep-tracer-go"
	"github.com/lightstep/lightstep-tracer-go/collectorpb"

	"github.com/openzipkin/zipkin-go/model"
)

func convertRequest(
	request *tracespan.HandleTraceSpanRequest,
) (*collectorpb.ReportRequest, error) {
	if request == nil {
		return nil, nil
	}

	spans, err := convertSpans(request.Instances)
	if err != nil {
		return nil, err
	}

	return &collectorpb.ReportRequest{
		Reporter: convertReporter(request.AdapterConfig),
		Spans:    spans,
	}, nil
}

func convertReporter(
	adapterConfig *types.Any,
) *collectorpb.Reporter {
	if adapterConfig == nil {
		return nil
	}

	return &collectorpb.Reporter{}
}

func convertSpans(messages []*tracespan.InstanceMsg) ([]*collectorpb.Span, error) {
	if len(messages) == 0 {
		return nil, nil
	}

	convertedSpans := make([]*collectorpb.Span, 0, len(messages))
	for _, message := range messages {
		convertedSpan, err := convertSpan(message)
		if err != nil {
			return nil, err
		}
		convertedSpans = append(convertedSpans, convertedSpan)
	}

	return convertedSpans, nil
}

func convertSpan(message *tracespan.InstanceMsg) (*collectorpb.Span, error) {
	if message == nil {
		return nil, nil
	}

	spanContext, err := convertSpanContext(message.TraceId, message.SpanId)
	if err != nil {
		return nil, err
	}

	references, err := convertReferences(message.TraceId, message.ParentSpanId)
	if err != nil {
		return nil, err
	}

	startTime, err := convertTimeStamp(message.StartTime)
	if err != nil {
		return nil, err
	}

	duration, err := convertDuration(message.StartTime, message.EndTime)
	if err != nil {
		return nil, err
	}

	tags, err := convertTags(message.Name, message.ParentSpanId, message.SpanTags)
	if err != nil {
		return nil, err
	}

	return &collectorpb.Span{
		SpanContext:    spanContext,
		OperationName:  message.SpanName,
		References:     references,
		StartTimestamp: startTime,
		DurationMicros: duration,
		Tags:           tags,
	}, nil
}

func convertSpanContext(traceIDHex string, spanIDHex string) (*collectorpb.SpanContext, error) {
	traceID, err := convertTraceID(traceIDHex)
	if err != nil {
		return nil, err
	}

	spanID, err := convertSpanID(spanIDHex)
	if err != nil {
		return nil, err
	}

	return &collectorpb.SpanContext{
		TraceId: traceID,
		SpanId:  spanID,
	}, nil
}

func convertTraceID(traceIDHex string) (uint64, error) {
	traceID, err := model.TraceIDFromHex(traceIDHex)
	if err != nil {
		return 0, err
	}

	if traceID.High > 0 {
		return 0, fmt.Errorf("128 bit trace ids not supported by LightStep")
	}

	return traceID.Low, nil
}

func convertSpanID(spanID string) (uint64, error) {
	return strconv.ParseUint(spanID, 16, 64)
}

func convertReferences(traceID string, parentSpanID string) ([]*collectorpb.Reference, error) {
	if len(parentSpanID) == 0 {
		return nil, nil
	}

	parentSpanContext, err := convertSpanContext(traceID, parentSpanID)
	if err != nil {
		return nil, err
	}

	return []*collectorpb.Reference{
		{
			Relationship: collectorpb.Reference_CHILD_OF,
			SpanContext:  parentSpanContext,
		},
	}, nil
}

func convertTimeStamp(input *v1beta1.TimeStamp) (*timestamp.Timestamp, error) {
	if input == nil {
		return nil, nil
	}

	inputTime, err := types.TimestampFromProto(input.Value)
	if err != nil {
		return nil, nil
	}

	output, err := ptypes.TimestampProto(inputTime)
	if err != nil {
		return nil, err
	}

	return output, nil
}

func convertDuration(
	start *v1beta1.TimeStamp,
	end *v1beta1.TimeStamp,
) (uint64, error) {
	if start == nil || start.Value == nil || end == nil || end.Value == nil {
		return 0, fmt.Errorf("start and end time expected")
	}

	startTime, err := types.TimestampFromProto(start.Value)
	if err != nil {
		return 0, err
	}

	endTime, err := types.TimestampFromProto(end.Value)
	if err != nil {
		return 0, err
	}

	duration := endTime.Sub(startTime)
	if duration < 0 {
		return 0, fmt.Errorf("expected start time to be before end time")
	}

	return uint64(duration / time.Microsecond), nil
}

func convertTags(
	componentName string,
	parentSpanGUID string,
	input map[string]*v1beta1.Value,
) ([]*collectorpb.KeyValue, error) {
	if _, ok := input[lightstep.ComponentNameKey]; !ok {
		input[lightstep.ComponentNameKey] = &v1beta1.Value{
			Value: &v1beta1.Value_StringValue{
				StringValue: componentName,
			},
		}
	}

	if _, ok := input[lightstep.ParentSpanGUIDKey]; !ok {
		input[lightstep.ParentSpanGUIDKey] = &v1beta1.Value{
			Value: &v1beta1.Value_StringValue{
				StringValue: parentSpanGUID,
			},
		}
	}

	output := make([]*collectorpb.KeyValue, 0, len(input))
	for key, value := range input {
		keyValue, err := convertKeyValue(key, value)
		if err != nil {
			return nil, err
		}
		output = append(output, keyValue)
	}

	return output, nil
}

func convertKeyValue(key string, value *v1beta1.Value) (*collectorpb.KeyValue, error) {
	switch v := value.Value.(type) {
	case *v1beta1.Value_StringValue:
		return &collectorpb.KeyValue{
			Key: key,
			Value: &collectorpb.KeyValue_StringValue{
				StringValue: v.StringValue,
			},
		}, nil
	case *v1beta1.Value_Int64Value:
		return &collectorpb.KeyValue{
			Key: key,
			Value: &collectorpb.KeyValue_IntValue{
				IntValue: v.Int64Value,
			},
		}, nil
	case *v1beta1.Value_DoubleValue:
		return &collectorpb.KeyValue{
			Key: key,
			Value: &collectorpb.KeyValue_DoubleValue{
				DoubleValue: v.DoubleValue,
			},
		}, nil
	case *v1beta1.Value_BoolValue:
		return &collectorpb.KeyValue{
			Key: key,
			Value: &collectorpb.KeyValue_BoolValue{
				BoolValue: v.BoolValue,
			},
		}, nil
	case *v1beta1.Value_IpAddressValue:
		return &collectorpb.KeyValue{
			Key: key,
			Value: &collectorpb.KeyValue_StringValue{
				StringValue: convertIpAddress(v.IpAddressValue),
			},
		}, nil
	case *v1beta1.Value_TimestampValue:
		return &collectorpb.KeyValue{
			Key: key,
			Value: &collectorpb.KeyValue_StringValue{
				StringValue: convertTimestampValue(v.TimestampValue),
			},
		}, nil
	case *v1beta1.Value_DurationValue:
		return &collectorpb.KeyValue{
			Key: key,
			Value: &collectorpb.KeyValue_StringValue{
				StringValue: convertDurationValue(v.DurationValue),
			},
		}, nil
	case *v1beta1.Value_EmailAddressValue:
		return &collectorpb.KeyValue{
			Key: key,
			Value: &collectorpb.KeyValue_StringValue{
				StringValue: convertEmailAddress(v.EmailAddressValue),
			},
		}, nil
	case *v1beta1.Value_DnsNameValue:
		return &collectorpb.KeyValue{
			Key: key,
			Value: &collectorpb.KeyValue_StringValue{
				StringValue: convertDnsNameValue(v.DnsNameValue),
			},
		}, nil
	case *v1beta1.Value_UriValue:
		return &collectorpb.KeyValue{
			Key: key,
			Value: &collectorpb.KeyValue_StringValue{
				StringValue: v.UriValue.String(),
			},
		}, nil
	default:
		return nil, fmt.Errorf("unsupported type %T", value)
	}
}

func convertIpAddress(address *v1beta1.IPAddress) string {
	return net.IP(address.Value).String()
}

func convertTimestampValue(timestamp *v1beta1.TimeStamp) string {
	return types.TimestampString(timestamp.Value)
}

func convertDurationValue(durationProto *v1beta1.Duration) string {
	duration, err := types.DurationFromProto(durationProto.Value)
	if err != nil {
		return fmt.Sprintf("(%v)", err)
	}
	return duration.String()
}

func convertEmailAddress(address *v1beta1.EmailAddress) string {
	if address == nil {
		return ""
	}
	return address.Value
}

func convertDnsNameValue(name *v1beta1.DNSName) string {
	if name != nil {
		return ""
	}
	return name.Value
}
