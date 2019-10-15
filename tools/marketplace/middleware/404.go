package middleware

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func NotFoundMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// This middleware always runs after the request is handled.
		// We call gin.Context.Next() to allow the pipeline to carry on.
		c.Next()

		caughtErrors := c.Errors.ByType(gin.ErrorTypeAny)

		// We care about handling errors here
		if len(caughtErrors) == 0 {
			return
		}

		err := caughtErrors[0].Err

		switch err.(type) {
		case NotFoundError:
			c.HTML(http.StatusNotFound, "404", nil)
			return

		default:
			return
		}
	}
}
