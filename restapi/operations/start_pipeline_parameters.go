// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/validate"
)

// NewStartPipelineParams creates a new StartPipelineParams object
//
// There are no default values defined in the spec.
func NewStartPipelineParams() StartPipelineParams {

	return StartPipelineParams{}
}

// StartPipelineParams contains all the bound params for the start pipeline operation
// typically these are obtained from a http.Request
//
// swagger:parameters startPipeline
type StartPipelineParams struct {

	// HTTP Request Object
	HTTPRequest *http.Request `json:"-"`

	/*
	  In: body
	*/
	Data interface{}
	/*
	  Required: true
	  In: path
	*/
	Name string
	/*The reason for the build.
	  Required: true
	  In: query
	*/
	Reason string
}

// BindRequest both binds and validates a request, it assumes that complex things implement a Validatable(strfmt.Registry) error interface
// for simple values it will use straight method calls.
//
// To ensure default values, the struct must have been initialized with NewStartPipelineParams() beforehand.
func (o *StartPipelineParams) BindRequest(r *http.Request, route *middleware.MatchedRoute) error {
	var res []error

	o.HTTPRequest = r

	qs := runtime.Values(r.URL.Query())

	if runtime.HasBody(r) {
		defer r.Body.Close()
		var body interface{}
		if err := route.Consumer.Consume(r.Body, &body); err != nil {
			res = append(res, errors.NewParseError("data", "body", "", err))
		} else {
			// no validation on generic interface
			o.Data = body
		}
	}

	rName, rhkName, _ := route.Params.GetOK("name")
	if err := o.bindName(rName, rhkName, route.Formats); err != nil {
		res = append(res, err)
	}

	qReason, qhkReason, _ := qs.GetOK("reason")
	if err := o.bindReason(qReason, qhkReason, route.Formats); err != nil {
		res = append(res, err)
	}
	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

// bindName binds and validates parameter Name from path.
func (o *StartPipelineParams) bindName(rawData []string, hasKey bool, formats strfmt.Registry) error {
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}

	// Required: true
	// Parameter is provided by construction from the route
	o.Name = raw

	return nil
}

// bindReason binds and validates parameter Reason from query.
func (o *StartPipelineParams) bindReason(rawData []string, hasKey bool, formats strfmt.Registry) error {
	if !hasKey {
		return errors.Required("reason", "query", rawData)
	}
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}

	// Required: true
	// AllowEmptyValue: false

	if err := validate.RequiredString("reason", "query", raw); err != nil {
		return err
	}
	o.Reason = raw

	return nil
}
