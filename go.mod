module github.com/go-a2a/a2a

go 1.24

replace google.golang.org/a2a => ./grpc/google.golang.org/a2a/

require (
	github.com/go-json-experiment/json v0.0.0-20250714165856-be8212f5270d
	github.com/google/go-cmp v0.7.0
	google.golang.org/a2a v0.0.0-00010101000000-000000000000
	google.golang.org/protobuf v1.36.6
)

require (
	golang.org/x/net v0.42.0 // indirect
	golang.org/x/sys v0.34.0 // indirect
	golang.org/x/text v0.27.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20250707201910-8d1bb00bc6a7 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250603155806-513f23925822 // indirect
	google.golang.org/grpc v1.73.0 // indirect
)
