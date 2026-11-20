#!/usr/bin/env python3
# -*- coding: utf-8 -*-
# Copyright 2026 Huawei Technologies Co., Ltd
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
# ==============================================================================

"""Offline tests for the four query tools: parsers and status aggregation."""

# pylint: disable=too-many-lines,duplicate-code  # one file per query area keeps fixtures local; shared prologue is intentional

import json

import pytest

from agent_core.datasources.clusterd.device_summary import parse_cluster_device_shards
from agent_core.datasources.query_context import QueryContext
from agent_core.datasources.gateway import SourceReadResult
from agent_core.datasources.public_faults import parse_public_faults
from agent_core.datasources.utils import decode_json_string
from agent_core.query_schemas import (
    FaultQuery,
    IsolationQuery,
    QueryFailure,
    ResourceQuery,
)
from agent_core.tools.cluster_resource import cluster_resource
from agent_core.tools.faulty_device import faulty_device
from agent_core.tools.isolated_device import isolated_device

NPU = "huawei.com/Ascend910"
NPU2 = "huawei.com/npu"


class FakeGateway:
    """Minimal read-only gateway double that hands out source ids like the real one."""

    def __init__(self, ctx, cms=None, nodes=None, pods=None):
        self.ctx = ctx
        self.cms = dict(cms or {})
        self.nodes = list(nodes or [])
        self.pods = list(pods or [])

    def _source_id(self):
        return self.ctx.next_source_id()

    def get_cm(self, namespace, name, absent_expected=False, *, use_cache=True):
        data = self.cms.get((namespace, name))
        if data is not None:
            status = "ok" if data else "empty"
            return SourceReadResult(
                status=status,
                source_id=self._source_id(),
                items=[{"name": name, "data": data}],
                resource_version="1",
                observed=len(data),
            )
        status = "absent_expected" if absent_expected else "missing"
        return SourceReadResult(
            status=status,
            source_id=self._source_id(),
            error_code="SOURCE_NOT_FOUND" if status == "missing" else None,
        )

    def list_cms(self, namespace, label_selector=None, name_prefix=None):
        items = [
            {"name": name, "data": data}
            for (ns, name), data in self.cms.items()
            if ns == namespace and (not name_prefix or name.startswith(name_prefix))
        ]
        status = "ok" if items else "empty"
        return SourceReadResult(status=status, source_id=self._source_id(), items=items, observed=len(items))

    def list_nodes(self, names=None):
        if names is not None:
            items = [node for node in self.nodes if node["metadata"]["name"] in names]
            missing = [name for name in names if name not in {n["metadata"]["name"] for n in items}]
            return SourceReadResult(status="ok", source_id=self._source_id(), items=items, missing=missing)
        status = "ok" if self.nodes else "empty"
        return SourceReadResult(status=status, source_id=self._source_id(), items=self.nodes)

    def list_pods(self, namespace=None):
        items = [pod for pod in self.pods if namespace is None or pod["metadata"]["namespace"] == namespace]
        status = "ok" if items else "empty"
        return SourceReadResult(status=status, source_id=self._source_id(), items=items)


def make_ctx(query, cms=None, nodes=None, pods=None):
    ctx = QueryContext(request=query)
    ctx.check_scope(query)
    if nodes is None:
        # Scoped queries validate that the named nodes exist; give every requested node a
        # Node object unless the test explicitly supplies the cluster's node set.
        nodes = [{"metadata": {"name": name}} for name in (ctx.requested_nodes or [])]
    ctx.gateway = FakeGateway(ctx, cms=cms, nodes=nodes, pods=pods)
    return ctx


def supplement_cms(nodes=()):
    """Present-but-empty supplemental ConfigMaps; ``nodes`` get a healthy NodeStatus."""
    node_cm = {
        "cluster-info-node-cm": json.dumps({f"mindx-dl-nodeinfo-{node}": {"NodeStatus": "Healthy"} for node in nodes})
    }
    return {
        ("mindx-dl", "cluster-info-node-cm"): node_cm if nodes else {},
        ("mindx-dl", "cluster-info-switch-0"): {},
        ("mindx-dl", "cluster-info-dpu-0"): {},
        ("cluster-system", "statistic-fault-info"): {},
    }


def cluster_device_list():
    return {
        NPU: "Ascend910-0,Ascend910-1",
        NPU + "-Fault": json.dumps(
            [
                {
                    "fault_type": "CardUnhealthy",
                    "npu_name": "Ascend910-1",
                    "fault_handling": "SeparateNPU",
                    "fault_code": "00776711",
                    "fault_time_and_level_map": {"00776711": {"fault_time": 1718702470, "fault_level": "SeparateNPU"}},
                }
            ]
        ),
        NPU + "-Unhealthy": "Ascend910-1",
    }


def pre_separate_device_list():
    """Two canonical PreSeparateNPU devices, plus devices that must not be counted."""
    faults = [
        {
            "fault_type": "CardUnhealthy",
            "npu_name": "Ascend910-0",
            "fault_handling": "PreSeparateNPU",
            "fault_code": "code-0a",
        },
        {
            # Same device again: identity-based dedup must count it once.
            "fault_type": "CardUnhealthy",
            "npu_name": "Ascend910-0",
            "fault_handling": "PreSeparateNPU",
            "fault_code": "code-0b",
        },
        {
            "fault_type": "CardUnhealthy",
            "npu_name": "Ascend910-1",
            "fault_handling": "SeparateNPU",
            "fault_code": "code-1",
        },
        {
            "fault_type": "CardUnhealthy",
            "npu_name": "Ascend910-2",
            "fault_handling": "ManuallySeparateNPU",
            "fault_code": "code-2",
        },
        {
            "fault_type": "CardUnhealthy",
            "npu_name": "Ascend910-3",
            "fault_level": "PreSeparateNPU",
            "fault_code": "code-3",
        },
        {
            # Canonical handling is fault_handling (SeparateNPU); it must not be counted.
            "fault_type": "CardUnhealthy",
            "npu_name": "Ascend910-4",
            "fault_handling": "SeparateNPU",
            "fault_level": "PreSeparateNPU",
            "fault_code": "code-4",
        },
    ]
    return {
        NPU: "Ascend910-0,Ascend910-1,Ascend910-2,Ascend910-3,Ascend910-4",
        NPU + "-Fault": json.dumps(faults),
    }


def test_faults_fromcluster_summary_is_ok():
    query = FaultQuery()
    cms = {
        ("mindx-dl", "cluster-info-device-0"): {
            "cluster-info-device-0": json.dumps(
                {
                    "mindx-dl-deviceinfo-worker-a": {
                        "DeviceList": cluster_device_list(),
                        "UpdateTime": 1718702470,
                        "SuperPodID": 3,
                        "RackID": 1,
                        "ServerIndex": 2,
                    }
                }
            )
        }
    }
    cms.update(supplement_cms())
    result = faulty_device(query, make_ctx(query, cms=cms))

    assert result.status == "ok"
    node = result.data.node["worker-a"]
    assert len(node.devices) == 1
    device = node.devices[0]
    assert device.device_id == "Ascend910-1"
    assert device.device_type == "npu"
    assert device.chip_type == "Ascend910"
    assert device.fault_type == "CardUnhealthy"
    assert device.faults[0].fault_code == "00776711"
    assert device.faults[0].fault_handling == "SeparateNPU"
    assert device.superpod.model_dump() == {"super_pod_id": 3, "rack_id": 1, "server_index": 2}


