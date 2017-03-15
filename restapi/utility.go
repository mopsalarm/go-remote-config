package restapi

import (
	"encoding/json"
	"net/http"
	"reflect"
	"strconv"

	"github.com/julienschmidt/httprouter"
	"github.com/rcrowley/go-metrics"
)

type ErrorResponse struct {
	Status int    `json:"status"`
	Error  string `json:"error"`
}

func WriteJSON(w http.ResponseWriter, status int, value interface{}) error {
	body, err := json.Marshal(value)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err)
		return err
	}

	return WriteBody(w, status, "application/json", body)
}

func WriteError(writer http.ResponseWriter, status int, err error) error {
	return WriteJSON(writer, status, ErrorResponse{
		Status: status,
		Error:  err.Error(),
	})
}

func WriteBody(writer http.ResponseWriter, status int, contentType string, content []byte) error {
	writer.Header().Set("Content-Length", strconv.Itoa(len(content)))
	writer.Header().Set("Content-Type", contentType)

	// write header writes the status code and all previously set headers.
	writer.WriteHeader(status)

	_, err := writer.Write(content)
	return err
}

// for a value that is wrapped into an interface, this method
// tries to check, if the value is a slice and if that one is nil.
func isSliceAndNil(value interface{}) bool {
	return value != nil && reflect.TypeOf(value).Kind() == reflect.Slice && reflect.ValueOf(value).IsNil()
}

func metricWrap(metricsTag string, handler httprouter.Handle) httprouter.Handle {
	timer := metrics.GetOrRegisterTimer("rest."+metricsTag, nil)
	return func(w http.ResponseWriter, req *http.Request, params httprouter.Params) {
		timer.Time(func() {
			handler(w, req, params)
		})
	}
}
