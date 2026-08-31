// Copyright 2015 CNI authors
// Modified by Huawei Technologies Co.,Ltd in 2026
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/vishvananda/netlink"

	ver "ascend-common/common-utils/version"
	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types"
	current "github.com/containernetworking/cni/pkg/types/100"
	"github.com/containernetworking/cni/pkg/version"
	"github.com/containernetworking/plugins/pkg/ip"
	"github.com/containernetworking/plugins/pkg/ipam"
	"github.com/containernetworking/plugins/pkg/netlinksafe"
	"github.com/containernetworking/plugins/pkg/ns"
	bv "github.com/containernetworking/plugins/pkg/utils/buildversion"
)

const (
	pluginName = "ub-host-device"
)

var (
	sysBusPCI       = "/sys/bus/pci/devices"
	sysBusAuxiliary = "/sys/bus/auxiliary/devices"
	sysBusUb        = "/sys/bus/ub/devices"
	// Array of different linux drivers bound to network device needed for DPDK
	userspaceDrivers = []string{"vfio-pci", "uio_pci_generic", "igb_uio"}
)

// NetConf for host-device config, look the README to learn how to use those parameters
type NetConf struct {
	types.NetConf
	Device        string `json:"device"` // Device-Name, something like eth0 or can0 etc.
	HWAddr        string `json:"hwaddr"` // MAC Address of target network interface
	DPDKMode      bool
	KernelPath    string `json:"kernelpath"` // Kernelpath of the device
	PCIAddr       string `json:"pciBusID"`   // PCI Address of target network device
	RuntimeConfig struct {
		DeviceID string `json:"deviceID,omitempty"`
	} `json:"runtimeConfig,omitempty"`

	// UBMode enables DPU (UB) mounting; devices come from the NAD "device"
	// or the kubelet-injected runtimeConfig.deviceID.
	UBMode bool `json:"ubMode,omitempty"`
	// InheritHostIP keeps the host IP addresses on the mounted UB interface
	// instead of requesting new ones from IPAM.
	InheritHostIP bool `json:"inheritHostIP,omitempty"`

	// for internal use
	auxDevice string `json:"-"` // Auxiliary device name as appears on Auxiliary bus (/sys/bus/auxiliary)
}

func init() {
	// this ensures that main runs only on main thread (thread group leader).
	// since namespace ops (unshare, setns) are done for a single thread, we
	// must ensure that the goroutine does not jump from OS thread to thread
	runtime.LockOSThread()
}

// handleDeviceID updates netconf fields with DeviceID runtime config
func handleDeviceID(netconf *NetConf) error {
	deviceID := netconf.RuntimeConfig.DeviceID
	if deviceID == "" {
		return nil
	}

	// UB mode consumes deviceID as a fallback in getUBDeviceIDs.
	if netconf.UBMode {
		return nil
	}

	// Check if deviceID is a PCI device
	pciPath := filepath.Join(sysBusPCI, deviceID)
	if _, err := os.Stat(pciPath); err == nil {
		netconf.PCIAddr = deviceID
		return nil
	}

	// Check if deviceID is an Auxiliary device
	auxPath := filepath.Join(sysBusAuxiliary, deviceID)
	if _, err := os.Stat(auxPath); err == nil {
		netconf.PCIAddr = ""
		netconf.auxDevice = deviceID
		return nil
	}

	return fmt.Errorf("runtime config DeviceID %s not found or unsupported", deviceID)
}

func loadConf(bytes []byte) (*NetConf, error) {
	n := &NetConf{}
	var err error
	if err = json.Unmarshal(bytes, n); err != nil {
		return nil, fmt.Errorf("failed to load netconf: %v", err)
	}

	// Apply the runtimeConfig DeviceID to PCIAddr/auxDevice (non-UB only).
	if err := handleDeviceID(n); err != nil {
		return nil, err
	}

	if !n.UBMode && n.Device == "" && n.HWAddr == "" && n.KernelPath == "" && n.PCIAddr == "" && n.auxDevice == "" {
		return nil, fmt.Errorf(`specify either "device", "hwaddr", "kernelpath" or "pciBusID"`)
	}

	// DPDK detection only applies to the PCI (non-UB) path.
	if !n.UBMode && len(n.PCIAddr) > 0 {
		n.DPDKMode, err = hasDpdkDriver(n.PCIAddr)
		if err != nil {
			return nil, fmt.Errorf("error with host device: %v", err)
		}
	}

	return n, nil
}