def test_device_faults_keep_each_source_entry_and_canonical_handling():
    query = FaultQuery(node=["worker-a"])
    cms = supplement_cms(("worker-a",))
    cms[("mindx-dl", "cluster-info-device-0")] = {
        "cluster-info-device-0": json.dumps(
            {
                "mindx-dl-deviceinfo-worker-a": {
                    "DeviceList": pre_separate_device_list(),
                    "UpdateTime": 1718702470,
                }
            }
        )
    }
    result = faulty_device(query, make_ctx(query, cms=cms))

    assert result.status == "ok"
    devices = result.data.node["worker-a"].devices
    # One record per source fault entry; the same card appears once per entry.
    assert len(devices) == 6
    handling = {(device.device_id, device.faults[0].fault_code): device.faults[0].fault_handling for device in devices}
    assert handling == {
        ("Ascend910-0", "code-0a"): "PreSeparateNPU",
        ("Ascend910-0", "code-0b"): "PreSeparateNPU",
        ("Ascend910-1", "code-1"): "SeparateNPU",
        ("Ascend910-2", "code-2"): "ManuallySeparateNPU",
        ("Ascend910-3", "code-3"): "PreSeparateNPU",
        ("Ascend910-4", "code-4"): "SeparateNPU",
    }


def test_dp_fallback_builds_node_devices():
    query = FaultQuery(node=["worker-b"])
    cms = supplement_cms(("worker-b",))
    cms[("kube-system", "mindx-dl-deviceinfo-worker-b")] = {
        "DeviceInfoCfg": json.dumps(
            {
                "DeviceInfo": {"DeviceList": pre_separate_device_list(), "UpdateTime": 1718702470},
                "CheckCode": "x",
            }
        )
    }
    result = faulty_device(query, make_ctx(query, cms=cms))

    assert result.status == "ok"
    node = result.data.node["worker-b"]
    assert {device.device_id for device in node.devices} == {
        "Ascend910-0",
        "Ascend910-1",
        "Ascend910-2",
        "Ascend910-3",
        "Ascend910-4",
    }


def resource_stats_cms():
    """Dual-bucket device facts: two Ascend910 faults (one also isolated) and a clean npu."""
    faults = [
        {
            "fault_type": "CardUnhealthy",
            "npu_name": name,
            "fault_handling": "SeparateNPU",
            "fault_code": f"code-{index}",
        }
        for index, name in enumerate(("Ascend910-1", "Ascend910-2"), start=1)
    ]
    cms = supplement_cms()
    cms[("mindx-dl", "cluster-info-device-0")] = {
        "cluster-info-device-0": json.dumps(
            {
                "mindx-dl-deviceinfo-worker-a": {
                    "DeviceList": {
                        NPU: "Ascend910-0,Ascend910-1,Ascend910-2",
                        NPU + "-Fault": json.dumps(faults),
                        NPU2: "npu-0,npu-1",
                    },
                    "UpdateTime": 1718702470,
                }
            }
        )
    }
    cms[("cluster-system", "clusterd-manual-info-cm")] = {
        "worker-a": json.dumps(
            {
                "Total": ["Ascend910-2"],
                "Detail": {"Ascend910-2": [{"FaultLevel": "ManuallySeparateNPU", "FaultCode": "code-2"}]},
            }
        )
    }
    return cms


def test_uncovered_explicit_target_is_partial():
    """Both nodes exist; one has device facts and the other has only some sources."""
    query = FaultQuery(node=["worker-a", "worker-missing"])
    cms = supplement_cms(("worker-a", "worker-missing"))
    cms[("mindx-dl", "cluster-info-device-0")] = {
        "cluster-info-device-0": json.dumps(
            {
                "mindx-dl-deviceinfo-worker-a": {
                    "DeviceList": pre_separate_device_list(),
                    "UpdateTime": 1718702470,
                }
            }
        )
    }
    nodes = [{"metadata": {"name": name}} for name in ("worker-a", "worker-missing")]
    result = faulty_device(query, make_ctx(query, cms=cms, nodes=nodes))

    assert result.status == "partial"
    assert result.data.node["worker-a"].devices
    # A node with no device source keeps only the fields some source produced.
    assert result.data.node["worker-missing"].devices is None


@pytest.mark.parametrize(
    "handler,query_cls",
    [
        (faulty_device, FaultQuery),
        (isolated_device, IsolationQuery),
        (cluster_resource, ResourceQuery),
    ],
)
def test_scoped_query_with_no_target_node_is_error(handler, query_cls):
    """Every requested node missing: the request has no target, so it fails."""
    query = query_cls(node=["worker-a", "worker-b"])
    with pytest.raises(QueryFailure) as exc:
        handler(query, make_ctx(query, cms={}, nodes=[]))

    assert exc.value.code == "TARGET_NOT_FOUND"
    assert exc.value.http_status == 404
    assert exc.value.path == "node"
    assert exc.value.message == "所有节点都找不到"


def test_scoped_query_with_one_absent_node_is_partial_and_reports_it():
    """A missing node does not sink the surviving nodes: partial + one issue per node."""
    query = FaultQuery(node=["worker-a", "worker-missing"])
    cms = supplement_cms(("worker-a",))
    cms[("mindx-dl", "cluster-info-device-0")] = _cluster_summary("worker-a", 1718702470, cluster_device_list())
    result = faulty_device(
        query,
        make_ctx(query, cms=cms, nodes=[{"metadata": {"name": "worker-a"}}]),
    )

    assert result.status == "partial"
    assert [(issue.code, issue.path, issue.message) for issue in result.errors] == [
        ("TARGET_NOT_FOUND", "node", "节点 worker-missing 不存在")
    ]
    assert list(result.data.node) == ["worker-a"]


def test_cluster_resource_partial_when_some_nodes_absent():
    query = ResourceQuery(node=["worker-a", "worker-missing"])
    node = {"metadata": {"name": "worker-a"}, "status": {"capacity": {NPU: "8"}, "allocatable": {NPU: "8"}}}
    result = cluster_resource(query, make_ctx(query, cms=supplement_cms(("worker-a",)), nodes=[node]))

    assert result.status == "partial"
    assert [issue.message for issue in result.errors] == ["节点 worker-missing 不存在"]
    assert result.data.node["worker-a"].model_extra["Ascend910"].capacity == 8


def test_isolated_device_partial_when_some_nodes_absent():
    query = IsolationQuery(node=["worker-a", "worker-missing"])
    cms = {
        ("cluster-system", "clusterd-manual-info-cm"): {
            "worker-a": json.dumps({"Total": ["Ascend910-1"], "Detail": {}})
        }
    }
    result = isolated_device(query, make_ctx(query, cms=cms, nodes=[{"metadata": {"name": "worker-a"}}]))

    assert result.status == "partial"
    assert [issue.message for issue in result.errors] == ["节点 worker-missing 不存在"]
    assert list(result.data.node) == ["worker-a"]


def test_node_validation_is_skipped_when_node_listing_is_unavailable():
    query = FaultQuery(node=["worker-a"])
    ctx = make_ctx(query, cms=supplement_cms(("worker-a",)))
    ctx.gateway.list_nodes = lambda names=None: SourceReadResult(status="unavailable", error_code="SOURCE_UNAVAILABLE")

    result = faulty_device(query, ctx)

    assert result.status == "partial"
    assert result.errors == []


