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

from ascend_fd import diag_root_cluster, parse_root_cluster
from ascend_fd.pkg.diag.root_cluster import fault_description
from ascend_fd.pkg.diag.root_cluster.utils import ErrorChecker

PLOG_LINE = (
    "[ERROR] HCCL(69395,python3):2024-04-08-02:55:04.903.511 [task_exception_handler.cc:452]"
    "[69395][TaskExceptionHandler][Callback]Task run failed, base information is streamID:[3]"
)


def build_parse_input(missing=None):
    item = {"item_type": "plog", "pid": 3199, "device_id": 0, "log_lines": [PLOG_LINE]}
    domain = {"server": "x.x.x.x", "instance_id": "xxxx1"}
    if missing == "item_type":
        item.pop("item_type")
    if missing == "log_lines":
        item.pop("log_lines")
    if missing == "pid":
        item.pop("pid")
    if missing == "server":
        domain.pop("server")
    if missing == "instance_id":
        domain.pop("instance_id")
    payload = {"log_domain": domain, "log_items": [item]}
    if missing == "log_domain":
        payload.pop("log_domain")
    if missing == "log_items":
        payload.pop("log_items")
    return [payload]


def build_base(server_id="server-a", **kwargs):
    base = {
        "logic_device_id": "0",
        "phy_device_id": "0",
        "device_ip": "1.0.0.0",
        "server_id": server_id,
        "rank_map": {"group_x": {"rank_num": 1, "rank_id": "0"}},
        "root_list": ["group_x"],
        "timeout_param": {},
    }
    base.update(kwargs)
    return base


def build_error(**kwargs):
    error = {
        "first_error_time": "2024-04-08-02:55:04.903511",
        "first_error_module": "HCCL",
        "timeout_error_events_list": [],
        "cqe_links": [],
        "cluster_exception": {},
    }
    error.update(kwargs)
    return error


def build_diag_input(missing=None, base=None, error=None, show_logs=None):
    item = {
        "base": build_base() if base is None else base,
        "error": build_error() if error is None else error,
        "show_logs": {"normal": [], "error": []} if show_logs is None else show_logs,
    }
    if missing:
        item.pop(missing)
    return [{"3199": item}]


class STParseRootCluster(unittest.TestCase):
    def assert_parse_failed(self, payload, expected):
        results, err_msg_list = parse_root_cluster(payload)
        self.assertEqual(results, [])
        self.assertTrue(err_msg_list)
        self.assertIn(expected, err_msg_list[0])

    def test_security_0013_missing_instance_id(self):
        """Ascend_Fd_SDK_Security_0013"""
        self.assert_parse_failed(build_parse_input("instance_id"), "instance_id")

    def test_security_0014_missing_server(self):
        """Ascend_Fd_SDK_Security_0014"""
        self.assert_parse_failed(build_parse_input("server"), "log_domain.server")

    def test_security_0015_missing_log_domain(self):
        """Ascend_Fd_SDK_Security_0015"""
        self.assert_parse_failed(build_parse_input("log_domain"), "log_domain")

    def test_security_0016_missing_item_type(self):
        """Ascend_Fd_SDK_Security_0016"""
        self.assert_parse_failed(build_parse_input("item_type"), "item_type")

    def test_security_0017_missing_log_lines(self):
        """Ascend_Fd_SDK_Security_0017"""
        self.assert_parse_failed(build_parse_input("log_lines"), "log_lines")

    def test_security_0018_missing_pid(self):
        """Ascend_Fd_SDK_Security_0018"""
        self.assert_parse_failed(build_parse_input("pid"), "pid")

    def test_security_0019_missing_log_items(self):
        """Ascend_Fd_SDK_Security_0019"""
        self.assert_parse_failed(build_parse_input("log_items"), "log_items")


class STDiagRootCluster(unittest.TestCase):
    def setUp(self):
        # ErrorChecker keeps the timeout info map as class-level state; isolate and restore it.
        original_timeout_map = ErrorChecker.timeout_error_info_map
        ErrorChecker.timeout_error_info_map = {}
        self.addCleanup(setattr, ErrorChecker, "timeout_error_info_map", original_timeout_map)

    def test_reliability_0007_empty_list(self):
        """Ascend_Fd_SDK_Reliability_0007"""
        results, err_msg_list = diag_root_cluster([])
        self.assertEqual(results, {})
        self.assertIn("The input list is empty", err_msg_list[0])

    def test_security_0023_server_id_empty(self):
        """Ascend_Fd_SDK_Security_0023"""
        results, _ = diag_root_cluster(build_diag_input(base=build_base(server_id="")))
        self.assertTrue(results.get("analyze_success"))
        self.assertTrue(results.get("first_error_device", "").startswith("unknown_server_0"))

    def test_security_0024_first_error_time_not_time_format(self):
        """Ascend_Fd_SDK_Security_0024"""
        results, _ = diag_root_cluster(build_diag_input(error=build_error(first_error_time="not_a_time")))
        self.assertTrue(results.get("analyze_success"))
        self.assertIn(
            results.get("fault_description", {}).get("code"),
            (fault_description.PART_ERROR_WITH_NO_TIMEOUT.code, fault_description.ALL_NO_ERROR.code),
        )

    def test_security_0025_show_logs_error_empty_list(self):
        """Ascend_Fd_SDK_Security_0025"""
        results, _ = diag_root_cluster(build_diag_input(show_logs={"normal": [], "error": []}))
        self.assertTrue(results.get("analyze_success"))
        self.assertIn(
            results.get("fault_description", {}).get("code"),
            (fault_description.PART_ERROR_WITH_NO_TIMEOUT.code, fault_description.ALL_NO_ERROR.code),
        )

    def test_security_0020_server_id_not_str(self):
        """Ascend_Fd_SDK_Security_0020"""
        results, _ = diag_root_cluster(build_diag_input(base=build_base(server_id=123)))
        self.assertTrue(results.get("analyze_success"))
        self.assertTrue(results.get("first_error_device", "").startswith("unknown_server_0"))

    def test_security_0021_show_logs_error_not_list(self):
        """Ascend_Fd_SDK_Security_0021"""
        results, _ = diag_root_cluster(build_diag_input(show_logs={"normal": [], "error": {}}))
        self.assertTrue(results.get("analyze_success"))

    def test_security_0022_start_train_time_not_str(self):
        """Ascend_Fd_SDK_Security_0022"""
        item = build_diag_input()[0]["3199"]
        item["start_train_time"] = 1
        results, err_msg_list = diag_root_cluster([{"3199": item}])
        self.assertEqual(results, {})
        self.assertIn("Root Cluster diagnosis job failed", err_msg_list[0])

    def test_security_0026_base_missing(self):
        """Ascend_Fd_SDK_Security_0026"""
        results, _ = diag_root_cluster(build_diag_input(missing="base"))
        self.assertEqual(results.get("fault_description", {}).get("code"), fault_description.INVALID_DEVICE_ERROR.code)

    def test_security_0027_error_missing(self):
        """Ascend_Fd_SDK_Security_0027"""
        results, _ = diag_root_cluster(build_diag_input(missing="error"))
        self.assertEqual(results.get("fault_description", {}).get("code"), fault_description.ALL_NO_ERROR.code)

    def test_security_0028_show_logs_missing(self):
        """Ascend_Fd_SDK_Security_0028"""
        # The interface currently reads show_logs without a default value.
        with self.assertRaises(AttributeError):
            diag_root_cluster(build_diag_input(missing="show_logs"))


if __name__ == "__main__":
    unittest.main()
