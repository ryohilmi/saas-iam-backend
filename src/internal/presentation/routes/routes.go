package routes

import (
	"database/sql"
	"iyaem/internal/infrastructure/database/postgresql"
	"iyaem/internal/presentation/controller"
	"iyaem/internal/providers"
	"net/http"

	"github.com/gin-gonic/gin"
)

func NewRouter(auth *providers.Authenticator, db *sql.DB) *gin.Engine {
	r := gin.Default()

	authController := controller.NewAuthController(auth, db)
	orgController := controller.NewOrganizationController(db, postgresql.NewOrganizationQuery(db))
	userController := controller.NewUserController(db)
	tenantController := controller.NewTenantController(db)
	roleController := controller.NewRoleController(db)
	groupController := controller.NewGroupController(db)

	r.Use(providers.CORSMiddleware())

	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "hello")
	})

	r.GET("/login", authController.Login)
	r.POST("/callback", authController.Callback)
	r.GET("/logout", authController.Logout)

	isManager := providers.IsOrganizationManager(db)
	isTenantValid := providers.IsTenantValid(db)

	r.GET("/organization", orgController.GetAffiliatedOrganizations)
	r.POST("/organization", orgController.CreateOrganization)

	r.GET("/organization/statistics", isManager, orgController.Statistics)

	r.POST("/organization/add-user", orgController.AddUser)
	r.POST("/organization/create-user", orgController.CreateUser)
	r.GET("/organization/level", orgController.UserLevel)
	r.GET("/organization/users", orgController.GetUsersInOrganization)
	r.GET("/organization/recent-users", orgController.GetRecentUsersInOrganization)
	r.DELETE("/organization/remove-user", orgController.RemoveUser)

	r.GET("/tenants", isManager, tenantController.TenantList)
	r.GET("/tenant/roles", isManager, isTenantValid, tenantController.Roles)
	r.GET("/tenant/groups", isManager, isTenantValid, tenantController.Groups)

	r.POST("/user/role", isManager, userController.AssignRole)
	r.DELETE("/user/role", isManager, userController.RemoveRole)

	r.POST("/user/group", isManager, isTenantValid, userController.AssignGroup)
	r.DELETE("/user/group", isManager, isTenantValid, userController.RemoveGroup)

	r.GET("/user", userController.DoesUserExist)
	r.GET("/user/details", isManager, userController.UserDetails)
	r.GET("/user/roles", isManager, userController.UserRoles)
	r.GET("/user/groups", isManager, userController.UserGroups)

	r.PUT("/user/promote", isManager, userController.Promote)
	r.PUT("/user/demote", isManager, userController.Demote)

	r.GET("/role/users", isManager, roleController.UsersWithRole)
	r.GET("/group/users", isManager, groupController.UsersWithGroup)

	return r
}
