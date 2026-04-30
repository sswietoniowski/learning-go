package common

import (
	"errors"
	"net/http"
	"strings"

	echo "github.com/labstack/echo/v4"

	"eats/backend/common/log"
)

func EchoErrorHandler(err error, c echo.Context) {
	if c.Response().Committed {
		return
	}

	httpErrorResponse, httpStatus := httpErrorResponseFromErr(err)

	log.FromContext(c.Request().Context()).With("err", err).Error("Handling HTTP error")

	if err := c.JSON(httpStatus, httpErrorResponse); err != nil {
		log.FromContext(c.Request().Context()).With("error", err).Error("Failed to send error response")
	}
}

type HttpErrorResponse struct {
	Message string            `json:"message"`
	Slug    string            `json:"slug"`
	Details []HttpErrorDetail `json:"details,omitempty"`
}

type HttpErrorDetail struct {
	EntityType string `json:"entity_type,omitempty"`
	EntityID   string `json:"entity_id,omitempty"`
	ErrorSlug  string `json:"error_slug,omitempty"`
	Message    string `json:"message,omitempty"`
}

func httpErrorResponseFromErr(err error) (HttpErrorResponse, int) {
	publicError := "Internal Server Error"
	statusCode := http.StatusInternalServerError

	var he *echo.HTTPError
	if errors.As(err, &he) {
		statusCode = he.Code
		publicError = http.StatusText(statusCode)
	}
	errorSlug := strings.ToLower(strings.ReplaceAll(publicError, " ", "_"))

	var commonErr Error
	if errors.As(err, &commonErr) {
		if commonErr.PublicError != "" {
			publicError = commonErr.PublicError
		}
		if commonErr.ErrorSlug != "" {
			errorSlug = commonErr.ErrorSlug
		}
		if commonErr.HttpErrorCode != 0 {
			statusCode = commonErr.HttpErrorCode
		}
	}

	httpDetails := make([]HttpErrorDetail, 0, len(commonErr.Details))
	for _, detail := range commonErr.Details {
		httpDetails = append(httpDetails, HttpErrorDetail(detail))
	}

	httpErrorResponse := HttpErrorResponse{
		Slug:    errorSlug,
		Message: publicError,
		Details: httpDetails,
	}

	return httpErrorResponse, statusCode
}
