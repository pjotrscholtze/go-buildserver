// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"
)

// StartPipelineOKCode is the HTTP code returned for type StartPipelineOK
const StartPipelineOKCode int = 200

/*StartPipelineOK Queued pipeline

swagger:response startPipelineOK
*/
type StartPipelineOK struct {
}

// NewStartPipelineOK creates StartPipelineOK with default headers values
func NewStartPipelineOK() *StartPipelineOK {

	return &StartPipelineOK{}
}

// WriteResponse to the client
func (o *StartPipelineOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.Header().Del(runtime.HeaderContentType) //Remove Content-Type on empty responses

	rw.WriteHeader(200)
}
