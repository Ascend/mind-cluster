#  Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.
import logging
import logging.handlers
import sys
import re

LOGGER = logging.getLogger('torch_mindio_logger')
LOGGER.propagate = False


class EscapeSpecialCharsFilter(logging.Filter):
    def filter(self, record):
        # Blacklist
        record.msg = re.sub(r'(%0a|%0A|%0d|%0D|%0c|%0C|%08|%09|%7f|%7F|%0b|%0B)',
                            '', str(record.msg))
        # Whitelist
        allowed_chars_pattern = re.compile(r'[^\x20-\x7E]')
        record.msg = allowed_chars_pattern.sub('', record.msg)
        # Replace multiple consecutive spaces with one space.
        record.msg = re.sub(r'\s+', ' ', record.msg)
        return True


LOGGER.addFilter(EscapeSpecialCharsFilter())
if not LOGGER.handlers:
    FORMATTER = logging.Formatter('[%(asctime)s][%(process)d][%(levelname)s][%(module)s:%(lineno)d]%(message)s')
    STREAM_HANDLER = logging.StreamHandler(stream=sys.stdout)
    STREAM_HANDLER.setFormatter(FORMATTER)
    LOGGER.addHandler(STREAM_HANDLER)
