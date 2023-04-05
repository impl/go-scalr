package scalr

import "time"

type StatusTransition struct {
	ID        string    `jsonapi:"primary,status-transitions"`
	Status    RunStatus `jsonapi:"attr,status"`
	Reason    *string   `jsonapi:"attr,reason,omitempty"`
	CreatedAt time.Time `jsonapi:"attr,occurred-at,iso8601"`
}
