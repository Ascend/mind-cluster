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

package healthz

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	"ascend-common/common-utils/hwlog"
)

func init() {
	if err := hwlog.InitRunLogger(&hwlog.LogConfig{OnlyToStdout: true}, context.Background()); err != nil {
		fmt.Printf("init run logger failed: %v\n", err)
	}
}

type validateConfigTestCase struct {
	name    string
	cfg     *Config
	wantErr bool
}

func newValidateConfigTestCases() []validateConfigTestCase {
	return []validateConfigTestCase{
		{
			name:    "nil config should return error",
			cfg:     nil,
			wantErr: true,
		},
		{
			name:    "disabled healthz should pass",
			cfg:     &Config{EnableHealthz: false},
			wantErr: false,
		},
		{
			name:    "valid default config should pass",
			cfg:     NewConfig(),
			wantErr: false,
		},
		{
			name:    "empty address should return error",
			cfg:     &Config{EnableHealthz: true, HealthzAddress: ""},
			wantErr: true,
		},
		{
			name:    "address with colon prefix should return error",
			cfg:     &Config{EnableHealthz: true, HealthzAddress: ":11251"},
			wantErr: true,
		},
		{
			name:    "invalid port should return error",
			cfg:     &Config{EnableHealthz: true, HealthzAddress: "abc"},
			wantErr: true,
		},
		{
			name:    "port out of range low should return error",
			cfg:     &Config{EnableHealthz: true, HealthzAddress: "80"},
			wantErr: true,
		},
		{
			name:    "port out of range high should return error",
			cfg:     &Config{EnableHealthz: true, HealthzAddress: "70000"},
			wantErr: true,
		},
		{
			name: "only cert file set should return error",
			cfg: &Config{
				EnableHealthz: true, HealthzAddress: "11251",
				TLSCertFile: "/path/to/cert",
			},
			wantErr: true,
		},
		{
			name: "only key file set should return error",
			cfg: &Config{
				EnableHealthz: true, HealthzAddress: "11251",
				TLSPrivateKeyFile: "/path/to/key",
			},
			wantErr: true,
		},
		{
			name: "both tls files set should pass",
			cfg: &Config{
				EnableHealthz: true, HealthzAddress: "11251",
				TLSCertFile: "/path/to/cert", TLSPrivateKeyFile: "/path/to/key",
			},
			wantErr: false,
		},
		{
			name:    "port at min boundary should pass",
			cfg:     &Config{EnableHealthz: true, HealthzAddress: "1025"},
			wantErr: false,
		},
		{
			name:    "port at max boundary should pass",
			cfg:     &Config{EnableHealthz: true, HealthzAddress: "65535"},
			wantErr: false,
		},
	}
}

