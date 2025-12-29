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
import json
import logging
import os

from ascend_fd.utils.status import ParamError
from ascend_fd.utils.tool import safe_read_json, safe_write_open
from ascend_fd.configuration.config import DEFAULT_USER_CONF
from ascend_fd.pkg.customize.custom_entity.valid import RULE_DST_NAME


logger = logging.getLogger("FAULT_DIAG")
echo = logging.getLogger("ECHO")
MAX_OP_NUM = 1000


def delete_entity(code_list, force_flag):
    """
    Delete the fault entities by input code list
    :param code_list: List[str], the code list need to delete
    :param force_flag: bool, indicates whether to skip the deletion confirmation step
    """
    if not force_flag:
        check_res = input("The deleted entity cannot be restored. Please enter Y/y to confirm the deletion: ")
        if check_res.strip().lower() != "y":
            echo.info("The deletion operation is canceled by user.")
            logger.info("The deletion operation is canceled by user.")
            return
    user_conf = safe_read_json(DEFAULT_USER_CONF) if os.path.exists(DEFAULT_USER_CONF) else {}
    user_entities = user_conf.setdefault("knowledge-repository", {})

    failed_codes = []
    deleted_codes = []
    op_time = 0
    for code in code_list:
        if code in (deleted_codes + failed_codes):
            continue
        op_time += 1
        if op_time > MAX_OP_NUM:
            echo.warning("A maximum of %s codes can be deleted at a time. Extra codes will be ignored.", MAX_OP_NUM)
            logger.warning("A maximum of %s codes can be deleted at a time. Extra codes will be ignored.", MAX_OP_NUM)
            break
        entity = user_entities.pop(code, {})
        if not entity:
            logger.error("Can not find fault code %s in user-defined fault entity set "
                         "when the delete operation is performed.", code)
            failed_codes.append(code)
            continue
        deleted_codes.append(code)
        logger.info("Fault entity (code %s) is deleted successfully.", code)

    delete_rules_connect_to_codes(deleted_codes, user_entities)
    with safe_write_open(DEFAULT_USER_CONF, mode="w+", encoding="utf-8") as file_stream:
        file_stream.write(json.dumps(user_conf, ensure_ascii=False, indent=4))
    if failed_codes:
        # if all codes delete failed, raise err.
        if not deleted_codes:
            raise ParamError(
                "All codes delete failed, please check whether the codes exists in user-defined fault entity set.")
        # if part of codes delete failed, print failure codes.
        echo.error("Fault codes %s does not exist in user-defined fault entity set, these codes fail to be deleted.",
                   failed_codes)
    else:
        echo.info("Deleted entity successfully.")


def delete_rules_connect_to_codes(deleted_code, user_entities):
    """
    Delete the fault entities' rule related to the deleted code
    :param deleted_code: List[str], the fault entity need to delete
    :param user_entities: dict, user entities conf
    """
    for _, entity_attr in user_entities.items():
        new_rule = []
        for single_rule in entity_attr.get("rule", []):
            if single_rule.get(RULE_DST_NAME) in deleted_code:
                continue
            new_rule.append(single_rule)
        entity_attr["rule"] = new_rule
