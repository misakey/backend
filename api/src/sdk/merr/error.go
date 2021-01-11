// Package merr is Misakey Error package, it help us to meet contracted error format defined in our convention
// it also allow us to define error code linked to our domain that would never change and will be used by consumers
package merr

import (
	"fmt"

	"github.com/pkg/errors"
)

//
// Global scheme of errors (in JSON):
// {
//      "code": "{Code}",
//      "origin": "{Origin}"
//      "desc": "free format description",
//      "details": {
//          {string}: {string},
//          {string}: {string},
//      },
// }
//

// Error defines internal errors type we deal with in misakey domain layers
type Error struct {
	error
	Co          Code              `json:"code"`
	Origin      Origin            `json:"origin"`
	Description string            `json:"desc"`
	Details     map[string]string `json:"details"`
}

// Cause returns the raw error contained within Error
func (e *Error) Cause() error {
	return e.error
}

// Error returns the error as a printable string
func (e Error) Error() string {
	return e.Description + ": " + e.error.Error()
}

// End triggers the end of a If method - unfreeze updates of the Error.
func (e Error) End() Error {
	return e
}

// Code set code attribute
func (e Error) Code(c Code) Error {
	e.Co = c
	return e
}

// Ori set the Origin attributes - it can be set only if current origin is NotDefined
func (e Error) Ori(ori Origin) Error {
	if e.Origin != OriNotDefined {
		return e
	}
	e.Origin = ori
	return e
}

// Desc the Error (concat with existing one)
func (e Error) Desc(desc string) Error {
	if len(e.Description) > 0 {
		desc = fmt.Sprintf("%s: %s", desc, e.Description)
	}
	e.Description = desc
	return e
}

// Descfribe the Error using Sprintf (concat with existing one)
func (e Error) Descf(desc string, a ...interface{}) Error {
	return e.Desc(fmt.Sprintf(desc, a...))
}

// Flush all details of the Error
func (e Error) Flush() Error {
	e.Description = ""
	e.Details = nil
	e.Origin = OriNotDefined
	return e
}

// Add a key/value detail to the error Details map
func (e Error) Add(k string, v string) Error {
	e.Details[k] = v
	return e
}

//
// Helpers...
//

// From returns a merr.Error using the received err
func From(err error) Error {
	return transform(err)
}

// Cause returns the underlying cause of the error, if possible.
// An error value has a cause if it implements the following
// interface:
//
//     type causer interface {
//            Cause() error
//     }
//
// If the error does not implement Cause, the original error will
// be returned. If the error is nil, nil will be returned without further
// investigation.
func Cause(err error) error {
	for err != nil {
		cause, ok := err.(Error)
		if !ok {
			break
		}
		err = cause.Cause()
	}
	// finally use external errors.Cause after our merr layer has vanished
	return errors.Cause(err)
}

// transform a native error in to a merr.Error
func transform(err error) Error {
	merr, ok := err.(Error)
	if !ok {
		return Error{
			error:       err,
			Co:          ToCode(err),
			Origin:      OriNotDefined,
			Description: err.Error(),
			Details:     make(map[string]string),
		}
	}
	return merr
}
