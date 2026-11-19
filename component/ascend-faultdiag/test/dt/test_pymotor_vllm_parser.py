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
import unittest
import tempfile
import os

from ascend_fd.pkg.parse.knowledge_graph.parser.pymotor_vllm_parser import PyMotorVLLMParser


class TestPyMotorVLLMParser(unittest.TestCase):
    """Test cases for PYMOTOR VLLM parser"""

    def _create_parser(self):
        return PyMotorVLLMParser({})

    def _create_temp_file(self, content, prefix="vllm-test"):
        with tempfile.NamedTemporaryFile(mode='w', suffix='.log', prefix=prefix + "-", delete=False) as tmp_file:
            tmp_file.write(content)
            return tmp_file.name

    def test_parse_vllm_device_info(self):
        """Test parsing vllm device info from log line"""
        log_line = "2026-05-26 15:15:49,051 - INFO - device_id: 0, device_ip_info: ['ipaddr:10.0.245.11\\n', 'netmask:255.255.255.0\\n']"
        parser = self._create_parser()
        device_info = parser.find_device_info(log_line)

        self.assertIsNotNone(device_info)
        self.assertEqual(device_info.device_id, "0")
        self.assertEqual(device_info.device_ip, "10.0.245.11")

    def test_parse_motor_log_time(self):
        """Test parsing motor format timestamp"""
        log_line = "2026-05-26 15:16:20  [INFO][root][logger.py:170][proc:MainProcess]  Internal logs of pod"
        parser = self._create_parser()
        time_str = parser.get_log_time(log_line)

        self.assertTrue(time_str.startswith("2026-05-26 15:16:20"))

    def test_parse_vllm_log_time(self):
        """Test parsing vllm format timestamp"""
        log_line = "2026-05-26 15:15:48,667 - INFO - start /mnt/configmap/hccl_tools.py"
        parser = self._create_parser()
        time_str = parser.get_log_time(log_line)

        self.assertTrue(time_str.startswith("2026-05-26 15:15:48"))

    def test_extract_host_pod_ip(self):
        """Test extracting host_ip and pod_ip"""
        log_line_host = "2026-05-26 15:15:48,694 - INFO - host_ip: 90.90.97.42"
        log_line_pod = "2026-05-26 15:15:48,694 - INFO - pod_ip: 192.168.222.203"

        parser = self._create_parser()
        parser.extract_host_pod_ip(log_line_host)
        parser.extract_host_pod_ip(log_line_pod)

        self.assertEqual(parser.host_ip, "90.90.97.42")
        self.assertEqual(parser.pod_ip, "192.168.222.203")

    def test_invalid_ip_rejected(self):
        """Test that invalid IP addresses are rejected"""
        log_line = "2026-05-26 15:15:49,051 - INFO - device_id: 0, device_ip_info: ['ipaddr:invalid_ip\\n', 'netmask:255.255.255.0\\n']"
        parser = self._create_parser()
        device_info = parser.find_device_info(log_line)

        self.assertIsNone(device_info)

    def test_multiple_devices_parsed(self):
        """Test parsing multiple devices from log content"""
        log_content = """2026-05-26 15:15:49,051 - INFO - device_id: 0, device_ip_info: ['ipaddr:10.0.245.11\\n', 'netmask:255.255.255.0\\n']
2026-05-26 15:15:50,285 - INFO - device_id: 1, device_ip_info: ['ipaddr:10.0.244.11\\n', 'netmask:255.255.255.0\\n']
2026-05-26 15:15:51,175 - INFO - device_id: 10, device_ip_info: ['ipaddr:10.0.245.16\\n', 'netmask:255.255.255.0\\n']"""

        parser = self._create_parser()
        file_path = self._create_temp_file(log_content)

        try:
            result = parser._parse_file(file_path)

            self.assertEqual(len(result.device_info_list), 3)
            device_ids = {d.device_id for d in result.device_info_list}
            self.assertEqual(device_ids, {"0", "1", "10"})
        finally:
            os.unlink(file_path)

    def test_container_ip_from_pod_ip(self):
        """Test that container IP is set from pod_ip"""
        log_content = """2026-05-26 15:15:48,694 - INFO - host_ip: 90.90.97.42
2026-05-26 15:15:48,694 - INFO - pod_ip: 192.168.222.203
2026-05-26 15:15:49,051 - INFO - device_id: 0, device_ip_info: ['ipaddr:10.0.245.11\\n', 'netmask:255.255.255.0\\n']"""

        parser = self._create_parser()
        file_path = self._create_temp_file(log_content)

        try:
            result = parser._parse_file(file_path)

            self.assertEqual(result.container_ip, "192.168.222.203")
        finally:
            os.unlink(file_path)


