package context

import "github.com/gin-gonic/gin"

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -o fakes/fake_cf_client.go ../vendor/github.com/cloudfoundry-community/go-cfclient/client_interface.go CloudFoundryClient

// StubbedContext decorates the handlers.Context type
// in order to stub out the HTML method.
//
// This is necessary in these tests because we
// want to test what parameters were given to
// the method, and it would otherwise be a
// black box.
//
// We don't care about what HTML would normally do.
type StubbedContext struct{
	*Context
	HTMLCallback func(int, string, interface{})
	Parameters gin.Params
}

func (c StubbedContext) HTML(code int, name string, obj interface{}) {
	if c.HTMLCallback != nil {
		c.HTMLCallback(code, name, obj)
	}
}

func (c StubbedContext) Params() gin.Params {
	return c.Parameters
}
