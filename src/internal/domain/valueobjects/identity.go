package valueobjects

type Identity struct {
	idpId  string
	userId UserId
}

func NewIdentity(idpId string, userId UserId) Identity {
	return Identity{idpId, userId}
}

func (i *Identity) IdpId() string {
	return i.idpId
}

func (i *Identity) UserId() UserId {
	return i.userId
}
