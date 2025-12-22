package api

import "skymind/api/smart_query"

var ApiGroupApp = new(ApiGroup)

type ApiGroup struct {
	SmartQueryApiGroup smart_query.ApiGroup
}
