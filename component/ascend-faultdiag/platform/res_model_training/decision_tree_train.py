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
from typing import List

import joblib
import pandas as pd
from sklearn.tree import DecisionTreeClassifier

from ascend_fd.pkg.diag.node_anomaly.resource_preemption.rp_detection import preprocess_data

PWD_PATH = os.path.dirname(os.path.realpath(__file__))
DATA_PATH = os.path.join(PWD_PATH, 'data')
MODEL_FILE_PATH = os.path.join(PWD_PATH, 'cpu_decision_tree')
FAULT_FLAG = "fault_flag"
TIME_COLUMN_NAME = "time"
FLAG = os.O_WRONLY | os.O_CREAT

parser = argparse.ArgumentParser()
parser.add_argument('-model_path', type=str, default=MODEL_FILE_PATH)
parser.add_argument('-data_path', type=str, default=DATA_PATH)
args = parser.parse_args()


def decision_tree_train(train_df, saved_model_path, train_feature: List[str] = None, max_depth: int = 5):
    """
    Train decision tree models
    :param train_df: Dataframe of selected features and labels
    :param saved_model_path: path for saving models
    :param train_feature: features used for training
    :param max_depth：Max depth of the tree
    """
    if train_feature:
        if FAULT_FLAG not in train_feature:
            train_feature.append(FAULT_FLAG)
        train_df = train_df[train_feature]
    model = DecisionTreeClassifier(random_state=0, max_depth=max_depth)
    # Split the data into x_train and y_train
    if TIME_COLUMN_NAME in train_df.columns:
        train_df = train_df.drop(TIME_COLUMN_NAME, axis=1)
    x_train = train_df.drop(FAULT_FLAG, axis=1)
    y_train = train_df[FAULT_FLAG]
    # Train the model
    model.fit(x_train, y_train)
    # Save model
    with os.fdopen(os.open(saved_model_path, FLAG, 0o640), 'wb') as model_path:
        joblib.dump(model, model_path)


if __name__ == '__main__':
    process_pre_df_list = list()
    for dirname in os.listdir(args.data_path):
        process_df = pd.read_csv(os.path.join(args.data_path, dirname, 'process.csv'))
        process_pre_df = preprocess_data(process_df)
        process_pre_df_list.append(process_pre_df)
    process_input_df = pd.concat(process_pre_df_list)
    # cpu preemption model training
    decision_tree_train(process_input_df, args.model_path, ['cpu'])
