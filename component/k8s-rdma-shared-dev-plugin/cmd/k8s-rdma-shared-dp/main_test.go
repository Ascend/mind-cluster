// main_test.go
// k8s-rdma-shared-dev-plugin 主包单元测试
//
// 说明:
//  1. 与 main.go 同包(package main), 以访问未导出函数/包级变量;
//  2. fault 包与 ub_device 接口的部分签名按 main.go 调用方式推断
//     (GetHcaDiscoverChan 返回 chan string、StartFaultReporting 入参 *kubernetes.Clientset),
//     若与实际定义不符, 编译期会报 "does not implement" 或运行期 gomonkey panic,
//     按注释提示同步调整 mock/打桩函数签名即可;
//  3. AddAnnotation 使用 httptest 假 apiserver 走真实 client-go 序列化路径,
//     可精确校验 JSON Patch 请求体与重试次数。
package main

import (
	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	ver "ascend-common/common-utils/version"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/Mellanox/k8s-rdma-shared-dev-plugin/pkg/fault"
	"github.com/Mellanox/k8s-rdma-shared-dev-plugin/pkg/resources/common"
	"github.com/Mellanox/k8s-rdma-shared-dev-plugin/pkg/resources/ub_device"
	"github.com/Mellanox/k8s-rdma-shared-dev-plugin/pkg/types"
	"github.com/Mellanox/k8s-rdma-shared-dev-plugin/pkg/utils"
	"github.com/agiledragon/gomonkey/v2"
	. "github.com/smartystreets/goconvey/convey"
	patchType "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
)

// TestMain 所有用例执行前初始化 hwlog,
// 保证被测函数内部的 hwlog.RunLog.Infof/Errorf 不会因日志未初始化而 panic。
func TestMain(m *testing.M) {
	tempDir, err := os.MkdirTemp("", "rdma-shared-dp-test")
	if err == nil {
		_ = hwlog.InitRunLogger(&hwlog.LogConfig{
			LogFileName:   filepath.Join(tempDir, "ut.log"),
			LogLevel:      0,
			MaxBackups:    1,
			MaxAge:        7,
			MaxLineLength: 1024,
		}, context.Background())
	}
	code := m.Run()
	if tempDir != "" {
		_ = os.RemoveAll(tempDir)
	}
	os.Exit(code)
}

// captureStdout 将标准输出重定向到管道, 返回读取函数与恢复函数,
// 用于捕获被测函数通过 fmt.Printf 打印的内容。
func captureStdout() (readFn func() string, restoreFn func(), err error) {
	r, w, e := os.Pipe()
	if e != nil {
		return nil, nil, e
	}
	orig := os.Stdout
	os.Stdout = w
	done := make(chan string, 1)
	go func() {
		data, _ := io.ReadAll(r)
		done <- string(data)
	}()
	readFn = func() string { return <-done }
	restoreFn = func() {
		os.Stdout = orig
		_ = w.Close()
	}
	return readFn, restoreFn, nil
}

// mockResourceManager types.ResourceManager 的 mock 实现。
// 通过内嵌接口规避未使用方法, 仅重写 initAndStartDevices 中会被调用的方法。
type mockResourceManager struct {
	types.ResourceManager
	mu                  sync.Mutex
	calls               []string
	readConfigErr       error
	validateConfigsErr  error
	validateRdmaModeErr error
	discoverDevicesErr  error
	initServersErr      error
	startAllServersErr  error
}

// 编译期校验: 若报 "does not implement", 说明 types.ResourceManager 方法签名与推断不一致, 请按实际接口调整。
var _ types.ResourceManager = (*mockResourceManager)(nil)

func (m *mockResourceManager) record(call string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls = append(m.calls, call)
}
func (m *mockResourceManager) getCalls() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]string(nil), m.calls...)
}
func (m *mockResourceManager) ReadConfig() error { m.record("ReadConfig"); return m.readConfigErr }
func (m *mockResourceManager) ValidateConfigs() error {
	m.record("ValidateConfigs")
	return m.validateConfigsErr
}
func (m *mockResourceManager) ValidateRdmaSystemMode() error {
	m.record("ValidateRdmaSystemMode")
	return m.validateRdmaModeErr
}
func (m *mockResourceManager) DiscoverHostDevices() error {
	m.record("DiscoverHostDevices")
	return m.discoverDevicesErr
}
func (m *mockResourceManager) InitServers() error { m.record("InitServers"); return m.initServersErr }
func (m *mockResourceManager) StartAllServers() error {
	m.record("StartAllServers")
	return m.startAllServersErr
}
func (m *mockResourceManager) PeriodicUpdate() func() { m.record("PeriodicUpdate"); return func() {} }

// mockUbResourceManager ub_device.UbResourceManager 的 mock 实现。
// 注意: GetHcaDiscoverChan 返回类型按 chan string 推断, 如实际接口不同请调整。
type mockUbResourceManager struct {
	ub_device.UbResourceManager
	mu             sync.Mutex
	hcaNames       []string
	unhealthyCalls [][]string
	configList     []*types.UserConfig
}

