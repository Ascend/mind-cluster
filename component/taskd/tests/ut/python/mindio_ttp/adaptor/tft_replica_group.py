#!/usr/bin/env python3
# -*- coding: utf-8 -*-
# Copyright (c) 2025, Huawei Technologies Co., Ltd. All rights reserved.
def get_repair_group():
    return None


def build_repair_group(group_list):
    pass


def destroy_repair_group():
    pass


def get_local_embedding_group():
    return None


def ttp_get_dp_cp_replica_group():
    return "dp_cp_replica_group"


def ttp_get_replica_dp_num():
    return 2


def ttp_get_dp_cp_replica_group_gloo():
    return 'dp_cp'


def ttp_get_dp_ep_replica_group():
    return 'dp_ep'


def ttp_get_dp_ep_replica_group_gloo():
    return 'dp_ep'


DP_CP_REPLICA_GROUP = 0
DP_CP_REPLICA_GROUP_GLOO = 0
REPLICA_NUM = 2
DP_EP_REPLICA_GROUP = None
DP_EP_REPLICA_GROUP_GLOO = None