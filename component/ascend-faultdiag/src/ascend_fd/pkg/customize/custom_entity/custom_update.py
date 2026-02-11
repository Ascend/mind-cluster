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
import logging

from ascend_fd.utils.status import ParamError
from ascend_fd.utils.tool import safe_read_json, safe_write_open
from ascend_fd.configuration.config import DEFAULT_USER_CONF, KNOWLEDGE_GRAPH_CONF
from ascend_fd.pkg.customize.custom_entity.valid import CHECK_MAP, check_missing_attribute_when_add, code_check

logger = logging.getLogger("FAULT_DIAG")
echo = logging.getLogger("ECHO")
MAX_OP_NUM = 1000


def update_entity(data_path: str = None, sdk_entity: dict = None, output_dict: dict = None):
    """
    Update the fault entities by input data file
    :param data_path: str, the updated data file path
    :param sdk_entity: sdk entity to update
    :param output_dict: output dict for updated results
    :return operation result
    """
    origin_conf = safe_read_json(KNOWLEDGE_GRAPH_CONF)
    origin_entities = origin_conf.setdefault("knowledge-re pository", {})
    origin_entity_codes = set(origin_entities.keys())

    user_conf = safe_read_json(DEFAULT_USER_CONF) if os.path.exists(DEFAULT_USER_CONF) else {}
    user_entities = user_conf.setdefault("knowledge-repository", {})

    update_data = safe_read_json(data_path) if data_path is not None else sdk_entity
    exist_failed_codes = []
    add_failed_codes = []
    update_failed_codes = []
    op_time = 0

    for code, entity in update_data.items():
        op_time += 1
        if op_time > MAX_OP_NUM:
            echo.warning("A maximum of %s codes can be updated at a time. Extra codes will be ignored.", MAX_OP_NUM)
            logger.warning("A maximum of %s codes can be updated at a time. Extra codes will be ignored.", MAX_OP_NUM)
            break
        if code in origin_entity_codes:
            logger.error("Entity(%s) already exists in the default fault entity set.", code)
            exist_failed_codes.append(code)
            continue
        old_entity = user_entities.get(code, {})
        if not old_entity and not _add_single_entity(code, entity, user_entities):
            add_failed_codes.append(code)
        if old_entity and not _update_single_entity(code, entity, user_entities):
            update_failed_codes.append(code)
    if data_path is not None:
        with safe_write_open(DEFAULT_USER_CONF, mode="w+", encoding="utf-8") as file_stream:
            file_stream.write(json.dumps(user_conf, ensure_ascii=False, indent=4))
    if output_dict is not None:
        output_dict.update(user_conf)
    if not any((*exist_failed_codes, *add_failed_codes, *update_failed_codes)):
        echo.info("Updated entity successfully.")
        return

    if exist_failed_codes:
        echo.error("Fault codes %s already exist in the default fault entity set, these codes fail to be updated.",
                   exist_failed_codes)
    if add_failed_codes:
        echo.error("Fault codes %s fail to verify parameters, these codes fail to be add.", add_failed_codes)
    if update_failed_codes:
        echo.error("Fault codes %s fail to verify parameters, these codes fail to be updated.", update_failed_codes)

    if len(exist_failed_codes) + len(add_failed_codes) + len(update_failed_codes) == len(update_data.keys()):
        raise ParamError("All codes update failed, please check the input json file content.")


def _update_single_entity(code, updated_entity, user_entities):
    """
    Update the existing fault entity
    :param code: str, the fault code that need to updated
    :param updated_entity: dict, the updated data
    :param user_entities: dict, user entities conf
    :return: indicates whether the update is successful. bool value.
    """
    entity = user_entities.get(code, {})
    user_entity_codes = set(user_entities.keys())
    if not _check_and_update_entity(code, updated_entity, entity, user_entity_codes):
        logger.error("Update entity(%s) failed because some attribute fail to be verified.", code)
        return False
    user_entities.update({code: entity})
    logger.info("Update entity(%s) successfully.", code)
    return True


def _add_single_entity(code, add_entity, user_entities):
    """
    Add a new fault entity
    :param code: str, the fault code that need to updated
    :param add_entity: dict, the add data
    :param user_entities: dict, user entities conf
    :return: indicates whether the adding is successful. bool value.
    """
    if not code_check(code):
        logger.error("The new code '%s' is invalid when add new entity.", code)
        return False
    missing_attr = check_missing_attribute_when_add(set(add_entity.keys()))
    if missing_attr:
        logger.error("Add entity(%s) failed because some required attribute %s are missing.", code, missing_attr)
        return False
    new_entity = {}
    user_entity_codes = set(user_entities.keys())
    if not _check_and_update_entity(code, add_entity, new_entity, user_entity_codes):
        logger.error("Add entity(%s) failed because some attribute fail to be verified.", code)
        return False
    user_entities[code] = new_entity
    logger.info("Add entity(%s) successfully.", code)
    return True


def _check_and_update_entity(code, input_entity, new_entity, user_entity_codes):
    """
    Check the input entity and update the data that passes the verification
    :param code: customized entity code
    :param input_entity: dict, input entity data from updated file
    :param new_entity: dict, the new entity after updating
    :param user_entity_codes: set, user entity codes set, contain origin_entity and user_entity
    :return: indicates whether the verification is successful. bool value.
    """
    for key, value in input_entity.items():
        if key not in CHECK_MAP:
            logger.warning("The key '%s' is invalid.", key)
            continue
        if key == "rule":
            check_res = CHECK_MAP.get(key)(code, value, user_entity_codes)
        else:
            check_res = CHECK_MAP.get(key)(value)
        if not check_res:
            logger.error("The validity of '%s' is false.", key)
            return False
        if "." not in key:
            new_entity[key] = value
            continue
        major_part, minor_part = key.split(".")
        new_entity.setdefault(major_part, {})[minor_part] = value
    return True
