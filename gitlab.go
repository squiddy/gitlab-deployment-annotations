package main

import (
	"fmt"
	"time"

	"github.com/xanzy/go-gitlab"
)

// ListDeployments extends ListProjectDeploymentsOptions with parameters that
// go-gitlab doesn't yet support.
type ListDeployments struct {
	gitlab.ListProjectDeploymentsOptions
	Status        *string `url:"status,omitempty"`
	Environment   *string `url:"environment,omitempty"`
	UpdatedBefore *string `url:"updated_before,omitempty"`
	UpdatedAfter  *string `url:"updated_after,omitempty"`
}

// GetFilteredDeployments returns a list of deployments for the given project in
// the given time interval.
func GetFilteredDeployments(client *gitlab.Client, projectID int,
	environment string, from time.Time, to time.Time) ([]*gitlab.Deployment, error) {

	opts := &ListDeployments{
		Status:        gitlab.String("success"),
		Environment:   gitlab.String(environment),
		UpdatedAfter:  gitlab.String(from.Format(time.RFC3339)),
		UpdatedBefore: gitlab.String(to.Format(time.RFC3339)),

		ListProjectDeploymentsOptions: gitlab.ListProjectDeploymentsOptions{
			OrderBy: gitlab.String("created_at"),
			Sort:    gitlab.String("desc"),
		},
	}

	// TODO: Use ListProjectDeployments from go-gitlab once it supports the
	// 		 other filter parameters
	path := fmt.Sprintf("projects/%d/deployments", projectID)
	req, err := client.NewRequest("GET", path, opts, nil)
	if err != nil {
		return nil, err
	}

	var deployments []*gitlab.Deployment
	_, err = client.Do(req, &deployments)
	if err != nil {
		return nil, err
	}

	return deployments, nil
}
