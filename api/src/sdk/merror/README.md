# Misakey Error Package

merror package allows us to share same data format for API errors.

It allows us to use internal code to identify some specific behaviour to improve UX or our alerting system.

Misakey Errors are recognized by our error handlers to be treaten and formatted correctly.

#### Convention

Our error convention is described [here](https://gitlab.misakey.dev/Misakey/dev-manifesto/blob/master/conventions.md#error-conventions)

#### Error Introspection

We use the [error cause principles](https://godoc.org/github.com/pkg/errors#Cause) to introspect a Misakey Error.
