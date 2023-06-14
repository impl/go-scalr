package scalr

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"
)

// Compile-time proof of interface implementation.
var _ StateVersions = (*stateVersions)(nil)

type StateVersions interface {
	// Read gets a state version resource from its ID.
	Read(ctx context.Context, stateVersionID string) (*StateVersion, error)

	// ReadCurrentForWorkspace gets the current state version for a given
	// workspace.
	ReadCurrentForWorkspace(ctx context.Context, workspaceID string) (*StateVersion, error)
}

// stateVersions implements StateVersions.
type stateVersions struct {
	client *Client
}

// StateVersionOutput describes a particular output of a state version.
type StateVersionOutput struct {
	Name      string      `json:"name"`
	Value     interface{} `json:"value"`
	Sensitive bool        `json:"sensitive"`
}

// StateVersionResource describes a resource in a state version.
type StateVersionResource struct {
	Type    string `json:"type"`
	Module  string `json:"module,omitempty"`
	Address string `json:"address"`
}

// StateVersion is a particular instance of Terraform state.
type StateVersion struct {
	ID        string                  `jsonapi:"primary,state-versions"`
	Outputs   []*StateVersionOutput   `jsonapi:"attr,outputs"`
	Resources []*StateVersionResource `jsonapi:"attr,resources"`
	Force     bool                    `jsonapi:"attr,force"`
	Lineage   string                  `jsonapi:"attr,lineage"`
	MD5       string                  `jsonapi:"attr,md5"`
	Serial    int                     `jsonapi:"attr,serial"`
	Size      int                     `jsonapi:"attr,size"`
	CreatedAt time.Time               `jsonapi:"attr,created-at,iso8601"`

	// Relations
	Run                  *Run          `jsonapi:"relation,run,omitempty"`
	NextStateVersion     *StateVersion `jsonapi:"relation,next-state-version,omitempty"`
	PreviousStateVersion *StateVersion `jsonapi:"relation,previous-state-version,omitempty"`
	Workspace            *Workspace    `jsonapi:"relation,workspace"`
}

func (s *stateVersions) Read(ctx context.Context, stateVersionID string) (*StateVersion, error) {
	if !validStringID(&stateVersionID) {
		return nil, errors.New("invalid value for state version ID")
	}

	u := fmt.Sprintf("state-versions/%s", url.PathEscape(stateVersionID))
	req, err := s.client.newRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}

	sv := &StateVersion{}
	if err := s.client.do(ctx, req, sv); err != nil {
		return nil, err
	}

	return sv, nil
}

func (s *stateVersions) ReadCurrentForWorkspace(ctx context.Context, workspaceID string) (*StateVersion, error) {
	if !validStringID(&workspaceID) {
		return nil, errors.New("invalid value for workspace ID")
	}

	u := fmt.Sprintf("workspaces/%s/current-state-version", url.PathEscape(workspaceID))
	req, err := s.client.newRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}

	sv := &StateVersion{}
	if err := s.client.do(ctx, req, sv); err != nil {
		return nil, err
	}

	return sv, nil
}