func cmdAdd(args *skel.CmdArgs) error {
	cfg, err := loadConf(args.StdinData)
	if err != nil {
		return err
	}
	if cfg.UBMode {
		return cmdAddUB(args, cfg)
	}

	containerNs, err := ns.GetNS(args.Netns)
	if err != nil {
		return fmt.Errorf("failed to open netns %q: %v", args.Netns, err)
	}
	defer containerNs.Close()

	result := &current.Result{}
	result.Interfaces = []*current.Interface{{
		Name:    args.IfName,
		Sandbox: containerNs.Path(),
	}}

	var contDev netlink.Link
	if !cfg.DPDKMode {
		contDev, err = moveHostDeviceIn(cfg, containerNs, args.IfName, result)
		if err != nil {
			return err
		}
	}

	if cfg.IPAM.Type == "" {
		return printNoIPAMResult(cfg, result, contDev, containerNs)
	}

	newResult, err := runIPAM(cfg, args, result.Interfaces)
	if err != nil {
		return err
	}

	return applyIPAMResult(cfg, containerNs, args.IfName, newResult)
}

// printNoIPAMResult prints the result when no IPAM plugin is configured:
// DPDK mode reports the pre-built interface list, otherwise the moved link.
func printNoIPAMResult(cfg *NetConf, result *current.Result, contDev netlink.Link, containerNs ns.NetNS) error {
	if cfg.DPDKMode {
		return types.PrintResult(result, cfg.CNIVersion)
	}
	return printLink(contDev, cfg.CNIVersion, containerNs)
}

// applyIPAMResult applies the IPAM result to the container interface (in the
// non-DPDK path) and prints the final result.
func applyIPAMResult(cfg *NetConf, containerNs ns.NetNS, ifName string, newResult *current.Result) error {
	if !cfg.DPDKMode {
		if err := configureIfaceInNS(containerNs, ifName, newResult); err != nil {
			return err
		}
	}
	newResult.DNS = cfg.DNS
	return types.PrintResult(newResult, cfg.CNIVersion)
}

// configureIfaceInNS applies the IPAM result (addresses/routes) to the given
// interface inside the container namespace.
func configureIfaceInNS(containerNs ns.NetNS, ifName string, res *current.Result) error {
	return containerNs.Do(func(_ ns.NetNS) error {
		return ipam.ConfigureIface(ifName, res)
	})
}

// moveHostDeviceIn moves the configured host device into the container
// namespace and records the resulting interface in the result.
func moveHostDeviceIn(cfg *NetConf, containerNs ns.NetNS, ifName string, result *current.Result) (netlink.Link, error) {
	hostDev, err := getLink(cfg.Device, cfg.HWAddr, cfg.KernelPath, cfg.PCIAddr, cfg.auxDevice)
	if err != nil {
		return nil, fmt.Errorf("failed to find host device: %v", err)
	}
	contDev, err := moveLinkIn(hostDev, containerNs, ifName)
	if err != nil {
		return nil, fmt.Errorf("failed to move link %v", err)
	}
	// Override the device name with the name in the container namespace and
	// set the MAC address of the interface.
	result.Interfaces[0].Name = contDev.Attrs().Name
	result.Interfaces[0].Mac = contDev.Attrs().HardwareAddr.String()
	return contDev, nil
}

// runIPAM allocates IPs via the IPAM plugin, releasing them on error.
func runIPAM(cfg *NetConf, args *skel.CmdArgs, interfaces []*current.Interface) (*current.Result, error) {
	r, err := ipam.ExecAdd(cfg.IPAM.Type, args.StdinData)
	if err != nil {
		return nil, err
	}
	// Invoke ipam del if err to avoid ip leak.
	defer func() {
		if err != nil {
			ipam.ExecDel(cfg.IPAM.Type, args.StdinData)
		}
	}()

	newResult, err := current.NewResultFromResult(r)
	if err != nil {
		return nil, err
	}
	if len(newResult.IPs) == 0 {
		return nil, errors.New("IPAM plugin returned missing IP config")
	}
	for _, ipc := range newResult.IPs {
		// All addresses apply to the container interface (moved from host).
		ipc.Interface = current.Int(0)
	}
	newResult.Interfaces = interfaces
	return newResult, nil
}

func cmdDel(args *skel.CmdArgs) error {
	cfg, err := loadConf(args.StdinData)
	if err != nil {
		return err
	}
	if args.Netns == "" {
		return nil
	}

	if cfg.UBMode {
		return cmdDelUB(args, cfg)
	}

	containerNs, err := ns.GetNS(args.Netns)
	if err != nil {
		return fmt.Errorf("failed to open netns %q: %v", args.Netns, err)
	}
	defer containerNs.Close()

	if cfg.IPAM.Type != "" {
		if err := ipam.ExecDel(cfg.IPAM.Type, args.StdinData); err != nil {
			return err
		}
	}

	if !cfg.DPDKMode {
		if err := moveLinkOut(containerNs, args.IfName); err != nil {
			return err
		}
	}

	return nil
}

