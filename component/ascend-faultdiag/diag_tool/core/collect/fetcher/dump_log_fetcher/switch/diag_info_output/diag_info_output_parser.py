from diag_tool.core.collect.collect_config import SwiCliOutputDataType
from diag_tool.core.collect.fetcher.dump_log_fetcher.switch.base import SwitchOutputParser
from diag_tool.utils import helpers


class SwitchDiagnoseInformationParser(SwitchOutputParser):
    _SPLIT_PATTERN = '=' * 79

    def __init__(self, file_content: str):
        super().__init__()
        self.file_content = file_content

    def parse(self) -> dict:
        self.find_name(self.file_content)
        self.find_ip(self.file_content)
        parts = helpers.split_str(self.file_content, self._SPLIT_PATTERN)
        if not parts or len(parts) % 2 != 0:
            return {}
        new_parts = [parts[i] + parts[i + 1] for i in range(0, len(parts), 2)]
        for part in new_parts:
            self.parse_cli_output_part(part)
        return self.parse_data.get_data_dict()

    # 找名字
    def find_name(self, file_content: str):
        search = helpers.find_pattern_after_substrings(
            file_content, ["display current-configuration", "sysname"], r".*"
        )
        if search:
            self.parse_data.add_data([SwiCliOutputDataType.SWI_NAME.name], search.group().strip())
