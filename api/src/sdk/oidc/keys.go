package oidc

// AccessContextKey is the key where model.AccessClaims will be set in Request Contenxt
type accessContextKey struct{}

// JWTStaticSignature is the key used to sign JWT internally, we don't mind having a secret one for now
// since the jwt goes only inside our internal & private network
const JWTStaticSignature = "wedontmindaboutsigningfornow"