// 编译期校验: 若报 "does not implement", 请按实际 UbResourceManager 接口调整 mock 方法签名。
var _ ub_device.UbResourceManager = (*mockUbResourceManager)(nil)

func newMockUbResourceManager(hcaNames []string, configs ...*types.UserConfig) *mockUbResourceManager {
	return &mockUbResourceManager{hcaNames: hcaNames, configList: configs}
}
func (m *mockUbResourceManager) GetHcaNames() []string              { return m.hcaNames }
func (m *mockUbResourceManager) GetHcaDiscoverChan() chan string    { return make(chan string, 1) }
func (m *mockUbResourceManager) GetConfigList() []*types.UserConfig { return m.configList }
func (m *mockUbResourceManager) MarkDevicesUnhealthy(hcaNames []string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.unhealthyCalls = append(m.unhealthyCalls, hcaNames)
}
func (m *mockUbResourceManager) getUnhealthyCalls() [][]string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([][]string(nil), m.unhealthyCalls...)
}

// patchRecorder 模拟 kube-apiserver 的 Node PATCH 接口:
// 记录请求路径/请求体/Content-Type, 支持前 failTimes 次返回 500 以验证重试逻辑。
type patchRecorder struct {
	mu           sync.Mutex
	calls        int
	paths        []string
	bodies       []string
	contentTypes []string
	failTimes    int
}

func (p *patchRecorder) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	body, _ := io.ReadAll(req.Body)
	p.mu.Lock()
	p.calls++
	p.paths = append(p.paths, req.URL.Path)
	p.bodies = append(p.bodies, string(body))
	p.contentTypes = append(p.contentTypes, req.Header.Get("Content-Type"))
	call, failTimes := p.calls, p.failTimes
	p.mu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	if call <= failTimes {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"kind":"Status","apiVersion":"v1","status":"Failure","message":"mock error","code":500}`))
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"kind":"Node","apiVersion":"v1","metadata":{"name":"mock-node"}}`))
}
func (p *patchRecorder) snapshot() (int, []string, []string, []string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.calls,
		append([]string(nil), p.paths...),
		append([]string(nil), p.bodies...),
		append([]string(nil), p.contentTypes...)
}

// TestPrintVersionString 验证版本信息拼接格式。
func TestPrintVersionString(t *testing.T) {
	Convey("printVersionString 版本信息拼接格式", t, func() {
		origVersion, origCommit, origDate := version, commit, date
		Reset(func() { version, commit, date = origVersion, origCommit, origDate })
		// 测试条件: 设置包级变量 version/commit/date 为已知值。
		version = "v9.9.9"
		commit = "abcdef1"
		date = "2026-09-29"
		// 期望结果: 输出 "k8s-rdma-shared-dev-plugin version:<v>, commit:<c>, date:<d>"。
		So(printVersionString(), ShouldEqual,
			"k8s-rdma-shared-dev-plugin version:v9.9.9, commit:abcdef1, date:2026-09-29")
	})
}

