// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
	"github.com/go-openapi/validate"
)

// Job job
//
// swagger:model Job
type Job struct {

	// build reason
	BuildReason string `json:"BuildReason,omitempty"`

	// origin
	Origin string `json:"Origin,omitempty"`

	// queue time
	// Format: date-time
	QueueTime strfmt.DateTime `json:"QueueTime,omitempty"`

	// repo name
	RepoName string `json:"RepoName,omitempty"`
}

// Validate validates this job
func (m *Job) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateQueueTime(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *Job) validateQueueTime(formats strfmt.Registry) error {
	if swag.IsZero(m.QueueTime) { // not required
		return nil
	}

	if err := validate.FormatOf("QueueTime", "body", "date-time", m.QueueTime.String(), formats); err != nil {
		return err
	}

	return nil
}

// ContextValidate validates this job based on context it is used
func (m *Job) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *Job) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *Job) UnmarshalBinary(b []byte) error {
	var res Job
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}