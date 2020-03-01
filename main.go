package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/xanzy/go-gitlab"
)

var client *gitlab.Client

type Query struct {
	ProjectID   int    `json:"project_id"`
	Environment string `json:"environment"`
}

func main() {
	clientHTTP := (*http.Client)(nil)
	if os.Getenv("HTTPS_SKIP_VERIFY") == "1" {
		transCfg := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		clientHTTP = &http.Client{Transport: transCfg}
	}

	client = gitlab.NewClient(clientHTTP, os.Getenv("GITLAB_TOKEN"))
	client.SetBaseURL(os.Getenv("GITLAB_URL"))

	ds := &DataSource{}
	ds.HandleAnnotation(func(req AnnotationsRequest) ([]AnnotationsResponse, error) {
		var query Query
		if err := json.Unmarshal([]byte(req.Annotation.Query), &query); err != nil {
			return nil, fmt.Errorf("Couldn't parse query: %w", err)
		}

		ds, err := GetFilteredDeployments(client, query.ProjectID, query.Environment, req.Range.From, req.Range.To)
		if err != nil {
			return nil, fmt.Errorf("Failed to get deployments: %w", err)
		}

		result := make([]AnnotationsResponse, 0)
		for _, d := range ds {
			result = append(result, AnnotationsResponse{
				Annotation: req.Annotation,
				Time:       d.CreatedAt.Unix() * 1000,
				Title:      d.Deployable.Commit.Title,
			})
		}
		return result, nil
	})
	if err := ds.ListenAndServe(os.Getenv("HTTP_ADDRESS")); err != nil {
		log.Fatalf("Error running server: %v", err)
	}
}