// TestCheckLogParams 验证日志参数合法性校验(含全部边界值)。
func TestCheckLogParams(t *testing.T) {
	Convey("checkLogParams 日志参数合法性校验", t, func() {
		origLevel, origBackups, origAge := logLevel, logMaxBackups, logMaxAge
		Reset(func() { logLevel, logMaxBackups, logMaxAge = origLevel, origBackups, origAge })
		// 用例1: logLevel 合法下边界
		// 测试条件: logLevel=-1, 其余参数合法;
		// 期望结果: 返回 true。
		Convey("logLevel=-1 (合法下边界) 应返回 true", func() {
			logLevel, logMaxBackups, logMaxAge = -1, 3, 7
			So(checkLogParams(), ShouldBeTrue)
		})
		// 用例2: logLevel 合法上边界
		// 测试条件: logLevel=3, 其余参数合法;
		// 期望结果: 返回 true。
		Convey("logLevel=3 (合法上边界) 应返回 true", func() {
			logLevel, logMaxBackups, logMaxAge = 3, 3, 7
			So(checkLogParams(), ShouldBeTrue)
		})
		// 用例3: logLevel 低于下边界
		// 测试条件: logLevel=-2;
		// 期望结果: 返回 false。
		Convey("logLevel=-2 (低于下边界) 应返回 false", func() {
			logLevel, logMaxBackups, logMaxAge = -2, 3, 7
			So(checkLogParams(), ShouldBeFalse)
		})
		// 用例4: logLevel 高于上边界
		// 测试条件: logLevel=4;
		// 期望结果: 返回 false。
		Convey("logLevel=4 (高于上边界) 应返回 false", func() {
			logLevel, logMaxBackups, logMaxAge = 4, 3, 7
			So(checkLogParams(), ShouldBeFalse)
		})
		// 用例5: maxBackups 合法下边界
		// 测试条件: maxBackups=1;
		// 期望结果: 返回 true。
		Convey("maxBackups=1 (合法下边界) 应返回 true", func() {
			logLevel, logMaxBackups, logMaxAge = 0, 1, 7
			So(checkLogParams(), ShouldBeTrue)
		})
		// 用例6: maxBackups 合法上边界
		// 测试条件: maxBackups=hwlog.DefaultMaxBackups;
		// 期望结果: 返回 true。
		Convey(fmt.Sprintf("maxBackups=%d (合法上边界) 应返回 true", hwlog.DefaultMaxBackups), func() {
			logLevel, logMaxBackups, logMaxAge = 0, hwlog.DefaultMaxBackups, 7
			So(checkLogParams(), ShouldBeTrue)
		})
		// 用例7: maxBackups 低于下边界
		// 测试条件: maxBackups=0;
		// 期望结果: 返回 false。
		Convey("maxBackups=0 (低于下边界) 应返回 false", func() {
			logLevel, logMaxBackups, logMaxAge = 0, 0, 7
			So(checkLogParams(), ShouldBeFalse)
		})
		// 用例8: maxBackups 超过上限
		// 测试条件: maxBackups=hwlog.DefaultMaxBackups+1;
		// 期望结果: 返回 false。
		Convey(fmt.Sprintf("maxBackups=%d (超过上限) 应返回 false", hwlog.DefaultMaxBackups+1), func() {
			logLevel, logMaxBackups, logMaxAge = 0, hwlog.DefaultMaxBackups+1, 7
			So(checkLogParams(), ShouldBeFalse)
		})
		// 用例9: maxAge 合法下边界
		// 测试条件: maxAge=7;
		// 期望结果: 返回 true。
		Convey("maxAge=7 (合法下边界) 应返回 true", func() {
			logLevel, logMaxBackups, logMaxAge = 0, 3, 7
			So(checkLogParams(), ShouldBeTrue)
		})
		// 用例10: maxAge 合法上边界
		// 测试条件: maxAge=700;
		// 期望结果: 返回 true。
		Convey("maxAge=700 (合法上边界) 应返回 true", func() {
			logLevel, logMaxBackups, logMaxAge = 0, 3, 700
			So(checkLogParams(), ShouldBeTrue)
		})
		// 用例11: maxAge 低于下边界
		// 测试条件: maxAge=6;
		// 期望结果: 返回 false。
		Convey("maxAge=6 (低于下边界) 应返回 false", func() {
			logLevel, logMaxBackups, logMaxAge = 0, 3, 6
			So(checkLogParams(), ShouldBeFalse)
		})
		// 用例12: maxAge 超过上边界
		// 测试条件: maxAge=701;
		// 期望结果: 返回 false。
		Convey("maxAge=701 (超过上边界) 应返回 false", func() {
			logLevel, logMaxBackups, logMaxAge = 0, 3, 701
			So(checkLogParams(), ShouldBeFalse)
		})
	})
}

// TestInitFlags 验证命令行参数注册、默认值以及 -version/-v 绑定同一变量。
func TestInitFlags(t *testing.T) {
	Convey("initFlags 命令行参数注册", t, func() {
		// 用例1: 全部参数注册且默认值正确
		// 测试条件: 调用 initFlags 重建 FlagSet;
		// 期望结果: 所有业务参数可通过 flag.Lookup 查到且 DefValue 正确, hzFlags 完成初始化。
		Convey("应注册全部参数并设置正确默认值", func() {
			initFlags()
			expectedDefaults := map[string]string{
				"version":      "false",
				"v":            "false",
				"config-file":  common.DefaultConfigFilePath,
				"use-cdi":      "false",
				"ub-excl-mode": "false",
				"logLevel":     fmt.Sprintf("%d", defaultLogLevel),
				"maxBackups":   fmt.Sprintf("%d", defaultLogMaxBackups),
				"maxAge":       fmt.Sprintf("%d", defaultLogMaxAge),
				"logFile":      defaultLogFile,
			}
			for name, want := range expectedDefaults {
				f := flag.Lookup(name)
				So(f, ShouldNotBeNil)
				if f != nil {
					So(f.DefValue, ShouldEqual, want)
				}
			}
			So(hzFlags, ShouldNotBeNil)
		})
		// 用例2: -version 与 -v 绑定同一变量
		// 测试条件: 分别以 -version / -v 解析命令行;
		// 期望结果: 两种写法均将 versionOpt 置为 true。
		Convey("-version 与 -v 应绑定同一变量", func() {
			initFlags()
			So(flag.CommandLine.Parse([]string{"-version"}), ShouldBeNil)
			So(versionOpt, ShouldBeTrue)
			versionOpt = false
			So(flag.CommandLine.Parse([]string{"-v"}), ShouldBeNil)
			So(versionOpt, ShouldBeTrue)
		})
	})
}