class TestPyMotorVLLMParserIPv6(unittest.TestCase):
    """Test cases for PYMOTOR VLLM parser with IPv6 addresses"""

    def _create_parser(self):
        return PyMotorVLLMParser({})

    def _create_temp_file(self, content, prefix="vllm-"):
        with tempfile.NamedTemporaryFile(mode='w', suffix='.log', prefix=prefix + "-", delete=False) as tmp_file:
            tmp_file.write(content)
            return tmp_file.name

    def test_find_device_info_ipv6(self):
        """Test parsing device info with IPv6 address"""
        log_line = (
            "2026-06-08 10:00:00,000 - INFO - "
            "device_id: 0, device_ip_info: ['ipaddr:fe80::1c01:3c0\\n', 'netmask:ffff:ffff::\\n']"
        )
        parser = self._create_parser()
        device_info = parser.find_device_info(log_line)

        self.assertIsNotNone(device_info)
        self.assertEqual(device_info.device_id, "0")
        self.assertEqual(device_info.device_ip, "fe80::1c01:3c0")

    def test_find_device_info_ipv6_full(self):
        """Test parsing device info with full-format IPv6 address"""
        log_line = (
            "2026-06-08 10:00:01,000 - INFO - "
            "device_id: 3, device_ip_info: ['ipaddr:2001:0db8:85a3:0000:0000:8a2e:0370:7334\\n', "
            "'netmask:ffff:ffff:ffff::\\n']"
        )
        parser = self._create_parser()
        device_info = parser.find_device_info(log_line)

        self.assertIsNotNone(device_info)
        self.assertEqual(device_info.device_id, "3")
        self.assertEqual(device_info.device_ip, "2001:0db8:85a3:0000:0000:8a2e:0370:7334")

    def test_extract_host_pod_ip_ipv6(self):
        """Test extracting host_ip and pod_ip with IPv6"""
        log_line_host = "2026-06-08 10:00:02,000 - INFO - host_ip: fe80::1c01:3c0"
        log_line_pod = "2026-06-08 10:00:02,000 - INFO - pod_ip: 2001:db8::1"

        parser = self._create_parser()
        parser.extract_host_pod_ip(log_line_host)
        parser.extract_host_pod_ip(log_line_pod)

        self.assertEqual(parser.host_ip, "fe80::1c01:3c0")
        self.assertEqual(parser.pod_ip, "2001:db8::1")

    def test_container_ip_from_pod_ip_ipv6(self):
        """Test that container IP is set from IPv6 pod_ip"""
        log_content = """2026-06-08 10:00:03,000 - INFO - host_ip: fe80::1c01:3c0
2026-06-08 10:00:03,000 - INFO - pod_ip: 2001:db8::1
2026-06-08 10:00:04,000 - INFO - device_id: 0, device_ip_info: ['ipaddr:fe80::1c01:3c0\\n', 'netmask:ffff:ffff::\\n']"""

        parser = self._create_parser()
        file_path = self._create_temp_file(log_content)

        try:
            result = parser._parse_file(file_path)

            self.assertEqual(result.container_ip, "2001:db8::1")
        finally:
            os.unlink(file_path)

    def test_multiple_devices_parsed_ipv6(self):
        """Test parsing multiple devices with IPv6 addresses"""
        log_content = (
            "2026-06-08 10:00:05,000 - INFO - device_id: 0, device_ip_info: ['ipaddr:fe80::1c01:3c0\\n', 'netmask:ffff:ffff::\\n']\n"
            "2026-06-08 10:00:06,000 - INFO - device_id: 1, device_ip_info: ['ipaddr:2001:db8::2\\n', 'netmask:ffff:ffff::\\n']\n"
            "2026-06-08 10:00:07,000 - INFO - device_id: 2, device_ip_info: ['ipaddr:fd00::100\\n', 'netmask:ffff:ffff::\\n']\n"
        )

        parser = self._create_parser()
        file_path = self._create_temp_file(log_content)

        try:
            result = parser._parse_file(file_path)

            self.assertEqual(len(result.device_info_list), 3)
            device_ids = {d.device_id for d in result.device_info_list}
            self.assertEqual(device_ids, {"0", "1", "2"})
            ips = {d.device_ip for d in result.device_info_list}
            self.assertIn("fe80::1c01:3c0", ips)
            self.assertIn("2001:db8::2", ips)
            self.assertIn("fd00::100", ips)
        finally:
            os.unlink(file_path)


