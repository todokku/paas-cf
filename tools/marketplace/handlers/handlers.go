package handlers

import (
	"code.cloudfoundry.org/lager"
	"fmt"
	"github.com/alphagov/paas-cf/tools/marketplace/context"
	. "github.com/alphagov/paas-cf/tools/marketplace/middleware"
	"github.com/alphagov/paas-cf/tools/marketplace/viewmodels"
	cfClient "github.com/cloudfoundry-community/go-cfclient"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/url"
)

type HttpHandlers struct {
	logger   lager.Logger
	cfClient cfClient.CloudFoundryClient
}

type HandlerError struct {
	StatusCode int
	Err        error
}

func (h HandlerError) Error() string {
	return fmt.Sprintf("status code: %d message: %s", h.StatusCode, h.Err.Error())
}

type ViewData struct {
	Services map[string]cfClient.Service
	Plans    map[string][]cfClient.ServicePlan
}

func NewHttpHandlers(cfClient cfClient.CloudFoundryClient, logger lager.Logger) HttpHandlers {
	return HttpHandlers{
		cfClient: cfClient,
		logger:   logger.Session("http-handlers"),
	}
}

// Takes an HTTP handler and marshals the context
// and error types.
func GinWrap(fn func(context.ContextInterface) error) func(c *gin.Context) {
	return func(c *gin.Context) {
		err := fn(&context.Context{c})

		if err != nil {
			switch err.(type) {
			case *HandlerError:
				handleErr := err.(*HandlerError)
				c.Status(handleErr.StatusCode)

				if handleErr.Err != nil {
					c.Error(handleErr.Err)
				}
				return

			default:
				c.Error(err)
			}

		}
	}
}

func (h *HttpHandlers) Index(c context.ContextInterface) error {
	lsess := h.logger.Session("index")

	services, err := h.cfClient.ListServices()
	if err != nil {
		lsess.Error("list-services", err)
		return &HandlerError{StatusCode: http.StatusInternalServerError}
	}
	
	plans, err := h.cfClient.ListServicePlans()
	if err != nil {
		lsess.Error("list-service-plans", err)
		return &HandlerError{StatusCode: http.StatusInternalServerError}
	}

	viewModel := viewmodels.NewServicesViewModel(services, plans)

	c.HTML(http.StatusOK, "index", viewModel)
	return nil
}

func (h *HttpHandlers) GetService(c context.ContextInterface) error {
	serviceName := c.Params().ByName("service")
	if serviceName == "" {
		return &HandlerError{
			StatusCode: http.StatusNotFound,
			Err:        ErrNotFound("service parameter missing"),
		}
	}

	services, err := h.cfClient.ListServicesByQuery(url.Values{
		"q": []string{fmt.Sprintf("label:%s", serviceName)},
	})

	if err != nil {
		return &HandlerError{500, err}
	}

	if len(services) < 1 {
		return &HandlerError{
			http.StatusNotFound,
			ErrNotFound(fmt.Sprintf("service '%s' cannot be found", serviceName)),
		}
	}

	if len(services) > 1 {
		return &HandlerError{
			http.StatusInternalServerError,
			fmt.Errorf("service '%s' labels more than one service, and is ambiguous", serviceName),
		}
	}
	service := services[0]

	servicePlans, err := h.cfClient.ListServicePlansByQuery(url.Values{
		"q": []string{fmt.Sprintf("service_guid:%s", service.Guid)},
	})

	if err != nil {
		return &HandlerError{http.StatusInternalServerError, err}
	}

	viewModel := viewmodels.Service{
		CFService: service,
		Plans:     servicePlans,
	}

	c.HTML(http.StatusOK, "service", viewModel)

	return nil
}

func (h *HttpHandlers) NoRoute(c context.ContextInterface) error {
	// Push not found errors through the same 404
	// path as other not founds for consistency.
	//
	// Use `c.Request.URL.Path` because
	// `c.FullPath()` only contains a path when
	// a route has been matched.
	return &HandlerError{
		StatusCode: http.StatusNotFound,
		Err:        ErrNotFound(c.Request().URL.Path),
	}
}

func (h *HttpHandlers) Example(c context.ContextInterface) error {
	c.HTML(http.StatusOK, "example", nil)
	return nil
}
