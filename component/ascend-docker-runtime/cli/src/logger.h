/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2020-2022. All rights reserved.
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
#ifndef LOGGER_H
#define LOGGER_H

void Logger(const char *msg, int level, int screen);

#ifdef GOOGLE_TEST
STATIC void WriteLogFile(const char* filename, const long maxSize, const char* buffer, unsigned bufferSize);
STATIC long GetLogSize(const char* filename);
STATIC int CreateLog(const char* filename);
STATIC int GetCurrentLocalTime(char* buffer, int length);
STATIC bool LogConvertStorage(const char* filename, const long maxSize);
STATIC void DivertAndWrite(const char *logPath, const char *msg, const int level);
STATIC long GetLogSizeProcess(const char* path);
#endif

#endif
