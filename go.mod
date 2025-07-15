module github.com/go-a2a/a2a

go 1.24

replace google.golang.org/a2a => ./grpc/google.golang.org/a2a/

require (
	github.com/go-json-experiment/json v0.0.0-20250714165856-be8212f5270d
	github.com/google/go-cmp v0.7.0
	github.com/google/uuid v1.6.0
	github.com/lestrrat-go/jwx/v3 v3.0.8
	google.golang.org/a2a v0.0.0-00010101000000-000000000000
	google.golang.org/protobuf v1.36.6
	gorm.io/gorm v1.30.0
)

require (
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.4.0 // indirect
	github.com/goccy/go-json v0.10.3 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/lestrrat-go/blackmagic v1.0.4 // indirect
	github.com/lestrrat-go/httpcc v1.0.1 // indirect
	github.com/lestrrat-go/httprc/v3 v3.0.0 // indirect
	github.com/lestrrat-go/option v1.0.1 // indirect
	github.com/lestrrat-go/option/v2 v2.0.0 // indirect
	github.com/segmentio/asm v1.2.0 // indirect
	github.com/valyala/fastjson v1.6.4 // indirect
	golang.org/x/crypto v0.40.0 // indirect
	golang.org/x/net v0.42.0 // indirect
	golang.org/x/sys v0.34.0 // indirect
	golang.org/x/text v0.27.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20250707201910-8d1bb00bc6a7 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250603155806-513f23925822 // indirect
	google.golang.org/grpc v1.73.0 // indirect
)
