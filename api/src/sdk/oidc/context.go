package oidc

import (
	"github.com/volatiletech/null/v8"
)

// Context format used to forward information to Open ID server
type Context map[string]interface{}

func NewContext() Context {
	return make(map[string]interface{})
}

func (ctx Context) SetACRValues(acrs ClassRefs) Context {
	ctx["acr_values"] = acrs
	return ctx
}

func (ctx Context) SetACRValue(acr ClassRef) Context {
	ctx["acr_values"] = NewClassRefs(acr)
	return ctx
}

func (ctx Context) ACRValues() ClassRefs {
	acrs, ok := ctx["acr_values"]
	if !ok {
		return ClassRefs{}
	}
	// if the context has been built by a json marshaling - it is []interface{}
	// this case is handled by resetting it properly
	ret, ok := acrs.(ClassRefs)
	if !ok {
		acr := ACR0
		for _, strACR := range acrs.([]interface{}) {
			acr = ClassRef(strACR.(string))
			break
		}
		ctx.SetACRValue(acr)
		return NewClassRefs(acr)
	}
	return ret
}

func (ctx Context) SetAMRs(amrs MethodRefs) Context {
	ctx["amrs"] = amrs
	return ctx
}

func (ctx Context) AddAMR(amr MethodRef) Context {
	stored, ok := ctx["amrs"]
	if !ok {
		stored = MethodRefs{}
	}
	amrs := stored.(MethodRefs)
	amrs.Add(amr)
	return ctx.SetAMRs(amrs)
}

func (ctx Context) AMRs() MethodRefs {
	amrs, ok := ctx["amrs"]
	if !ok {
		return MethodRefs{}
	}
	// if the context has been built by a json marshaling - it is []interface{}
	// this case is handled by resetting it properly
	ret, ok := amrs.(MethodRefs)
	if !ok {
		newAMRs := MethodRefs{}
		for _, strAMR := range amrs.([]interface{}) {
			newAMRs.Add(MethodRef(strAMR.(string)))
		}
		ctx.SetAMRs(newAMRs)
		return newAMRs
	}
	return ret
}

func (ctx Context) SetAID(accountID null.String) Context {
	// ignore account id set if acr < 2
	if ctx.ACRValues().Get().LessThan(ACR2) {
		return ctx
	}
	ctx["aid"] = accountID.String
	return ctx
}

func (ctx Context) AID() null.String {
	aid, ok := ctx["aid"]
	if !ok {
		return null.String{}
	}
	return null.StringFrom(aid.(string))
}

func (ctx Context) SetMID(identityID string) Context {
	ctx["mid"] = identityID
	return ctx
}

func (ctx Context) MID() string {
	mid, ok := ctx["mid"]
	if !ok {
		return ""
	}
	return mid.(string)
}

func (ctx Context) SetLoginHint(loginHint string) Context {
	ctx["login_hint"] = loginHint
	return ctx
}

func (ctx Context) LoginHint() string {
	lh, ok := ctx["login_hint"]
	if !ok {
		return ""
	}
	return lh.(string)
}
