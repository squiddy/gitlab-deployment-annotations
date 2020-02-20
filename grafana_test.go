package main

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestIndex(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	index(w, req)

	resp := w.Result()
	assert.Equal(t, 200, resp.StatusCode)
}

func handleAnnotationRequest(payload string, fn AnnotationHandler) *http.Response {
	reqBody := strings.NewReader(payload)
	req := httptest.NewRequest("POST", "/", reqBody)
	w := httptest.NewRecorder()

	ds := &DataSource{}
	ds.HandleAnnotation(fn)
	handler := withDataSource(http.HandlerFunc(annotations), *ds)
	handler.ServeHTTP(w, req)

	return w.Result()
}

func TestAnnotationsInvalidRequest(t *testing.T) {
	resp := handleAnnotationRequest("invalid", nil)
	body, _ := ioutil.ReadAll(resp.Body)
	assert.Equal(t, 400, resp.StatusCode)
	assert.Equal(t, "decode error invalid character 'i' looking for beginning of value\n", string(body))
}

func TestAnnotationsValidRequest(t *testing.T) {
	resp := handleAnnotationRequest(`
	{
		"annotation": {
			"name": "Deploys",
			"Datasource": "My datasource",
			"IconColor": "red",
			"Enable": true,
			"Query": "query"
		},
		"range": {
			"from": "2020-01-02T12:00:00.0Z",
			"to": "2020-01-07T20:00:00.0Z"
		}
	}
	`, func(req AnnotationsRequest) ([]AnnotationsResponse, error) {
		assert.Equal(t, Annotation{Name: "Deploys", Datasource: "My datasource",
			IconColor: "red", Enable: true, Query: "query"}, req.Annotation)
		assert.Equal(t, time.Date(2020, 1, 2, 12, 0, 0, 0, time.UTC), req.Range.From)
		assert.Equal(t, time.Date(2020, 1, 7, 20, 0, 0, 0, time.UTC), req.Range.To)

		res := []AnnotationsResponse{
			AnnotationsResponse{
				Annotation: req.Annotation,
				Time:       123123,
				Title:      "skynet activation",
			},
		}
		return res, nil
	})

	body, _ := ioutil.ReadAll(resp.Body)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, `[{"annotation":{"name":"Deploys","datasource":"My datasource","iconColor":"red","enable":true,"query":"query"},"time":123123,"title":"skynet activation"}]
`, string(body))
}

func TestAnnotationsValidRequestErrorResponse(t *testing.T) {
	resp := handleAnnotationRequest(`
	{
		"annotation": {
			"name": "Deploys",
			"Datasource": "My datasource",
			"IconColor": "red",
			"Enable": true,
			"Query": "query"
		},
		"range": {
			"from": "2020-01-02T12:00:00.0Z",
			"to": "2020-01-07T20:00:00.0Z"
		}
	}
	`, func(req AnnotationsRequest) ([]AnnotationsResponse, error) {
		return nil, errors.New("something bad happened")
	})

	body, _ := ioutil.ReadAll(resp.Body)
	assert.Equal(t, 400, resp.StatusCode)
	assert.Equal(t, "something bad happened\n", string(body))
}