// getUbInterfaceName resolves the host NIC name of a UB device from its
// sysfs path (e.g. "00015" -> "/sys/bus/ub/devices/00015/net").
func getUbInterfaceName(deviceID string) (string, error) {
	netDir := filepath.Join(sysBusUb, deviceID, "net")
	entries, err := os.ReadDir(netDir)
	if err != nil {
		return "", fmt.Errorf("failed to read UB net dir for %q (%s): %v", deviceID, netDir, err)
	}
	if len(entries) == 0 {
		return "", fmt.Errorf("UB device %q sysfs path %s has no net interface", deviceID, netDir)
	}
	return entries[0].Name(), nil
}

// getHostLink resolves the host netlink.Link of a UB device by its device ID.
func getHostLink(deviceID string) (netlink.Link, error) {
	ifName, err := getUbInterfaceName(deviceID)
	if err != nil {
		return nil, err
	}
	return netlinksafe.LinkByName(ifName)
}

// hostAddrs captures a host interface's IPs/routes to re-apply in the container.
type hostAddrs struct {
	addrs  []netlink.Addr
	routes []netlink.Route
}

// listHostAddrs reads the addresses/routes of the host interface before the move.
func listHostAddrs(link netlink.Link) (*hostAddrs, error) {
	addrs, err := netlink.AddrList(link, netlink.FAMILY_ALL)
	if err != nil {
		return nil, fmt.Errorf("failed to list addrs of %q: %v", link.Attrs().Name, err)
	}
	routes, err := netlink.RouteList(link, netlink.FAMILY_ALL)
	if err != nil {
		return nil, fmt.Errorf("failed to list routes of %q: %v", link.Attrs().Name, err)
	}
	return &hostAddrs{addrs: addrs, routes: routes}, nil
}

// applyHostAddrs re-applies the captured host addresses inside the container.
func applyHostAddrs(containerNs ns.NetNS, ifName string, ha *hostAddrs) error {
	return containerNs.Do(func(_ ns.NetNS) error {
		link, err := netlinksafe.LinkByName(ifName)
		if err != nil {
			return fmt.Errorf("failed to find %q in container ns: %v", ifName, err)
		}
		for i := range ha.addrs {
			// Link-local addrs are re-derived from the MAC; adding them fails (EPERM).
			if ha.addrs[i].IP.IsLinkLocalUnicast() {
				continue
			}
			if err := netlink.AddrAdd(link, &ha.addrs[i]); err != nil && !os.IsExist(err) {
				return fmt.Errorf("failed to add addr %s to %q: %v", ha.addrs[i].IPNet, ifName, err)
			}
		}
		for i := range ha.routes {
			// Point the route at the moved interface and re-add it; EEXIST means
			// the route is already present, so it can be safely skipped.
			ha.routes[i].LinkIndex = link.Attrs().Index
			if err := netlink.RouteAdd(&ha.routes[i]); err != nil && !os.IsExist(err) {
				return fmt.Errorf("failed to add route %v to %q: %v", ha.routes[i], ifName, err)
			}
		}
		return nil
	})
}

// cmdAddUB mounts the UB devices into the container netns.
func cmdAddUB(args *skel.CmdArgs, cfg *NetConf) error {
	deviceIDs, err := getUBDeviceIDs(args, cfg)
	if err != nil {
		return err
	}

	containerNs, err := ns.GetNS(args.Netns)
	if err != nil {
		return fmt.Errorf("failed to open netns %q: %v", args.Netns, err)
	}
	defer containerNs.Close()

	result := &current.Result{}
	for _, deviceID := range deviceIDs {
		if err := setupUBDevice(cfg, args, containerNs, deviceID, result); err != nil {
			return err
		}
	}

	result.DNS = cfg.DNS
	return types.PrintResult(result, cfg.CNIVersion)
}

// getUBDeviceIDs returns the UB devices to mount: NAD "device" first,
// else runtimeConfig.deviceID.
func getUBDeviceIDs(_ *skel.CmdArgs, cfg *NetConf) ([]string, error) {
	// An explicitly configured NAD "device" is authoritative.
	if id, ok, err := ubDeviceIDFromNAD(cfg); err != nil {
		return nil, err
	} else if ok {
		return []string{id}, nil
	}

	// Fall back to runtimeConfig.deviceID (a UB address used as-is).
	if cfg.RuntimeConfig.DeviceID != "" {
		return []string{cfg.RuntimeConfig.DeviceID}, nil
	}

	return nil, fmt.Errorf("no allocated DPU devices found for the pod")
}