def test_empty_scope_does_not_validate_nodes():
    query = FaultQuery(node=[])

    def _no_read(*_args, **_kwargs):
        raise AssertionError("an empty scope must not read the Node list")

    ctx = make_ctx(query, cms={})
    ctx.gateway.list_nodes = _no_read
    result = faulty_device(query, ctx)

    assert result.status == "ok"
    assert result.data.node == {}


def test_node_level_fallback_does_not_require_device_coverage():
    """An uncovered in-scope node still publishes the node-level facts read from the node summary."""
    query = FaultQuery(node=["worker-a"])
    cms = supplement_cms(("worker-a",))
    cms[("mindx-dl", "cluster-info-node-cm")] = {
        "cluster-info-node-cm": json.dumps(
            {
                "mindx-dl-nodeinfo-worker-a": {
                    "NodeStatus": "Healthy",
                    "FaultDevList": [
                        {
                            "DeviceType": "PSU",
                            "DeviceId": "3",
                            "FaultCode": ["0300000D"],
                            "FaultLevel": "NotHandleFault",
                        }
                    ],
                }
            }
        )
    }
    result = faulty_device(query, make_ctx(query, cms=cms))

    # No NPU device facts at all: the answer stays partial, but the node-level fact is published.
    assert result.status == "partial"
    devices = result.data.node["worker-a"].devices
    assert [(device.device_id, device.device_type) for device in devices] == [("3", "psu")]


def test_health_device_without_fault_details_keeps_record(caplog):
    query = FaultQuery(node=["worker-a", "worker-b"])
    cms = supplement_cms(("worker-a", "worker-b"))
    cms[("mindx-dl", "cluster-info-device-0")] = {
        "cluster-info-device-0": json.dumps(
            {
                "mindx-dl-deviceinfo-worker-a": {
                    "DeviceList": pre_separate_device_list(),
                    "UpdateTime": 1718702470,
                },
                "mindx-dl-deviceinfo-worker-b": {
                    # Health list without any Fault object: the device is still abnormal.
                    "DeviceList": {NPU: "Ascend910-0", NPU + "-Unhealthy": "Ascend910-0"},
                    "UpdateTime": 1718702470,
                },
            }
        )
    }
    result = faulty_device(query, make_ctx(query, cms=cms))

    assert result.status == "ok"
    worker_b = [device for device in result.data.node["worker-b"].devices if device.device_id == "Ascend910-0"]
    assert len(worker_b) == 1
    assert worker_b[0].faults == []
    assert "query parse issue" in caplog.text


def test_complete_device_facts_have_no_pre_separation_handling():
    query = FaultQuery(node=["worker-a"])
    cms = supplement_cms(("worker-a",))
    cms[("mindx-dl", "cluster-info-device-0")] = {
        "cluster-info-device-0": json.dumps(
            {
                "mindx-dl-deviceinfo-worker-a": {
                    "DeviceList": cluster_device_list(),
                    "UpdateTime": 1718702470,
                }
            }
        )
    }
    result = faulty_device(query, make_ctx(query, cms=cms))

    assert not any(
        detail.fault_handling == "PreSeparateNPU"
        for device in result.data.node["worker-a"].devices
        for detail in device.faults
    )


def test_device_record_keeps_canonical_fault_handling():
    query = FaultQuery(node=["worker-a"])
    fault = {
        "fault_type": "CardUnhealthy",
        "npu_name": "Ascend910-1",
        "fault_handling": "SeparateNPU",
        "large_model_fault_level": "SubHealthFault",
        "fault_level": "RestartNPU",
        "fault_code": "00776711",
    }
    cms = supplement_cms(("worker-a",))
    cms[("mindx-dl", "cluster-info-device-0")] = {
        "cluster-info-device-0": json.dumps(
            {
                "mindx-dl-deviceinfo-worker-a": {
                    "DeviceList": {
                        NPU: "Ascend910-1",
                        NPU + "-Fault": json.dumps([fault]),
                    },
                    "UpdateTime": 1718702470,
                }
            }
        )
    }
    result = faulty_device(query, make_ctx(query, cms=cms))

    device = result.data.node["worker-a"].devices[0]
    assert device.faults[0].fault_handling == "SeparateNPU"
    assert set(device.model_dump()) == {"device_id", "device_type", "chip_type", "fault_type", "faults", "superpod"}
    assert set(device.faults[0].model_dump()) == {"fault_code", "fault_handling", "last_time"}
    assert result.status == "ok"


def test_faults_fall_back_to_dp_for_uncovered_node():
    query = FaultQuery(node=["worker-b"])
    cms = supplement_cms(("worker-b",))
    cms[("kube-system", "mindx-dl-deviceinfo-worker-b")] = {
        "DeviceInfoCfg": json.dumps(
            {
                "DeviceInfo": {"DeviceList": cluster_device_list(), "UpdateTime": 1718702470},
                "CheckCode": "x",
            }
        )
    }
    result = faulty_device(query, make_ctx(query, cms=cms))

    assert result.status == "ok"
    assert len(result.data.node["worker-b"].devices) == 1


def test_faults_without_any_source_is_partial():
    query = FaultQuery(node=["worker-a"])
    result = faulty_device(query, make_ctx(query, cms=supplement_cms()))

    assert result.status == "partial"
    assert result.errors == []
    assert result.data.node == {}


def _cluster_summary(node, update_time, device_list=None):
    entry = {"DeviceList": device_list or {NPU: "Ascend910-0,Ascend910-1"}, "UpdateTime": update_time}
    return {"cluster-info-device-0": json.dumps({f"mindx-dl-deviceinfo-{node}": entry})}


def test_producer_fault_time_is_exposed_as_public_last_time():
    query = FaultQuery(node=["worker-a"])
    cms = supplement_cms(("worker-a",))
    cms[("mindx-dl", "cluster-info-device-0")] = _cluster_summary("worker-a", 1718702470, cluster_device_list())
    result = faulty_device(query, make_ctx(query, cms=cms))

    assert result.status == "ok"
    assert result.data.node["worker-a"].devices[0].faults[0].last_time == "2024-06-18T09:21:10Z"


def test_primary_covered_node_is_not_reread_from_dp():
    query = FaultQuery(node=["worker-a"])
    cms = supplement_cms(("worker-a",))
    cms[("mindx-dl", "cluster-info-device-0")] = _cluster_summary("worker-a", 1718702470, cluster_device_list())
    result = faulty_device(query, make_ctx(query, cms=cms))

    assert result.status == "ok"
    assert [item.device_id for item in result.data.node["worker-a"].devices] == ["Ascend910-1"]