class TestPyMotorVLLMParserSDKInput(unittest.TestCase):
    """Test that PyMotorVLLMParser works with SDK input (LogInfoSaver objects)"""

    CONFIG_PATH = os.path.join(
        os.path.dirname(__file__), '..', '..', 'src', 'ascend_fd', 'configuration', 'kg-config.json'
    )

    def _load_config(self):
        from ascend_fd.utils.load_kg_config import ParseRegexMap

        parse_regex = ParseRegexMap(config_pkg_list=[self.CONFIG_PATH]).get_parse_regex()
        return parse_regex.get(PyMotorVLLMParser.SOURCE_FILE, {})

    def _create_parser(self):
        config = self._load_config()
        return PyMotorVLLMParser(
            {
                "default_conf": {PyMotorVLLMParser.SOURCE_FILE: config},
                "user_conf": {},
            }
        )

    def test_sdk_input_eplb_001(self):
        """SDK input with LogInfoSaver should match EPLB fault without filename filtering"""
        from ascend_fd.pkg.parse.parser_saver import LogInfoSaver
        from ascend_fd.model.context import KGParseCtx
        from ascend_fd.model.parse_info import KGParseFilePath

        log_saver = LogInfoSaver(
            source_file="PyMotor_vLLM_Log",
            path="/log/debug/vllm-2025_202503252036.log",
            device_id=0,
            log_lines=[
                "2026-05-26 15:15:49.409000",
                "(APIServer pid=2056) AssertionError: EPLB must used in expert parallelism.",
            ],
            modification_time="",
            component="",
        )
        parse_file_path = KGParseFilePath(pymotor_vllm_log_path=[log_saver])
        parse_ctx = KGParseCtx(parse_file_path=parse_file_path, is_sdk_input=True)

        parser = self._create_parser()
        events, _ = parser.parse(parse_ctx, "test-task")

        event_codes = {ev.get("event_code") for ev in events}
        self.assertIn("AISW_VLLM_ASCEND_EPLB_001", event_codes)

    def test_sdk_input_no_match(self):
        """SDK input with irrelevant log lines should produce no events"""
        from ascend_fd.pkg.parse.parser_saver import LogInfoSaver
        from ascend_fd.model.context import KGParseCtx
        from ascend_fd.model.parse_info import KGParseFilePath

        log_saver = LogInfoSaver(
            source_file="PyMotor_vLLM_Log",
            path="/log/debug/vllm-normal.log",
            device_id=0,
            log_lines=[
                "2026-05-26 15:15:49.409000",
                "INFO - Engine running normally",
            ],
            modification_time="",
            component="",
        )
        parse_file_path = KGParseFilePath(pymotor_vllm_log_path=[log_saver])
        parse_ctx = KGParseCtx(parse_file_path=parse_file_path, is_sdk_input=True)

        parser = self._create_parser()
        events, _ = parser.parse(parse_ctx, "test-task")

        self.assertEqual(len(events), 0)


