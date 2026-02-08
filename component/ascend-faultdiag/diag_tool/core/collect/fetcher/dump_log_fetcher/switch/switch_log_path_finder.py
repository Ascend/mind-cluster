import os.path
from pathlib import Path
from typing import List, Tuple

from diag_tool.utils import file_tool


class SwitchLogPathFinder:
    _SLOT_DIR = "slot_1_0"
    _SEARCH_DEPTH = 4
    _EXCLUDE_TXT_SUBFIXES = ("zip", "gz", "dat")

    @classmethod
    def find(cls, root_dir: str) -> Tuple[List[str], List[str]]:
        slot_dirs = file_tool.find_all_sub_paths(root_dir, cls._SLOT_DIR, cls._SEARCH_DEPTH)
        diag_info_dirs = [os.path.dirname(slot_dir) for slot_dir in slot_dirs]
        cli_output_txt_paths = cls._find_cli_output_txt(root_dir, diag_info_dirs)
        return diag_info_dirs, cli_output_txt_paths

    @classmethod
    def _find_cli_output_txt(cls, root_dir: str, exclude_folders: List[str]):
        result_files = []
        root_path = Path(root_dir).resolve()
        exclude_names = set(exclude_folders)
        for current_dir, dirs, files in os.walk(root_path, topdown=True):
            # 检查当前目录是否应该被排除
            if current_dir in exclude_names:
                dirs.clear()  # 清空子目录列表，停止继续遍历
                continue

            # 过滤掉排除的文件夹名
            dirs[:] = [d for d in dirs if d not in exclude_names]

            for file_name in files:
                if file_name.endswith(cls._EXCLUDE_TXT_SUBFIXES):
                    continue
                file_path = os.path.join(current_dir, file_name)
                result_files.append(file_path)
        return result_files
