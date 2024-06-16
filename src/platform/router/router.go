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
	auth_token "iyaem/platform/token"
	"iyaem/web/app/controllers"
)

func New(auth *authenticator.Authenticator, db *sql.DB) *gin.Engine {
	r := gin.Default()

	gob.Register(map[string]interface{}{})

	orgController := controllers.NewOrganizationController(auth, db)
	userController := controllers.NewUserController(auth, db)
	tenantController := controllers.NewTenantController(auth, db)
	roleController := controllers.NewRoleController(db)

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

	r.GET("/organization/statistics", middleware.IsOrganizationManager(db), orgController.Statistics)

	r.POST("/organization/add-user", orgController.AddUser)
	r.POST("/organization/create-user", orgController.CreateUser)
	r.GET("/organization/level", orgController.UserLevel)
	r.GET("/organization/users", orgController.GetUsersInOrganization)
	r.GET("/organization/recent-users", orgController.GetRecentUsersInOrganization)

	r.GET("/tenants", middleware.IsOrganizationManager(db), tenantController.TenantList)
	r.GET("/tenant/roles", middleware.IsOrganizationManager(db), tenantController.Roles)
	r.GET("/tenant/groups", middleware.IsOrganizationManager(db), tenantController.Groups)

	r.POST("/user/role", middleware.IsOrganizationManager(db), userController.AssignRole)
	r.DELETE("/user/role", middleware.IsOrganizationManager(db), userController.RemoveRole)
	r.GET("/user", userController.DoesUserExist)

	r.PUT("/user/promote", middleware.IsOrganizationManager(db), userController.Promote)
	r.PUT("/user/demote", middleware.IsOrganizationManager(db), userController.Demote)

	r.GET("/role/users", middleware.IsOrganizationManager(db), roleController.UsersWithRole)

	r.GET("/get-token", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"token": auth_token.GetTokenSingleton().Token,
		})
	})
	return r
}
