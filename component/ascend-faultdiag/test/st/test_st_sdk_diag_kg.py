#!/usr/bin/env python3
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
import unittest

from ascend_fd import diag_knowledge_graph, parse_knowledge_graph

PLOG_ERR_LINE = (
    "[ERROR] RUNTIME(27514,python):2024-03-16-15:52:17.572.654 [task_info.cc:5261]28555 "
    "PrintErrorInfoForNotifyWaitTask:[COMP][DEFAULT]Notify wait execute failed, device_id=1"
)
CHAIN_SOURCE_CODE = "Comp_Network_Custom_01"
CHAIN_DST_CODE = "AISW_CANN_HCCL_027"
OCCUR_TIME = "2024-04-08 16:24:48"


def build_ccae_fault(event_code=CHAIN_SOURCE_CODE, source_device="0", **kwargs):
    fault = {
        "event_code": event_code,
        "key_info": "",
        "source_file": "train.log",
        "source_device": source_device,
        "occur_time": OCCUR_TIME,
    }
    fault.update(kwargs)
    return fault


def build_ccae_input(server="node-1", faults=None, **kwargs):
    payload = {"server": server, "source": "CCAE", "fault": [build_ccae_fault()] if faults is None else faults}
    payload.update(kwargs)
    return payload


def build_kg_analyzer():
    results, _ = parse_knowledge_graph(
        [{"log_domain": {"server": "x.x.x.x"}, "log_items": [{"item_type": "CANN_Plog", "log_lines": [PLOG_ERR_LINE]}]}]
    )
    return results[0]["fault"][0]


