package merr

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefautlError(t *testing.T) {
	t.Run("different addresses", func(t *testing.T) {
		err1 := NotFound()
		err2 := NotFound()
		addErr1 := fmt.Sprintf("%p", &err1)
		addErr2 := fmt.Sprintf("%p", &err2)
		assert.NotEqual(t, addErr1, addErr2)
		addDet1 := fmt.Sprintf("%p", err1.Details)
		addDet2 := fmt.Sprintf("%p", err2.Details)
		assert.NotEqual(t, addDet1, addDet2)
	})
}
