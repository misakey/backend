package oidc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContext(t *testing.T) {
	ctx := NewContext()

	// set ACR value
	ctx.SetACRValue(ACR1)
	assert.Equal(t, NewClassRefs(ACR1), ctx.ACRValues(), "set/get acr values")

	// set ACR Values
	ctx.SetACRValues(NewClassRefs(ACR2))
	assert.Equal(t, NewClassRefs(ACR2), ctx.ACRValues(), "set/get acr value")

	// set AMRs
	ctx.SetAMRs(MethodRefs{AMRPrehashedPassword})
	assert.Equal(t, MethodRefs{AMRPrehashedPassword}, ctx.AMRs(), "set/get amrs")

	// add AMR
	ctx.AddAMR(AMREmailedCode)
	assert.Equal(t, MethodRefs{AMRPrehashedPassword, AMREmailedCode}, ctx.AMRs(), "add/get amrs")

	// add AMR on unset AMRs
	ctx = NewContext()
	ctx.AddAMR(AMREmailedCode)
	assert.Equal(t, MethodRefs{AMREmailedCode}, ctx.AMRs(), "add/get amrs")

}