// ubDeviceIDFromNAD resolves the NAD "device" (host NIC name) to its UB
// device address; ok=false when no device is configured (pciBusID is ignored).
func ubDeviceIDFromNAD(cfg *NetConf) (string, bool, error) {
	if cfg.Device != "" {
		addr, err := ubAddressForInterfaceName(cfg.Device)
		if err != nil {
			return "", false, fmt.Errorf("cannot resolve UB device from config device %q: %v", cfg.Device, err)
		}
		return addr, true, nil
	}
	return "", false, nil
}

// ubAddressForInterfaceName finds the UB device owning a host NIC by
// scanning /sys/bus/ub/devices.
func ubAddressForInterfaceName(ifName string) (string, error) {
	devices, err := os.ReadDir(sysBusUb)
	if err != nil {
		return "", fmt.Errorf("failed to read %s: %v", sysBusUb, err)
	}
	for _, dev := range devices {
		netDir := filepath.Join(sysBusUb, dev.Name(), "net")
		entries, err := os.ReadDir(netDir)
		if err != nil {
			continue
		}
		if len(entries) > 0 && entries[0].Name() == ifName {
			return dev.Name(), nil
		}
	}
	return "", fmt.Errorf("no UB device owns interface %q", ifName)
}

// setupUBDevice mounts one UB device, allocating IPs before the move.
func setupUBDevice(cfg *NetConf, args *skel.CmdArgs, containerNs ns.NetNS, deviceID string, result *current.Result) (err error) {
	// UB mode must either allocate IPs via IPAM or inherit the host IPs,
	// otherwise applyUBIPAM would receive a nil result and panic.
	if cfg.IPAM.Type == "" && !cfg.InheritHostIP {
		return fmt.Errorf("ubMode requires either ipam or inheritHostIP")
	}
	hostDev, err := getHostLink(deviceID)
	if err != nil {
		return fmt.Errorf("failed to find host device for UB %q: %v", deviceID, err)
	}

	var ipamResult *current.Result
	if cfg.IPAM.Type != "" && !cfg.InheritHostIP {
		ipamResult, err = allocateUBIPAM(cfg, args)
		if err != nil {
			return err
		}
		// Release the IPAM allocation if a later step fails.
		defer func() {
			if err != nil {
				_ = ipam.ExecDel(cfg.IPAM.Type, args.StdinData)
			}
		}()
	}

	var ha *hostAddrs
	if cfg.InheritHostIP {
		ha, err = listHostAddrs(hostDev)
		if err != nil {
			return err
		}
	}

	// Keep the host interface name inside the container namespace.
	contDev, err := moveLinkIn(hostDev, containerNs, hostDev.Attrs().Name)
	if err != nil {
		return fmt.Errorf("failed to move link %v", err)
	}

	iface := &current.Interface{
		Name:    contDev.Attrs().Name,
		Mac:     contDev.Attrs().HardwareAddr.String(),
		Sandbox: containerNs.Path(),
	}
	result.Interfaces = append(result.Interfaces, iface)

	if cfg.InheritHostIP {
		return applyInheritedAddrs(containerNs, iface, ha, result)
	}
	return applyUBIPAM(containerNs, iface, ipamResult, result)
}

// allocateUBIPAM requests IPs from the configured IPAM plugin for the device.
func allocateUBIPAM(cfg *NetConf, args *skel.CmdArgs) (*current.Result, error) {
	r, err := ipam.ExecAdd(cfg.IPAM.Type, args.StdinData)
	if err != nil {
		return nil, fmt.Errorf("failed to allocate IPs via %q IPAM: %v", cfg.IPAM.Type, err)
	}
	return current.NewResultFromResult(r)
}

// applyUBIPAM applies the IPAM result to the mounted interface, filling in
// the interface when the IPAM omits it.
func applyUBIPAM(containerNs ns.NetNS, iface *current.Interface, ipamResult *current.Result, result *current.Result) error {
	if len(ipamResult.Interfaces) == 0 {
		ipamResult.Interfaces = []*current.Interface{
			{Name: iface.Name, Mac: iface.Mac, Sandbox: iface.Sandbox},
		}
		for _, ipc := range ipamResult.IPs {
			if ipc.Interface == nil {
				idx := 0
				ipc.Interface = &idx
			}
		}
	}
	if err := configureIfaceInNS(containerNs, iface.Name, ipamResult); err != nil {
		return err
	}
	result.IPs = append(result.IPs, ipamResult.IPs...)
	result.Routes = append(result.Routes, ipamResult.Routes...)
	return nil
}

