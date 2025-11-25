module container-manager

go 1.18

require (
	ascend-common v0.0.0
	github.com/agiledragon/gomonkey/v2 v2.8.0
	github.com/gogo/protobuf v1.3.2
	github.com/smartystreets/goconvey v1.6.4
)

require (
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/gopherjs/gopherjs v0.0.0-20181017120253-0766667cb4d1 // indirect
	github.com/jtolds/gls v4.20.0+incompatible // indirect
	github.com/smartystreets/assertions v1.2.0 // indirect
	golang.org/x/sys v0.8.0 // indirect
	k8s.io/apimachinery v0.25.3 // indirect
)

replace ascend-common => ../ascend-common
