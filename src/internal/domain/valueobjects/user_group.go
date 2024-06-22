package valueobjects

type UserGroup struct {
	membershipId MembershipId
	groupId      GroupId
	tenantId     TenantId
}

func NewUserGroup(membershipId MembershipId, roleId GroupId, tenantId TenantId) UserGroup {
	return UserGroup{membershipId, roleId, tenantId}
}

func (ur *UserGroup) MembershipId() MembershipId {
	return ur.membershipId
}

func (ur *UserGroup) GroupId() GroupId {
	return ur.groupId
}

func (ur *UserGroup) TenantId() TenantId {
	return ur.tenantId
}

func (ur *UserGroup) Equals(other UserGroup) bool {
	return ur.membershipId.Equals(other.membershipId) &&
		ur.groupId.Equals(other.groupId) &&
		ur.tenantId.Equals(other.tenantId)
}
