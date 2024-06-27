package providers

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/golang-jwt/jwt"
)

// IsAuthenticated is a middleware that checks if
// the user has already been authenticated previously.
func IsAuthenticated(ctx *gin.Context) {
	authorizationHeader := ctx.Request.Header.Get("Authorization")

	if authorizationHeader == "" || len("Bearer ") >= len(authorizationHeader) {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"message": "Unauthorized",
		})
		return
	}

	token := authorizationHeader[len("Bearer "):]
	claims, err := decodeJWT(token)
	if err != nil {
		log.Printf("Error 9876: %v", err)
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"message": "Invalid Token",
		})
		return
	}
	ctx.Set("token", claims)
	ctx.Next()
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Content-Type", "application/json")
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Max, Set-Cookie")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(200)
		} else {
			c.Next()
		}
	}
}

func IsOrganizationManager(db *sql.DB) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		type OrgParams struct {
			OrganizationId string `json:"organization_id" form:"organization_id" binding:"required"`
		}

		var params OrgParams

		if ctx.Request.Method == "GET" {
			err := ctx.ShouldBindQuery(&params)
			if err != nil {
				log.Printf("Error 6969: %v", err)
				ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
					"message": "Invalid request body",
				})
				return
			}
		} else {
			err := ctx.ShouldBindBodyWith(&params, binding.JSON)
			if err != nil {
				log.Printf("Error 6970: %v", err)
				ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
					"message": "Invalid request body",
				})
				return
			}
		}

		authorizationHeader := ctx.Request.Header.Get("Authorization")

		if authorizationHeader == "" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"message": "Unauthorized",
			})
			return
		}

		token := authorizationHeader[len("Bearer "):]
		claims, err := DecodeJWT(token)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"message": "Unauthorized",
			})
			return
		}

		var level string

		err = db.QueryRow("SELECT level FROM user_organization uo left join public.user u on uo.user_id = u.id  WHERE u.email=$1 and uo.organization_id=$2;", claims["email"], params.OrganizationId).Scan(&level)
		if err != nil {
			log.Printf("Error 6901: %v", err)
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": "Internal Server Error",
			})
			return
		}

		log.Printf("Level: %v", level)

		if level != "owner" && level != "manager" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"message": "Unauthorized, only owner or manager can perform this action.",
			})
			return
		}

		ctx.Next()
	}
}

func IsTenantValid(db *sql.DB) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		type OrgParams struct {
			OrganizationId string `json:"organization_id" form:"organization_id" binding:"required"`
			TenantId       string `json:"tenant_id" form:"tenant_id" binding:"required"`
		}

		var params OrgParams

		if ctx.Request.Method == "GET" {
			err := ctx.ShouldBindQuery(&params)
			if err != nil {
				log.Printf("Error 6969: %v", err)
				ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
					"message": "Invalid request body",
				})
				return
			}
		} else {
			err := ctx.ShouldBindBodyWith(&params, binding.JSON)
			if err != nil {
				log.Printf("Error 6970: %v", err)
				ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
					"message": "Invalid request body",
				})
				return
			}
		}

		// Check if user exists in organization
		row := db.QueryRow(`
			SELECT EXISTS(
			SELECT 1 FROM tenant t 
			left join organization o on t.org_id = o.id
			where t.id=$1 and o.id=$2);`, params.TenantId, params.OrganizationId)

		var exists bool
		err := row.Scan(&exists)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": "Internal Server Error",
			})
			return
		}

		if !exists {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"status":  "error",
				"message": "Tenant not found",
			})
			return
		}

		ctx.Next()
	}
}

func SetSubDomain(auth *Authenticator) gin.HandlerFunc {
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

func DecodeJWT(token string) (map[string]interface{}, error) {
	tokenInstance, _, err := new(jwt.Parser).ParseUnverified(token, jwt.MapClaims{})
	if err != nil {
		return nil, err
	}

	claims, ok := tokenInstance.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid claims")
	}

	return claims, nil
}

func IsMachine() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		authorizationHeader := ctx.Request.Header.Get("Authorization")

		token := authorizationHeader[len("Bearer "):]

		log.Printf("API Token: %v\n", token)

		if token != "iam" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"message": "Unauthorized",
			})

			return
		}

		ctx.Next()
	}
}
