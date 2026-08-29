package app

import (
    "errors"
    "net/http"
    "os"
    "path/filepath"
    "strings"

    "github.com/gin-gonic/gin"
)

const (
    swaggerPath                 = "/swagger"
    swaggerIndexPath            = "/swagger/index.html"
    swaggerStorefrontSpecPath   = "/swagger/openapi/storefront.yaml"
    swaggerOperationsSpecPath   = "/swagger/openapi/operations.yaml"
    swaggerSpecContentType      = "application/yaml; charset=utf-8"
    swaggerIndexContentType     = "text/html; charset=utf-8"
    swaggerSpecDirectoryEnvName = "OPENAPI_DIR"
)

const swaggerUIIndexHTML = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Qurban Commerce API</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css">
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-standalone-preset.js"></script>
  <script>
    window.onload = function () {
      window.ui = SwaggerUIBundle({
        urls: [
          { name: "Storefront API", url: "/swagger/openapi/storefront.yaml" },
          { name: "Operations API", url: "/swagger/openapi/operations.yaml" }
        ],
        dom_id: "#swagger-ui",
        deepLinking: true,
        presets: [
          SwaggerUIBundle.presets.apis,
          SwaggerUIStandalonePreset
        ],
        layout: "StandaloneLayout",
        validatorUrl: "none",
        requestInterceptor: function (request) {
          const placeholderOrigin = "https://api.example.invalid";
          const target = new URL(request.url, window.location.origin);
          if (target.origin === placeholderOrigin) {
            request.url = window.location.origin + target.pathname + target.search;
          }
          return request;
        }
      });
    };
  </script>
</body>
</html>`

func registerSwaggerRoutes(router *gin.Engine, directory string) {
    router.GET(swaggerPath, redirectToSwaggerUI)
    router.GET(swaggerPath+"/", redirectToSwaggerUI)
    router.GET(swaggerIndexPath, serveSwaggerUI)
    router.GET(swaggerStorefrontSpecPath, serveSwaggerSpec(directory, "storefront.yaml"))
    router.GET(swaggerOperationsSpecPath, serveSwaggerSpec(directory, "operations.yaml"))
}

func redirectToSwaggerUI(c *gin.Context) {
    c.Redirect(http.StatusPermanentRedirect, swaggerIndexPath)
}

func serveSwaggerUI(c *gin.Context) {
    setSwaggerHeaders(c)
    c.Data(http.StatusOK, swaggerIndexContentType, []byte(swaggerUIIndexHTML))
}

func serveSwaggerSpec(directory, filename string) gin.HandlerFunc {
    return func(c *gin.Context) {
        body, err := os.ReadFile(filepath.Join(directory, filename))
        if errors.Is(err, os.ErrNotExist) {
            http.NotFound(c.Writer, c.Request)
            return
        }
        if err != nil {
            http.Error(c.Writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
            return
        }
        setSwaggerHeaders(c)
        c.Data(http.StatusOK, swaggerSpecContentType, body)
    }
}

func setSwaggerHeaders(c *gin.Context) {
    c.Header("Cache-Control", "no-store")
    c.Header("Content-Security-Policy", "default-src 'self'; connect-src 'self'; font-src 'self' https://unpkg.com; img-src 'self' data: https://validator.swagger.io; script-src 'self' 'unsafe-inline' https://unpkg.com; style-src 'self' 'unsafe-inline' https://unpkg.com")
    c.Header("Referrer-Policy", "no-referrer")
    c.Header("X-Content-Type-Options", "nosniff")
}

func resolveSwaggerSpecDirectory() string {
    if configured := strings.TrimSpace(os.Getenv(swaggerSpecDirectoryEnvName)); configured != "" {
        return configured
    }
    for _, candidate := range []string{
        "contracts/openapi",
        "../../contracts/openapi",
        "../../../../contracts/openapi",
    } {
        if hasSwaggerSpecs(candidate) {
            return candidate
        }
    }
    return "contracts/openapi"
}

func hasSwaggerSpecs(directory string) bool {
    for _, filename := range []string{"storefront.yaml", "operations.yaml"} {
        info, err := os.Stat(filepath.Join(directory, filename))
        if err != nil || info.IsDir() {
            return false
        }
    }
    return true
}
