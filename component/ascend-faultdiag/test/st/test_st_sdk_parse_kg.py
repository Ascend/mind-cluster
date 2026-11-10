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
import time
import unittest

from ascend_fd import (
    diag_knowledge_graph,
    diag_root_cluster,
    parse_fault_type,
    parse_knowledge_graph,
    parse_root_cluster,
)
from ascend_fd.utils.regular_table import (
    AMCT_SOURCE,
    BMC_APP_DUMP_SOURCE,
    BMC_DEVICE_DUMP_SOURCE,
    BMC_LOG_DUMP_SOURCE,
    BMC_SOURCE,
    CANN_DEVICE_SOURCE,
    CANN_PLOG_SOURCE,
    DEVICEPLUGIN_SOURCE,
    DOCKER_RUNTIME_SOURCE,
    LCNE_SOURCE,
    MINDIE_SOURCE,
    NODEDLOG_SOURCE,
    NPU_DEVICE_SOURCE,
    NPU_EXPORTER_SOURCE,
    NPU_OS_SOURCE,
    OS_DEMESG_SOURCE,
    OS_SOURCE,
    OS_SYSMON_SOURCE,
    OS_VMCORE_DMESG_SOURCE,
    TRAIN_LOG_SOURCE,
    VOLCANO_CONTROLLER_SOURCE,
    VOLCANO_SCHEDULER_SOURCE,
)

PLOG_LINE = (
    "[ERROR] RUNTIME(27514,python):2024-03-16-15:52:17.572.654 [task_info.cc:5261]28555 "
    "PrintErrorInfoForNotifyWaitTask:[COMP][DEFAULT]Notify wait execute failed, device_id=1"
)
PLOG_SECOND_LINE = "[ERROR] HCCL(11262,python3):2023-01-01-02:01:00.000.000 There is no subscribe thread."
MINDIE_LINE = (
    "[2024-11-05 12:00:00.123456] [11] [1] [MS] [1] [ERROR] : [MIE03E400008] [HttpServer] Http server on timeout."
)
MINDIE_SECOND_LINE = (
    "[2024-11-05 12:00:00.123456] [11] [1] [MS] [1] [ERROR] : [MIE03E400002] [HttpServer] Http server on timeout."
)
NOTIFY_ERR_LOG = (
    "[ERROR] HCCL(69395,python3):2024-04-08-02:55:04.903.511 [task_exception_handler.cc:452]"
    "[69395][TaskExceptionHandler][Callback]Task run failed, base information is streamID:[3]"
)
OS_MODIFICATION_TIME = "2022-10-10 23:00:00.000000"
CUSTOM_ENTITY = {
    "entity_one": {
        "attribute.class": "Software",
        "attribute.component": "NPU",
        "attribute.module": "NPU",
        "attribute.cause_zh": "原因",
        "attribute.description_zh": "描述",
        "attribute.suggestion_zh": "建议",
        "source_file": "Check",
        "rule": [],
        "regex.in": [["Machine Check"]],
    },
    "entity_two": {
        "attribute.class": "Software",
        "attribute.component": "NPU",
        "attribute.module": "NPU",
        "attribute.cause_zh": "原因",
        "attribute.description_zh": "描述",
        "attribute.suggestion_zh": "建议",
        "source_file": "Check",
        "rule": [],
        "regex.in": [["Second Check"]],
    },
}


def build_item(item_type, lines, **kwargs):
    item = {"item_type": item_type, "log_lines": lines}
    item.update(kwargs)
    return item


def build_input(item_type, lines, server="1.1.1.1", **item_kwargs):
    return {"log_domain": {"server": server}, "log_items": [build_item(item_type, lines, **item_kwargs)]}


def get_root_causes(results):
    if not results:
        return {}
    return results[0].get("fault", [{}])[0].get("response", {}) or {}


def get_events(results):
    events = []
    for device_info in get_root_causes(results).values():
        for code, code_info in device_info.get("root_causes", {}).items():
            for event in code_info.get("events_attribute", []):
                events.append((code, event))
    return events


def get_codes(results):
    return [code for code, _ in get_events(results)]


def get_occur_time(results, code):
    for event_code, event in get_events(results):
        if event_code == code:
            return event.get("occur_time", "")
    return ""


