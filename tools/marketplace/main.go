package main

import (
	"bytes"
	"code.cloudfoundry.org/lager"
	"fmt"
	"github.com/alphagov/paas-cf/tools/marketplace/handlers"
	"github.com/alphagov/paas-cf/tools/marketplace/middleware"
	"github.com/cloudfoundry-community/go-cfclient"
	"html/template"
	"io/ioutil"
	"path"
	"path/filepath"
	"strings"

	"github.com/foolin/goview"
	"github.com/foolin/goview/supports/ginview"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
)

func main() {
	logger := lager.NewLogger("paas-cf-marketplace")
	logger.RegisterSink(lager.NewWriterSink(os.Stdout, lager.INFO))

	port := envVar("PORT", logger)
	api := envVar("CF_API", logger)
	clientName := envVar("CF_CLIENT", logger)
	clientSecret := envVar("CF_CLIENT_SECRET", logger)

	cfClient, err := cfclient.NewClient(&cfclient.Config{
		ApiAddress:   api,
		ClientID:     clientName,
		ClientSecret: clientSecret,
	})

	if err != nil {
		logger.Fatal("create-cf-client", err)
	}

	httpHandlers := handlers.NewHttpHandlers(cfClient, logger)

	ginRouter := configureGinRouter(httpHandlers)

	logger.Info("listen-and-serve", lager.Data{"port": port})
	err = ginRouter.Run(":" + port)

	if err != nil {
		logger.Error("listen-and-serve", err)
	}
	logger.Info("exiting")
}

func configureGinRouter(httpHandlers handlers.HttpHandlers) *gin.Engine {
	ginRouter := gin.Default()

	var goViewEngine *ginview.ViewEngine = nil

	goViewConfig := goview.Config{
		Root:      "templates",
		Extension: ".gohtml",
		Master:    "layout",
		Partials:  discoverPartials(),
		Funcs: map[string]interface{}{
			"partial": func(layout string, data interface{}) (template.HTML, error) {
				// goview renders a template using a master file
				// when it receives a file name without the extension.
				// We don't want it to render partials using the
				// master, but it's goview's convention to pass without.
				// To avoid that, we force the extension here.

				layout = strings.TrimSuffix(layout, ".gohtml") + ".gohtml"

				buf := new(bytes.Buffer)
				err := goViewEngine.RenderWriter(buf, layout, data)
				return template.HTML(buf.String()), err
			},
		},
	}
	goViewEngine = ginview.New(goViewConfig)

	ginRouter.HTMLRender = goViewEngine
	ginRouter.Use(middleware.NotFoundMiddleware())

	ginRouter.StaticFS("/assets/", http.Dir("./assets"))
	ginRouter.StaticFS("/javascript/", http.Dir("./javascript"))
	ginRouter.StaticFS("/stylesheets/", http.Dir("./stylesheets"))

	ginRouter.GET("/services/:service", handlers.GinWrap(httpHandlers.GetService))
	ginRouter.GET("/example", handlers.GinWrap(httpHandlers.Example))
	ginRouter.GET("/", handlers.GinWrap(httpHandlers.Index))

	ginRouter.NoRoute(handlers.GinWrap(httpHandlers.NoRoute))
	return ginRouter
}

func discoverPartials() []string {
	fileInfos, err := ioutil.ReadDir("templates/partials/")
	if err != nil {
		panic(err)
	}

	// goview wants the partial names giving
	// * without a file extension (".gohtml")
	// * relative to the template root ("templates")
	workingDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	root := path.Join(workingDir, "templates")
	var relativeFilePaths []string
	for _, info := range fileInfos {
		fullPath := path.Join(root, "partials", info.Name())
		relPath, err := filepath.Rel(root, fullPath)
		if err != nil {
			panic(err)
		}
		strippedPath := strings.TrimSuffix(relPath, ".gohtml")
		relativeFilePaths = append(relativeFilePaths, strippedPath)
	}

	return relativeFilePaths
}

func envVar(name string, logger lager.Logger) string {
	val, found := os.LookupEnv(name)

	if !found {
		logger.Fatal("lookup-env-var", fmt.Errorf("environment var not set: %s", name))
	}

	return val
}
