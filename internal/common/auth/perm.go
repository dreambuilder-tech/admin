package auth

type PermCode string

const (
	MemberList       PermCode = "member-list"
	MemberListExport PermCode = "member-list-export"
)

var AllRouterPerms = make(map[string]PermCode)

func IsValidPerm(perm PermCode) bool {
	for _, p := range AllRouterPerms {
		if p == perm {
			return true
		}
	}
	return false
}

func GetAllPerms() map[string]PermCode {
	return AllRouterPerms
}
