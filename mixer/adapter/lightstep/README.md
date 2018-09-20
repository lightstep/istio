# Test Instructions
Follow build instructions here: https://github.com/istio/istio/wiki/Mixer-Out-Of-Process-Adapter-Walkthrough

## Tab 1 
```bash
go run cmd/main.go -serverAddress=38355 -accessToken={PUBLIC_LIGHTSTEP_ACCESS_TOKEN} -socketAddress=collector-grpc.lightstep.com:8080
```

## Tab 2 
```bash
$GOPATH/out/darwin_amd64/release/mixs server --configStoreURL=fs://$(pwd)/mixer/adapter/lightstep/testdata
```

## Tab 3 
```bash
pushd $ISTIO/istio && make mixc && $GOPATH/out/darwin_amd64/release/mixc report -s destination.service="svc.cluster.local" --stringmap_attributes request.headers=x-b3-traceid:463ac35c9f6413ad\;x-b3-spanid:a2fb4a1d1a96d312 -t request.time=2006-01-02T15:04:05Z,response.time=2006-01-02T15:04:25Z
```

## Current State
Receiving this error
```
Report RPC returned Unknown (1 error occurred:

* rpc error: code = Unknown desc = failed to send report: rpc error: code = Unknown desc = Invalid runtime in Report (missing or empty guid))
```