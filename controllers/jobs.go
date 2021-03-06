package controllers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	
	"github.com/gorilla/mux"
	"github.com/unrolled/render"

	"github.com/briandowns/aion/config"
	"github.com/briandowns/aion/database"
	"github.com/briandowns/aion/dispatcher"
)

// JobsRouteHandler provides the handler for jobs data
func JobsRouteHandler(ren *render.Render, conf *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		db, err := database.NewDatabase(conf)
		if err != nil {
			log.Println(err)
			return 
		}
		defer db.Conn.Close()
		ren.JSON(w, http.StatusOK, map[string]interface{}{"jobs": db.GetJobs()})
	}
}

// NewJobRouteHandler creates a new job with the POST'd data
func NewJobRouteHandler(ren *render.Render, dispatcher *dispatcher.Dispatcher) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var nj database.Job

		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&nj)
		if err != nil {
			ren.JSON(w, 400, map[string]error{"error": err})
			return
		}
		defer r.Body.Close()
		switch {
		case nj.Name == "":
			ren.JSON(w, 400, map[string]error{"error": ErrMissingNameField})
			return
		}

		dispatcher.SenderChan <- &nj
		ren.JSON(w, http.StatusOK, map[string]database.Job{"job": nj})
	}
}

// JobByIDRouteHandler provides the handler for jobs data for the given ID
func JobByIDRouteHandler(ren *render.Render, conf *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		jid := vars["id"]

		jobID, err := strconv.Atoi(jid)
		if err != nil {
			log.Println(err)
			ren.JSON(w, 400, map[string]error{"error": err})
			return
		}

		db, err := database.NewDatabase(conf)
		if err != nil {
			log.Println(err)
			ren.JSON(w, http.StatusOK, map[string]error{"error": err})
			return
		}
		defer db.Conn.Close()

		if t := db.GetJobByID(jobID); len(t) > 0 {
			ren.JSON(w, http.StatusOK, map[string]interface{}{"task": t})
		} else {
			ren.JSON(w, http.StatusOK, map[string]interface{}{"task": ErrNoJobsFound.Error()})
		}
	}
}

// JobDeleteByIDRouteHandler deletes the job data for the given ID
func JobDeleteByIDRouteHandler(ren *render.Render, conf *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		jid := vars["id"]
		jobID, err := strconv.Atoi(jid)
		if err != nil {
			log.Println(err)
			return
		}
		db, err := database.NewDatabase(conf)
		if err != nil {
			log.Println(err)
			return
		}
		defer db.Conn.Close()

		db.DeleteJob(jobID)

		ren.JSON(w, http.StatusOK, map[string]interface{}{"task": jobID})
	}
}