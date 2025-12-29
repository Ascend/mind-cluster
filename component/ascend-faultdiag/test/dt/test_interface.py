#!/usr/bin/env python3
# -*- coding: utf-8 -*-
# Copyright 2025 Huawei Technologies Co., Ltd
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

from ascend_fd import parse_fault_type
from ascend_fd import parse_knowledge_graph
from ascend_fd import parse_root_cluster
from ascend_fd import diag_knowledge_graph
from ascend_fd import diag_root_cluster
from ascend_fd.utils.regular_table import MINDIE_SOURCE, CANN_PLOG_SOURCE, CANN_DEVICE_SOURCE, TRAIN_LOG_SOURCE, \
    AMCT_SOURCE, DOCKER_RUNTIME_SOURCE, NPU_EXPORTER_SOURCE, DEVICEPLUGIN_SOURCE, LCNE_SOURCE, NODEDLOG_SOURCE, \
    VOLCANO_SCHEDULER_SOURCE, NPU_OS_SOURCE, NPU_DEVICE_SOURCE, OS_SOURCE, OS_SYSMON_SOURCE, OS_VMCORE_DMESG_SOURCE, \
    OS_DEMESG_SOURCE, NPU_INFO_SOURCE


class TestFaultTypeParser(unittest.TestCase):
    @staticmethod
    def get_output_fault_output(output):
        fault_list, _ = output
        return [fault.get("error_type", "") for fault in fault_list]

    def setUp(self) -> None:
        pass

    def test_invalid_input_log_list(self):
        _, err = parse_fault_type({})
        self.assertIn("Invalid parameter type for 'input_log_list', it should be 'list'.", err[0])

    def test_common_use(self):
        input_list = [
            {
                "log_domain": {
                    "server": "x.x.x.x",
                    "device": ["0"]
                },
                "log_items": [
                    {
                        "item_type": MINDIE_SOURCE,
                        "log_lines": [
                            "[error] MIE03E400008"
                        ]
                    }
                ]
            }
        ]
        fault_list = self.get_output_fault_output(parse_fault_type(input_list))
        self.assertIn("AISW_MindIE_MS_HttpServer_01", fault_list)

    def test_multiline_matching(self):
        input_list = [
            {
                "log_domain": {
                    "server": "x.x.x.x",
                    "device": ["0"]
                },
                "log_items": [
                    {
                        "item_type": TRAIN_LOG_SOURCE,
                        "log_lines": [
                            "The parameters number of the function is",
                            "but the number of provided arguments is"
                        ]
                    }
                ]
            }
        ]
        fault_list = self.get_output_fault_output(parse_fault_type(input_list))
        self.assertIn("AISW_MindSpore_Compiler_09", fault_list)

    def test_source_check_err_handle(self):
        input_list = [
            {
                "log_domain": {
                    "server": "x.x.x.x",
                    "device": ["0"]
                },
                "log_items": [
                    {
                        "item_type": "Mindie|2|3|4|5|6|7|8|9|10|11",
                        "log_lines": []
                    }
                ]
            }
        ]
        _, err_msg_list = parse_fault_type(input_list)
        self.assertIn("Invalid item_type", err_msg_list[0])

    def test_list_exceeds_limit(self):
        list_out_of_limit = [i for i in range(300)]
        _, err_msg_list = parse_fault_type(list_out_of_limit)
        self.assertIn("the input list exceeds the limit", err_msg_list[0])

    def test_invalid_item_type(self):
        input_list = [
            {
                "log_domain": {
                    "server": "x.x.x.x",
                    "device": ["0"]
                },
                "log_items": [
                    {
                        "item_type": "",
                        "log_lines": []
                    }
                ]
            }
        ]
        _, err_msg_list = parse_fault_type(input_list)
        self.assertIn("both 'item_type' and 'log_lines' should be available and valid", err_msg_list[0])

    def test_missing_domain(self):
        input_list = [
            {
                "log_items": [
                    {
                        "item_type": MINDIE_SOURCE,
                        "log_lines": []
                    }
                ]
            }
        ]
        _, err_msg_list = parse_fault_type(input_list)
        self.assertIn("both 'log_domain' and 'log_items' should be available and valid", err_msg_list[0])

    def test_invalid_server(self):
        input_list = [
            {
                "log_domain": {
                    "server": "",
                    "device": ["0"]
                },
                "log_items": [
                    {
                        "item_type": MINDIE_SOURCE,
                        "log_lines": []
                    }
                ]
            }
        ]
        _, err_msg_list = parse_fault_type(input_list)
        self.assertIn("'server' should be available and valid for 'log_domain'", err_msg_list[0])

    def test_no_device(self):
        input_list = [
            {
                "log_domain": {
                    "server": "x.x.x.x"
                },
                "log_items": [
                    {
                        "item_type": MINDIE_SOURCE,
                        "log_lines": []
                    }
                ]
            }
        ]
        _, err_msg_list = parse_fault_type(input_list)
        self.assertIn("'device' should be a filed of 'log_domain'", err_msg_list[0])

    def test_no_log_lines(self):
        input_list = [
            {
                "log_domain": {
                    "server": "x.x.x.x",
                    "device": ["0"]
                },
                "log_items": [
                    {
                        "item_type": MINDIE_SOURCE
                    }
                ]
            }
        ]
        _, err_msg_list = parse_fault_type(input_list)
        self.assertIn("'log_lines' need to exist for a 'log_items' element", err_msg_list[0])

    def tearDown(self) -> None:
        pass


