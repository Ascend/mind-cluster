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
	"testing"

	"github.com/smartystreets/goconvey/convey"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

const (
	ascend910 = "huawei.com/Ascend910"
)

// npuQuantity builds a Quantity for the given NPU count.
func npuQuantity(count int) resource.Quantity {
	q := resource.NewQuantity(int64(count), resource.DecimalSI)
	return resource.MustParse(q.String())
}

// npuContainer builds a container with the given NPU request count.
func npuContainer(name string, count int) v1.Container {
	return v1.Container{
		Name:  name,
		Image: "test",
		Resources: v1.ResourceRequirements{
			Requests: v1.ResourceList{
				v1.ResourceName(ascend910): npuQuantity(count),
			},
		},
	}
}

// npuLimitContainer builds a container with only NPU limits.
func npuLimitContainer(name string, count int) v1.Container {
	return v1.Container{
		Name:  name,
		Image: "test",
		Resources: v1.ResourceRequirements{
			Limits: v1.ResourceList{
				v1.ResourceName(ascend910): npuQuantity(count),
			},
		},
	}
}

// assertNPU asserts the NPU value in result equals expected.
func assertNPU(result *v1.ResourceList, expected int64) {
	npu := (*result)[v1.ResourceName(ascend910)]
	convey.So(npu.Value(), convey.ShouldEqual, expected)
}

// podWithContainers builds a PodSpec from the given containers.
func podWithContainers(containers ...v1.Container) v1.PodSpec {
	return v1.PodSpec{Containers: containers}
}

// mixedContainer builds a container with cpu req and cpu+npu limits.
func mixedContainer() v1.Container {
	return v1.Container{
		Name:  "infer",
		Image: "test",
		Resources: v1.ResourceRequirements{
			Requests: v1.ResourceList{
				v1.ResourceCPU: resource.MustParse("2"),
			},
			Limits: v1.ResourceList{
				v1.ResourceCPU:             resource.MustParse("4"),
				v1.ResourceName(ascend910): npuQuantity(8),
			},
		},
	}
}

// fullContainer builds a container with cpu/memory/npu req and limits.
func fullContainer() v1.Container {
	return v1.Container{
		Name:  "infer",
		Image: "test",
		Resources: v1.ResourceRequirements{
			Requests: v1.ResourceList{
				v1.ResourceCPU:             resource.MustParse("16"),
				v1.ResourceMemory:          resource.MustParse("64Gi"),
				v1.ResourceName(ascend910): npuQuantity(1),
			},
			Limits: v1.ResourceList{
				v1.ResourceCPU:             resource.MustParse("32"),
				v1.ResourceMemory:          resource.MustParse("128Gi"),
				v1.ResourceName(ascend910): npuQuantity(1),
			},
		},
	}
}

// TestCalcMinResourcesReturnNil tests nil-returning edge cases.
func TestCalcMinResourcesReturnNil(t *testing.T) {
	convey.Convey("should return nil when replicas <= 0", t, func() {
		convey.So(CalcMinResources(0, v1.PodSpec{}), convey.ShouldBeNil)
		convey.So(CalcMinResources(-1, v1.PodSpec{}), convey.ShouldBeNil)
	})

	convey.Convey("should return nil when no requests declared", t, func() {
		podSpec := podWithContainers(
			v1.Container{Name: "test", Image: "test"},
		)
		convey.So(CalcMinResources(2, podSpec), convey.ShouldBeNil)
	})
}

// TestCalcMinResourcesSingleContainer tests single-container cases.
func TestCalcMinResourcesSingleContainer(t *testing.T) {
	convey.Convey("should calculate NPU for replicas=2", t, func() {
		podSpec := podWithContainers(npuContainer("infer", 8))
		assertNPU(CalcMinResources(2, podSpec), 16)
	})

	convey.Convey("should calculate NPU for replicas=1", t, func() {
		podSpec := podWithContainers(npuContainer("infer", 8))
		assertNPU(CalcMinResources(1, podSpec), 8)
	})
}

// TestCalcMinResourcesMultiContainer tests multi-container accumulation.
func TestCalcMinResourcesMultiContainer(t *testing.T) {
	convey.Convey("should sum NPU across containers, then multiply", t, func() {
		// (4 + 4) * 2 = 16
		podSpec := podWithContainers(
			npuContainer("a", 4),
			npuContainer("b", 4),
		)
		assertNPU(CalcMinResources(2, podSpec), 16)
	})
}

