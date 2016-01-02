package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"

	"code.google.com/p/go-uuid/uuid"

	"github.com/codegangsta/negroni"
	"github.com/goincremental/negroni-sessions"
	"github.com/goincremental/negroni-sessions/cookiestore"
	"github.com/gorilla/mux"
	"github.com/unrolled/render"
)

// CLI flags
var (
	portFlag      string
	queueHostFlag string
	setupFlag     bool
	dbUserFlag    string
	dbPassFlag    string
	dbHostFlag    string
	dbNameFlag    string
	dbSetupFlag   bool
)

var jobRegistryChan = make(chan Job)
var taskRegistryChan = make(chan Task)
var signalsChan = make(chan os.Signal, 1)

func init() {
	flag.StringVar(&queueHostFlag, "nsq-host", "", "NSQ server to connect to")
	flag.StringVar(&portFlag, "port", ":9898", "port to run the server")
	flag.StringVar(&dbUserFlag, "db-user", "aion", "database user")
	flag.StringVar(&dbPassFlag, "db-pass", "aion", "database pass")
	flag.StringVar(&dbHostFlag, "db-host", "aion", "database host")
	flag.StringVar(&dbNameFlag, "db-name", "aion", "database name")
	flag.BoolVar(&dbSetupFlag, "db-setup", false, "intial DB configuration")
}

func main() {
	flag.Parse()

	signal.Notify(signalsChan, os.Interrupt)

	go func() {
		for sig := range signalsChan {
			log.Printf("Exiting... %v\n", sig)
			signalsChan = nil
			os.Exit(1)
		}
	}()

	if setupFlag {
		db, err := NewDatabase(dbUserFlag, dbPassFlag, dbHostFlag, dbNameFlag)
		if err != nil {
			log.Fatalln(err)
		}
		db.Setup()
		os.Exit(0)
	}

	// setup the renderer for returning our JSON
	ren := render.New(render.Options{})

	store := cookiestore.New([]byte(uuid.NewUUID().String()))

	// initialize the web framework
	n := negroni.New(
		negroni.NewRecovery(),
		negroni.NewLogger(),
		negroni.NewStatic(http.Dir("public")),
	)

	n.Use(sessions.Sessions("session", store))

	// create a router to handle the requests coming in to our endpoints
	router := mux.NewRouter()

	// Frontend Entry Point
	router.HandleFunc(frontEnd, FrontendHandler()).Methods("GET")

	// Jobs Route
	router.HandleFunc(JobsPath, JobsRouteHandler(ren)).Methods("GET")

	// New Jobs Route
	router.HandleFunc(JobsPath, NewJobsRouteHandler(ren)).Methods("POST")

	// Tasks Route
	router.HandleFunc(TasksPath, TasksRouteHandler(ren)).Methods("GET")

	// New Tasks Route
	router.HandleFunc(TasksPath, NewTasksRouteHandler(ren)).Methods("POST")

	n.UseHandler(router)
	n.Run(portFlag)
}
