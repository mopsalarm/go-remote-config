package restapi

import (
	"encoding/json"
	"github.com/goji/httpauth"
	"github.com/julienschmidt/httprouter"
	"github.com/mopsalarm/go-remote-config/config"
	"github.com/pkg/errors"
	"net/http"
	"strconv"
	"sync"
)

type Service interface {
	// some service definition
	StartupCount() (int, error)
}

type RestApi struct {
	lock  sync.RWMutex
	rules []config.Rule
}

func Setup(router *httprouter.Router, password string, rules []config.Rule) {
	api := &RestApi{
		rules: rules,
	}

	router.GET("/version/:version/hash/:hash/config.json", metricWrap("config", api.handleGetConfig))
	router.GET("/rules", metricWrap("rules.get", api.handleGetRules))

	// secure a route to write a new config.
	router.Handler("PUT", "/rules",
		httpauth.SimpleBasicAuth("admin", password)(http.HandlerFunc(api.handlePutRules)))
}

func (api *RestApi) handleGetConfig(w http.ResponseWriter, request *http.Request, params httprouter.Params) {
	// parse version from request
	version, err := strconv.Atoi(params.ByName("version"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, errors.WithMessage(err, "Could not parse version number."))
		return
	}

	// get the rules
	api.lock.RLock()
	rules := api.rules
	api.lock.RUnlock()

	// build config from rule.
	conf := config.Apply(rules, config.Context{
		Version:    version,
		DeviceHash: params.ByName("hash"),
		Beta:       request.URL.Query().Get("beta") == "true",
	})

	WriteJSON(w, http.StatusOK, conf)
}

func (api *RestApi) handleGetRules(w http.ResponseWriter, request *http.Request, params httprouter.Params) {
	// get the rules
	api.lock.RLock()
	rules := api.rules
	api.lock.RUnlock()

	WriteJSON(w, http.StatusOK, rules)
}

func (api *RestApi) handlePutRules(w http.ResponseWriter, request *http.Request) {
	var rules []config.Rule
	if err := json.NewDecoder(request.Body).Decode(&rules); err != nil {
		WriteError(w, http.StatusBadRequest, errors.WithMessage(err, "Could not decode body."))
		return
	}

	// update the rules.
	api.lock.Lock()
	api.rules = rules
	api.lock.Unlock()

	w.WriteHeader(http.StatusAccepted)
}
