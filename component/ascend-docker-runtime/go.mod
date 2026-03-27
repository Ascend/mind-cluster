module ascend-docker-runtime

go 1.18

require (
	ascend-common v0.0.0
	github.com/agiledragon/gomonkey/v2 v2.8.0
	github.com/containerd/containerd v1.7.20
	github.com/opencontainers/runtime-spec v1.1.0
	github.com/pelletier/go-toml v1.9.5
	github.com/prashantv/gostub v0.0.0-00010101000000-000000000000
	github.com/smartystreets/goconvey v1.6.4
	github.com/stretchr/testify v1.8.4
)

require (
	dario.cat/mergo v1.0.0 // indirect
	github.com/Microsoft/go-winio v0.6.2 // indirect
	github.com/Microsoft/hcsshim v0.11.7 // indirect
	github.com/containerd/cgroups v1.1.0 // indirect
	github.com/containerd/continuity v0.4.2 // indirect
	github.com/containerd/errdefs v0.1.0 // indirect
	github.com/containerd/log v0.1.0 // indirect
	github.com/containerd/platforms v0.2.1 // indirect
	github.com/containerd/ttrpc v1.2.5 // indirect
	github.com/containerd/typeurl/v2 v2.1.1 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/docker/go-events v0.0.0-20190806004212-e31b211e4f1c // indirect
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/gopherjs/gopherjs v0.0.0-20181017120253-0766667cb4d1 // indirect
	github.com/jtolds/gls v4.20.0+incompatible // indirect
	github.com/klauspost/compress v1.16.7 // indirect
	github.com/moby/sys/mountinfo v0.6.2 // indirect
	github.com/moby/sys/user v0.1.0 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/smartystreets/assertions v0.0.0-20180927180507-b2de0cb4f26d // indirect
	go.opencensus.io v0.24.0 // indirect
	golang.org/x/net v0.24.0 // indirect
	golang.org/x/sync v0.5.0 // indirect
	golang.org/x/sys v0.19.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20231212172506-995d672761c0 // indirect
	google.golang.org/grpc v1.59.0 // indirect
	google.golang.org/protobuf v1.33.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/apimachinery v0.26.2 // indirect
	k8s.io/utils v0.0.0-20230220204549-a5ecb0141aa5 // indirect
)

replace (
	ascend-common => ../ascend-common
	github.com/prashantv/gostub => github.com/prashantv/gostub v1.0.1-0.20191007164320-bbe3712b9c4a
)
