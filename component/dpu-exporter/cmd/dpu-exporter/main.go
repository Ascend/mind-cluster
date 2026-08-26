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

// dpu-exporter is an exporter for Huawei DPU (Data Processing Unit) metrics.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"huawei.com/dpu-exporter/pkg/collector/dpucollector"
	"huawei.com/dpu-exporter/pkg/collector/metricscollector"
	"huawei.com/dpu-exporter/pkg/configmanager"
	"huawei.com/dpu-exporter/pkg/device"
	prompkg "huawei.com/dpu-exporter/pkg/platform/prometheus"
	"huawei.com/dpu-exporter/utils/logger"
)

const (
	// shutdownTimeout is the max time to wait for graceful shutdown
	shutdownTimeout = 5 * time.Second
)

func main() {
	configFile := flag.String("config", "", "path to config file (default: /etc/dpu-exporter/config.json)")
	port := flag.Int("port", 8080, "HTTP port for metrics server")
	cardType := flag.String("cardType", device.CardTypeHuawei, "DPU card type (e.g. 'huawei', currently only 'huawei' is supported)")
	flag.IntVar(&logger.HwLogConfig.LogLevel, "logLevel", 0, "log level (-1-debug, 0-info, 1-warning, 2-error 3-critical)")
	flag.IntVar(&logger.HwLogConfig.MaxAge, "maxAge", logger.HwLogConfig.MaxAge, "max age of backup logs in days, range is [7, 700]")
	flag.IntVar(&logger.HwLogConfig.MaxBackups, "maxBackups", logger.HwLogConfig.MaxBackups, "max number of backup log files, range is (0, 180]")
	flag.Parse()

	httpPort := *port
	// Override config file path if provided via CLI
	if *configFile != "" {
		configmanager.SetConfigFilePath(*configFile)
	}

	// Initialize logger
	if err := logger.InitLogger(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to init logger: %v\n", err)
		os.Exit(1)
	}
	logger.Info("starting dpu-exporter")

	// Load configuration
	if err := configmanager.LoadConfig(); err != nil {
		logger.Errorf("failed to load config: %v", err)
		os.Exit(1)
	}

	// Initialize device manager (card-type-specific)
	dmgr, err := device.AutoInit(*cardType)
	if err != nil {
		logger.Errorf("failed to init device manager: %v", err)
		os.Exit(1)
	}

	// Create the central DpuCollector
	n := dpucollector.NewDpuCollector(dmgr)

	// Set global context for collectors that need it
	metricscollector.SetGlobalContext(n)

	// Build initial collector chains
	initChains()

	// Register config reload hook (rebuild chains on config change)
	configmanager.RegisterReloadHook(func() {
		rebuildChains()
	})

	// Setup graceful shutdown context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup

	// Start config hot-reload watcher
	if err := configmanager.StartHotReload(ctx, &wg); err != nil {
		logger.Errorf("failed to start config hot-reload: %v", err)
	}

	// Initialize DPU device list (must run before StartCollect)
	dpucollector.InitDpuList(&wg, ctx, n)

	// Start all collection goroutines (dpu metrics, interface metrics)
	dpucollector.StartCollect(&wg, ctx, n)

	// Setup Prometheus platform
	promCollector := prompkg.NewPrometheusCollector()
	dpuMetricsChain, ifaceMetricsChain := dpucollector.GetChainsSnapshot()
	for _, c := range dpuMetricsChain {
		promCollector.Register(c)
	}
	for _, c := range ifaceMetricsChain {
		promCollector.Register(c)
	}

	// Start HTTP server in a goroutine
	serverErr := make(chan error, 1)
	go func() {
		serverErr <- prompkg.StartPrometheus(strconv.Itoa(httpPort), promCollector)
	}()

	logger.Infof("dpu-exporter started, serving metrics on port %d", httpPort)

	// Wait for SIGTERM/SIGINT or server error
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-sigChan:
		logger.Info("received stop signal, shutting down...")
	case err := <-serverErr:
		logger.Errorf("prometheus server stopped: %v", err)
		os.Exit(1)
	}

	// Cancel context to stop all goroutines
	cancel()

	// Wait for goroutines to finish with timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		logger.Info("all goroutines stopped gracefully")
	case <-time.After(shutdownTimeout):
		logger.Warnf("shutdown timeout reached after %v, forcing exit", shutdownTimeout)
	}

	logger.Info("dpu-exporter stopped")
}

// initChains builds the initial collector chains based on the current config.
func initChains() {
	cfg := configmanager.GetConfig()
	cfg.IntervalConfig.Apply()

	dpuMetricsChain := []metricscollector.MetricsCollector{&metricscollector.Hinicadm5Collector{}}
	ifaceMetricsChain := []metricscollector.MetricsCollector{&metricscollector.SysfsCollector{}}

	dpucollector.SetChains(dpuMetricsChain, ifaceMetricsChain)
	logger.Info("collector chains initialized")
}

// rebuildChains rebuilds the collector chains on config hot-reload
func rebuildChains() {
	dpuMetricsChain := []metricscollector.MetricsCollector{&metricscollector.Hinicadm5Collector{}}
	ifaceMetricsChain := []metricscollector.MetricsCollector{&metricscollector.SysfsCollector{}}

	dpucollector.SetChains(dpuMetricsChain, ifaceMetricsChain)
	logger.Info("collector chains rebuilt after config reload")
}
