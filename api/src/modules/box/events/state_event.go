package events

import (
	"context"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/volatiletech/sqlboiler/types"
)

type stateLifecycleContent struct {
	State string `json:"state"`
}

func (c *stateLifecycleContent) Unmarshal(content types.JSON) error {
	return content.Unmarshal(c)
}

func (c stateLifecycleContent) Validate() error {
	return v.ValidateStruct(&c,
		v.Field(&c.State, v.Required, v.In("closed")),
	)
}

// today, state if only about lifecycle
func (c *boxComputer) playState(ctx context.Context, e Event) error {
	lifecycleContent := stateLifecycleContent{}
	if err := lifecycleContent.Unmarshal(e.Content); err != nil {
		return err
	}
	c.box.Lifecycle = lifecycleContent.State
	return nil
}
