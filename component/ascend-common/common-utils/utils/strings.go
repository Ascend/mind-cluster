/* Copyright(C) 2023. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package utils provides the util func
package utils

import (
	"crypto/sha256"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"ascend-common/api"
)

const (
	maskLen = 2
)

// ReplacePrefix replace string with prefix
func ReplacePrefix(source, prefix string) string {
	if prefix == "" {
		prefix = "****"
	}
	if len(source) <= maskLen {
		return prefix
	}
	end := string([]rune(source)[maskLen:len(source)])
	return prefix + end
}

// MaskPrefix mask string prefix with ****
func MaskPrefix(source string) string {
	return ReplacePrefix(source, "")
}

// GetSha256Code return the sha256 hash bytes
func GetSha256Code(data []byte) []byte {
	hash256 := sha256.New()
	if _, err := hash256.Write(data); err != nil {
		fmt.Println(err)
		return nil
	}
	return hash256.Sum(nil)
}

// ReverseString reverse string
func ReverseString(s string) string {
	runes := []rune(s)
	for start, end := 0, len(runes)-1; start < end; start, end = start+1, end-1 {
		runes[start], runes[end] = runes[end], runes[start]
	}
	return string(runes)
}

// IsDigitString return string is all digit
func IsDigitString(s string) bool {
	if len(s) == 0 {
		return false
	}
	for _, c := range s {
		if !unicode.IsDigit(c) {
			return false
		}
	}
	return true
}

// GetLastNumberFromString get last number for string
// eg："mu-test-0-decode-0-0" -> "0"
func GetLastNumberFromString(name string) (string, error) {
	parts := strings.Split(name, "-")
	if len(parts) == 0 {
		return "", fmt.Errorf("wrong format of string")
	}

	lastPart := parts[len(parts)-1]

	_, err := strconv.Atoi(lastPart)
	if err != nil {
		return "", fmt.Errorf("no number found at the end of string: %s, %v", name, err)
	}

	return lastPart, nil
}

// devTypeMaskMap maps the original devType to the masked devType
var devTypeMaskMap = map[string]string{
	api.Ascend910A5: api.VersionNPU,
}

// MaskDevType masks the devType based on the devTypeMaskMap
func MaskDevType(devType string) string {
	if masked, ok := devTypeMaskMap[devType]; ok {
		return masked
	}
	return devType
}

// npuResourceNameSet fixed NPU resource names, including whole-card and
// dynamic segmentation (npu-core) resources, excluding other resources
// with the same prefix (such as DPU and UB NIC resources)
var npuResourceNameSet = map[string]struct{}{
	api.HuaweiNPU:        {},
	api.HuaweiNPUCore:    {},
	api.HuaweiAscend910:  {},
	api.HuaweiAscend310:  {},
	api.HuaweiAscend310P: {},
}

// staticCutResourceRegexp matches static segmentation (vNPU) resource names
// with valid hardware tiers, like Ascend910-2c, Ascend310P-1c
var staticCutResourceRegexp = regexp.MustCompile(`^(Ascend910-([2-6]|8|10|12|16)|Ascend310P-(1|2|4))c$`)

// IsNPUResourceName check whether the resource name is an NPU resource,
// including fixed names and static segmentation names like huawei.com/Ascend910-2c
func IsNPUResourceName(name string) bool {
	if _, ok := npuResourceNameSet[name]; ok {
		return true
	}
	return IsStaticCutResourceName(name)
}

// IsStaticCutResourceName check whether the resource name is a static
// segmentation (vNPU) resource name, like huawei.com/Ascend910-2c
func IsStaticCutResourceName(name string) bool {
	if !strings.HasPrefix(name, api.ResourceNamePrefix) {
		return false
	}
	return staticCutResourceRegexp.MatchString(strings.TrimPrefix(name, api.ResourceNamePrefix))
}