def test_dual_chip_prefixes_are_reported_per_chip_type():
    """Both card prefixes are read independently; each record carries its own chip_type."""
    npu_key = "huawei.com/npu"
    query = FaultQuery(node=["worker-a"])
    cms = supplement_cms(("worker-a",))
    device_list = {
        NPU: "Ascend910-0,Ascend910-1",
        NPU + "-Fault": json.dumps(
            [
                {
                    "fault_type": "CardUnhealthy",
                    "npu_name": "Ascend910-1",
                    "fault_handling": "SeparateNPU",
                    "fault_code": "00776711",
                }
            ]
        ),
        npu_key: "npu-0,npu-1,npu-2",
        npu_key + "-Fault": json.dumps(
            [
                {
                    "fault_type": "CardUnhealthy",
                    "npu_name": "npu-2",
                    "fault_handling": "SeparateNPU",
                    "fault_code": "00776711",
                }
            ]
        ),
        # Alias double-write: the same id published under the other prefix stays a second
        # record instead of being merged away.
        npu_key + "-Unhealthy": "npu-2,Ascend910-1",
    }
    cms[("mindx-dl", "cluster-info-device-0")] = _cluster_summary("worker-a", 1718702470, device_list)
    result = faulty_device(query, make_ctx(query, cms=cms))

    assert result.status == "ok"
    devices = result.data.node["worker-a"].devices
    chips: dict = {}
    for device in devices:
        chips.setdefault(device.chip_type, set()).add(device.device_id)
    assert chips == {"Ascend910": {"Ascend910-1"}, "npu": {"npu-2", "Ascend910-1"}}
    assert len([device for device in devices if device.device_id == "Ascend910-1"]) == 2


def test_dp_fallback_merges_per_prefix():
    """Prefixes the primary parsed keep their records; DP only fills the unparsable prefix."""
    npu_key = "huawei.com/npu"
    query = FaultQuery(node=["worker-a"])
    cms = supplement_cms(("worker-a",))
    device_list = {
        NPU: "Ascend910-0,Ascend910-1",
        NPU + "-Fault": json.dumps(
            [
                {
                    "fault_type": "CardUnhealthy",
                    "npu_name": "Ascend910-1",
                    "fault_handling": "SeparateNPU",
                    "fault_code": "00776711",
                }
            ]
        ),
        # The npu roster is structurally broken, so the primary does not cover the node.
        npu_key + "-Fault": "{}",
    }
    cms[("mindx-dl", "cluster-info-device-0")] = _cluster_summary("worker-a", 1718702470, device_list)
    cms[("kube-system", "mindx-dl-deviceinfo-worker-a")] = {
        "DeviceInfoCfg": json.dumps(
            {
                "DeviceInfo": {
                    "DeviceList": {
                        npu_key: "npu-0,npu-1",
                        npu_key + "-Fault": json.dumps(
                            [
                                {
                                    "fault_type": "CardUnhealthy",
                                    "npu_name": "npu-1",
                                    "fault_handling": "SeparateNPU",
                                    "fault_code": "00776711",
                                }
                            ]
                        ),
                        # DP also carries the prefix the primary already answered: not taken.
                        NPU: "Ascend910-0,Ascend910-1",
                        NPU + "-Fault": json.dumps(
                            [
                                {
                                    "fault_type": "CardUnhealthy",
                                    "npu_name": "Ascend910-0",
                                    "fault_handling": "SeparateNPU",
                                    "fault_code": "00776711",
                                }
                            ]
                        ),
                    },
                    "UpdateTime": 1718702470,
                }
            }
        )
    }
    result = faulty_device(query, make_ctx(query, cms=cms))

    assert result.status == "ok"
    devices = result.data.node["worker-a"].devices
    chips: dict = {}
    for device in devices:
        chips.setdefault(device.chip_type, set()).add(device.device_id)
    # Ascend910 keeps the primary's Ascend910-1 (not the DP's Ascend910-0); npu comes from DP.
    assert chips == {"Ascend910": {"Ascend910-1"}, "npu": {"npu-1"}}


def test_node_level_device_summary_record_has_no_chip_type():
    """Device-summary fault entries without npu_name are node-level facts: chip_type is null."""
    query = FaultQuery(node=["worker-a"])
    cms = supplement_cms(("worker-a",))
    device_list = {
        NPU: "Ascend910-0",
        NPU + "-Fault": json.dumps(
            [
                {"fault_type": "Node", "fault_code": "99", "fault_handling": "NotHandleFault"},
                {
                    "fault_type": "CardUnhealthy",
                    "npu_name": "Ascend910-0",
                    "fault_handling": "SeparateNPU",
                    "fault_code": "00776711",
                },
            ]
        ),
    }
    cms[("mindx-dl", "cluster-info-device-0")] = _cluster_summary("worker-a", 1718702470, device_list)
    result = faulty_device(query, make_ctx(query, cms=cms))

    assert result.status == "ok"
    devices = result.data.node["worker-a"].devices
    node_level = [device for device in devices if device.device_id == ""]
    assert len(node_level) == 1
    assert (node_level[0].device_type, node_level[0].chip_type) == ("node", None)
    assert node_level[0].faults[0].fault_code == "99"


def test_unusable_device_list_is_an_empty_set_with_a_logged_issue():
    """An unusable DeviceList is an empty roster: devices []、status ok，只记日志。"""
    for entry in (
        {"DeviceList": "huawei.com/Ascend910=Ascend910-0", "UpdateTime": 1718702470},
        {"UpdateTime": 1718702470},
    ):
        issues = []
        items = [
            {
                "name": "cluster-info-device-0",
                "data": {"cluster-info-device-0": json.dumps({"mindx-dl-deviceinfo-worker-a": entry})},
            }
        ]
        entries = parse_cluster_device_shards(items, issues, "s1")

        assert entries["worker-a"].device_list == {}
        assert [(issue.code, issue.path) for issue in issues] == [
            ("SOURCE_INVALID", "cluster-info-device-0.mindx-dl-deviceinfo-worker-a.DeviceList")
        ]

    query = FaultQuery(node=["worker-a"])
    cms = supplement_cms(("worker-a",))
    cms[("mindx-dl", "cluster-info-device-0")] = {
        "cluster-info-device-0": json.dumps(
            {"mindx-dl-deviceinfo-worker-a": {"DeviceList": None, "UpdateTime": 1718702470}}
        )
    }
    result = faulty_device(query, make_ctx(query, cms=cms))

    assert result.status == "ok"
    assert result.data.node["worker-a"].devices == []

    # cluster_resource reads the same facts: an unusable roster is a clean zero, not a gap.
    resource_query = ResourceQuery(node=["worker-a"])
    node = {
        "metadata": {"name": "worker-a"},
        "status": {"capacity": {"cpu": "16", NPU: "8"}, "allocatable": {"cpu": "16", NPU: "8"}},
    }
    resource = cluster_resource(resource_query, make_ctx(resource_query, cms=cms, nodes=[node]))

    assert resource.status == "ok"
    assert resource.data.node["worker-a"].buckets()["Ascend910"].fault == 0


def test_isolated_devices_total_with_detail_is_reported():
    query = IsolationQuery()
    cms = {
        ("cluster-system", "clusterd-manual-info-cm"): {
            "worker-a": json.dumps(
                {
                    # A duplicate roster name is one record; members are name-sorted.
                    "Total": ["Ascend910-4", "Ascend910-1", "Ascend910-4"],
                    "Detail": {
                        "Ascend910-1": [
                            {
                                "FaultCode": "00776711",
                                "FaultLevel": "ManuallySeparateNPU",
                                "LastSeparateTime": 1718702470000,
                            }
                        ],
                        # Several detail entries aggregate: codes dedup, time keeps the latest.
                        "Ascend910-4": [
                            {"FaultCode": "8", "FaultLevel": "SilentFault", "LastSeparateTime": 1750000100000},
                            {"FaultCode": "8", "FaultLevel": "SilentFault", "LastSeparateTime": 1750000200000},
                        ],
                    },
                }
            )
        }
    }
    result = isolated_device(query, make_ctx(query, cms=cms))

    assert result.status == "ok"
    devices = result.data.node["worker-a"].devices
    assert [device.device_id for device in devices] == ["Ascend910-1", "Ascend910-4"]
    assert devices[0].fault_level == "ManuallySeparateNPU"
    assert devices[0].fault_code == "00776711"
    assert devices[0].last_separate_time == "2024-06-18T09:21:10Z"
    assert devices[1].fault_level == "SilentFault"
    assert devices[1].fault_code == "8"
    assert devices[1].last_separate_time == "2025-06-15T15:10:00Z"


