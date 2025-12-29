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
import argparse
import warnings

import joblib
import pandas as pd
from sklearn.ensemble import RandomForestClassifier

warnings.filterwarnings("ignore", category=FutureWarning, module="sklearn", lineno=193)

FLAG = os.O_WRONLY | os.O_CREAT
PWD_PATH = os.path.dirname(os.path.realpath(__file__))
MODEL_FILE_PATH = os.path.join(PWD_PATH, 'net_rf_model.pt')

parser = argparse.ArgumentParser()
parser.add_argument('-model_path', type=str, default=MODEL_FILE_PATH)
parser.add_argument('-dataset_path', type=str, default=PWD_PATH)
parser.add_argument('-action', type=str, default='train')
args = parser.parse_args()

if __name__ == "__main__":
    input_X = pd.read_csv(os.path.join(args.dataset_path, args.action + '.csv'), index_col=0)
    input_X = input_X.sort_index()
    input_Y = input_X['label'].values.astype(int)
    input_X = input_X.drop(columns=['label'])

    if args.action == 'train':
        classifier = RandomForestClassifier(
            n_jobs=-1,
            class_weight="balanced",
            n_estimators=500,
            random_state=0
        )
        classifier.fit(input_X, input_Y)
        with os.fdopen(os.open(args.model_path, FLAG, 0o640), 'wb') as model_path:
            joblib.dump(classifier, model_path)


