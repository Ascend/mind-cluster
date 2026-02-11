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
import json
import logging
import os
from typing import Dict, List

from ascend_fd.utils.tool import safe_read_open
from ascend_fd.utils.status import InfoIncorrectError, FileOpenError
from ascend_fd.utils.i18n import LANG

main_logger = logging.getLogger("FAULT_DIAG")


class EntityAttribute:
    def __init__(self, attributes: Dict):
        """
        Init entity attribute
        :param attributes: entity attribute
        """
        self.class_ = attributes.get("class", "")
        self.component = attributes.get("component", "")
        self.module = attributes.get("module", "")
        self.cause_zh = attributes.get("cause_zh", "")
        self.description_zh = attributes.get("description_zh", "")
        self.suggestion_zh = attributes.get("suggestion_zh", "")
        self.cause_en = attributes.get("cause_en", "")
        self.description_en = attributes.get("description_en", "")
        self.suggestion_en = attributes.get("suggestion_en", "")
        self.error_case = attributes.get("error_case", "")
        self.fixed_case = attributes.get("fixed_case", "")

    def to_json(self):
        """
        Converting class variables to json
        :return: json format of entity attribute
        """
        json_dict = {}
        for k, v in vars(self).items():
            if k == "class_":  # skip class_ key
                continue
            if k in ('error_case', 'fixed_case') and not v:  # skip the error_case\fixed_case when the value is null
                continue
            if k.startswith(("cause", "description", "suggestion")) and not k.endswith(LANG):  # skip different language
                continue
            json_dict.update({k: v})
        json_dict.update({'class': self.class_})  # use the class key to save class_ value
        return json_dict


class Rule:
    def __init__(self, dst_code=None, expression=None, **kwargs):
        """
        Init rule
        :param dst_code: code of destination entity
        :param expression: expression rule between source entity and destination entity
        :param kwargs: reserved parameter
        """
        self.dst_code = dst_code
        self.expression = expression


class SchemaEntity:

    def __init__(self, entity_code: str = "", attribute=None, rule: List[Rule] = None,
                 source_file="", regex: Dict = None, attr_regex: str = "", **kwargs):
        """
        Init Schema entity
        :param entity_code: code of entity
        :param attribute: attribute of entity
        :param rule: rule of entity
        :param parse_attr:
        :param source_file: file type of parse keywords
        :param regex: parse keywords of entity
        :param kwargs: reserved parameter
        """
        self.entity_code = entity_code
        self.attribute = EntityAttribute(attribute or {})
        self.rule = [Rule(**r) for r in (rule or [])]
        self.source_file = source_file
        self.regex = regex or {}
        self.attr_regex = attr_regex


class KgConfigParser:
    _KNOW_REPO = "knowledge-repository"

    def __init__(self, config_path_list: List = None, sdk_config_repo: dict = None):
        """
        Init kg-config.json parser
        :param config_path_list: path list of kg-config.json
        """
        self.entity_map: Dict[str, SchemaEntity] = {}
        self.load_config(config_path_list)
        if sdk_config_repo:
            self.add_single_config(sdk_config_repo)

    @abc.abstractmethod
    def add_single_config(self, json_obj):
        """
        Add single configuration
        :param json_obj:  json obj
        """
        pass

    @abc.abstractmethod
    def _save_single_entity(self, entity_code, schema_entity):
        """
        Save  of single entity
        :param entity_code:
        :param schema_entity:
        """
        pass

    def load_config(self, config_path_list):
        """
        Load all kg-config.json in config path list
        :param config_path_list: path list of kg-config.json
        """
        for config_path in config_path_list:
            if not os.path.exists(config_path) or not os.path.isfile(config_path):
                continue
            with safe_read_open(config_path, 'rb') as file_stream:
                try:
                    self.add_single_config(json.load(file_stream))
                except InfoIncorrectError as err:
                    main_logger.warning(
                        'The content obtained from the %s file is not a JSON.', os.path.basename(config_path))
                    raise InfoIncorrectError(
                        f"The content obtained from the [{os.path.basename(config_path)}] file is not a JSON.") from err
                except Exception as err:
                    main_logger.warning('Open %s failed', os.path.basename(config_path))
                    raise FileOpenError(f"Open {os.path.basename(config_path)} failed: {err}") from err
        main_logger.info("Add schema entities success.")

    def _load_know_repo(self, json_obj):
        """
        Load knowledge-repository json of kg-config.json
        :param json_obj: knowledge-repository json context
        """
        know_repo = json_obj.get(self._KNOW_REPO, dict())
        for entity_code, entity in know_repo.items():
            source_files = entity.get('source_file', '')
            source_file_list = [f.strip() for f in source_files.split("|") if f.strip()]
            if not source_file_list:
                source_file_list = [""]
            for single_source_file in source_file_list:
                new_entity = {**entity}
                new_entity["source_file"] = single_source_file
                schema_entity = SchemaEntity(**new_entity)
                if self.__class__ == Schema and entity_code in self.entity_map:
                    continue
                self._save_single_entity(entity_code, schema_entity)


class ParseRegexMap(KgConfigParser):
    _BLACKLIST = "blacklist"

    def __init__(self, config_pkg_list: List = None, sdk_config_repo: dict = None):
        """
        Init parse regex map
        :param config_pkg_list: path list of kg-config.json
        """
        self.log_mask_off_rules = list()
        self.parse_regex: Dict[str, Dict] = {}
        super().__init__(config_pkg_list, sdk_config_repo)

    def add_single_config(self, json_obj):
        self._load_mask_off_rules(json_obj)
        self._load_know_repo(json_obj)

    def get_mask_off_rules(self):
        return self.log_mask_off_rules

    def get_parse_regex(self):
        return self.parse_regex

    def _load_mask_off_rules(self, json_obj):
        self.log_mask_off_rules.extend(list(json_obj.get(self._BLACKLIST, dict()).values()))

    def _save_single_entity(self, entity_code, schema_entity):
        if schema_entity.source_file and schema_entity.regex:
            source_file_to_regex = self.parse_regex.setdefault(schema_entity.source_file, {})
            source_file_to_regex.setdefault(entity_code, {}).update(schema_entity.regex)
            if schema_entity.attr_regex:
                source_file_to_regex.setdefault(entity_code, {}).update({"attr_regex": schema_entity.attr_regex})


class Schema(KgConfigParser):

    def __init__(self, config_pkg_list: List = None, sdk_config_repo: dict = None):
        super().__init__(config_pkg_list, sdk_config_repo)

    def add_single_config(self, json_obj):
        self._load_know_repo(json_obj)

    def add_custom_event_to_schema(self, entity_code, schema_entity):
        schema_entity.entity_code = entity_code
        self.entity_map.update({entity_code: schema_entity})

    def get_schema_entity(self, entity_code):
        return self.entity_map.get(entity_code)

    def _save_single_entity(self, entity_code, schema_entity):
        self.entity_map[entity_code] = schema_entity
