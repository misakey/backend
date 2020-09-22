package ajwt

// AccessContextKey is the key where model.AccessClaims will be set in Request Contenxt
type accessContextKey struct{}

// JWTStaticSignature is the key used to sign JWT internally, we don't mind having a secret one for now
// since the jwt goes only inside our internal & private network
const JWTStaticSignature = "wedontmindaboutsigningfornow"

type scope string

const (
	openIDScope      scope = "openid"
	userScope        scope = "user"
	applicationScope scope = "application"
	serviceScope     scope = "service"
	misadminScope    scope = "misadmin"
)

const (
	adminRoleLabel string = "admin"
	dpoRoleLabel   string = "dpo"
)

// scopePrefix defines possible prefix for special scopes:
// roles or application purposes
// it is uses by helpers to evaluate a scope type
type scopePrefix string

const (
	adminRolePrefix  scopePrefix = "rol.admin"
	dpoRolePrefix    scopePrefix = "rol.dpo"
	oldPurposePrefix scopePrefix = "pur:"
	purposePrefix    scopePrefix = "pur."
)
