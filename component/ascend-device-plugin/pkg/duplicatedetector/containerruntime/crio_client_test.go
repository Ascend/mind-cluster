/* Copyright(C) 2026. Huawei Technologies Co.,Ltd. All rights reserved.
   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package containerruntime

import (
	"context"
	"errors"
	"net"
	"os"
	"path/filepath"
	"reflect"
	"syscall"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/util/sets"
	runtimeapi "k8s.io/cri-api/pkg/apis/runtime/v1"

	"Ascend-device-plugin/pkg/duplicatedetector/types"
	"Ascend-device-plugin/pkg/kubeclient"
)

type fakeCrioRuntime struct {
	runtimeapi.RuntimeServiceClient
	list   func(context.Context, *runtimeapi.ListContainersRequest) (*runtimeapi.ListContainersResponse, error)
	status func(context.Context, *runtimeapi.ContainerStatusRequest) (*runtimeapi.ContainerStatusResponse, error)
}

func (f *fakeCrioRuntime) ListContainers(ctx context.Context, req *runtimeapi.ListContainersRequest,
	_ ...grpc.CallOption) (*runtimeapi.ListContainersResponse, error) {
	return f.list(ctx, req)
}

func (f *fakeCrioRuntime) ContainerStatus(ctx context.Context, req *runtimeapi.ContainerStatusRequest,
	_ ...grpc.CallOption) (*runtimeapi.ContainerStatusResponse, error) {
	return f.status(ctx, req)
}

func crioStatus(id, detail string) *runtimeapi.ContainerStatusResponse {
	return &runtimeapi.ContainerStatusResponse{
		Status: &runtimeapi.ContainerStatus{
			Id: id, State: runtimeapi.ContainerState_CONTAINER_RUNNING,
			Metadata: &runtimeapi.ContainerMetadata{Name: "fallback"},
			Labels: map[string]string{"io.kubernetes.container.name": "worker",
				"io.kubernetes.pod.name": "pod", "io.kubernetes.pod.namespace": "default"},
		}, Info: map[string]string{"info": detail},
	}
}

const crioEnvSpec = `{"runtimeSpec":{"process":{"env":["ASCEND_VISIBLE_DEVICES=1,2"]}}}`

func TestCrioParseSingleContainer(t *testing.T) {
	stopped := crioStatus("one", crioEnvSpec)
	stopped.Status.State = runtimeapi.ContainerState_CONTAINER_EXITED
	for _, tc := range []struct {
		name     string
		id       string
		response *runtimeapi.ContainerStatusResponse
		err      error
		wantErr  bool
		notFound bool
	}{
		{name: "normal", id: "one", response: crioStatus("one", crioEnvSpec)},
		{name: "empty ID", wantErr: true},
		{name: "RPC failure", id: "one", err: errors.New("unavailable"), wantErr: true},
		{name: "removed", id: "one", err: status.Error(codes.NotFound, "container disappeared"),
			wantErr: true, notFound: true},
		{name: "exited", id: "one", response: stopped, wantErr: true, notFound: true},
		{name: "nil response", id: "one", wantErr: true},
		{name: "missing status", id: "one", response: &runtimeapi.ContainerStatusResponse{}, wantErr: true},
		{name: "missing ID", id: "one", response: crioStatus("", crioEnvSpec), wantErr: true},
		{name: "missing info", id: "one", response: crioStatus("one", ""), wantErr: true},
		{name: "bad JSON", id: "one", response: crioStatus("one", "{"), wantErr: true},
		{name: "missing spec", id: "one", response: crioStatus("one", "{}"), wantErr: true},
		{name: "invalid spec", id: "one", response: crioStatus("one", `{"runtimeSpec":42}`), wantErr: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fake := &fakeCrioRuntime{status: func(ctx context.Context, req *runtimeapi.ContainerStatusRequest) (
				*runtimeapi.ContainerStatusResponse, error) {
				if !req.Verbose || req.ContainerId != tc.id {
					t.Fatal("expected verbose status for requested ID")
				}
				if _, ok := ctx.Deadline(); !ok {
					t.Fatal("status RPC must have a deadline")
				}
				return tc.response, tc.err
			}}
			info, err := (&crioClient{client: fake}).ParseSingleContainer(context.Background(), tc.id)
			if (err != nil) != tc.wantErr || IsNotFound(err) != tc.notFound {
				t.Fatalf("unexpected error: %v", err)
			}
			if !tc.wantErr && (info.ID != "one" || info.Name != "worker" || info.PodName != "pod" ||
				info.PodNS != "default" || info.Namespace != crioNamespace || !reflect.DeepEqual(info.Devices, []int{1, 2})) {
				t.Fatalf("unexpected container information: %+v", info)
			}
		})
	}
}

func TestCrioSpecFallback(t *testing.T) {
	patch := gomonkey.ApplyFuncReturn(npuMajor, sets.NewString("195"))
	defer patch.Reset()
	const cgroupDevices = `{"runtimeSpec":{"linux":{"resources":{"devices":[` +
		`{"allow":true,"type":"c","major":195,"minor":6,"access":"rwm"},` +
		`{"allow":true,"type":"c","major":1,"minor":3,"access":"rwm"}]}}}}`
	for _, tc := range []struct {
		name   string
		detail string
		want   []int
	}{
		{"no process", `{"runtimeSpec":{}}`, nil},
		{"no matching env", `{"runtimeSpec":{"process":{"env":["OTHER=1"]}}}`, nil},
		{"invalid env", `{"runtimeSpec":{"process":{"env":["ASCEND_VISIBLE_DEVICES=bad"]}}}`, nil},
		{"last env wins", `{"runtimeSpec":{"process":{"env":["ASCEND_VISIBLE_DEVICES=1","ASCEND_VISIBLE_DEVICES=3"]}}}`, []int{3}},
		{"exact env key", `{"runtimeSpec":{"process":{"env":["NOT_ASCEND_VISIBLE_DEVICES=8"]}}}`, nil},
		{"cgroup devices", cgroupDevices, []int{6}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			resp := crioStatus("one", tc.detail)
			resp.Status.Labels = nil
			fake := &fakeCrioRuntime{status: func(context.Context, *runtimeapi.ContainerStatusRequest) (
				*runtimeapi.ContainerStatusResponse, error) {
				return resp, nil
			}}
			info, err := (&crioClient{client: fake}).ParseSingleContainer(context.Background(), "one")
			if err != nil || info.Name != "fallback" || !reflect.DeepEqual(info.Devices, tc.want) {
				t.Fatalf("info=%+v error=%v", info, err)
			}
		})
	}
}

func crioList(ids ...string) *runtimeapi.ListContainersResponse {
	resp := &runtimeapi.ListContainersResponse{}
	for _, id := range ids {
		resp.Containers = append(resp.Containers, &runtimeapi.Container{
			Id: id, State: runtimeapi.ContainerState_CONTAINER_RUNNING})
	}
	return resp
}

func TestCrioParseAllContainers(t *testing.T) {
	failures := map[string]error{
		"gone":    status.Error(codes.NotFound, "container disappeared"),
		"bad":     errors.New("status unavailable"),
		"timeout": status.Error(codes.DeadlineExceeded, "status request timed out"),
	}
	fake := &fakeCrioRuntime{list: func(ctx context.Context, req *runtimeapi.ListContainersRequest) (
		*runtimeapi.ListContainersResponse, error) {
		if req.GetFilter().GetState().GetState() != runtimeapi.ContainerState_CONTAINER_RUNNING {
			t.Fatal("must list running containers")
		}
		if _, ok := ctx.Deadline(); !ok {
			t.Fatal("missing deadline")
		}
		resp := crioList("one", "gone", "bad", "timeout")
		resp.Containers = append(resp.Containers, nil, &runtimeapi.Container{},
			&runtimeapi.Container{Id: "exited", State: runtimeapi.ContainerState_CONTAINER_EXITED})
		return resp, nil
	}, status: func(_ context.Context, req *runtimeapi.ContainerStatusRequest) (*runtimeapi.ContainerStatusResponse, error) {
		if err := failures[req.ContainerId]; err != nil {
			return nil, err
		}
		return crioStatus(req.ContainerId, crioEnvSpec), nil
	}}
	c := &crioClient{client: fake}
	infos, err := c.ParseAllContainers(context.Background())
	if err != nil || len(infos) != 1 || infos["one"] == nil || len(c.initialContainers) != 1 {
		t.Fatalf("containers=%v error=%v", infos, err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	calls := 0
	fake.status = func(_ context.Context, req *runtimeapi.ContainerStatusRequest) (*runtimeapi.ContainerStatusResponse, error) {
		calls++
		if calls == 2 {
			cancel()
			return nil, status.Error(codes.Canceled, "scan canceled")
		}
		return crioStatus(req.ContainerId, crioEnvSpec), nil
	}
	if partial, err := c.ParseAllContainers(ctx); !errors.Is(err, context.Canceled) || partial != nil || calls != 2 {
		t.Fatalf("canceled scan returned containers=%v, error=%v, calls=%d", partial, err, calls)
	}
	if !reflect.DeepEqual(c.initialContainers, map[string]struct{}{"one": {}}) {
		t.Fatalf("canceled scan replaced the initial snapshot: %v", c.initialContainers)
	}
}

func TestCrioEmptyContainerLists(t *testing.T) {
	for _, tc := range []struct {
		response *runtimeapi.ListContainersResponse
		err      error
		wantErr  bool
	}{{crioList(), nil, false}, {nil, nil, true}, {nil, errors.New("unavailable"), true}} {
		fake := &fakeCrioRuntime{list: func(context.Context, *runtimeapi.ListContainersRequest) (
			*runtimeapi.ListContainersResponse, error) {
			return tc.response, tc.err
		}}
		infos, err := (&crioClient{client: fake}).ParseAllContainers(context.Background())
		if (err != nil) != tc.wantErr || len(infos) != 0 {
			t.Fatalf("list response %v, error %v: got %v, %v", tc.response, tc.err, infos, err)
		}
	}
}

func TestCrioPollingRetriesAndTransitions(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	snapshots := [][]string{nil, {"old", "new"}, {"old", "new"}, {"new"}}
	lists, statuses := 0, 0
	fake := &fakeCrioRuntime{list: func(context.Context, *runtimeapi.ListContainersRequest) (
		*runtimeapi.ListContainersResponse, error) {
		ids := snapshots[lists]
		lists++
		if ids == nil {
			return nil, errors.New("temporary failure")
		}
		return crioList(ids...), nil
	}, status: func(_ context.Context, req *runtimeapi.ContainerStatusRequest) (*runtimeapi.ContainerStatusResponse, error) {
		statuses++
		if statuses == 1 {
			return nil, errors.New("not ready")
		}
		return crioStatus(req.ContainerId, crioEnvSpec), nil
	}}
	c := &crioClient{client: fake, initialContainers: map[string]struct{}{"old": {}}}
	ticks := make(chan time.Time, 3)
	for i := 0; i < 3; i++ {
		ticks <- time.Now()
	}
	var events []types.ContainerEvent
	c.pollContainerEvents(ctx, func(e types.ContainerEvent) {
		events = append(events, e)
		if e.Type == types.ContainerEventDestroy {
			cancel()
		}
	}, ticks)
	if len(events) != 2 || events[0].Type != types.ContainerEventCreate || events[0].ContainerID != "new" ||
		events[1].Type != types.ContainerEventDestroy || events[1].ContainerID != "old" || statuses != 2 || lists != 4 {
		t.Fatalf("unexpected polling result: events=%+v lists=%d statuses=%d", events, lists, statuses)
	}
	for _, e := range events {
		if e.Namespace != crioNamespace || e.Timestamp.IsZero() {
			t.Fatalf("invalid event: %+v", e)
		}
	}
	if events[0].ContainerInfo == nil || events[0].ContainerInfo.ID != "new" ||
		!reflect.DeepEqual(events[0].ContainerInfo.Devices, []int{1, 2}) || events[1].ContainerInfo != nil {
		t.Fatalf("creation must carry the parsed snapshot, destruction must not: %+v", events)
	}
}

func TestNewCrioClient(t *testing.T) {
	if _, err := NewCrioClient("relative.sock"); err == nil {
		t.Fatal("expected absolute path error")
	}
	var checked string
	patch := gomonkey.ApplyFunc(checkSockFile, func(path string) error { checked = path; return nil })
	defer patch.Reset()
	factoryPatch := gomonkey.ApplyFuncReturn(autoDetectOciEndpoint, "", errors.New("unexpected containerd lookup"))
	defer factoryPatch.Reset()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	(&crioClient{}).WatchContainerEvents(ctx, nil)
	for _, endpoint := range []string{"", "unix:///custom/crio.sock", "/custom/crio.sock"} {
		cli, err := NewClient(&types.DetectorConfig{RuntimeType: kubeclient.CrioRuntime, CriEndpoint: endpoint})
		if err != nil {
			t.Fatal(err)
		}
		c := cli.(*crioClient)
		want := "/custom/crio.sock"
		if endpoint == "" {
			want, err = normalizeCrioSocketPath(defaultCrioAddr)
			if err != nil {
				t.Fatal(err)
			}
		}
		if checked != want || c.conn.Target() != unixPre+want {
			t.Fatalf("incorrect endpoint: %s", checked)
		}
		c.WatchContainerEvents(ctx, func(types.ContainerEvent) { t.Fatal("event after cancellation") })
	}
	grpcPatch := gomonkey.ApplyFuncReturn(grpc.NewClient, nil, errors.New("client setup failed"))
	defer grpcPatch.Reset()
	if _, err := NewCrioClient(""); err == nil {
		t.Fatal("expected gRPC client setup error")
	}
}

type crioTestServer struct {
	runtimeapi.UnimplementedRuntimeServiceServer
}

func (*crioTestServer) ListContainers(context.Context, *runtimeapi.ListContainersRequest) (
	*runtimeapi.ListContainersResponse, error) {
	return crioList("rpc-container"), nil
}

func (*crioTestServer) ContainerStatus(_ context.Context, req *runtimeapi.ContainerStatusRequest) (
	*runtimeapi.ContainerStatusResponse, error) {
	if !req.Verbose {
		return nil, errors.New("verbose status required")
	}
	return crioStatus(req.ContainerId, crioEnvSpec), nil
}

func TestCrioUnixEndpoint(t *testing.T) {
	dir := t.TempDir()
	socket := filepath.Join(dir, "crio.sock")
	listener, err := net.Listen("unix", socket)
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	alias := filepath.Join(dir, "alias")
	if err := os.Symlink(dir, alias); err != nil {
		t.Fatal(err)
	}
	if _, err := NewCrioClient(filepath.Join(alias, "crio.sock")); err == nil {
		t.Fatal("arbitrary symlink escaped socket validation")
	}
	if os.Getuid() != 0 {
		if _, err := NewCrioClient(socket); err == nil {
			t.Fatal("accepted a non-root-owned socket")
		}
		patch := gomonkey.ApplyFuncReturn(checkSockFile, nil)
		defer patch.Reset()
	}
	server := grpc.NewServer()
	defer server.Stop()
	runtimeapi.RegisterRuntimeServiceServer(server, &crioTestServer{})
	go func() { _ = server.Serve(listener) }()
	client, err := NewCrioClient(socket)
	if err != nil {
		t.Fatal(err)
	}
	defer client.(*crioClient).conn.Close()
	infos, err := client.ParseAllContainers(context.Background())
	if err != nil || infos["rpc-container"] == nil ||
		!reflect.DeepEqual(infos["rpc-container"].Devices, []int{1, 2}) {
		t.Fatalf("gRPC round trip: infos=%+v error=%v", infos, err)
	}
}

type crioEndpointInfo struct {
	os.FileInfo
	mode os.FileMode
	uid  uint32
}

func (i crioEndpointInfo) Mode() os.FileMode { return i.mode }
func (i crioEndpointInfo) IsDir() bool       { return i.mode.IsDir() }
func (i crioEndpointInfo) Sys() interface{}  { return &syscall.Stat_t{Uid: i.uid} }

func TestNormalizeCrioSocketPath(t *testing.T) {
	for _, tc := range []struct {
		name, path, target, want string
		mode, parentMode         os.FileMode
		uid                      uint32
		statErr                  error
	}{
		{name: "directory", path: defaultCrioAddr, mode: os.ModeDir | 0755, want: defaultCrioAddr},
		{name: "standard alias", path: defaultCrioAddr, mode: os.ModeSymlink, target: "/run", want: "/run/crio/crio.sock"},
		{name: "relative alias", path: defaultCrioAddr, mode: os.ModeSymlink, target: "../run", want: "/run/crio/crio.sock"},
		{name: "other alias", path: defaultCrioAddr, mode: os.ModeSymlink, target: "/other"},
		{name: "unowned alias", path: defaultCrioAddr, mode: os.ModeSymlink, target: "/run", uid: 1000},
		{name: "writable parent", path: defaultCrioAddr, mode: os.ModeSymlink, target: "/run", parentMode: os.ModeDir | 0777},
		{name: "symlink parent", path: defaultCrioAddr, mode: os.ModeSymlink, target: "/run", parentMode: os.ModeSymlink},
		{name: "missing alias", path: defaultCrioAddr, statErr: os.ErrNotExist, want: defaultCrioAddr},
		{name: "unreadable alias", path: defaultCrioAddr, statErr: os.ErrPermission},
		{name: "relative path", path: "crio.sock"},
		{name: "traversal", path: "/var/run/../crio.sock"},
		{name: "custom", path: "/custom/crio.sock", want: "/custom/crio.sock"},
		{name: "similar prefix", path: "/var/runtime/crio.sock", want: "/var/runtime/crio.sock"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			patch := gomonkey.ApplyFunc(os.Lstat, func(path string) (os.FileInfo, error) {
				if path == "/var" {
					mode := tc.parentMode
					if mode == 0 {
						mode = os.ModeDir | 0755
					}
					return crioEndpointInfo{mode: mode}, nil
				}
				return crioEndpointInfo{mode: tc.mode, uid: tc.uid}, tc.statErr
			}).ApplyFunc(os.Readlink, func(string) (string, error) { return tc.target, nil })
			defer patch.Reset()
			got, err := normalizeCrioSocketPath(tc.path)
			if got != tc.want || (err != nil) != (tc.want == "") {
				t.Fatalf("path=%q error=%v, want %q", got, err, tc.want)
			}
		})
	}
}