class STParseKnowledgeGraphItemType(unittest.TestCase):
    def assert_item_type_parsed(self, item_type, lines, expected_code, extra_items=None, **item_kwargs):
        payload = build_input(item_type, lines, **item_kwargs)
        payload["log_items"].extend(extra_items or [])
        results, err_msg_list = parse_knowledge_graph([payload])
        self.assertEqual(err_msg_list, [])
        self.assertEqual(len(results), 1)
        self.assertIn(expected_code, get_codes(results))
        return results

    def assert_item_type_without_event(self, item_type, lines, **item_kwargs):
        """Assert the item type is accepted even though no knowledge rule is configured for it."""
        results, err_msg_list = parse_knowledge_graph([build_input(item_type, lines, **item_kwargs)])
        self.assertEqual(err_msg_list, [])
        self.assertEqual(len(results), 1)
        self.assertEqual(get_events(results), [])

    def test_func_0030_noded_log(self):
        """Ascend_Fd_SDK_Func_0030"""
        line = (
            "[INFO]     2999/10/01 03:39:17.760916 305     ipmimonitor/ipmi_monitor.go:206    "
            "get fault event, [device type: BMC]-[device id: 255]-[error code: 2C000031]"
        )
        self.assert_item_type_parsed(NODEDLOG_SOURCE, [line], "0x2C000031")

    def test_func_0031_device_plugin(self):
        """Ascend_Fd_SDK_Func_0031"""
        line = (
            '[WARN]     2024/01/10 03:30:38.643330 17      deviceswitch/ascend_switch.go:190    '
            'switch subscribe got fault:deviceswitch.SwitchFaultEvent{EventType:0x4, FaultID:0xf1ff09, '
            'AssembledFaultCode:"[0x00f1ff09,155912,npu,na]", PeerPortDevice:0x1, PeerPortId:0x2, '
            'SwitchChipId:0x0, SwitchPortId:0x0, Severity:0x0, Assertion:0x1, EventSerialNum:0, '
            'NotifySerialNum:0, AlarmRaisedTime:0, AdditionalParam:"", AdditionalInfo:""}, '
            'faultCode:[0x00f1ff09,155912,npu,na]'
        )
        self.assert_item_type_parsed(DEVICEPLUGIN_SOURCE, [line], "Comp_Switch_L1_06")

    def test_func_0032_volcano_scheduler(self):
        """Ascend_Fd_SDK_Func_0032"""
        line = "E0111 03:29:15.991928       1 reschedule.go:533]    Failed to get plugin volcano-npu_xxx_linux-xxx"
        self.assert_item_type_parsed(VOLCANO_SCHEDULER_SOURCE, [line], "AISW_MindX_Volcano_03")

    def test_func_0033_volcano_controller(self):
        """Ascend_Fd_SDK_Func_0033"""
        line = "E0111 03:29:15.700191       1 queue_controller.go:258]    Failed to create pod for Job"
        self.assert_item_type_parsed(VOLCANO_CONTROLLER_SOURCE, [line], "AISW_MindX_Volcano_10")

    def test_func_0034_docker_runtime(self):
        """Ascend_Fd_SDK_Func_0034"""
        line = (
            "[INFO]     2024/01/10 03:29:09.354927 1       hwlog/api.go:108    failed to check files, "
            "owner not right /usr/bin/runc xxxx"
        )
        self.assert_item_type_parsed(DOCKER_RUNTIME_SOURCE, [line], "AISW_MindX_Docker_Runtime_02")

    def test_func_0035_npu_exporter(self):
        """Ascend_Fd_SDK_Func_0035"""
        line = (
            "[ERROR]    2024/01/10 02:29:39.034681 60      collector/npu_collector.go:1924    "
            "deviceManager init failed, prepare dcmi failed, err: cannot found valid driver lib, "
            "fromEnv: lib path is invalid, [/usr/local/Ascend/driver/lib64/driver/libdcmi.so: "
            "check uid or mode failed; /usr/local/dcmi/libdcmi.so: check uid or mode failed;]"
        )
        self.assert_item_type_parsed(NPU_EXPORTER_SOURCE, [line], "AISW_MindX_Npu_Exporter_01")

    def test_func_0036_os(self):
        """Ascend_Fd_SDK_Func_0036"""
        line = "Mar 15 03:40:50 sha2uvp-gert01 kernel: [16794.247724] detected stalls on"
        self.assert_item_type_parsed(OS_SOURCE, [line], "Comp_OS_Kernel_CPU_06")

    def test_func_0037_os_vmcore_dmesg(self):
        """Ascend_Fd_SDK_Func_0037"""
        line = "[  293.039556] NMI watchdog: Watchdog detected hard LOCKUP on cpu 34"
        self.assert_item_type_parsed(OS_VMCORE_DMESG_SOURCE, [line], "Comp_OS_Kernel_CPU_04")

    def test_func_0038_os_sysmon(self):
        """Ascend_Fd_SDK_Func_0038"""
        line = "2024-08-30T17:01:48.559312+00:00|info|sysmonitor[12706]: zombie process"
        self.assert_item_type_parsed(OS_SYSMON_SOURCE, [line], "Comp_OS_Kernel_CPU_11")

    def test_func_0039_os_dmesg(self):
        """Ascend_Fd_SDK_Func_0039"""
        line = "[Mon Sep 30 15:00:00 2024] systemd[1]: license-manager.service: I/O error."
        self.assert_item_type_parsed(OS_DEMESG_SOURCE, [line], "Comp_OS_HW_Stor_02")

    def test_func_0040_npu_device(self):
        """Ascend_Fd_SDK_Func_0040"""
        line = (
            "[EVENT] IMP(-2,IMP): 2024-12-04-08:21:54.994.714 70 (device_id:1 die_id:0) [IMP] present 1, "
            "mac_link 1, pcs_link 1, rf_lf 1, pcs_err_cnt 5, pcs_64_66b 4681"
        )
        self.assert_item_type_parsed(NPU_DEVICE_SOURCE, [line], "Comp_Network_Custom_11")

    def test_func_0041_npu_os(self):
        """Ascend_Fd_SDK_Func_0041"""
        line = (
            "[ERROR] KERNEL(3551,sklogd):2021-10-05-23:07:08.350.235 [klogd.c:242][583733.716915] event_id=0x80c78008"
        )
        self.assert_item_type_parsed(NPU_OS_SOURCE, [line], "0x80C78008")

    def test_func_0042_train_log(self):
        """Ascend_Fd_SDK_Func_0042"""
        self.assert_item_type_parsed(TRAIN_LOG_SOURCE, ["get socket timeout"], "AISW_CANN_ERRMSG_Custom_04")

    def test_func_0043_cann_amct(self):
        """Ascend_Fd_SDK_Func_0043"""
        line = (
            "2024-10-23 10:47:11,976 - INFO - [AMCT]:[AMCT]: Layer batchmatmul_2 does not support "
            "asymmetric quant, set symmetric, No layer support in the graph"
        )
        self.assert_item_type_parsed(AMCT_SOURCE, [line], "AISW_CANN_AMCT_ALL_001")

    def test_func_0044_cann_device(self):
        """Ascend_Fd_SDK_Func_0044"""
        line = "[ERROR] AICPU(25602,aicpu_scheduler):2023-12-01-01:47:18.061.111 load stream and task failed."
        plog_item = build_item(
            CANN_PLOG_SOURCE,
            [PLOG_LINE],
            path="debug/plog/plog-51_20231129113051449.log",
            device_id=3,
        )
        self.assert_item_type_parsed(
            CANN_DEVICE_SOURCE,
            [line],
            "AISW_CANN_AICPU_04",
            extra_items=[plog_item],
            path="debug/device-2/device-51_20231129113051449.log",
        )

    def test_func_0045_lcne_log(self):
        """Ascend_Fd_SDK_Func_0045"""
        line = (
            "Aug 12 2025 08:53:46+08:00 Z3_22R_99_30 %%01CLI/5/CMDRECORD(s):CID=0x80ca2713;"
            "Recorded command information. alarmID=0x00f1ff09"
        )
        self.assert_item_type_parsed(LCNE_SOURCE, [line], "Comp_Switch_L1_02")

    def test_func_0046_bmc_log(self):
        """Ascend_Fd_SDK_Func_0046"""
        self.assert_item_type_without_event(BMC_SOURCE, ["2024-01-01 00:00:00, ERROR, BMC fault occurred"])

    def test_func_0047_bmc_app_dump_log(self):
        """Ascend_Fd_SDK_Func_0047"""
        self.assert_item_type_without_event(
            BMC_APP_DUMP_SOURCE, ["2024-01-01 00:00:00, ERROR, app dump fault occurred"]
        )

    def test_func_0048_bmc_device_dump_log(self):
        """Ascend_Fd_SDK_Func_0048"""
        self.assert_item_type_without_event(
            BMC_DEVICE_DUMP_SOURCE, ["2024-01-01 00:00:00, ERROR, device dump fault occurred"]
        )

    def test_func_0049_bmc_log_dump_log(self):
        """Ascend_Fd_SDK_Func_0049"""
        line = "2026-01-01T00:00:00.123456+08:00 0x080000E7 Asserted"
        self.assert_item_type_parsed(BMC_LOG_DUMP_SOURCE, [line], "0x080000E7")

    def test_func_0052_device_plugin_recovery(self):
        """Ascend_Fd_SDK_Func_0052"""
        lines = [
            (
                '[WARN]     2024/01/10 03:30:38.643330 17      deviceswitch/ascend_switch.go:190    '
                'switch subscribe got fault:deviceswitch.SwitchFaultEvent{EventType:0x4, FaultID:0xf1ff09, '
                'AssembledFaultCode:"[0x00f1ff09,155912,npu,na]", PeerPortDevice:0x1, PeerPortId:0x2, '
                'SwitchChipId:0x0, SwitchPortId:0x0, Severity:0x0, Assertion:0x1, EventSerialNum:0, '
                'NotifySerialNum:0, AlarmRaisedTime:0, AdditionalParam:"", AdditionalInfo:""}, '
                'faultCode:[0x00f1ff09,155912,npu,na]'
            ),
            (
                '[WARN]     2024/01/10 03:31:38.643290 17      deviceswitch/ascend_switch.go:190    '
                'switch subscribe got fault:deviceswitch.SwitchFaultEvent{EventType:0x4, FaultID:0xf103b0, '
                'AssembledFaultCode:"[0x00f103b0,155907,na,na]", PeerPortDevice:0x0, PeerPortId:0x2, '
                'SwitchChipId:0x0, SwitchPortId:0x0, Severity:0x0, Assertion:0x1, EventSerialNum:0, '
                'NotifySerialNum:0, AlarmRaisedTime:0, AdditionalParam:"", AdditionalInfo:""}, '
                'faultCode:[0x00f103b0,155907,na,na]'
            ),
            (
                '[WARN]     2024/01/10 03:32:38.643290 17      deviceswitch/ascend_switch.go:190    '
                'switch subscribe got fault:deviceswitch.SwitchFaultEvent{EventType:0x4, FaultID:0xf103b0, '
                'AssembledFaultCode:"[0x00f103b0,155907,na,na]", PeerPortDevice:0x0, PeerPortId:0x2, '
                'SwitchChipId:0x0, SwitchPortId:0x0, Severity:0x0, Assertion:0x0, EventSerialNum:0, '
                'NotifySerialNum:0, AlarmRaisedTime:0, AdditionalParam:"", AdditionalInfo:""}, '
                'faultCode:[0x00f103b0,155907,na,na]'
            ),
        ]
        results = self.assert_item_type_parsed(DEVICEPLUGIN_SOURCE, lines, "Comp_Switch_L1_06")
        self.assertNotIn("Comp_Switch_L1_01", get_codes(results))


