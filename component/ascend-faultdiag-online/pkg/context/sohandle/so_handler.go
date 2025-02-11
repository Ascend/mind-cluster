/*
Copyright(C)2025. Huawei Technologies Co.,Ltd. All rights reserved.

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

/*
Package sohandle 包含了加载和使用动态链接库（.so）的功能。
*/
package sohandle

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"unsafe"
)

// 定义SO文件常量
const (
	GetType = "getType"
	Execute = "execute"
)

// SoHandler 结构体，用于管理动态链接库的句柄、类型以及主执行函数。
type SoHandler struct {
	SoHandle    syscall.Handle                           // .so 文件句柄
	SoType      string                                   // .so 文件类型
	ExecuteFunc func(input, output *string) (int, error) // .so 文件中的主执行函数
}

// NewSoHandler 创建一个新的 SoHandler
func NewSoHandler(soPath string) (*SoHandler, error) {
	// 加载 .so 文件
	handle, err := syscall.LoadLibrary(soPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load .so file: %v", err)
	}

	// 获取 .so 文件类型
	soType, err := getSoType(handle, soPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get .so file type: %v", err)
	}

	// 获取主执行函数
	executeFunc, err := getExecuteFunc(handle, soPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get execute function: %v", err)
	}

	return &SoHandler{
		SoHandle:    handle,
		SoType:      soType,
		ExecuteFunc: executeFunc,
	}, nil
}

// getSoType 获取 .so 文件类型
func getSoType(handle syscall.Handle, soPath string) (string, error) {
	fn, err := syscall.GetProcAddress(handle, GetType)
	if err != nil {
		return "", err
	}
	var typeName string
	_, _, ret := syscall.SyscallN(fn, uintptr(unsafe.Pointer(&typeName)))
	if ret != 0 {
		return "", errors.New(fmt.Sprintf("Call [%s] func [%s] failed, return code [%d]", soPath, GetType, ret))
	}
	return fmt.Sprintf("%s", typeName), nil
}

// getExecuteFunc 获取主执行函数
func getExecuteFunc(handle syscall.Handle, soPath string) (func(input, output *string) (int, error), error) {
	fn, err := syscall.GetProcAddress(handle, Execute)
	if err != nil {
		return nil, err
	}
	return func(input, output *string) (int, error) {
		inputPtr := uintptr(unsafe.Pointer(input))
		outputPtr := uintptr(unsafe.Pointer(output))
		_, _, ret := syscall.SyscallN(fn, 1, inputPtr, outputPtr, 0, 0)
		if ret != 0 {
			return -1, errors.New(fmt.Sprintf("Call [%s] func [%s] failed, return code [%d]", soPath, Execute, ret))
		}
		return 0, nil
	}, nil
}

// Close 释放 .so 文件句柄
func (h *SoHandler) Close() error {
	if h.SoHandle != 0 {
		return syscall.FreeLibrary(h.SoHandle)
	}
	return nil
}

// 筛选 .so 文件的函数
func filterSOFiles(soDir string) ([]string, error) {
	var soFiles []string
	// 使用 filepath.Walk 递归遍历目录
	err := filepath.Walk(soDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// 检查是否为普通文件且扩展名是 .so
		if !info.IsDir() && filepath.Ext(info.Name()) == ".so" {
			soFiles = append(soFiles, path)
		}
		return nil
	})
	return soFiles, err
}

// GenerateSoHandlerMap 生成 .so 文件句柄映射表
func GenerateSoHandlerMap(soDir string) (map[string]*SoHandler, error) {
	soFiles, err := filterSOFiles(soDir)
	if err != nil {
		return nil, err
	}
	soHandlerMap := make(map[string]*SoHandler)
	for _, soFile := range soFiles {
		soHandler, err := NewSoHandler(soFile)
		if err != nil {
			return nil, err
		}
		soHandlerMap[soHandler.SoType] = soHandler
	}
	return soHandlerMap, err
}
