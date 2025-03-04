package eventapi

import (
	"ascend-faultdiag-online/pkg/context/contextdata"
	"ascend-faultdiag-online/pkg/context/diagcontext"
	"ascend-faultdiag-online/pkg/model/diagmodel"
	"ascend-faultdiag-online/pkg/service/request"
	"ascend-faultdiag-online/pkg/service/servicecore"
)

const apiDiagEvent = "diag"

// GetDiagEventApi 获取添加指标的api
func GetDiagEventApi() *servicecore.Api {
	return servicecore.BuildApi(apiDiagEvent, &diagmodel.DiagModel{}, apiDiagEventFunc, nil)
}

// apiDiagEventFunc 诊断事件
func apiDiagEventFunc(ctxData *contextdata.CtxData, diagCtx *diagcontext.DiagContext, reqCtx *request.Context, model *diagmodel.DiagModel) error {
	diagCtx.StartDiag(ctxData)
	return nil
}