class STParseKnowledgeGraphFunctional(unittest.TestCase):
    def test_func_0012_same_fault_multiple_devices(self):
        """Ascend_Fd_SDK_Func_0012"""
        input_list = [
            build_input(CANN_PLOG_SOURCE, [PLOG_LINE], device_id=0),
            build_input(CANN_PLOG_SOURCE, [PLOG_LINE], device_id=1),
        ]
        input_list[0]["log_items"].extend(input_list[1]["log_items"])
        results, err_msg_list = parse_knowledge_graph([input_list[0]])
        self.assertEqual(err_msg_list, [])
        devices = [event.get("source_device") for code, event in get_events(results) if code == "AISW_CANN_Runtime_021"]
        self.assertEqual(sorted(devices), ["0", "1"])

    def test_func_0013_same_fault_multiple_empty_device_id(self):
        """Ascend_Fd_SDK_Func_0013"""
        input_list = {
            "log_domain": {"server": "1.1.1.1"},
            "log_items": [build_item(CANN_PLOG_SOURCE, [PLOG_LINE]), build_item(CANN_PLOG_SOURCE, [PLOG_LINE])],
        }
        results, err_msg_list = parse_knowledge_graph([input_list])
        self.assertEqual(err_msg_list, [])
        devices = [event.get("source_device") for code, event in get_events(results)]
        self.assertEqual(set(devices), {"Unknown"})

    def test_func_0016_one_device_multiple_faults(self):
        """Ascend_Fd_SDK_Func_0016"""
        results, err_msg_list = parse_knowledge_graph(
            [build_input(CANN_PLOG_SOURCE, [PLOG_LINE, PLOG_SECOND_LINE], device_id=0)]
        )
        self.assertEqual(err_msg_list, [])
        codes = get_codes(results)
        self.assertIn("AISW_CANN_Runtime_021", codes)
        self.assertIn("AISW_CANN_DRV_ESCH_008", codes)

    def test_func_0018_same_fault_in_different_servers(self):
        """Ascend_Fd_SDK_Func_0018"""
        payload = [
            build_input(CANN_PLOG_SOURCE, [PLOG_LINE], server="server-1"),
            build_input(CANN_PLOG_SOURCE, [PLOG_LINE], server="server-2"),
        ]
        results, err_msg_list = parse_knowledge_graph(payload)
        self.assertEqual(err_msg_list, [])
        self.assertEqual(len(results), 2)
        for result in results:
            self.assertIn("AISW_CANN_Runtime_021", get_codes([result]))

    def test_func_0019_path_present_without_device_id(self):
        """Ascend_Fd_SDK_Func_0019"""
        results = parse_knowledge_graph(
            [build_input(CANN_PLOG_SOURCE, [PLOG_LINE], path="process_log/plog-51_20231129113051449.log")]
        )[0]
        self.assertIn("AISW_CANN_Runtime_021", get_codes(results))

    def test_func_0014_mixed_plog_and_mindie(self):
        """Ascend_Fd_SDK_Func_0014"""
        payload = {
            "log_domain": {"server": "1.1.1.1"},
            "log_items": [
                build_item(CANN_PLOG_SOURCE, [PLOG_LINE], device_id=0),
                build_item(MINDIE_SOURCE, [MINDIE_LINE], device_id=1),
            ],
        }
        results, err_msg_list = parse_knowledge_graph([payload])
        self.assertEqual(err_msg_list, [])
        codes = get_codes(results)
        self.assertIn("AISW_CANN_Runtime_021", codes)
        self.assertIn("AISW_MindIE_MS_HttpServer_01", codes)

    def test_func_0015_mindie_one_device_multiple_faults(self):
        """Ascend_Fd_SDK_Func_0015"""
        results, err_msg_list = parse_knowledge_graph(
            [build_input(MINDIE_SOURCE, [MINDIE_LINE, MINDIE_SECOND_LINE], device_id=0)]
        )
        self.assertEqual(err_msg_list, [])
        codes = get_codes(results)
        self.assertIn("AISW_MindIE_MS_HttpServer_01", codes)
        self.assertIn("AISW_MindIE_MS_HttpServer_02", codes)

    def test_reliability_0013_graph_type_mindie_fault(self):
        """Ascend_Fd_SDK_Reliability_0013"""
        results, err_msg_list = parse_knowledge_graph([build_input(MINDIE_SOURCE, [MINDIE_LINE], device_id=0)])
        self.assertEqual(err_msg_list, [])
        self.assertIn("AISW_MindIE_MS_HttpServer_01", get_codes(results))

    def test_reliability_0011_mixed_types_with_foreign_log(self):
        """Ascend_Fd_SDK_Reliability_0011"""
        payload = {
            "log_domain": {"server": "1.1.1.1"},
            "log_items": [
                build_item(CANN_PLOG_SOURCE, [MINDIE_LINE], device_id=0),
                build_item(MINDIE_SOURCE, [PLOG_LINE], device_id=1),
            ],
        }
        results, err_msg_list = parse_knowledge_graph([payload])
        self.assertEqual(err_msg_list, [])
        self.assertEqual(len(results), 1)

    def test_func_0050_custom_type_multiple_faults(self):
        """Ascend_Fd_SDK_Func_0050"""
        payload = [
            build_input(
                "Check", ["Machine Check", "Second Check"], device_id=0, modification_time="2025-01-01 20:22:33.999999"
            )
        ]
        results, err_msg_list = parse_knowledge_graph(payload, CUSTOM_ENTITY)
        self.assertEqual(err_msg_list, [])
        codes = get_codes(results)
        self.assertIn("entity_one", codes)
        self.assertIn("entity_two", codes)

    def test_func_0061_custom_type_modification_time(self):
        """Ascend_Fd_SDK_Func_0061"""
        expected_time = "2025-01-01 20:22:33.999999"
        payload = [build_input("Check", ["Machine Check"], modification_time=expected_time)]
        results, err_msg_list = parse_knowledge_graph(payload, CUSTOM_ENTITY)
        self.assertEqual(err_msg_list, [])
        self.assertEqual(get_occur_time(results, "entity_one"), expected_time)

    def test_func_0062_one_time_custom_fault_in_multi_items(self):
        """Ascend_Fd_SDK_Func_0062"""
        payload = {
            "log_domain": {"server": "1.1.1.1"},
            "log_items": [build_item("Check", ["Machine Check"]), build_item(MINDIE_SOURCE, [MINDIE_LINE])],
        }
        results, err_msg_list = parse_knowledge_graph([payload], CUSTOM_ENTITY)
        self.assertEqual(err_msg_list, [])
        codes = get_codes(results)
        self.assertIn("entity_one", codes)
        self.assertIn("AISW_MindIE_MS_HttpServer_01", codes)

    def test_reliability_0023_custom_type_not_in_scope(self):
        """Ascend_Fd_SDK_Reliability_0023"""
        payload = [build_input("NotSupportedType", ["some fault"])]
        results, _ = parse_knowledge_graph(payload, CUSTOM_ENTITY)
        # An item type outside the custom entity scope produces no fault information.
        self.assertEqual(get_codes(results), [])

    def test_func_0051_multiple_legal_item_types(self):
        """Ascend_Fd_SDK_Func_0051"""
        payload = {
            "log_domain": {"server": "1.1.1.1"},
            "log_items": [
                build_item(MINDIE_SOURCE, [MINDIE_LINE], device_id=0),
                build_item(CANN_PLOG_SOURCE, [PLOG_LINE], device_id=1),
            ],
        }
        results, err_msg_list = parse_knowledge_graph([payload])
        self.assertEqual(err_msg_list, [])
        codes = get_codes(results)
        self.assertIn("AISW_MindIE_MS_HttpServer_01", codes)
        self.assertIn("AISW_CANN_Runtime_021", codes)

    def test_func_0053_train_log_multiline(self):
        """Ascend_Fd_SDK_Func_0053"""
        lines = ["The parameters number of the function is", "but the number of provided arguments is"]
        results, err_msg_list = parse_knowledge_graph([build_input(TRAIN_LOG_SOURCE, lines)])
        self.assertEqual(err_msg_list, [])
        self.assertIn("AISW_MindSpore_Compiler_09", get_codes(results))

    def test_func_0054_modification_time_validation(self):
        """Ascend_Fd_SDK_Func_0054"""
        expected_time = "2025-01-01 20:22:33.000000"
        valid = [build_input(TRAIN_LOG_SOURCE, ["get socket timeout"], modification_time="2025-01-01 20:22:33")]
        results, err_msg_list = parse_knowledge_graph(valid)
        self.assertEqual(err_msg_list, [])
        self.assertEqual(get_occur_time(results, "AISW_CANN_ERRMSG_Custom_04"), expected_time)
        invalid = [build_input(TRAIN_LOG_SOURCE, ["get socket timeout"], modification_time="invalid")]
        _, invalid_err = parse_knowledge_graph(invalid)
        self.assertIn("modification_time is invalid", invalid_err[0])

    def test_func_0055_train_log_modification_time(self):
        """Ascend_Fd_SDK_Func_0055"""
        expected_time = "2022-10-10 23:00:00.000000"
        payload = [build_input(TRAIN_LOG_SOURCE, ["get socket timeout"], modification_time=expected_time)]
        results, err_msg_list = parse_knowledge_graph(payload)
        self.assertEqual(err_msg_list, [])
        self.assertEqual(get_occur_time(results, "AISW_CANN_ERRMSG_Custom_04"), expected_time)

    def test_func_0056_mindie_modification_time(self):
        """Ascend_Fd_SDK_Func_0056"""
        # The MindIE log carries its own fault time, which takes precedence over modification_time.
        payload = [build_input(MINDIE_SOURCE, [MINDIE_LINE], device_id=0, modification_time=OS_MODIFICATION_TIME)]
        results, err_msg_list = parse_knowledge_graph(payload)
        self.assertEqual(err_msg_list, [])
        self.assertEqual(get_occur_time(results, "AISW_MindIE_MS_HttpServer_01"), "2024-11-05 12:00:00.123456")

    def test_func_0057_os_modification_time(self):
        """Ascend_Fd_SDK_Func_0057"""
        line = "Mar 15 03:40:50 sha2uvp-gert01 kernel: [16794.247724] detected stalls on"
        payload = [build_input(OS_SOURCE, [line], modification_time=OS_MODIFICATION_TIME)]
        results, err_msg_list = parse_knowledge_graph(payload)
        self.assertEqual(err_msg_list, [])
        self.assertEqual(get_occur_time(results, "Comp_OS_Kernel_CPU_06"), "2022-03-15 03:40:50.000000")

    def test_func_0058_vmcore_modification_time(self):
        """Ascend_Fd_SDK_Func_0058"""
        line = "[  293.039556] NMI watchdog: Watchdog detected hard LOCKUP on cpu 34"
        payload = [build_input(OS_VMCORE_DMESG_SOURCE, [line], modification_time=OS_MODIFICATION_TIME)]
        results, err_msg_list = parse_knowledge_graph(payload)
        self.assertEqual(err_msg_list, [])
        self.assertEqual(get_occur_time(results, "Comp_OS_Kernel_CPU_04"), OS_MODIFICATION_TIME)

    def test_func_0059_sysmon_modification_time(self):
        """Ascend_Fd_SDK_Func_0059"""
        line = "2024-08-30T17:01:48.559312+00:00|info|sysmonitor[12706]: zombie process"
        payload = [build_input(OS_SYSMON_SOURCE, [line], modification_time=OS_MODIFICATION_TIME)]
        results, err_msg_list = parse_knowledge_graph(payload)
        self.assertEqual(err_msg_list, [])
        self.assertNotEqual(get_occur_time(results, "Comp_OS_Kernel_CPU_11"), OS_MODIFICATION_TIME)

    def test_func_0060_dmesg_modification_time(self):
        """Ascend_Fd_SDK_Func_0060"""
        line = "[Mon Sep 30 15:00:00 2024] systemd[1]: license-manager.service: I/O error."
        payload = [build_input(OS_DEMESG_SOURCE, [line], modification_time=OS_MODIFICATION_TIME)]
        results, err_msg_list = parse_knowledge_graph(payload)
        self.assertEqual(err_msg_list, [])
        self.assertEqual(get_occur_time(results, "Comp_OS_HW_Stor_02"), "2024-09-30 15:00:00.000000")

    def test_reliability_0022_plog_modification_time(self):
        """Ascend_Fd_SDK_Reliability_0022"""
        payload = [build_input(CANN_PLOG_SOURCE, [PLOG_LINE], modification_time=OS_MODIFICATION_TIME)]
        results, err_msg_list = parse_knowledge_graph(payload)
        self.assertEqual(err_msg_list, [])
        self.assertNotEqual(get_occur_time(results, "AISW_CANN_Runtime_021"), OS_MODIFICATION_TIME)