def test_isolated_devices_ignore_dp_manually_separate():
    """ClusterD Total is the only roster; DP observations never add members."""
    query = IsolationQuery()
    cms = {
        ("cluster-system", "clusterd-manual-info-cm"): {
            "worker-a": json.dumps({"Total": ["Ascend910-1"], "Detail": {}})
        },
        ("kube-system", "mindx-dl-deviceinfo-worker-a"): {"ManuallySeparateNPU": "Ascend910-2"},
    }
    result = isolated_device(query, make_ctx(query, cms=cms))

    assert result.status == "ok"
    devices = result.data.node["worker-a"].devices
    assert [device.device_id for device in devices] == ["Ascend910-1"]
    assert devices[0].fault_level == "unknown"
    assert devices[0].fault_code is None
    assert devices[0].last_separate_time is None


def test_isolated_devices_missing_manual_cm_is_ok_when_nothing_isolated():
    query = IsolationQuery()
    cms = {("kube-system", "mindx-dl-deviceinfo-worker-a"): {"ManuallySeparateNPU": "Ascend910-2"}}
    result = isolated_device(query, make_ctx(query, cms=cms))

    assert result.status == "ok"
    assert result.data.node == {}


def test_isolated_devices_target_without_cm_key_is_ok_and_empty():
    """A target node that exists but is absent from the CM is a legitimate empty roster."""
    query = IsolationQuery(node=["worker-missing"])
    cms = {
        ("cluster-system", "clusterd-manual-info-cm"): {
            "worker-a": json.dumps({"Total": ["Ascend910-1"], "Detail": {}})
        }
    }
    result = isolated_device(query, make_ctx(query, cms=cms, nodes=[{"metadata": {"name": "worker-missing"}}]))

    assert result.status == "ok"
    assert result.data.node == {}


def test_isolated_devices_empty_scope_returns_empty_without_reading():
    query = IsolationQuery(node=[])

    def _no_read(*_args, **_kwargs):
        raise AssertionError("the isolation CM must not be read for an empty scope")

    ctx = make_ctx(query, cms={})
    ctx.gateway.get_cm = _no_read
    result = isolated_device(query, ctx)

    assert result.status == "ok"
    assert result.data.node == {}


def test_isolated_devices_cm_read_failure_is_partial():
    query = IsolationQuery()
    ctx = make_ctx(query, cms={})
    ctx.gateway.get_cm = lambda *_args, **_kwargs: SourceReadResult(
        status="forbidden", source_id=ctx.next_source_id(), error_code="SOURCE_FORBIDDEN"
    )
    result = isolated_device(query, ctx)

    # Even with nothing read at all, a finished query stays partial, never error.
    assert result.status == "partial"
    assert result.data.node == {}


def test_isolated_devices_broken_node_entry_is_partial_and_keeps_valid_nodes():
    query = IsolationQuery()
    cms = {
        ("cluster-system", "clusterd-manual-info-cm"): {
            "worker-a": json.dumps({"Total": ["Ascend910-1"], "Detail": {}}),
            "worker-b": "not-json{",
        }
    }
    result = isolated_device(query, make_ctx(query, cms=cms))

    assert result.status == "partial"
    assert [device.device_id for device in result.data.node["worker-a"].devices] == ["Ascend910-1"]
    assert "worker-b" not in result.data.node


def test_isolated_devices_missing_total_roster_is_partial():
    """A missing Total is never rebuilt from Detail; the target stays uncovered."""
    query = IsolationQuery()
    cms = {
        ("cluster-system", "clusterd-manual-info-cm"): {
            "worker-a": json.dumps({"Detail": {"Ascend910-1": [{"FaultLevel": "ManuallySeparateNPU"}]}}),
        }
    }
    result = isolated_device(query, make_ctx(query, cms=cms))

    assert result.status == "partial"
    assert result.data.node == {}


def test_isolated_devices_invalid_total_member_is_partial_and_keeps_valid_members():
    query = IsolationQuery()
    cms = {
        ("cluster-system", "clusterd-manual-info-cm"): {
            "worker-a": json.dumps({"Total": ["Ascend910-1", ""], "Detail": {}})
        }
    }
    result = isolated_device(query, make_ctx(query, cms=cms))

    assert result.status == "partial"
    assert [device.device_id for device in result.data.node["worker-a"].devices] == ["Ascend910-1"]


def test_manual_total_detail_mismatch_ignores_detail_only_members(caplog):
    """Total-only members stay with unknown facts; Detail-only members are dropped."""
    query = IsolationQuery()
    cms = {
        ("cluster-system", "clusterd-manual-info-cm"): {
            "worker-a": json.dumps(
                {
                    "Total": ["Ascend910-1"],
                    "Detail": {"Ascend910-2": [{"FaultCode": "00776711", "FaultLevel": "ManuallySeparateNPU"}]},
                }
            )
        }
    }
    result = isolated_device(query, make_ctx(query, cms=cms))

    # The roster itself is complete, so detail mismatches never downgrade the status.
    assert result.status == "ok"
    devices = result.data.node["worker-a"].devices
    assert [device.device_id for device in devices] == ["Ascend910-1"]
    assert devices[0].fault_level == "unknown"
    assert devices[0].fault_code is None
    assert devices[0].last_separate_time is None
    assert "query parse issue" in caplog.text


def test_isolated_devices_second_magnitude_time_falls_back(caplog):
    """LastSeparateTime is contracted in ms; a seconds-magnitude value still renders."""
    query = IsolationQuery()
    cms = {
        ("cluster-system", "clusterd-manual-info-cm"): {
            "worker-a": json.dumps(
                {
                    "Total": ["Ascend910-1"],
                    "Detail": {
                        "Ascend910-1": [
                            {
                                "FaultCode": "00776711",
                                "FaultLevel": "ManuallySeparateNPU",
                                "LastSeparateTime": 1718702470,
                            }
                        ]
                    },
                }
            )
        }
    }
    result = isolated_device(query, make_ctx(query, cms=cms))

    assert result.status == "ok"
    device = result.data.node["worker-a"].devices[0]
    assert device.last_separate_time == "2024-06-18T09:21:10Z"
    assert "documented contract is ms" in caplog.text


def test_resource_stats_empty_cluster_is_a_complete_answer():
    query = ResourceQuery()
    result = cluster_resource(query, make_ctx(query, cms=supplement_cms(), nodes=[], pods=[]))

    assert result.status == "ok"
    assert result.data.node == {}


