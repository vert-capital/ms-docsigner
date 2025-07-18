package handlers

import (
	"net/http"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
)

// func AllRoutesTextHandler(c *gin.Context) {

// 	routes := c.Routes()
// 	response := ""
// 	for _, route := range routes {
// 		response += fmt.Sprintf("%s %s\n", route.Method, route.Path)
// 	}
// 	c.String(200, response)

// }

func HomeHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Hello World",
	})

	c.Request.Response.Header.Add("Content-Type", "application/json")
}

func XmlHandler(c *gin.Context) {
	c.XML(http.StatusOK, gin.H{
		"message": "Hello World",
	})

	c.Request.Response.Header.Add("Content-Type", "application/xml")
}

func TextHandler(c *gin.Context) {
	c.String(http.StatusOK, "Hello World")

	c.Request.Response.Header.Add("Content-Type", "text/plain")
}

func YamlHandler(c *gin.Context) {
	c.YAML(http.StatusOK, gin.H{
		"message": "Hello World",
	})

	c.Request.Response.Header.Add("Content-Type", "application/x-yaml")
}

func ProtobufHandler(c *gin.Context) {
	c.ProtoBuf(http.StatusOK, gin.H{
		"message": "Hello World",
	})

	c.Request.Response.Header.Add("Content-Type", "application/x-protobuf")
}

func ServerSideEventsHandler(c *gin.Context) {
	c.SSEvent("message", gin.H{
		"message": "Hello World",
	})

	c.Request.Response.Header.Add("Content-Type", "text/event-stream")
}

func SecretHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "The secret ingredient is ...",
	})

	c.Request.Response.Header.Add("Content-Type", "application/json")
}

func MountSamplesHandlers(r *gin.Engine) {

	group := r.Group("/samples")

	group.GET("/", HomeHandler)
	group.GET("/xml", XmlHandler)
	group.GET("/text", TextHandler)
	group.GET("/yaml", YamlHandler)
	group.GET("/protobuf", ProtobufHandler)
	group.GET("/server-side-events", ServerSideEventsHandler)
	group.GET("/secret", SecretHandler)
	// group.GET("/routes", AllRoutesTextHandler)

	r.GET("/routes", func(ctx *gin.Context) {

		type Route struct {
			Method  string
			Path    string
			Handler string
		}

		if gin.Mode() == gin.DebugMode {

			routes := make([]Route, 0)
			var response string

			for _, route := range r.Routes() {
				routes = append(routes, Route{
					Method:  route.Method,
					Path:    route.Path,
					Handler: route.Handler,
				})
			}

			sort.Slice(routes, func(i, j int) bool {
				return routes[i].Path < routes[j].Path
			})

			for _, route := range routes {
				method := route.Method

				if len(method) < 7 {
					method = method + strings.Repeat(" ", 7-len(method))
				}

				response += method + " " + route.Path + "\n"
			}

			ctx.String(http.StatusOK, "%v", response)

		}
	})
}
