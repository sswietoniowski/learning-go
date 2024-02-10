package helper

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
)

const ContentTypeHeader = "Content-Type"
const JsonContentType = "application/json"

// IsValidMethod checks if the request method is the expected one and returns a boolean.
func IsValidMethod(w http.ResponseWriter, r *http.Request, expectedMethod string) bool {
	if r.Method != expectedMethod {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return false
	}

	return true
}

// IsValidContentType checks if the request content type is the expected one and returns a boolean.
func IsValidContentType(w http.ResponseWriter, r *http.Request, expectedContentType string) bool {
	if r.Header.Get(ContentTypeHeader) != expectedContentType {
		http.Error(w, "Invalid Content-Type, expected 'application/json'", http.StatusUnsupportedMediaType)
		return false
	}

	return true
}

// ParseJsonRequest reads the request body and decodes the JSON data into the provided interface data and returns an error if any.
func ParseJsonRequest(w http.ResponseWriter, r *http.Request, data interface{}) error {
	const maxBodySize = 1_048_576 // 1MB
	r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)

	// we must use a JSON decoder to enforce strict JSON parsing
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(data); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return errors.New("could not decode request data from JSON")
	}

	if err := dec.Decode(&struct{}{}); err != io.EOF {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return errors.New("request body must only contain a single JSON object")
	}

	return nil
}

// ExtractIdFromRoute extracts the id from the request path and returns it as an int64 or an error if any.
func ExtractIdFromRoute(w http.ResponseWriter, r *http.Request, path string) (int64, error) {
	id := r.URL.Path[len(path):]
	idInt, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return 0, errors.New("bad request, unable to extract id from path")
	}

	return idInt, nil
}

// SendJsonResponse encodes the provided data to JSON and sends it as the response with the provided status code and returns an error if any.
func SendJsonResponse(w http.ResponseWriter, statusCode int, data interface{}) error {
	dataJSON, err := json.Marshal(data)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return errors.New("could not encode response data to JSON")
	}

	w.Header().Set(ContentTypeHeader, JsonContentType)
	if statusCode > 0 {
		w.WriteHeader(statusCode)
	}
	if data != nil {
		w.Write(dataJSON)
	}

	return nil
}