def test_resource_stats_six_metrics():
    query = ResourceQuery()
    nodes = [
        {
            "metadata": {"name": "worker-a"},
            "status": {
                "capacity": {"cpu": "16", NPU: "8", NPU2: "4"},
                "allocatable": {"cpu": "15", NPU: "6", NPU2: "3"},
            },
        }
    ]
    pods = [
        {
            "metadata": {"name": "pod-a", "namespace": "default"},
            "spec": {"nodeName": "worker-a", "containers": [{"resources": {"requests": {NPU: "2"}}}]},
            "status": {"phase": "Running"},
        },
        {
            "metadata": {"name": "pod-b", "namespace": "default"},
            "spec": {"nodeName": "worker-a", "containers": [{"resources": {"requests": {NPU: "4"}}}]},
            "status": {"phase": "Pending"},
        },
    ]
    result = cluster_resource(query, make_ctx(query, cms=resource_stats_cms(), nodes=nodes, pods=pods))

    assert result.status == "ok"
    node = result.data.node["worker-a"]
    assert (node.cpu.capacity_cores, node.cpu.allocatable_cores, node.cpu.used_cores, node.cpu.idle_cores) == (
        16.0,
        15.0,
        0.0,
        15.0,
    )
    # Pending requests never count; the faulted card that is also isolated counts once in
    # each field (fault never subtracts isolation).
    ascend = node.buckets()["Ascend910"]
    assert (ascend.capacity, ascend.allocatable, ascend.used, ascend.idle) == (8, 6, 2, 4)
    assert (ascend.fault, ascend.isolated) == (2, 1)
    # The second whole-card prefix is an independent bucket, never merged into Ascend910.
    npu = node.buckets()["npu"]
    assert (npu.capacity, npu.allocatable, npu.used, npu.idle) == (4, 3, 0, 3)
    assert (npu.fault, npu.isolated) == (0, 0)


def test_resource_stats_alias_device_counts_once_per_bucket():
    """An isolated id written under both prefixes counts once per bucket, never merged."""
    query = ResourceQuery()
    cms = supplement_cms()
    cms[("mindx-dl", "cluster-info-device-0")] = {
        "cluster-info-device-0": json.dumps(
            {
                "mindx-dl-deviceinfo-worker-a": {
                    "DeviceList": {
                        NPU: "Ascend910-0,Ascend910-2",
                        NPU + "-Fault": json.dumps(
                            [
                                {
                                    "npu_name": "Ascend910-0",
                                    "fault_type": "CardUnhealthy",
                                    "fault_handling": "SeparateNPU",
                                    "fault_code": "8",
                                }
                            ]
                        ),
                        NPU2: "Ascend910-2,npu-0",
                    },
                    "UpdateTime": 1718702470,
                }
            }
        )
    }
    cms[("cluster-system", "clusterd-manual-info-cm")] = {
        "worker-a": json.dumps(
            {
                "Total": ["Ascend910-2"],
                "Detail": {"Ascend910-2": [{"FaultLevel": "ManuallySeparateNPU", "FaultCode": "8"}]},
            }
        )
    }
    nodes = [
        {
            "metadata": {"name": "worker-a"},
            "status": {
                "capacity": {"cpu": "8", NPU: "2", NPU2: "2"},
                "allocatable": {"cpu": "8", NPU: "2", NPU2: "2"},
            },
        }
    ]
    result = cluster_resource(query, make_ctx(query, cms=cms, nodes=nodes, pods=[]))

    assert result.status == "ok"
    resources = result.data.node["worker-a"].buckets()
    assert (resources["Ascend910"].fault, resources["Ascend910"].isolated) == (1, 1)
    assert (resources["npu"].fault, resources["npu"].isolated) == (0, 1)


def test_resource_stats_unattributable_isolation_nulls_only_isolated(caplog):
    """An isolated name outside every roster nulls isolated; fault stays independent."""
    query = ResourceQuery()
    cms = supplement_cms()
    cms[("mindx-dl", "cluster-info-device-0")] = {
        "cluster-info-device-0": json.dumps(
            {
                "mindx-dl-deviceinfo-worker-a": {
                    "DeviceList": {NPU: "Ascend910-0"},
                    "UpdateTime": 1718702470,
                }
            }
        )
    }
    cms[("cluster-system", "clusterd-manual-info-cm")] = {
        "worker-a": json.dumps(
            {
                "Total": ["Ascend910-9"],
                "Detail": {"Ascend910-9": [{"FaultLevel": "ManuallySeparateNPU", "FaultCode": "8"}]},
            }
        )
    }
    nodes = [
        {
            "metadata": {"name": "worker-a"},
            "status": {"capacity": {"cpu": "8", NPU: "8"}, "allocatable": {"cpu": "8", NPU: "8"}},
        }
    ]
    result = cluster_resource(query, make_ctx(query, cms=cms, nodes=nodes, pods=[]))

    bucket = result.data.node["worker-a"].buckets()["Ascend910"]
    assert result.status == "ok"  # an evidence gap never changes the envelope status
    assert bucket.fault == 0  # the device facts are readable and carry no fault identity
    assert bucket.isolated is None
    assert "query parse issue" in caplog.text


def test_resource_stats_isolated_card_moved_out_of_roster():
    """A genuinely isolated card leaves the allocatable roster; it is still attributed."""
    query = ResourceQuery()
    cms = supplement_cms()
    cms[("mindx-dl", "cluster-info-device-0")] = {
        "cluster-info-device-0": json.dumps(
            {
                "mindx-dl-deviceinfo-worker-a": {
                    "DeviceList": {
                        NPU: "Ascend910-0,Ascend910-1",  # card 2 has been moved out
                        NPU + "-Unhealthy": "Ascend910-2",
                        NPU + "-Fault": json.dumps(
                            [
                                {
                                    "npu_name": "Ascend910-2",
                                    "fault_type": "CardUnhealthy",
                                    "fault_handling": "ManuallySeparateNPU",
                                    "fault_code": "",
                                }
                            ]
                        ),
                    },
                    "UpdateTime": 1718702470,
                }
            }
        )
    }
    cms[("cluster-system", "clusterd-manual-info-cm")] = {
        "worker-a": json.dumps(
            {
                "Total": ["Ascend910-2"],
                "Detail": {"Ascend910-2": [{"FaultLevel": "ManuallySeparateNPU", "FaultCode": "ManuallySeparateNPU"}]},
            }
        )
    }
    nodes = [
        {
            "metadata": {"name": "worker-a"},
            "status": {"capacity": {"cpu": "8", NPU: "8"}, "allocatable": {"cpu": "8", NPU: "8"}},
        }
    ]
    result = cluster_resource(query, make_ctx(query, cms=cms, nodes=nodes, pods=[]))

    bucket = result.data.node["worker-a"].buckets()["Ascend910"]
    assert result.status == "ok"
    assert bucket.fault == 1
    assert bucket.isolated == 1  # was None while the roster only listed allocatable devices


def test_resource_stats_missing_device_facts_null_fault_but_publish_isolated():
    query = ResourceQuery()
    nodes = [
        {
            "metadata": {"name": "worker-a"},
            "status": {"capacity": {"cpu": "8", NPU: "8"}, "allocatable": {"cpu": "8", NPU: "8"}},
        }
    ]
    result = cluster_resource(query, make_ctx(query, cms=supplement_cms(), nodes=nodes, pods=[]))

    bucket = result.data.node["worker-a"].buckets()["Ascend910"]
    assert result.status == "ok"
    assert (bucket.capacity, bucket.allocatable, bucket.used, bucket.idle) == (8, 8, 0, 8)
    assert bucket.fault is None  # the fault device source was not readable
    assert bucket.isolated == 0  # the isolation source is readable and lists nobody