// applyInheritedAddrs re-applies the captured host addresses and reports them.
func applyInheritedAddrs(containerNs ns.NetNS, iface *current.Interface, ha *hostAddrs, result *current.Result) error {
	if err := applyHostAddrs(containerNs, iface.Name, ha); err != nil {
		return err
	}
	for _, a := range ha.addrs {
		result.IPs = append(result.IPs, &current.IPConfig{
			Interface: current.Int(len(result.Interfaces) - 1),
			Address:   *a.IPNet,
		})
	}
	return nil
}

// cmdDelUB moves the UB interfaces (host names, not args.IfName) back to host.
func cmdDelUB(args *skel.CmdArgs, cfg *NetConf) error {
	containerNs, err := ns.GetNS(args.Netns)
	if err != nil {
		return fmt.Errorf("failed to open netns %q: %v", args.Netns, err)
	}
	defer containerNs.Close()

	if cfg.IPAM.Type != "" && !cfg.InheritHostIP {
		if err := ipam.ExecDel(cfg.IPAM.Type, args.StdinData); err != nil {
			return err
		}
	}

	// Interfaces keep their host names, so move out every UB-owned one.
	ifNames, err := ubInterfaceNamesInNetns(containerNs)
	if err != nil {
		return err
	}
	for _, ifName := range ifNames {
		if err := moveLinkOut(containerNs, ifName); err != nil {
			return err
		}
	}
	return nil
}

// ubInterfaceNamesInNetns lists the UB-owned interfaces in a namespace.
func ubInterfaceNamesInNetns(containerNs ns.NetNS) ([]string, error) {
	var names []string
	err := containerNs.Do(func(_ ns.NetNS) error {
		links, err := netlink.LinkList()
		if err != nil {
			return err
		}
		for _, l := range links {
			if isUbNetdev(l.Attrs().Name) {
				names = append(names, l.Attrs().Name)
			}
		}
		return nil
	})
	return names, err
}

// isUbNetdev reports whether a netdev's sysfs device link points into
// /sys/bus/ub/devices (raw target; EvalSymlinks would break the match).
func isUbNetdev(ifName string) bool {
	target, err := os.Readlink(filepath.Join("/sys/class/net", ifName, "device"))
	if err != nil {
		return false
	}
	return strings.Contains(target, "/bus/ub/devices/")
}

func moveLinkIn(hostDev netlink.Link, containerNs ns.NetNS, containerIfName string) (netlink.Link, error) {
	hostDevName := hostDev.Attrs().Name

	// Rename via a tempNS: direct rapid renames race with udev/NM.
	tempNS, err := ns.TempNetNS()
	if err != nil {
		return nil, fmt.Errorf("failed to create tempNS: %v", err)
	}
	defer tempNS.Close()

	// Restore the up state on error (moving ns sets the link down).
	if hostDev.Attrs().Flags&net.FlagUp == net.FlagUp {
		defer func() {
			if err != nil {
				if hostDev, err := netlinksafe.LinkByName(hostDevName); err == nil {
					_ = netlink.LinkSetUp(hostDev)
				}
			}
		}()
	}

	// Move the host device into tempNS
	if err = netlink.LinkSetNsFd(hostDev, int(tempNS.Fd())); err != nil {
		return nil, fmt.Errorf("failed to move %q to tempNS: %v", hostDevName, err)
	}

	return renameAndMoveToContainer(tempNS, containerNs, hostDevName, containerIfName)
}

// renameAndMoveToContainer renames the device in tempNS, sets the alias, and
// moves it into the container ns (rolling back on error).
func renameAndMoveToContainer(tempNS, containerNs ns.NetNS, hostDevName, containerIfName string) (netlink.Link, error) {
	var contDev netlink.Link
	err := tempNS.Do(func(hostNS ns.NetNS) error {
		var err error
		contDev, err = renameAndMoveInTempNS(hostNS, tempNS, containerNs, hostDevName, containerIfName)
		return err
	})
	if err != nil {
		return nil, err
	}
	return contDev, nil
}

