package oidc

import "strings"

// Context format used to forward information to Open ID server
type Context map[string]string

func NewContext() Context {
	return make(map[string]string)
}

func (ctx Context) SetAMR(amr MethodRefs) Context {
	ctx["amr"] = amr.String()
	return ctx
}

func (ctx Context) GetAMR() []string {
	return strings.Split(ctx["amr"], " ")
}
