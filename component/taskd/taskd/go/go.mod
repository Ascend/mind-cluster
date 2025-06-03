module taskd

go 1.21

require (
	ascend-common v0.0.0
	clusterd v0.0.0-00010101000000-000000000000
	github.com/agiledragon/gomonkey/v2 v2.12.0
	github.com/golang/protobuf v1.5.4
	github.com/google/uuid v1.3.0
	github.com/smartystreets/goconvey v1.8.1
	github.com/stretchr/testify v1.10.0
	golang.org/x/time v0.3.0
	google.golang.org/grpc v1.57.2
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/gopherjs/gopherjs v1.17.2 // indirect
	github.com/jtolds/gls v4.20.0+incompatible // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/smarty/assertions v1.15.0 // indirect
	golang.org/x/net v0.35.0 // indirect
	golang.org/x/sys v0.30.0 // indirect
	golang.org/x/text v0.22.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240318140521-94a12d6c2237 // indirect
	google.golang.org/protobuf v1.36.5 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace ascend-common => ../../../ascend-common

replace clusterd => ../../../clusterd
