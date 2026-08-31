module ub-host-device-cni

go 1.26

require (
	ascend-common v0.0.0
	github.com/containernetworking/cni v1.3.0
	github.com/containernetworking/plugins v1.9.1
	github.com/vishvananda/netlink v1.3.1
)

replace ascend-common => ../ascend-common

require (
	github.com/agiledragon/gomonkey/v2 v2.12.0 // indirect
	github.com/coreos/go-iptables v0.8.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/safchain/ethtool v0.6.2 // indirect
	github.com/vishvananda/netns v0.0.5 // indirect
	golang.org/x/sys v0.35.0 // indirect
	sigs.k8s.io/knftables v0.0.18 // indirect
)
