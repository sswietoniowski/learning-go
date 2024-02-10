package helper

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
)

const ContentTypeHeader = "Content-Type"
const JsonContentType = "application/json"

func IsValidMethod(w http.ResponseWriter, r *http.Request, expectedMethod string) bool {
	if r.Method != expectedMethod {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return false
	}

	return true
}

func IsValidContentType(w http.ResponseWriter, r *http.Request, expectedContentType string) bool {
	if r.Header.Get(ContentTypeHeader) != expectedContentType {
		http.Error(w, "Invalid Content-Type, expected 'application/json'", http.StatusUnsupportedMediaType)
		return false
	}

	return true
}

func ParseJsonRequest(w http.ResponseWriter, r *http.Request, data interface{}) error {
	err := json.NewDecoder(r.Body).Decode(data)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return errors.New("could not decode request data from JSON")
	}

	return nil
}

func ExtractIdFromRoute(w http.ResponseWriter, r *http.Request, path string) (int64, error) {
	id := r.URL.Path[len(path):]
	idInt, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return 0, errors.New("bad request, unable to extract id from path")
	}

	return idInt, nil
}

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
