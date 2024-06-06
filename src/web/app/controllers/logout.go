package controllers

import (
	"net/http"
	"net/url"
	"os"

	"github.com/gin-gonic/gin"
)

func LogoutHandler(ctx *gin.Context) {
	logoutUrl, err := url.Parse("https://" + os.Getenv("AUTH0_DOMAIN") + "/v2/logout")
	if err != nil {
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}

	// scheme := "https"
	// if ctx.Request.TLS != nil {
	// 	scheme = "https"
	// }

	returnTo, err := url.Parse(ctx.Request.Header.Get("Origin"))
	if err != nil {
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}

	if returnTo.String() == "" {
		returnTo, err = url.Parse("https://" + os.Getenv("AUTH0_LOGOUT_URL"))
		if err != nil {
			ctx.String(http.StatusInternalServerError, err.Error())
			return
		}
	}

	parameters := url.Values{}
	parameters.Add("returnTo", returnTo.String())
	parameters.Add("client_id", os.Getenv("AUTH0_CLIENT_ID"))
	logoutUrl.RawQuery = parameters.Encode()

	if ctx.Request.Header.Get("Origin") == "" {
		ctx.Redirect(http.StatusTemporaryRedirect, logoutUrl.String())
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"url": logoutUrl.String(),
	})
}
