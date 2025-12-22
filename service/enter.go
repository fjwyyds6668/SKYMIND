package service

import "skymind/service/smart_query"

var ServiceGroupApp = new(ServiceGroup)

type ServiceGroup struct {
	SmartQueryServiceGroup smart_query.ServiceGroup
}
