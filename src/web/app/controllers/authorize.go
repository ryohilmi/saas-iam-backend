package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"iyaem/platform/authenticator"
)

func AuthorizeHandler(auth *authenticator.Authenticator) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authURL := ctx.Query("authorizationURL")
		if authURL == "" {
			ctx.String(http.StatusBadRequest, "Invalid authorization URL.")
			return
		}

		ctx.Redirect(http.StatusTemporaryRedirect, authURL)
	}
}
