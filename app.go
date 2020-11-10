package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

type App struct {
	Router *mux.Router
	config *Env
}

// init of app
func (a *App) Initialize(e *Env) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	a.Router = mux.NewRouter()
	a.config = e
	a.initializeRoutes()
}

// init a route
func (a *App) initializeRoutes() {
	a.Router.HandleFunc("/api/shorten", a.createShortlink).Methods("POST")
	log.Println("POST /api/shorten")
	a.Router.HandleFunc("/api/info", a.getShortlinkInfo).Methods("GET")
	log.Println("GET /api/info")
	a.Router.HandleFunc("/{shortlink:[a-zA-Z0-9]{1,11}}", a.redirect).Methods("GET")
	log.Println("GET /{shortlink:[a-zA-Z0-9]{1,11}}")
}

func (a *App) Run(addr string) {
	log.Printf("RUN at [%s]", addr)
	log.Fatal(http.ListenAndServe(addr, a.Router))
}

type shortenReq struct {
	URL                 string `json:"url" validate:"required"`
	ExpirationInMinutes int64  `json:"expiration_in_minutes" validate:"min=0"`
}

type shortlinkResp struct {
	Shortlink string `json:"shortlink"`
}

func (a *App) createShortlink(w http.ResponseWriter, r *http.Request) {
	var req shortenReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, StatusError{http.StatusBadRequest,
			fmt.Errorf("parse parameters failed %v", r.Body)})
		return
	}
	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		respondWithError(w, StatusError{http.StatusBadRequest,
			fmt.Errorf("validate parameters failed %v", req)})
		return
	}
	defer r.Body.Close()

	fmt.Printf("%#v\n", req)

	shortenURL, err := a.config.s.Shorten(req.URL, req.ExpirationInMinutes)
	if err != nil {
		respondWithError(w, err)
	} else {
		respondWithJSON(w, http.StatusOK, shortlinkResp{Shortlink: shortenURL})
	}
}

func (a *App) getShortlinkInfo(w http.ResponseWriter, r *http.Request) {
	values := r.URL.Query()
	shortenURL := values.Get("shortlink")

	shortlinkInfo, err := a.config.s.ShortlinkInfo(shortenURL)
	if err != nil {
		respondWithError(w, err)
	} else {
		respondWithJSON(w, http.StatusOK, shortlinkInfo)
	}
}

func (a *App) redirect(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	fmt.Printf("shortlink=%s\n", vars["shortlink"])

	shortenURL := vars["shortlink"]
	URL, err := a.config.s.Unshorten(shortenURL)
	if err != nil {
		respondWithError(w, err)
	} else {
		// 零时重定向
		http.Redirect(w, r, URL, http.StatusTemporaryRedirect)
	}
}

func respondWithError(w http.ResponseWriter, err error) {
	switch e := err.(type) {
	case Error:
		log.Printf("HTTP %d - %s", e.Status(), e)
		respondWithJSON(w, e.Status(), e.Error())
	default:
		respondWithJSON(w, http.StatusInternalServerError,
			http.StatusText(http.StatusInternalServerError))
	}
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	resp, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(resp)
}
