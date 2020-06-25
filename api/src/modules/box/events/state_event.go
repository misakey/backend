package events

import (
	"context"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/volatiletech/sqlboiler/boil"
	"github.com/volatiletech/sqlboiler/types"
)

type StateLifecycleContent struct {
	State string `json:"state"`
}

func (c *StateLifecycleContent) Unmarshal(content types.JSON) error {
	return content.Unmarshal(c)
}

func (c StateLifecycleContent) Validate() error {
	return v.ValidateStruct(&c,
		v.Field(&c.State, v.Required, v.In("closed")),
	)
}

func GetStateLifecycle(
	ctx context.Context,
	exec boil.ContextExecutor,
	boxID, lifecycle string,
) (Event, error) {
	jsonQuery := `{"state": "` + lifecycle + `"}`
	return findByTypeContent(ctx, exec, boxID, "state.lifecycle", &jsonQuery)
}