// TestInitLogModule 验证日志模块初始化的成败两条路径。
func TestInitLogModule(t *testing.T) {
	Convey("initLogModule 日志模块初始化", t, func() {
		origLogFile := logFile
		Reset(func() { logFile = origLogFile })
		// 用例1: 初始化成功
		// 测试条件: gomonkey 打桩 hwlog.InitRunLogger 返回 nil;
		// 期望结果: initLogModule 返回 nil, 且传给 hwlog 的配置与包级参数一致。
		Convey("hwlog 初始化成功时应返回 nil 且配置透传正确", func() {
			var gotConfig *hwlog.LogConfig
			patches := gomonkey.ApplyFunc(hwlog.InitRunLogger,
				func(cfg *hwlog.LogConfig, ctx context.Context) error {
					gotConfig = cfg
					return nil
				})
			defer patches.Reset()
			logFile = filepath.Join(t.TempDir(), "success.log")
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			err := initLogModule(ctx)
			So(err, ShouldBeNil)
			So(gotConfig, ShouldNotBeNil)
			if gotConfig != nil {
				So(gotConfig.LogFileName, ShouldEqual, logFile)
				So(gotConfig.LogLevel, ShouldEqual, logLevel)
				So(gotConfig.MaxBackups, ShouldEqual, logMaxBackups)
				So(gotConfig.MaxAge, ShouldEqual, logMaxAge)
				So(gotConfig.MaxLineLength, ShouldEqual, maxLogLineLength)
			}
		})
		// 用例2: 初始化失败
		// 测试条件: 打桩 hwlog.InitRunLogger 返回指定错误;
		// 期望结果: initLogModule 透传同一错误对象。
		Convey("hwlog 初始化失败时应透传错误", func() {
			wantErr := fmt.Errorf("hwlog init failed for ut")
			patches := gomonkey.ApplyFunc(hwlog.InitRunLogger,
				func(cfg *hwlog.LogConfig, ctx context.Context) error { return wantErr })
			defer patches.Reset()
			logFile = filepath.Join(t.TempDir(), "failure.log")
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			So(initLogModule(ctx), ShouldEqual, wantErr)
		})
	})
}

// TestCreateK8sClient 验证集群内客户端构建。
func TestCreateK8sClient(t *testing.T) {
	Convey("createK8sClient 集群内客户端构建", t, func() {
		// 用例1: InClusterConfig 失败
		// 测试条件: 打桩 rest.InClusterConfig 返回错误;
		// 期望结果: 返回 nil 客户端及包含 "failed to get in-cluster config" 的错误。
		Convey("InClusterConfig 失败时应返回包装错误", func() {
			patches := gomonkey.ApplyFunc(rest.InClusterConfig, func() (*rest.Config, error) {
				return nil, fmt.Errorf("unable to load in-cluster configuration")
			})
			defer patches.Reset()
			client, err := createK8sClient()
			So(client, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "failed to get in-cluster config")
		})
		// 用例2: 正常构建
		// 测试条件: 打桩 rest.InClusterConfig 返回合法配置(NewForConfig 真实执行, 不会发起网络请求);
		// 期望结果: 返回非 nil 的 *kubernetes.Clientset 且无错误。
		Convey("配置正常时应返回可用 clientset", func() {
			patches := gomonkey.ApplyFunc(rest.InClusterConfig, func() (*rest.Config, error) {
				return &rest.Config{Host: "http://127.0.0.1:1"}, nil
			})
			defer patches.Reset()
			client, err := createK8sClient()
			So(err, ShouldBeNil)
			So(client, ShouldNotBeNil)
		})
	})
}