class TestKnowledgeGraphParser(unittest.TestCase):
    @staticmethod
    def get_root_cause(device_id: str, output):
        kg_analyzer_list, _ = output
        root_cause = kg_analyzer_list[0].get("fault", [])[0].get("response", {}).get(device_id, {}).get("root_causes",
                                                                                                        {})
        return root_cause

    def setUp(self) -> None:
        pass

    def test_invalid_input_log_list(self):
        _, err = parse_knowledge_graph({})
        self.assertIn("Invalid parameter type for 'input_log_list', it should be 'list'.", err[0])

    def test_invalid_custom_entity(self):
        _, err = parse_knowledge_graph([], [{}])
        self.assertIn("Invalid parameter type for 'custom_entity', it should be 'dict'.", err[0])

    def test_mindie_common_use(self):
        input_list = [
            {
                "log_domain": {
                    "server": "x.x.x.x"
                },
                "log_items": [
                    {
                        "item_type": MINDIE_SOURCE,
                        "device_id": 0,
                        "log_lines": [
                            '[2024-11-05 12:00:00.123456] [11] [1] [MS] [1] [ERROR] : '
                            '[MIE03E400008] [HttpServer] Http server on timeout.'
                        ],
                        "component": "Controller"
                    },
                    {
                        "item_type": MINDIE_SOURCE,
                        "device_id": 1,
                        "log_lines": [
                            '[2024-11-05 12:00:00.123456] [11] [1] [MS] [1] [ERROR] : '
                            '[MIE03E400002] [HttpServer] Http server on timeout.'
                        ]
                    }
                ]
            }
        ]
        root_cause = self.get_root_cause("0", parse_knowledge_graph(input_list))
        self.assertIn("AISW_MindIE_MS_HttpServer_01", root_cause.keys())
        fault_type = root_cause.get("AISW_MindIE_MS_HttpServer_01", {}).get("events_attribute", [{}])[0].get("type", "")
        self.assertEqual(fault_type, "MindIE_Controller")
        root_cause = self.get_root_cause("1", parse_knowledge_graph(input_list))
        self.assertIn("AISW_MindIE_MS_HttpServer_02", root_cause.keys())
        fault_type = root_cause.get("AISW_MindIE_MS_HttpServer_02", {}).get("events_attribute", [{}])[0].get("type", "")
        self.assertEqual(fault_type, "MindIE")

    def test_internationalization(self):
        input_list = [
            {
                "log_domain": {
                    "server": "x.x.x.x"
                },
                "log_items": [
                    {
                        "item_type": MINDIE_SOURCE,
                        "log_lines": [
                            '[2024-11-05 12:00:00.123456] [11] [1] [MS] [1] [ERROR] : '
                            '[MIE03E400008] [HttpServer] Http server on timeout.'
                        ]
                    }
                ]
            }
        ]
        root_cause = self.get_root_cause("0", parse_knowledge_graph(input_list))
        entities_attribute = root_cause.get("AISW_MindIE_MS_HttpServer_01", {}).get("entities_attribute", {})
        self.assertNotIn("cause_en", entities_attribute)
        self.assertNotIn("description_en", entities_attribute)
        self.assertNotIn("suggestion_en", entities_attribute)

    def test_component_validator(self):
        input_list = [
            {
                "log_domain": {
                    "server": "x.x.x.x"
                },
                "log_items": [
                    {
                        "item_type": MINDIE_SOURCE,
                        "log_lines": ["xxx"],
                        "component": "Invalid Component"
                    }
                ]
            }
        ]
        _, err_msg_list = parse_knowledge_graph(input_list)
        self.assertIn("not in the allowed choices: ['Controller', 'Coordinator']", err_msg_list[0])

    def test_plog_device_assignment(self):
        input_list = [
            {
                "log_domain": {
                    "server": "x.x.x.x"
                },
                "log_items": [
                    {
                        "item_type": CANN_PLOG_SOURCE,
                        "path": "",
                        "device_id": 0,
                        "log_lines": [
                            '[ERROR] RUNTIME(27514,python):2024-03-16-15:52:17.572.654 [task_info.cc:5261]28555 '
                            'PrintErrorInfoForNotifyWaitTask:[COMP][DEFAULT]Notify wait execute failed, device_id=1, '
                            'stream_id=20, task_id=6, flip_num=0, notify_id=3'
                        ]
                    }
                ]
            }
        ]
        expected_device_id = "0"
        root_cause = self.get_root_cause(expected_device_id, parse_knowledge_graph(input_list))
        self.assertIn("AISW_CANN_Runtime_021", root_cause.keys())
        device_id = root_cause.get("AISW_CANN_Runtime_021", {}).get("events_attribute", {})[0].get("source_device", "")
        self.assertEqual(expected_device_id, device_id)

    def test_log_lines_type_err(self):
        input_list = [
            {
                "log_domain": {
                    "server": "x.x.x.x"
                },
                "log_items": [
                    {
                        "item_type": MINDIE_SOURCE,
                        "path": "",
                        "device_id": 0,
                        "log_lines": [5]
                    }
                ]
            }
        ]
        _, err_msg_list = parse_knowledge_graph(input_list)
        self.assertIn("Type mismatch for 'input_log_list[0].log_items[0].log_lines[0]', "
                      "expected str, got int instead", err_msg_list[0])

    def test_plog_device_id_reference(self):
        input_list = [
            {
                "log_domain": {
                    "server": "x.x.x.x"
                },
                "log_items": [
                    {
                        "item_type": CANN_DEVICE_SOURCE,
                        "path": "debug/device-2/device-51_20231129113051449.log",
                        "log_lines": [
                            "[ERROR] AICPU(25602,aicpu_scheduler):2023-12-01-01:47:18.061.111 load stream "
                            "and task failed."
                        ]
                    },
                    {
                        "item_type": CANN_PLOG_SOURCE,
                        "path": "debug/plog/plog-51_20231129113051449.log",
                        "device_id": 3,
                        "log_lines": [
                            "[INFO] HCCL(69402,python3):2024-04-07-18:35:53.395.735 [topoinfo_detect.cc:208]"
                            "[69402][HCCL_TRACE]SetupAgent rankNum[3], rank[0], rootInfo "
                            "identifier[10.136.181.175%enp179s0f0_60000_0_1712529353144389], "
                            "server[10.136.181.175%enp179s0f0], logicDevId[0], phydevId[2], deviceIp[192.168.13.16]"
                        ]
                    }
                ]
            }
        ]
        expected_device_id = "3"
        root_cause = self.get_root_cause(expected_device_id, parse_knowledge_graph(input_list))
        self.assertIn("AISW_CANN_AICPU_04", root_cause.keys())
        device_id = root_cause.get("AISW_CANN_AICPU_04", {}).get("events_attribute", {})[0].get("source_device", "")
        self.assertEqual(expected_device_id, device_id)

    def test_resuming_training(self):
        input_list = [
            {
                "log_domain": {
                    "server": "x.x.x.x"
                },
                "log_items": [
                    {
                        "item_type": CANN_PLOG_SOURCE,
                        "path": "process_log/run/plog-123_20240110032852709.log",
                        "log_lines": [
                            "[INFO] HCCL(11262,python3):2023-01-01-01:48:31.246.088 "
                            "[trace_attr.c:189](tid:229) attr init success, timeout=0ms.",
                            "[INFO] HCCL(11262,python3):2023-01-01-02:00:00.000.000 "
                            "[trace_attr.c:189](tid:229) attr init success, timeout=0ms."
                        ]
                    },
                    {
                        "item_type": CANN_PLOG_SOURCE,
                        "path": "process_log/debug/plog-11262_20240110032852701.log",
                        "log_lines": [
                            "[ERROR] HCCL(11262,python3):2023-01-01-01:00:00.000.000 Not sched cpu.",
                            "[ERROR] HCCL(11262,python3):2023-01-01-02:01:00.000.000 There is no subscribe thread."
                        ]
                    }
                ]
            }
        ]
        expected_device_id = "Unknown"
        root_cause = self.get_root_cause(expected_device_id, parse_knowledge_graph(input_list))
        self.assertNotIn("AISW_CANN_DRV_ESCH_013", root_cause.keys())
        self.assertIn("AISW_CANN_DRV_ESCH_008", root_cause.keys())

    def test_train_log_parse(self):
        input_list = [
            {
                "log_domain": {
                    "server": "x.x.x.x"
                },
                "log_items": [
                    {
                        "item_type": TRAIN_LOG_SOURCE,
                        "device_id": 2,
                        "log_lines": [
                            "get socket timeout",
                            "The name is not defined  Call runtime aclrtSynchronizeStreamWithTimeout error",
                            "or not supported in graph mode"
                        ]
                    }
                ]
            }
        ]
        expected_device_id = "2"
        root_cause = self.get_root_cause(expected_device_id, parse_knowledge_graph(input_list))
        self.assertNotIn("AISW_MindSpore_Ascend_Backend_26", root_cause.keys())
        self.assertIn("AISW_MindSpore_Compiler_01", root_cause.keys())
        self.assertIn("AISW_CANN_ERRMSG_Custom_04", root_cause.keys())

    def test_os_parse(self):
        expected_time = "2022-10-10 23:00:00.000000"
        input_list = [
            {
                "log_domain": {
                    "server": "x.x.x.x"
                },
                "log_items": [
                    {
                        "item_type": OS_SOURCE,
                        "modification_time": expected_time,
                        "log_lines": [
                            "Mar 15 03:40:50 sha2uvp-gert01 kernel: [16794.247724] detected stalls on"
                        ]
                    },
                    {
                        "item_type": OS_VMCORE_DMESG_SOURCE,
                        "modification_time": expected_time,
                        "log_lines": [
                            "[  293.039556] NMI watchdog: Watchdog detected hard LOCKUP on cpu 34"
                        ]
                    },
                    {
                        "item_type": OS_SYSMON_SOURCE,
                        "modification_time": expected_time,
                        "log_lines": [
                            "2024-08-30T17:01:48.559312+00:00|info|sysmonitor[12706]: zombie process"
                        ]
                    },
                    {
                        "item_type": OS_DEMESG_SOURCE,
                        "modification_time": expected_time,
                        "log_lines": [
                            "[Mon Sep 30 15:00:00 2024] systemd[1]: license-manager.service: I/O error."
                        ]
                    }
                ]
            }
        ]
        expected_device_id = "Unknown"
        root_cause = self.get_root_cause(expected_device_id, parse_knowledge_graph(input_list))
        exp_os_code = "Comp_OS_Kernel_CPU_06"
        self.assertIn(exp_os_code, root_cause.keys())
        os_occur_time = root_cause.get(exp_os_code, {}).get("events_attribute", {})[0].get("occur_time", "")
        self.assertEqual("2022-03-15 03:40:50.000000", os_occur_time)

        exp_vmcore_code = "Comp_OS_Kernel_CPU_04"
        self.assertIn(exp_vmcore_code, root_cause.keys())
        vmcore_occur_time = root_cause.get(exp_vmcore_code, {}).get("events_attribute", {})[0].get("occur_time", "")
        self.assertEqual(expected_time, vmcore_occur_time)

        exp_sysmon_code = "Comp_OS_Kernel_CPU_11"
        self.assertIn(exp_sysmon_code, root_cause.keys())
        sysmon_occur_time = root_cause.get(exp_sysmon_code, {}).get("events_attribute", {})[0].get("occur_time", "")
        self.assertNotEqual(expected_time, sysmon_occur_time)

        exp_dmesg_code = "Comp_OS_HW_Stor_02"
        self.assertIn(exp_dmesg_code, root_cause.keys())
        dmesg_occur_time = root_cause.get(exp_dmesg_code, {}).get("events_attribute", {})[0].get("occur_time", "")
        self.assertEqual("2024-09-30 15:00:00.000000", dmesg_occur_time)

    def test_os_group_filter(self):
        input_list = [
            {
                "log_domain": {
                    "server": "x.x.x.x"
                },
                "log_items": [
                    {
                        "item_type": OS_SOURCE,
                        "path": "/var/log/messages",
                        "log_lines": [
                            "Mar 15 03:40:50 sha2uvp-gert01 kernel: [16794.247724] detected stalls on"
                        ]
                    },
                    {
                        "item_type": OS_SOURCE,
                        "path": "",
                        "log_lines": [
                            "Mar 16 03:40:50 sha2uvp-gert01 kernel: [16794.247724] "
                            "format error or unsupported for import"
                        ]
                    },
                    {
                        "item_type": OS_SOURCE,
                        "path": "/var/log/messages",
                        "log_lines": [
                            "Mar 17 03:40:50 sha2uvp-gert01 kernel: [16794.247724] "
                            "ready to call notify chain to isolate page"
                        ]
                    },
                    {
                        "item_type": OS_SOURCE,
                        "path": "/var/log/messages-1",
                        "log_lines": [
                            "Mar 17 03:40:50 sha2uvp-gert01 kernel: [16794.247724] No space left on device"
                        ]
                    }
                ]
            }
        ]
        expected_device_id = "Unknown"
        root_cause = self.get_root_cause(expected_device_id, parse_knowledge_graph(input_list))
        self.assertIn("Comp_OS_Kernel_CPU_06", root_cause)
        self.assertIn("Comp_OS_Custom_04", root_cause)
        self.assertIn("Comp_OS_Service_Container_04", root_cause)
        self.assertNotIn("Comp_OS_Service_Container_03", root_cause)

    def test_npu_device_parse(self):
        input_list = [
            {
                "log_domain": {
                    "server": "x.x.x.x"
                },
                "log_items": [
                    {
                        "item_type": NPU_OS_SOURCE,
                        "log_lines": [
                            "[ERROR] KERNEL(3551,sklogd):2021-10-05-23:07:08.350.235 "
                            "[klogd.c:242][583733.716915] event_id=0x80c78008"
                        ]
                    },
                    {
                        "item_type": NPU_DEVICE_SOURCE,
                        "log_lines": [
                            "[EVENT] IMP(-2,IMP): 2024-12-04-08:21:54.994.714 70 "
                            "(device_id:1 die_id:0) [IMP] present 1, mac_link 1, "
                            "pcs_link 1, rf_lf 1, pcs_err_cnt 5, pcs_64_66b 4681"
                        ]
                    }
                ]
            }
        ]
        expected_device_id = "Unknown"
        root_cause = self.get_root_cause(expected_device_id, parse_knowledge_graph(input_list))
        self.assertIn("0x80C78008", root_cause.keys())
        self.assertIn("Comp_Network_Custom_11", root_cause.keys())

    def test_npu_os_parse(self):
        input_list = [
            {
                "log_domain": {
                    "server": "x.x.x.x"
                },
                "log_items": [
                    {
                        "item_type": NPU_OS_SOURCE,
                        "path": "/test/event/event_20251125162247511.log",
                        "log_lines": [
                            "[EVENT] ROCE(3464,dmp_daemon):2025-11-25-16:22:49.053.785 [xsfp.c:896]xsfp_get_info(896) "
                            ": xsfp_optical_ready_flag has changed from -1 to 0.(dev_id=0)"
                        ]
                    }
                ]
            }
        ]
        _, err_msg = parse_knowledge_graph(input_list)
        self.assertEqual(err_msg, [])

    def test_npu_device_parse_path_id(self):
        input_list = [
            {
                "log_domain": {
                    "server": "x.x.x.x"
                },
                "log_items": [
                    {
                        "item_type": NPU_OS_SOURCE,
                        "path": "log/slog/dev-os-3/debug/device-os/device-os_20240312200914967.log",
                        "log_lines": [
                            "[ERROR] KERNEL(3551,sklogd):2021-10-05-23:07:08.350.235 "
                            "[klogd.c:242][583733.716915] event_id=0x80c78008"
                        ]
                    }
                ]
            }
        ]
        expected_device_id = "3"
        root_cause = self.get_root_cause(expected_device_id, parse_knowledge_graph(input_list))
        device_id = root_cause.get("0x80C78008", {}).get("events_attribute", {})[0].get("source_device", "")
        self.assertEqual(expected_device_id, device_id)

    def test_modification_time(self):
        expected_time = "2022-10-10 23:00:00.000000"
        input_list = [
            {
                "log_domain": {
                    "server": "x.x.x.x"
                },
                "log_items": [
                    {
                        "item_type": TRAIN_LOG_SOURCE,
                        "modification_time": expected_time,
                        "log_lines": [
                            "get socket timeout"
                        ]
                    }
                ]
            }
        ]
        expected_device_id = "Unknown"
        root_cause = self.get_root_cause(expected_device_id, parse_knowledge_graph(input_list))
        occur_time = root_cause.get("AISW_CANN_ERRMSG_Custom_04", {}).get(
            "events_attribute", {})[0].get("occur_time", "")
        self.assertEqual(expected_time, occur_time)

    def test_parse_amct(self):
        input_list = [
            {
                "log_domain": {
                    "server": "x.x.x.x"
                },
                "log_items": [
                    {
                        "item_type": AMCT_SOURCE,
                        "log_lines": [
                            "2024-10-23 10:47:11,976 - INFO - [AMCT]:[AMCT]: Layer batchmatmul_2 does not support "
                            "asymmetric quant, set symmetric, No layer support in the graph"
                        ]
                    }
                ]
            }
        ]
        expected_device_id = "Unknown"
        root_cause = self.get_root_cause(expected_device_id, parse_knowledge_graph(input_list))
        self.assertIn("AISW_CANN_AMCT_ALL_001", root_cause.keys())

    def test_parse_docker_runtime(self):
        input_list = [
            {
                "log_domain": {
                    "server": "x.x.x.x"
                },
                "log_items": [
                    {
                        "item_type": DOCKER_RUNTIME_SOURCE,
                        "log_lines": [
                            "[INFO]     2024/01/10 03:29:09.354927 1       hwlog/api.go:108    failed to check files, "
                            "owner not right /usr/bin/runc xxxx"
                        ]
                    }
                ]
            }
        ]
        expected_device_id = "Unknown"
        root_cause = self.get_root_cause(expected_device_id, parse_knowledge_graph(input_list))
        self.assertIn("AISW_MindX_Docker_Runtime_02", root_cause.keys())

    def test_parse_npu_exporter(self):
        input_list = [
            {
                "log_domain": {
                    "server": "x.x.x.x"
                },
                "log_items": [
                    {
                        "item_type": NPU_EXPORTER_SOURCE,
                        "log_lines": [
                            "[ERROR]    2024/01/10 02:29:39.034681 60      collector/npu_collector.go:1924    "
                            "deviceManager init failed, prepare dcmi failed, err: cannot found valid driver lib, "
                            "fromEnv: lib path is invalid, [/usr/local/Ascend/driver/lib64/driver/libdcmi.so: "
                            "check uid or mode failed; /usr/local/dcmi/libdcmi.so: check uid or mode failed;]"
                        ]
                    }
                ]
            }
        ]
        expected_device_id = "Unknown"
        root_cause = self.get_root_cause(expected_device_id, parse_knowledge_graph(input_list))
        self.assertIn("AISW_MindX_Npu_Exporter_01", root_cause.keys())

    def test_device_plugin_recovery(self):
        input_list = [
            {
                "log_domain": {
                    "server": "x.x.x.x"
                },
                "log_items": [
                    {
                        "item_type": DEVICEPLUGIN_SOURCE,
                        "log_lines": [
                            '[WARN]     2024/01/10 03:30:38.643330 17      deviceswitch/ascend_switch.go:190    '
                            'switch subscribe got fault:deviceswitch.SwitchFaultEvent{EventType:0x4, FaultID:0xf1ff09, '
                            'AssembledFaultCode:"[0x00f1ff09,155912,npu,na]", PeerPortDevice:0x1, PeerPortId:0x2, '
                            'SwitchChipId:0x0, SwitchPortId:0x0, Severity:0x0, Assertion:0x1, EventSerialNum:0, '
                            'NotifySerialNum:0, AlarmRaisedTime:0, AdditionalParam:"", AdditionalInfo:""}, '
                            'faultCode:[0x00f1ff09,155912,npu,na]',
                            '[WARN]     2024/01/10 03:31:38.643290 17      deviceswitch/ascend_switch.go:190    '
                            'switch subscribe got fault:deviceswitch.SwitchFaultEvent{EventType:0x4, '
                            'FaultID:0xf103b0, AssembledFaultCode:"[0x00f103b0,155907,na,na]", '
                            'PeerPortDevice:0x0, PeerPortId:0x2, SwitchChipId:0x0, SwitchPortId:0x0, Severity:0x0, '
                            'Assertion:0x1, EventSerialNum:0, NotifySerialNum:0, AlarmRaisedTime:0, '
                            'AdditionalParam:"", AdditionalInfo:""}, faultCode:[0x00f103b0,155907,na,na]',
                            '[WARN]     2024/01/10 03:32:38.643290 17      deviceswitch/ascend_switch.go:190    '
                            'switch subscribe got fault:deviceswitch.SwitchFaultEvent{EventType:0x4, '
                            'FaultID:0xf103b0, AssembledFaultCode:"[0x00f103b0,155907,na,na]", PeerPortDevice:0x0, '
                            'PeerPortId:0x2, SwitchChipId:0x0, SwitchPortId:0x0, Severity:0x0, Assertion:0x0, '
                            'EventSerialNum:0, NotifySerialNum:0, AlarmRaisedTime:0, AdditionalParam:"", '
                            'AdditionalInfo:""}, faultCode:[0x00f103b0,155907,na,na]'

                        ]
                    }
                ]
            }
        ]
        expected_device_id = "Unknown"
        root_cause = self.get_root_cause(expected_device_id, parse_knowledge_graph(input_list))
        self.assertIn("Comp_Switch_L1_06", root_cause.keys())
        self.assertNotIn("Comp_Switch_L1_01", root_cause.keys())

    def test_parse_lcne(self):
        input_list = [
            {
                "log_domain": {
                    "server": "x.x.x.x"
                },
                "log_items": [
                    {
                        "item_type": LCNE_SOURCE,
                        "log_lines": [
                            "Aug 12 2025 08:53:46+08:00 Z3_22R_99_30 %%01CLI/5/CMDRECORD(s):CID=0x80ca2713;"
                            "Recorded command information. alarmID=0x00f1ff09"
                        ]
                    }
                ]
            }
        ]
        expected_device_id = "Unknown"
        root_cause = self.get_root_cause(expected_device_id, parse_knowledge_graph(input_list))
        self.assertIn("Comp_Switch_L1_02", root_cause.keys())

    def test_parse_noded(self):
        input_list = [
            {
                "log_domain": {
                    "server": "x.x.x.x"
                },
                "log_items": [
                    {
                        "item_type": NODEDLOG_SOURCE,
                        "log_lines": [
                            "[INFO]     2999/10/01 03:39:17.760916 305     ipmimonitor/ipmi_monitor.go:206    "
                            "get fault event, [device type: BMC]-[device id: 255]-[error code: 2C000031]"
                        ]
                    }
                ]
            }
        ]
        expected_device_id = "Unknown"
        root_cause = self.get_root_cause(expected_device_id, parse_knowledge_graph(input_list))
        self.assertIn("0x2C000031", root_cause.keys())

    def test_parse_volcano_component(self):
        input_list = [
            {
                "log_domain": {
                    "server": "x.x.x.x"
                },
                "log_items": [
                    {
                        "item_type": VOLCANO_SCHEDULER_SOURCE,
                        "log_lines": [
                            "E0111 03:29:15.991928       1 reschedule.go:533]    "
                            "Failed to get plugin volcano-npu_xxx_linux-xxx"
                        ]
                    }
                ]
            }
        ]
        expected_device_id = "Unknown"
        root_cause = self.get_root_cause(expected_device_id, parse_knowledge_graph(input_list))
        self.assertIn("AISW_MindX_Volcano_03", root_cause.keys())

    def test_parse_npu_info(self):
        input_list = [
            {
                "log_domain": {
                    "server": "x.x.x.x"
                },
                "log_items": [
                    {
                        "item_type": NPU_INFO_SOURCE,
                        "path": "npu_info_before.txt",
                        "log_lines": [
                            "/usr/local/Ascend/driver/tools/hccn_tool -i 0 -fec -g",
                            "fec mode: no FEC mode"
                        ]
                    },
                    {
                        "item_type": NPU_INFO_SOURCE,
                        "path": "npu_info_after.txt",
                        "log_lines": [
                            "xxxx"
                        ]
                    }
                ]
            }
        ]
        expected_device_id = "0"
        root_cause = self.get_root_cause(expected_device_id, parse_knowledge_graph(input_list))
        self.assertIn("Comp_Network_Custom_02", root_cause.keys())

    def test_temp_entity(self):
        input_list = [
            {
                "log_domain": {
                    "server": "x.x.x.x"
                },
                "log_items": [
                    {
                        "item_type": "Check",
                        "modification_time": "2025-01-01 20:22:33.999999",
                        "log_lines": [
                            "Machine Check"
                        ]
                    }
                ]
            }
        ]

        entity = {
            "test_entity": {
                "attribute.class": "Software",
                "attribute.component": "NPU",
                "attribute.module": "NPU",
                "attribute.cause_zh": "111",
                "attribute.description_zh": "22222",
                "attribute.suggestion_zh": "333333",
                "attribute.cause_en": "--",
                "attribute.description_en": "--",
                "attribute.suggestion_en": "--",
                "source_file": "Check",
                "rule": [],
                "regex.in": [
                    [
                        "Machine Check"
                    ]
                ]
            }
        }
        expected_device_id = "Unknown"
        root_cause = self.get_root_cause(expected_device_id, parse_knowledge_graph(input_list, entity))
        self.assertIn("test_entity", root_cause.keys())

    def test_modification_time_validation(self):
        input_list = [
            {
                "log_domain": {
                    "server": "x.x.x.x"
                },
                "log_items": [
                    {
                        "item_type": TRAIN_LOG_SOURCE,
                        "modification_time": "xxx",
                        "log_lines": [
                            "get socket timeout"
                        ]
                    }
                ]
            }
        ]
        _, err_msg_list = parse_knowledge_graph(input_list)
        self.assertIn("modification_time is invalid", err_msg_list[0])

    def test_modification_time_handle(self):
        actual_occur_time = "2025-01-01 20:22:33.999999"
        input_list = [
            {
                "log_domain": {
                    "server": "x.x.x.x"
                },
                "log_items": [
                    {
                        "item_type": TRAIN_LOG_SOURCE,
                        "modification_time": f"   {actual_occur_time}  ",
                        "log_lines": [
                            "get socket timeout"
                        ]
                    }
                ]
            }
        ]
        expected_device_id = "Unknown"
        root_cause = self.get_root_cause(expected_device_id, parse_knowledge_graph(input_list))
        expected_code = "AISW_CANN_ERRMSG_Custom_04"
        occur_time = root_cause.get(expected_code, {}).get("events_attribute", {})[0].get("occur_time", "")
        self.assertEqual(actual_occur_time, occur_time)

    def test_modification_time_tolerance(self):
        input_list = [
            {
                "log_domain": {
                    "server": "x.x.x.x"
                },
                "log_items": [
                    {
                        "item_type": TRAIN_LOG_SOURCE,
                        "modification_time": f"2025-01-01 20:22:33",
                        "log_lines": [
                            "synchronize_and_free_events"
                        ]
                    },
                    {
                        "item_type": TRAIN_LOG_SOURCE,
                        "modification_time": f"2025-01-01 20:22:33.333",
                        "log_lines": [
                            "npuSynchronizeDevice"
                        ]
                    }
                ]
            }
        ]
        expected_device_id = "Unknown"
        root_cause = self.get_root_cause(expected_device_id, parse_knowledge_graph(input_list))
        exp_time_1 = root_cause.get("AISW_PyTorch_Train_25", {}).get("events_attribute", {})[0] \
            .get("occur_time", "")
        self.assertEqual("2025-01-01 20:22:33.000000", exp_time_1)
        root_cause = self.get_root_cause(expected_device_id, parse_knowledge_graph(input_list))
        exp_time_2 = root_cause.get("AISW_PyTorch_Train_23", {}).get("events_attribute", {})[0] \
            .get("occur_time", "")
        self.assertEqual("2025-01-01 20:22:33.333000", exp_time_2)

    def test_temp_entity_check_all_failed(self):
        input_list = [
            {
                "log_domain": {
                    "server": "x.x.x.x"
                },
                "log_items": [
                    {
                        "item_type": "Check",
                        "modification_time": "2025-01-01 20:22:33.999999",
                        "log_lines": [
                            "Machine Check"
                        ]
                    }
                ]
            }
        ]

        entity = {
            "test_entity": {
                "attribute.class": "Software",
                "attribute.component": "NPU"
            },
            "another_test_entity": {
                "attribute.class": "Software",
                "attribute.component": "NPU"
            }
        }
        _, err_msg_list = parse_knowledge_graph(input_list, entity)
        self.assertIn("All codes update failed", err_msg_list[0])

    def test_temp_entity_check_partially_failed(self):
        input_list = [
            {
                "log_domain": {
                    "server": "x.x.x.x"
                },
                "log_items": [
                    {
                        "item_type": "Check",
                        "modification_time": "2025-01-01 20:22:33.999999",
                        "log_lines": [
                            "Machine Check"
                        ]
                    }
                ]
            }
        ]

        entity = {
            "test_entity": {
                "attribute.class": "Software",
                "attribute.component": "NPU",
                "attribute.module": "NPU",
                "attribute.cause_zh": "111",
                "attribute.description_zh": "22222",
                "attribute.suggestion_zh": "333333",
                "attribute.cause_en": "--",
                "attribute.description_en": "--",
                "attribute.suggestion_en": "--",
                "source_file": "Check",
                "rule": [],
                "regex.in": [
                    [
                        "Machine Check"
                    ]
                ]
            },
            "another_test_entity": {
                "attribute.class": "Software",
                "attribute.component": "NPU"
            }
        }
        result_list, err_msgs = parse_knowledge_graph(input_list, entity)
        self.assertEqual("Some entities failed to update, please check the input: ['another_test_entity']", err_msgs[0])
        root_cause = result_list[0].get("fault", [])[0].get("response", {}).get("Unknown", {}).get("root_causes", {})
        self.assertIn("test_entity", root_cause.keys())

    def test_multi_custom_with_default_fault(self):
        input_list = [
            {
                "log_domain": {
                    "server": "x.x.x.x"
                },
                "log_items": [
                    {
                        "item_type": "Check",
                        "modification_time": "2025-01-01 20:22:33.999999",
                        "log_lines": [
                            "Machine Check"
                        ]
                    },
                    {
                        "item_type": "EntityCheck",
                        "log_lines": [
                            "EntityCheck"
                        ]
                    },
                    {
                        "item_type": "MindIE",
                        "log_lines": [
                            "[2024-12-30 07:23:21.034+00:00] : [MIE03E400005] [HttpServer] "
                            "Certificate file is not exist."
                        ]
                    },
                    {
                        "item_type": "entity_should_not_in",
                        "log_lines": [
                            "Some keywords similar that can be matched"
                        ]
                    }
                ]
            }
        ]
        entity = {
            "test_entity": {
                "attribute.class": "Software",
                "attribute.component": "NPU",
                "attribute.module": "NPU",
                "attribute.cause_zh": "111",
                "attribute.description_zh": "22222",
                "attribute.suggestion_zh": "333333",
                "attribute.cause_en": "--",
                "attribute.description_en": "--",
                "attribute.suggestion_en": "--",
                "source_file": "Check",
                "rule": [],
                "regex.in": [["Machine Check"]]
            },
            "another_test_entity": {
                "attribute.class": "Software",
                "attribute.component": "NPU",
                "attribute.module": "NPU",
                "attribute.cause_zh": "111",
                "attribute.description_zh": "22222",
                "attribute.suggestion_zh": "333333",
                "attribute.cause_en": "--",
                "attribute.description_en": "--",
                "attribute.suggestion_en": "--",
                "source_file": "EntityCheck",
                "rule": [],
                "regex.in": [["EntityCheck"]]
            },
            "yet_another_test_entity": {
                "attribute.class": "Software",
                "attribute.component": "NPU",
                "attribute.module": "NPU",
                "attribute.cause_zh": "111",
                "attribute.description_zh": "22222",
                "attribute.suggestion_zh": "333333",
                "attribute.cause_en": "--",
                "attribute.description_en": "--",
                "attribute.suggestion_en": "--",
                "source_file": "TureType",
                "rule": [],
                "regex.in": [["Some keywords similar that can be matched"]]
            }
        }
        result_list, err_msgs = parse_knowledge_graph(input_list, entity)
        root_cause = result_list[0].get("fault", [])[0].get("response", {}).get("Unknown", {}).get("root_causes", {})
        self.assertIn("AISW_MindIE_MS_HttpServer_06", root_cause)
        self.assertIn("test_entity", root_cause)
        self.assertIn("another_test_entity", root_cause)
        self.assertNotIn("yet_another_test_entity", root_cause)
        self.assertIn("The following item types are unsupported", err_msgs[0])

    def tearDown(self) -> None:
        pass


