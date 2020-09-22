package bubble

// Needle is a component able to transform a given error into a better one.
// The analysis of the input error is up to the implementation, as the final returned error is.
// The implementation shall return nil if the input error is unknown.
type Needle interface {
	// Explode is a function testing all known needles in the giving box
	// it takes in argument the error that bubbled up.
	Explode(err error) error
}

var needles []Needle
var lock bool

// AddNeedle to current list of needles - if not locked
func AddNeedle(n Needle) {
	if lock {
		return
	}
	needles = append(needles, n)
}

// Lock the possibility to add needles to the list
// should be used at the end of all needles instantiation
func Lock() {
	lock = true
}

// Explode iterates on all registered needles
// if the returned error is not nil, it means the needle found an error he could handle
// since we iterate with a simple loop on the needles, first appended needles are the first to be used
func Explode(bubble error) error {
	for _, needle := range needles {
		if detailedBubble := needle.Explode(bubble); detailedBubble != nil {
			return detailedBubble
		}
	}
	return bubble
}
