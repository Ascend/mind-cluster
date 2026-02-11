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
import logging

from ascend_fd.utils.status import FileOpenError
from ascend_fd.utils.tool import safe_read_json
from ascend_fd.configuration.config import KNOWLEDGE_GRAPH_CONF
from ascend_fd.pkg.customize.custom_entity.valid import code_check, CHECK_MAP
from ascend_fd.utils.i18n import LANG


echo = logging.getLogger("ECHO")
logger = logging.getLogger("FAULT_DIAG")


def check_entity(conf_path):
    """
    Check the user-defined entity file
    :param conf_path: str, the entity file path that need to check
    :return: indicates whether the verification of user-defined entity file is successful. bool value.
    """
    try:
        user_entity_conf = safe_read_json(conf_path)
    except Exception as err:
        logger.error("Open check file path failed. The reason is: %s", err)
        raise FileOpenError(f"Open check file path failed: {err}") from err
    user_entities = user_entity_conf.setdefault("knowledge-repository", {})
    if not user_entities:
        echo.warning("Custom entity is empty. Please check whether the file content is correct.")
    origin_conf = safe_read_json(KNOWLEDGE_GRAPH_CONF)
    origin_entities = origin_conf.setdefault("knowledge-repository", {})

    entity_checker = EntityChecker(origin_entities, user_entities)
    if not entity_checker.check():
        return False
    echo.info("Custom entity verification passed.")
    logger.info("Custom entity verification passed.")
    return True


class EntityChecker:
    def __init__(self, origin_entities, user_entities):
        """
        Init Entity Checker
        :param origin_entities: dict, original entities
        :param user_entities: dict, user-defined entities
        """
        self.origin_entities = origin_entities
        self.user_entities = user_entities
        self.all_entity_codes = set(self.origin_entities.keys() | set(user_entities.keys()))

        self.current_entity_code = None
        self.required_attributes = None

    def check(self):
        for entity_code, entity_attr in self.user_entities.items():
            if entity_code in self.origin_entities:
                echo.error("Entity(%s) already exists in the default fault entity set. Check failed.", entity_code)
                logger.error("Entity(%s) already exists in the default fault entity set. Check failed.", entity_code)
                return False
            if not code_check(entity_code):
                echo.error("The entity code '%s' is invalid in user-defined entity. Check failed.", entity_code)
                logger.error("The entity code '%s' is invalid in user-defined entity. Check failed.", entity_code)
                return False
            self.current_entity_code = entity_code
            if not self.entity_check(entity_attr):
                return False
        return True

    def entity_check(self, entity):
        """
        Check each entity data
        :param entity: dict, each single entity attr. Contain attribute, rule, source_file and regex
        :return: indicates whether the verification is successful. bool value.
        """
        self.required_attributes = {
            "attribute.class", "attribute.component", "attribute.module", f"attribute.cause_{LANG}",
            f"attribute.description_{LANG}", f"attribute.suggestion_{LANG}", "source_file", "regex.in"
        }
        for key, value in entity.items():
            if key == "attribute" and not self._attribute_check(value):
                return False
            if key == "regex" and not self._regex_check(value):
                return False
            if key == "source_file" and not self._source_file_check(value):
                return False
            if key == "rule" and not CHECK_MAP.get("rule")(self.current_entity_code, value, self.all_entity_codes):
                echo.error("The 'rule' field of entity [%s] fails to be verified. Check failed.",
                           self.current_entity_code)
                logger.error("The 'rule' field of entity [%s] fails to be verified. Check failed.",
                             self.current_entity_code)
                return False
        if bool(self.required_attributes):
            echo.error("Some required attribute %s of entity [%s] are missing. Check failed.",
                       self.required_attributes, self.current_entity_code)
            logger.error("Some required attribute %s of entity [%s] are missing. Check failed.",
                         self.required_attributes, self.current_entity_code)
            return False
        return True

    def _source_file_check(self, source_file):
        """
        Check entity source file data
        :param source_file: str, entity source_file value
        :return: check result. bool value.
        """
        if not CHECK_MAP.get("source_file")(source_file):
            echo.error("The 'source_file' field of entity [%s] fails to be verified. Check failed.",
                       self.current_entity_code)
            logger.error("The 'source_file' field of entity [%s] fails to be verified. Check failed.",
                         self.current_entity_code)
            return False
        self.required_attributes.discard("source_file")
        return True

    def _attribute_check(self, attr):
        """
        Check entity attribute dict
        :param attr: dict, entity attribute dict
        :return: check result. bool value.
        """
        if not isinstance(attr, dict):
            echo.error("The 'attribute' field of entity [%s] fails to be verified. Check failed.",
                       self.current_entity_code)
            logger.error("The 'attribute' field of entity [%s] fails to be verified. Check failed.",
                         self.current_entity_code)
            return False
        for key, value in attr.items():
            name = f"attribute.{key}"
            if name not in CHECK_MAP:
                logger.warning("The '%s' field name of entity [%s] is invalid.", name, self.current_entity_code)
                continue
            if not CHECK_MAP.get(name)(value):
                echo.error("The '%s' field of entity [%s] fails to be verified. Check failed.", name,
                           self.current_entity_code)
                logger.error("The '%s' field of entity [%s] fails to be verified. Check failed.", name,
                             self.current_entity_code)
                return False
            self.required_attributes.discard(name)
        return True

    def _regex_check(self, regex):
        """
        Check entity regex dict
        :param regex: dict, entity regex dict
        :return: check result. bool value.
        """
        if not isinstance(regex, dict):
            echo.error("The 'regex' field of entity [%s] fails to be verified. Check failed.",
                       self.current_entity_code)
            logger.error("The 'regex' field of entity [%s] fails to be verified. Check failed.",
                         self.current_entity_code)
            return False
        for key, value in regex.items():
            name = f"regex.{key}"
            if name == "regex.regex":
                echo.error("The 'regex.regex' of entity [%s] does not support user-defined. Check failed.",
                           self.current_entity_code)
                logger.error("The 'regex.regex' of entity [%s] does not support user-defined. Check failed.",
                             self.current_entity_code)
                return False
            if name not in CHECK_MAP:
                logger.warning("The '%s' field name of entity [%s] is invalid.", name, self.current_entity_code)
                continue
            if not CHECK_MAP.get(name)(value):
                echo.error("The '%s' field of entity [%s] fails to be verified. Check failed.", name,
                           self.current_entity_code)
                logger.error("The '%s' field of entity [%s] fails to be verified. Check failed.", name,
                             self.current_entity_code)
                return False
            self.required_attributes.discard(name)
        return True
