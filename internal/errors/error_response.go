package errors

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type ErrorResponse struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	Cause     string `json:"cause"`
	Fix       string `json:"fix"`
	Timestamp string `json:"timestamp"`
}

type errorMeta struct {
	Code    string
	Message string
	Cause   string
	Fix     string
	Status  int
}

var errorRegistry = []struct {
	err  error
	meta errorMeta
}{
	{ErrInvalidID, errorMeta{
		Code: "INVALID_ID", Message: "Invalid identifier",
		Cause: "Invalid or non-existent ID provided.", Fix: "Check request parameters.", Status: http.StatusBadRequest,
	}},
	{ErrAuthHeaderMissing, errorMeta{
		Code: "AUTH_HEADER_MISSING", Message: "Missing or invalid authorization header",
		Cause: "Authorization header is missing or malformed.", Fix: "Add header: Authorization: Bearer <token>.", Status: http.StatusUnauthorized,
	}},
	{ErrUnauthorizedToken, errorMeta{
		Code: "UNAUTHORIZED_TOKEN", Message: "Token invalid or missing",
		Cause: "Token expired, invalid signature, wrong issuer, wrong audience, or missing.", Fix: "Request a new Keycloak access token.", Status: http.StatusUnauthorized,
	}},
	{ErrClaimsNotFound, errorMeta{
		Code: "CLAIMS_NOT_FOUND", Message: "Token claims not found",
		Cause: "Token missing expected fields.", Fix: "Request a new Keycloak access token.", Status: http.StatusUnauthorized,
	}},
	{ErrInvalidClaims, errorMeta{
		Code: "INVALID_CLAIMS", Message: "Invalid token claims",
		Cause: "Token content corrupted or malformed.", Fix: "Log in again.", Status: http.StatusUnauthorized,
	}},
	{ErrAdminOnly, errorMeta{
		Code: "ADMIN_ONLY", Message: "Admin access only",
		Cause: "Administrator privileges required for this action.", Fix: "Contact an administrator.", Status: http.StatusForbidden,
	}},
}

// defaultMeta — for unknown and validation errors.
var defaultMeta = errorMeta{
	Code: "INTERNAL_ERROR", Message: "Internal server error",
	Cause: "An unexpected error occurred.", Fix: "Retry the request later or contact support.", Status: http.StatusInternalServerError,
}

var validationMeta = errorMeta{
	Code: "VALIDATION_ERROR", Message: "Data validation error",
	Cause: "Submitted data does not match the expected format or rules.", Fix: "Check the request body and correct fields according to API documentation.", Status: http.StatusBadRequest,
}

func resolveMeta(err error) (errorMeta, int) {
	for _, e := range errorRegistry {
		if errors.Is(err, e.err) {
			return e.meta, e.meta.Status
		}
	}
	return defaultMeta, defaultMeta.Status
}

func BuildErrorResponse(err error, statusOverride int) (ErrorResponse, int) {
	meta, status := resolveMeta(err)
	if statusOverride > 0 {
		status = statusOverride
	}
	return ErrorResponse{
		Code:      meta.Code,
		Message:   meta.Message,
		Cause:     meta.Cause,
		Fix:       meta.Fix,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}, status
}

func WriteError(c *gin.Context, statusOverride int, err error) {
	resp, status := BuildErrorResponse(err, statusOverride)
	c.JSON(status, resp)
}

func WriteValidationError(c *gin.Context, detailMessage string) {
	resp := ErrorResponse{
		Code:      validationMeta.Code,
		Message:   validationMeta.Message + ": " + detailMessage,
		Cause:     validationMeta.Cause,
		Fix:       validationMeta.Fix,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
	c.JSON(validationMeta.Status, resp)
}
