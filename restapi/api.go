package restapi

import (
	"github.com/julienschmidt/httprouter"
	"github.com/mopsalarm/go-remote-config/config"
	"github.com/pkg/errors"
	"net/http"
	"strconv"
)

type Service interface {
	// some service definition
	StartupCount() (int, error)
}

type RestApi struct {
	rules []config.Rule
}

func Setup(router *httprouter.Router, rules []config.Rule) {
	api := &RestApi{
		rules: rules,
	}

	router.GET("/version/:version/hash/:hash/config.json", metricWrap("rest.config", api.handleGetConfig))
}

func (api *RestApi) handleGetConfig(w http.ResponseWriter, request *http.Request, params httprouter.Params) {
	// parse version from request
	version, err := strconv.Atoi(params.ByName("version"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, errors.WithMessage(err, "Could not parse version number."))
		return
	}

	// build config from rule.
	conf := config.Apply(api.rules, config.Context{
		Version:    version,
		DeviceHash: params.ByName("hash"),
	})

	WriteJSON(w, http.StatusOK, conf)
}