class TestKnowledgeDiagnosis(unittest.TestCase):
    def setUp(self) -> None:
        pass

    def test_invalid_input_log_list(self):
        _, err = diag_knowledge_graph({})
        self.assertIn("Invalid parameter type for 'input_log_list', it should be 'list'.", err[0])

    def test_custom_event_input_with_ccae(self):
        input_list = [{
            "server": "node-97-17",
            "source": "ccae",
            "fault": [{
                "event_code": "AISW_CANN_ERRCODE_Custom_ERR02200",
                "key_info": "[ERROR] 2025-09-04-14:33:31 (PID:1041251, Device:6, RankID:6) "
                            "ERR02200 DIST call hccl api failed.",
                "source_file": "/home/train.log",
                "source_device": "Unknown",
                "occur_time": "2025-09-04 14:33:58.000000"
            }]
        }]
        diag_results, _ = diag_knowledge_graph(input_list)
        fault_dict = diag_results[0].get("fault", [{}])[0]
        self.assertNotEqual(fault_dict.get("description_zh", ""), "")

    def test_internationalization(self):
        input_list = [{
            "server": "node-97-17",
            "source": "ccae",
            "fault": [{
                "event_code": "AISW_MindSpore_MindData_16",
                "key_info": "",
                "source_file": "test1.txt",
                "source_device": "Unknown",
                "occur_time": "2024-04-08 16:24:48"
            }]
        }]
        diag_results, _ = diag_knowledge_graph(input_list)
        fault_dict = diag_results[0].get("fault", [{}])[0]
        self.assertNotIn("cause_en", fault_dict)
        self.assertNotIn("description_en", fault_dict)
        self.assertNotIn("suggestion_en", fault_dict)

    def tearDown(self) -> None:
        pass