// TestAddAnnotation 验证节点注解 JSON Patch 写入: 键转义/请求体/重试/适配器委托。
func TestAddAnnotation(t *testing.T) {
	Convey("AddAnnotation 节点注解 JSON Patch 写入", t, func() {
		origNodeName, hadNodeName := os.LookupEnv(api.NodeNameEnv)
		_ = os.Setenv(api.NodeNameEnv, "mock-node")
		Reset(func() {
			if hadNodeName {
				_ = os.Setenv(api.NodeNameEnv, origNodeName)
			} else {
				_ = os.Unsetenv(api.NodeNameEnv)
			}
		})
		// 构造指向假 apiserver 的真实 clientset, 走真实序列化路径。
		newClient := func(recorder *patchRecorder) (*kubernetes.Clientset, func()) {
			server := httptest.NewServer(recorder)
			client, err := kubernetes.NewForConfig(&rest.Config{Host: server.URL})
			So(err, ShouldBeNil)
			return client, server.Close
		}
		// 用例1: 首次成功 + 特殊字符转义
		// 测试条件: key 同时包含 "~" 与 "/", apiserver 正常返回;
		// 期望结果: 无错误; PATCH 路径为 /api/v1/nodes/mock-node;
		//          Content-Type 为 application/json-patch+json;
		//          请求体中 "~"→"~0"、"/"→"~1", op=replace。
		Convey("key 含 ~/ 应转义且请求路径/类型/请求体正确", func() {
			recorder := &patchRecorder{}
			client, closeServer := newClient(recorder)
			defer closeServer()
			err := AddAnnotation(client, "example.com/key~with/slash", "value-1")
			So(err, ShouldBeNil)
			calls, paths, bodies, contentTypes := recorder.snapshot()
			So(calls, ShouldEqual, 1)
			So(paths[0], ShouldEqual, "/api/v1/nodes/mock-node")
			So(contentTypes[0], ShouldEqual, string(patchType.JSONPatchType))
			want, _ := json.Marshal([]interface{}{map[string]string{
				"op":    "replace",
				"path":  "/metadata/annotations/example.com~1key~0with~1slash",
				"value": "value-1",
			}})
			So(bodies[0], ShouldEqual, string(want))
		})
		// 用例2: 普通 key 无需转义
		// 测试条件: key 不含 "~" 与 "/";
		// 期望结果: path 原样拼接, 请求体正确。
		Convey("普通 key 无需转义", func() {
			recorder := &patchRecorder{}
			client, closeServer := newClient(recorder)
			defer closeServer()
			So(AddAnnotation(client, "plain-annotation", "v"), ShouldBeNil)
			_, _, bodies, _ := recorder.snapshot()
			want, _ := json.Marshal([]interface{}{map[string]string{
				"op": "replace", "path": "/metadata/annotations/plain-annotation", "value": "v",
			}})
			So(bodies[0], ShouldEqual, string(want))
		})
		// 用例3: 重试后成功
		// 测试条件: apiserver 前 2 次返回 500, 第 3 次成功; 打桩 time.Sleep 消除等待;
		// 期望结果: 恰好重试 3 次, 最终返回 nil。
		Convey("前 2 次失败第 3 次成功应恰好重试 3 次且返回 nil", func() {
			sleepPatch := gomonkey.ApplyFunc(time.Sleep, func(time.Duration) {})
			defer sleepPatch.Reset()
			recorder := &patchRecorder{failTimes: 2}
			client, closeServer := newClient(recorder)
			defer closeServer()
			So(AddAnnotation(client, "retry-key", "v"), ShouldBeNil)
			calls, _, _, _ := recorder.snapshot()
			So(calls, ShouldEqual, retryTime)
		})
		// 用例4: 重试耗尽
		// 测试条件: apiserver 连续 3 次返回 500;
		// 期望结果: 恰好调用 3 次后停止, 返回非 nil 错误。
		Convey("连续 3 次失败应恰好重试 3 次并返回错误", func() {
			sleepPatch := gomonkey.ApplyFunc(time.Sleep, func(time.Duration) {})
			defer sleepPatch.Reset()
			recorder := &patchRecorder{failTimes: retryTime}
			client, closeServer := newClient(recorder)
			defer closeServer()
			So(AddAnnotation(client, "always-fail", "v"), ShouldNotBeNil)
			calls, _, _, _ := recorder.snapshot()
			So(calls, ShouldEqual, retryTime)
		})
		// 用例5: 适配器委托
		// 测试条件: 使用 clientsetAnnotationAdder 适配器调用 AddAnnotation;
		// 期望结果: 委托到包级 AddAnnotation, 请求体转义与赋值正确且无错误。
		Convey("clientsetAnnotationAdder 适配器应正确委托", func() {
			recorder := &patchRecorder{}
			client, closeServer := newClient(recorder)
			defer closeServer()
			adder := clientsetAnnotationAdder{clientset: client}
			So(adder.AddAnnotation("adapter/key", "adapter-value"), ShouldBeNil)
			_, _, bodies, _ := recorder.snapshot()
			want, _ := json.Marshal([]interface{}{map[string]string{
				"op": "replace", "path": "/metadata/annotations/adapter~1key", "value": "adapter-value",
			}})
			So(bodies[0], ShouldEqual, string(want))
		})
	})
}

