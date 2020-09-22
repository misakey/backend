// TO COPY - TO ADAPT
// merror is Misakey Error package, it help us to meet contracted error format defined in our convention
// it also allow us to define error code linked to our domain that would never change and will be used by consumers
package merror

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
	Co      Code              `json:"code"`
	Ori     Origin            `json:"origin"`
	Desc    string            `json:"desc"`
	Details map[string]string `json:"details"`

	// internal logic to know if the error is supposed to be updated or not
	// by default: false (we do update)
	// can be modified after the use of If or End function.
	noUpdate bool
}

// Cause returns the raw error contained within Error
func (e *Error) Cause() error {
	return e.error
}

// Error returns the error as a printable string
func (e Error) Error() string {
	return e.Desc + ": " + e.error.Error()
}

// Clear returns a boolean establishing if Error state is clear or not
func (e *Error) Clear() bool {
	// 1. if noUpdate is set to true, the If method has been called without calling End method afterwards.
	return !e.noUpdate
}

// Freeze possible updates if the Error has not same cause as sent error.
// Unfreeze is possible using End or this same method.
func (e Error) If(match error) Error {
	if Cause(e) != Cause(match) {
		e.noUpdate = true
	} else {
		e.noUpdate = false
	}
	return e
}

// End triggers the end of a If method - unfreeze updates of the Error.
func (e Error) End() Error {
	e.noUpdate = false
	return e
}

// From set Origin attributes
func (e Error) Code(c Code) Error {
	if e.noUpdate {
		return e
	}
	e.Co = c
	return e
}

// From set Origin attributes - it can be set only if current origin is NotDefined
func (e Error) From(ori Origin) Error {
	if e.noUpdate || e.Ori != OriNotDefined {
		return e
	}
	e.Ori = ori
	return e
}

// Add desc to the Error (concat with existing one)
func (e Error) Describe(desc string) Error {
	if e.noUpdate {
		return e
	}
	if len(e.Desc) > 0 {
		desc = fmt.Sprintf("%s: %s", desc, e.Desc)
	}
	e.Desc = desc
	return e
}

// Set desc to the Error using Sprintf (concat with existing one)
func (e Error) Describef(desc string, a ...interface{}) Error {
	return e.Describe(fmt.Sprintf(desc, a...))
}

// FlushDesc of the Error
func (e Error) FlushDesc() Error {
	e.Desc = ""
	return e
}

// Flush all details of the Error
func (e Error) Flush() Error {
	e.Desc = ""
	e.Details = nil
	e.Ori = OriNotDefined
	return e
}

// Add a key/value detail to the error Details map
func (e Error) Detail(k string, v string) Error {
	if e.noUpdate {
		return e
	}
	e.Details[k] = v
	return e
}

// Reset Detail values
func (e Error) ResetDetails() Error {
	if e.noUpdate {
		return e
	}
	e.Details = make(map[string]string)
	return e
}

//
// Helpers...
//

// Transform transforms err into merror
func Transform(err error) Error {
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
	// finally use external errors.Cause after our merror layer has vanished
	return errors.Cause(err)
}

// transform a native error in to a merror.Error
func transform(err error) Error {
	merr, ok := err.(Error)
	if !ok {
		return Error{
			error:   err,
			Co:      ToCode(err),
			Ori:     OriNotDefined,
			Desc:    err.Error(),
			Details: make(map[string]string),
		}
	}
	return merr
}
