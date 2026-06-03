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

// Package healthz provides a lightweight HTTP(S) health check server
// for Kubernetes liveness probes. It supports custom health check
// callbacks for business-level health detection.
package healthz

import (
	"context"
	"crypto/tls"
	"errors"
	"flag"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"golang.org/x/time/rate"

	"ascend-common/common-utils/hwlog"
)

const (
	defaultHealthzAddress  = "11251"
	healthzPath            = "/"
	httpScheme             = "http"
	httpsScheme            = "https"
	defaultShutDownTimeout = 5 * time.Second
	defaultReadTimeout     = 10 * time.Second
	defaultWriteTimeout    = 10 * time.Second
	defaultCheckerTimeout  = time.Second
	minPort                = 1025
	maxPort                = 65535
	defaultRateLimit       = 1
	defaultRateBurst       = 5
)

// HealthChecker is the interface for registering custom health check callbacks.
// Components can implement this interface to provide business-level health checks.
type HealthChecker interface {
	// Check performs a health check and returns nil if healthy,
	// or an error describing the failure if unhealthy.
	Check(ctx context.Context) error
}

// Config holds the configuration for the healthz server.
// It also serves as the parsed result of command-line flags via RegisterFlags.
type Config struct {
	EnableHealthz     bool
	HealthzAddress    string
	TLSCertFile       string
	TLSPrivateKeyFile string
}

var (
	globalCheckers []HealthChecker
	checkerMu      sync.RWMutex
	globalLimiter  = rate.NewLimiter(rate.Limit(defaultRateLimit), defaultRateBurst)
)

// RegisterHealthChecker appends a health check callback to the checker chain.
// All registered checkers must pass for the probe to return healthy.
// This function is thread-safe and can be called at any time.
func RegisterHealthChecker(checker HealthChecker) {
	checkerMu.Lock()
	defer checkerMu.Unlock()
	globalCheckers = append(globalCheckers, checker)
}

// ClearHealthCheckers removes all registered health check callbacks.
// Primarily useful for tests to reset state between cases.
func ClearHealthCheckers() {
	checkerMu.Lock()
	defer checkerMu.Unlock()
	globalCheckers = nil
}

// ResetLimiter replaces the global rate limiter with a fresh token bucket.
// Primarily useful for tests to ensure deterministic rate limiting behavior.
func ResetLimiter() {
	globalLimiter = rate.NewLimiter(rate.Limit(defaultRateLimit), defaultRateBurst)
}

func getHealthCheckers() []HealthChecker {
	checkerMu.RLock()
	defer checkerMu.RUnlock()
	return globalCheckers
}

// NewConfig creates a Config with default values.
func NewConfig() *Config {
	return &Config{
		EnableHealthz:     true,
		HealthzAddress:    defaultHealthzAddress,
		TLSCertFile:       "",
		TLSPrivateKeyFile: "",
	}
}

// RegisterFlags registers the four healthz flags with the global flag.CommandLine
// and returns a Config that will be populated by flag.Parse().
// Call before flag.Parse(), typically in a var block or init().
func RegisterFlags() *Config {
	cfg := &Config{
		EnableHealthz:  false,
		HealthzAddress: defaultHealthzAddress,
	}
	flag.BoolVar(&cfg.EnableHealthz, "enable-healthz", cfg.EnableHealthz,
		"Whether to enable health check service")
	flag.StringVar(&cfg.HealthzAddress, "healthz-address", cfg.HealthzAddress,
		"Health check service listen port")
	flag.StringVar(&cfg.TLSCertFile, "tls-cert-file", cfg.TLSCertFile,
		"TLS certificate file path for HTTPS")
	flag.StringVar(&cfg.TLSPrivateKeyFile, "tls-private-key-file", cfg.TLSPrivateKeyFile,
		"TLS private key file path for HTTPS")
	return cfg
}

// Serve starts the healthz server using the Config values.
func (c *Config) Serve(ctx context.Context) error {
	return StartServe(ctx, c)
}

