/* Copyright(C) 2022. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package process arses the environment variables to obtain the files and directories to be mounted
// and transfers the files and directories to the CLI for mounting before the container is started.
package process

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/opencontainers/runtime-spec/specs-go"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	"ascend-common/common-utils/limiter"
	"ascend-docker-runtime/mindxcheckutils"
)

const (
	runLogPath             = api.HookRunLogPath
	ascendRuntimeOptions   = api.AscendRuntimeOptionsEnv
	ascendRuntimeMounts    = api.AscendRuntimeMountsEnv
	ascendVisibleDevices   = api.AscendVisibleDevicesEnv
	ascendAllowLink        = api.AscendAllowLinkEnv
	ascendDockerCli        = "ascend-docker-cli"
	defaultAscendDockerCli = "/usr/local/bin/ascend-docker-cli"
	configDir              = api.RunTimeDConfigPath
	baseConfig             = "base"
	configFileSuffix       = "list"

	kvPairSize = 2
	// MaxCommandLength is the max length of command.
	MaxCommandLength = 65535
)

var (
	limitReader                = io.LimitReader(os.Stdin, limiter.DefaultDataLimit)
	doExec                     = syscall.Exec
	ascendDockerCliName        = ascendDockerCli
	defaultAscendDockerCliName = defaultAscendDockerCli
)

var validRuntimeOptions = [...]string{
	"NODRV",
	"VIRTUAL",
}

type containerConfig struct {
	Pid    int
	Rootfs string
	Env    []string
}

// InitLogModule initializes some logging configuration.
func InitLogModule(ctx context.Context) error {
	const backups = 2
	const logMaxAge = 365
	const fileMaxSize = 2
	runLogConfig := hwlog.LogConfig{
		LogFileName: runLogPath,
		LogLevel:    0,
		MaxBackups:  backups,
		MaxAge:      logMaxAge,
		OnlyToFile:  true,
		FileMaxSize: fileMaxSize,
	}
	if err := hwlog.InitRunLogger(&runLogConfig, ctx); err != nil {
		fmt.Printf("log init failed, error is %v", err)
		return err
	}
	return nil
}

func parseMounts(mounts string) []string {
	if mounts == "" {
		return []string{baseConfig}
	}
	const maxMountLength = 128
	if len(mounts) > maxMountLength {
		return []string{baseConfig}
	}

	mountConfigs := make([]string, 0)
	for _, m := range strings.Split(mounts, ",") {
		m = strings.TrimSpace(m)
		m = strings.ToLower(m)
		mountConfigs = append(mountConfigs, m)
	}

	return mountConfigs
}

func isRuntimeOptionValid(option string) bool {
	for _, validOption := range validRuntimeOptions {
		if option == validOption {
			return true
		}
	}

	return false
}

func parseRuntimeOptions(runtimeOptions string) ([]string, error) {
	parsedOptions := make([]string, 0)

	if runtimeOptions == "" {
		return parsedOptions, nil
	}
	const maxLength = 128
	if len(runtimeOptions) > maxLength {
		hwlog.RunLog.Errorf("length of %v value is invalid, its length: %v",
			api.AscendRuntimeOptionsEnv, len(runtimeOptions))
		return nil, fmt.Errorf("invalid runtime option, the length exceeds 128 characters")
	}

	for _, option := range strings.Split(runtimeOptions, ",") {
		option = strings.TrimSpace(option)
		if !isRuntimeOptionValid(option) {
			hwlog.RunLog.Errorf("value of %v is not in valid option list, value: %v",
				api.AscendRuntimeOptionsEnv, option)
			return nil, fmt.Errorf("invalid runtime option of invalid input value")
		}

		parsedOptions = append(parsedOptions, option)
	}

	return parsedOptions, nil
}

func parseSoftLinkMode(allowLink string) (string, error) {
	if allowLink == "True" {
		return "True", nil
	}
	if allowLink == "" || allowLink == "False" {
		return "False", nil
	}

	return "", fmt.Errorf("invalid soft link option")
}

func parseOciSpecFile(file string) (*specs.Spec, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, fmt.Errorf("failed to open the OCI config file: %s, err: %v", file, err)
	}
	defer f.Close()

	spec := new(specs.Spec)
	if err := json.NewDecoder(f).Decode(spec); err != nil {
		return nil, fmt.Errorf("failed to parse OCI config file: %s, caused by: %v", file, err)
	}

	if spec.Process == nil {
		return nil, fmt.Errorf("invalid OCI spec for empty process")
	}

	if spec.Root == nil {
		return nil, fmt.Errorf("invalid OCI spec for empty root")
	}

	return spec, nil
}

var getContainerConfig = func() (*containerConfig, error) {
	state := new(specs.State)
	decoder := json.NewDecoder(limitReader)

	if err := decoder.Decode(state); err != nil {
		return nil, fmt.Errorf("failed to parse the container's state")
	}

	configPath := path.Join(state.Bundle, "config.json")
	if _, err := mindxcheckutils.RealFileChecker(configPath, true, true, mindxcheckutils.DefaultSize); err != nil {
		return nil, err
	}

	ociSpec, err := parseOciSpecFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse OCI spec: %v", err)
	}
	if len(ociSpec.Process.Env) > MaxCommandLength {
		return nil, fmt.Errorf("too many items in spec file")
	}
	// when use ctr->containerd. the rootfs in config.json is a relative path
	rfs := ociSpec.Root.Path
	if !filepath.IsAbs(rfs) {
		rfs = path.Join(state.Bundle, ociSpec.Root.Path)
	}

	ret := &containerConfig{
		Pid:    state.Pid,
		Rootfs: rfs,
		Env:    ociSpec.Process.Env,
	}

	return ret, nil
}

func getValueByKey(data []string, name string) string {
	splitNumber := 2
	for _, s := range data {
		p := strings.SplitN(s, "=", splitNumber)
		if len(p) != kvPairSize {
			hwlog.RunLog.Errorf("env is not key-value mode, env: %v", s)
			log.Panicln("environment error")
		}

		if p[0] == name && len(p) == kvPairSize {
			return p[1]
		}
	}

	return ""
}

func readMountConfig(dir string, name string) ([]string, []string, error) {
	configFileName := fmt.Sprintf("%s.%s", name, configFileSuffix)
	baseConfigFilePath, err := filepath.Abs(filepath.Join(dir, configFileName))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to assemble base config file path: %v", err)
	}

	fileInfo, err := os.Stat(baseConfigFilePath)
	if _, err := mindxcheckutils.RealFileChecker(baseConfigFilePath, true, false,
		mindxcheckutils.DefaultSize); err != nil {
		return nil, nil, err
	}
	if err != nil {
		return nil, nil, fmt.Errorf("cannot stat base configuration file %s : %v", baseConfigFilePath, err)
	}

	if !fileInfo.Mode().IsRegular() {
		return nil, nil, fmt.Errorf("base configuration file damaged because is not a regular file")
	}

	f, err := os.Open(baseConfigFilePath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open base configuration file %s: %v", baseConfigFilePath, err)
	}
	defer f.Close()

	fileMountList, dirMountList := make([]string, 0), make([]string, 0)
	const maxEntryNumber = 128
	entryCount := 0
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		mountPath := scanner.Text()
		entryCount = entryCount + 1
		if entryCount > maxEntryNumber {
			return nil, nil, fmt.Errorf("mount list too long")
		}
		absMountPath, err := filepath.Abs(mountPath)
		if err != nil {
			continue // skipping files/dirs with any problems
		}
		mountPath = absMountPath

		stat, err := os.Stat(mountPath)
		if err != nil {
			continue // skipping files/dirs with any problems
		}

		if stat.Mode().IsRegular() {
			fileMountList = append(fileMountList, mountPath)
		} else if stat.Mode().IsDir() {
			dirMountList = append(dirMountList, mountPath)
		}
	}

	return fileMountList, dirMountList, nil
}

func readConfigsOfDir(dir string, configs []string) ([]string, []string, error) {
	fileInfo, err := os.Stat(dir)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot stat configuration directory %s : %v", dir, err)
	}

	if !fileInfo.Mode().IsDir() {
		return nil, nil, fmt.Errorf("%s should be a dir for docker runtime, but now it is not", dir)
	}

	fileMountList := make([]string, 0)
	dirMountList := make([]string, 0)

	for _, config := range configs {
		fileList, dirList, err := readMountConfig(dir, config)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to process config %s: %v", config, err)
		}

		fileMountList = append(fileMountList, fileList...)
		dirMountList = append(dirMountList, dirList...)
	}

	return fileMountList, dirMountList, nil
}

func getArgs(cliPath string, containerConfig *containerConfig, fileMountList []string,
	dirMountList []string, allowLink string) []string {
	args := append([]string{cliPath},
		"--allow-link", allowLink, "--pid", fmt.Sprintf("%d", containerConfig.Pid),
		"--rootfs", containerConfig.Rootfs)
	for _, filePath := range fileMountList {
		args = append(args, "--mount-file", filePath)
	}
	for _, dirPath := range dirMountList {
		args = append(args, "--mount-dir", dirPath)
	}
	return args
}

// DoPrestartHook parses the environment variables in the container to obtain the files and directories to be mounted.
func DoPrestartHook() error {
	containerConfig, err := getContainerConfig()
	if err != nil {
		return fmt.Errorf("failed to get container config: %#v", err)
	}

	if visibleDevices := getValueByKey(containerConfig.Env, ascendVisibleDevices); visibleDevices == "" {
		return nil
	}

	mountConfigs := parseMounts(getValueByKey(containerConfig.Env, ascendRuntimeMounts))

	fileMountList, dirMountList, err := readConfigsOfDir(configDir, mountConfigs)
	if err != nil {
		return fmt.Errorf("failed to read configuration from config directory: %#v", err)
	}

	parsedOptions, err := parseRuntimeOptions(getValueByKey(containerConfig.Env, ascendRuntimeOptions))
	if err != nil {
		return fmt.Errorf("failed to parse runtime options: %#v", err)
	}

	allowLink, err := parseSoftLinkMode(getValueByKey(containerConfig.Env, ascendAllowLink))
	if err != nil {
		return fmt.Errorf("failed to parse soft link mode: %#v", err)
	}

	currentExecPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("cannot get the path of docker-hook: %#v", err)
	}

	cliPath := path.Join(path.Dir(currentExecPath), ascendDockerCliName)
	if _, err = os.Stat(cliPath); err != nil {
		return fmt.Errorf("cannot find docker-cli executable file at %s: %#v", cliPath, err)
	}
	if _, err := mindxcheckutils.RealFileChecker(cliPath, true, false, mindxcheckutils.DefaultSize); err != nil {
		return err
	}
	args := getArgs(cliPath, containerConfig, fileMountList, dirMountList, allowLink)
	if len(parsedOptions) > 0 {
		args = append(args, "--options", strings.Join(parsedOptions, ","))
	}
	hwlog.RunLog.Info("docker hook success, will start cli")
	if err := mindxcheckutils.ChangeRuntimeLogMode("hook-run-"); err != nil {
		return err
	}
	if err := doExec(cliPath, args, os.Environ()); err != nil {
		return fmt.Errorf("failed to exec docker-cli %v: %v", args, err)
	}
	return nil
}