class TestRootClusterParser(unittest.TestCase):
    def setUp(self) -> None:
        self.notify_err_log = "[ERROR] HCCL(69395,python3):2024-04-08-02:55:04.903.511 [task_exception_handler.cc:" \
                              + "452][69395][TaskExceptionHandler][Callback]Task run failed, base information is " \
                              + "streamID:[3], taskID[77726298], taskType[Notify Wait], tag[AllReduce_10.136.181.175" \
                              + "%enp179s0f0_60000_0_1712529353144389], index[3], AlgType(level 0-1-2):" \
                              + "[null-null-ring]."
        self.mindie_err_log_1 = ("[2025-05-30 21:47:44,982] [134] [281466201825504] [llm] [ERROR] [logging.py-56] :"
                                 " Link from 2.0.0.0 to 1.0.0.0 failed, error code is MIE05E01001B.")
        self.mindie_err_log_2 = ("[2025-05-30 21:47:44,982] [134] [281466201825504] [llm] [ERROR] [logging.py-56] :"
                                 " Link from 1.0.0.0 to 2.0.0.0 failed, error code is MIE05E01001B.")
        self.mindie_err_log_3 = ("[2025-06-03 14:21:44,857] [7855] [281464498917600] [llm] [ERROR] [logging.py-56] :"
                                 " 5.0.0.0 pull kv from 6.0.0.0 failed, error code is MIE05E01001A.")

    def test_invalid_input_log_list(self):
        _, err = parse_root_cluster({})
        self.assertIn("Invalid parameter type for 'input_log_list', it should be 'list'.", err[0])

    def test_common_use(self):
        input_list = [
            {
                "log_domain": {
                    "server": "x.x.x.x",
                    "instance_id": "xxxx1"
                },
                "log_items": [
                    {
                        "item_type": "plog",
                        "pid": 3199,
                        "rank_id": 0,
                        "device_id": 0,
                        "log_lines": [self.notify_err_log]
                    },
                    {
                        "item_type": "mindie",
                        "pid": 3199,
                        "rank_id": 0,
                        "device_id": 0,
                        "log_lines": [self.mindie_err_log_1,
                                      self.mindie_err_log_2,
                                      self.mindie_err_log_3]
                    }
                ]
            }
        ]
        parse_result, _ = parse_root_cluster(input_list)
        rc_parser = parse_result[0].get("3199")
        base_info = rc_parser.get("base", {})
        self.assertEqual("x.x.x.x", base_info.get("server_id", ""))
        self.assertEqual("0", base_info.get("logic_device_id", ""))
        self.assertEqual("0", base_info.get("phy_device_id", ""))
        self.assertEqual({"xxxx1": {"rank_num": 1, "rank_id": '0'}}, base_info.get("rank_map", {}))
        self.assertIn("xxxx1", base_info.get("root_list", []))
        err_info = rc_parser.get("error", {})
        self.assertEqual("Notify", err_info.get("timeout_error_events_list", [{}])[0].get("error_type"))
        # mindie相关信息检查
        mindie_parse_result = parse_result[1]
        self.assertTrue(mindie_parse_result.get("mindie", False))
        self.assertEqual(2, len(mindie_parse_result.get("link_error_info_map", {})))
        self.assertEqual(1, len(mindie_parse_result.get("pull_kv_error_map", {})))

    def test_rank_id_allocation(self):
        input_list = [
            {
                "log_domain": {
                    "server": "x.x.x.x",
                    "instance_id": "xxxx1"
                },
                "log_items": [
                    {
                        "item_type": "plog",
                        "pid": 3199,
                        "rank_id": 1,
                        "device_id": 0,
                        "log_lines": [self.notify_err_log]
                    },
                    {
                        "item_type": "plog",
                        "pid": 3399,
                        "device_id": 0,
                        "log_lines": ["xxxx"]
                    }
                ]
            },
            {
                "log_domain": {
                    "server": "x.y.x.x",
                    "instance_id": "xxxx1"
                },
                "log_items": [
                    {
                        "item_type": "plog",
                        "pid": 3299,
                        "rank_id": 3,
                        "device_id": 0,
                        "log_lines": ["xxxx"]
                    }
                ]
            }
        ]
        parse_result, _ = parse_root_cluster(input_list)
        rc_parser = parse_result[0].get("3399")
        base_info = rc_parser.get("base", {})
        self.assertEqual({"xxxx1": {"rank_num": 3, "rank_id": '0'}}, base_info.get("rank_map", {}))

    def test_tolerance_in_partial_dev_id_case(self):
        input_list = [
            {
                "log_domain": {
                    "server": "x.x.x.x",
                    "instance_id": "xxx"
                },
                "log_items": [
                    {
                        "item_type": "plog",
                        "pid": 241,
                        "device_id": 0,
                        "rank_id": 0,
                        "log_lines": [
                            "[INFO] HCCL(241,python3.8):2025-08-26-16:26:47.977.325 [op_base.cc:830] "
                            "[241][HCCL_TRACE]HcclGetRootInfo success, take time [6134]us"
                        ]
                    },
                    {
                        "item_type": "plog",
                        "pid": 753,
                        "log_lines": [
                            "[ERROR] TBE(753,python3.8):2025-08-26-16:37:03.889.560 raise error"
                        ]
                    }
                ]
            }
        ]
        parse_result, _ = parse_root_cluster(input_list)
        pid_with_dev_id = parse_result[0].get("241")
        base_info = pid_with_dev_id.get("base", {})
        self.assertEqual({"xxx": {"rank_num": 1, "rank_id": '0'}}, base_info.get("rank_map", {}))
        pid_without_dev_id = parse_result[0].get("753")
        base_info = pid_without_dev_id.get("base", {})
        self.assertEqual({}, base_info.get("rank_map", {}))

    def test_validation_in_non_first_item_case(self):
        input_list = [
            {
                "log_domain": {
                    "server": "x.x.x.x"
                },
                "log_items": [
                    {
                        "item_type": MINDIE_SOURCE,
                        "log_lines": [" "]
                    },
                    {
                        "item_type": MINDIE_SOURCE,
                        "log_lines": []
                    }
                ]
            }
        ]
        _, err_msg_list = parse_knowledge_graph(input_list)
        self.assertIn("Empty value not allowed for input_log_list[0].log_items[1].log_lines", err_msg_list[0])

    def test_validation_in_non_first_server_case(self):
        input_list = [
            {
                "log_domain": {
                    "server": "x.x.x.x"
                },
                "log_items": [
                    {
                        "item_type": MINDIE_SOURCE,
                        "log_lines": [" "]
                    }
                ]
            },
            {
                "log_domain": {},
                "log_items": [
                    {
                        "item_type": MINDIE_SOURCE,
                        "log_lines": [" "]
                    }
                ]
            }
        ]
        _, err_msg_list = parse_knowledge_graph(input_list)
        self.assertIn("Empty value not allowed for input_log_list[1].log_domain", err_msg_list[0])

    def tearDown(self) -> None:
        pass