// TestWriteDpuResourceAnnotation 验证 DPU 资源注解写入节点逻辑。
func TestWriteDpuResourceAnnotation(t *testing.T) {
	Convey("writeDpuResourceAnnotation DPU 资源注解写入", t, func() {
		// 用例1: 获取节点名失败
		// 测试条件: 打桩 utils.GetNodeName 返回错误;
		// 期望结果: 原样透传该错误, utils.UpdateNodeAnnotation 不被调用。
		Convey("获取节点名失败时应透传错误且不写注解", func() {
			updateCalled := false
			patches := gomonkey.
				ApplyFunc(utils.GetNodeName, func() (string, error) { return "", fmt.Errorf("no node name") }).
				ApplyFunc(utils.UpdateNodeAnnotation,
					func(kubernetes.Interface, string, string, string) error { updateCalled = true; return nil })
			defer patches.Reset()
			err := writeDpuResourceAnnotation(fake.NewSimpleClientset(), nil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "no node name")
			So(updateCalled, ShouldBeFalse)
		})
		// 用例2: 成功写入
		// 测试条件: GetNodeName 返回 node-1, configList 含两条资源配置;
		// 期望结果: key 为 common.DPUResourceAnnotationKey,
		//          value 按 prefix/name 拼接并以逗号分隔, 顺序保持不变。
		Convey("成功时应按 prefix/name 逗号拼接写入注解", func() {
			var gotNode, gotKey, gotValue string
			patches := gomonkey.
				ApplyFunc(utils.GetNodeName, func() (string, error) { return "node-1", nil }).
				ApplyFunc(utils.UpdateNodeAnnotation,
					func(_ kubernetes.Interface, node, key, value string) error {
						gotNode, gotKey, gotValue = node, key, value
						return nil
					})
			defer patches.Reset()
			configs := []*types.UserConfig{
				{ResourcePrefix: "rdma", ResourceName: "hca_shared"},
				{ResourcePrefix: "rdma", ResourceName: "ub_dev"},
			}
			So(writeDpuResourceAnnotation(fake.NewSimpleClientset(), configs), ShouldBeNil)
			So(gotNode, ShouldEqual, "node-1")
			So(gotKey, ShouldEqual, common.DPUResourceAnnotationKey)
			So(gotValue, ShouldEqual, "rdma/hca_shared,rdma/ub_dev")
		})
		// 用例3: 更新注解失败
		// 测试条件: 打桩 utils.UpdateNodeAnnotation 返回错误;
		// 期望结果: 原样透传该错误。
		Convey("更新注解失败时应透传错误", func() {
			wantErr := fmt.Errorf("update annotation failed")
			patches := gomonkey.
				ApplyFunc(utils.GetNodeName, func() (string, error) { return "node-1", nil }).
				ApplyFunc(utils.UpdateNodeAnnotation,
					func(kubernetes.Interface, string, string, string) error { return wantErr })
			defer patches.Reset()
			So(writeDpuResourceAnnotation(fake.NewSimpleClientset(), nil), ShouldEqual, wantErr)
		})
	})
}

// TestStartFaultDetection 验证故障检测的启动条件与回调联动。
// 注意: fault 包函数签名按 main.go 调用方式推断, 若实际定义不同请同步调整打桩签名。
func TestStartFaultDetection(t *testing.T) {
	Convey("startFaultDetection 故障检测启停控制", t, func() {
		// 用例1: 探测周期低于最小值
		// 测试条件: faultDetectPeriod < common.MinFaultDetectionPeriod;
		// 期望结果: fault.StartFaultDetection 与 fault.StartFaultReporting 均不被调用。
		Convey("探测周期低于最小值时应禁用故障检测", func() {
			detectCalled, reportCalled := false, false
			patches := gomonkey.
				ApplyFunc(fault.StartFaultDetection,
					func(context.Context, func() []string, chan string, int, func([]string)) { detectCalled = true }).
				ApplyFunc(fault.StartFaultReporting,
					func(context.Context, *kubernetes.Clientset) { reportCalled = true })
			defer patches.Reset()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			startFaultDetection(ctx, newMockUbResourceManager([]string{"hca-0"}),
				common.MinFaultDetectionPeriod-1, nil)
			So(detectCalled, ShouldBeFalse)
			So(reportCalled, ShouldBeFalse)
		})
		// 用例2: k8s 客户端为 nil
		// 测试条件: 探测周期合法但 k8sClient 为 nil;
		// 期望结果: 不启动故障检测与上报, 函数安全返回。
		Convey("k8s 客户端为 nil 时不应启动故障检测", func() {
			detectCalled, reportCalled := false, false
			patches := gomonkey.
				ApplyFunc(fault.StartFaultDetection,
					func(context.Context, func() []string, chan string, int, func([]string)) { detectCalled = true }).
				ApplyFunc(fault.StartFaultReporting,
					func(context.Context, *kubernetes.Clientset) { reportCalled = true })
			defer patches.Reset()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			startFaultDetection(ctx, newMockUbResourceManager([]string{"hca-0"}),
				common.MinFaultDetectionPeriod, nil)
			So(detectCalled, ShouldBeFalse)
			So(reportCalled, ShouldBeFalse)
		})
		// 用例3: 正常启动
		// 测试条件: 探测周期合法且 k8s 客户端有效, 打桩两个 fault 入口函数;
		// 期望结果: 异步启动检测与上报; 打桩函数收到的 HCA 列表与周期正确;
		//          通过健康回调触发 MarkDevicesUnhealthy, 记录内容与入参一致。
		Convey("正常场景应异步启动检测/上报并正确联动健康回调", func() {
			started := make(chan struct{}, 2)
			var gotPeriod int
			var gotHcaNames []string
			var healthCb func([]string)
			patches := gomonkey.
				ApplyFunc(fault.StartFaultDetection,
					func(ctx context.Context, getHca func() []string, ch chan string, period int, cb func([]string)) {
						gotPeriod = period
						gotHcaNames = getHca()
						healthCb = cb
						started <- struct{}{}
					}).
				ApplyFunc(fault.StartFaultReporting,
					func(ctx context.Context, client *kubernetes.Clientset) { started <- struct{}{} })
			defer patches.Reset()
			rm := newMockUbResourceManager([]string{"hca-0", "hca-1"})
			realClient, err := kubernetes.NewForConfig(&rest.Config{Host: "http://127.0.0.1:1"})
			So(err, ShouldBeNil)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			startFaultDetection(ctx, rm, common.MinFaultDetectionPeriod+10, realClient)
			<-started // 等待检测协程启动
			<-started // 等待上报协程启动
			So(gotPeriod, ShouldEqual, common.MinFaultDetectionPeriod+10)
			So(gotHcaNames, ShouldResemble, []string{"hca-0", "hca-1"})
			healthCb([]string{"hca-0"})
			So(rm.getUnhealthyCalls(), ShouldResemble, [][]string{{"hca-0"}})
		})
	})
}