// renameAndMoveInTempNS renames/aliases/moves the device inside tempNS,
// rolling back on error.
func renameAndMoveInTempNS(hostNS, tempNS, containerNs ns.NetNS, hostDevName, containerIfName string) (netlink.Link, error) {
	var contDev netlink.Link
	var err error

	// lookup the device in tempNS (index might have changed)
	tempNSDev, err := netlinksafe.LinkByName(hostDevName)
	if err != nil {
		return nil, fmt.Errorf("failed to find %q in tempNS: %v", hostDevName, err)
	}
	// destroying a non empty tempNS would move physical devices back to
	// the initial net namespace, not the parent one, so move it back on error
	defer func() {
		if err != nil && tempNSDev != nil {
			_ = netlink.LinkSetNsFd(tempNSDev, int(hostNS.Fd()))
		}
	}()
	// Rename the device to the wanted name
	if err = netlink.LinkSetName(tempNSDev, containerIfName); err != nil {
		return nil, fmt.Errorf("failed to rename host device %q to %q: %v", hostDevName, containerIfName, err)
	}
	defer func() {
		if err != nil && tempNSDev != nil {
			_ = netlink.LinkSetName(tempNSDev, hostDevName)
		}
	}()
	// Save host device name into the container device's alias property
	if err = netlink.LinkSetAlias(tempNSDev, hostDevName); err != nil {
		return nil, fmt.Errorf("failed to set alias to %q: %v", hostDevName, err)
	}
	defer func() {
		if err != nil && tempNSDev != nil {
			_ = netlink.LinkSetAlias(tempNSDev, "")
		}
	}()
	// Move the device to the containerNS
	if err = netlink.LinkSetNsFd(tempNSDev, int(containerNs.Fd())); err != nil {
		return nil, fmt.Errorf("failed to move %q (host: %q) to container NS: %v", containerIfName, hostDevName, err)
	}
	defer func() {
		if err != nil {
			tempNSDev, _ = netlinksafe.LinkByName(containerIfName)
		}
	}()
	contDev, err = bringLinkUp(containerNs, tempNS, containerIfName)
	return contDev, err
}

// bringLinkUp looks up the device in the container namespace and brings it
// up, moving it back to tempNS on error.
func bringLinkUp(containerNs, tempNS ns.NetNS, containerIfName string) (netlink.Link, error) {
	var contDev netlink.Link
	err := containerNs.Do(func(_ ns.NetNS) error {
		var err error
		contDev, err = netlinksafe.LinkByName(containerIfName)
		if err != nil {
			return fmt.Errorf("failed to find %q in container NS: %v", containerIfName, err)
		}
		// Move the interface back to tempNS on error
		defer func() {
			if err != nil {
				_ = netlink.LinkSetNsFd(contDev, int(tempNS.Fd()))
			}
		}()
		// Bring the device up; this must be done in the containerNS.
		if err = netlink.LinkSetUp(contDev); err != nil {
			return fmt.Errorf("failed to set %q up: %v", containerIfName, err)
		}
		return nil
	})
	return contDev, err
}

func moveLinkOut(containerNs ns.NetNS, containerIfName string) error {
	// Rename via a tempNS: rapid renames race with udev/NM.
	tempNS, err := ns.TempNetNS()
	if err != nil {
		return fmt.Errorf("failed to create tempNS: %v", err)
	}
	defer tempNS.Close()

	var contDev netlink.Link

	// Restore the up state on error (moving ns sets the link down).
	defer func() {
		if err != nil && contDev != nil && contDev.Attrs().Flags&net.FlagUp == net.FlagUp {
			containerNs.Do(func(_ ns.NetNS) error {
				if contDev, err := netlinksafe.LinkByName(containerIfName); err == nil {
					_ = netlink.LinkSetUp(contDev)
				}
				return nil
			})
		}
	}()

	contDev, err = moveLinkOutToTempNS(containerNs, tempNS, containerIfName)
	if err != nil {
		return err
	}

	return moveLinkOutToHost(tempNS, containerNs, containerIfName)
}

// moveLinkOutToTempNS looks up the container interface and moves it into
// tempNS, verifying the original host name is recorded in the alias.
func moveLinkOutToTempNS(containerNs, tempNS ns.NetNS, containerIfName string) (netlink.Link, error) {
	var contDev netlink.Link
	err := containerNs.Do(func(_ ns.NetNS) error {
		var err error
		contDev, err = netlinksafe.LinkByName(containerIfName)
		if err != nil {
			return fmt.Errorf("failed to find %q in containerNS: %v", containerIfName, err)
		}
		// Verify we have the original name
		if contDev.Attrs().Alias == "" {
			return fmt.Errorf("failed to find original ifname for %q (alias is not set)", containerIfName)
		}
		if err = netlink.LinkSetNsFd(contDev, int(tempNS.Fd())); err != nil {
			return fmt.Errorf("failed to move %q to tempNS: %v", containerIfName, err)
		}
		return nil
	})
	return contDev, err
}