class TestRootClusterDiagnosis(unittest.TestCase):

    def setUp(self):
        pass

    def test_invalid_input_log_list(self):
        _, err = diag_root_cluster({})
        self.assertIn("Invalid parameter type for 'input_log_list', it should be 'list'.", err[0])

    def test_common_use(self):
        rc_parser_list = [
            {
                '3199': {
                    'base': {'logic_device_id': '0', 'phy_device_id': '0', 'device_ip': '', 'server_id': 'x.x.x.x',
                             'rank_map': {'xxxx1': {'rank_num': 1, 'rank_id': '0'}}, 'root_list': ['xxxx1'],
                             'timeout_param': {}
                             },
                    'error': {'first_error_time': '2024-04-08-02:55:04.903511',
                              'first_error_module': 'HCCL',
                              'timeout_error_events_list': [
                                  {'error_type': 'Notify', 'error_time': '2024-04-08-02:55:04.903511',
                                   'key_info': '', 'identifier': '', 'tag': 'AllReduce_x.x.x.x%enp11111',
                                   'index': '3', 'remote_rank': ''}
                              ],
                              'cqe_links': [], 'cluster_exception': {}},
                    'show_logs': {'normal': [], 'error': []},
                    'plog_parsed_name': 'plog-parser-3199-1.log',
                    'start_train_time': '2024-04-08-02:55:04.903511',
                    'end_train_time': '2024-04-08-02:55:04.903511',
                    'start_resumable_training_time': '0000-01-01-00:00:00.000000',
                    'recovery_success_time': '0000-01-01-00:00:00.000000',
                    'lagging_time': '0000-01-01-00:00:00.000000'
                }
            }
        ]
        rc_diag_result, _ = diag_root_cluster(rc_parser_list)
        fault_code = rc_diag_result.get("fault_description", {}).get("code", "")
        target_code = 111
        self.assertEqual(fault_code, target_code)
        rc_diag_filed_list = ["analyze_success", "fault_description", "root_cause_device", "device_link",
                              "remote_link", "first_error_device", "last_error_device"]
        for field in rc_diag_filed_list:
            self.assertIn(field, rc_diag_result.keys())

    def tearDown(self):
        pass