class TestPyMotorVLLMParserMultiLineFaultMode(unittest.TestCase):
    """Regression tests for cross-line fault mode matching (all/opt/max_lines)"""

    CONFIG_PATH = os.path.join(
        os.path.dirname(__file__), '..', '..', 'src', 'ascend_fd', 'configuration', 'kg-config.json'
    )

    def _load_config(self):
        from ascend_fd.utils.load_kg_config import ParseRegexMap

        parse_regex = ParseRegexMap(config_pkg_list=[self.CONFIG_PATH]).get_parse_regex()
        return parse_regex.get(PyMotorVLLMParser.SOURCE_FILE, {})

    def _create_parser(self):
        config = self._load_config()
        return PyMotorVLLMParser(
            {
                "default_conf": {PyMotorVLLMParser.SOURCE_FILE: config},
                "user_conf": {},
            }
        )

    def _create_temp_file(self, content, prefix="vllm-mooncake"):
        with tempfile.NamedTemporaryFile(mode='w', suffix='.log', prefix=prefix + "-", delete=False) as tmp_file:
            tmp_file.write(content)
            return tmp_file.name

    def test_mooncake_002_cross_line_match(self):
        """MOONCAKE_002 opt/all groups on different lines must match (regression)"""
        log_content = (
            "\x1b[93m2026-06-24 10:46:50.237365 WARNING  \x1b[0m\x1b[0K[3591] "
            "[client_pool.hpp:445] send request to 80.48.37.140:50060 failed. connection refused. \n"
            "E20260624 10:46:50.237413  3591 master_client.cpp:234] Client not available \n"
            "(Worker_TP1 pid=3193) ERROR 06-24 10:46:50 [mooncake_backend.py:146] "
            "[vllm-ascend] [distributed] - Initialize mooncake failed. ret=-600, "
            "metadata_server=P2PHANDSHAKE. Check mooncake config and network. \n"
        )
        parser = self._create_parser()
        file_path = self._create_temp_file(log_content)
        try:
            result = parser._parse_file(file_path)
            event_codes = {ev.get("event_code") for ev in result.event_list}
            self.assertIn("AISW_VLLM_ASCEND_KV_POOL_MOONCAKE_002", event_codes)
        finally:
            os.unlink(file_path)

    def test_mooncake_002_no_match_when_group_missing(self):
        """MOONCAKE_002 must not match when one all-group is absent"""
        log_content = (
            "\x1b[93m2026-06-24 10:46:50.237365 WARNING  \x1b[0m\x1b[0K[3591] "
            "[client_pool.hpp:445] send request to 80.48.37.140:50060 failed. connection refused. \n"
            "E20260624 10:46:50.237413  3591 master_client.cpp:234] Client not available \n"
        )
        parser = self._create_parser()
        file_path = self._create_temp_file(log_content)
        try:
            result = parser._parse_file(file_path)
            event_codes = {ev.get("event_code") for ev in result.event_list}
            self.assertNotIn("AISW_VLLM_ASCEND_KV_POOL_MOONCAKE_002", event_codes)
        finally:
            os.unlink(file_path)


MEM_FAULT_LOG_CASES = {
    "MF-GEN-001": [
        ("src/smem/csrc/smem_bm/smem_bm_entry.cpp:300 SmemBmEntry::Join timeout waiting for PROMOTE_TO_ACTIVE, rank=3"),
        (
            "src/smem/csrc/smem_trans/smem_trans_connection.cpp:649 "
            "SmemTransConnection::Join timeout waiting for PROMOTE_TO_ACTIVE, rank=3 name:pod-0"
        ),
    ],
    "MF-SDMA-001": [
        ("src/hybm/csrc/mm/hybm_vmm_based_segment.cpp:155 Try HalMemCreate ret:507899 dev:0 spend time:12 size:4096"),
    ],
    "MF-SDMA-002": [
        (
            "src/hybm/csrc/ts_engine/hybm_stream.cpp:384 "
            "task exec failed, stream:1 sqeType:11 cqeErrorCode:8(sdma move COMPERR ERROR) cqeErrorType:2"
        ),
        (
            "src/hybm/csrc/ts_engine/hybm_stream.cpp:384 "
            "task exec failed, stream:1 sqeType:11 cqeErrorCode:9(sdma move COMPDATAERR ERROR) cqeErrorType:2"
        ),
    ],
    "MF-RDMA-001": [
        (
            "src/hybm/csrc/mm/hybm_conn_based_segment.cpp:116 "
            "Unable to alloc size:1073741824 with hugepage via mmap error:12"
        ),
    ],
    "MF-SHM-001": [
        ("src/smem/csrc/smem_bm/smem_bm.cpp:313 HOST_SHM op type only supports DRAM shared memory without HBM"),
        ("src/smem/csrc/smem_bm/smem_bm.cpp:321 HOST_SHM op type does not support mixing with other data op types"),
    ],
    "MEMCACHE-002": [
        "src/memcache/csrc/client/mmc_client.cpp:120 client 12 batch get lease expired.",
        "src/memcache/csrc/client/mmc_client.cpp:150 client 12 batch copy read lease expired.",
        "src/memcache/csrc/client/mmc_client.cpp:180 client 12 get 34 lease expired after read.",
    ],
    "MF-CLUSTER-001": [
        (
            "src/memcache/csrc/meta_service/mmc_meta_service_process.cpp:321 "
            "Fatal signal caught, signo=11 code=1 errno=0 sender_pid=123 sender_uid=0 addr=0x1 addr_lsb=0"
        ),
        (
            "src/memcache/csrc/meta_service/mmc_meta_service.cpp:200 "
            "HaConfigStore elected as leader, starting MetaNetServer and enabling KvEvents"
        ),
        (
            "src/memcache/csrc/meta_service/mmc_meta_service.cpp:210 "
            "HaConfigStore became standby/follower, stopping MetaNetServer and disabling KvEvents"
        ),
    ],
    "MF-CLUSTER-002": [
        "src/memcache/csrc/meta_service/mmc_meta_net_client.cpp:346 10.0.0.1 link broken",
        "src/memcache/csrc/meta_service/mmc_meta_net_client.cpp:231 10.0.0.1 server changed notify",
    ],
}


