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

package container

import (
	"context"
	"errors"
	"net"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	criv1 "k8s.io/cri-api/pkg/apis/runtime/v1"
	"k8s.io/cri-api/pkg/apis/runtime/v1alpha2"

	"ascend-common/api"
	commonutils "ascend-common/common-utils/utils"
)

const crioTestInfo = `{"pid":1234,"runtimeSpec":{"process":{"env":["ASCEND_VISIBLE_DEVICES=0"]},"linux":{"resources":{"devices":[{"allow":true,"type":"c","major":236,"minor":0,"access":"rwm"}]}}}}`

type crioV1Server struct {
	criv1.UnimplementedRuntimeServiceServer
	info  string
	err   error
	calls []*criv1.ContainerStatusRequest
}

func (s *crioV1Server) ContainerStatus(ctx context.Context, req *criv1.ContainerStatusRequest) (*criv1.ContainerStatusResponse, error) {
	s.calls = append(s.calls, req)
	if s.err != nil {
		return nil, s.err
	}
	return &criv1.ContainerStatusResponse{Info: map[string]string{"info": s.info}}, nil
}

func (s *crioV1Server) ListContainers(context.Context, *criv1.ListContainersRequest) (*criv1.ListContainersResponse, error) {
	return &criv1.ListContainersResponse{Containers: []*criv1.Container{{Id: "crio-id"}}}, nil
}

type crioAlphaServer struct {
	v1alpha2.UnimplementedRuntimeServiceServer
	info  string
	err   error
	calls []*v1alpha2.ContainerStatusRequest
}

func (s *crioAlphaServer) ContainerStatus(ctx context.Context, req *v1alpha2.ContainerStatusRequest) (*v1alpha2.ContainerStatusResponse, error) {
	s.calls = append(s.calls, req)
	if s.err != nil {
		return nil, s.err
	}
	return &v1alpha2.ContainerStatusResponse{Info: map[string]string{"info": s.info}}, nil
}

func (s *crioAlphaServer) ListContainers(context.Context, *v1alpha2.ListContainersRequest) (*v1alpha2.ListContainersResponse, error) {
	return &v1alpha2.ListContainersResponse{Containers: []*v1alpha2.Container{{Id: "crio-id"}}}, nil
}

func newCrioTestOperator(t *testing.T, v *crioV1Server, a *crioAlphaServer) *RuntimeOperatorTool {
	t.Helper()
	listener := bufconn.Listen(1024 * 1024)
	server := grpc.NewServer()
	if v != nil {
		criv1.RegisterRuntimeServiceServer(server, v)
	}
	if a != nil {
		v1alpha2.RegisterRuntimeServiceServer(server, a)
	}
	go func() { _ = server.Serve(listener) }()
	t.Cleanup(func() { server.Stop(); _ = listener.Close() })
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	conn, err := grpc.DialContext(ctx, "passthrough:///crio-test", grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) { return listener.Dial() }), grpc.WithBlock())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	return &RuntimeOperatorTool{EndpointType: EndpointTypeCrio, criConn: conn,
		criClient: v1alpha2.NewRuntimeServiceClient(conn)}
}

