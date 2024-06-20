package controller

import "github.com/gin-gonic/gin"

func GetBearerToken(ctx *gin.Context) string {
	authorizationHeader := ctx.Request.Header.Get("Authorization")

	token := authorizationHeader[len("Bearer "):]

	return token
}
