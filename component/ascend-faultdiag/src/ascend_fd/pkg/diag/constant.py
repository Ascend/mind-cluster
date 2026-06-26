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
# ==============================================================================

DECISION_TREE_MODULES = frozenset(
    [
        ('sklearn.ensemble._forest', 'RandomForestClassifier'),
        ('sklearn.tree._classes', 'DecisionTreeClassifier'),
        ('numpy', 'ndarray'),
        ('numpy', 'dtype'),
        ('numpy._core.numeric', '_frombuffer'),
        ('numpy.core.numeric', '_frombuffer'),
        ('numpy.core.multiarray', 'scalar'),
        ('numpy._core.multiarray', 'scalar'),
        ('numpy._core.multiarray', '_reconstruct'),
        ('numpy.core.multiarray', '_reconstruct'),
        ('sklearn.tree._tree', 'Tree'),
    ]
)
