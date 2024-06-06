package controllers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

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

		var tenant_id string
		var organization_id string
		var org_name string
		var user_id string
		var user_roles []string

		origin := ctx.Request.Header.Get("Origin")
		host, _ := url.Parse(origin)
		subdomain := strings.Split(host.Host, ".")

		rows, err := db.Query(`	select t.id as tenant_id, o.id as organization_id, o.name as org_name from tenant t
								left join application a on a.id  = t.app_id
								left join organization o on o.id = t.org_id
								where host=$2 and o.subdomain=$1 limit 1`, subdomain[0], subdomain[1])

		if err != nil {
			log.Printf("err: %v", err)
		}

		for rows.Next() {
			err = rows.Scan(&tenant_id, &organization_id, &org_name)
			if err != nil {
				log.Print(err)
			}
		}

		row := db.QueryRow(`select id from public.user where idp_id=$1`, claims["sub"])
		err = row.Scan(&user_id)
		if err != nil {
			log.Print(err)
		}

		log.Println("User ID: ", user_id)

		rows, err = db.Query(`	select r.name as role
								from role r
								left join user_role ur on r.id = ur.role_id 
								left join user_organization uo on uo.id = ur.user_org_id 
								left join "user" u on u.id = uo.user_id 
								left join organization o on o.id = uo.organization_id 
								left join tenant t on t.id = r.tenant_id 
								where u.id = $1
								and t.id = $2
								and o.id = $3
							`, user_id, tenant_id, organization_id)
		if err != nil {
			log.Print(err)
		}

		log.Println("User Roles: ", user_roles)

		var role string
		for rows.Next() {
			err = rows.Scan(&role)
			if err != nil {
				log.Print(err)
			}

			user_roles = append(user_roles, role)
		}

		key = []byte(os.Getenv("JWT_SECRET"))
		t = jwt.NewWithClaims(jwt.SigningMethodHS256,
			jwt.MapClaims{
				"iss":       "iam.sashore.com",
				"sub":       user_id,
				"exp":       claims["exp"],
				"iat":       claims["iat"],
				"name":      claims["name"],
				"tenant_id": tenant_id,
				"org_id":    organization_id,
				"org_name":  org_name,
				"roles":     user_roles,
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
