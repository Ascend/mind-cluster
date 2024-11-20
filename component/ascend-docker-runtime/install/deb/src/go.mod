module main

go 1.18

require (
	huawei.com/npu-exporter/v5 v5.0.0
	mindxcheckutils v1.0.0
)

require (
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	golang.org/x/sys v0.8.0 // indirect
)

replace (
	huawei.com/npu-exporter/v5 => gitee.com/ascend/ascend-npu-exporter/v5 v5.0.0
	mindxcheckutils => ../../../mindxcheckutils
)
