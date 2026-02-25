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

import tarfile
import zipfile
from pathlib import Path

from toolkit.utils.logger import DIAG_LOGGER


class CompressTool:

    @staticmethod
    def extract_zip_recursive(zip_path: str, extract_to: str = None, max_depth=3, current_depth=1,
                              remove_original_zip=False):
        """
        递归解压ZIP文件，支持层级控制

        Args:
            zip_path: 主ZIP文件路径
            extract_to: 解压目标目录
            max_depth: 最大解压深度（1表示只解压主ZIP，不处理内部ZIP）
            current_depth: 当前深度（内部使用，不要手动设置）
            remove_original_zip: 解压后是否删除原始ZIP文件
        """
        zip_path = Path(zip_path)

        if extract_to is None:
            extract_to = zip_path.parent / zip_path.stem
        else:
            extract_to = Path(extract_to) / zip_path.stem
        extract_to = Path(extract_to)
        if extract_to.exists():
            return 0
        extract_to.mkdir(parents=True, exist_ok=True)

        # print(f"{'  ' * (current_depth - 1)}[深度{current_depth}] 解压: {zip_path.name} -> {extract_to.name}")

        # 检查是否达到最大深度
        if current_depth > max_depth:
            # print(f"{'  ' * (current_depth - 1)}达到最大深度 {max_depth}，停止解压内部ZIP")
            return 0

        zip_files_found = 0

        try:
            # 解压当前ZIP文件
            with zipfile.ZipFile(zip_path, 'r') as zip_ref:
                zip_ref.extractall(extract_to)

            # 递归查找并解压内部的ZIP文件（如果未达到最大深度）
            if current_depth < max_depth:
                zip_files_found = CompressTool._find_and_extract_zip_files(
                    extract_to, max_depth, current_depth, remove_original_zip
                )
            else:
                # 只统计但不解压
                zip_files_found = CompressTool._count_zip_files(extract_to)
                # print(f"{'  ' * current_depth}找到 {zip_files_found} 个内部ZIP文件（因达到最大深度未解压）")

            if remove_original_zip:
                zip_path.unlink()
                # print(f"{'  ' * (current_depth - 1)}已删除原始ZIP: {zip_path.name}")

        except zipfile.BadZipFile as e:
            DIAG_LOGGER.error(f"{'  ' * (current_depth - 1)}错误: {zip_path.name} 不是有效的ZIP文件 - {e}")
        except Exception as e:
            DIAG_LOGGER.error(f"{'  ' * (current_depth - 1)}解压错误: {zip_path.name} - {e}")

        return zip_files_found

    @staticmethod
    def _sanitize_filename(name: str) -> str:
        """
        替换文件名中的不支持字符
        :param name: 原始文件名
        :return: 处理后的文件名
        """
        # Windows和Linux不支持的字符
        invalid_chars = '<>'
        # 替换不支持的字符为下划线
        for char in invalid_chars:
            name = name.replace(char, '_')
        return name

    @staticmethod
    def _count_zip_files(directory):
        """统计目录中的ZIP文件数量"""
        directory = Path(directory)
        return sum(1 for item in directory.rglob('*')
                   if item.is_file() and item.suffix.lower() == '.zip')

    @staticmethod
    def _find_and_extract_zip_files(directory, max_depth, current_depth, remove_original_zip=False):
        """
        在目录中递归查找并解压ZIP文件（带深度控制）
        """
        directory = Path(directory)
        zip_files_found = 0

        for item in directory.rglob('*'):
            if item.is_file() and item.suffix.lower() == '.zip':
                zip_files_found += 1
                try:
                    # 递归解压（深度+1）
                    inner_zips = CompressTool.extract_zip_recursive(
                        str(item), str(item.parent), max_depth, current_depth + 1, remove_original_zip
                    )
                    zip_files_found += inner_zips

                except Exception as e:
                    DIAG_LOGGER.error(f"{'  ' * current_depth}解压错误: {item.name} - {e}")

        return zip_files_found

    @classmethod
    def extract_tar_gz(cls, tar_gz_path: str, extract_to: str):
        """
        在进程池中执行的解压任务
        :param tar_gz_path: .tar.gz文件路径
        :param extract_to: 解压目标目录
        """
        DIAG_LOGGER.info(f"开始解压: {tar_gz_path} 到 {extract_to}")
        try:
            # 创建不含.tar.gz后缀的同名文件夹
            tar_path = Path(tar_gz_path)
            # 从后往前去除.tar.gz后缀，以正确处理文件名中包含其他点号的情况
            folder_name = tar_path.name
            if folder_name.endswith('.tar.gz'):
                folder_name = folder_name[:-7]  # 去除.tar.gz后缀
            extract_folder = Path(extract_to) / folder_name
            # 如果文件夹已存在，跳过解压
            if extract_folder.exists():
                return True

            # 使用tarfile直接解压.tar.gz文件到指定文件夹
            with tarfile.open(tar_gz_path, 'r:gz') as tar:
                # 遍历压缩包内的所有成员
                for member in tar.getmembers():
                    # 检查成员名称是否包含特殊字符并替换
                    sanitized_name = cls._sanitize_filename(member.name)
                    # 如果名称被修改，则更新成员名称
                    if sanitized_name != member.name:
                        member.name = sanitized_name
                    # 特别处理链接目标名称
                    if member.issym() or member.islnk():
                        continue
                    tar.extract(member, path=str(extract_folder))
            DIAG_LOGGER.info(f"成功解压: {tar_gz_path} 到 {extract_folder}")
            return True
        except Exception as e:
            DIAG_LOGGER.error(f"解压失败 {tar_gz_path}: {e}")
            return False
