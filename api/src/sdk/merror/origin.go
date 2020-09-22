package merror

// Origin is an information about where the error does come from.
// Once set, the Origin of an error should not be spoiled.
type Origin string

const (
	OriACR      Origin = "acr"      // the error comes from the authorization token acr
	OriBody     Origin = "body"     // the error comes from body parameter
	OriHeaders  Origin = "headers"  // the error comes from headers
	OriInternal Origin = "internal" // the error comes from internal logic
	OriQuery    Origin = "query"    // the error comes from query parameters
	OriPath     Origin = "path"     // the error comes from path parameters

	OriNotDefined Origin = "not_defined" // the error has no origin defined yet
)
