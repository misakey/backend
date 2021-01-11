package merr

// Origin is an information about where the error does come from.
// Once set, the Origin of an error should not be spoiled.
type Origin string

// origin constants
const (
	OriACR     Origin = "acr"     // the error comes from the authorization token acr
	OriBody    Origin = "body"    // the error comes from body parameter
	OriHeaders Origin = "headers" // the error comes from headers
	OriQuery   Origin = "query"   // the error comes from query parameters
	OriPath    Origin = "path"    // the error comes from path parameters
	OriCookies Origin = "cookies" // the error comes from cookies

	OriNotDefined Origin = "not_defined" // the error has no origin defined yet
)
