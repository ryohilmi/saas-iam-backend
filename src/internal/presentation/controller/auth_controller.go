package controller

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/url"
	"os"

	"iyaem/internal/providers"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

type AuthController struct {
	auth *providers.Authenticator
	db   *sql.DB
}

var (
	key []byte
	t   *jwt.Token
	s   string
)

func NewAuthController(auth *providers.Authenticator, db *sql.DB) *AuthController {
	return &AuthController{auth, db}
}

func (c *AuthController) Login(ctx *gin.Context) {
	state, err := generateRandomState()
	if err != nil {
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}

	log.Printf("Client req: %v", ctx.Request.Header.Get("Origin"))

	if ctx.Request.Header.Get("Origin") == "" {
		c.auth.Config.RedirectURL = "https://" + os.Getenv("AUTH0_CALLBACK_URL")
		authorizationURL := c.auth.AuthCodeURL(state)

		ctx.Redirect(http.StatusTemporaryRedirect, authorizationURL)
		return
	} else {
		c.auth.Config.RedirectURL = ctx.Request.Header.Get("origin") + "/callback"
	}

	authorizationURL := c.auth.AuthCodeURL(state) + "&app=" + ctx.Request.Header.Get("Origin")

	ctx.JSON(http.StatusTemporaryRedirect, gin.H{
		"url": authorizationURL,
	})

}

func (c *AuthController) Callback(ctx *gin.Context) {
	jsonData, err := ctx.GetRawData()
	if err != nil {
		log.Printf("Error: %v", err)
	}

	type Params struct {
		Code  string `json:"code"`
		State string `json:"state"`
	}

	var params Params
	err = json.Unmarshal(jsonData, &params)
	if err != nil {
		log.Printf("Error: %v", err)
	}

	token, err := c.auth.Exchange(ctx.Request.Context(), params.Code)
	if err != nil {
		ctx.String(http.StatusUnauthorized, "Failed to exchange an authorization code for a token.")
		return
	}

	_, rawIdToken, err := c.auth.VerifyIDToken(ctx.Request.Context(), token)
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
	var email string

	row := c.db.QueryRow(`select u.id, picture, email from public.user u
		LEFT JOIN user_identity ui
		on u.id = ui.user_id
		where ui.idp_id=$1`, claims["sub"])
	err = row.Scan(&user_id, &picture, &email)
	if err != nil {
		log.Printf("Error 4321: %v", err)
	}

	log.Println("User ID: ", user_id)

	key = []byte(os.Getenv("JWT_SECRET"))
	t = jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"iss":     "iam.sashore.com",
			"sub":     user_id,
			"picture": picture,
			"email":   email,
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

func (c *AuthController) Logout(ctx *gin.Context) {
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

	ctx.JSON(http.StatusTemporaryRedirect, gin.H{
		"url": logoutUrl.String(),
	})
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
