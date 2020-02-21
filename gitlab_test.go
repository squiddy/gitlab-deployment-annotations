package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/xanzy/go-gitlab"
)

func setup() (*http.ServeMux, *httptest.Server, *gitlab.Client) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)

	client := gitlab.NewClient(nil, "")
	client.SetBaseURL(server.URL)

	return mux, server, client
}

func teardown(server *httptest.Server) {
	server.Close()
}

func TestGetFilteredDeploymentsSuccess(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/projects/16/deployments", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, r.Method, "GET")
		assert.Equal(t, r.URL.RawQuery, "environment=Live&order_by=created_at&sort=desc&status=success&updated_after=2020-01-10T05%3A00%3A00Z&updated_before=2020-02-03T12%3A00%3A00Z")
		fmt.Fprint(w, `[{"id": 123}]`)
	})

	from := time.Date(2020, 1, 10, 5, 0, 0, 0, time.UTC)
	to := time.Date(2020, 2, 3, 12, 0, 0, 0, time.UTC)
	deployments, err := GetFilteredDeployments(client, 16, "Live", from, to)
	assert.Nil(t, err)
	assert.Equal(t, deployments, []*gitlab.Deployment{
		&gitlab.Deployment{ID: 123},
	})
}

func TestGetFilteredDeploymentsFailure(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/projects/16/deployments", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "failure", http.StatusBadRequest)
	})

	deployments, err := GetFilteredDeployments(client, 16, "Live", time.Now(), time.Now())
	assert.Equal(t, err.(*gitlab.ErrorResponse).Response.StatusCode, 400)
	assert.Nil(t, deployments)
}
