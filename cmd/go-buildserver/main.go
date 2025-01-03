package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/go-openapi/loads"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/gorilla/mux"
	flags "github.com/jessevdk/go-flags"
	"github.com/jmoiron/sqlx"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file" //
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
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
	mux  *mux.Router
}

func (m *mock) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix((*r).RequestURI, "/api/") ||
		strings.HasPrefix((*r).RequestURI, "/swagger.json") {
		m.next.ServeHTTP(w, r)
		return
	}
	m.mux.ServeHTTP(w, r)
}

func contains(slice []string, str string) bool {
	for _, v := range slice {
		if v == str {
			return true
		}
	}
	return false
}
func getDatabaseDriver(db *sql.DB, driverName string) (database.Driver, error) {
	switch driverName {
	case "postgres":
		return postgres.WithInstance(db, &postgres.Config{})
	case "mysql":
		return mysql.WithInstance(db, &mysql.Config{})
	case "sqlite3":
		return sqlite3.WithInstance(db, &sqlite3.Config{})
	// Add more drivers here as needed.
	default:
		return nil, fmt.Errorf("unsupported driver: %s", driverName)
	}
}
func getDBConnection(c config.Config) (*sqlx.DB, error) {
	db, err := sqlx.Open(c.SQLDriver, c.SQLConnectionString)

	if err != nil {
		log.Println("Failed to create SQL connection!")
		log.Panic(err)
		return nil, err
	}
	db.SetConnMaxIdleTime(-1)
	db.SetConnMaxLifetime(-1)
	db.SetMaxOpenConns(1)
	if err := db.Ping(); err != nil {
		log.Printf("Failed to ping the database: %v", err)
		return nil, err
	}

	return db, nil
}

func main() {
	if len(os.Args) != 2 {
		println("Usage: app <config-path.yaml>")
		return
	}
	path := os.Args[1]
	// path := "../../example/config_mysql.yaml"
	// path := "../../example/config.yaml"
	log.Println("Starting buildserver")

	log.Printf("Loading config: %s\n", path)
	c := config.LoadConfig(path)
	log.Println("")
	cr := cron.New(cron.WithSeconds())
	cr.Start()
	wm := websocketmanager.NewWebsocketManager()
	log.Println("Known SQLDrivers: " + strings.Join(sql.Drivers(), ", "))
	log.Println("Selected SQLDriver: " + c.SQLDriver)
	knownSQLDriver := contains(sql.Drivers(), c.SQLDriver)
	log.Println("Is known SQLDriver: " + strconv.FormatBool(knownSQLDriver))
	if !knownSQLDriver {
		log.Println("An unkown SQL Driver was provided, please provide a known driver")
		return
	}
	db, err := getDBConnection(c)
	defer db.Close()

	driver, err := getDatabaseDriver(db.DB, c.SQLDriver)
	m, err := migrate.NewWithDatabaseInstance(
		"file://"+c.DBMigrations,
		c.SQLDriver,
		driver,
	)
	if err != nil {
		log.Fatalf("Failed to initialize migrate: %v", err)
	}

	// Apply migrations
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("Failed to apply migrations: %v", err)
	}

	fmt.Println("Migrations applied successfully!")
	if !(c.SQLDriver == "sqlite3" && strings.Contains(c.SQLConnectionString, "memory")) {
		fmt.Println("Reconnecting to the database")
		db.Close()
		db, err = getDBConnection(c)
		defer db.Close()
		if err != nil {
			panic(err)
		}
	}

	fmt.Println("Starting server")

	dbRepo := repo.NewDatabaseRepo(db)
	buildResultRepo := repo.NewBuildResultRepo(dbRepo, &wm)
	buildRepo := repo.NewPipelineRepo(&c, &wm, buildResultRepo, dbRepo)
	bq := repo.NewJobQueue(buildRepo, cr, &wm, dbRepo)
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
		mux:  controller.RegisterUIController(buildRepo, bq, wm, c),
	}
	server.SetHandler(t)

	if err := server.Serve(); err != nil {
		log.Fatalln(err)
	}
}
