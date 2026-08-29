package cors

import (
    "net/http"
    "strings"

    "github.com/Kangditya/persona-apps/apps/api/internal/platform/httpx"
    "github.com/gin-gonic/gin"
)

type Policy struct {
    Origins     map[string]struct{}
    Methods     []string
    Headers     []string
    Credentials bool
}

func Middleware(policy Policy) gin.HandlerFunc {
    methods := allowedSet(policy.Methods)
    headers := allowedSet(policy.Headers)
    return func(c *gin.Context) {
        origin := c.GetHeader("Origin")
        if origin == "" {
            c.Next()
            return
        }
        addVary(c, "Origin")
        if _, allowed := policy.Origins[origin]; !allowed {
            rejectPreflight(c)
            return
        }

        c.Header("Access-Control-Allow-Origin", origin)
        if policy.Credentials {
            c.Header("Access-Control-Allow-Credentials", "true")
        }
        c.Header("Access-Control-Expose-Headers", "X-Request-ID, Retry-After")
        if c.Request.Method != http.MethodOptions || c.GetHeader("Access-Control-Request-Method") == "" {
            c.Next()
            return
        }

        requestedMethod := strings.ToLower(strings.TrimSpace(c.GetHeader("Access-Control-Request-Method")))
        if _, allowed := methods[requestedMethod]; !allowed || !requestedHeadersAllowed(c.GetHeader("Access-Control-Request-Headers"), headers) {
            writeForbidden(c)
            return
        }
        addVary(c, "Access-Control-Request-Method")
        addVary(c, "Access-Control-Request-Headers")
        c.Header("Access-Control-Allow-Methods", strings.Join(policy.Methods, ", "))
        c.Header("Access-Control-Allow-Headers", strings.Join(policy.Headers, ", "))
        c.Status(http.StatusNoContent)
        c.Abort()
    }
}

func rejectPreflight(c *gin.Context) {
    if c.Request.Method == http.MethodOptions && c.GetHeader("Access-Control-Request-Method") != "" {
        writeForbidden(c)
        return
    }
    c.Next()
}

func writeForbidden(c *gin.Context) {
    httpx.WriteError(c, http.StatusForbidden, "forbidden", "request is not authorized", httpx.RequestID(c.Request.Context()), nil)
    c.Abort()
}

func requestedHeadersAllowed(value string, allowed map[string]struct{}) bool {
    if strings.TrimSpace(value) == "" {
        return true
    }
    for _, header := range strings.Split(value, ",") {
        if _, found := allowed[strings.ToLower(strings.TrimSpace(header))]; !found {
            return false
        }
    }
    return true
}

func allowedSet(values []string) map[string]struct{} {
    result := make(map[string]struct{}, len(values))
    for _, value := range values {
        result[strings.ToLower(value)] = struct{}{}
    }
    return result
}

func addVary(c *gin.Context, value string) {
    for _, current := range c.Writer.Header().Values("Vary") {
        for _, item := range strings.Split(current, ",") {
            if strings.EqualFold(strings.TrimSpace(item), value) {
                return
            }
        }
    }
    c.Writer.Header().Add("Vary", value)
}
