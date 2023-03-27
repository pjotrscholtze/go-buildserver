// Code generated by go-swagger; DO NOT EDIT.

package restapi

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"encoding/json"
)

var (
	// SwaggerJSON embedded version of the swagger document used at generation time
	SwaggerJSON json.RawMessage
	// FlatSwaggerJSON embedded flattened version of the swagger document used at generation time
	FlatSwaggerJSON json.RawMessage
)

func init() {
	SwaggerJSON = json.RawMessage([]byte(`{
  "swagger": "2.0",
  "info": {
    "title": "Go Buildserver",
    "version": "1.0.0"
  },
  "basePath": "/api",
  "paths": {
    "/repos": {
      "get": {
        "produces": [
          "application/json",
          "application/xml"
        ],
        "summary": "Get repos",
        "operationId": "listRepos",
        "responses": {
          "200": {
            "description": "Successful operation",
            "schema": {
              "type": "array",
              "items": {
                "$ref": "#/definitions/Repo"
              },
              "xml": {
                "name": "addresses",
                "wrapped": true
              }
            }
          }
        }
      }
    },
    "/repos/{name}": {
      "post": {
        "consumes": [
          "application/json",
          "application/xml",
          "application/x-www-form-urlencoded"
        ],
        "summary": "Start build",
        "operationId": "startBuild",
        "parameters": [
          {
            "type": "string",
            "name": "name",
            "in": "path",
            "required": true
          },
          {
            "name": "data",
            "in": "body",
            "required": true,
            "schema": {
              "type": "object",
              "additionalProperties": true
            }
          },
          {
            "type": "string",
            "description": "The reason for the build.",
            "name": "reason",
            "in": "query",
            "required": true
          }
        ],
        "responses": {
          "200": {
            "description": "Started build"
          }
        }
      }
    }
  },
  "definitions": {
    "BuildResult": {
      "type": "object",
      "properties": {
        "Lines": {
          "type": "array",
          "items": {
            "$ref": "#/definitions/BuildResultLine"
          },
          "xml": {
            "wrapped": true
          }
        },
        "Reason": {
          "type": "string"
        },
        "StartTime": {
          "type": "string",
          "format": "date-time"
        },
        "Status": {
          "type": "string"
        }
      }
    },
    "BuildResultLine": {
      "type": "object",
      "properties": {
        "Line": {
          "type": "string"
        },
        "Pipe": {
          "type": "string"
        },
        "Time": {
          "type": "string",
          "format": "date-time"
        }
      }
    },
    "Repo": {
      "type": "object",
      "properties": {
        "BuildScript": {
          "type": "string"
        },
        "ForceCleanBuild": {
          "type": "boolean"
        },
        "LastBuildResult": {
          "type": "array",
          "items": {
            "$ref": "#/definitions/BuildResult"
          },
          "xml": {
            "wrapped": true
          }
        },
        "Name": {
          "type": "string"
        },
        "Triggers": {
          "type": "array",
          "items": {
            "$ref": "#/definitions/Trigger"
          },
          "xml": {
            "wrapped": true
          }
        },
        "URL": {
          "type": "string"
        }
      }
    },
    "Trigger": {
      "type": "object",
      "properties": {
        "Kind": {
          "type": "string"
        },
        "Schedule": {
          "type": "string"
        }
      }
    }
  }
}`))
	FlatSwaggerJSON = json.RawMessage([]byte(`{
  "swagger": "2.0",
  "info": {
    "title": "Go Buildserver",
    "version": "1.0.0"
  },
  "basePath": "/api",
  "paths": {
    "/repos": {
      "get": {
        "produces": [
          "application/json",
          "application/xml"
        ],
        "summary": "Get repos",
        "operationId": "listRepos",
        "responses": {
          "200": {
            "description": "Successful operation",
            "schema": {
              "type": "array",
              "items": {
                "$ref": "#/definitions/Repo"
              },
              "xml": {
                "name": "addresses",
                "wrapped": true
              }
            }
          }
        }
      }
    },
    "/repos/{name}": {
      "post": {
        "consumes": [
          "application/json",
          "application/x-www-form-urlencoded",
          "application/xml"
        ],
        "summary": "Start build",
        "operationId": "startBuild",
        "parameters": [
          {
            "type": "string",
            "name": "name",
            "in": "path",
            "required": true
          },
          {
            "name": "data",
            "in": "body",
            "required": true,
            "schema": {
              "type": "object",
              "additionalProperties": true
            }
          },
          {
            "type": "string",
            "description": "The reason for the build.",
            "name": "reason",
            "in": "query",
            "required": true
          }
        ],
        "responses": {
          "200": {
            "description": "Started build"
          }
        }
      }
    }
  },
  "definitions": {
    "BuildResult": {
      "type": "object",
      "properties": {
        "Lines": {
          "type": "array",
          "items": {
            "$ref": "#/definitions/BuildResultLine"
          },
          "xml": {
            "wrapped": true
          }
        },
        "Reason": {
          "type": "string"
        },
        "StartTime": {
          "type": "string",
          "format": "date-time"
        },
        "Status": {
          "type": "string"
        }
      }
    },
    "BuildResultLine": {
      "type": "object",
      "properties": {
        "Line": {
          "type": "string"
        },
        "Pipe": {
          "type": "string"
        },
        "Time": {
          "type": "string",
          "format": "date-time"
        }
      }
    },
    "Repo": {
      "type": "object",
      "properties": {
        "BuildScript": {
          "type": "string"
        },
        "ForceCleanBuild": {
          "type": "boolean"
        },
        "LastBuildResult": {
          "type": "array",
          "items": {
            "$ref": "#/definitions/BuildResult"
          },
          "xml": {
            "wrapped": true
          }
        },
        "Name": {
          "type": "string"
        },
        "Triggers": {
          "type": "array",
          "items": {
            "$ref": "#/definitions/Trigger"
          },
          "xml": {
            "wrapped": true
          }
        },
        "URL": {
          "type": "string"
        }
      }
    },
    "Trigger": {
      "type": "object",
      "properties": {
        "Kind": {
          "type": "string"
        },
        "Schedule": {
          "type": "string"
        }
      }
    }
  }
}`))
}
