package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/go-openapi/loads"
	flags "github.com/jessevdk/go-flags"

	"github.com/pjotrscholtze/go-buildserver/cmd/go-buildserver/config"
	"github.com/pjotrscholtze/go-buildserver/cmd/go-buildserver/controller"
	"github.com/pjotrscholtze/go-buildserver/cmd/go-buildserver/repo"
	"github.com/pjotrscholtze/go-buildserver/cmd/go-buildserver/websocketmanager"
	"github.com/pjotrscholtze/go-buildserver/restapi"
	"github.com/pjotrscholtze/go-buildserver/restapi/operations"
	"github.com/robfig/cron/v3"
)

type mock struct {
	next http.Handler
	mux  *http.ServeMux
}

func (m *mock) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix((*r).RequestURI, "/api/") ||
		strings.HasPrefix((*r).RequestURI, "/swagger.json") {
		m.next.ServeHTTP(w, r)
		return
	}
	m.mux.ServeHTTP(w, r)
}

func main() {
	// if len(os.Args) != 2 {
	// 	println("Usage: app <config-path.yaml>")
	// 	return
	// }
	// path := os.Args[1]
	path := "../../example/config.yaml"
	log.Println("Starting buildserver")

	log.Printf("Loading config: %s\n", path)
	c := config.LoadConfig(path)
	log.Println("")
	cr := cron.New(cron.WithSeconds())
	cr.Start()
	wm := websocketmanager.NewWebsocketManager()

	buildRepo := repo.NewPipelineRepo(&c, &wm)
	bq := repo.NewBuildQueue(buildRepo, cr, &wm)
	go bq.Run()

	swaggerSpec, err := loads.Embedded(restapi.SwaggerJSON, restapi.FlatSwaggerJSON)
	if err != nil {
		log.Fatalln(err)
	}

	api := operations.NewGoBuildserverAPI(swaggerSpec)
	controller.ConnectControllers(api, buildRepo, bq)
	server := restapi.NewServer(api)
	defer server.Shutdown()

	parser := flags.NewParser(server, flags.Default)
	parser.ShortDescription = "Go Buildserver"
	parser.LongDescription = swaggerSpec.Spec().Info.Description
	server.ConfigureFlags()
	for _, optsGroup := range api.CommandLineOptionsGroups {
		_, err := parser.AddGroup(optsGroup.ShortDescription, optsGroup.LongDescription, optsGroup.Options)
		if err != nil {
			log.Fatalln(err)
		}
	}

	if _, err := parser.Parse(); err != nil {
		code := 1
		if fe, ok := err.(*flags.Error); ok {
			if fe.Type == flags.ErrHelp {
				code = 0
			}
		}
		os.Exit(code)
	}

	server.ConfigureAPI()
	server.Port = c.HTTPPort
	server.Host = c.HTTPHost

	t := &mock{
		next: api.Serve(nil),
		mux:  controller.RegisterUIController(buildRepo, bq, wm),
	}
	server.SetHandler(t)

	if err := server.Serve(); err != nil {
		log.Fatalln(err)
	}
}
