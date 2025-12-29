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
import abc
import inspect
import json
from typing import Dict, Any
from ascend_fd.utils.custom_typing import get_origin, get_args


class JsonObj(metaclass=abc.ABCMeta):

    @staticmethod
    def _is_json_dict_in_typing_list(instance, sign_class):
        return isinstance(instance, list) and \
            hasattr(sign_class, "__args__") and \
            isinstance(getattr(sign_class, "__args__"), tuple) and \
            len(sign_class.__args__) == 1 and \
            issubclass(sign_class.__args__[0], JsonObj)

    @staticmethod
    def _is_json_obj_in_typing_dict(instance, sign_class):
        return isinstance(instance, dict) and \
            hasattr(sign_class, "__args__") and \
            isinstance(getattr(sign_class, "__args__"), tuple) and \
            len(sign_class.__args__) == 2 and \
            inspect.isclass(sign_class.__args__[1]) and \
            issubclass(sign_class.__args__[1], JsonObj)

    @classmethod
    def from_dict(cls, json_dict: Dict, check_parameter=False):
        sigs = inspect.signature(cls.__init__)
        args = []
        for arg_name, parameter in sigs.parameters.items():
            if arg_name == "self":
                continue
            value = json_dict.get(arg_name)
            if value is None:
                args.append(parameter.default)
                continue
            sign_class = parameter.annotation
            if check_parameter and not cls._check_type(value, sign_class):
                raise TypeError(
                    f"Field '{arg_name}' type mismatch. expected: {sign_class}, actual: {type(value)}, value: {value}"
                )
            if isinstance(sign_class, type) and issubclass(sign_class, JsonObj):
                value = sign_class.from_dict(value, check_parameter)
            elif cls._is_json_dict_in_typing_list(value, sign_class):
                value = [sign_class.__args__[0].from_dict(item, check_parameter) for item in value]
            elif cls._is_json_obj_in_typing_dict(value, sign_class):
                value = {key: sign_class.__args__[1].from_dict(val, check_parameter) for key, val in value.items()}
            args.append(value)
        return cls(*args)

    @classmethod
    def _check_type(cls, value: Any, expected_type: Any) -> bool:
        origin = get_origin(expected_type)
        args = get_args(expected_type)

        if origin is list:
            if not isinstance(value, list):
                return False
            if not args:
                return True
            item_type = args[0]
            return all(cls._check_type(v, item_type) for v in value)

        # JsonDict 子类
        if isinstance(expected_type, type) and issubclass(expected_type, JsonObj):
            return isinstance(value, dict)

        # 基本类型（int, str, bool等）
        return isinstance(value, expected_type)

    def to_dict(self):
        members = {}
        for key, value in vars(self).items():
            new_value = value
            if key.startswith('__'):
                continue
            if isinstance(value, JsonObj):
                new_value = value.to_dict()
            elif isinstance(value, list):
                new_value = []
                for item in value:
                    if isinstance(item, JsonObj):
                        new_value.append(item.to_dict())
                    else:
                        new_value.append(item)
            elif isinstance(value, dict) and value:
                new_value = {}
                for k, v in value.items():
                    if isinstance(v, JsonObj):
                        new_value[k] = v.to_dict()
                    else:
                        new_value[k] = v
            members[key] = new_value
        return members

    def to_json(self):
        return json.dumps(self.to_dict())
