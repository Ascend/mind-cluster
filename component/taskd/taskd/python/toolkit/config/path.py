#!/usr/bin/python3
# -*- coding: utf-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2022. All rights reserved.
import os

WORK_DIR = os.path.dirname(os.path.dirname(__file__))
CFG_DIR = os.path.join(WORK_DIR, "configuration")
LOG_CFG_FILE = os.path.join(CFG_DIR, "logger.yaml")
TORCH_EXTENSIONS_CACHE_DIR = os.path.expanduser("~/.cache/torch_extensions")
