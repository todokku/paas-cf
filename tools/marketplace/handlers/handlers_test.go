package handlers_test


import (
	"code.cloudfoundry.org/lager"
	"fmt"
	"github.com/alphagov/paas-cf/tools/marketplace/context"
	"github.com/alphagov/paas-cf/tools/marketplace/context/fakes"
	"github.com/alphagov/paas-cf/tools/marketplace/handlers"
	"github.com/alphagov/paas-cf/tools/marketplace/middleware"
	"github.com/alphagov/paas-cf/tools/marketplace/viewmodels"
	cfClient "github.com/cloudfoundry-community/go-cfclient"
	"github.com/gin-gonic/gin"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"net/http"
	"net/http/httptest"
)


var _ = Describe("Handlers", func() {
	var (
		stubContext  context.StubbedContext
		respWriter   *httptest.ResponseRecorder
		httpHandlers handlers.HttpHandlers
		fakeCF       *fakes.FakeCloudFoundryClient
	)
	BeforeEach(func(){
		logger := lager.NewLogger("test")
		logger.RegisterSink(lager.NewWriterSink(GinkgoWriter, lager.DEBUG))

		respWriter = httptest.NewRecorder()
		c, _ := gin.CreateTestContext(respWriter)
		ctx := context.Context{c}

		stubContext = context.StubbedContext{&ctx, nil, gin.Params{}}
		fakeCF = &fakes.FakeCloudFoundryClient{}

		httpHandlers = handlers.NewHttpHandlers(fakeCF, logger)
	})


	Describe("GetService", func(){
		It("returns a not found error if the service parameter is empty", func(){
			stubContext.Parameters = gin.Params{
				{Key: "service", Value: ""},
			}

			err := httpHandlers.GetService(stubContext)

			herr := err.(*handlers.HandlerError)
			Expect(herr).ToNot(BeNil())
			Expect(herr.Err).To(MatchError(middleware.ErrNotFound("service parameter missing")))
		})

		It("return a not found error when the service cannot be found", func(){
			fakeCF.ListServicesByQueryReturns([]cfClient.Service{}, nil)

			stubContext.Parameters = gin.Params{
				{"service", "foo"},
			}
			err := httpHandlers.GetService(stubContext)

			herr := err.(*handlers.HandlerError)
			Expect(herr).ToNot(BeNil())
			Expect(herr.Err).To(MatchError(
				middleware.ErrNotFound(fmt.Sprintf("service '%s' cannot be found", "foo")),
			))
		})

		It("returns a 500 status code when the name finds > 1 result", func(){
			fakeCF.ListServicesByQueryReturns([]cfClient.Service{
				{Guid: "svc-1", Label: "service"},
				{Guid: "svc-2", Label: "service"},
			}, nil)

			stubContext.Parameters = gin.Params{
				{"service", "service"},
			}

			err := httpHandlers.GetService(stubContext)
			herr := err.(*handlers.HandlerError)
			Expect(herr).ToNot(BeNil())
			Expect(herr.StatusCode).To(Equal(http.StatusInternalServerError))
		})

		It("renders the service and all of its plans when all is well", func(){
			service := cfClient.Service{Guid: "svc-1", Label: "service"}
			fakeCF.ListServicesByQueryReturns([]cfClient.Service{
				service,
			}, nil)

			plans := []cfClient.ServicePlan{
				{Guid: "plan-1", ServiceGuid: "svc-1"},
				{Guid: "plan-2", ServiceGuid: "svc-1"},
			}
			fakeCF.ListServicePlansByQueryReturns(plans, nil)

			stubContext.Parameters = gin.Params{
				{"service", "service-1"},
			}

			stubContext.HTMLCallback = func (code int, name string, data interface{}){
				Expect(data).To(BeAssignableToTypeOf(viewmodels.Service{}))

				vm := data.(viewmodels.Service)
				Expect(vm.CFService).To(Equal(service))
				Expect(vm.Plans).To(Equal(plans))
			}

			err := httpHandlers.GetService(stubContext)
			Expect(err).ToNot(HaveOccurred())

		})
	})
})
