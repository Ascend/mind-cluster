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
import shutil
import os

from ascend_fd.model.context import KGParseCtx
from ascend_fd.model.parse_info import KGParseFilePath
from ascend_fd.pkg.parse.knowledge_graph.parser.amct_log_parser import AMCTLogParser
from ascend_fd.pkg.parse.knowledge_graph.parser.lcne_parser import LCNEParser
from ascend_fd.pkg.parse.knowledge_graph.parser.mindio_parser import MindIOLogParser
from ascend_fd.pkg.parse.knowledge_graph.parser.noded_log_parser import NodeDLogParser
from ascend_fd.pkg.parse.knowledge_graph.parser.npu_device_parse import NpuOsLogParser, NpuDeviceLogParser
from ascend_fd.pkg.parse.knowledge_graph.parser.train_log_parser import TrainLogParser
from ascend_fd.pkg.parse.knowledge_graph.parser.cann_log_parser import CANNPlogParser
from ascend_fd.pkg.parse.knowledge_graph.parser.host_os_parser import HostMsgParser, HostVmCoreParser
from ascend_fd.pkg.parse.knowledge_graph.parser.npu_info_parser import NpuInfoParser
from ascend_fd.pkg.parse.knowledge_graph.parser.device_plugin_parser import DevicePluginParser
from ascend_fd.pkg.parse.knowledge_graph.parser.volcano_parser import VolcanoSchedulerParser, VolcanoControllerParser
from ascend_fd.pkg.parse.knowledge_graph.parser.common_dl_parser import DockerRuntimeParser, NpuExporterParser
from ascend_fd.pkg.parse.knowledge_graph.parser.mindie_parser import MindieParser

TEST_DIR = os.path.dirname(os.path.dirname(os.path.realpath(__file__)))
TESTCASE_KG_PARSE_INPUT = os.path.join(TEST_DIR, "st_module_testcase", "kg_parse")
EVENT_CODE = "event_code"
OCCUR_TIME = "occur_time"
SOURCE_DEVICE = "source_device"
WARN = "WARN"
ASSERTION = "Assertion"


