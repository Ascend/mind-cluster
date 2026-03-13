/* Copyright(C) 2021-2023. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package hccn this for npu hccn info
package hccn

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"ascend-common/common-utils/hwlog"
	"ascend-common/common-utils/limiter"
	"ascend-common/common-utils/utils"
	"ascend-common/devmanager/common"
)

const (
	space   = " "
	newLine = "\n"
	colon   = ":"

	// LinkUp npu interface up
	LinkUp string = "UP"
	// LinkDown npu interface down
	LinkDown string = "DOWN"

	opticalPartLen  = 2
	netSpeedPartLen = 2
	firstIndex      = 1
	secondIndex     = 2
	fourthIndex     = 4
	linkStatusPart  = 3
	base64          = 64

	cardHealthy = 0

	normalCode   = 1
	abnormalCode = 0

	naValue    = "NA"
	notSupport = "not supported"
	unknownStr = "Unknown!"

	limitSize = 1024 * 1024
)

func getInfoFromHccnTool(args ...string) (string, error) {
	const hccnTool = "/usr/local/Ascend/driver/tools/hccn_tool"
	if _, err := utils.CheckPath(hccnTool); err != nil {
		return "", err
	}
	cmd := exec.Command(hccnTool, args...)
	cmd.Env = []string{
		"PATH=" + os.Getenv("PATH"),
		utils.LdLibPath + "=" + os.Getenv(utils.LdLibPath),
	}
	limitStdout := limiter.NewLimitedWriter(limitSize)
	cmd.Stdout = limitStdout
	cmd.Stderr = limiter.NewLimitedWriter(limitSize)
	err := cmd.Run()
	if err != nil {
		return "", err
	}

	return string(limitStdout.GetBufferBytes()), nil
}

// GetNPULinkStatus exec "hccn_tool -i * -link -g" to get link status
func GetNPULinkStatus(phyID int32) (string, error) {
	args := []string{"-i", strconv.Itoa(int(phyID)), "-link", "-g"}
	// command example: hccn_tool -i 0 -link -g
	// success result example is: link status: DOWN
	outStr, err := getInfoFromHccnTool(args...)
	hwlog.RunLog.Debugf("hccn_tool command exec result: %v", outStr)
	if err != nil {
		return common.Abnormal, buildHccnErr(phyID, "link status", err)
	}
	replacedStr := strings.ReplaceAll(outStr, newLine, "")
	outArr := strings.Split(replacedStr, space)
	if len(outArr) != linkStatusPart {
		return common.Abnormal, buildHccnErr(phyID, "link status",
			fmt.Errorf("length of output %v is not equal to %v", outArr, linkStatusPart))
	}

	status := outArr[secondIndex]
	hwlog.RunLog.Debugf("hccn_tool get npu link status: %s", status)
	return status, nil
}

// GetNPULinkSpeed exec "hccn_tool -i * -speed -g" to get link speed
func GetNPULinkSpeed(phyID int32) (int, error) {
	args := []string{"-i", strconv.Itoa(int(phyID)), "-speed", "-g"}
	// command example: hccn_tool -i 0 -speed -g
	// success result example is: Speed: 100000 Mb/s
	outStr, err := getInfoFromHccnTool(args...)
	if err != nil {
		return common.RetError, buildHccnErr(phyID, "link speed", err)
	}
	return getSpeedFromOutStr(outStr, phyID)
}

func getSpeedFromOutStr(outStr string, phyID int32) (int, error) {
	if strings.Contains(outStr, unknownStr) {
		return common.RetError, buildHccnErr(phyID, "link speed", fmt.Errorf("npu link speed is unknown"))
	}
	replacedStr := strings.ReplaceAll(outStr, newLine, "")
	outArr := strings.Split(replacedStr, space)
	if len(outArr) != linkStatusPart {
		return common.RetError, buildHccnErr(phyID, "link speed", fmt.Errorf("length of output %v is not equal to %v",
			outArr, linkStatusPart))
	}
	const midIndex = 1
	speed, err := strconv.Atoi(outArr[midIndex])
	if err != nil {
		return common.RetError, buildHccnErr(phyID, "link speed", fmt.Errorf("covert speed from string failed: %s", err))
	}

	return speed, nil
}

// GetNPULinkUpNum exec "hccn_tool -i * -link_stat -g" to get link up count
func GetNPULinkUpNum(phyID int32) (int, error) {
	args := []string{"-i", strconv.Itoa(int(phyID)), "-link_stat", "-g"}
	// command example: hccn_tool -i 0 -link_stat -g
	// success result include: [device x]link up count : y
	outStr, err := getInfoFromHccnTool(args...)
	if err != nil {
		return common.RetError, buildHccnErr(phyID, "link stat", err)
	}

	const (
		linkUpArrLen = 6
		linkUpStr    = "link up count"
	)
	linkUPCount := 0
	lines := strings.Split(outStr, newLine)
	for _, line := range lines {
		if line == "" || !strings.Contains(line, linkUpStr) {
			continue
		}

		linkUpArr := strings.Fields(line)
		if len(linkUpArr) != linkUpArrLen {
			return common.RetError, buildHccnErr(phyID, "link up num", fmt.Errorf("length of output %v is not "+
				"equal to %v", linkUpArr, linkUpArrLen))
		}
		if linkUPCount, err = strconv.Atoi(linkUpArr[linkUpArrLen-1]); err != nil {
			return common.RetError, buildHccnErr(phyID, "link up num",
				fmt.Errorf("covert link up num from string failed: %s", err))
		}
		return linkUPCount, nil
	}

	return common.RetError, buildHccnErr(phyID, "link up num", fmt.Errorf("did not find link up count"))
}

// GetNPUStatInfo exec "hccn_tool -i * -stat -g" to get stat info
func GetNPUStatInfo(phyID int32) (map[string]int, error) {
	args := []string{"-i", strconv.Itoa(int(phyID)), "-stat", "-g"}
	// command example: hccn_tool -i 0 -stat -g
	// success result include: [device x]link up count : y
	outStr, err := getInfoFromHccnTool(args...)
	if err != nil {
		return nil, buildHccnErr(phyID, "stat", err)
	}
	lines := strings.Split(outStr, newLine)
	statInfoMap := make(map[string]int)
	const statPartLen = 2
	for _, line := range lines {
		statParts := strings.Split(line, colon)
		if len(statParts) != statPartLen || statParts[1] == "" {
			continue
		}
		statNum, err := strconv.Atoi(statParts[1])
		if err != nil {
			hwlog.RunLog.Errorf("covert stat num of [%s] from string failed: %s", statParts[1], err)
			continue
		}
		statInfoMap[statParts[0]] = statNum
	}

	return statInfoMap, nil
}

// GetNPUOpticalInfo exec "hccn_tool -i * -optical -g" to get optical info
func GetNPUOpticalInfo(phyID int32) (map[string]string, error) {
	args := []string{"-i", strconv.Itoa(int(phyID)), "-optical", "-g"}
	// command example: hccn_tool -i 0 -optical -g
	// success result include: [device x]link up count : y
	outStr, err := getInfoFromHccnTool(args...)
	if err != nil {
		return nil, buildHccnErr(phyID, "optical", err)
	}
	lines := strings.Split(outStr, newLine)
	opticalInfoMap := make(map[string]string)
	for _, line := range lines {
		opticalParts := strings.Split(line, colon)
		if len(opticalParts) != opticalPartLen {
			continue
		}
		opticalKey := strings.ReplaceAll(strings.TrimSpace(opticalParts[0]), space, "_")
		opticalValue := strings.TrimSpace(opticalParts[1])
		opticalInfoMap[opticalKey] = opticalValue
	}

	return opticalInfoMap, nil
}

// GetNPUInterfaceTraffic exec "hccn_tool -i * -bandwidth -g" to get bandwidth info
func GetNPUInterfaceTraffic(phyID int32) (float64, float64, error) {
	const (
		noTraffic      = common.RetError
		trafficPartLen = 4
		txStr          = "TX:"
		rxStr          = "RX:"
	)

	args := []string{"-i", strconv.Itoa(int(phyID)), "-bandwidth", "-g"}
	// command example: hccn_tool -i 0 -bandwidth -g
	// success result has two lines:
	// Bandwidth TX: 0.00 MB/sec
	// Bandwidth RX: 0.00 MB/sec
	outStr, err := getInfoFromHccnTool(args...)
	hwlog.RunLog.Debugf("hccn_tool command exec result: %v", outStr)
	if err != nil {
		return noTraffic, noTraffic, buildHccnErr(phyID, "interface traffic", err)
	}

	var (
		tx = float64(noTraffic)
		rx = float64(noTraffic)
	)

	lines := strings.Split(outStr, newLine)
	for _, line := range lines {
		if line == "" {
			continue
		}

		trafficArr := strings.Fields(line)
		hwlog.RunLog.Debugf("npu bandwidth split as: %v", trafficArr)
		if len(trafficArr) != trafficPartLen {
			continue
		}
		if strings.Contains(line, txStr) {
			tmpTx, err := strconv.ParseFloat(trafficArr[secondIndex], base64)
			if err != nil {
				hwlog.RunLog.Errorf("get float data from Bandwidth TX err: %s", err)
				continue
			}
			tx = tmpTx
		}
		if strings.Contains(line, rxStr) {
			tmpRx, err := strconv.ParseFloat(trafficArr[secondIndex], base64)
			if err != nil {
				hwlog.RunLog.Errorf("get float data from Bandwidth RX err: %s", err)
				continue
			}
			rx = tmpRx
		}
	}
	return tx, rx, nil
}

// GetFloatDataFromStr get float data from string with space
func GetFloatDataFromStr(str, dataType string) float64 {
	if str == "" || strings.Contains(str, naValue) || strings.Contains(str, notSupport) {
		return common.RetError
	}
	dataParts := strings.Split(str, space)
	if len(dataParts) != opticalPartLen {
		errMsg := fmt.Sprintf("convert %v optical data type failed, "+
			"the length of optical data %v is %v not equal to %d. ", dataType, dataParts, len(dataParts), opticalPartLen)
		hwlog.RunLog.Error(errMsg)
		return common.RetError
	}
	floatData, err := strconv.ParseFloat(dataParts[0], base64)
	if err != nil {
		hwlog.RunLog.Errorf("convert %v optical data type to a floating-point number failed, "+
			"get float data from string %v failed, err: %v", dataType, dataParts[0], err)
		return common.RetError
	}
	return floatData
}

// GetHealthCode return union healthy code
func GetHealthCode(healthCode uint32) int {
	if healthCode == common.UnRetError {
		return common.RetError
	}

	if healthCode == cardHealthy {
		return normalCode
	}
	return abnormalCode
}

// GetLinkStatusCode return union link status code
func GetLinkStatusCode(status string) int {
	if status == common.Abnormal {
		return common.RetError
	}

	if status == LinkUp {
		return normalCode
	}
	return abnormalCode
}

// GetNetworkHealthy return union network healthy code
func GetNetworkHealthy(netCode uint32) int {
	if netCode == common.UnRetError {
		return common.RetError
	}

	if netCode == common.NetworkInit || netCode == common.NetworkSuccess {
		return normalCode
	}
	return abnormalCode
}

func buildHccnErr(phyID int32, msg string, err error) error {
	return fmt.Errorf("phyID(%d),get npu %s info failed,error is :%v", phyID, msg, err)
}

// GetNPULinkStatusA5 exec "hccn_tool -g -link -i device_id -u udie_id -p port_id" to get link status
func GetNPULinkStatusA5(logicID, udieID, portID int32) (string, error) {
	args := []string{"-g", "-link", "-i", strconv.Itoa(int(logicID)), "-u", strconv.Itoa(int(udieID)), "-p", strconv.Itoa(int(portID))}
	// command example: hccn_tool -g -link -i 58 -u 0 -p 4
	// success result example is: link status: DOWN
	outStr, err := getInfoFromHccnTool(args...)
	hwlog.RunLog.Debugf("hccn_tool command: %v exec result: %v", args, outStr)
	if err != nil {
		return common.Abnormal, buildHccnErrA5("link status", err)
	}
	replacedStr := strings.ReplaceAll(outStr, newLine, "")
	outArr := strings.Split(replacedStr, space)
	if len(outArr) != linkStatusPart {
		return common.Abnormal, buildHccnErrA5("link status",
			fmt.Errorf("length of output %v is not equal to %v", outArr, linkStatusPart))
	}

	status := outArr[secondIndex]
	hwlog.RunLog.Debugf("hccn_tool get npu link status: %s", status)
	return status, nil
}

// GetNPUInterfaceTrafficA5 exec "hccn_tool -g -bandwidth -i device_id -u udie_id -p port_id -time [1-226]" to get bandwidth info
func GetNPUInterfaceTrafficA5(logicID, udieID, portID, durationTime int32) (float64, float64, error) {
	const (
		noTraffic      = common.RetError
		trafficPartLen = 4
		txStr          = "TX:"
		rxStr          = "RX:"
	)
	args := []string{"-g", "-bandwidth", "-i", strconv.Itoa(int(logicID)), "-u", strconv.Itoa(int(udieID)), "-p",
		strconv.Itoa(int(portID)), "-time", strconv.Itoa(int(durationTime))}
	// success result has two lines:
	// Bandwidth TX: 0.00 MB/sec
	// Bandwidth RX: 0.00 MB/sec
	outStr, err := getInfoFromHccnTool(args...)
	hwlog.RunLog.Debugf("hccn_tool command exec result: %v", outStr)
	if err != nil {
		return noTraffic, noTraffic, buildHccnErrA5("interface traffic", err)
	}
	var (
		tx = float64(noTraffic)
		rx = float64(noTraffic)
	)
	lines := strings.Split(outStr, newLine)
	for _, line := range lines {
		if line == "" {
			continue
		}
		trafficArr := strings.Fields(line)
		hwlog.RunLog.Debugf("npu bandwidth split as: %v", trafficArr)
		if len(trafficArr) != trafficPartLen {
			continue
		}
		if strings.Contains(line, txStr) {
			tmpTx, err := strconv.ParseFloat(trafficArr[secondIndex], base64)
			if err != nil {
				hwlog.RunLog.Errorf("get float data from Bandwidth TX err: %v", err)
				continue
			}
			tx = tmpTx
		}
		if strings.Contains(line, rxStr) {
			tmpRx, err := strconv.ParseFloat(trafficArr[secondIndex], base64)
			if err != nil {
				hwlog.RunLog.Errorf("get float data from Bandwidth RX err: %v", err)
				continue
			}
			rx = tmpRx
		}
	}
	return tx, rx, nil
}

// GetNPULinkSpeedA5 exec "hccn_tool -g -speed -i phy_id -u udie_id -p port_id" to get link speed
func GetNPULinkSpeedA5(logicID, udieID, portID int32) (int, error) {
	args := []string{"-g", "-speed", "-i", strconv.Itoa(int(logicID)), "-u", strconv.Itoa(int(udieID)),
		"-p", strconv.Itoa(int(portID))}
	// command example: hccn_tool -g -speed -i 56 -u 0 -p 4
	outStr, err := getInfoFromHccnTool(args...)
	if err != nil {
		return common.RetError, buildHccnErrA5("link speed", err)
	}
	return getSpeedFromStrA5(outStr)
}

func getSpeedFromStrA5(str string) (int, error) {
	if strings.Contains(str, naValue) {
		return common.RetError, buildHccnErrA5("link speed", fmt.Errorf("npu link speed is unknown, port is down"))
	}

	lines := strings.Split(strings.TrimSpace(str), "\n")
	var dataRows []string
	for _, line := range lines {
		if line != "" && !strings.Contains(line, "+") && strings.Contains(line, "|") {
			dataRows = append(dataRows, line)
		}
	}
	parts := strings.Split(dataRows[firstIndex], "|")

	speedInfo := strings.Split(strings.TrimSpace(parts[fourthIndex]), space)
	if len(speedInfo) != netSpeedPartLen {
		return common.RetError, buildHccnErrA5("port speed ",
			fmt.Errorf("length of output %v is not equal to %v", speedInfo, netSpeedPartLen))
	}
	speed, err := strconv.Atoi(speedInfo[0])
	if err != nil {
		return common.RetError, buildHccnErrA5("convert speed from string failed", err)
	}

	return speed, nil
}

func buildHccnErrA5(msg string, err error) error {
	return fmt.Errorf("get npu %s info failed,error is :%v", msg, err)
}

// GetNPUUbStatInfo  get npu ub stat information
func GetNPUUbStatInfo(logicID, udieID, portID int32) (map[string]string, error) {
	args := []string{"-g", "-stat", "-i", strconv.Itoa(int(logicID)), "-u", strconv.Itoa(int(udieID)),
		"-p", strconv.Itoa(int(portID))}
	// command example: hccn_tool -g -stat -i 0 -u 0 -p 4
	outStr, err := getInfoFromHccnTool(args...)
	if err != nil {
		return nil, buildHccnErrA5("ub stat info", err)
	}
	lines := strings.Split(outStr, newLine)
	ubStatInfoMap := make(map[string]string)
	for _, line := range lines {
		ubParts := strings.Split(line, colon)
		if len(ubParts) != opticalPartLen {
			continue
		}
		ubKey := strings.TrimSpace(ubParts[0])
		ubValue := strings.TrimSpace(ubParts[1])
		ubStatInfoMap[ubKey] = ubValue
	}
	return ubStatInfoMap, nil
}

// GetIntDataFromStr get int data from string with space
func GetIntDataFromStr(str, dataType string) int {
	if str == "" || strings.Contains(str, naValue) || strings.Contains(str, notSupport) {
		return common.RetError
	}
	intData, err := strconv.Atoi(str)
	if err != nil {
		hwlog.RunLog.Errorf("convert %v ub data type: %v to an int number failed, err: %v", str, dataType, err)
		return common.RetError
	}
	return intData
}
