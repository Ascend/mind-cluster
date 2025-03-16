package busconfig

import (
	"context"
	"fmt"
	"sync"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/domain/common"
	"clusterd/pkg/interface/grpc/config"
)

var server *BusinessConfigServer

type BusinessConfigServer struct {
	serviceCtx   context.Context
	configHolder map[string]*ConfigHolder
	lock         sync.RWMutex
	config.UnimplementedConfigServer
}

func NewBusinessConfigServer(serviceCtx context.Context) *BusinessConfigServer {
	server = &BusinessConfigServer{
		serviceCtx:   serviceCtx,
		configHolder: make(map[string]*ConfigHolder),
		lock:         sync.RWMutex{},
	}
	return server
}

// Register is task register service
func (c *BusinessConfigServer) Register(ctx context.Context, req *config.ClientInfo) (*config.Status, error) {
	hwlog.RunLog.Infof("config service receive Register request, jobId=%s, role=%s", req.JobId, req.Role)
	c.lock.RLock()
	defer c.lock.RUnlock()
	_, ok := c.configHolder[req.JobId]
	if ok {
		return &config.Status{Code: int32(common.OK), Info: "register success"}, nil
	}
	holder := NewConfigHolder(req.JobId, c.serviceCtx)
	c.configHolder[req.JobId] = holder
	return &config.Status{Code: int32(common.OK), Info: "register success"}, nil
}

// SubscribeRankTable subscribe rank table from ClusterD
func (c *BusinessConfigServer) SubscribeRankTable(request *config.ClientInfo,
	stream config.Config_SubscribeRankTableServer) error {
	requestInfo := fmt.Sprintf("taskId=%s, rule=%s", request.JobId, request.Role)
	hwlog.RunLog.Infof("receive Subscribe ranktable request, %s", requestInfo)
	holder, ok := c.configHolder[request.JobId]
	if !ok {
		return fmt.Errorf("jobId=%s not registed", request.JobId)
	}
	table := GetData(request.JobId)
	if len(table) > 0 {
		err := stream.Send(&config.RankTableStream{JobId: request.JobId, RankTable: table})
		if err != nil {
			hwlog.RunLog.Infof("send global ranktable failed, error: %v", err)
		}
	}

	// reset controller
	holder.listenRankTableChan(stream)
	return nil
}

func (c *BusinessConfigServer) ranktableChanged(jobId, rankTable string) {
	c.lock.RLock()
	holder, ok := c.configHolder[jobId]
	if !ok {
		return
	}
	c.lock.RUnlock()
	data := &config.RankTableStream{
		JobId:     jobId,
		RankTable: rankTable,
	}
	holder.notify(data)
}

func dataChanged(jobId, rankTable string) {
	//server.ranktableChanged(jobId, rankTable)
}
