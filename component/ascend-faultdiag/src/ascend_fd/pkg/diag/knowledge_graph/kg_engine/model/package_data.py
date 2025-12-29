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
import uuid
import logging
from typing import Dict

from ascend_fd.utils.tool import load_json_data

kg_logger = logging.getLogger("KG_ENGINE")


class PackageData:
    _EVENT_CODE = "event_code"
    _EVENT_ID = "event_id"
    _EVENT_SOURCE = "source_device"

    def __init__(self, root_device_list, pkg_data_path=""):
        """
        Init package data model
        :param root_device_list: root devices
        :param pkg_data_path: kg-parser file path
        """
        self.event_map = dict()
        self.event_codes = set()
        self.root_device_list = root_device_list
        self.fault_devices = set()
        if pkg_data_path:
            events_json = load_json_data(pkg_data_path)
            self.load_events(events_json)

    def load_events(self, events: Dict):
        """
        Load events context
        :param events: fault events context, the key is code
        """
        for event_code, event_list in events.items():
            if not isinstance(event_list, list):
                continue
            for event in event_list:
                event_source = event.get(self._EVENT_SOURCE, "Unknown")
                if not self._source_is_root_device(event_source):
                    continue
                if event_source != "Unknown":
                    self.fault_devices.add(event_source)
                event_id = event.get(self._EVENT_ID, str(uuid.uuid4()))
                event.update({self._EVENT_CODE: event_code})
                self.event_map.update({event_id: dict(event)})  # 保存为一个新的字典，避免引用传递
                self.event_codes.add(event_code)

    def load_single_device_events(self, all_event_map, root_device):
        """
        load single device events
        :param all_event_map: all events map
        :param root_device: root device
        """
        for event in all_event_map.values():
            source_device = event.get(self._EVENT_SOURCE, "Unknown")
            if source_device == root_device or source_device == "Unknown":
                event_id = event.get(self._EVENT_ID, str(uuid.uuid4()))
                self.event_map.update({event_id: dict(event)})  # 保存为一个新的字典，避免引用传递
                self.event_codes.add(event.get(self._EVENT_CODE, ""))

    def _source_is_root_device(self, event_source):
        if not self.root_device_list:
            return True
        if event_source != "Unknown" and event_source not in self.root_device_list:
            return False
        return True