func validateConfig(cfg *Config) error {
	if cfg == nil {
		return fmt.Errorf("healthz config is nil")
	}
	if !cfg.EnableHealthz {
		return nil
	}
	if cfg.HealthzAddress == "" {
		return fmt.Errorf("healthzAddress cannot be empty")
	}
	port, err := strconv.Atoi(cfg.HealthzAddress)
	if err != nil {
		return fmt.Errorf("invalid port in healthzAddress: %s", cfg.HealthzAddress)
	}
	if port < minPort || port > maxPort {
		return fmt.Errorf("port %d out of range [%d, %d]", port, minPort, maxPort)
	}
	if (cfg.TLSCertFile != "" && cfg.TLSPrivateKeyFile == "") ||
		(cfg.TLSCertFile == "" && cfg.TLSPrivateKeyFile != "") {
		return fmt.Errorf("tlsCertFile and tlsPrivateKeyFile must be both set or both empty")
	}
	return nil
}

// StartServe starts the healthz HTTP(S) server.
// It binds the port synchronously and returns an error if binding fails
// (e.g. port already in use). The server runs in a goroutine and will be
// gracefully shut down when ctx is cancelled.
func StartServe(ctx context.Context, cfg *Config) error {
	if err := validateConfig(cfg); err != nil {
		return err
	}
	if !cfg.EnableHealthz {
		hwlog.RunLog.Info("healthz server is disabled")
		return nil
	}

	mux := http.NewServeMux()
	mux.HandleFunc(healthzPath, healthzHandler)

	server := &http.Server{
		Handler:      mux,
		ReadTimeout:  defaultReadTimeout,
		WriteTimeout: defaultWriteTimeout,
	}

	if cfg.TLSCertFile != "" && cfg.TLSPrivateKeyFile != "" {
		server.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
	}

	ln, err := net.Listen("tcp", ":"+cfg.HealthzAddress)
	if err != nil {
		return fmt.Errorf("healthz server failed to listen on %s: %w", cfg.HealthzAddress, err)
	}

	go func() {
		scheme := httpScheme
		var serveErr error
		if cfg.TLSCertFile != "" && cfg.TLSPrivateKeyFile != "" {
			scheme = httpsScheme
			serveErr = server.ServeTLS(ln, cfg.TLSCertFile, cfg.TLSPrivateKeyFile)
		} else {
			serveErr = server.Serve(ln)
		}
		if serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
			hwlog.RunLog.Errorf("healthz server (%s) stopped with error: %v", scheme, serveErr)
		}
	}()

	go func() {
		<-ctx.Done()
		shutCtx, cancel := context.WithTimeout(context.Background(), defaultShutDownTimeout)
		defer cancel()
		if err := server.Shutdown(shutCtx); err != nil {
			hwlog.RunLog.Errorf("healthz server shutdown error: %v", err)
		}
	}()

	scheme := httpScheme
	if cfg.TLSCertFile != "" && cfg.TLSPrivateKeyFile != "" {
		scheme = httpsScheme
	}
	hwlog.RunLog.Infof("healthz server started on %s (%s)", cfg.HealthzAddress, scheme)
	return nil
}

func healthzHandler(w http.ResponseWriter, r *http.Request) {
	if !globalLimiter.Allow() {
		hwlog.RunLog.Warnf("healthz request rate limit exceeded, dropping request from %s", r.RemoteAddr)
		w.WriteHeader(http.StatusTooManyRequests)
		return
	}
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if r.URL.Path != "/" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	checkers := getHealthCheckers()
	for _, checker := range checkers {
		ctx, cancel := context.WithTimeout(r.Context(), defaultCheckerTimeout)
		done := make(chan error, 1)
		go func() {
			done <- checker.Check(ctx)
		}()

		var checkErr error
		select {
		case checkErr = <-done:
		case <-ctx.Done():
			checkErr = ctx.Err()
			hwlog.RunLog.Warnf("healthz checker timed out after %v", defaultCheckerTimeout)
		}
		cancel()

		if checkErr != nil {
			hwlog.RunLog.Errorf("healthz custom check failed: %v", checkErr)
			w.WriteHeader(http.StatusServiceUnavailable)
			if _, writeErr := w.Write([]byte(fmt.Sprintf("unhealthy: %v", checkErr))); writeErr != nil {
				hwlog.RunLog.Errorf("healthz write response failed: %v", writeErr)
			}
			return
		}
	}
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("ok")); err != nil {
		hwlog.RunLog.Errorf("healthz write response failed: %v", err)
	}
}
