package httpx

import (
    "context"
    "database/sql"
    "database/sql/driver"
    "encoding/json"
    "errors"
    "net"
    "net/http"
    "strings"

    "github.com/gin-gonic/gin"
    "github.com/jackc/pgx/v5/pgconn"
)

var ErrDependencyUnavailable = errors.New("dependency unavailable")

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

func IsDependencyUnavailable(err error) bool {
    if errors.Is(err, ErrDependencyUnavailable) || errors.Is(err, context.DeadlineExceeded) || errors.Is(err, sql.ErrConnDone) || errors.Is(err, driver.ErrBadConn) {
        return true
    }
    var postgresError *pgconn.PgError
    if errors.As(err, &postgresError) {
        return strings.HasPrefix(postgresError.Code, "08") || postgresError.Code == "57P01" || postgresError.Code == "57P02" || postgresError.Code == "57P03"
    }
    var networkError net.Error
    return errors.As(err, &networkError)
}
