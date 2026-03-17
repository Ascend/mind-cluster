#!/usr/bin/env python3
# coding: utf-8
# Copyright 2026 Huawei Technologies Co., Ltd
import os.path

from tests.st.lib.dl_deployer.dl import Installer


class NodedInstaller(Installer):
    component_name = 'noded'

    @staticmethod
    def get_labels():
        return ["nodeDEnable=on", ]

