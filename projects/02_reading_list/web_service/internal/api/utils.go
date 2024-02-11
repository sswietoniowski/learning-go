package api

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
)

const contentTypeHeader = "Content-Type"
const jsonContentType = "application/json"

// IsValidMethod checks if the request method is the expected one
// and returns a boolean.
func IsValidMethod(w http.ResponseWriter, r *http.Request, expectedMethod string) bool {
	return r.Method == expectedMethod
}

// IsValidContentType checks if the request content type is the expected one
// and returns a boolean.
func IsValidContentType(w http.ResponseWriter, r *http.Request, expectedContentType string) bool {
	return r.Header.Get(contentTypeHeader) == expectedContentType
}

// ParseJsonRequest reads the request body and decodes the JSON data into the
// provided interface data and returns an error if any.
func ParseJsonRequest(w http.ResponseWriter, r *http.Request, data interface{}) error {
	const maxBodySize = 1_048_576 // 1MB
	r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)

	// we must use a JSON decoder to enforce strict JSON parsing
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(data); err != nil {
		return errors.New("could not decode request data from JSON")
	}

	if err := dec.Decode(&struct{}{}); err != io.EOF {
		return errors.New("request body must only contain a single JSON object")
	}

	return nil
}

// ExtractIdFromRoute extracts the id from the request path and returns
// it as an int64 or an error if any.
func ExtractIdFromRoute(w http.ResponseWriter, r *http.Request, path string) (int64, error) {
	id := r.URL.Path[len(path):]
	idInt, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return 0, errors.New("bad request, unable to extract id from path")
	}

	return idInt, nil
}

// sendJsonResponse encodes the provided data to JSON and sends it as the
// response with the provided status code and returns an error if any.
func sendJsonResponse(w http.ResponseWriter, statusCode int, data interface{}, headers http.Header) error {
	dataJSON, err := json.Marshal(data)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return errors.New("could not encode response data to JSON")
	}

	for key, value := range headers {
		w.Header()[key] = value
	}

	w.Header().Set(contentTypeHeader, jsonContentType)
	if statusCode > 0 {
		w.WriteHeader(statusCode)
	}
	if data != nil {
		w.Write(dataJSON)
	}

	return nil
}

func SendOk(w http.ResponseWriter, data interface{}) error {
	return sendJsonResponse(w, http.StatusOK, data, nil)
}

func SendCreated(w http.ResponseWriter, data interface{}, location string) error {
	headers := make(http.Header)
	headers.Set("Location", location)
	return sendJsonResponse(w, http.StatusCreated, data, headers)
}

func SendNoContent(w http.ResponseWriter) error {
	return sendJsonResponse(w, http.StatusNoContent, nil, nil)
}

func SendNotFound(w http.ResponseWriter) error {
	return sendJsonResponse(w, http.StatusNotFound, nil, nil)
}

func SendBadRequest(w http.ResponseWriter) error {
	return sendJsonResponse(w, http.StatusBadRequest, nil, nil)
}

func SendMethodNotAllowed(w http.ResponseWriter) error {
	return sendJsonResponse(w, http.StatusMethodNotAllowed, nil, nil)
}

func SendUnsupportedMediaType(w http.ResponseWriter) error {
	return sendJsonResponse(w, http.StatusUnsupportedMediaType, nil, nil)
}

func SendInternalServerError(w http.ResponseWriter) error {
	return sendJsonResponse(w, http.StatusInternalServerError, nil, nil)
}