// TestCalcMinResourcesFallback tests basic Limits fallback.
func TestCalcMinResourcesFallback(t *testing.T) {
	convey.Convey("should fall back to Limits when Requests omitted", t, func() {
		// Limits=8, replicas=2, expect 16
		podSpec := podWithContainers(npuLimitContainer("infer", 8))
		assertNPU(CalcMinResources(2, podSpec), 16)
	})

	convey.Convey("should use Requests when both set", t, func() {
		// Requests=4, Limits=8, replicas=2, expect 8
		c := npuContainer("infer", 4)
		c.Resources.Limits = v1.ResourceList{
			v1.ResourceName(ascend910): npuQuantity(8),
		}
		podSpec := podWithContainers(c)
		assertNPU(CalcMinResources(2, podSpec), 8)
	})
}

// TestCalcMinResourcesPerResourceFallback tests per-resource fallback.
func TestCalcMinResourcesPerResourceFallback(t *testing.T) {
	convey.Convey("should fall back per-resource for mixed req/limit", t, func() {
		// req has cpu only; limit has cpu + npu
		// cpu uses req(2), npu falls back to limit(8); replicas=2
		c := mixedContainer()
		podSpec := podWithContainers(c)
		result := CalcMinResources(2, podSpec)
		convey.So(result, convey.ShouldNotBeNil)
		cpu := (*result)[v1.ResourceCPU]
		convey.So(cpu.Value(), convey.ShouldEqual, 4)
		assertNPU(result, 16)
	})
}

// TestCalcMinResourcesMultiResource tests cpu/memory/npu together.
func TestCalcMinResourcesMultiResource(t *testing.T) {
	convey.Convey("should handle cpu/memory/npu together", t, func() {
		c := fullContainer()
		podSpec := podWithContainers(c)
		result := CalcMinResources(2, podSpec)
		convey.So(result, convey.ShouldNotBeNil)
		cpu := (*result)[v1.ResourceCPU]
		convey.So(cpu.Value(), convey.ShouldEqual, 32)
		mem := (*result)[v1.ResourceMemory]
		convey.So(mem.Value(), convey.ShouldEqual, 64*2*1024*1024*1024)
		assertNPU(result, 2)
	})
}

// TestAddResourceList tests the AddResourceList function.
func TestAddResourceList(t *testing.T) {
	convey.Convey("should add requests to list", t, func() {
		list := v1.ResourceList{}
		req := v1.ResourceList{
			v1.ResourceCPU: resource.MustParse("4"),
		}
		AddResourceList(list, req, nil)
		cpu := list[v1.ResourceCPU]
		convey.So(cpu.Value(), convey.ShouldEqual, 4)
	})

	convey.Convey("should accumulate when adding same resource twice", t, func() {
		list := v1.ResourceList{}
		req := v1.ResourceList{
			v1.ResourceCPU: resource.MustParse("2"),
		}
		AddResourceList(list, req, nil)
		AddResourceList(list, req, nil)
		cpu := list[v1.ResourceCPU]
		convey.So(cpu.Value(), convey.ShouldEqual, 4)
	})

	convey.Convey("should fall back to limits when req is nil", t, func() {
		list := v1.ResourceList{}
		limit := v1.ResourceList{
			v1.ResourceName(ascend910): resource.MustParse("8"),
		}
		AddResourceList(list, nil, limit)
		npu := list[v1.ResourceName(ascend910)]
		convey.So(npu.Value(), convey.ShouldEqual, 8)
	})

	convey.Convey("should prefer req over limit for same resource", t, func() {
		list := v1.ResourceList{}
		req := v1.ResourceList{
			v1.ResourceCPU: resource.MustParse("2"),
		}
		limit := v1.ResourceList{
			v1.ResourceCPU: resource.MustParse("8"),
		}
		AddResourceList(list, req, limit)
		cpu := list[v1.ResourceCPU]
		convey.So(cpu.Value(), convey.ShouldEqual, 2)
	})
}

// TestAddResourceListPerResourceFallback tests per-resource fallback:
// resources present in req use req; resources only in limit use limit.
func TestAddResourceListPerResourceFallback(t *testing.T) {
	convey.Convey("should use limit for resources absent from req", t, func() {
		list := v1.ResourceList{}
		req := v1.ResourceList{
			v1.ResourceCPU: resource.MustParse("2"),
		}
		limit := v1.ResourceList{
			v1.ResourceCPU:             resource.MustParse("4"),
			v1.ResourceName(ascend910): npuQuantity(8),
		}
		AddResourceList(list, req, limit)
		cpu := list[v1.ResourceCPU]
		convey.So(cpu.Value(), convey.ShouldEqual, 2)
		npu := list[v1.ResourceName(ascend910)]
		convey.So(npu.Value(), convey.ShouldEqual, 8)
	})
}