class STParseKnowledgeGraphReliability(unittest.TestCase):
    def test_reliability_0008_duplicate_device_and_item_type(self):
        """Ascend_Fd_SDK_Reliability_0008"""
        payload = {
            "log_domain": {"server": "1.1.1.1"},
            "log_items": [
                build_item(CANN_PLOG_SOURCE, [PLOG_LINE], device_id=0),
                build_item(CANN_PLOG_SOURCE, [PLOG_LINE], device_id=0),
            ],
        }
        results, err_msg_list = parse_knowledge_graph([payload])
        self.assertEqual(err_msg_list, [])
        self.assertEqual(get_codes(results).count("AISW_CANN_Runtime_021"), 1)

    def test_reliability_0009_empty_list(self):
        """Ascend_Fd_SDK_Reliability_0009"""
        results, err_msg_list = parse_knowledge_graph([])
        self.assertEqual(results, [])
        self.assertEqual(err_msg_list, [])

    def test_reliability_0010_path_not_standard_without_device_id(self):
        """Ascend_Fd_SDK_Reliability_0010"""
        payload = [build_input(CANN_PLOG_SOURCE, [PLOG_LINE], path="some/not/standard/path.log")]
        results, err_msg_list = parse_knowledge_graph(payload)
        self.assertEqual(err_msg_list, [])
        self.assertIn("AISW_CANN_Runtime_021", get_codes(results))

    def test_reliability_0012_same_server_same_path_different_lines(self):
        """Ascend_Fd_SDK_Reliability_0012"""
        payload = {
            "log_domain": {"server": "1.1.1.1"},
            "log_items": [
                build_item(CANN_PLOG_SOURCE, [PLOG_LINE], path="a/plog-1_20240101000000.log"),
                build_item(CANN_PLOG_SOURCE, [PLOG_SECOND_LINE], path="a/plog-1_20240101000000.log"),
            ],
        }
        results, err_msg_list = parse_knowledge_graph([payload])
        self.assertEqual(err_msg_list, [])
        self.assertEqual(len(results), 1)

    def test_reliability_0020_same_server_different_path_and_lines(self):
        """Ascend_Fd_SDK_Reliability_0020"""
        payload = {
            "log_domain": {"server": "1.1.1.1"},
            "log_items": [
                build_item(CANN_PLOG_SOURCE, [PLOG_LINE], path="a/plog-1_20240101000000.log"),
                build_item(CANN_PLOG_SOURCE, [PLOG_SECOND_LINE], path="b/plog-2_20240101000000.log"),
            ],
        }
        results, err_msg_list = parse_knowledge_graph([payload])
        self.assertEqual(err_msg_list, [])
        self.assertEqual(len(results), 1)

    def test_reliability_0021_multiple_items_with_illegal_item_type(self):
        """Ascend_Fd_SDK_Reliability_0021"""
        payload = {
            "log_domain": {"server": "1.1.1.1"},
            "log_items": [build_item(MINDIE_SOURCE, [MINDIE_LINE], device_id=0), build_item("IllegalType", ["fault"])],
        }
        results, err_msg_list = parse_knowledge_graph([payload])
        self.assertTrue(err_msg_list)
        self.assertIn("unsupported", err_msg_list[0])
        self.assertIn("AISW_MindIE_MS_HttpServer_01", get_codes(results))

    def test_perf_0003_256_servers_16_items_10_lines(self):
        """Ascend_Fd_SDK_Perf_0003"""
        payload = []
        for index in range(256):
            items = [build_item(CANN_PLOG_SOURCE, [PLOG_LINE] * 10, device_id=0) for _ in range(16)]
            payload.append({"log_domain": {"server": f"1.1.1.{index}"}, "log_items": items})
        start_time = time.time()
        results, err_msg_list = parse_knowledge_graph(payload)
        cost_seconds = time.time() - start_time
        self.assertEqual(err_msg_list, [])
        self.assertEqual(len(results), 256)
        self.assertLess(cost_seconds, 120)


