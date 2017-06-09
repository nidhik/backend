package db

const PUBLIC_KEY = "*"
const READ_PERM = "_rperm"
const WRITE_PERM = "_wperm"

type Permission struct {
	Read  bool `json:"r" bson:"r"`
	Write bool `json:"w" bson:"w"`
}

type ACL struct {
	ACL         map[string]Permission `json:"_acl" bson:"_acl,omitempty"`
	ReadAccess  []string              `json:"_rperm" bson:"_rperm,omitempty"`
	WriteAccess []string              `json:"_wperm" bson:"_wperm,omitempty"`
}

func NewACL() *ACL {
	return &ACL{ACL: make(map[string]Permission)}
}

func (acl ACL) IsZero() bool {
	return len(acl.ACL) == 0 && len(acl.ReadAccess) == 0 && len(acl.WriteAccess) == 0
}

func (acl *ACL) AddRead(name string) {
	perm := acl.ACL[name]
	perm.Read = true
	acl.ACL[name] = perm
	acl.ReadAccess = append(acl.ReadAccess, name)
}

func (acl *ACL) AddWrite(name string) {

	perm := acl.ACL[name]
	perm.Write = true
	acl.ACL[name] = perm
	acl.WriteAccess = append(acl.WriteAccess, name)
}

func (acl *ACL) SetPublicRead() {
	acl.AddRead(PUBLIC_KEY)
}

func (acl *ACL) SetPublicWrite() {
	acl.AddWrite(PUBLIC_KEY)
}

func (acl *ACL) SetPublicReadWrite() {
	acl.SetPublicRead()
	acl.SetPublicWrite()
}

func (acl *ACL) CanRead(name string) bool {
	return acl.ACL[name].Read || acl.ACL[PUBLIC_KEY].Read
}

func (acl *ACL) CanWrite(name string) bool {
	return acl.ACL[name].Write || acl.ACL[PUBLIC_KEY].Write
}