class TestRootClusterMindIEDiagnosis(unittest.TestCase):
    RC_PARSE_BASE = {
        '1': {
            'base': {
                'logic_device_id': '0',
                'phy_device_id': '0',
                'device_ip': '1.0.0.0',
                'server_id': 'x.x.x.x',
                'server_name': 'x.x.x.x',
                'rank_map': {
                    'xxxx1':
                        {'rank_num': 3, 'rank_id': '0'}
                },
                'root_list': ['xxxx1'],
                'timeout_param': {}
            },
            'error': {},
            'show_logs': {'normal': [], 'error': []},
            'plog_parsed_name': 'plog-parser-3199-1.log',
            'start_train_time': '2024-04-08-02:55:04.903511',
            'end_train_time': '2024-04-08-02:55:04.903511',
            'start_resumable_training_time': '0000-01-01-00:00:00.000000',
            'recovery_success_time': '0000-01-01-00:00:00.000000',
            'lagging_time': '0000-01-01-00:00:00.000000'
        },
        '2': {
            'base': {
                'logic_device_id': '1',
                'phy_device_id': '1',
                'device_ip': '2.0.0.0',
                'server_id': 'x.x.x.x',
                'server_name': 'x.x.x.x',
                'rank_map': {
                    'xxxx1':
                        {'rank_num': 3, 'rank_id': '1'}
                },
                'root_list': ['xxxx1'],
                'timeout_param': {}
            },
            'error': {},
            'show_logs': {'normal': [], 'error': []},
            'plog_parsed_name': 'plog-parser-3199-1.log',
            'start_train_time': '2024-04-08-02:55:04.903511',
            'end_train_time': '2024-04-08-02:55:04.903511',
            'start_resumable_training_time': '0000-01-01-00:00:00.000000',
            'recovery_success_time': '0000-01-01-00:00:00.000000',
            'lagging_time': '0000-01-01-00:00:00.000000'
        },
        '3': {
            'base': {
                'logic_device_id': '2',
                'phy_device_id': '2',
                'device_ip': '3.0.0.0',
                'server_id': 'x.x.x.x',
                'server_name': 'x.x.x.x',
                'rank_map': {
                    'xxxx1':
                        {'rank_num': 3, 'rank_id': '2'}
                },
                'root_list': ['xxxx1'],
                'timeout_param': {}
            },
            'error': {},
            'show_logs': {'normal': [], 'error': []},
            'plog_parsed_name': 'plog-parser-3199-1.log',
            'start_train_time': '2024-04-08-02:55:04.903511',
            'end_train_time': '2024-04-08-02:55:04.903511',
            'start_resumable_training_time': '0000-01-01-00:00:00.000000',
            'recovery_success_time': '0000-01-01-00:00:00.000000',
            'lagging_time': '0000-01-01-00:00:00.000000'
        }
    }

    def setUp(self):
        pass

    def test_pull_kv_error(self):
        rc_parser_list = [{
            'mindie': True,
            'link_error_info_map': {},
            'pull_kv_error_map': {
                '3.0.0.0': ['6.0.0.0']
            }
        }, self.RC_PARSE_BASE]
        rc_diag_result, _ = diag_root_cluster(rc_parser_list)
        fault_code = rc_diag_result.get("fault_description", {}).get("code", "")
        target_code = 129
        self.assertEqual(fault_code, target_code)

    def test_link_error(self):
        rc_parser_list = [{
            'mindie': True,
            'link_error_info_map': {
                '2.0.0.0': ['1.0.0.0'],
                '1.0.0.0': ['2.0.0.0']
            },
            'pull_kv_error_map': {}
        }, self.RC_PARSE_BASE]
        rc_diag_result, _ = diag_root_cluster(rc_parser_list)
        fault_code = rc_diag_result.get("fault_description", {}).get("code", "")
        target_code = 102
        self.assertEqual(fault_code, target_code)
        self.assertIn("x.x.x.x device-0", rc_diag_result.get("root_cause_device", []))
        self.assertIn("x.x.x.x device-1", rc_diag_result.get("root_cause_device", []))
        self.assertNotIn("x.x.x.x device-3", rc_diag_result.get("root_cause_device", []))

    def test_one_link_error(self):
        rc_parser_list = [{
            'mindie': True,
            'link_error_info_map': {
                '1.0.0.0': ['2.0.0.0']
            },
            'pull_kv_error_map': {}
        }, self.RC_PARSE_BASE]
        rc_diag_result, _ = diag_root_cluster(rc_parser_list)
        fault_code = rc_diag_result.get("fault_description", {}).get("code", "")
        target_code = 102
        self.assertEqual(fault_code, target_code)
        self.assertNotIn("x.x.x.x device-0", rc_diag_result.get("root_cause_device", []))
        self.assertIn("x.x.x.x device-1", rc_diag_result.get("root_cause_device", []))
        self.assertNotIn("x.x.x.x device-3", rc_diag_result.get("root_cause_device", []))

    def tearDown(self):
        pass
