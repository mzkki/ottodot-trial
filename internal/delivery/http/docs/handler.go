package docs

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const scalarHTML = `<!doctype html>
<html>
  <head>
    <title>OttoDot API Documentation</title>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
  </head>
  <body>
    <script
      id="api-reference"
      data-url="/api/v1/docs/openapi.json"
      data-configuration='{
        "showSidebar": true,
        "hideDownloadButton": false,
        "hideClientButton": false,
        "withDefaultFonts": true
      }'
    ></script>
    <script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"></script>
  </body>
</html>`

func DocsHandler(c *gin.Context) {
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, scalarHTML)
}
