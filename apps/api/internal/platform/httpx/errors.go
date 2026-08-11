package httpx

import (
    "encoding/json"
    "net/http"

    "github.com/gin-gonic/gin"
)

type errorEnvelope struct {
    Error errorBody `json:"error"`
}

type errorBody struct {
    Code      string         `json:"code"`
    Message   string         `json:"message"`
    RequestID string         `json:"request_id"`
    Details   map[string]any `json:"details"`
}

func WriteError(c *gin.Context, status int, code, message, requestID string, details map[string]any) {
    if details == nil {
        details = map[string]any{}
    }
    if status >= http.StatusInternalServerError {
        message = "internal server error"
        details = map[string]any{}
    }

    c.Header("Content-Type", "application/json")
    c.Status(status)
    _ = json.NewEncoder(c.Writer).Encode(errorEnvelope{Error: errorBody{
        Code: code, Message: message, RequestID: requestID, Details: details,
    }})
}
