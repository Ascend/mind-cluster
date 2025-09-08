#!/usr/bin/env python3
# -*- coding: utf-8 -*-
# Copyright (c) 2025, Huawei Technologies Co., Ltd. All rights reserved.
class swap_prefetch():
    prefetch_data_ptr_list = []
    prefetch_list = []
    slice_tensor_storage_ptr_list = []
    swap_tensors = []
    data_ptr = {}
    cur_micro_num = 0
    remove_num = 0
    forward_flag = False
    slice_tensor_storage_ptr = {}


class SwapPrefetch():
    swap_prefetch = swap_prefetch()


