/*
 * Copyright(C) 2026. Huawei Technologies Co.,Ltd. All rights reserved.
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at

 * http://www.apache.org/licenses/LICENSE-2.0

 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package plugin

import (
	"context"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"ascend-common/common-utils/healthz"
)

// TestNewDraHealthChecker verifies the constructor stores the config pointer.
func TestNewDraHealthChecker(t *testing.T) {
	cfg := &healthz.Config{EnableHealthz: true, HealthzAddress: "11251"}

	adp := NewDraHealthChecker(cfg)

	if adp == nil {
		t.Fatal("NewDraHealthChecker() = nil, want non-nil manager")
	}
	if adp.DraHealthzConfig != cfg {
		t.Errorf("DraHealthzConfig = %+v, want the config passed in", adp.DraHealthzConfig)
	}
}

// TestDraHealthChecker_Check verifies the built-in checker reports healthy.
func TestDraHealthChecker_Check(t *testing.T) {
	if err := (draHealthChecker{}).Check(context.Background()); err != nil {
		t.Errorf("draHealthChecker.Check() = %v, want nil", err)
	}
}

// TestStartHealthyCheck_Errors verifies StartHealthyCheck propagates config
// validation errors from the healthz package and returns nil when healthz is
// disabled. All cases fail before any network binding, so they are
// deterministic.
func TestStartHealthyCheck_Errors(t *testing.T) {
	tests := []struct {
		name        string
		cfg         *healthz.Config
		errContains string
	}{
		{"nil config", nil, "healthz config is nil"},
		{"empty address", &healthz.Config{EnableHealthz: true},
			"healthzAddress cannot be empty"},
		{"non-numeric port", &healthz.Config{EnableHealthz: true, HealthzAddress: "http"},
			"invalid port in healthzAddress"},
		{"port out of range", &healthz.Config{EnableHealthz: true, HealthzAddress: "80"},
			"out of range"},
		{"tls cert without key",
			&healthz.Config{EnableHealthz: true, HealthzAddress: "11251", TLSCertFile: "cert.pem"},
			"must be both set or both empty"},
		{"disabled returns nil", &healthz.Config{EnableHealthz: false}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adp := NewDraHealthChecker(tt.cfg)

			err := adp.StartHealthyCheck(context.Background())
			if tt.errContains == "" {
				if err != nil {
					t.Fatalf("StartHealthyCheck() = %v, want nil", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("StartHealthyCheck() = nil, want error containing %q", tt.errContains)
			}
			if !strings.Contains(err.Error(), tt.errContains) {
				t.Errorf("StartHealthyCheck() error = %v, want containing %q", err, tt.errContains)
			}
		})
	}
}

// TestRegisterHealthChecker verifies the registered DRA checkers keep the
// healthz probe healthy: start the server on a free port, GET /, and expect
// 200 "ok". A checker that always returns nil is not individually observable
// through the vendored package's public API, so this exercises the full
// register-and-serve wiring instead.
func TestRegisterHealthChecker(t *testing.T) {
	healthz.ClearHealthCheckers()
	defer healthz.ClearHealthCheckers()
	healthz.ResetLimiter()

	port := freeTCPPort(t)
	adp := NewDraHealthChecker(&healthz.Config{EnableHealthz: true, HealthzAddress: port})
	adp.RegisterHealthChecker()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := adp.StartHealthyCheck(ctx); err != nil {
		t.Fatalf("StartHealthyCheck() error = %v, want nil", err)
	}

	resp := probeHealthz(t, "http://127.0.0.1:"+port+"/")
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read probe body: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("probe status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	if string(body) != "ok" {
		t.Errorf("probe body = %q, want %q", body, "ok")
	}
}

// freeTCPPort asks the kernel for a free TCP port via an ephemeral listener.
func freeTCPPort(t *testing.T) string {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen on ephemeral port: %v", err)
	}
	port := strconv.Itoa(l.Addr().(*net.TCPAddr).Port)
	if err := l.Close(); err != nil {
		t.Fatalf("close ephemeral listener: %v", err)
	}
	return port
}

// probeHealthz GETs the healthz endpoint, retrying briefly while the server
// goroutine is still starting up.
func probeHealthz(t *testing.T, url string) *http.Response {
	t.Helper()
	for i := 0; i < 100; i++ {
		resp, err := http.Get(url)
		if err == nil {
			return resp
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("GET %s did not succeed within 1s", url)
	return nil
}