class STDiagKnowledgeGraph(unittest.TestCase):
    @staticmethod
    def get_codes(result):
        return [fault.get("code") for fault in result.get("fault", [])]

    def assert_diag_success(self, payload, expected_results=1):
        results, err_msg_list = diag_knowledge_graph(payload)
        self.assertEqual(err_msg_list, [])
        self.assertEqual(len(results), expected_results)
        for result in results:
            self.assertTrue(result.get("analyze_success"))
            self.assertTrue(result.get("fault"))
        return results

    def assert_diag_failed(self, payload, expected):
        results, err_msg_list = diag_knowledge_graph(payload)
        self.assertEqual(results, [])
        self.assertTrue(err_msg_list)
        self.assertIn(expected, err_msg_list[0])

    def test_func_0065_source_not_ccae_with_multiple_fault_types(self):
        """Ascend_Fd_SDK_Func_0065"""
        results = self.assert_diag_success([{"server": "node-1", "source": "OTHER", "fault": [build_kg_analyzer()]}])
        self.assertIn("AISW_CANN_Runtime_021", self.get_codes(results[0]))

    def test_func_0020_ccae_same_device_different_event_code(self):
        """Ascend_Fd_SDK_Func_0020"""
        faults = [build_ccae_fault(CHAIN_SOURCE_CODE, "0"), build_ccae_fault(CHAIN_DST_CODE, "0")]
        results = self.assert_diag_success([build_ccae_input(faults=faults)])
        self.assertIn(CHAIN_SOURCE_CODE, self.get_codes(results[0]))
        chains = results[0]["fault"][0].get("fault_chains", [])
        self.assertTrue(any(CHAIN_DST_CODE in str(chain) for chain in chains))

    def test_func_0025_ccae_same_event_code_different_devices(self):
        """Ascend_Fd_SDK_Func_0025"""
        faults = [build_ccae_fault(CHAIN_SOURCE_CODE, "0"), build_ccae_fault(CHAIN_SOURCE_CODE, "1")]
        results = self.assert_diag_success([build_ccae_input(faults=faults)])
        self.assertIn(CHAIN_SOURCE_CODE, self.get_codes(results[0]))

    def test_func_0027_ccae_multiple_servers_same_event_code(self):
        """Ascend_Fd_SDK_Func_0027"""
        payload = [
            build_ccae_input(server="node-1"),
            build_ccae_input(server="node-2"),
        ]
        self.assert_diag_success(payload, expected_results=2)

    def test_func_0063_ccae_multiple_fault_types_different_devices(self):
        """Ascend_Fd_SDK_Func_0063"""
        faults = [build_ccae_fault(CHAIN_SOURCE_CODE, "0"), build_ccae_fault(CHAIN_DST_CODE, "1")]
        results = self.assert_diag_success([build_ccae_input(faults=faults)])
        self.assertIn(CHAIN_SOURCE_CODE, self.get_codes(results[0]))

    def test_func_0064_ccae_multiple_fault_types_same_device(self):
        """Ascend_Fd_SDK_Func_0064"""
        faults = [build_ccae_fault(CHAIN_SOURCE_CODE, "0"), build_ccae_fault(CHAIN_DST_CODE, "0")]
        results = self.assert_diag_success([build_ccae_input(faults=faults)])
        self.assertIn(CHAIN_SOURCE_CODE, self.get_codes(results[0]))

    def test_func_0024_ccae_same_server_same_device_chain(self):
        """Ascend_Fd_SDK_Func_0024"""
        faults = [build_ccae_fault(CHAIN_SOURCE_CODE, "0"), build_ccae_fault(CHAIN_DST_CODE, "0")]
        results = self.assert_diag_success([build_ccae_input(faults=faults)])
        chains = results[0]["fault"][0].get("fault_chains", [])
        self.assertTrue(any(CHAIN_DST_CODE in str(chain) for chain in chains))

    def test_func_0021_ccae_same_server_multiple_devices_chain(self):
        """Ascend_Fd_SDK_Func_0021"""
        faults = [build_ccae_fault(CHAIN_SOURCE_CODE, "0"), build_ccae_fault(CHAIN_DST_CODE, "1")]
        results = self.assert_diag_success([build_ccae_input(faults=faults)])
        self.assertIn(CHAIN_SOURCE_CODE, self.get_codes(results[0]))

    def test_func_0022_source_not_ccae_multiple_servers_same_event_code(self):
        """Ascend_Fd_SDK_Func_0022"""
        analyzer = build_kg_analyzer()
        payload = [
            {"server": "node-1", "source": "OTHER", "fault": [analyzer]},
            {"server": "node-2", "source": "OTHER", "fault": [analyzer]},
        ]
        self.assert_diag_success(payload, expected_results=2)

    def test_func_0028_no_source_same_event_code_different_device(self):
        """Ascend_Fd_SDK_Func_0028"""
        results = self.assert_diag_success([{"server": "node-1", "fault": [build_kg_analyzer()]}])
        self.assertTrue(results[0].get("analyze_success"))

    def test_func_0023_no_source_same_device_chain_event_code(self):
        """Ascend_Fd_SDK_Func_0023"""
        results = self.assert_diag_success([{"server": "node-1", "fault": [build_kg_analyzer()]}])
        self.assertTrue(results[0].get("fault"))

    def test_reliability_0014_mixed_ccae_and_non_ccae(self):
        """Ascend_Fd_SDK_Reliability_0014"""
        payload = [
            build_ccae_input(server="node-1"),
            {"server": "node-2", "source": "OTHER", "fault": [build_kg_analyzer()]},
        ]
        self.assert_diag_success(payload, expected_results=2)

    def test_reliability_0015_duplicate_event_code_and_device(self):
        """Ascend_Fd_SDK_Reliability_0015"""
        fault = build_ccae_fault()
        results = self.assert_diag_success([build_ccae_input(faults=[fault, dict(fault)])])
        self.assertIn(CHAIN_SOURCE_CODE, self.get_codes(results[0]))

    def test_reliability_0016_one_server_fault_empty(self):
        """Ascend_Fd_SDK_Reliability_0016"""
        payload = [build_ccae_input(server="node-1"), build_ccae_input(server="node-2", faults=[])]
        results, err_msg_list = diag_knowledge_graph(payload)
        self.assertEqual(len(results), 1)
        self.assertTrue(err_msg_list)

    def test_reliability_0017_multiple_elements_same_server(self):
        """Ascend_Fd_SDK_Reliability_0017"""
        payload = [build_ccae_input(server="node-1"), build_ccae_input(server="node-1")]
        self.assert_diag_success(payload, expected_results=2)

    def test_reliability_0018_empty_list(self):
        """Ascend_Fd_SDK_Reliability_0018"""
        results, err_msg_list = diag_knowledge_graph([])
        self.assertEqual(results, [])
        self.assertIn("All input invalid or empty input", err_msg_list[0])

    def test_reliability_0025_single_fault_more_than_1000(self):
        """Ascend_Fd_SDK_Reliability_0025"""
        faults = [build_ccae_fault(source_device=str(index % 16)) for index in range(1001)]
        results = self.assert_diag_success([build_ccae_input(faults=faults)])
        self.assertIn(CHAIN_SOURCE_CODE, self.get_codes(results[0]))

    def test_reliability_0024_max_4096_nodes(self):
        """Ascend_Fd_SDK_Reliability_0024"""
        payload = [{"server": f"node-{index}", "source": "OTHER", "fault": [{}]} for index in range(4096)]
        results, err_msg_list = diag_knowledge_graph(payload)
        self.assertEqual(err_msg_list, [])
        self.assertEqual(len(results), 4096)

    def test_security_0040_server_type_not_str(self):
        """Ascend_Fd_SDK_Security_0040"""
        self.assert_diag_failed([build_ccae_input(server=1)], "input_log_list[0].server")

    def test_security_0041_source_type_not_str(self):
        """Ascend_Fd_SDK_Security_0041"""
        self.assert_diag_failed([build_ccae_input(source=1)], "input_log_list[0].source")

    def test_security_0042_ccae_fault_type_not_list(self):
        """Ascend_Fd_SDK_Security_0042"""
        self.assert_diag_failed([build_ccae_input(faults={})], "input_log_list[0].fault")

    def test_security_0043_event_code_type_not_str(self):
        """Ascend_Fd_SDK_Security_0043"""
        self.assert_diag_failed([build_ccae_input(faults=[build_ccae_fault(event_code=1)])], "fault[0].event_code")

    def test_security_0044_key_info_type_not_str(self):
        """Ascend_Fd_SDK_Security_0044"""
        self.assert_diag_failed([build_ccae_input(faults=[build_ccae_fault(key_info=1)])], "fault[0].key_info")

    def test_security_0045_source_file_type_not_str(self):
        """Ascend_Fd_SDK_Security_0045"""
        self.assert_diag_failed([build_ccae_input(faults=[build_ccae_fault(source_file=1)])], "fault[0].source_file")

    def test_security_0046_source_device_type_not_str(self):
        """Ascend_Fd_SDK_Security_0046"""
        self.assert_diag_failed(
            [build_ccae_input(faults=[build_ccae_fault(source_device=1)])], "fault[0].source_device"
        )

    def test_security_0047_occur_time_type_not_str(self):
        """Ascend_Fd_SDK_Security_0047"""
        self.assert_diag_failed([build_ccae_input(faults=[build_ccae_fault(occur_time=1)])], "fault[0].occur_time")

    def test_security_0048_server_empty(self):
        """Ascend_Fd_SDK_Security_0048"""
        self.assert_diag_failed([build_ccae_input(server="")], "Empty value not allowed")

    def test_security_0049_source_empty(self):
        """Ascend_Fd_SDK_Security_0049"""
        results = self.assert_diag_success([{"server": "node-1", "source": "", "fault": [build_kg_analyzer()]}])
        self.assertIn("AISW_CANN_Runtime_021", self.get_codes(results[0]))

    def test_security_0050_source_not_ccae(self):
        """Ascend_Fd_SDK_Security_0050"""
        self.assert_diag_success([{"server": "node-1", "source": "OTHER", "fault": [build_kg_analyzer()]}])

    def test_security_0051_source_device_empty(self):
        """Ascend_Fd_SDK_Security_0051"""
        self.assert_diag_failed(
            [build_ccae_input(faults=[build_ccae_fault(source_device="")])], "Empty value not allowed"
        )

    def test_security_0052_occur_time_empty(self):
        """Ascend_Fd_SDK_Security_0052"""
        self.assert_diag_failed([build_ccae_input(faults=[build_ccae_fault(occur_time="")])], "Empty value not allowed")

    def test_security_0053_server_missing(self):
        """Ascend_Fd_SDK_Security_0053"""
        payload = build_ccae_input()
        payload.pop("server")
        self.assert_diag_failed([payload], "Missing required field: 'input_log_list[0].server'")

    def test_security_0054_fault_missing(self):
        """Ascend_Fd_SDK_Security_0054"""
        payload = build_ccae_input()
        payload.pop("fault")
        self.assert_diag_failed([payload], "Missing required field: 'input_log_list[0].fault'")

    def test_security_0055_event_code_missing(self):
        """Ascend_Fd_SDK_Security_0055"""
        fault = build_ccae_fault()
        fault.pop("event_code")
        self.assert_diag_failed([build_ccae_input(faults=[fault])], "fault[0].event_code")

    def test_security_0056_key_info_missing(self):
        """Ascend_Fd_SDK_Security_0056"""
        fault = build_ccae_fault()
        fault.pop("key_info")
        self.assert_diag_failed([build_ccae_input(faults=[fault])], "fault[0].key_info")

    def test_security_0057_source_file_missing(self):
        """Ascend_Fd_SDK_Security_0057"""
        fault = build_ccae_fault()
        fault.pop("source_file")
        self.assert_diag_failed([build_ccae_input(faults=[fault])], "fault[0].source_file")

    def test_security_0058_source_device_missing(self):
        """Ascend_Fd_SDK_Security_0058"""
        fault = build_ccae_fault()
        fault.pop("source_device")
        self.assert_diag_failed([build_ccae_input(faults=[fault])], "fault[0].source_device")

    def test_security_0059_occur_time_missing(self):
        """Ascend_Fd_SDK_Security_0059"""
        fault = build_ccae_fault()
        fault.pop("occur_time")
        self.assert_diag_failed([build_ccae_input(faults=[fault])], "fault[0].occur_time")


if __name__ == "__main__":
    unittest.main()