def _review_node(name="worker-a", allocatable="7"):
    status = {"capacity": {"cpu": "16", NPU: "8"}, "allocatable": {"cpu": "16"}}
    if allocatable is not None:
        status["allocatable"][NPU] = allocatable
    return {"metadata": {"name": name}, "status": status}


def test_bad_fault_json_is_partial_and_no_device_facts():
    query = FaultQuery(node=["worker-a"])
    cms = supplement_cms(("worker-a",))
    cms[("mindx-dl", "cluster-info-device-0")] = {
        "cluster-info-device-0": json.dumps(
            {"mindx-dl-deviceinfo-worker-a": {"DeviceList": {NPU: "Ascend910-0", NPU + "-Fault": "not-json"}}}
        )
    }
    result = faulty_device(query, make_ctx(query, cms=cms))

    assert result.status == "partial"
    assert result.data.node["worker-a"].devices == []


def test_decode_json_string_empty_requires_explicit_allowance():
    issues = []
    assert decode_json_string("", "path", issues) is None
    assert [issue.code for issue in issues] == ["SOURCE_INVALID"]

    issues = []
    assert decode_json_string("   ", "path", issues, allow_empty=True) is None
    assert not issues

    issues = []
    assert decode_json_string("not-json", "path", issues, allow_empty=True) is None
    assert [issue.code for issue in issues] == ["SOURCE_INVALID"]


def test_empty_clusterd_shards_are_not_reported_as_corrupt():
    """ClusterD clears an empty/unused shard to ""; that is valid 'no data', not bad JSON."""
    query = FaultQuery()
    cms = supplement_cms()
    cms[("mindx-dl", "cluster-info-device-0")] = {"cluster-info-device-0": ""}
    cms[("mindx-dl", "cluster-info-device-1")] = {
        "cluster-info-device-1": json.dumps({"mindx-dl-deviceinfo-worker-a": {"DeviceList": cluster_device_list()}})
    }
    cms[("mindx-dl", "cluster-info-node-cm")] = {"cluster-info-node-cm": ""}
    cms[("mindx-dl", "cluster-info-switch-0")] = {"cluster-info-switch-0": " "}
    cms[("mindx-dl", "cluster-info-dpu-0")] = {"cluster-info-dpu-0": ""}
    result = faulty_device(query, make_ctx(query, cms=cms))

    assert result.status == "ok"
    assert [device.device_id for device in result.data.node["worker-a"].devices] == ["Ascend910-1"]


def test_public_fault_empty_string_still_reports_invalid():
    """The allowance is source-specific; public-fault truncation must remain visible."""
    issues = []
    records = parse_public_faults({"PublicFaults": "", "FaultNum": ""}, None, "s1", issues)

    assert records == {}
    assert [issue.code for issue in issues] == ["SOURCE_INVALID"]


def test_full_cluster_reads_dp_for_nodes_missing_from_summary():
    query = FaultQuery()
    cms = supplement_cms()
    cms[("mindx-dl", "cluster-info-device-0")] = {
        "cluster-info-device-0": json.dumps(
            {"mindx-dl-deviceinfo-worker-a": {"DeviceList": {NPU: "Ascend910-0", NPU + "-Fault": "[]"}}}
        )
    }
    cms[("kube-system", "mindx-dl-deviceinfo-worker-b")] = {
        "DeviceInfoCfg": json.dumps(
            {
                "DeviceInfo": {
                    "DeviceList": {
                        NPU: "Ascend910-1",
                        NPU + "-Fault": json.dumps(
                            [
                                {
                                    "npu_name": "Ascend910-1",
                                    "fault_type": "CardUnhealthy",
                                    "fault_handling": "SeparateNPU",
                                    "fault_code": "123",
                                }
                            ]
                        ),
                    }
                }
            }
        )
    }
    result = faulty_device(query, make_ctx(query, cms=cms, nodes=[_review_node(), _review_node("worker-b")]))

    assert result.data.node["worker-b"].devices[0].device_id == "Ascend910-1"


def test_missing_allocatable_is_unknown_not_zero(caplog):
    query = ResourceQuery()
    result = cluster_resource(query, make_ctx(query, nodes=[_review_node(allocatable=None)]))
    bucket = result.data.node["worker-a"].buckets()["Ascend910"]

    assert result.status == "partial"
    assert "query parse issue" in caplog.text
    assert bucket.capacity == 8
    assert bucket.allocatable is None
    assert bucket.idle is None


def test_missing_capacity_is_unknown_not_zero(caplog):
    query = ResourceQuery()
    node = {
        "metadata": {"name": "worker-a"},
        "status": {"capacity": {"cpu": "16"}, "allocatable": {"cpu": "16", NPU: "8"}},
    }
    result = cluster_resource(query, make_ctx(query, nodes=[node]))
    bucket = result.data.node["worker-a"].buckets()["Ascend910"]

    assert result.status == "partial"
    assert bucket.capacity is None
    assert bucket.allocatable == 8
    assert bucket.idle == 8
    assert "query parse issue" in caplog.text


def test_negative_capacity_difference_is_flagged(caplog):
    query = ResourceQuery()
    node = {
        "metadata": {"name": "worker-a"},
        "status": {"capacity": {"cpu": "16", NPU: "0"}, "allocatable": {"cpu": "16", NPU: "8"}},
    }
    result = cluster_resource(query, make_ctx(query, nodes=[node]))

    assert result.status == "partial"
    assert result.data.node["worker-a"].buckets()["Ascend910"].capacity == 0
    assert "query parse issue" in caplog.text


def test_unsupported_npu_resource_is_excluded_and_flagged(caplog):
    query = ResourceQuery()
    virtual = "huawei.com/Ascend910-virtual"
    node = {
        "metadata": {"name": "worker-a"},
        "status": {
            "capacity": {"cpu": "16", NPU: "8", virtual: "4"},
            "allocatable": {"cpu": "16", NPU: "8", virtual: "4"},
        },
    }
    result = cluster_resource(query, make_ctx(query, nodes=[node]))
    bucket = result.data.node["worker-a"].buckets()["Ascend910"]

    assert bucket.capacity == 8
    assert "query parse issue" in caplog.text  # the virtual key is logged and excluded
    assert result.status == "ok"  # a recognized bucket still answers the question

    # An explicit target carrying only unknown card keys has no answer at all.
    only_virtual = {
        "metadata": {"name": "worker-a"},
        "status": {"capacity": {virtual: "4"}, "allocatable": {virtual: "4"}},
    }
    query = ResourceQuery(node=["worker-a"])
    result = cluster_resource(query, make_ctx(query, nodes=[only_virtual]))
    assert result.status == "partial"
    assert result.data.node == {}


