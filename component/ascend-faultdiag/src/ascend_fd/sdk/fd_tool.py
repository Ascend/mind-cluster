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
import os
import json
import ctypes

from ascend_fd.utils.status import InnerError

UTF8 = 'utf-8'


class AttrKey:
    EVENT_ATTRIBUTE = 0
    EVENT_RULE = 1
    EVENT_ATTR_REGEX = 2


class InputLog(ctypes.Structure):
    _fields_ = [
        ("logData", ctypes.c_char_p),
        ("logType", ctypes.c_char_p),
        ("framework", ctypes.c_char_p)
    ]


class FDTool:
    CODE_BUFFER_LEN = 100
    INFO_BUFFER_LEN = 8192

    def __init__(self):
        so_dir = os.path.join(os.path.dirname(os.path.dirname(__file__)), "lib")
        so_file = "libfaultdiag.so"
        self._fd_tools_dll = ctypes.CDLL(os.path.join(so_dir, so_file))

    def data_match(self, log_data, log_type, framework_name=""):
        """
        Data Match.
        Args:
            logData:        <InputLog*> input log data struct
            matchEventCode: <char*> the event code which matched keywords
            codeLen:        <int> the code buffer length
            matchKeyInfo:   <char*> the keyInfo which matched
            infoLen:        <int> the info buffer length
        Returns:
            ret:            <int> 0 means successful, 1 means not matched, 2 means detected debugging
        """
        if not log_data:
            return "", ""
        self._fd_tools_dll.DataMatch.argtypes = [
            ctypes.POINTER(InputLog), ctypes.c_char_p, ctypes.c_int, ctypes.c_char_p, ctypes.c_int
        ]
        self._fd_tools_dll.DataMatch.restype = ctypes.c_int

        input_log = InputLog(log_data.encode(UTF8), log_type.encode(UTF8), framework_name.encode(UTF8))

        match_event_code = ctypes.create_string_buffer(self.CODE_BUFFER_LEN)
        match_key_info = ctypes.create_string_buffer(self.INFO_BUFFER_LEN)

        ret = self._fd_tools_dll.DataMatch(
            ctypes.byref(input_log), match_event_code, self.CODE_BUFFER_LEN, match_key_info, self.INFO_BUFFER_LEN
        )
        if ret == 0:
            match_code = match_event_code.value.decode(UTF8)
            match_key_info = match_key_info.value.decode(UTF8)
            return match_code, match_key_info
        if ret == 2:
            raise InnerError("Debugger detected, program terminated.")
        return "", ""

    def get_attr_value_dict(self, event_code, attr_key):
        """
        Get the dictionary attribute of the event, "attribute" or "rule"
        :param event_code: event code
        :param attr_key: key of the event attribute
        :return: the value of the event attribute
        """
        if attr_key != AttrKey.EVENT_ATTRIBUTE and attr_key != AttrKey.EVENT_RULE:
            return None
        attr_value = self.get_event_attr(event_code, attr_key)
        if not attr_value:
            return None
        try:
            attr_value_dict = json.loads(attr_value)
        except json.JSONDecodeError:
            attr_value_dict = None
        return attr_value_dict

    def get_attr_value_str(self, event_code, attr_key):
        """
        Get the string attribute of the event, "regex"
        :param event_code: event code
        :param attr_key: key of the event attribute
        :return: the value of the event attribute
        """
        if attr_key != AttrKey.EVENT_ATTR_REGEX:
            return None
        return self.get_event_attr(event_code, attr_key)

    def get_event_attr(self, event_code, attr_key):
        """
        Get the event attribute.
        Args:
            eventCode:     <const char*> event code
            key:           <int> attribute key. 0: EVENT_ATTRIBUTE; 1: EVENT_RULE; 2: EVENT_ATTR_REGEX.
            value:         <char*> the attribute string of key. If EVENT_ATTRIBUTE, return dict string; if EVENT_RULE,
                                   return list string; if EVENT_ATTR_REGEX, return string
            valueLen:      <int> the value buffer length
        Returns:
            ret:           <int> 0 means successful, 1 means inner error, 2 means detected debugging
        """
        self._fd_tools_dll.GetEventAttribute.argtypes = [ctypes.c_char_p, ctypes.c_int, ctypes.c_char_p, ctypes.c_int]
        self._fd_tools_dll.GetEventAttribute.restype = ctypes.c_int

        event_code = event_code.encode(UTF8)
        attr_value = ctypes.create_string_buffer(self.INFO_BUFFER_LEN)

        ret = self._fd_tools_dll.GetEventAttribute(event_code, attr_key, attr_value, self.INFO_BUFFER_LEN)
        if ret == 0:
            return attr_value.value.decode(UTF8)
        if ret == 2:
            raise InnerError("Debugger detected, program terminated.")
        return None

    def is_code_exist(self, event_code):
        """
        Get the event attribute.
        Args:
            eventCode:     <const char*> event code
        Returns:
            ret:           <int> 0 means exist, 1 means not exist, 2 means detected debugging
        """
        self._fd_tools_dll.IsCodeExist.argtypes = [ctypes.c_char_p]
        self._fd_tools_dll.IsCodeExist.restype = ctypes.c_int
        event_code = event_code.encode(UTF8)
        ret = self._fd_tools_dll.IsCodeExist(event_code)
        if ret == 2:
            raise InnerError("Debugger detected, program terminated.")
        return ret == 0
