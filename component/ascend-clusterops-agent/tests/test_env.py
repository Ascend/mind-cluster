#!/usr/bin/env python3
# -*- coding: utf-8 -*-
# Copyright 2026 Huawei Technologies Co., Ltd
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

"""有界环境变量解析：缺省、非法、越界、nan/inf 与正常值。"""

import math

import pytest

from agent_core.https.bounded_env import env_float, env_int

NAME = "CLUSTEROPS_TEST_ENV"


def test_missing_value_uses_default(monkeypatch):
    monkeypatch.delenv(NAME, raising=False)
    assert env_int(NAME, 7, 1, 10) == 7
    assert env_float(NAME, 1.5, 0.0, 10.0) == 1.5


def test_invalid_value_falls_back_with_warning(monkeypatch, caplog):
    monkeypatch.setenv(NAME, "not-a-number")
    assert env_int(NAME, 7, 1, 10) == 7
    assert env_float(NAME, 1.5, 0.0, 10.0) == 1.5
    assert NAME in caplog.text


@pytest.mark.parametrize("value", ["0", "11", "-3"])
def test_out_of_range_integer_falls_back(monkeypatch, caplog, value):
    monkeypatch.setenv(NAME, value)
    assert env_int(NAME, 7, 1, 10) == 7
    assert NAME in caplog.text


@pytest.mark.parametrize("value", ["nan", "inf", "-inf", "10.5"])
def test_non_finite_or_out_of_range_float_falls_back(monkeypatch, value):
    monkeypatch.setenv(NAME, value)
    result = env_float(NAME, 1.5, 0.0, 10.0)
    assert result == 1.5
    assert math.isfinite(result)


def test_valid_values_are_returned_unchanged(monkeypatch):
    monkeypatch.setenv(NAME, "5")
    assert env_int(NAME, 7, 1, 10) == 5
    monkeypatch.setenv(NAME, "2.5")
    assert env_float(NAME, 1.5, 0.0, 10.0) == 2.5
