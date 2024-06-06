package controllers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"golang.org/x/oauth2"

	"iyaem/platform/authenticator"
)

var (
	key []byte
	t   *jwt.Token
	s   string
)

type Params struct {
	Code  string `json:"code"`
	State string `json:"state"`
}

func CallbackPostHandler(auth *authenticator.Authenticator, db *sql.DB) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		jsonData, err := ctx.GetRawData()
		if err != nil {
			log.Printf("Error: %v", err)
		}

		var params Params
		err = json.Unmarshal(jsonData, &params)
		if err != nil {
			log.Printf("Error: %v", err)
		}

		token, err := auth.Exchange(ctx.Request.Context(), params.Code)
		if err != nil {
			ctx.String(http.StatusUnauthorized, "Failed to exchange an authorization code for a token.")
			return
		}

		_, rawIdToken, err := auth.VerifyIDToken(ctx.Request.Context(), token)
		if err != nil {
			ctx.String(http.StatusInternalServerError, "Failed to verify ID Token.")
			return
		}

		claims, err := DecodeJWT(rawIdToken)
		if err != nil {
			log.Print(err)
		}

		var picture string
		var user_id string

		row := db.QueryRow(`select id, picture from public.user 
							LEFT JOIN public.user_identity
							on public.user.id = public.user_identity.user_id
							where idp_id=$1`,
			claims["sub"])
		err = row.Scan(&user_id, &picture)
		if err != nil {
			log.Print(err)
		}

		log.Println("User ID: ", user_id)

		key = []byte(os.Getenv("JWT_SECRET"))
		t = jwt.NewWithClaims(jwt.SigningMethodHS256,
			jwt.MapClaims{
				"iss":     "iam.sashore.com",
				"sub":     user_id,
				"picture": picture,
				"exp":     claims["exp"],
				"iat":     claims["iat"],
				"name":    claims["name"],
			})
		s, err = t.SignedString(key)
		if err != nil {
			log.Print(err)
		}

		ctx.JSON(http.StatusOK, gin.H{"token": s})
	}
}

func CallbackGetHandler(auth *authenticator.Authenticator) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		session := sessions.Default(ctx)

		log.Printf("Session: %v", session)

		if ctx.Query("state") != session.Get("state") {
			log.Printf("Invalid state parameter: %s != %s\n", ctx.Query("state"), session.Get("state"))
			ctx.String(http.StatusBadRequest, "Invalid state parameter."+ctx.Query("state")+" "+session.Get("state").(string))
			// return
		}

		// Exchange an authorization code for a token.
		CustomParam := oauth2.SetAuthURLParam("hello", "hello world")

		token, err := auth.Exchange(ctx.Request.Context(), ctx.Query("code"), CustomParam)
		if err != nil {
			ctx.String(http.StatusUnauthorized, "Failed to exchange an authorization code for a token.")
			return
		}

		// log.Printf("Token: %s", token.AccessToken)

		idToken, rawIdToken, err := auth.VerifyIDToken(ctx.Request.Context(), token)
		if err != nil {
			ctx.String(http.StatusInternalServerError, "Failed to verify ID Token.")
			return
		}

		var profile map[string]interface{}
		if err := idToken.Claims(&profile); err != nil {
			ctx.String(http.StatusInternalServerError, err.Error())
			return
		}

		log.Printf("Profile: %v", idToken.AccessTokenHash)

		session.Set("access_token", token.AccessToken)
		session.Set("profile", profile)
		if err := session.Save(); err != nil {
			ctx.String(http.StatusInternalServerError, err.Error())
			return
		}

		//Redirect to logged in page.
		ctx.SetCookie("id_token", rawIdToken, 3600, "/", "localhost", false, true)
		ctx.Redirect(http.StatusTemporaryRedirect, "/user")
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