class TestPyMotorVLLMParserMemFaultModes(unittest.TestCase):
    """Fault mode tests for MemFabric and Memcache entries in kg-config"""

    CONFIG_PATH = os.path.join(
        os.path.dirname(__file__), '..', '..', 'src', 'ascend_fd', 'configuration', 'kg-config.json'
    )

    @classmethod
    def setUpClass(cls):
        from ascend_fd.utils.load_kg_config import ParseRegexMap

        parse_regex = ParseRegexMap(config_pkg_list=[cls.CONFIG_PATH]).get_parse_regex()
        cls.config = parse_regex.get(PyMotorVLLMParser.SOURCE_FILE, {})

    def _create_parser(self):
        return PyMotorVLLMParser(
            {
                "default_conf": {PyMotorVLLMParser.SOURCE_FILE: self.config},
                "user_conf": {},
            }
        )

    def _create_temp_file(self, content, prefix="vllm-memfault"):
        with tempfile.NamedTemporaryFile(mode='w', suffix='.log', prefix=prefix + "-", delete=False) as tmp_file:
            tmp_file.write(content)
            return tmp_file.name

    def _parse_event_codes(self, log_lines):
        content = "".join(f"2026-11-11 10:00:00,000 - ERROR - {line}\n" for line in log_lines)
        file_path = self._create_temp_file(content)
        try:
            result = self._create_parser()._parse_file(file_path)
            return {ev.get("event_code") for ev in result.event_list}
        finally:
            os.unlink(file_path)

    def _assert_fault_matched(self, event_code):
        for log_line in MEM_FAULT_LOG_CASES[event_code]:
            with self.subTest(event_code=event_code, log_line=log_line):
                self.assertIn(event_code, self._parse_event_codes([log_line]))

    def test_mf_gen_001_join_timeout(self):
        """MF-GEN-001: Join timeout of either entry should be matched"""
        self._assert_fault_matched("MF-GEN-001")

    def test_mf_sdma_001_hal_mem_create_failure(self):
        """MF-SDMA-001: HalMemCreate failure log should be matched"""
        self._assert_fault_matched("MF-SDMA-001")

    def test_mf_sdma_002_sdma_move_error(self):
        """MF-SDMA-002: Both COMPERR and COMPDATAERR cqe error codes should be matched"""
        self._assert_fault_matched("MF-SDMA-002")

    def test_mf_rdma_001_alloc_local_memory_failure(self):
        """MF-RDMA-001: Segment slice allocation failure should be matched"""
        self._assert_fault_matched("MF-RDMA-001")

    def test_mf_shm_001_host_shm_param_conflict(self):
        """MF-SHM-001: Both HOST_SHM parameter conflict logs should be matched"""
        self._assert_fault_matched("MF-SHM-001")

    def test_memcache_002_kvcache_lease_expired(self):
        """MEMCACHE-002: All lease expired key logs should be matched"""
        self._assert_fault_matched("MEMCACHE-002")

    def test_mf_cluster_001_leader_process_crash(self):
        """MF-CLUSTER-001: All leader crash and switchover key logs should be matched"""
        self._assert_fault_matched("MF-CLUSTER-001")

    def test_mf_cluster_002_management_link_broken(self):
        """MF-CLUSTER-002: Link broken and server changed notify logs should be matched"""
        self._assert_fault_matched("MF-CLUSTER-002")

    def test_normal_log_not_matched(self):
        """Normal MemFabric and Memcache logs should not trigger the new fault modes"""
        event_codes = self._parse_event_codes(
            [
                "src/smem/csrc/smem_bm/smem_bm_entry.cpp:100 SmemBmEntry init success, rank=0",
                "src/memcache/csrc/meta_service/mmc_meta_service.cpp:100 meta service running normally",
            ]
        )
        self.assertFalse(event_codes & set(MEM_FAULT_LOG_CASES))


