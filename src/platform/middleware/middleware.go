// Save this file in ./platform/middleware/isAuthenticated.go

package middleware

import (
	"iyaem/platform/authenticator"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func IsAuthenticated(ctx *gin.Context) {
	if sessions.Default(ctx).Get("profile") == nil {
		ctx.Redirect(http.StatusSeeOther, "/")
	} else {
		ctx.Next()
	}
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Content-Type", "application/json")
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Max")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(200)
		} else {
			c.Next()
		}
	}
}

func SetSubDomain(auth *authenticator.Authenticator) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		host, err := url.Parse(ctx.Request.Host)
		if err != nil {
			log.Fatal(err)
		}

		log.Printf("Host: %v", host)

		hostUrl := strings.TrimSpace(host.String())
		//Figure out if a subdomain exists in the host given.
		hostParts := strings.Split(hostUrl, ".")

		log.Printf("Host Parts: %v", hostParts)

		if len(hostParts) > 1 {
			auth.SetSubDomain(hostParts[0])
		} else {
			auth.ResetSubDomain()
		}

		ctx.Next()
	}
}