func TestCrioContainerStatus(t *testing.T) {
	tests := []struct {
		name                string
		v1                  bool
		version             string
		v1Err               error
		alphaErr            error
		wantErr             codes.Code
		v1Calls, alphaCalls int
	}{
		{name: "v1", v1: true, v1Calls: 1},
		{name: "v1alpha2 only", alphaCalls: 1},
		{name: "cached v1alpha2", v1: true, version: criVersionV1alpha2, alphaCalls: 1},
		{name: "unimplemented fallback", v1: true, v1Err: status.Error(codes.Unimplemented, "unsupported method"), v1Calls: 1, alphaCalls: 1},
		{name: "not found does not fallback", v1: true, v1Err: status.Error(codes.NotFound, "gone"), wantErr: codes.NotFound, v1Calls: 1},
		{name: "permission denied does not fallback", v1: true, v1Err: status.Error(codes.PermissionDenied, "denied"), wantErr: codes.PermissionDenied, v1Calls: 1},
		{name: "unavailable does not fallback", v1: true, v1Err: status.Error(codes.Unavailable, "offline"), wantErr: codes.Unavailable, v1Calls: 1},
		{name: "alpha error", alphaErr: status.Error(codes.NotFound, "gone"), wantErr: codes.NotFound, alphaCalls: 1},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			v := &crioV1Server{info: crioTestInfo, err: tc.v1Err}
			a := &crioAlphaServer{info: crioTestInfo, err: tc.alphaErr}
			var registered *crioV1Server
			if tc.v1 {
				registered = v
			}
			op := newCrioTestOperator(t, registered, a)
			op.criVersion = tc.version
			spec, err := op.GetContainerInfoByID(context.Background(), "crio-id")
			if status.Code(err) != tc.wantErr {
				t.Fatalf("error = %v, want %v", err, tc.wantErr)
			}
			if tc.wantErr == codes.OK {
				if spec.Process == nil || len(spec.Process.Env) != 1 || spec.Linux == nil || spec.Linux.Resources == nil {
					t.Fatalf("incomplete spec: %+v", spec)
				}
				device := spec.Linux.Resources.Devices[0]
				if *device.Major != 236 || *device.Minor != 0 {
					t.Fatalf("wrong device: %+v", device)
				}
			}
			if len(v.calls) != tc.v1Calls || len(a.calls) != tc.alphaCalls {
				t.Fatalf("calls: v1=%d alpha=%d", len(v.calls), len(a.calls))
			}
			for _, req := range v.calls {
				if req.ContainerId != "crio-id" || !req.Verbose {
					t.Fatalf("invalid request: %+v", req)
				}
			}
			for _, req := range a.calls {
				if req.ContainerId != "crio-id" || !req.Verbose {
					t.Fatalf("invalid request: %+v", req)
				}
			}
		})
	}
}

