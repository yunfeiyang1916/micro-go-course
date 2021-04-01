package transport

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/transport"
	"github.com/gorilla/mux"
	"github.com/yunfeiyang1916/micro-go-course/hystrix/comment/endpoint"
)

var (
	ErrorBadRequest = errors.New("invalid request parameter")
)

// MakeHttpHandler make http handler use mux
func MakeHttpHandler(ctx context.Context, endpoints *endpoint.CommentsEndpoints) http.Handler {
	r := mux.NewRouter()

	kitLog := log.NewLogfmtLogger(os.Stderr)

	kitLog = log.With(kitLog, "ts", log.DefaultTimestampUTC)
	kitLog = log.With(kitLog, "caller", log.DefaultCaller)

	options := []kithttp.ServerOption{
		kithttp.ServerErrorHandler(transport.NewLogErrorHandler(kitLog)),
		kithttp.ServerErrorEncoder(encodeError),
	}

	r.Methods("GET").Path("/comments/detail").Handler(kithttp.NewServer(
		endpoints.CommentsListEndpoint,
		decodeCommentsListRequest,
		encodeJSONResponse,
		options...,
	))
	return r
}
func decodeCommentsListRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	id := r.URL.Query().Get("id")

	if id == "" {
		return nil, ErrorBadRequest
	}
	return endpoint.CommentsListRequest{
		Id: id,
	}, nil
}

func encodeJSONResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}

func encodeError(_ context.Context, err error, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	switch err {
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}
