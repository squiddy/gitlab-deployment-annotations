package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Range struct {
	From time.Time `json:"from"`
	To   time.Time `json:"to"`
}

type Annotation struct {
	Name       string `json:"name"`
	Datasource string `json:"datasource"`
	IconColor  string `json:"iconColor"`
	Enable     bool   `json:"enable"`
	Query      string `json:"query"`
}

type AnnotationsRequest struct {
	Annotation Annotation `json:"annotation"`
	Range      Range      `json:"range"`
}

type AnnotationsResponse struct {
	Annotation Annotation `json:"annotation"`
	Time       int64      `json:"time"`
	Title      string     `json:"title"`
}

func annotations(res http.ResponseWriter, req *http.Request) {
	var parsed AnnotationsRequest
	if err := json.NewDecoder(req.Body).Decode(&parsed); err != nil {
		http.Error(res, fmt.Sprintf("decode error %v", err), http.StatusBadRequest)
		return
	}

	dataSource := req.Context().Value(ctxDataSourceKey).(*DataSource)
	result, err := dataSource.AnnotationHandler(parsed)
	if err != nil {
		http.Error(res, fmt.Sprintf("%v", err), http.StatusBadRequest)
		return
	}

	json.NewEncoder(res).Encode(&result)
}

func index(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "ok")
}

type AnnotationHandler func(req AnnotationsRequest) ([]AnnotationsResponse, error)

type DataSource struct {
	AnnotationHandler AnnotationHandler
}

// HandleAnnotation assigns a function that will handle incoming annotation
// requests.
func (ds *DataSource) HandleAnnotation(fn AnnotationHandler) {
	ds.AnnotationHandler = fn
}

type key int

const ctxDataSourceKey = key(1)

func withDataSource(next http.Handler, dataSource DataSource) http.Handler {
	return http.HandlerFunc(
		func(res http.ResponseWriter, req *http.Request) {
			ctx := context.WithValue(req.Context(), ctxDataSourceKey, &dataSource)
			req = req.WithContext(ctx)
			next.ServeHTTP(res, req)
		},
	)
}

// ListenAndServe listens on the given address for incoming requests from
// grafana.
func (ds DataSource) ListenAndServe(addr string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", index)
	mux.HandleFunc("/annotations", annotations)
	return http.ListenAndServe(addr, withDataSource(mux, ds))
}
