package scalr

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"
)

// Compile-time proof of interface implementation.
var _ Runs = (*runs)(nil)

// Runs describes all the run related methods that the Scalr API supports.
type Runs interface {

	// Read a run by its ID.
	Read(ctx context.Context, runID string) (*Run, error)
	// Create a new run with the given options.
	Create(ctx context.Context, options RunCreateOptions) (*Run, error)
	// Apply a run that is paused waiting for confirmation after a plan.
	Apply(ctx context.Context, runID string, options RunApplyOptions) error
	// Interrupt a run that is currently planning or applying.
	Cancel(ctx context.Context, runID string, options RunCancelOptions) error
	// Skip any remaining work on runs that are paused waiting for confirmation or priority.
	Discard(ctx context.Context, runID string, options RunDiscardOptions) error
}

// runs implements Runs.
type runs struct {
	client *Client
}

// RunStatus represents a run state.
type RunStatus string

// List all available run statuses.
const (
	RunApplied            RunStatus = "applied"
	RunApplyQueued        RunStatus = "apply_queued"
	RunApplying           RunStatus = "applying"
	RunCanceled           RunStatus = "canceled"
	RunConfirmed          RunStatus = "confirmed"
	RunCostEstimated      RunStatus = "cost_estimated"
	RunCostEstimating     RunStatus = "cost_estimating"
	RunDiscarded          RunStatus = "discarded"
	RunErrored            RunStatus = "errored"
	RunPending            RunStatus = "pending"
	RunPlanQueued         RunStatus = "plan_queued"
	RunPlanned            RunStatus = "planned"
	RunPlannedAndFinished RunStatus = "planned_and_finished"
	RunPlanning           RunStatus = "planning"
	RunPolicyChecked      RunStatus = "policy_checked"
	RunPolicyChecking     RunStatus = "policy_checking"
	RunPolicyOverride     RunStatus = "policy_override"
	RunPolicySoftFailed   RunStatus = "policy_soft_failed"
)

// RunSource represents a source type of a run.
type RunSource string

// List all available run sources.
const (
	RunSourceAPI                  RunSource = "api"
	RunSourceConfigurationVersion RunSource = "configuration-version"
	RunSourceUI                   RunSource = "ui"
	RunSourceVCS                  RunSource = "vcs"
	RunSourceCLI                  RunSource = "cli"
)

// Run represents a Scalr run.
type Run struct {
	ID        string    `jsonapi:"primary,runs"`
	Source    RunSource `jsonapi:"attr,source"`
	Message   string    `jsonapi:"attr,message"`
	IsDestroy bool      `jsonapi:"attr,is-destroy"`
	CreatedAt time.Time `jsonapi:"attr,created-at,iso8601"`
	Status    RunStatus `jsonapi:"attr,status"`

	// Relations
	VcsRevision          *VcsRevision          `jsonapi:"relation,vcs-revision"`
	Apply                *Apply                `jsonapi:"relation,apply"`
	ConfigurationVersion *ConfigurationVersion `jsonapi:"relation,configuration-version"`
	CostEstimate         *CostEstimate         `jsonapi:"relation,cost-estimate"`
	Plan                 *Plan                 `jsonapi:"relation,plan"`
	PolicyChecks         []*PolicyCheck        `jsonapi:"relation,policy-checks"`
	Workspace            *Workspace            `jsonapi:"relation,workspace"`
	StatusTransitions    []*StatusTransition   `jsonapi:"relation,status-transitions"`
}

// RunCreateOptions represents the options for creating a new run.
type RunCreateOptions struct {
	// For internal use only!
	ID string `jsonapi:"primary,runs"`

	// Specifies the configuration version to use for this run.
	ConfigurationVersion *ConfigurationVersion `jsonapi:"relation,configuration-version"`
	// Specifies the workspace where the run will be executed.
	Workspace *Workspace `jsonapi:"relation,workspace"`
}

func (o RunCreateOptions) valid() error {
	if o.Workspace == nil {
		return errors.New("workspace is required")
	}
	if !validStringID(&o.Workspace.ID) {
		return errors.New("invalid value for workspace ID")
	}
	if o.ConfigurationVersion == nil {
		return errors.New("configuration-version is required")
	}
	if !validStringID(&o.ConfigurationVersion.ID) {
		return errors.New("invalid value for configuration-version ID")
	}
	return nil
}

// Create a new run with the given options.
func (s *runs) Create(ctx context.Context, options RunCreateOptions) (*Run, error) {
	if err := options.valid(); err != nil {
		return nil, err
	}

	// Make sure we don't send a user provided ID.
	options.ID = ""

	req, err := s.client.newRequest("POST", "runs", &options)
	if err != nil {
		return nil, err
	}

	r := &Run{}
	err = s.client.do(ctx, req, r)
	if err != nil {
		return nil, err
	}

	return r, nil
}

// Read a run by its ID.
func (s *runs) Read(ctx context.Context, runID string) (*Run, error) {
	if !validStringID(&runID) {
		return nil, errors.New("invalid value for run ID")
	}

	options := struct {
		Include string `url:"include"`
	}{
		Include: "vcs-revision,status-transitions",
	}

	u := fmt.Sprintf("runs/%s", url.QueryEscape(runID))
	req, err := s.client.newRequest("GET", u, options)
	if err != nil {
		return nil, err
	}

	r := &Run{}
	err = s.client.do(ctx, req, r)
	if err != nil {
		return nil, err
	}

	return r, nil
}

// RunApplyOptions represents the options for applying a Run.
type RunApplyOptions struct {
	Comment string     `jsonapi:"attr,comment"`
	ApplyAt *time.Time `jsonapi:"attr,apply-at,iso8601"`
}

// Apply a run that is paused waiting for confirmation after a plan.
func (s *runs) Apply(ctx context.Context, runID string, options RunApplyOptions) error {
	if !validStringID(&runID) {
		return errors.New("invalid value for run ID")
	}

	u := fmt.Sprintf("runs/%s/actions/apply", url.PathEscape(runID))
	req, err := s.client.newRequest("POST", u, &options)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	return s.client.do(ctx, req, nil)
}

// RunCancelOptions represents the options for canceling a Run.
type RunCancelOptions struct {
	Comment string `jsonapi:"attr,comment"`
}

// Cancel interrupts a run that is currently planning or applying.
func (s *runs) Cancel(ctx context.Context, runID string, options RunCancelOptions) error {
	if !validStringID(&runID) {
		return errors.New("invalid value for run ID")
	}

	u := fmt.Sprintf("runs/%s/actions/cancel", url.PathEscape(runID))
	req, err := s.client.newRequest("POST", u, &options)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	return s.client.do(ctx, req, nil)
}

// RunDiscardOptions represents the options for discarding a Run.
type RunDiscardOptions struct {
	Comment string `jsonapi:"attr,comment"`
}

// Discard skips any remaining work on runs that are paused waiting for confirmation or priority.
func (s *runs) Discard(ctx context.Context, runID string, options RunDiscardOptions) error {
	if !validStringID(&runID) {
		return errors.New("invalid value for run ID")
	}

	u := fmt.Sprintf("runs/%s/actions/discard", url.PathEscape(runID))
	req, err := s.client.newRequest("POST", u, &options)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	return s.client.do(ctx, req, nil)
}
