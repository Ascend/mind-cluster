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

package rdma

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"ascend-common/common-utils/utils"
)

const (
	rdmaDeviceTableName        = "rdma_devices.conf"
	rdmaConfigMaxSizeMB        = 100
	rdmaDeviceTableMaxLineSize = 64 * 1024
)

var rdmaCollectionPattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_-]{0,63}$`)

type rdmaHardwareIdentity struct {
	vendorID  string
	deviceID  string
	classCode string
}

type rdmaDeviceRule struct {
	identity   rdmaHardwareIdentity
	collection string
	rank       int
}

// DiscoverRdmaCollections selects driver collections in HCA request order and deduplicates them.
func DiscoverRdmaCollections(dir string, hcaNames []string) ([]string, error) {
	if len(hcaNames) == 0 {
		return nil, nil
	}

	rules, err := loadRdmaDeviceTable(dir)
	if err != nil {
		return nil, err
	}
	collections := make([]string, 0)
	seen := make(map[string]bool)
	for _, hcaName := range hcaNames {
		hcaSysfsPath, err := selectedRdmaHCAPath(hcaName)
		if err != nil {
			return nil, err
		}
		identity, err := readRdmaIdentity(hcaSysfsPath)
		if err != nil {
			return nil, fmt.Errorf("HCA %s: %w", hcaName, err)
		}
		matchedCollections := matchRdmaCollections(rules, identity)
		if len(matchedCollections) == 0 {
			return nil, fmt.Errorf("no matching RDMA collection for HCA %s identity %v in %s",
				hcaName, identity, filepath.Join(dir, rdmaDeviceTableName))
		}
		for _, collection := range matchedCollections {
			if !seen[collection] {
				seen[collection] = true
				collections = append(collections, collection)
			}
		}
	}
	return collections, nil
}

func loadRdmaDeviceTable(dir string) ([]rdmaDeviceRule, error) {
	path := filepath.Join(dir, rdmaDeviceTableName)
	if err := checkRdmaConfigFile(path); err != nil {
		return nil, err
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open RDMA device table %s: %w", path, err)
	}
	defer file.Close()
	return parseRdmaDeviceTable(file, path)
}

func parseRdmaDeviceTable(reader io.Reader, path string) ([]rdmaDeviceRule, error) {
	scanner := bufio.NewScanner(reader)
	// Bound each line separately from the file size, allowing LF or CRLF at the limit.
	scanner.Buffer(make([]byte, 4096), rdmaDeviceTableMaxLineSize+2)
	rules := make([]rdmaDeviceRule, 0)
	lineNo := 0
	for scanner.Scan() {
		lineNo++
		if len(scanner.Bytes()) > rdmaDeviceTableMaxLineSize {
			return nil, rdmaDeviceTableLineSizeError(path, lineNo)
		}
		line := strings.TrimSpace(strings.SplitN(scanner.Text(), "#", 2)[0])
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) != 4 {
			return nil, fmt.Errorf("RDMA device table %s line %d: expected four columns, got %d",
				path, lineNo, len(fields))
		}
		collection, err := normalizeRdmaCollection(fields[3])
		if err != nil {
			return nil, fmt.Errorf("RDMA device table %s line %d: %w", path, lineNo, err)
		}
		rule := rdmaDeviceRule{collection: collection, identity: rdmaHardwareIdentity{
			vendorID:  normalizeRdmaIdentity(fields[0]),
			deviceID:  normalizeRdmaIdentity(fields[1]),
			classCode: normalizeRdmaIdentity(fields[2]),
		}}
		for _, field := range []string{rule.identity.vendorID, rule.identity.deviceID, rule.identity.classCode} {
			if field != "*" {
				rule.rank++
			}
		}
		rules = append(rules, rule)
	}
	if err := scanner.Err(); err != nil {
		if errors.Is(err, bufio.ErrTooLong) {
			return nil, rdmaDeviceTableLineSizeError(path, lineNo+1)
		}
		return nil, fmt.Errorf("scan RDMA device table %s near line %d: %w", path, lineNo+1, err)
	}
	if len(rules) == 0 {
		return nil, fmt.Errorf("RDMA device table %s has no records", path)
	}
	return rules, nil
}

func rdmaDeviceTableLineSizeError(path string, lineNo int) error {
	return fmt.Errorf("RDMA device table %s line %d exceeds maximum line length of %d bytes",
		path, lineNo, rdmaDeviceTableMaxLineSize)
}

func readRdmaIdentity(hcaSysfsPath string) (rdmaHardwareIdentity, error) {
	var identity rdmaHardwareIdentity
	var err error
	if identity.vendorID, err = readRdmaIdentityField(hcaSysfsPath, "vendor"); err != nil {
		return identity, err
	}
	if identity.deviceID, err = readRdmaIdentityField(hcaSysfsPath, "device"); err != nil {
		return identity, err
	}
	if identity.classCode, err = readRdmaIdentityField(hcaSysfsPath, "class_code"); err != nil {
		return identity, err
	}
	return identity, nil
}

func readRdmaIdentityField(hcaSysfsPath, fieldName string) (string, error) {
	data, err := os.ReadFile(filepath.Join(hcaSysfsPath, hcaDeviceDir, fieldName))
	if err != nil {
		return "", fmt.Errorf("read field %s: %w", fieldName, err)
	}
	value := normalizeRdmaIdentity(string(data))
	if value == "" {
		return "", fmt.Errorf("read field %s: empty identity", fieldName)
	}
	return value, nil
}

// matchRdmaCollections selects the most specific matching tuple and collects all its mapped collections.
// Table order breaks ties between different tuples; callers deduplicate names, and no match yields an empty list.
func matchRdmaCollections(rules []rdmaDeviceRule, identity rdmaHardwareIdentity) []string {
	var collections []string
	var bestIdentity rdmaHardwareIdentity
	bestRank := -1
	for _, rule := range rules {
		matched := matchesRdmaIdentity(rule.identity, identity)
		if !matched || rule.rank < bestRank {
			continue
		}
		if rule.rank > bestRank {
			bestRank, bestIdentity = rule.rank, rule.identity
			collections = nil
		}
		if rule.identity == bestIdentity {
			collections = append(collections, rule.collection)
		}
	}
	return collections
}

func matchesRdmaIdentity(rule, identity rdmaHardwareIdentity) bool {
	return (rule.vendorID == "*" || rule.vendorID == identity.vendorID) &&
		(rule.deviceID == "*" || rule.deviceID == identity.deviceID) &&
		(rule.classCode == "*" || rule.classCode == identity.classCode)
}

// checkRdmaConfigFile checks ownership, parent permissions, and the 100 MB limit; symlinks are rejected.
func checkRdmaConfigFile(path string) error {
	if _, err := utils.RealFileChecker(path, true, false, rdmaConfigMaxSizeMB); err != nil {
		return fmt.Errorf("check RDMA configuration file %s: %w", path, err)
	}
	return nil
}

func normalizeRdmaCollection(name string) (string, error) {
	if !rdmaCollectionPattern.MatchString(name) {
		return "", fmt.Errorf("invalid RDMA collection name %q", name)
	}
	return strings.ToLower(name), nil
}

func normalizeRdmaIdentity(value string) string {
	return strings.TrimPrefix(strings.ToLower(strings.TrimSpace(value)), "0x")
}