// moveLinkOutToHost restores the host name from the alias and moves the
// device back to the host ns (rolling back on error).
func moveLinkOutToHost(tempNS, containerNs ns.NetNS, containerIfName string) error {
	return tempNS.Do(func(hostNS ns.NetNS) error {
		// Lookup the device in tempNS (index might have changed)
		tempNSDev, err := netlinksafe.LinkByName(containerIfName)
		if err != nil {
			return fmt.Errorf("failed to find %q in tempNS: %v", containerIfName, err)
		}
		// Move the device back to containerNS on error
		defer func() {
			if err != nil {
				_ = netlink.LinkSetNsFd(tempNSDev, int(containerNs.Fd()))
			}
		}()
		hostDevName := tempNSDev.Attrs().Alias
		// Rename container device to hostDevName
		if err = netlink.LinkSetName(tempNSDev, hostDevName); err != nil {
			return fmt.Errorf("failed to rename device %q to %q: %v", containerIfName, hostDevName, err)
		}
		defer func() {
			if err != nil {
				_ = netlink.LinkSetName(tempNSDev, containerIfName)
			}
		}()
		// Unset device's alias property
		if err = netlink.LinkSetAlias(tempNSDev, ""); err != nil {
			return fmt.Errorf("failed to unset alias of %q: %v", hostDevName, err)
		}
		defer func() {
			if err != nil {
				_ = netlink.LinkSetAlias(tempNSDev, hostDevName)
			}
		}()
		// Finally move the device to the hostNS
		if err = netlink.LinkSetNsFd(tempNSDev, int(hostNS.Fd())); err != nil {
			return fmt.Errorf("failed to move %q to hostNS: %v", hostDevName, err)
		}
		// As we don't know the previous state, leave the link down
		return nil
	})
}

func hasDpdkDriver(pciaddr string) (bool, error) {
	driverLink := filepath.Join(sysBusPCI, pciaddr, "driver")
	driverPath, err := filepath.EvalSymlinks(driverLink)
	if err != nil {
		return false, err
	}
	driverStat, err := os.Stat(driverPath)
	if err != nil {
		return false, err
	}
	driverName := driverStat.Name()
	for _, drv := range userspaceDrivers {
		if driverName == drv {
			return true, nil
		}
	}
	return false, nil
}

func printLink(dev netlink.Link, cniVersion string, containerNs ns.NetNS) error {
	result := current.Result{
		CNIVersion: current.ImplementedSpecVersion,
		Interfaces: []*current.Interface{
			{
				Name:    dev.Attrs().Name,
				Mac:     dev.Attrs().HardwareAddr.String(),
				Sandbox: containerNs.Path(),
			},
		},
	}
	return types.PrintResult(&result, cniVersion)
}

func linkFromPath(path string) (netlink.Link, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %q", path, err)
	}
	if len(entries) > 0 {
		// grab the first net device
		return netlinksafe.LinkByName(entries[0].Name())
	}
	return nil, fmt.Errorf("failed to find network device in path %s", path)
}

func getLink(devname, hwaddr, kernelpath, pciaddr string, auxDev string) (netlink.Link, error) {
	switch {

	case len(devname) > 0:
		return netlinksafe.LinkByName(devname)
	case len(hwaddr) > 0:
		hwAddr, err := net.ParseMAC(hwaddr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse MAC address %q: %v", hwaddr, err)
		}

		links, err := netlinksafe.LinkList()
		if err != nil {
			return nil, fmt.Errorf("failed to list node links: %v", err)
		}

		for _, link := range links {
			if bytes.Equal(link.Attrs().HardwareAddr, hwAddr) {
				return link, nil
			}
		}
	case len(kernelpath) > 0:
		if !filepath.IsAbs(kernelpath) || !strings.HasPrefix(kernelpath, "/sys/devices/") {
			return nil, fmt.Errorf("kernel device path %q must be absolute and begin with /sys/devices/", kernelpath)
		}
		netDir := filepath.Join(kernelpath, "net")
		return linkFromPath(netDir)
	case len(pciaddr) > 0:
		netDir := filepath.Join(sysBusPCI, pciaddr, "net")
		if _, err := os.Lstat(netDir); err != nil {
			virtioNetDir := filepath.Join(sysBusPCI, pciaddr, "virtio*", "net")
			matches, err := filepath.Glob(virtioNetDir)
			if matches == nil || err != nil {
				return nil, fmt.Errorf("no net directory under pci device %s", pciaddr)
			}
			netDir = matches[0]
		}
		return linkFromPath(netDir)
	case len(auxDev) > 0:
		netDir := filepath.Join(sysBusAuxiliary, auxDev, "net")
		return linkFromPath(netDir)
	}

	return nil, fmt.Errorf("failed to find physical interface")
}

