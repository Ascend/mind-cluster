# coding=utf-8
# Copyright (c) 2024, HUAWEI CORPORATION.  All rights reserved.

from mindspeed.patch_utils import Patch


class PatchManager:
    """
    Manage patches and solve invalid patch problem caused by duplicate patch and other complicated problems.
    """
    _patch_info_collection = {}

    @staticmethod
    def register_patch(orig_func_name, new_func=None, force_patch=False, create_dummy=False):
        if orig_func_name not in PatchManager._patch_info_collection:
            PatchManager._patch_info_collection[orig_func_name] = Patch(orig_func_name, new_func, create_dummy)
        else:
            PatchManager._patch_info_collection.get(orig_func_name).set_patch_func(new_func, force_patch)

    @staticmethod
    def apply_patches():
        for patch in PatchManager._patch_info_collection.values():
            patch.apply_patch()
