package routes

import (
	"database/sql"
	"iyaem/internal/app/commands"
	"iyaem/internal/infrastructure/database/postgresql"
	"iyaem/internal/presentation/controller"
	"iyaem/internal/providers"
	"net/http"

	"github.com/gin-gonic/gin"
)

func NewRouter(auth *providers.Authenticator, db *sql.DB) *gin.Engine {
	r := gin.Default()

	orgRepo := postgresql.NewOrganizationRepository(db)
	userRepo := postgresql.NewUserRepository(db)
	memRepo := postgresql.NewMembershipRepository(db)

	createOrgCommand := commands.NewCreateOrganizationCommand(orgRepo)
	promoteUserCommand := commands.NewPromoteUserCommand(orgRepo, memRepo)
	demoteUserCommand := commands.NewDemoteUserCommand(orgRepo, memRepo)
	addOrgUserCommand := commands.NewAddOrganizationUserCommand(orgRepo, memRepo, userRepo)
	createUserCommand := commands.NewCreateUserCommand(userRepo)
	addRoleCommand := commands.NewAddRoleToMemberCommand(orgRepo, memRepo)
	removeRoleCommand := commands.NewRemoveRoleFromMemberCommand(orgRepo, memRepo)
	addGroupCommand := commands.NewAddGroupToMemberCommand(orgRepo)
	removeGroupCommand := commands.NewRemoveGroupFromMemberCommand(orgRepo)

	authController := controller.NewAuthController(auth, db)
	orgController := controller.NewOrganizationController(
		db,
		createOrgCommand,
		addOrgUserCommand,
		createUserCommand,
		addRoleCommand,
		postgresql.NewOrganizationQuery(db),
	)
	userController := controller.NewUserController(
		db,
		promoteUserCommand,
		demoteUserCommand,
		addRoleCommand,
		removeRoleCommand,
		addGroupCommand,
		removeGroupCommand,
		postgresql.NewUserQuery(db),
	)
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

	r.GET("/api/organization", providers.IsMachine(), orgController.GetAllOrganizations)

	r.Use(providers.IsAuthenticated)

	r.GET("/organization/:id", orgController.FindById)

	r.GET("/organization", orgController.GetAffiliatedOrganizations)
	r.POST("/organization", orgController.CreateOrganization)

	r.GET("/organization/statistics", orgController.Statistics)

	r.POST("/organization/add-user", orgController.AddUser)
	r.POST("/organization/create-user", orgController.CreateUser)
	r.GET("/organization/level", userController.UserLevel)
	r.GET("/organization/users", orgController.GetUsers)
	r.GET("/organization/recent-users", orgController.GetRecentUsers)

	r.GET("/user", userController.DoesUserExist)

	r.GET("/tenants", tenantController.TenantList)
	r.GET("/tenant/roles", isTenantValid, tenantController.Roles)
	r.GET("/tenant/groups", isTenantValid, tenantController.Groups)

	r.GET("/user/details", userController.UserDetails)
	r.GET("/user/roles", userController.UserRoles)
	r.GET("/user/groups", userController.UserGroups)

	r.GET("/role/users", roleController.UsersWithRole)
	r.GET("/group/users", groupController.UsersWithGroup)

	r.Use(isManager)

	r.DELETE("/organization/remove-user", orgController.RemoveUser)

	r.POST("/user/role", userController.AssignRole)
	r.DELETE("/user/role", userController.RemoveRole)

	r.POST("/user/group", isTenantValid, userController.AssignGroup)
	r.DELETE("/user/group", isTenantValid, userController.RemoveGroup)

	r.PUT("/user/promote", userController.Promote)
	r.PUT("/user/demote", userController.Demote)

	return r
}
