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

from ascend_fd.utils.status import ParamError, InnerError
from ascend_fd.utils.tool import safe_write_open, safe_read_json, DEFAULT_BLACKLIST_CONF

bl_logger = logging.getLogger('FAULT_DIAG')
echo = logging.getLogger("ECHO")
MAX_BLACKLIST_NUM = 50
MAX_WORD_NUM = 10
MIN_WORD_LENGTH = 1
MAX_WORD_LENGTH = 200
DANGER_WORD_LENGTH = 10


def start_blacklist_job(args):
    """
    Start blacklist job
    :param args: blacklist option
    :return: None
    json format
    { blacklist:[['ERR1','CONTEXT1'],
    ['ERR2','CONTEXT2']] }
    """
    if args.force and not (args.delete or args.file):
        raise ParamError("The '-f --force' parameter can only be used together with '-d --delete' or '-f --file'.")
    blacklist_manager = BlackListManager()
    if args.add:
        blacklist_manager.add_blacklist(args.add)
        echo.info('add blacklist success')
        return
    if args.delete:
        blacklist = blacklist_manager.get_blacklist()
        contents = _delete_items_show(args.delete, blacklist)
        if args.force or _user_confirm("delete items", contents):
            blacklist_manager.delete_blacklist_keywords_by_key(args.delete)
            echo.info('delete blacklist success')
        return
    if args.show:
        blacklist_manager.show_blacklist()
        return
    if args.file:
        if args.force or _user_confirm("apply file", args.file, ', this will replace the former file'):
            blacklist_manager.switch_custom_file(args.file)
            echo.info('apply blacklist success')
        return


def _delete_items_show(ids: list, blacklist: list):
    """
    delete items return
    :param ids: int array like [1,2,3]
    :return: formatted string as show function
    1. xxxx, xxxx
    """
    id_set = set(ids)
    return_string_list = []
    for i, blacklist_item in enumerate(blacklist):
        if i in id_set:
            return_string_list.append("{}. {}".format(str(i), ', '.join(blacklist_item)))
    return_string = "\n".join(return_string_list)
    return "\n{}\n".format(return_string)


def _user_confirm(operation, contents, message=''):
    if input(f"please confirm to {operation}: {contents} {message}(enter 'y' or 'n'): ").lower() == 'y':
        return True
    echo.info(f"user cancelled to {operation}")
    bl_logger.info("user cancelled to %s", operation)
    return False


class BlackListManager:
    DEFAULT_BLACKLIST = [
        ["Ioctl", "failed"],
        ["Failed", "to", "get", "tgid", "by", "pid"],
        ["DRV", "cannot", "find", "docker", "in"]
    ]

    def __init__(self):
        self._blacklist = self._read_blacklist()

    @staticmethod
    def _write_blacklist(blacklist):
        max_length = min(len(blacklist), MAX_BLACKLIST_NUM)
        if len(blacklist) > max_length:
            echo.warning('blacklist is over %d , the oldest will be replaced !', MAX_BLACKLIST_NUM)
        new_blacklist = blacklist[-max_length:]
        for ignore_words in new_blacklist:
            if not isinstance(ignore_words, list):
                bl_logger.error('invalid blacklist format!')
                raise ParamError('blacklist item must be list')
            if len(ignore_words) > MAX_WORD_NUM:
                bl_logger.error('blacklist items over limit!')
                raise ParamError('blacklist length must under {}'.format(MAX_WORD_NUM))
            for word in ignore_words:
                if not isinstance(word, str):
                    bl_logger.error('invalid blacklist format!')
                    raise ParamError('word in blacklist must be string')
                if len(word) > MAX_WORD_LENGTH or len(word) < MIN_WORD_LENGTH:
                    raise ParamError(f"word length in blacklist must between {MIN_WORD_LENGTH} and {MAX_WORD_LENGTH}")

        config_dict = {'blacklist': new_blacklist}
        with safe_write_open(DEFAULT_BLACKLIST_CONF, mode="w+", encoding="utf-8") as file_stream:
            json.dump(config_dict, file_stream, ensure_ascii=False, indent=4)

    @classmethod
    def _read_blacklist(cls):
        try:
            config_dict = safe_read_json(DEFAULT_BLACKLIST_CONF)
        except FileNotFoundError:
            config_dict = {}
            with safe_write_open(DEFAULT_BLACKLIST_CONF, mode="w+", encoding="utf-8") as file_stream:
                json.dump(config_dict, file_stream, ensure_ascii=False, indent=4)
        blacklist = config_dict.get('blacklist', [])
        max_length = min(len(blacklist), MAX_BLACKLIST_NUM)
        return blacklist[-max_length:]

    def is_log_line_need_ignore(self, line: str):
        bid, _ = self._get_first_black_list_by_line(line)
        return bid is not None

    def get_blacklist(self):
        return self._blacklist

    def add_blacklist(self, keywords):
        if len(keywords) == 1 and len(keywords[0]) < DANGER_WORD_LENGTH:
            echo.warning('you may block most lines by ignoring this word')
        blacklist = self._blacklist
        blacklist.append(keywords)
        self._write_blacklist(blacklist)
        bl_logger.info('create black list success. the key words are %s', ','.join(keywords))

    def show_blacklist(self):
        blacklist = self._blacklist
        echo.info('[BLACKLIST]')
        for index, ignore_words in enumerate(blacklist):
            if not isinstance(ignore_words, list):
                raise InnerError('the json file is not illegal with the value {}'.format(str(ignore_words)))

            blacklist_line = ', '.join(ignore_words)
            echo.info('%d. %s', index, blacklist_line)

        bl_logger.info('show blacklist of custom success')

    def switch_custom_file(self, file_path):
        custom_file = safe_read_json(file_path)
        custom_blacklist = custom_file.get('blacklist', [])
        self._write_blacklist(custom_blacklist)
        bl_logger.info('apply blacklist of custom success')

    def delete_blacklist_keywords_by_key(self, keywords_ids):
        new_list = []
        blacklist = self._blacklist
        delete_ids = set(keywords_ids)
        delete_ids_len = len(delete_ids)
        for index, ignore_words in enumerate(blacklist):
            if index not in delete_ids:
                new_list.append(ignore_words)
            else:
                delete_ids.remove(index)

        # 如果所有的id都删除失败，抛出异常
        if len(delete_ids) == delete_ids_len:
            bl_logger.warning('All ids delete failed, please check whether the ids exists.')
            raise ParamError('All ids delete failed, please check whether the ids exists.')

        # 判断用户输入中是否有不存在的数字
        if len(delete_ids) != 0:
            echo.warning('you have chosen the number that may not exist, which includes %s',
                         ','.join(str(i) for i in delete_ids))
            bl_logger.warning('the input ids are not all deleted')
        self._write_blacklist(new_list)
        bl_logger.info('blacklist deleted success')

    def _get_first_black_list_by_line(self, line: str):
        for bid, keywords in enumerate(self._blacklist + self.DEFAULT_BLACKLIST):
            all_keywords_present = all(keyword in line for keyword in keywords)
            if all_keywords_present:
                return bid, keywords
        return None, None