func TestCrioListAndStatus(t *testing.T) {
	for _, alphaOnly := range []bool{false, true} {
		t.Run(map[bool]string{false: "v1", true: "v1alpha2"}[alphaOnly], func(t *testing.T) {
			var v *crioV1Server
			if !alphaOnly {
				v = &crioV1Server{info: crioTestInfo}
			}
			op := newCrioTestOperator(t, v, &crioAlphaServer{info: crioTestInfo})
			containers, err := op.GetContainers(context.Background())
			if err != nil || len(containers) != 1 || containers[0].Id != "crio-id" {
				t.Fatalf("containers=%v err=%v", containers, err)
			}
			version := criVersionV1
			if alphaOnly {
				version = criVersionV1alpha2
			}
			if op.criVersion != version {
				t.Fatalf("version=%s", op.criVersion)
			}
			if _, err := op.GetContainerInfoByID(context.Background(), containers[0].Id); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestCrioParseInfo(t *testing.T) {
	tests := []struct {
		name, raw string
		valid     bool
	}{
		{"valid", crioTestInfo, true},
		{"no devices", `{"runtimeSpec":{"process":{},"linux":{"resources":{}}}}`, true},
		{"invalid pid does not block devices", strings.Replace(crioTestInfo, "1234", `"unknown"`, 1), true},
		{"empty", "", false}, {"invalid JSON", "{", false}, {"null", "null", false},
		{"no spec", `{"pid":1234}`, false}, {"null spec", `{"runtimeSpec":null}`, false},
		{"string spec", `{"runtimeSpec":"bad"}`, false},
		{"missing process", `{"runtimeSpec":{"linux":{"resources":{}}}}`, false},
		{"missing linux", `{"runtimeSpec":{"process":{}}}`, false},
		{"missing resources", `{"runtimeSpec":{"process":{},"linux":{}}}`, false},
		{"invalid env", `{"runtimeSpec":{"process":{"env":1},"linux":{"resources":{}}}}`, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := parseCrioContainerInfo(map[string]string{"info": tc.raw})
			if (err == nil) != tc.valid {
				t.Fatalf("err=%v", err)
			}
		})
	}
	if _, err := parseCrioContainerInfo(nil); err == nil {
		t.Fatal("expected missing info error")
	}
	op := &RuntimeOperatorTool{EndpointType: EndpointTypeCrio}
	if _, err := op.GetContainerInfoByID(context.Background(), "id"); err == nil {
		t.Fatal("expected connection error")
	}
}

func TestCrioPIDs(t *testing.T) {
	for _, pid := range []string{"1234", "0", "-1", "2147483648", "4294967296", "null", `"bad"`} {
		t.Run(pid, func(t *testing.T) {
			op := newCrioTestOperator(t, &crioV1Server{info: strings.Replace(crioTestInfo, "1234", pid, 1)}, nil)
			pids, err := op.GetContainerPIDs(context.Background(), "crio-id")
			if pid == "1234" {
				if err != nil || !reflect.DeepEqual(pids, []uint32{1234}) {
					t.Fatalf("pids=%v err=%v", pids, err)
				}
			} else if err == nil {
				t.Fatal("expected PID error")
			}
		})
	}
	op := newCrioTestOperator(t, &crioV1Server{err: status.Error(codes.NotFound, "gone")}, nil)
	if _, err := op.GetContainerPIDs(context.Background(), "id"); err == nil {
		t.Fatal("expected status error")
	}
}

func TestCrioParserReuse(t *testing.T) {
	raw := strings.Replace(crioTestInfo, "ASCEND_VISIBLE_DEVICES", api.AscendDeviceInfo, 1)
	op := newCrioTestOperator(t, &crioV1Server{info: raw}, nil)
	parser := &DevicesParser{RuntimeOperator: op}
	results := make(chan DevicesInfo, 1)
	c := &CommonContainer{Id: "crio-id", Labels: map[string]string{
		labelK8sPodNamespace: "default", labelK8sPodName: "npu-pod", labelContainerName: "worker"}}
	if err := parser.parseDevices(context.Background(), c, results); err != nil {
		t.Fatal(err)
	}
	result := <-results
	if result.ID != c.Id || result.Name != "default_npu-pod_worker" || !reflect.DeepEqual(result.Devices, []int{0}) {
		t.Fatalf("result=%+v", result)
	}
	if _, ok := result.PIDs[1234]; !ok {
		t.Fatalf("PIDs=%v", result.PIDs)
	}
}

func TestCrioInitAndSocket(t *testing.T) {
	parser := MakeDevicesParser(CntNpuMonitorOpts{EndpointType: EndpointTypeCrio, CriEndpoint: "unix:///custom/crio.sock",
		UseCriBackup: false, UseOciBackup: false})
	op := parser.RuntimeOperator.(*RuntimeOperatorTool)
	if op.GetContainerType() != CrioContainer || op.UseCriBackup || op.UseOciBackup {
		t.Fatalf("operator=%+v", op)
	}
	patches := gomonkey.NewPatches()
	defer patches.Reset()
	paths := []string{}
	patches.ApplyFunc(commonutils.CheckPath, func(path string) (string, error) { paths = append(paths, path); return path, nil })
	patches.ApplyFuncReturn(commonutils.DoCheckOwnerAndPermission, nil)
	if err := sockCheck(op); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(paths, []string{"/custom/crio.sock"}) {
		t.Fatalf("paths=%v", paths)
	}
	patches.ApplyFuncReturn(syscall.Getuid, 0)
	patches.ApplyFuncReturn((*RuntimeOperatorTool).initCriClient, nil)
	patches.ApplyFunc((*RuntimeOperatorTool).initOciClient, func(*RuntimeOperatorTool) error {
		t.Error("CRI-O must not initialize a containerd client")
		return errors.New("unexpected OCI client")
	})
	if err := op.Init(); err != nil {
		t.Fatal(err)
	}
	if err := op.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestCrioCancellationAndClose(t *testing.T) {
	op := newCrioTestOperator(t, &crioV1Server{info: crioTestInfo}, nil)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := op.GetContainerInfoByID(ctx, "id"); status.Code(err) != codes.Canceled {
		t.Fatalf("err=%v", err)
	}
	if err := op.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestCrioParserWithoutAscendEnv(t *testing.T) {
	// The shared fixture contains ASCEND_VISIBLE_DEVICES. Remove it so this
	// test cannot silently pass via the environment-variable fast path.
	raw := strings.Replace(crioTestInfo, "ASCEND_VISIBLE_DEVICES=0", "PATH=/usr/bin", 1)
	raw = strings.Replace(raw, `"minor":0`, `"minor":7`, 1)
	if strings.Contains(raw, api.AscendDeviceInfo) {
		t.Fatal("fixture must not contain the Ascend device environment variable")
	}
	tests := []struct {
		name        string
		info        string
		wantDevices []int
	}{
		{name: "NPU device rule", info: raw, wantDevices: []int{7}},
		{name: "non-NPU device rule", info: strings.Replace(raw, `"major":236`, `"major":1`, 1)},
		{name: "no device rules", info: `{"pid":1234,"runtimeSpec":{"process":{"env":[]},"linux":{"resources":{"devices":[]}}}}`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			op := newCrioTestOperator(t, &crioV1Server{info: tc.info}, nil)
			parser := &DevicesParser{RuntimeOperator: op}
			deviceRulesChecked := false
			patch := gomonkey.ApplyFunc(npuMajor, func() []string {
				deviceRulesChecked = true
				return []string{"236"}
			})
			defer patch.Reset()
			results := make(chan DevicesInfo, 1)
			c := &CommonContainer{Id: "crio-id", Labels: map[string]string{
				labelK8sPodNamespace: "default", labelK8sPodName: "npu-pod", labelContainerName: "worker"}}
			if err := parser.parseDevices(context.Background(), c, results); err != nil {
				t.Fatal(err)
			}
			if !deviceRulesChecked {
				t.Fatal("device-rule branch was not exercised")
			}
			result := <-results
			if !reflect.DeepEqual(result.Devices, tc.wantDevices) {
				t.Fatalf("devices=%v, want %v", result.Devices, tc.wantDevices)
			}
			if len(tc.wantDevices) > 0 {
				if result.ID != c.Id || result.Name != "default_npu-pod_worker" {
					t.Fatalf("result=%+v", result)
				}
			} else if result.ID != "" {
				t.Fatalf("non-NPU container must not produce a device binding: %+v", result)
			}
		})
	}
}

func TestCrioSocketValidation(t *testing.T) {
	for _, tc := range []struct {
		name                   string
		pathErr, permissionErr error
	}{
		{"valid", nil, nil},
		{"nested symlink rejected", errors.New("can't support symlinks"), nil},
		{"owner or permission rejected", nil, errors.New("invalid permission")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			patches := gomonkey.ApplyFuncReturn(os.Readlink, "/run", nil)
			defer patches.Reset()
			patches.ApplyFunc(commonutils.CheckPath, func(path string) (string, error) {
				if path != "/run/crio/crio.sock" {
					t.Fatalf("checked path=%s", path)
				}
				return path, tc.pathErr
			})
			permissionChecked := false
			patches.ApplyFunc(commonutils.DoCheckOwnerAndPermission, func(path string, mode os.FileMode, uid uint32) error {
				permissionChecked = true
				if path != "/run/crio/crio.sock" || mode != excludePermissions || uid != 0 {
					t.Fatal("incorrect permission check")
				}
				return tc.permissionErr
			})
			op := &RuntimeOperatorTool{EndpointType: EndpointTypeCrio, CriEndpoint: DefaultCrioAddr}
			err := sockCheck(op)
			wantErr := tc.pathErr != nil || tc.permissionErr != nil
			if (err != nil) != wantErr {
				t.Fatalf("error=%v", err)
			}
			if permissionChecked != (tc.pathErr == nil) {
				t.Fatal("incorrect validation order")
			}
			wantEndpoint := DefaultCrioAddr
			if !wantErr {
				wantEndpoint = "unix:///run/crio/crio.sock"
			}
			if op.CriEndpoint != wantEndpoint {
				t.Fatalf("dial endpoint=%s", op.CriEndpoint)
			}
		})
	}
}

func TestCrioCustomSocketSymlinkRejected(t *testing.T) {
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(root, "real")
	if err := os.Mkdir(target, 0700); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(root, "link")
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}
	op := &RuntimeOperatorTool{EndpointType: EndpointTypeCrio,
		CriEndpoint: "unix://" + filepath.Join(link, "crio.sock")}
	if err := sockCheck(op); err == nil || !strings.Contains(err.Error(), "symlinks") {
		t.Fatalf("expected real filesystem symlink rejection, got %v", err)
	}
}
