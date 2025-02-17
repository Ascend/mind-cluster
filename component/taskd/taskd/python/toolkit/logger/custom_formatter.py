#   Copyright (C)  2022. Huawei Technologies Co., Ltd. All rights reserved.
import logging


# Truncate the log, if the length of the log within a line exceeds MAX_LINE_LENGTH
class MaxLengthFormatter(logging.Formatter):
    def __init__(self, fmt, max_length):
        super().__init__(fmt=fmt)
        self.max_length = max_length

    def format(self, record):
        msg = super().format(record)
        # The repr() function will escape characters like \r and \n.
        # The repr() function adds quotation marks at the beginning and end of a string; these need to be removed.
        msg = repr(msg)[1:-1]
        if len(msg) > self.max_length:
            return msg[:self.max_length] + '...'
        return msg