package viewmodels

import (
	cfClient "github.com/cloudfoundry-community/go-cfclient"
	"strings"
)

type ServicesViewModel struct {
	Services []Service
}

type Service struct {
	CFService cfClient.Service
	Plans []cfClient.ServicePlan
}

func NewServicesViewModel(cfServices []cfClient.Service, cfServicePlans []cfClient.ServicePlan, ) *ServicesViewModel {
	mapSvcGuid := make(map[string]cfClient.Service, 0)
	mapSvcPlanGuid := make(map[string][]cfClient.ServicePlan, 0)

	for _, svc := range cfServices {
		if isAcceptanceTestBrokerService(svc) {
			continue
		}

		mapSvcGuid[svc.Guid] = svc
		mapSvcPlanGuid[svc.Guid] = []cfClient.ServicePlan{}
	}

	for _, plan := range cfServicePlans {
		// Only add the plans for services which haven't been excluded
		if _, ok := mapSvcPlanGuid[plan.ServiceGuid]; ok {
			mapSvcPlanGuid[plan.ServiceGuid] = append(mapSvcPlanGuid[plan.ServiceGuid], plan)
		}
	}

	servicesVMs := []Service{}

	for _, svc := range mapSvcGuid {
		plans := []cfClient.ServicePlan{}

		if p, ok := mapSvcPlanGuid[svc.Guid]; ok {
			plans = p
		}

		servicesVMs = append(servicesVMs, Service{
			CFService: svc,
			Plans:     plans,
		})
	}

	return &ServicesViewModel{Services: servicesVMs}
}

func isAcceptanceTestBrokerService(svc cfClient.Service) bool {
	return strings.Contains(svc.Label, "CATS-")
}
