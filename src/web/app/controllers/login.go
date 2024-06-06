package controllers

import (
	"crypto/rand"
	"encoding/base64"
	"log"
	"net/http"
	"os"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"

	"iyaem/platform/authenticator"
)

func LoginHandler(auth *authenticator.Authenticator) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		state, err := generateRandomState()
		if err != nil {
			ctx.String(http.StatusInternalServerError, err.Error())
			return
		}

		session := sessions.Default(ctx)
		session.Set("state", state)

		log.Printf("State: %s", state)

		if err := session.Save(); err != nil {
			ctx.String(http.StatusInternalServerError, err.Error())
			return
		}

		log.Printf("Client req: %v", ctx.Request.Header.Get("Origin"))

		if ctx.Request.Header.Get("Origin") == "" {
			auth.Config.RedirectURL = "https://" + os.Getenv("AUTH0_CALLBACK_URL")
			authorizationURL := auth.AuthCodeURL(state) + "&app=" + ctx.Request.Host

			ctx.Redirect(http.StatusTemporaryRedirect, authorizationURL)
			return
		} else {
			auth.Config.RedirectURL = ctx.Request.Header.Get("Origin") + "/callback"
		}

		authorizationURL := auth.AuthCodeURL(state) + "&app=" + ctx.Request.Header.Get("Origin")

		ctx.JSON(http.StatusOK, gin.H{
			"url": authorizationURL,
		})
	}
}

func generateRandomState() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	state := base64.StdEncoding.EncodeToString(b)

	return state, nil
}
