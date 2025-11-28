/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2025. All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

#ifndef OCK_TTP_PTHREAD_TIMEDWAIT_H
#define OCK_TTP_PTHREAD_TIMEDWAIT_H

#include <mutex>
#include "common_constants.h"
namespace ock {
namespace ttp {

class PthreadTimedwait {    // wait signal or overtime, instead of sem_timedwait
public:
    PthreadTimedwait() = default;
    ~PthreadTimedwait() = default;
    TResult Initialize()
    {
        signalFlag = false;
        int32_t attrInitRet = pthread_condattr_init(&cattr);
        int32_t setLockRet = pthread_condattr_setclock(&cattr, CLOCK_MONOTONIC);
        int32_t condInitRet = pthread_cond_init(&condTimeChecker, &cattr);
        int32_t mutexInitRet = pthread_mutex_init(&timeCheckerMutex, nullptr);
        if (attrInitRet || setLockRet || condInitRet || mutexInitRet) {
            return TTP_ERROR;
        }
        return TTP_OK;
    }

    int32_t PthreadTimedwaitSecs(long secs)
    {
        struct timespec ts {0, 0};
        int32_t ret = 0;

        pthread_mutex_lock(&this->timeCheckerMutex);
        clock_gettime(CLOCK_MONOTONIC, &ts);

        ts.tv_sec += secs;
        while (!this->signalFlag) {    // avoid spurious wakeup
            ret = pthread_cond_timedwait(&this->condTimeChecker, &this->timeCheckerMutex, &ts);
            if (ret == ETIMEDOUT) {    // avoid infinite loop
                break;
            }
        }
        this->signalFlag = false;
        pthread_mutex_unlock(&this->timeCheckerMutex);

        return ret;
    }

    // signal will NOT lost when call PthreadSignal before PthreadTimedwaitSecs, so we can proactive cleanup
    void SignalClean()
    {
        signalFlag = false;
    }

    int32_t PthreadSignal()
    {
        int32_t signalRet = 0;
        pthread_mutex_lock(&this->timeCheckerMutex);
        signalFlag = true;
        signalRet = pthread_cond_signal(&this->condTimeChecker);
        pthread_mutex_unlock(&this->timeCheckerMutex);
        return signalRet;
    }
private:
    pthread_condattr_t cattr;
    pthread_cond_t condTimeChecker;
    pthread_mutex_t timeCheckerMutex;
    bool signalFlag { false };  // signal will NOT lost when call PthreadSignal before PthreadTimedwaitSecs
};

}  // namespace ttp
}  // namespace ock
#endif // OCK_TTP_PTHREAD_TIMEDWAIT_H
