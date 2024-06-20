package valueobjects

type UserRole struct {
	membershipId MembershipId
	roleId       RoleId
	tenantId     TenantId
}

func NewUserRole(membershipId MembershipId, roleId RoleId, tenantId TenantId) UserRole {
	return UserRole{membershipId, roleId, tenantId}
}

func (ur *UserRole) MembershipId() MembershipId {
	return ur.membershipId
}

func (ur *UserRole) RoleId() RoleId {
	return ur.roleId
}

func (ur *UserRole) TenantId() TenantId {
	return ur.tenantId
}