class TestPyMotorVLLMParserMMCLog(unittest.TestCase):
    """Test cases for mmc-*.log (Memcache) log file support"""

    CONFIG_PATH = os.path.join(
        os.path.dirname(__file__), '..', '..', 'src', 'ascend_fd', 'configuration', 'kg-config.json'
    )

    def _load_config(self):
        from ascend_fd.utils.load_kg_config import ParseRegexMap

        parse_regex = ParseRegexMap(config_pkg_list=[self.CONFIG_PATH]).get_parse_regex()
        return parse_regex.get(PyMotorVLLMParser.SOURCE_FILE, {})

    def _create_parser(self):
        config = self._load_config()
        return PyMotorVLLMParser(
            {
                "default_conf": {PyMotorVLLMParser.SOURCE_FILE: config},
                "user_conf": {},
            }
        )

    def test_mmc_filename_accepted(self):
        """Any log filename starting with mmc should be accepted by the parser"""
        for filename in (
            "mmc_local_8600.log",
            "mmc-090.log",
            "mmc_090.log",
            "mmc090.log",
            "mmc-d0-20250101.log",
        ):
            self.assertTrue(PyMotorVLLMParser._is_accepted_file(filename), filename)

    def test_non_mmc_filename_rejected(self):
        """Filenames not starting with mmc or not ending with .log should be rejected"""
        for filename in ("amc-090.log", "mm-090.log", "mmc-090.txt", "mmc_090", "other-090.log"):
            self.assertFalse(PyMotorVLLMParser._is_accepted_file(filename), filename)

    def test_mmc_log_in_subdir_collected(self):
        """mmc*.log in a subdirectory should be collected by the log saver"""
        from ascend_fd.pkg.parse.parser_saver import PyMotorVLLMLogSaver

        with tempfile.TemporaryDirectory() as tmp_dir:
            sub_dir = os.path.join(tmp_dir, "MF-GEN-001")
            os.makedirs(sub_dir)
            mmc_path = os.path.join(sub_dir, "mmc_local_8600.log")
            vllm_path = os.path.join(tmp_dir, "vllm-d0.log")
            ignored_path = os.path.join(tmp_dir, "unrelated.log")
            for path in (mmc_path, vllm_path, ignored_path):
                with open(path, "w", encoding="utf-8") as tmp_file:
                    tmp_file.write("INFO - no event\n")

            saver = PyMotorVLLMLogSaver()
            saver.filter_log(tmp_dir)
            collected = saver.get_pymotor_vllm_log_list()

            self.assertIn(mmc_path, collected)
            self.assertIn(vllm_path, collected)
            self.assertNotIn(ignored_path, collected)

    def test_parse_mmc_log_in_subdir(self):
        """Parsing a mmc-*.log file in a subdirectory should produce events"""
        from ascend_fd.model.context import KGParseCtx
        from ascend_fd.model.parse_info import KGParseFilePath

        log_content = (
            "2026-05-26 15:15:49,409 - ERROR - "
            "(APIServer pid=2056) AssertionError: EPLB must used in expert parallelism.\n"
        )
        with tempfile.TemporaryDirectory() as tmp_dir:
            sub_dir = os.path.join(tmp_dir, "MF-GEN-001")
            os.makedirs(sub_dir)
            file_path = os.path.join(sub_dir, "mmc-090.log")
            with open(file_path, "w", encoding="utf-8") as tmp_file:
                tmp_file.write(log_content)

            parse_file_path = KGParseFilePath(pymotor_vllm_log_path=[file_path])
            parse_ctx = KGParseCtx(parse_file_path=parse_file_path, is_sdk_input=False)

            events, _ = self._create_parser().parse(parse_ctx, "test-task")

            event_codes = {ev.get("event_code") for ev in events}
            self.assertIn("AISW_VLLM_ASCEND_EPLB_001", event_codes)


if __name__ == '__main__':
    unittest.main()
