/*
   Copyright(C) 2026. Huawei Technologies Co.,Ltd. All rights reserved.
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

package utils

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"sync"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"ascend-common/common-utils/hwlog"
	"ascend-common/common-utils/utils"
)

func init() {
	_ = hwlog.InitRunLogger(&hwlog.LogConfig{OnlyToStdout: true}, context.Background())
}

func resetNpuNicMappingCache() {
	npuNicMappingCache = nil
	npuNicMappingErr = nil
	npuNicMappingOnce = sync.Once{}
}

func TestLoadNpuNicMappingSuccess(t *testing.T) {
	patches := gomonkey.ApplyFunc(utils.LoadFile, func(name string) ([]byte, error) {
		return []byte(`{"npuNics":[{"npuId":0,"nicNames":["ens2f0","ens0f2","ens1f0"]}]}`), nil
	})
	defer patches.Reset()

	convey.Convey("Given a valid mapping file on disk", t, func() {
		mapping, err := loadNpuNicMapping()
		convey.Convey("Then mapping should be parsed without error", func() {
			convey.So(err, convey.ShouldBeNil)
			convey.So(mapping, convey.ShouldNotBeNil)
			convey.So(len(mapping.NpuNics), convey.ShouldEqual, 1)
			convey.So(mapping.NpuNics[0].NpuId, convey.ShouldEqual, 0)
			convey.So(mapping.NpuNics[0].NicNames, convey.ShouldResemble, []string{"ens2f0", "ens0f2", "ens1f0"})
		})
	})
}

func TestLoadNpuNicMappingNotFound(t *testing.T) {
	patches := gomonkey.ApplyFunc(utils.LoadFile, func(name string) ([]byte, error) {
		return nil, nil
	})
	defer patches.Reset()

	convey.Convey("Given the mapping file does not exist", t, func() {
		mapping, err := loadNpuNicMapping()
		convey.Convey("Then nil mapping and nil error should be returned", func() {
			convey.So(err, convey.ShouldBeNil)
			convey.So(mapping, convey.ShouldBeNil)
		})
	})
}

// TestLoadNpuNicMappingReadError tests loadNpuNicMapping when reading the mapping file fails.
func TestLoadNpuNicMappingReadError(t *testing.T) {
	patches := gomonkey.ApplyFunc(utils.LoadFile, func(name string) ([]byte, error) {
		return nil, errors.New("read error")
	})
	defer patches.Reset()

	convey.Convey("Given reading the mapping file fails", t, func() {
		_, err := loadNpuNicMapping()
		convey.Convey("Then an error should be returned", func() {
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

func TestLoadNpuNicMappingParseError(t *testing.T) {
	patches := gomonkey.ApplyFunc(utils.LoadFile, func(name string) ([]byte, error) {
		return []byte(`invalid json`), nil
	})
	defer patches.Reset()

	convey.Convey("Given the mapping file contains invalid JSON", t, func() {
		_, err := loadNpuNicMapping()
		convey.Convey("Then a parse error should be returned", func() {
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

func TestGetNicNamesConfigNotFound(t *testing.T) {
	resetNpuNicMappingCache()

	patches := gomonkey.ApplyFunc(utils.LoadFile, func(name string) ([]byte, error) {
		return nil, nil
	})
	defer patches.Reset()

	convey.Convey("Given the mapping config file does not exist", t, func() {
		_, err := GetNicNames(0)
		convey.Convey("Then an error should be returned", func() {
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

func TestReadCardType(t *testing.T) {
	convey.Convey("Given /sys/class/net paths for a NIC", t, func() {
		convey.Convey("When device/card_type exists, it should be used", func() {
			patches := gomonkey.ApplyFunc(os.ReadFile, func(name string) ([]byte, error) {
				convey.So(name, convey.ShouldEqual, "/sys/class/net/ens0f0/device/card_type")
				return []byte("A5Server\n"), nil
			})
			defer patches.Reset()

			cardType, ok := readCardType("ens0f0")
			convey.So(ok, convey.ShouldBeTrue)
			convey.So(cardType, convey.ShouldEqual, "A5Server")
		})

		convey.Convey("When device/card_type is missing but card_type exists, it should fall back", func() {
			calls := 0
			patches := gomonkey.ApplyFunc(os.ReadFile, func(name string) ([]byte, error) {
				calls++
				if calls == 1 {
					return nil, errors.New("no such file")
				}
				convey.So(name, convey.ShouldEqual, "/sys/class/net/ens0f0/card_type")
				return []byte("A5Pod200G\n"), nil
			})
			defer patches.Reset()

			cardType, ok := readCardType("ens0f0")
			convey.So(ok, convey.ShouldBeTrue)
			convey.So(cardType, convey.ShouldEqual, "A5Pod200G")
		})

		convey.Convey("When both paths fail, ok should be false", func() {
			patches := gomonkey.ApplyFunc(os.ReadFile, func(name string) ([]byte, error) {
				return nil, errors.New("no such file")
			})
			defer patches.Reset()

			_, ok := readCardType("ens0f0")
			convey.So(ok, convey.ShouldBeFalse)
		})
	})
}

func TestLoadNpuNicMappingByForm(t *testing.T) {
	patches := gomonkey.ApplyFunc(utils.LoadFile, func(name string) ([]byte, error) {
		return []byte(`[
			{"productType":"Server","npuNics":[{"npuId":0,"nicNames":["ens2f0","ens0f2"]}]},
			{"productType":"PoD","npuNics":[{"npuId":1,"nicNames":["ens2f1","ens0f0"]}]}
		]`), nil
	}).ApplyFunc(machineType, func() (string, error) {
		return "Server", nil
	})
	defer patches.Reset()

	convey.Convey("Given a per-form mapping config", t, func() {
		convey.Convey("When the current machine type matches a product", func() {
			mapping, err := loadNpuNicMapping()
			convey.Convey("Then the mapping of the matching form should be returned", func() {
				convey.So(err, convey.ShouldBeNil)
				convey.So(mapping, convey.ShouldNotBeNil)
				convey.So(len(mapping.NpuNics), convey.ShouldEqual, 1)
				convey.So(mapping.NpuNics[0].NpuId, convey.ShouldEqual, 0)
				convey.So(mapping.NpuNics[0].NicNames, convey.ShouldResemble, []string{"ens2f0", "ens0f2"})
			})
		})
	})
}

func TestLoadNpuNicMappingFormNotFound(t *testing.T) {
	patches := gomonkey.ApplyFunc(utils.LoadFile, func(name string) ([]byte, error) {
		return []byte(`[{"productType":"PoD","npuNics":[{"npuId":1,"nicNames":["ens2f1"]}]}]`), nil
	}).ApplyFunc(machineType, func() (string, error) {
		return "Server", nil
	})
	defer patches.Reset()

	convey.Convey("Given a per-form mapping config without the current machine type", t, func() {
		convey.Convey("Then an error should be returned", func() {
			_, err := loadNpuNicMapping()
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

// TestBuildReverseIndex tests buildReverseIndex for nil, empty, single and merged-primary-NIC items.
func TestBuildReverseIndex(t *testing.T) {
	convey.Convey("Given a nil item list", t, func() {
		index := buildReverseIndex(nil)
		convey.Convey("Then an empty index is returned", func() {
			convey.So(index, convey.ShouldBeEmpty)
		})
	})
	convey.Convey("Given items with empty NicNames", t, func() {
		index := buildReverseIndex([]NpuNicItem{
			{NpuId: 0, NicNames: nil},
			{NpuId: 1, NicNames: []string{}},
		})
		convey.Convey("Then they are skipped", func() {
			convey.So(index, convey.ShouldBeEmpty)
		})
	})
	convey.Convey("Given a single item", t, func() {
		index := buildReverseIndex([]NpuNicItem{
			{NpuId: 2, NicNames: []string{"ens2f0", "ens0f2"}},
		})
		convey.Convey("Then the primary NIC maps to its NPU id", func() {
			convey.So(index, convey.ShouldResemble, map[string][]int{"ens2f0": {2}})
		})
	})
	convey.Convey("Given multiple items sharing the same primary NIC", t, func() {
		index := buildReverseIndex([]NpuNicItem{
			{NpuId: 0, NicNames: []string{"ens2f0"}},
			{NpuId: 1, NicNames: []string{"ens2f0"}},
		})
		convey.Convey("Then the NPU ids are merged under that NIC", func() {
			convey.So(index, convey.ShouldResemble, map[string][]int{"ens2f0": {0, 1}})
		})
	})
}

func TestValidateSysfsPath(t *testing.T) {
	convey.Convey("Given a path under /sys", t, func() {
		convey.Convey("Then it is valid", func() {
			convey.So(validateSysfsPath("/sys/class/net/ens0f0"), convey.ShouldBeTrue)
		})
	})
	convey.Convey("Given a path outside /sys", t, func() {
		convey.Convey("Then it is invalid", func() {
			convey.So(validateSysfsPath("/etc/hostname"), convey.ShouldBeFalse)
			convey.So(validateSysfsPath(""), convey.ShouldBeFalse)
		})
	})
}

// TestCardTypeOf tests cardTypeOf mapping raw card types to Server/PoD forms.
func TestCardTypeOf(t *testing.T) {
	convey.Convey("Given an A5Server card_type", t, func() {
		patches := gomonkey.ApplyFunc(os.ReadFile, func(name string) ([]byte, error) {
			return []byte("A5Server\n"), nil
		})
		defer patches.Reset()

		cardType, ok := cardTypeOf("ens0f0")
		convey.Convey("Then Server is returned", func() {
			convey.So(ok, convey.ShouldBeTrue)
			convey.So(cardType, convey.ShouldEqual, "Server")
		})
	})
	convey.Convey("Given an A5Pod card_type", t, func() {
		patches := gomonkey.ApplyFunc(os.ReadFile, func(name string) ([]byte, error) {
			return []byte("A5Pod200G\n"), nil
		})
		defer patches.Reset()

		cardType, ok := cardTypeOf("ens0f0")
		convey.Convey("Then PoD is returned", func() {
			convey.So(ok, convey.ShouldBeTrue)
			convey.So(cardType, convey.ShouldEqual, "PoD")
		})
	})
	convey.Convey("Given an unknown card_type", t, func() {
		patches := gomonkey.ApplyFunc(os.ReadFile, func(name string) ([]byte, error) {
			return []byte("Unknown\n"), nil
		})
		defer patches.Reset()

		_, ok := cardTypeOf("ens0f0")
		convey.Convey("Then it is not recognized", func() {
			convey.So(ok, convey.ShouldBeFalse)
		})
	})
}

type fakeDirEntry struct{ name string }

func (f fakeDirEntry) Name() string               { return f.name }
func (f fakeDirEntry) IsDir() bool                { return false }
func (f fakeDirEntry) Type() fs.FileMode          { return 0 }
func (f fakeDirEntry) Info() (fs.FileInfo, error) { return nil, nil }

func TestMachineType(t *testing.T) {
	convey.Convey("Given an ens device with a Server card type", t, func() {
		patches := gomonkey.ApplyFunc(os.ReadDir, func(name string) ([]os.DirEntry, error) {
			return []os.DirEntry{fakeDirEntry{name: "ens0f0"}}, nil
		}).ApplyFunc(os.ReadFile, func(name string) ([]byte, error) {
			return []byte("A5Server\n"), nil
		})
		defer patches.Reset()

		mt, err := machineType()
		convey.Convey("Then Server is returned", func() {
			convey.So(err, convey.ShouldBeNil)
			convey.So(mt, convey.ShouldEqual, "Server")
		})
	})
	convey.Convey("Given no ens device but a non-ens device with a Pod card type", t, func() {
		patches := gomonkey.ApplyFunc(os.ReadDir, func(name string) ([]os.DirEntry, error) {
			return []os.DirEntry{fakeDirEntry{name: "eth0"}}, nil
		}).ApplyFunc(os.ReadFile, func(name string) ([]byte, error) {
			return []byte("A5Pod100G\n"), nil
		})
		defer patches.Reset()

		mt, err := machineType()
		convey.Convey("Then PoD is returned", func() {
			convey.So(err, convey.ShouldBeNil)
			convey.So(mt, convey.ShouldEqual, "PoD")
		})
	})
	convey.Convey("Given no readable card type on any device", t, func() {
		patches := gomonkey.ApplyFunc(os.ReadDir, func(name string) ([]os.DirEntry, error) {
			return []os.DirEntry{fakeDirEntry{name: "ens0f0"}, fakeDirEntry{name: "eth0"}}, nil
		}).ApplyFunc(os.ReadFile, func(name string) ([]byte, error) {
			return nil, errors.New("no such file")
		})
		defer patches.Reset()

		_, err := machineType()
		convey.Convey("Then an error is returned", func() {
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

func TestSelectByForm(t *testing.T) {
	convey.Convey("Given products containing the current machine form", t, func() {
		patches := gomonkey.ApplyFunc(machineType, func() (string, error) {
			return "PoD", nil
		})
		defer patches.Reset()

		index, form, err := selectByForm([]ProductMapping{
			{ProductType: "Server", NpuNics: []NpuNicItem{{NpuId: 0, NicNames: []string{"ens2f0"}}}},
			{ProductType: "PoD", NpuNics: []NpuNicItem{{NpuId: 1, NicNames: []string{"ens2f1"}}}},
		})
		convey.Convey("Then the matching form's reverse index is returned", func() {
			convey.So(err, convey.ShouldBeNil)
			convey.So(form, convey.ShouldEqual, "PoD")
			convey.So(index, convey.ShouldResemble, map[string][]int{"ens2f1": {1}})
		})
	})
	convey.Convey("Given products without the current machine form", t, func() {
		patches := gomonkey.ApplyFunc(machineType, func() (string, error) {
			return "Server", nil
		})
		defer patches.Reset()

		_, _, err := selectByForm([]ProductMapping{
			{ProductType: "PoD", NpuNics: []NpuNicItem{{NpuId: 1, NicNames: []string{"ens2f1"}}}},
		})
		convey.Convey("Then an error is returned", func() {
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

func TestGetAffectedNPU(t *testing.T) {
	convey.Convey("Given a reverse index with a primary NIC", t, func() {
		dpuToNpuIDs = map[string][]int{"ens2f0": {0, 1}}
		convey.Convey("Then the affected NPU ids are returned", func() {
			convey.So(GetAffectedNPU("ens2f0"), convey.ShouldResemble, []int{0, 1})
		})
		convey.Convey("Then an unmapped NIC returns an empty list", func() {
			convey.So(GetAffectedNPU("ens9f0"), convey.ShouldBeEmpty)
		})
		convey.Convey("Then an empty NIC name returns an empty list", func() {
			convey.So(GetAffectedNPU(""), convey.ShouldBeEmpty)
		})
	})
}

func TestGetNicNamesSuccess(t *testing.T) {
	resetNpuNicMappingCache()

	patches := gomonkey.ApplyFunc(utils.LoadFile, func(name string) ([]byte, error) {
		return []byte(`{"npuNics":[{"npuId":0,"nicNames":["ens2f0","ens0f2"]},{"npuId":1,"nicNames":["ens2f1"]}]}`), nil
	})
	defer patches.Reset()

	convey.Convey("Given a valid mapping config", t, func() {
		convey.Convey("Then the NIC names of a mapped NPU id are returned", func() {
			names, err := GetNicNames(0)
			convey.So(err, convey.ShouldBeNil)
			convey.So(names, convey.ShouldResemble, []string{"ens2f0", "ens0f2"})
		})
		convey.Convey("Then an unmapped NPU id returns an error", func() {
			_, err := GetNicNames(99)
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}
