package scalr

import (
	"context"
	"errors"
	"fmt"
	"net/url"
)

// Compile-time proof of interface implementation.
var _ ConfigurationVersions = (*configurationVersions)(nil)

// ConfigurationVersions describes all the configuration version related
// methods that the Scalr API supports.
type ConfigurationVersions interface {
	// Create is used to create a new configuration version. The created
	// configuration version will be usable once data is uploaded to it.
	Create(ctx context.Context, options ConfigurationVersionCreateOptions) (*ConfigurationVersion, error)

	// Read a configuration version by its ID.
	Read(ctx context.Context, cvID string) (*ConfigurationVersion, error)
}

// configurationVersions implements ConfigurationVersions.
type configurationVersions struct {
	client *Client
}

// ConfigurationStatus represents a configuration version status.
type ConfigurationStatus string

// List all available configuration version statuses.
const (
	ConfigurationErrored  ConfigurationStatus = "errored"
	ConfigurationPending  ConfigurationStatus = "pending"
	ConfigurationUploaded ConfigurationStatus = "uploaded"
)

// ConfigurationVersion is a representation of an uploaded or ingressed
// Terraform configuration in Scalr. A workspace must have at least one
// configuration version before any runs may be queued on it.
type ConfigurationVersion struct {
	ID     string              `jsonapi:"primary,configuration-versions"`
	Status ConfigurationStatus `jsonapi:"attr,status"`
	// Relations
	Workspace   *Workspace   `jsonapi:"relation,workspace"`
	VcsRevision *VcsRevision `jsonapi:"relation,vcs-revision,omitempty"`
}

// ConfigurationVersionCreateOptions represents the options for creating a
// configuration version.
type ConfigurationVersionCreateOptions struct {
	// For internal use only!
	ID string `jsonapi:"primary,configuration-versions"`

	// Indicates if a run should automatically be queued when the configuration
	// has been uploaded.
	AutoQueueRuns *bool `jsonapi:"attr,auto-queue-runs,omitempty"`

	Workspace *Workspace `jsonapi:"relation,workspace"`
}

func (o ConfigurationVersionCreateOptions) valid() error {
	if o.Workspace == nil {
		return errors.New("workspace is required")
	}
	if !validStringID(&o.Workspace.ID) {
		return errors.New("invalid value for workspace ID")
	}
	return nil
}

// Create is used to create a new configuration version.
func (s *configurationVersions) Create(ctx context.Context, options ConfigurationVersionCreateOptions) (*ConfigurationVersion, error) {
	if err := options.valid(); err != nil {
		return nil, err
	}

	// Make sure we don't send a user provided ID.
	options.ID = ""

	req, err := s.client.newRequest("POST", "configuration-versions", &options)
	if err != nil {
		return nil, err
	}

	cv := &ConfigurationVersion{}
	err = s.client.do(ctx, req, cv)
	if err != nil {
		return nil, err
	}

	return cv, nil
}

// Read a configuration version by its ID.
func (s *configurationVersions) Read(ctx context.Context, cvID string) (*ConfigurationVersion, error) {
	if !validStringID(&cvID) {
		return nil, errors.New("invalid value for configuration version ID")
	}

	options := struct {
		Include string `url:"include"`
	}{
		Include: "vcs-revision",
	}
	u := fmt.Sprintf("configuration-versions/%s", url.QueryEscape(cvID))
	req, err := s.client.newRequest("GET", u, options)
	if err != nil {
		return nil, err
	}

	cv := &ConfigurationVersion{}
	err = s.client.do(ctx, req, cv)
	if err != nil {
		return nil, err
	}

	return cv, nil
}