def test_node_health_is_attached_to_each_node():
    cms = {
        ("mindx-dl", "cluster-info-device-0"): {
            "cluster-info-device-0": json.dumps(
                {
                    "mindx-dl-deviceinfo-worker-a": {
                        "DeviceList": cluster_device_list(),
                        "UpdateTime": 1718702470,
                    },
                    "mindx-dl-deviceinfo-worker-b": {
                        "DeviceList": {NPU: ""},
                        "UpdateTime": 1718702470,
                    },
                }
            )
        },
        ("mindx-dl", "cluster-info-node-cm"): {
            "cluster-info-node-cm": json.dumps(
                {
                    "mindx-dl-nodeinfo-worker-a": {
                        "NodeStatus": "UnHealthy",
                        "FaultDevList": [
                            {
                                "DeviceType": "PSU",
                                "DeviceId": 0,
                                "FaultCode": ["99"],
                                "FaultLevel": "NotHandleFault",
                            }
                        ],
                        "UpdateTime": 1718702470000,
                    },
                    "mindx-dl-nodeinfo-worker-b": {
                        "NodeStatus": "UnHealthy",
                        "FaultDevList": [],
                        "UpdateTime": 1718702470000,
                    },
                }
            )
        },
    }
    cms.update({key: value for key, value in supplement_cms().items() if key not in cms})
    nodes = [
        {"metadata": {"name": "worker-a"}, "status": {"capacity": {NPU: "8"}, "allocatable": {NPU: "8"}}},
        {"metadata": {"name": "worker-b"}, "status": {"capacity": {NPU: "8"}, "allocatable": {NPU: "8"}}},
    ]
    query = FaultQuery()
    result = faulty_device(query, make_ctx(query, cms=cms, nodes=nodes))

    assert result.status == "ok"
    # worker-b has NodeStatus=UnHealthy but an empty FaultDevList: health must still be attached.
    assert result.data.node["worker-a"].node_health.node_status == "UnHealthy"
    assert result.data.node["worker-b"].node_health.node_status == "UnHealthy"
    assert list(result.data.node["worker-a"].model_dump())[-1] == "node_health"
    psu = [device for device in result.data.node["worker-a"].devices if device.device_type == "psu"]
    assert len(psu) == 1
    assert (psu[0].device_id, psu[0].fault_type, psu[0].chip_type) == ("0", "PSU", None)
    assert psu[0].faults[0].fault_code == "99"
    assert any(device.device_id == "Ascend910-1" for device in result.data.node["worker-a"].devices)


def test_faulty_device_includes_supplementary_sources():
    cms = {
        ("mindx-dl", "cluster-info-device-0"): {
            "cluster-info-device-0": json.dumps(
                {
                    "mindx-dl-deviceinfo-worker-a": {
                        "DeviceList": {NPU: "Ascend910-0"},
                        "UpdateTime": 1718702470,
                    }
                }
            )
        },
        ("mindx-dl", "cluster-info-node-cm"): {
            "cluster-info-node-cm": json.dumps({"mindx-dl-nodeinfo-worker-a": {"NodeStatus": "Healthy"}})
        },
        ("mindx-dl", "cluster-info-switch-0"): {
            "cluster-info-switch-0": json.dumps(
                {
                    "worker-a": {
                        "CmName": "worker-a",
                        "FaultLevel": "Warning",
                        "FaultCode": [
                            json.dumps(
                                {
                                    "AssembledFaultCode": "switch-active",
                                    "SwitchChipId": "0",
                                    "SwitchPortId": "1",
                                    "Assertion": 1,
                                    "AlarmRaisedTime": 1718702470000,
                                }
                            ),
                            json.dumps({"AssembledFaultCode": "switch-cleared", "Assertion": 0}),
                        ],
                    }
                }
            )
        },
        ("mindx-dl", "cluster-info-dpu-0"): {
            "cluster-info-dpu-0": json.dumps(
                {
                    "dpuinfo-worker-a": {
                        "DPUInfo": {
                            "DPUList": [
                                {
                                    "DeviceID": "0",
                                    "HcaName": "hca0",
                                    "AffectedNPU": ["3"],
                                    "FaultList": [
                                        {
                                            "FaultCode": "dpu-link",
                                            "FaultLevel": "Warning",
                                            "Time": 1718702470000,
                                        }
                                    ],
                                }
                            ],
                            "NodeEvent": None,
                        }
                    },
                    "UpdateTime": 1718702470000,
                }
            )
        },
        ("cluster-system", "statistic-fault-info"): {
            "PublicFaults": json.dumps(
                {
                    "worker-a": [
                        {
                            "resource": "npu",
                            "devIds": ["5"],
                            "type": "public",
                            "faultCode": "public-code",
                            "level": "Warning",
                            "faultTime": 1718702470,
                        }
                    ]
                }
            ),
            "FaultNum": json.dumps({"publicFaultNum": 1}),
        },
    }
    nodes = [{"metadata": {"name": "worker-a"}, "status": {"capacity": {NPU: "8"}}}]
    query = FaultQuery()
    result = faulty_device(query, make_ctx(query, cms=cms, nodes=nodes))

    assert result.status == "ok"
    node = result.data.node["worker-a"]
    switch = {record.fault_code: record for record in node.switch}
    assert switch["switch-active"].device_id == "switch:0:1"
    # The ClusterD switch summary has no fault-type dimension, so the record must not
    # carry the key at all (null would imply "known field, unknown value").
    assert "fault_type" not in switch["switch-active"].model_dump()
    assert "switch-cleared" not in switch
    assert node.dpu[0].device_id == "dpu:0"
    assert "fault_type" not in node.dpu[0].model_dump()
    assert node.dpu[0].affected_npu == ["3"]
    assert node.public_faults[0].device_id == "5"
    assert node.public_faults[0].fault_type == "PublicFault"
    assert node.public_faults[0].last_time == "2024-06-18T09:21:10Z"
    assert node.node_health.node_status == "Healthy"
    assert node.devices == []


def test_query_data_builders_emit_only_public_fields():
    from agent_core.query_schemas import (
        CpuResource,
        IsolatedDeviceRecord,
        NodeIsolation,
        NodeResourceData,
        ResourceBucket,
    )
    from agent_core.tools.data_builders import (
        build_isolation_data,
        build_resource_data,
    )

    isolation_nodes = {
        "worker-b": NodeIsolation(devices=[]),
        "worker-a": NodeIsolation(
            devices=[
                IsolatedDeviceRecord(
                    device_id="Ascend910-6",
                    fault_level="ManuallySeparateNPU",
                    fault_code="81078603",
                    last_separate_time="2025-06-15T15:10:00Z",
                )
            ]
        ),
    }
    isolation_data = build_isolation_data(isolation_nodes)
    assert list(isolation_data.node) == ["worker-a", "worker-b"]
    record = isolation_data.node["worker-a"].devices[0]
    assert record.device_id == "Ascend910-6"
    assert set(record.model_dump()) == {"device_id", "fault_level", "fault_code", "last_separate_time"}

    bucket = ResourceBucket(capacity=8, allocatable=8, used=1, idle=7, fault=0, isolated=0)
    node_data = NodeResourceData(
        cpu=CpuResource(capacity_cores=16.0, allocatable_cores=16.0, used_cores=0.5, idle_cores=15.5),
        Ascend910=bucket,
    )
    resource_data = build_resource_data({"worker-b": NodeResourceData(), "worker-a": node_data})
    assert list(resource_data.node) == ["worker-a", "worker-b"]
    assert resource_data.node["worker-a"].cpu.model_dump() == {
        "capacity_cores": 16.0,
        "allocatable_cores": 16.0,
        "used_cores": 0.5,
        "idle_cores": 15.5,
    }
    # cpu and the whole-card bucket are sibling keys of the node object (no wrapper).
    assert resource_data.node["worker-a"].model_dump() == {
        "cpu": {
            "capacity_cores": 16.0,
            "allocatable_cores": 16.0,
            "used_cores": 0.5,
            "idle_cores": 15.5,
        },
        "Ascend910": {
            "capacity": 8,
            "allocatable": 8,
            "used": 1,
            "idle": 7,
            "fault": 0,
            "isolated": 0,
        },
    }
