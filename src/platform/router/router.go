package router

import (
	"database/sql"
	"encoding/gob"
	"net/http"

	_ "github.com/lib/pq"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"

	"iyaem/platform/authenticator"
	"iyaem/platform/middleware"
	"iyaem/web/app/controllers"
)

func New(auth *authenticator.Authenticator, db *sql.DB) *gin.Engine {
	r := gin.Default()

	gob.Register(map[string]interface{}{})

	orgController := controllers.NewOrganizationController(auth, db)

	store := cookie.NewStore([]byte("secret"))
	r.Use(sessions.Sessions("auth-session", store))

	r.Static("/public", "web/static")
	r.LoadHTMLGlob("web/template/*")

	r.GET("/", func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "home.html", nil)
	})

	r.Use(middleware.CORSMiddleware())

	r.GET("/login", controllers.LoginHandler(auth))
	r.GET("/authorize", controllers.AuthorizeHandler(auth))

	r.GET("/callback", controllers.CallbackGetHandler(auth))
	r.POST("/callback", controllers.CallbackPostHandler(auth, db))
	r.GET("/logout", controllers.LogoutHandler)

	r.GET("/organization", orgController.GetAffiliatedOrganizations)
	r.POST("/organization", orgController.CreateOrganization)

	r.GET("/organization/user", orgController.GetUsersInOrganization)
	return r
}
