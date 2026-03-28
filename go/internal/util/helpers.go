package util

import "errors"

// DeferErrJoin joins the error returned by the callable function to the
// provided error pointer. If the callable function returns an error, it is
// joined to the provided error pointer. If the callable function does not
// return an error, the provided error pointer is not modified.
// This function is useful for deferring the execution of a function that may
// return an error, and then joining the error with the provided error pointer.
func DeferErrJoin(callable func() error, err *error) *error {
	*err = errors.Join(*err, callable())
	return err
}
