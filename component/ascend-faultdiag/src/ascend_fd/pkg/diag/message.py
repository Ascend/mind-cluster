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
from ascend_fd.utils.i18n import get_note_msg_by_code
from ascend_fd.utils.json_dict import JsonObj


class NoteMsg(JsonObj):
    def __init__(self, code: str, note: str = ""):
        self.code = code
        self.note = note
        self.set_note_by_code(code)

    def __hash__(self):
        return hash(self.note)

    def set_note_by_code(self, code):
        self.note = get_note_msg_by_code(code)


class FormatNoteMsg(NoteMsg, JsonObj):
    def __init__(self, key: str):
        super().__init__(key)

    def format(self, *args, **kwargs):
        self.note = self.note.format(*args, **kwargs)
        return self


MULTI_RANK_NOTE_MSG = NoteMsg("MULTI_RANK_NOTE_MSG")

MAX_RANK_NOTE_MSG = NoteMsg("MAX_RANK_NOTE_MSG")

MAX_DEVICE_NOTE_MSG = NoteMsg("MAX_DEVICE_NOTE_MSG")

MAX_WORKER_CHAINS_NOTE_MSG = NoteMsg("MAX_WORKER_CHAINS_NOTE_MSG")

NET_SINGLE_WORKER_MSG = NoteMsg("NET_SINGLE_WORKER_MSG")

UNKNOWN_ROOT_ERROR_RANK = NoteMsg("UNKNOWN_ROOT_ERROR_RANK")

SOME_SUBTASKS_FAILED = NoteMsg("SOME_SUBTASKS_FAILED")

SOME_DEVICE_FAILED = NoteMsg("SOME_DEVICE_FAILED")

FAULT_CHAINS_NOTE = NoteMsg("FAULT_CHAINS_NOTE")

FAULT_CHAINS_MAX_NOTE = FormatNoteMsg("FAULT_CHAINS_MAX_NOTE")

NO_GROUP_RANK_INFO_NOTE = NoteMsg("NO_GROUP_RANK_INFO_NOTE")

REMOTE_LINKS_NOTE = NoteMsg("REMOTE_LINKS_NOTE")

REMOTE_LINKS_MAX_NOTE = NoteMsg("REMOTE_LINKS_MAX_NOTE")

MULTI_FAULT_IN_KNOWLEDGE_GRAPH = NoteMsg("MULTI_FAULT_IN_KNOWLEDGE_GRAPH")
