/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2023. All rights reserved.
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
#include "retry_task_pool.h"
#include <ctime>
#include <fcntl.h>
#include <utility>
#include "background_log.h"
#include "memfs_api.h"
#include "service_configure.h"

using namespace ock::bg::util;
using namespace ock::memfs;
using namespace ock::common::config;

static std::string g_reportFormat1 =
    "{ \"Effect\": \"Checkpoint cannot be saved in storage system.\", \"EventCategory\": \"MindIO\", "
    "\"EventDescription\": \"MindIO dump checkpoint failure.\", \"EventID\": \"0x1001001\", \"EventName\": "
    "\"MindIO Dump Checkpoint Failure\", \"EventSubject\": \"MindIO\", \"EventType\": \"alert\", \"Parts\": \"\", "
    "\"PossibleCause\": \"1.Storage service is not available on the computing node. 2. Checkpoint data is incomplete "
    "or corrupt in MindIO.\", \"Severity\": \"Major\", \"Suggestion\": \"1. Check storage system correctly mounted on "
    "the computing node. \n2. Check additional error log recorded by MindIO.\", \"Status\": \"";

static std::string g_reportFormat2 = "\", \"ChName\": \"MindIO持久化检查点数据异常\"， \"TimeStamp\":\"";

static std::string g_reportFormat3 = "\" }RECORD_END";

RetryTaskPool::RetryTaskPool(RetryTaskConfig &config) noexcept
    : threadPool(config.thCnt, config.name + "_ThreadPool"),
      timerExecutor(config.name + "_Timer"),
      autoEvictFile { config.autoEvictFile },
      maxFailCntForUnserviceable { config.maxFailCntForUnserviceable },
      maxRetryTimes { config.retryTimes },
      maxRetryIntervalMs { std::chrono::milliseconds(config.retryIntervalSec * 1000L) },
      firstWaitTime(std::chrono::milliseconds(config.firstWaitMs))
{}

RetryTaskPool::~RetryTaskPool() = default;

void RetryTaskPool::Submit(const std::function<bool()> &task) noexcept
{
    auto retryTask = std::make_shared<RetryTask>(task, *this);
    threadPool.Push([retryTask]() { retryTask->Process(retryTask); });
}

int RetryTaskPool::Start() noexcept
{
    reportPath = ServiceConfigure::GetInstance().GetWorkPath();

    std::string::size_type position = reportPath.find_last_of('/');
    if (position == std::string::npos) {
        BKG_LOG_ERROR("create logger failed : invalid folder path.");
        return -1;
    }
    reportPath.append("/ccae/ockiod_report_CCAE.json");

    BKG_LOG_INFO("report CCAE config file");
    return threadPool.Start();
}

void RetryTaskPool::ReportCCAE(bool serviceable) noexcept
{
    auto fd = open(reportPath.c_str(), O_CREAT | O_WRONLY | O_TRUNC, S_IRUSR | S_IWUSR | S_IRGRP);
    if (fd < 0) {
        BKG_LOG_ERROR("Failed to Open CCAE file, err " << errno);
        return;
    }

    std::time_t now = std::time(nullptr);
    int maxTimeBuff = 128;
    char formattedBuff[maxTimeBuff];
    bzero(formattedBuff, maxTimeBuff);
    std::strftime(formattedBuff, sizeof(formattedBuff), "%Y-%m-%dT%H:%M:%S%z", std::localtime(&now));

    std::string formattedTime = formattedBuff;
    std::string::size_type position = formattedTime.find_last_of('+');
    if (position == std::string::npos) {
        position = formattedTime.find_last_of('-');
        if (position == std::string::npos) {
            BKG_LOG_ERROR("Failed to format time trace, str " << formattedTime.c_str());
            close(fd);
            return;
        }
    }

    // change time zone "+0800" to "+08:00"
    formattedTime.insert(position + 3u, 1u, ':');

    std::string inputCtx;
    if (serviceable) {
        inputCtx = g_reportFormat1 + "Cleared" + g_reportFormat2 + formattedTime + g_reportFormat3;
    } else {
        inputCtx = g_reportFormat1 + "Uncleared" + g_reportFormat2 + formattedTime + g_reportFormat3;
    }

    auto numBytes = write(fd, inputCtx.c_str(), inputCtx.size());
    if (numBytes < 0) {
        BKG_LOG_ERROR("Failed to write CCAE file, err " << errno);
        close(fd);
        return;
    }

    close(fd);
}

RetryTask::RetryTask(std::function<bool()> task, RetryTaskPool &pool) noexcept
    : retryTimes { 0 }, realTask { std::move(task) }, retryTaskPool(pool)
{}

void RetryTask::Process(const std::shared_ptr<RetryTask> &task) noexcept
{
    if (task->realTask()) {
        auto oldCnt = task->retryTaskPool.failCnt.exchange(0);
        if (oldCnt >= task->retryTaskPool.maxFailCntForUnserviceable) {
            BKG_LOG_INFO("Enable file system serviceable");
            MemFsApi::Serviceable(true);
            task->retryTaskPool.ReportCCAE(true);
        }
        return;
    }

    auto failCnt = task->retryTaskPool.failCnt.fetch_add(1);
    if (failCnt == task->retryTaskPool.maxFailCntForUnserviceable) {
        BKG_LOG_INFO("Disable file system serviceable");
        MemFsApi::Serviceable(false);
        task->retryTaskPool.ReportCCAE(false);
    }

    if (++task->retryTimes >= task->retryTaskPool.maxRetryTimes && task->retryTaskPool.autoEvictFile) {
        return;
    }

    auto curTime = (1U << (task->retryTimes - 1)) * task->retryTaskPool.firstWaitTime;
    auto waitTime = std::min(curTime, task->retryTaskPool.maxRetryIntervalMs);
    auto timerTask = [task]() { Waiting(task); };
    task->retryTaskPool.timerExecutor.Submit(timerTask, waitTime);
}

void RetryTask::Waiting(const std::shared_ptr<RetryTask> &task) noexcept
{
    task->retryTaskPool.threadPool.Push([task]() { Process(task); });
}