// TestInitAndStartDevices 验证设备管理器初始化与启动全流程。
func TestInitAndStartDevices(t *testing.T) {
	Convey("initAndStartDevices 设备管理器初始化与启动流程", t, func() {
		// 用例1: PCI 类型完整流程
		// 测试条件: deviceType=PCI, 各步骤均成功;
		// 期望结果: ReadConfig→ValidateConfigs→ValidateRdmaSystemMode→DiscoverHostDevices
		//          →InitServers→StartAllServers→PeriodicUpdate 全部执行,
		//          返回原资源管理器与非 nil 的周期更新停止函数。
		Convey("PCI 类型应执行全部流程(含 ValidateRdmaSystemMode)", func() {
			rm := &mockResourceManager{}
			gotRm, stop := initAndStartDevices("PCI", func() types.ResourceManager { return rm })
			So(gotRm, ShouldEqual, rm)
			So(stop, ShouldNotBeNil)
			calls := rm.getCalls()
			So(calls, ShouldContain, "ReadConfig")
			So(calls, ShouldContain, "ValidateConfigs")
			So(calls, ShouldContain, "ValidateRdmaSystemMode")
			So(calls, ShouldContain, "DiscoverHostDevices")
			So(calls, ShouldContain, "InitServers")
			So(calls, ShouldContain, "StartAllServers")
			So(calls, ShouldContain, "PeriodicUpdate")
		})
		// 用例2: 非 PCI 类型
		// 测试条件: deviceType=UB;
		// 期望结果: 不执行 ValidateRdmaSystemMode, 其余步骤正常完成。
		Convey("非 PCI 类型不应执行 ValidateRdmaSystemMode", func() {
			rm := &mockResourceManager{}
			gotRm, stop := initAndStartDevices("UB", func() types.ResourceManager { return rm })
			So(gotRm, ShouldNotBeNil)
			So(stop, ShouldNotBeNil)
			calls := rm.getCalls()
			So(calls, ShouldNotContain, "ValidateRdmaSystemMode")
			So(calls, ShouldContain, "StartAllServers")
		})
		// 用例3: 中间步骤失败
		// 测试条件: 各步骤均返回错误(仅记录日志不提前返回);
		// 期望结果: 全部步骤(含 PeriodicUpdate)仍被执行, 不 panic, 正常返回。
		Convey("中间步骤失败不应中断启动流程", func() {
			rm := &mockResourceManager{
				readConfigErr:      fmt.Errorf("read config error"),
				validateConfigsErr: fmt.Errorf("invalid config"),
				discoverDevicesErr: fmt.Errorf("discover error"),
				initServersErr:     fmt.Errorf("init servers error"),
				startAllServersErr: fmt.Errorf("start servers error"),
			}
			gotRm, stop := initAndStartDevices("PCI", func() types.ResourceManager { return rm })
			So(gotRm, ShouldNotBeNil)
			So(stop, ShouldNotBeNil)
			calls := rm.getCalls()
			So(calls, ShouldContain, "StartAllServers")
			So(calls, ShouldContain, "PeriodicUpdate")
		})
	})
}

