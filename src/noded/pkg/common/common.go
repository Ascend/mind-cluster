/* Copyright(C) 2024. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package common for common function
package common

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"regexp"
	"strings"

	"huawei.com/npu-exporter/v6/common-utils/hwlog"
)

var (
	nodeDRegexp = map[string]*regexp.Regexp{
		RegexNodeNameKey:  regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$`),
		RegexFaultCodeKey: regexp.MustCompile(`^[a-zA-Z0-9]{8}$`),
	}
)

var (
	deviceType = map[byte]string{
		0x00: "CPU", 0x01: "Memory", 0x02: "Disk", 0x03: "PSU", 0x04: "Fan", 0x05: "Disk BP",
		0x06: "RAID Card", 0x07: "UNKNOWN", 0x08: "PCIe Card", 0x09: "AMC", 0x0A: "HBA", 0x0B: "Mezz Card",
		0x0C: "UNKNOWN", 0x0D: "NIC", 0x0E: "Memory Board", 0x0F: "PCIe Riser", 0x10: "Mainboard",
		0x11: "LCD", 0x12: "Chassis", 0x13: "NCM", 0x14: "Switch Module", 0x15: "Storage Board",
		0x16: "Chassis BP", 0x17: "HMM/CMC", 0x18: "Fan BP", 0x19: "PSU BP", 0x1A: "BMC", 0x1B: "MMC/MM",
		0x1C: "Twin Node Backplane", 0x1D: "Base Plane", 0x1E: "Fabric Plane", 0x1F: "Switch Mezz", 0x20: "LED",
		0x21: "SD Card", 0x22: "Security Module", 0x23: "I/O Board", 0x24: "CPU Board", 0x25: "RMC",
		0x26: "PCIe Adapter", 0x27: "PCH", 0x28: "Cable", 0x29: "Port", 0x2A: "LSW", 0x2B: "PHY", 0x2C: "System",
		0x2D: "M.2 Transfer Card", 0x2E: "LED Board", 0x2F: "LPM", 0x30: "PIC Card", 0x31: "Button", 0x32: "Expander",
		0x33: "CPI", 0x34: "ACM", 0x35: "CIM", 0x36: "PFM", 0x37: "KPAR", 0x38: "JC", 0x39: "SCM",
		0x3A: "Minisas HD channel", 0x3B: "SATA DOM channel", 0x3C: "GE channel", 0x3D: "XGE channel",
		0x3E: "PCIe Switch", 0x3F: "Interface Device", 0x40: "xPU Board", 0x41: "Disk BaseBoard",
		0x42: "VGA Intf Card", 0x43: "Pass-Through Card", 0x44: "Logical Driver", 0x45: "PCIe Retimer",
		0x46: "PCIe Repeater", 0x47: "SAS", 0x48: "Memory Channel", 0x49: "BMA", 0x4A: "LOM",
		0x4B: "Signal Adapter Board", 0x4C: "Horizontal Connection Board", 0x4D: "Node", 0x4E: "Asset Locate Board",
		0x4F: "Unit", 0x50: "RMM", 0x51: "Rack", 0x52: "BBU", 0x53: "OCP Card", 0x54: "Leakage Detection Card",
		0x55: "MESH Card", 0x56: "NPU", 0x57: "CIC Card", 0x58: "Expansion Module", 0x59: "Fan Module",
		0x5A: "AR Card", 0x5B: "Converge Board", 0x5C: "SoC Board", 0x5D: "ExpBoard", 0xC0: "BCU", 0xC1: "EXU",
		0xC2: "SEU", 0xC3: "IEU", 0xC4: "CLU",
	}
)

// GetDeviceType get device type str
func GetDeviceType(subject byte) string {
	if _, ok := deviceType[subject]; !ok {
		return UnknownDevice
	}
	return deviceType[subject]
}

// GetPattern return pattern map
func GetPattern() map[string]*regexp.Regexp {
	return nodeDRegexp
}

// CopyStringSlice copy a string slice
func CopyStringSlice(s []string) []string {
	result := make([]string, len(s))
	for i, str := range s {
		result[i] = strings.Clone(str)
	}
	return result
}

// MakeDataHash make data hash
func MakeDataHash(data interface{}) string {
	dataBuffer, err := json.Marshal(data)
	if err != nil {
		return ""
	}
	encode := sha256.New()
	if _, err := encode.Write(dataBuffer); err != nil {
		hwlog.RunLog.Errorf("hash data failed, err is %v", err)
		return ""
	}
	sum := encode.Sum(nil)
	return hex.EncodeToString(sum)
}

// DeepCopyFaultConfig deep copy fault config
func DeepCopyFaultConfig(oldConfig, newConfig *FaultConfig) {
	DeepCopyFaultTypeCode(oldConfig.FaultTypeCode, newConfig.FaultTypeCode)
}

// DeepCopyFaultTypeCode deep copy fault type code
func DeepCopyFaultTypeCode(oldFaultTypeCode, newFaultTypeCode *FaultTypeCode) {
	oldFaultTypeCode.NotHandleFaultCodes = CopyStringSlice(newFaultTypeCode.NotHandleFaultCodes)
	oldFaultTypeCode.PreSeparateFaultCodes = CopyStringSlice(newFaultTypeCode.PreSeparateFaultCodes)
	oldFaultTypeCode.SeparateFaultCodes = CopyStringSlice(newFaultTypeCode.SeparateFaultCodes)
}

// RemoveDuplicateString string deduplication
func RemoveDuplicateString(strSlice []string) []string {
	strMap := make(map[string]struct{})
	result := make([]string, 0)
	for _, str := range strSlice {
		if _, ok := strMap[str]; !ok {
			strMap[str] = struct{}{}
			result = append(result, str)
		}
	}
	return result
}

// RevertByteSlice revert the byte slice
func RevertByteSlice(bytes []byte) []byte {
	for i, j := 0, len(bytes)-1; i < j; i, j = i+1, j-1 {
		bytes[i], bytes[j] = bytes[j], bytes[i]
	}
	return bytes
}

// ConvertIntToTwoByteSlice convert int to two byte slice
func ConvertIntToTwoByteSlice(num int64) []byte {
	if num <= 0 || num > MaxSixTeenBitIntValue {
		return []byte{0x00, 0x00}
	}
	var result []byte
	for num > 0 {
		temp := num % HexByteBase
		result = append(result, byte(temp))
		num = num / HexByteBase
	}
	for len(result) < TwoByteSliceLength {
		result = append(result, byte(zeroByte))
	}
	return result
}

// NewSignalWatcher create a new signal watcher
func NewSignalWatcher(signals ...os.Signal) chan os.Signal {
	signalChan := make(chan os.Signal, 1)
	for _, sign := range signals {
		signal.Notify(signalChan, sign)
	}
	return signalChan
}

// FilterDuplicateFaultCodes filter duplicates fault codes in same level
func FilterDuplicateFaultCodes(faultTypeCode *FaultTypeCode) {
	faultTypeCode.NotHandleFaultCodes = RemoveDuplicateString(faultTypeCode.NotHandleFaultCodes)
	faultTypeCode.PreSeparateFaultCodes = RemoveDuplicateString(faultTypeCode.PreSeparateFaultCodes)
	faultTypeCode.SeparateFaultCodes = RemoveDuplicateString(faultTypeCode.SeparateFaultCodes)
}

// ToUpperFaultCodesStr convert fault type code str to upper
func ToUpperFaultCodesStr(faultTypeCode *FaultTypeCode) {
	StringSliceToUpper(faultTypeCode.NotHandleFaultCodes)
	StringSliceToUpper(faultTypeCode.PreSeparateFaultCodes)
	StringSliceToUpper(faultTypeCode.SeparateFaultCodes)
}

// StringSliceToUpper convert str to upper in a slice
func StringSliceToUpper(slice []string) {
	for i, str := range slice {
		slice[i] = strings.ToUpper(str)
	}
	return
}

// CheckFaultCodes check whether fault code is illegal
func CheckFaultCodes(codes []string) error {
	regex := GetPattern()[RegexFaultCodeKey]
	for _, code := range codes {
		if !regex.MatchString(code) {
			return fmt.Errorf("fault code %s contains illegal character", code)
		}
	}
	return nil
}
