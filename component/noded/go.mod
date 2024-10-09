module nodeD

go 1.19

require (
	github.com/agiledragon/gomonkey/v2 v2.8.0
	github.com/smartystreets/goconvey v1.6.4
	github.com/stretchr/testify v1.8.2
	github.com/u-root/u-root v0.11.0
	huawei.com/npu-exporter/v6 v6.0.0-RC1.b001
	k8s.io/api v0.25.13
	k8s.io/apimachinery v0.25.13
	k8s.io/client-go v0.25.13
)

replace (
	huawei.com/npu-exporter/v6 => gitee.com/ascend/ascend-npu-exporter/v6 v6.0.0-RC1.b001
	k8s.io/api => k8s.io/api v0.25.13
	k8s.io/apimachinery => k8s.io/apimachinery v0.25.13
	k8s.io/client-go => k8s.io/client-go v0.25.13
)