// TestAddVersionToNodeAnnotation 验证版本信息写入节点注解。
func TestAddVersionToNodeAnnotation(t *testing.T) {
	Convey("addVersionToNodeAnnotation 版本信息写入节点注解", t, func() {
		info := ver.Info{
			Version:   "v1.2.3",
			GitCommit: "abcdef1",
			GitBranch: "master",
			BuildOS:   "linux",
			BuildArch: "amd64",
			GoVersion: "go1.21",
		}
		// 用例1: InClusterConfig 失败
		// 测试条件: 打桩 rest.InClusterConfig 返回错误;
		// 期望结果: 函数安全返回(仅记录日志), ver.ReportVersionToNodeAnnotation 不被调用, 无 panic。
		Convey("InClusterConfig 失败时应安全返回且不上报版本", func() {
			reportCalled := false
			patches := gomonkey.
				ApplyFunc(rest.InClusterConfig,
					func() (*rest.Config, error) { return nil, fmt.Errorf("no in-cluster config") }).
				ApplyFunc(ver.ReportVersionToNodeAnnotation,
					func(ver.AnnotationAdder, ver.Info, string) error { reportCalled = true; return nil })
			defer patches.Reset()
			So(func() { addVersionToNodeAnnotation(info, "k8s-rdma-shared-dp") }, ShouldNotPanic)
			So(reportCalled, ShouldBeFalse)
		})
		// 用例2: 正常上报
		// 测试条件: InClusterConfig 返回合法配置, 打桩 ver.ReportVersionToNodeAnnotation;
		// 期望结果: 以组件名 k8s-rdma-shared-dp 与原始版本信息调用上报接口, 且传入非 nil 适配器。
		Convey("正常场景应以组件名 k8s-rdma-shared-dp 上报版本信息", func() {
			var gotAdder ver.AnnotationAdder
			var gotInfo ver.Info
			var gotName string
			patches := gomonkey.
				ApplyFunc(rest.InClusterConfig,
					func() (*rest.Config, error) { return &rest.Config{Host: "http://127.0.0.1:1"}, nil }).
				ApplyFunc(ver.ReportVersionToNodeAnnotation,
					func(adder ver.AnnotationAdder, i ver.Info, name string) error {
						gotAdder, gotInfo, gotName = adder, i, name
						return nil
					})
			defer patches.Reset()
			addVersionToNodeAnnotation(info, "k8s-rdma-shared-dp")
			So(gotName, ShouldEqual, "k8s-rdma-shared-dp")
			So(gotInfo.Version, ShouldEqual, "v1.2.3")
			So(gotAdder, ShouldNotBeNil)
		})
		// 用例3: 上报失败
		// 测试条件: 打桩 ver.ReportVersionToNodeAnnotation 返回错误;
		// 期望结果: 仅记录错误日志, 不 panic。
		Convey("版本上报失败时应记录错误且不 panic", func() {
			patches := gomonkey.
				ApplyFunc(rest.InClusterConfig,
					func() (*rest.Config, error) { return &rest.Config{Host: "http://127.0.0.1:1"}, nil }).
				ApplyFunc(ver.ReportVersionToNodeAnnotation,
					func(ver.AnnotationAdder, ver.Info, string) error { return fmt.Errorf("report version error") })
			defer patches.Reset()
			So(func() { addVersionToNodeAnnotation(info, "k8s-rdma-shared-dp") }, ShouldNotPanic)
		})
	})
}

// TestMainVersionFlag 验证 main 的版本参数短路逻辑。
func TestMainVersionFlag(t *testing.T) {
	Convey("main 传入版本参数时应打印版本信息后立即返回", t, func() {
		for _, flagName := range []string{"-v", "-version"} {
			// 用例: -v / -version 均应触发版本打印短路
			// 测试条件: os.Args 模拟版本参数, 打桩 ver.Get 返回已知信息;
			//          打桩 hwlog.InitRunLogger 返回错误以拦截"未短路"的情况;
			// 期望结果: stdout 输出包含 version/commit/branch 字段, 不进入日志初始化等后续流程。
			Convey(fmt.Sprintf("参数 %s 应打印版本信息后立即返回", flagName), func() {
				patches := gomonkey.
					ApplyGlobalVar(&os.Args, []string{"k8s-rdma-shared-dp", flagName}).
					ApplyFunc(ver.Get, func() ver.Info {
						return ver.Info{Version: "ut-version", GitCommit: "ut-commit", GitBranch: "ut-branch",
							BuildOS: "linux", BuildArch: "amd64", GoVersion: "go-ut"}
					}).
					ApplyFunc(hwlog.InitRunLogger,
						func(*hwlog.LogConfig, context.Context) error { return fmt.Errorf("should not be called") })
				defer patches.Reset()
				readOut, restoreOut, err := captureStdout()
				So(err, ShouldBeNil)
				main() // 应在打印版本后直接 return
				restoreOut()
				out := readOut()
				So(out, ShouldContainSubstring, "version=ut-version")
				So(out, ShouldContainSubstring, "commit=ut-commit")
				So(out, ShouldContainSubstring, "branch=ut-branch")
			})
		}
	})
}

// TestMainInvalidLogLevel 验证 main 对非法日志参数的短路逻辑。
func TestMainInvalidLogLevel(t *testing.T) {
	Convey("main 传入非法 logLevel 时应在初始化日志前直接返回", t, func() {
		// 测试条件: os.Args 模拟 [-logLevel 99](超出合法范围 [-1,3]),
		//          打桩 hwlog.InitRunLogger 以探测是否被调用;
		// 期望结果: stdout 提示 "Invalid logLevel 99", hwlog.InitRunLogger 不被调用, 无 panic。
		initLogCalled := false
		patches := gomonkey.
			ApplyGlobalVar(&os.Args, []string{"k8s-rdma-shared-dp", "-logLevel", "99"}).
			ApplyFunc(hwlog.InitRunLogger,
				func(*hwlog.LogConfig, context.Context) error { initLogCalled = true; return nil })
		defer patches.Reset()
		readOut, restoreOut, err := captureStdout()
		So(err, ShouldBeNil)
		So(func() { main() }, ShouldNotPanic)
		restoreOut()
		out := readOut()
		So(out, ShouldContainSubstring, "Invalid logLevel 99")
		So(initLogCalled, ShouldBeFalse)
	})
}
