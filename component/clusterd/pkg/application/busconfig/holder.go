package busconfig

import (
	"context"
	"sync"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/interface/grpc/config"
)

type ConfigHolder struct {
	jobId          string
	rankTableChan  chan *config.RankTableStream
	holderContext  context.Context
	ctxCancelFunc  context.CancelFunc
	serviceContext context.Context
	lock           sync.RWMutex
}

func NewConfigHolder(jobId string, serviceCtx context.Context) *ConfigHolder {
	holder := &ConfigHolder{
		jobId:          jobId,
		rankTableChan:  make(chan *config.RankTableStream, 1),
		serviceContext: serviceCtx,
		lock:           sync.RWMutex{},
	}
	holder.holderContext, holder.ctxCancelFunc = context.WithCancel(holder.serviceContext)
	return holder
}

func (c *ConfigHolder) listenRankTableChan(stream config.Config_SubscribeRankTableServer) {
	c.reset()
	hwlog.RunLog.Infof("start listen a new send channel, jobId=%s", c.jobId)
	for {
		if c.rankTableChan == nil {
			return
		}
		select {
		case <-c.serviceContext.Done():
			hwlog.RunLog.Infof("context done, jobId=%s break listen sendChan", c.jobId)
			return
		case signal, ok := <-c.rankTableChan:
			if ok {
				err := stream.Send(signal)
				hwlog.RunLog.Infof("send global ranktable failed, error: %v", err)
			} else {
				hwlog.RunLog.Infof("sendChan closed, jobId=%s break listen sendChan", c.jobId)
				return
			}
		}
	}
}

func (c *ConfigHolder) reset() {
	hwlog.RunLog.Infof("jobId=%s enter reset function", c.jobId)
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.ctxCancelFunc != nil {
		c.ctxCancelFunc()
	}
	close(c.rankTableChan)
	c.rankTableChan = make(chan *config.RankTableStream, 1)
	c.holderContext, c.ctxCancelFunc = context.WithCancel(c.serviceContext)
}

func (c *ConfigHolder) notify(data *config.RankTableStream) {
	if c.rankTableChan == nil {
		return
	}
	c.rankTableChan <- data
}
