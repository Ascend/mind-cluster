import os

from diag_tool.core.collect.fetcher.dump_log_fetcher.switch.cli_output_txt.cli_output_txt_parser import \
    SwiCliOutputTxtParser
from diag_tool.core.collect.fetcher.dump_log_fetcher.switch.diag_info_output.diag_info_output_parser import \
    SwitchDiagnoseInformationParser


class SwiCliOutputParser:
    _DIAG_INFO_START = "======"

    @classmethod
    def parse(cls, file_path: str) -> dict:
        if not os.path.exists(file_path):
            return {}
        try:
            with open(file_path, 'r', encoding="utf8") as f:
                content = f.read()
        except UnicodeDecodeError:
            return {}
        if not content:
            return {}
        if content.startswith(cls._DIAG_INFO_START):
            data = SwitchDiagnoseInformationParser(content).parse()
        else:
            data = SwiCliOutputTxtParser(content).parse()
        return data