class KgParseTestCase(unittest.TestCase):
    REGEX = "regex"
    IN = "in"

    def setUp(self) -> None:
        self.test_code = "AISW_CANN_ERRMSG_Custom_01"
        self.os_code = "Comp_OS_Kernel_CPU_04"
        self.custom_code = "AISW_CANN_ERRCODE_Custom_E30003"
        self.pytorch_code = "AISW_PyTorch_Init_02"
        self.pytorch_common_code = "AISW_PyTorch_ERRCODE_Common"
        self.pytorch_common_code_detail = "AISW_PyTorch_ERRCODE_Common_ERR00002"
        self.device_plugin_occur_code = "Comp_Switch_L1_05"
        self.device_plugin_recovered_code = "Comp_Switch_L1_01"
        self.device_plugin_cross_file_recovered_code = "Comp_Switch_L1_06"
        self.volcano_scheduler_code = "AISW_MindX_Volcano_03"
        self.volcano_controller_code = "AISW_MindX_Volcano_10"
        self.docker_runtime_code = "AISW_MindX_Docker_Runtime_02"
        self.npu_exporter_code = "AISW_MindX_Npu_Exporter_01"
        self.npu_exporter_dump_log_code = "AISW_MindX_Npu_Exporter_03"
        self.mindie_ms_fault_code = "AISW_MindIE_Ms_HttpServer_01"
        self.noded_code = "0x0000001D"
        self.amct_code = "AISW_CANN_AMCT_ALL_001"
        self.npu_device_code = "Comp_Network_Custom_11"
        self.mindio_code = "AISW_MindIO_TTP_01"
        self.test_regex = {
            self.test_code: {self.IN: ["E30008"]},
            self.custom_code: {
                self.REGEX: "[EWI][A-Z0-9][0-9]{4}",
                "attr_regex": "(?P<complement>[EWI][A-Z0-9][0-9]{4}):"
            },
            "0x8C17A005": {
                self.IN: [["Node Type=0x60b;", "Sensor Type=0xd0;", "event_state=5)"], ["Event id=0x8c17a005"]]
            },
            self.os_code: {self.IN: ["Watchdog detected hard LOCKUP"]},
            self.pytorch_common_code: {
                self.REGEX: "^\\[ERROR].{0,40}PID.{0,15}Device.{0,10}RankID.{0,15}ERR",
                "attr_regex": "^\\[ERROR]\\D{0,5}(?P<occur_time>[0-9-:]{19}).{1,40}Device:(?P<source_device>"
                              "\\d{1,4}),.{1,30}(?P<error_code>ERR\\d{5}).{0,5}(?P<module>PTA|OPS|DIST|GRAPH|PROF)"
            },
            self.pytorch_code: {self.REGEX: "_npu_setDevice"},
            self.device_plugin_occur_code: {
                self.IN: [WARN, "[0x00f1ff09,155912,cpu,na]", "Assertion"]
            },
            self.device_plugin_recovered_code: {
                self.IN: [WARN, "[0x00f103b0,155907,na,na]", ASSERTION]
            },
            self.device_plugin_cross_file_recovered_code: {
                self.IN: [WARN, "[0x00f1ff09,155912,npu,na]", ASSERTION]
            },
            self.volcano_scheduler_code: {
                self.IN: ["Failed to get plugin volcano-npu"]
            },
            self.volcano_controller_code: {
                self.IN: ["Failed to create pod for Job"]
            },
            self.docker_runtime_code: {
                self.IN: ["failed to check files", "owner not right /usr/bin/runc"]
            },
            self.npu_exporter_code: {
                self.IN: ["deviceManager init failed", "cannot found valid driver lib",
                          "fromEnv", "lib path is invalid"]
            },
            self.npu_exporter_dump_log_code: {
                self.IN: ["get npu link stat failed", "no such file or directory"]
            },
            self.mindie_ms_fault_code: {
                self.IN: ["MIE03E400008"]
            },
            self.noded_code: {
                self.IN: [["error", "code", "0000001D"]]
            },
            self.amct_code: {
                self.IN: [["No layer support", "in the graph"]]
            },
            self.npu_device_code: {
                self.IN: ["rf_lf", "pcs_err_cnt"],
                "attr_regex": ", rf_lf (?P<rf_lf>[12]), pcs_err_cnt (?P<pcs_err_cnt>\\d{1,10}), "
            },
            self.mindio_code: {
                self.IN: ["remove broken link"]
            }
        }
        self.params = {
            "regex_conf": {"TrainLog": self.test_regex, "CANN_Plog": self.test_regex, "OS": self.test_regex,
                           "NPU_OS": self.test_regex, "DL_DevicePlugin | LCNELog": self.test_regex,
                           "NodeDLog": self.test_regex,
                           "CANN_Amct": self.test_regex, "NPU_Device": self.test_regex, "MindIO": self.test_regex},
            "start_time": "1999-10-01 03:40:50.000000", "end_time": "2999-10-01 03:40:50.000000"
        }
        self.host_os_parser = HostMsgParser(self.params)
        self.host_os_input_file_list = [
            os.path.join(TESTCASE_KG_PARSE_INPUT, "messages"),
            os.path.join(TESTCASE_KG_PARSE_INPUT, "messages-2"),
            os.path.join(TESTCASE_KG_PARSE_INPUT, "messages-3")
        ]

        self.vmcore_dmesg_parser = HostVmCoreParser(self.params)
        self.crash_dir = os.path.join(TESTCASE_KG_PARSE_INPUT, "crash", "127.0.0.1-2024-09-23-11:25:29")
        if not os.path.exists(self.crash_dir):
            os.makedirs(self.crash_dir)
        shutil.copy(os.path.join(TESTCASE_KG_PARSE_INPUT, "vmcore-dmesg.txt"), self.crash_dir)
        self.vmcore_dmesg_input_file_list = [os.path.join(self.crash_dir, "vmcore-dmesg.txt")]

        self.train_parser = TrainLogParser(self.params)
        self.train_input_file_list = [os.path.join(TESTCASE_KG_PARSE_INPUT, "rank-0.txt")]
        self.train_input_file_list_1 = [os.path.join(TESTCASE_KG_PARSE_INPUT, "rank-1.txt")]

        self.cann_log_parser = CANNPlogParser(self.params)
        self.cann_input_file_list = (
            "test_pid", [os.path.join(TESTCASE_KG_PARSE_INPUT, "plog-10972_20230131024824126.log")]
        )

        self.npu_info_parser = NpuInfoParser(self.params)
        self.npu_info_file_list = [
            os.path.join(TESTCASE_KG_PARSE_INPUT, "npu_info_before.txt"),
            os.path.join(TESTCASE_KG_PARSE_INPUT, "npu_info_after.txt")
        ]
        self.npu_os_parser = NpuOsLogParser(self.params)
        self.npu_os_input_file_dict = {
            TESTCASE_KG_PARSE_INPUT: [os.path.join(TESTCASE_KG_PARSE_INPUT, "device-os_20231031164139951.log")]
        }
        self.npu_device_parser = NpuDeviceLogParser(self.params)
        self.npu_device_input_file_dict = {
            TESTCASE_KG_PARSE_INPUT: [
                os.path.join(TESTCASE_KG_PARSE_INPUT, "device-1", "device-1_20241204162140159.log")]
        }
        self.device_plugin_parser = DevicePluginParser(self.params)
        self.device_plugin_file_list = [
            os.path.join(TESTCASE_KG_PARSE_INPUT, "dl_log", "devicePlugin", "devicePlugin.log"),
            os.path.join(TESTCASE_KG_PARSE_INPUT, "dl_log", "devicePlugin", "devicePlugin-2024-01-10T03-30-45.197.log")
        ]
        self.volcano_scheduler_parser = VolcanoSchedulerParser(self.params)
        self.volcano_scheduler_file_list = [
            os.path.join(TESTCASE_KG_PARSE_INPUT, "dl_log", "volcano-scheduler", "volcano-scheduler.log")
        ]
        self.volcano_controller_parser = VolcanoControllerParser(self.params)
        self.volcano_controller_file_list = [
            os.path.join(TESTCASE_KG_PARSE_INPUT, "dl_log", "volcano-controller", "volcano-controller.log")
        ]
        self.docker_runtime_parser = DockerRuntimeParser(self.params)
        self.docker_runtime_file_list = [
            os.path.join(TESTCASE_KG_PARSE_INPUT, "dl_log", "ascend-docker-runtime", "runtime-run.log")
        ]
        self.npu_exporter_parser = NpuExporterParser(self.params)
        self.npu_exporter_file_list = [
            os.path.join(TESTCASE_KG_PARSE_INPUT, "dl_log", "npu-exporter", "npu-exporter.log"),
            os.path.join(TESTCASE_KG_PARSE_INPUT, "dl_log", "npu-exporter", "npu-exporter-2023-12-22T22-40-17.760.log")
        ]
        self.mindie_ms_parser = MindieParser(self.params)
        self.mindie_log_list = [
            os.path.join(TESTCASE_KG_PARSE_INPUT, "mindie", "log", "debug", "mindie-ms_11_202411061400.log")
        ]
        self.ndoed_log_parser = NodeDLogParser(self.params)
        self.noded_log_list = [
            os.path.join(TESTCASE_KG_PARSE_INPUT, "dl_log", "noded", "noded.log"),
            os.path.join(TESTCASE_KG_PARSE_INPUT, "dl_log", "noded", "noded-2999-10-01T03-39-17.760.log")
        ]
        self.noded_log_dict = {"noded_log_path": self.noded_log_list}
        self.mindio_log_parser = MindIOLogParser(self.params)
        self.mindio_log_list = [
            os.path.join(TESTCASE_KG_PARSE_INPUT, "ttp_log", "ttp_log.log.1")
        ]
        self.amct_log_parser = AMCTLogParser(self.params)
        self.amct_log_dict = {
            "amct_path": [os.path.join(TESTCASE_KG_PARSE_INPUT, "amct_log/amct_onnx.log")]
        }

    def test_train_log_parse_func(self):
        parse_ctx = KGParseCtx(parse_file_path=KGParseFilePath(train_log_path=self.train_input_file_list))
        self.train_parser.parse(parse_ctx, "test_task_id")
        event_result_list = self.train_parser._parse_single_file(self.train_input_file_list[0])
        self.assertEqual(self.test_code, event_result_list[0][EVENT_CODE])
        self.assertEqual(self.custom_code, event_result_list[1][EVENT_CODE])
        self.assertEqual("AISW_TRACEBACK_RuntimeError", event_result_list[2][EVENT_CODE])

    def test_pytorch_train_log_parse_func(self):
        event_result_list = self.train_parser._parse_single_file(self.train_input_file_list_1[0])
        self.assertEqual(self.pytorch_code, event_result_list[0][EVENT_CODE])
        self.assertEqual(self.pytorch_common_code_detail, event_result_list[1][EVENT_CODE])

    def test_host_os_parse_func(self):
        parse_ctx = KGParseCtx(parse_file_path=KGParseFilePath(host_log_path=self.host_os_input_file_list))
        self.host_os_parser.parse(parse_ctx, "test_task_id")
        event_result_list = self.host_os_parser._parse_file(self.host_os_input_file_list[0], "test_task_id")
        self.assertEqual(self.test_code, event_result_list[0][EVENT_CODE])
        event_result_list = self.host_os_parser._parse_chunk(self.host_os_input_file_list[0], 0, 1024 * 1024)
        self.assertEqual(self.test_code, event_result_list[0][EVENT_CODE])

    def test_vmcore_dmesg_parse_func(self):
        event_result_list = self.vmcore_dmesg_parser._parse_chunk(
            self.vmcore_dmesg_input_file_list[0], 0, 1024 * 1024)
        self.assertEqual(self.os_code, event_result_list[0][EVENT_CODE])

    def test_cann_log_parse_func(self):
        event_result_list = self.cann_log_parser._parse_files_of_pid(*self.cann_input_file_list)[0]
        self.assertEqual(self.test_code, event_result_list[0][EVENT_CODE])
        self.assertEqual("AISW_CANN_Runtime_054", event_result_list[1][EVENT_CODE])
        self.assertEqual("AISW_CANN_Memory_Info_Custom", event_result_list[2][EVENT_CODE])

    def test_npu_info_parse_func(self):
        parse_ctx = KGParseCtx(parse_file_path=KGParseFilePath(npu_info_path=self.npu_info_file_list))
        event_result_list, _ = self.npu_info_parser.parse(parse_ctx, "test_task_id")
        self.assertTrue("Comp_Network_Custom_01", event_result_list[0][EVENT_CODE])
        self.assertTrue("0x8C03A000", event_result_list[1][EVENT_CODE])
        self.assertTrue("0x8C084E00", event_result_list[2][EVENT_CODE])
        self.assertTrue("0xA4025021", event_result_list[3][EVENT_CODE])
        self.assertTrue("AISW_CANN_DRV_Custom_01", event_result_list[4][EVENT_CODE])
        self.assertTrue("Comp_Network_Custom_07", event_result_list[5][EVENT_CODE])
        self.assertTrue("Comp_NPU_DRV_Custom_01", event_result_list[6][EVENT_CODE])
        self.assertTrue("Comp_NPU_DRV_Custom_02", event_result_list[7][EVENT_CODE])
        self.assertTrue("Comp_Network_Custom_05", event_result_list[8][EVENT_CODE])
        self.assertTrue("Comp_Network_Custom_08", event_result_list[9][EVENT_CODE])
        self.assertTrue("Comp_Network_Custom_09", event_result_list[10][EVENT_CODE])
        self.assertTrue("Comp_Network_Custom_10", event_result_list[11][EVENT_CODE])
        self.assertTrue("Comp_Network_Custom_02", event_result_list[12][EVENT_CODE])
        self.assertTrue("Comp_Network_Custom_03", event_result_list[13][EVENT_CODE])
        self.assertTrue("23.0.7", event_result_list[14]["driver_version"])
        self.assertTrue("7.1.0.11.220", event_result_list[14]["firm_version"])
        self.assertTrue("7.0.T10", event_result_list[14]["cann_version"])
        self.assertTrue("8.0.RC3", event_result_list[14]["nnae_version"])
        self.assertTrue("1.11.0", event_result_list[14]["pytorch_version"])
        self.assertTrue("2.1.0.post8.dev20241009", event_result_list[14]["torch_npu_version"])
        self.assertTrue("2.3.0", event_result_list[14]["mindspore_version"])

    def test_npu_os_parse_func(self):
        parse_ctx = KGParseCtx(parse_file_path=KGParseFilePath(slog_path=self.npu_os_input_file_dict))
        event_result_list, _ = self.npu_os_parser.parse(parse_ctx, "test_task_id")
        self.assertEqual("0x8C17A005", event_result_list[0][EVENT_CODE])

    def test_npu_os_parse_single_file(self):
        file_path = self.npu_os_input_file_dict.get(TESTCASE_KG_PARSE_INPUT, [])[0]
        event_result_list = self.npu_os_parser._parse_single_file(file_path)
        self.assertEqual("0x8C17A005", event_result_list[0][EVENT_CODE])
        self.assertEqual("Unknown", event_result_list[0][SOURCE_DEVICE])

    def test_npu_device_parse_func(self):
        parse_ctx = KGParseCtx(parse_file_path=KGParseFilePath(slog_path=self.npu_device_input_file_dict))
        event_result_list, _ = self.npu_device_parser.parse(parse_ctx, "test_task_id")
        self.assertEqual("Comp_Network_Custom_11", event_result_list[0][EVENT_CODE])

    def test_npu_device_parse_single_file(self):
        file_path = self.npu_device_input_file_dict.get(TESTCASE_KG_PARSE_INPUT, [])[0]
        event_result_list = self.npu_device_parser._parse_single_file(file_path)
        self.assertEqual("Comp_Network_Custom_11", event_result_list[0][EVENT_CODE])
        self.assertEqual("1", event_result_list[0][SOURCE_DEVICE])

    def test_npu_0x40f84e00(self):
        parse_ctx = KGParseCtx(parse_file_path=KGParseFilePath(
            host_log_path=[os.path.join(TESTCASE_KG_PARSE_INPUT, "messages_0x40F84E00")]))
        event_list, _ = self.host_os_parser.parse(parse_ctx, "test_task_id")
        for event in event_list:
            self.assertEqual("0x40F84E00", event[EVENT_CODE])
            if "08-01" in event[OCCUR_TIME]:
                self.assertEqual("1", event[SOURCE_DEVICE])
            if "08-02" in event[OCCUR_TIME]:
                self.assertEqual("1000", event[SOURCE_DEVICE])
            if "08-03" in event[OCCUR_TIME]:
                self.assertEqual("Unknown", event[SOURCE_DEVICE])

    def test_mindie_ms_parse_func(self):
        parse_ctx = KGParseCtx(parse_file_path=KGParseFilePath(mindie_log_path=self.mindie_log_list))
        files_parse_info, _ = self.mindie_ms_parser.parse(parse_ctx, "test_task_id")
        self.assertTrue("AISW_MindIE_Ms_HttpServer_01", files_parse_info.event_list[0][EVENT_CODE])

    def test_mindie_ms_parse_file(self):
        single_file_parse_info = self.mindie_ms_parser._parse_file(self.mindie_log_list[0])
        self.assertTrue("AISW_MindIE_Ms_HttpServer_01", single_file_parse_info.event_list[0][EVENT_CODE])
        self.assertEqual("xxx.xx.xx.x", single_file_parse_info.device_info.device_ip)
        self.assertEqual("6", single_file_parse_info.device_info.phy_device_id)
        self.assertEqual("6", single_file_parse_info.device_info.logic_device_id)
        self.assertEqual("6", single_file_parse_info.device_info.device_id)
        self.assertEqual("11", single_file_parse_info.device_info.pid)
        self.assertTrue("AISW_MindIE_ERRCODE_Common_MIE1AE66666B", single_file_parse_info.event_list[0][EVENT_CODE])

    def test_device_plugin_parse_func(self):
        parse_ctx = KGParseCtx(parse_file_path=KGParseFilePath(device_plugin_path=self.device_plugin_file_list))
        event_result_list, _ = self.device_plugin_parser.parse(parse_ctx, "test_task_id")
        result_code = [event[EVENT_CODE] for event in event_result_list]
        self.assertIn(self.device_plugin_occur_code, result_code)
        self.assertNotIn(self.device_plugin_recovered_code, result_code)
        self.assertNotIn(self.device_plugin_cross_file_recovered_code, result_code)

    def test_volcano_scheduler_parse_func(self):
        parse_ctx = KGParseCtx(parse_file_path=KGParseFilePath(volcano_scheduler_path=self.volcano_scheduler_file_list))
        event_result_list, _ = self.volcano_scheduler_parser.parse(parse_ctx, "test_task_id")
        self.assertTrue("AISW_MindX_Volcano_03", event_result_list[0][EVENT_CODE])

    def test_volcano_controller_parse_func(self):
        parse_ctx = KGParseCtx(parse_file_path=KGParseFilePath(
            volcano_controller_path=self.volcano_controller_file_list))
        event_result_list, _ = self.volcano_controller_parser.parse(parse_ctx, "test_task_id")
        self.assertTrue("AISW_MindX_Volcano_10", event_result_list[0][EVENT_CODE])

    def test_docker_runtime_parse_func(self):
        parse_ctx = KGParseCtx(parse_file_path=KGParseFilePath(docker_runtime_path=self.docker_runtime_file_list))
        event_result_list, _ = self.docker_runtime_parser.parse(parse_ctx, "test_task_id")
        self.assertTrue("AISW_MindX_Docker_Runtime_02", event_result_list[0][EVENT_CODE])

    def test_dl_log_parse_file(self):
        event_result_list = self.docker_runtime_parser._parse_file(self.docker_runtime_file_list[0])
        self.assertTrue("AISW_MindX_Docker_Runtime_02", event_result_list[0][EVENT_CODE])

    def test_npu_exporter_parse_func(self):
        parse_ctx = KGParseCtx(parse_file_path=KGParseFilePath(npu_exporter_path=self.npu_exporter_file_list))
        event_result_list, _ = self.npu_exporter_parser.parse(parse_ctx, "test_task_id")
        self.assertTrue("AISW_MindX_Npu_Exporter_03", event_result_list[0][EVENT_CODE])
        self.assertTrue("AISW_MindX_Npu_Exporter_01", event_result_list[1][EVENT_CODE])

    def test_noded_log_parse_func(self):
        parse_ctx = KGParseCtx(parse_file_path=KGParseFilePath(noded_log_path=self.noded_log_list))
        event_list, _ = self.ndoed_log_parser.parse(parse_ctx, "test_task_id")
        self.assertEqual("0x0000001D", event_list[0][EVENT_CODE])
        self.assertEqual("0x2C000031", event_list[1][EVENT_CODE])

    def test_noded_log_parse_no_time(self):
        ndoed_log_parser = NodeDLogParser({"NodeDLog": self.test_regex})
        parse_ctx = KGParseCtx(parse_file_path=KGParseFilePath(noded_log_path=self.noded_log_list))
        event_list, _ = ndoed_log_parser.parse(parse_ctx, "test_task_id")
        self.assertEqual("0x0000001D", event_list[0][EVENT_CODE])

    def test_noded_log_parse_single_file(self):
        file_path = self.noded_log_dict.get("noded_log_path", [])[-1]
        event_dict = self.ndoed_log_parser._parse_single_file(file_path)
        key = ": BMC]-[device id: 255]-[error code: 2C000031]"
        self.assertEqual("0x2C000031", event_dict.get(key, {}).get("event_code"))

    def test_mindio_log_parse_func(self):
        parse_ctx = KGParseCtx(parse_file_path=KGParseFilePath(mindio_log_path=self.mindio_log_list))
        event_list, _ = self.mindio_log_parser.parse(parse_ctx, "test_task_id")
        self.assertEqual(self.mindio_code, event_list[0][EVENT_CODE])

    def test_amct_log_parse_func(self):
        parse_ctx = KGParseCtx(parse_file_path=KGParseFilePath(
            amct_path=[os.path.join(TESTCASE_KG_PARSE_INPUT, "amct_log/amct_onnx.log")]))
        event_list, _ = self.amct_log_parser.parse(parse_ctx, "test_task_id")
        self.assertEqual("AISW_CANN_AMCT_ALL_001", event_list[0][EVENT_CODE])

    def test_amct_log_parse_single_file(self):
        file_path = self.amct_log_dict.get("amct_path", [])[-1]
        event = self.amct_log_parser._parse_single_file(file_path)
        self.assertEqual("AISW_CANN_AMCT_ALL_001", event[0][EVENT_CODE])

    def test_lcne_log_parse_single_file(self):
        test_regex = {
            "Comp_Switch_L1_21": {
                "in": ["alarmID=0x00f10509"]
            }
        }
        params = {
            "regex_conf": {"LCNELog": test_regex}
        }
        file_path = os.path.join(TESTCASE_KG_PARSE_INPUT, "lcne_log", "log_1_20250814013158.log")

        lcne_log_parser = LCNEParser(params)
        lcne_log_parser.timezone_trans_flag = True
        event = lcne_log_parser._parse_file(file_path)
        self.assertEqual("Comp_Switch_L1_21", event[0][EVENT_CODE])
        self.assertEqual("2025-07-31 10:36:19.000000", event[0]["occur_time"])

        lcne_log_parser.timezone_trans_flag = False
        event = lcne_log_parser._parse_file(file_path)
        self.assertEqual("2025-07-31 18:36:19.000000", event[0]["occur_time"])

    def tearDown(self) -> None:
        if os.path.exists(self.crash_dir):
            shutil.rmtree(self.crash_dir)