class STParseKnowledgeGraphSecurity(unittest.TestCase):
    def assert_invalid(self, payload, expected):
        results, err_msg_list = parse_knowledge_graph(payload)
        self.assertEqual(results, [])
        self.assertTrue(err_msg_list)
        self.assertIn(expected, err_msg_list[0])

    def test_security_0029_server_empty(self):
        """Ascend_Fd_SDK_Security_0029"""
        payload = build_input(MINDIE_SOURCE, [MINDIE_LINE])
        payload["log_domain"]["server"] = ""
        self.assert_invalid([payload], "Empty value not allowed")

    def test_security_0035_server_type_not_str(self):
        """Ascend_Fd_SDK_Security_0035"""
        payload = build_input(MINDIE_SOURCE, [MINDIE_LINE])
        payload["log_domain"]["server"] = 1
        self.assert_invalid([payload], "expected str")

    def test_security_0036_item_type_type_not_str(self):
        """Ascend_Fd_SDK_Security_0036"""
        self.assert_invalid([build_input(1, [MINDIE_LINE])], "item_type")

    def test_security_0037_log_lines_type_not_list(self):
        """Ascend_Fd_SDK_Security_0037"""
        payload = build_input(MINDIE_SOURCE, MINDIE_LINE)
        self.assert_invalid([payload], "log_lines")

    def test_security_0038_path_type_not_str(self):
        """Ascend_Fd_SDK_Security_0038"""
        payload = build_input(MINDIE_SOURCE, [MINDIE_LINE], path=123)
        self.assert_invalid([payload], "path")

    def test_security_0039_device_id_type_not_int(self):
        """Ascend_Fd_SDK_Security_0039"""
        payload = build_input(MINDIE_SOURCE, [MINDIE_LINE], device_id="123")
        self.assert_invalid([payload], "device_id")

    def test_security_0030_log_domain_missing(self):
        """Ascend_Fd_SDK_Security_0030"""
        payload = build_input(MINDIE_SOURCE, [MINDIE_LINE])
        payload.pop("log_domain")
        self.assert_invalid([payload], "log_domain")

    def test_security_0031_server_missing(self):
        """Ascend_Fd_SDK_Security_0031"""
        payload = build_input(MINDIE_SOURCE, [MINDIE_LINE])
        payload["log_domain"].pop("server")
        self.assert_invalid([payload], "log_domain")

    def test_security_0032_log_item_missing(self):
        """Ascend_Fd_SDK_Security_0032"""
        payload = build_input(MINDIE_SOURCE, [MINDIE_LINE])
        payload.pop("log_items")
        self.assert_invalid([payload], "log_items")

    def test_security_0033_item_type_missing(self):
        """Ascend_Fd_SDK_Security_0033"""
        payload = {"log_domain": {"server": "1.1.1.1"}, "log_items": [{"log_lines": [MINDIE_LINE]}]}
        self.assert_invalid([payload], "item_type")

    def test_security_0034_log_lines_missing(self):
        """Ascend_Fd_SDK_Security_0034"""
        payload = {"log_domain": {"server": "1.1.1.1"}, "log_items": [{"item_type": MINDIE_SOURCE}]}
        self.assert_invalid([payload], "log_lines")


