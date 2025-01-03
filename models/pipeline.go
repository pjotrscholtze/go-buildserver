// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"
	"strconv"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
)

// Pipeline pipeline
//
// swagger:model Pipeline
type Pipeline struct {

	// build script
	BuildScript string `json:"BuildScript,omitempty"`

	// force clean build
	ForceCleanBuild bool `json:"ForceCleanBuild,omitempty"`

	// last build result
	LastBuildResult []*BuildResult `json:"LastBuildResult" xml:"LastBuildResult"`

	// name
	Name string `json:"Name,omitempty"`

	// path
	Path string `json:"Path,omitempty"`

	// triggers
	Triggers []*Trigger `json:"Triggers" xml:"Triggers"`

	// URL
	URL string `json:"URL,omitempty"`
}

// Validate validates this pipeline
func (m *Pipeline) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateLastBuildResult(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateTriggers(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *Pipeline) validateLastBuildResult(formats strfmt.Registry) error {
	if swag.IsZero(m.LastBuildResult) { // not required
		return nil
	}

	for i := 0; i < len(m.LastBuildResult); i++ {
		if swag.IsZero(m.LastBuildResult[i]) { // not required
			continue
		}

		if m.LastBuildResult[i] != nil {
			if err := m.LastBuildResult[i].Validate(formats); err != nil {
				if ve, ok := err.(*errors.Validation); ok {
					return ve.ValidateName("LastBuildResult" + "." + strconv.Itoa(i))
				}
				return err
			}
		}

	}

	return nil
}

func (m *Pipeline) validateTriggers(formats strfmt.Registry) error {
	if swag.IsZero(m.Triggers) { // not required
		return nil
	}

	for i := 0; i < len(m.Triggers); i++ {
		if swag.IsZero(m.Triggers[i]) { // not required
			continue
		}

		if m.Triggers[i] != nil {
			if err := m.Triggers[i].Validate(formats); err != nil {
				if ve, ok := err.(*errors.Validation); ok {
					return ve.ValidateName("Triggers" + "." + strconv.Itoa(i))
				}
				return err
			}
		}

	}

	return nil
}

// ContextValidate validate this pipeline based on the context it is used
func (m *Pipeline) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	var res []error

	if err := m.contextValidateLastBuildResult(ctx, formats); err != nil {
		res = append(res, err)
	}

	if err := m.contextValidateTriggers(ctx, formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *Pipeline) contextValidateLastBuildResult(ctx context.Context, formats strfmt.Registry) error {

	for i := 0; i < len(m.LastBuildResult); i++ {

		if m.LastBuildResult[i] != nil {
			if err := m.LastBuildResult[i].ContextValidate(ctx, formats); err != nil {
				if ve, ok := err.(*errors.Validation); ok {
					return ve.ValidateName("LastBuildResult" + "." + strconv.Itoa(i))
				}
				return err
			}
		}

	}

	return nil
}

func (m *Pipeline) contextValidateTriggers(ctx context.Context, formats strfmt.Registry) error {

	for i := 0; i < len(m.Triggers); i++ {

		if m.Triggers[i] != nil {
			if err := m.Triggers[i].ContextValidate(ctx, formats); err != nil {
				if ve, ok := err.(*errors.Validation); ok {
					return ve.ValidateName("Triggers" + "." + strconv.Itoa(i))
				}
				return err
			}
		}

	}

	return nil
}

// MarshalBinary interface implementation
func (m *Pipeline) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *Pipeline) UnmarshalBinary(b []byte) error {
	var res Pipeline
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
