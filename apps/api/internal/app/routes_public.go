package app

import "github.com/gin-gonic/gin"

func registerPublicRoutes(router *gin.Engine) {
    // The public contract has no implemented handlers yet; keep its group
    // separate so the first public module can register against this boundary.
    _ = router.Group(publicAPIPrefix)
}
