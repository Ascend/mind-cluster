/* Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package topology for write topology of Rack
package topology

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"

	"ascend-common/common-utils/hwlog"
	"ascend-common/common-utils/utils"
)

const (
	// Sha256HashLength sha256 hash length
	Sha256HashLength = 32
	// maxRackNumPerSuperPod max rack numbers of each super pod
	maxRackNumPerSuperPod = 256
	// maxSuperPodNum max super pod numbers
	maxSuperPodNum = 64

	defaultPerm         = 0644
	rackDirPerm         = 0755
	publishCmNamePrefix = "super-pod"
	size50M             = 50 * 1024 * 1024
)

// TopoInfo topo info for store topo file
type TopoInfo struct {
	// SuperPodId super pod id
	SuperPodId int
	// RackId rack id
	RackId int
	// TopoJsonFile topo file target path
	TopoJsonFile string
}

// ToString get json string of topo info
func topoFileToStr(orgFile string) (string, error) {
	topoData, err := utils.ReadLimitBytes(orgFile, size50M)
	if err != nil {
		return "", fmt.Errorf("read topo file failed, path:<%v>; err:<%v>", orgFile, err)
	}
	// check json and unmarshal
	if !json.Valid(topoData) {
		return "", fmt.Errorf("topo file is not json, path:<%v>", orgFile)
	}
	return string(topoData), nil
}

// get file sha256
func getFileHash(filePath string) ([Sha256HashLength]byte, error) {
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return [Sha256HashLength]byte{}, err
	}
	return sha256.Sum256(fileContent), nil
}

// ToFile write topo info to file
func ToFile(filePath string, orgFile string) error {
	newTopoStr, err := topoFileToStr(orgFile)
	if err != nil {
		return err
	}

	newHash := sha256.Sum256([]byte(newTopoStr))
	var originalHash [Sha256HashLength]byte

	if _, err := os.Stat(filePath); err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	} else {
		originalHash, err = getFileHash(filePath)
		if err != nil {
			return err
		}
	}

	// file content is same,no need to write
	if newHash == originalHash {
		hwlog.RunLog.Infof("The file %s is up to date, no need to write", filePath)
		return nil
	}

	f, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, defaultPerm)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err = f.WriteString(newTopoStr); err != nil {
		return err
	}

	if err = os.Chmod(filePath, defaultPerm); err != nil {
		return err
	}

	return nil
}