class STParseKnowledgeGraphIPv6(unittest.TestCase):
    def test_ipv6_func_0001_parse_mindie(self):
        """Ascend_MindCluster_Fd_IPv6_SDK_Func_0001"""
        payload = [
            {
                "log_domain": {"server": "2001:db8::1", "device": ["0"]},
                "log_items": [{"item_type": MINDIE_SOURCE, "log_lines": [MINDIE_LINE]}],
            }
        ]
        results, err_msg_list = parse_fault_type(payload)
        self.assertEqual(err_msg_list, [])
        self.assertTrue(results)

    def test_ipv6_func_0002_parse_root_cluster_multiple_instances(self):
        """Ascend_MindCluster_Fd_IPv6_SDK_Func_0002"""
        payload = [
            {
                "log_domain": {"server": "2001:db8::1", "instance_id": "inst1"},
                "log_items": [build_item("plog", [NOTIFY_ERR_LOG], pid=3199, device_id=0)],
            },
            {
                "log_domain": {"server": "2001:db8::1", "instance_id": "inst2"},
                "log_items": [build_item("plog", [NOTIFY_ERR_LOG], pid=3299, device_id=1)],
            },
        ]
        results, err_msg_list = parse_root_cluster(payload)
        self.assertEqual(err_msg_list, [])
        self.assertEqual(len(results), 2)

    def test_ipv6_func_0003_diag_root_cluster_from_parse_output(self):
        """Ascend_MindCluster_Fd_IPv6_SDK_Func_0003"""
        payload = [
            {
                "log_domain": {"server": "2001:db8::1", "instance_id": "inst1"},
                "log_items": [build_item("plog", [NOTIFY_ERR_LOG], pid=3199, device_id=0)],
            }
        ]
        parse_result, _ = parse_root_cluster(payload)
        diag_result, err_msg_list = diag_root_cluster(parse_result)
        self.assertEqual(err_msg_list, [])
        self.assertTrue(diag_result.get("analyze_success"))

    def test_ipv6_func_0004_parse_kg_full_format(self):
        """Ascend_MindCluster_Fd_IPv6_SDK_Func_0004"""
        server = "2001:0db8:0000:0000:0000:0000:0000:0001"
        payload = {
            "log_domain": {"server": server},
            "log_items": [
                build_item(CANN_PLOG_SOURCE, [PLOG_LINE], device_id=0),
                build_item(MINDIE_SOURCE, [MINDIE_LINE], device_id=1),
            ],
        }
        results, err_msg_list = parse_knowledge_graph([payload])
        self.assertEqual(err_msg_list, [])
        codes = get_codes(results)
        self.assertIn("AISW_CANN_Runtime_021", codes)
        self.assertIn("AISW_MindIE_MS_HttpServer_01", codes)

    def test_ipv6_func_0005_parse_kg_compressed_format(self):
        """Ascend_MindCluster_Fd_IPv6_SDK_Func_0005"""
        payload = {
            "log_domain": {"server": "2001:db8::1"},
            "log_items": [
                build_item(CANN_PLOG_SOURCE, [PLOG_LINE], device_id=0),
                build_item(MINDIE_SOURCE, [MINDIE_LINE], device_id=1),
            ],
        }
        results, err_msg_list = parse_knowledge_graph([payload])
        self.assertEqual(err_msg_list, [])
        self.assertIn("AISW_CANN_Runtime_021", get_codes(results))

    def test_ipv6_func_0006_cann_amct(self):
        """Ascend_MindCluster_Fd_IPv6_SDK_Func_0006"""
        line = (
            "2024-10-23 10:47:11,976 - INFO - [AMCT]:[AMCT]: Layer batchmatmul_2 does not support "
            "asymmetric quant, set symmetric, No layer support in the graph"
        )
        results, err_msg_list = parse_knowledge_graph([build_input(AMCT_SOURCE, [line], server="2001:db8::1")])
        self.assertEqual(err_msg_list, [])
        self.assertIn("AISW_CANN_AMCT_ALL_001", get_codes(results))

    def test_ipv6_func_0007_cann_device(self):
        """Ascend_MindCluster_Fd_IPv6_SDK_Func_0007"""
        line = "[ERROR] AICPU(25602,aicpu_scheduler):2023-12-01-01:47:18.061.111 load stream and task failed."
        results, err_msg_list = parse_knowledge_graph([build_input(CANN_DEVICE_SOURCE, [line], server="2001:db8::1")])
        self.assertEqual(err_msg_list, [])
        self.assertEqual(len(results), 1)

    def test_ipv6_func_0008_npu_device(self):
        """Ascend_MindCluster_Fd_IPv6_SDK_Func_0008"""
        line = (
            "[EVENT] IMP(-2,IMP): 2024-12-04-08:21:54.994.714 70 (device_id:1 die_id:0) [IMP] present 1, "
            "mac_link 1, pcs_link 1, rf_lf 1, pcs_err_cnt 5, pcs_64_66b 4681"
        )
        results, err_msg_list = parse_knowledge_graph([build_input(NPU_DEVICE_SOURCE, [line], server="2001:db8::1")])
        self.assertEqual(err_msg_list, [])
        self.assertIn("Comp_Network_Custom_11", get_codes(results))

    def test_ipv6_func_0009_device_plugin(self):
        """Ascend_MindCluster_Fd_IPv6_SDK_Func_0009"""
        line = (
            '[WARN]     2024/01/10 03:30:38.643330 17      deviceswitch/ascend_switch.go:190    '
            'switch subscribe got fault:deviceswitch.SwitchFaultEvent{EventType:0x4, FaultID:0xf1ff09, '
            'AssembledFaultCode:"[0x00f1ff09,155912,npu,na]", PeerPortDevice:0x1, PeerPortId:0x2, '
            'SwitchChipId:0x0, SwitchPortId:0x0, Severity:0x0, Assertion:0x1, EventSerialNum:0, '
            'NotifySerialNum:0, AlarmRaisedTime:0, AdditionalParam:"", AdditionalInfo:""}, '
            'faultCode:[0x00f1ff09,155912,npu,na]'
        )
        results, err_msg_list = parse_knowledge_graph([build_input(DEVICEPLUGIN_SOURCE, [line], server="2001:db8::1")])
        self.assertEqual(err_msg_list, [])
        self.assertIn("Comp_Switch_L1_06", get_codes(results))

    def test_ipv6_func_0010_diag_kg_chain(self):
        """Ascend_MindCluster_Fd_IPv6_SDK_Func_0010"""
        payload = [
            {
                "server": "2001:db8::1",
                "source": "CCAE",
                "fault": [
                    {
                        "event_code": "Comp_Network_Custom_01",
                        "key_info": "",
                        "source_file": "train.log",
                        "source_device": "0",
                        "occur_time": "2024-04-08 16:24:48",
                    },
                    {
                        "event_code": "AISW_CANN_HCCL_027",
                        "key_info": "",
                        "source_file": "train.log",
                        "source_device": "0",
                        "occur_time": "2024-04-08 16:24:48",
                    },
                ],
            }
        ]
        results, err_msg_list = diag_knowledge_graph(payload)
        self.assertEqual(err_msg_list, [])
        self.assertTrue(results[0].get("analyze_success"))
        chains = results[0]["fault"][0].get("fault_chains", [])
        self.assertTrue(any("AISW_CANN_HCCL_027" in str(chain) for chain in chains))

    def test_ipv6_reliability_0001_wrong_format(self):
        """Ascend_MindCluster_Fd_IPv6_SDK_Relb_0001"""
        payload = {
            "log_domain": {"server": "2001:db8:::1"},
            "log_items": [
                build_item(CANN_PLOG_SOURCE, [PLOG_LINE], device_id=0),
                build_item(MINDIE_SOURCE, [MINDIE_LINE], device_id=1),
            ],
        }
        results, err_msg_list = parse_knowledge_graph([payload])
        self.assertEqual(err_msg_list, [])
        self.assertEqual(len(results), 1)


if __name__ == "__main__":
    unittest.main()