func main() {
	// Show version information (single variable for both -version and -v).
	var showVersion bool
	flag.BoolVar(&showVersion, "version", false, "Print the version and exit")
	flag.BoolVar(&showVersion, "v", false, "Print the version and exit")
	flag.Parse()
	if showVersion {
		info := ver.Get()
		fmt.Printf("version=%s commit=%s branch=%s os=%s arch=%s goVersion=%s\n",
			info.Version, info.GitCommit, info.GitBranch, info.BuildOS, info.BuildArch, info.GoVersion)
		return
	}

	skel.PluginMainFuncs(skel.CNIFuncs{
		Add:    cmdAdd,
		Check:  cmdCheck,
		Del:    cmdDel,
		Status: cmdStatus,
		/* FIXME GC */
	}, version.All, bv.BuildString(pluginName))
}

func cmdCheck(args *skel.CmdArgs) error {
	cfg, err := loadConf(args.StdinData)
	if err != nil {
		return err
	}
	netns, err := ns.GetNS(args.Netns)
	if err != nil {
		return fmt.Errorf("failed to open netns %q: %v", args.Netns, err)
	}
	defer netns.Close()

	// run the IPAM plugin and get back the config to apply
	if cfg.IPAM.Type != "" {
		if err := ipam.ExecCheck(cfg.IPAM.Type, args.StdinData); err != nil {
			return err
		}
	}

	result, err := parsePrevResult(cfg)
	if err != nil {
		return err
	}

	if cfg.DPDKMode {
		return nil
	}

	contMap, err := findContainerInterface(result, args)
	if err != nil {
		return err
	}

	return netns.Do(func(_ ns.NetNS) error {
		return validateContainerState(contMap, args, result)
	})
}

// parsePrevResult parses the RawPrevResult of the previous CNI invocation
// into a usable current.Result.
func parsePrevResult(cfg *NetConf) (*current.Result, error) {
	if cfg.NetConf.RawPrevResult == nil {
		return nil, fmt.Errorf("Required prevResult missing")
	}
	if err := version.ParsePrevResult(&cfg.NetConf); err != nil {
		return nil, err
	}
	return current.NewResultFromResult(cfg.PrevResult)
}

// findContainerInterface returns the previous-result interface whose name and
// sandbox match the current invocation.
func findContainerInterface(result *current.Result, args *skel.CmdArgs) (*current.Interface, error) {
	var contMap *current.Interface
	for _, intf := range result.Interfaces {
		if args.IfName == intf.Name && args.Netns == intf.Sandbox {
			contMap = intf
			break
		}
	}
	// The namespace must be the same as what was configured
	if contMap == nil {
		return nil, fmt.Errorf("Sandbox in prevResult doesn't match configured netns: %s", args.Netns)
	}
	return contMap, nil
}

// validateContainerState checks the container interface, IPs and routes in
// the previous result against the values found in the container.
func validateContainerState(contMap *current.Interface, args *skel.CmdArgs, result *current.Result) error {
	if err := validateCniContainerInterface(*contMap); err != nil {
		return err
	}
	if err := ip.ValidateExpectedInterfaceIPs(args.IfName, result.IPs); err != nil {
		return err
	}
	return ip.ValidateExpectedRoute(result.Routes)
}

func validateCniContainerInterface(intf current.Interface) error {
	var link netlink.Link
	var err error

	if intf.Name == "" {
		return fmt.Errorf("Container interface name missing in prevResult: %v", intf.Name)
	}
	link, err = netlinksafe.LinkByName(intf.Name)
	if err != nil {
		return fmt.Errorf("Container Interface name in prevResult: %s not found", intf.Name)
	}
	if intf.Sandbox == "" {
		return fmt.Errorf("Error: Container interface %s should not be in host namespace", link.Attrs().Name)
	}

	if intf.Mac != "" {
		if intf.Mac != link.Attrs().HardwareAddr.String() {
			return fmt.Errorf("Interface %s Mac %s doesn't match container Mac: %s", intf.Name, intf.Mac, link.Attrs().HardwareAddr)
		}
	}

	return nil
}

func cmdStatus(args *skel.CmdArgs) error {
	conf := NetConf{}
	if err := json.Unmarshal(args.StdinData, &conf); err != nil {
		return fmt.Errorf("failed to load netconf: %w", err)
	}

	if conf.IPAM.Type != "" {
		if err := ipam.ExecStatus(conf.IPAM.Type, args.StdinData); err != nil {
			return err
		}
	}

	// TODO: Check if host device exists.

	return nil
}
