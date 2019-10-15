package viewmodels_test

import (
	"github.com/alphagov/paas-cf/tools/marketplace/viewmodels"
	cfClient "github.com/cloudfoundry-community/go-cfclient"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Services", func() {
	Describe("NewServicesViewModel", func() {
		It("correctly pairs up the services and their service plans", func() {
			services := []cfClient.Service{
				{
					Guid:        "svc-1",
					Label:       "svc-1",
					Description: "Service 1",
				},
				{
					Guid:        "svc-2",
					Label:       "svc-2",
					Description: "Service 2",
				},
			}

			plans := []cfClient.ServicePlan{
				{
					Name:        "Plan 1",
					Guid:        "plan-1",
					ServiceGuid: "svc-1",
				}, {
					Name:        "Plan 2",
					Guid:        "plan-2",
					ServiceGuid: "svc-1",
				}, {
					Name:        "Plan 3",
					Guid:        "plan-3",
					ServiceGuid: "svc-2",
				},
			}

			viewModel := viewmodels.NewServicesViewModel(services, plans)

			Expect(viewModel.Services[0].CFService.Guid).To(Equal("svc-1"))
			Expect(viewModel.Services[1].CFService.Guid).To(Equal("svc-2"))

			Expect(len(viewModel.Services[0].Plans)).To(Equal(2))
			Expect(viewModel.Services[0].Plans[0].Guid).To(Equal("plan-1"))
			Expect(viewModel.Services[0].Plans[1].Guid).To(Equal("plan-2"))
			Expect(viewModel.Services[1].Plans[0].Guid).To(Equal("plan-3"))
		})

		It("gives an empty slice of plans if no plans could be found", func() {
			services := []cfClient.Service{
				{Guid: "svc-1", Label: "svc-1"},
			}

			plans := []cfClient.ServicePlan{}

			viewModel := viewmodels.NewServicesViewModel(services, plans)
			Expect(viewModel.Services[0].Plans).To(Equal([]cfClient.ServicePlan{}))
		})

		It("filters out services and plans from smoke and acceptance tests", func() {
			services := []cfClient.Service{
				{
					Guid:        "svc-1",
					Label:       "svc-1",
					Description: "Service 1",
				},
				{
					Guid:        "svc-2",
					Label:       "CATS-SVC",
					Description: "Custom acceptance test service",
				},
			}

			plans := []cfClient.ServicePlan{
				{
					Name:        "Plan 1",
					Guid:        "plan-1",
					ServiceGuid: "svc-1",
				}, {
					Name:        "Plan 2",
					Guid:        "plan-2",
					ServiceGuid: "svc-1",
				}, {
					Name:        "Plan 3",
					Guid:        "plan-3",
					ServiceGuid: "svc-2",
				},
			}

			viewModel := viewmodels.NewServicesViewModel(services, plans)

			Expect(len(viewModel.Services)).To(Equal(1))
			Expect(viewModel.Services[0].CFService.Guid).To(Equal("svc-1"))

			Expect(len(viewModel.Services[0].Plans)).To(Equal(2))
			Expect(viewModel.Services[0].Plans[0].Guid).To(Equal("plan-1"))
			Expect(viewModel.Services[0].Plans[1].Guid).To(Equal("plan-2"))
		})
	})
})