func TestValidateConfig(t *testing.T) {
	for _, tc := range newValidateConfigTestCases() {
		t.Run(tc.name, func(t *testing.T) {
			err := validateConfig(tc.cfg)
			if (err != nil) != tc.wantErr {
				t.Errorf("validateConfig() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

func TestNewConfig(t *testing.T) {
	cfg := NewConfig()
	if !cfg.EnableHealthz {
		t.Error("default EnableHealthz should be true")
	}
	if cfg.HealthzAddress != defaultHealthzAddress {
		t.Errorf("default HealthzAddress = %s, want %s", cfg.HealthzAddress, defaultHealthzAddress)
	}
	if cfg.TLSCertFile != "" || cfg.TLSPrivateKeyFile != "" {
		t.Error("default TLS fields should be empty")
	}
}

func TestStartServeDisabled(t *testing.T) {
	cfg := &Config{EnableHealthz: false}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := StartServe(ctx, cfg); err != nil {
		t.Errorf("StartServe with disabled should not return error, got: %v", err)
	}
}

func TestStartServeAndRequest(t *testing.T) {
	ResetLimiter()
	cfg := NewConfig()
	cfg.HealthzAddress = "11252"
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := StartServe(ctx, cfg); err != nil {
		t.Fatalf("StartServe failed: %v", err)
	}

	resp, err := http.Get("http://localhost:11252/")
	if err != nil {
		t.Fatalf("GET / failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestHealthzHandlerMethodNotAllowed(t *testing.T) {
	ResetLimiter()
	cfg := NewConfig()
	cfg.HealthzAddress = "11253"
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := StartServe(ctx, cfg); err != nil {
		t.Fatalf("StartServe failed: %v", err)
	}

	req, _ := http.NewRequest(http.MethodPost, "http://localhost:11253/", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST / failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", resp.StatusCode)
	}
}

type mockHealthChecker struct {
	err error
}

func (m *mockHealthChecker) Check(_ context.Context) error {
	return m.err
}

func TestStartServeWithHealthyChecker(t *testing.T) {
	ResetLimiter()
	RegisterHealthChecker(&mockHealthChecker{err: nil})
	defer ClearHealthCheckers()

	cfg := NewConfig()
	cfg.HealthzAddress = "11254"
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := StartServe(ctx, cfg); err != nil {
		t.Fatalf("StartServe failed: %v", err)
	}

	resp, err := http.Get("http://localhost:11254/")
	if err != nil {
		t.Fatalf("GET / failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestStartServeWithUnhealthyChecker(t *testing.T) {
	ResetLimiter()
	RegisterHealthChecker(&mockHealthChecker{err: fmt.Errorf("db connection lost")})
	defer ClearHealthCheckers()

	cfg := NewConfig()
	cfg.HealthzAddress = "11255"
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := StartServe(ctx, cfg); err != nil {
		t.Fatalf("StartServe failed: %v", err)
	}

	resp, err := http.Get("http://localhost:11255/")
	if err != nil {
		t.Fatalf("GET / failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", resp.StatusCode)
	}
}

func TestStartServeGracefulShutdown(t *testing.T) {
	ResetLimiter()
	cfg := NewConfig()
	cfg.HealthzAddress = "11256"
	ctx, cancel := context.WithCancel(context.Background())
	if err := StartServe(ctx, cfg); err != nil {
		t.Fatalf("StartServe failed: %v", err)
	}

	resp, err := http.Get("http://localhost:11256/")
	if err != nil {
		t.Fatalf("GET / before shutdown failed: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 before shutdown, got %d", resp.StatusCode)
	}

	cancel()
	time.Sleep(1 * time.Second)

	_, err = http.Get("http://localhost:11256/")
	if err == nil {
		t.Error("expected connection error after shutdown")
	}
}

func TestRegisterHealthCheckerConcurrent(t *testing.T) {
	done := make(chan struct{})
	go func() {
		for i := 0; i < 100; i++ {
			RegisterHealthChecker(&mockHealthChecker{err: nil})
		}
		done <- struct{}{}
	}()
	go func() {
		for i := 0; i < 100; i++ {
			_ = getHealthCheckers()
		}
		done <- struct{}{}
	}()
	<-done
	<-done
	checkers := getHealthCheckers()
	if len(checkers) != 100 {
		t.Errorf("expected 100 checkers, got %d", len(checkers))
	}
	ClearHealthCheckers()
}

func TestMultipleCheckersAllHealthy(t *testing.T) {
	ResetLimiter()
	RegisterHealthChecker(&mockHealthChecker{err: nil})
	RegisterHealthChecker(&mockHealthChecker{err: nil})
	defer ClearHealthCheckers()

	cfg := NewConfig()
	cfg.HealthzAddress = "11257"
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := StartServe(ctx, cfg); err != nil {
		t.Fatalf("StartServe failed: %v", err)
	}

	resp, err := http.Get("http://localhost:11257/")
	if err != nil {
		t.Fatalf("GET / failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 with all checkers healthy, got %d", resp.StatusCode)
	}
}

func TestMultipleCheckersFirstFails(t *testing.T) {
	ResetLimiter()
	RegisterHealthChecker(&mockHealthChecker{err: fmt.Errorf("first failed")})
	RegisterHealthChecker(&mockHealthChecker{err: nil})
	defer ClearHealthCheckers()

	cfg := NewConfig()
	cfg.HealthzAddress = "11258"
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := StartServe(ctx, cfg); err != nil {
		t.Fatalf("StartServe failed: %v", err)
	}

	resp, err := http.Get("http://localhost:11258/")
	if err != nil {
		t.Fatalf("GET / failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("expected 503 when first checker fails, got %d", resp.StatusCode)
	}
}

func TestMultipleCheckersLastFails(t *testing.T) {
	ResetLimiter()
	RegisterHealthChecker(&mockHealthChecker{err: nil})
	RegisterHealthChecker(&mockHealthChecker{err: fmt.Errorf("last failed")})
	defer ClearHealthCheckers()

	cfg := NewConfig()
	cfg.HealthzAddress = "11259"
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := StartServe(ctx, cfg); err != nil {
		t.Fatalf("StartServe failed: %v", err)
	}

	resp, err := http.Get("http://localhost:11259/")
	if err != nil {
		t.Fatalf("GET / failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("expected 503 when last checker fails, got %d", resp.StatusCode)
	}
}

type slowHealthChecker struct{}

func (s *slowHealthChecker) Check(_ context.Context) error {
	time.Sleep(2 * time.Second)
	return nil
}

func TestStartServeCheckerTimeout(t *testing.T) {
	ResetLimiter()
	RegisterHealthChecker(&slowHealthChecker{})
	defer ClearHealthCheckers()

	cfg := NewConfig()
	cfg.HealthzAddress = "11262"
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := StartServe(ctx, cfg); err != nil {
		t.Fatalf("StartServe failed: %v", err)
	}

	resp, err := http.Get("http://localhost:11262/")
	if err != nil {
		t.Fatalf("GET / failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("expected 503 for timed-out checker, got %d", resp.StatusCode)
	}
}

func TestStartServePortConflict(t *testing.T) {
	ln, err := net.Listen("tcp", ":11260")
	if err != nil {
		t.Fatalf("failed to set up listener: %v", err)
	}
	defer ln.Close()

	cfg := NewConfig()
	cfg.HealthzAddress = "11260"
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := StartServe(ctx, cfg); err == nil {
		t.Error("expected error for port conflict, got nil")
	}
}

func TestClearHealthCheckers(t *testing.T) {
	RegisterHealthChecker(&mockHealthChecker{err: fmt.Errorf("should be cleared")})
	ClearHealthCheckers()
	checkers := getHealthCheckers()
	if len(checkers) != 0 {
		t.Errorf("expected 0 checkers after clear, got %d", len(checkers))
	}
}

func TestHealthzHandlerRateLimit(t *testing.T) {
	ResetLimiter()
	defer ResetLimiter()

	cfg := NewConfig()
	cfg.HealthzAddress = "11261"
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := StartServe(ctx, cfg); err != nil {
		t.Fatalf("StartServe failed: %v", err)
	}

	var rateLimited bool
	for i := 0; i < 20; i++ {
		resp, err := http.Get("http://localhost:11261/")
		if err != nil {
			t.Fatalf("GET / failed: %v", err)
		}
		if resp.StatusCode == http.StatusTooManyRequests {
			rateLimited = true
			resp.Body.Close()
			break
		}
		if resp.StatusCode != http.StatusOK {
			t.Errorf("unexpected status code: %d", resp.StatusCode)
		}
		resp.Body.Close()
	}
	if !rateLimited {
		t.Error("expected at least one rate-limited request (429), but all passed")
	}
}
