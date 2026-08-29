package app

import (
    "net/http"
    "net/http/httptest"
    "os"
    "path/filepath"
    "strings"
    "testing"

    "github.com/gin-gonic/gin"
)

func TestSwaggerRoutesServeUIAndCanonicalContracts(t *testing.T) {
    router := gin.New()
    registerSwaggerRoutes(router, resolveSwaggerSpecDirectory())

    redirect := httptest.NewRecorder()
    router.ServeHTTP(redirect, httptest.NewRequest(http.MethodGet, swaggerPath, nil))
    if redirect.Code != http.StatusPermanentRedirect || redirect.Header().Get("Location") != swaggerIndexPath {
        t.Fatalf("Swagger redirect = %d %q", redirect.Code, redirect.Header().Get("Location"))
    }

    page := httptest.NewRecorder()
    router.ServeHTTP(page, httptest.NewRequest(http.MethodGet, swaggerIndexPath, nil))
    if page.Code != http.StatusOK {
        t.Fatalf("Swagger UI status = %d, want %d", page.Code, http.StatusOK)
    }
    if page.Header().Get("Content-Type") != swaggerIndexContentType {
        t.Fatalf("Swagger UI content type = %q, want %q", page.Header().Get("Content-Type"), swaggerIndexContentType)
    }
    for _, expected := range []string{"SwaggerUIBundle", swaggerStorefrontSpecPath, swaggerOperationsSpecPath, "requestInterceptor"} {
        if !strings.Contains(page.Body.String(), expected) {
            t.Fatalf("Swagger UI page does not contain %q", expected)
        }
    }

    for _, test := range []struct {
        name string
        path string
        file string
    }{
        {name: "storefront", path: swaggerStorefrontSpecPath, file: "storefront.yaml"},
        {name: "operations", path: swaggerOperationsSpecPath, file: "operations.yaml"},
    } {
        t.Run(test.name, func(t *testing.T) {
            response := httptest.NewRecorder()
            router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, test.path, nil))
            if response.Code != http.StatusOK {
                t.Fatalf("spec status = %d, want %d", response.Code, http.StatusOK)
            }
            if response.Header().Get("Content-Type") != swaggerSpecContentType {
                t.Fatalf("spec content type = %q, want %q", response.Header().Get("Content-Type"), swaggerSpecContentType)
            }
            want, err := os.ReadFile(filepath.Join(resolveSwaggerSpecDirectory(), test.file))
            if err != nil {
                t.Fatal(err)
            }
            if response.Body.String() != string(want) {
                t.Fatalf("served spec differs from canonical %s", test.file)
            }
        })
    }
}

func TestSwaggerSpecReturnsNotFoundWhenContractIsMissing(t *testing.T) {
    router := gin.New()
    registerSwaggerRoutes(router, t.TempDir())

    response := httptest.NewRecorder()
    router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, swaggerStorefrontSpecPath, nil))
    if response.Code != http.StatusNotFound {
        t.Fatalf("missing spec status = %d, want %d", response.Code, http.StatusNotFound)
    }
}
