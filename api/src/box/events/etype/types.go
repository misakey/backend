package etype

// Event types constants
const (
	// event types
	Accessadd       = "access.add"
	Accessrm        = "access.rm"
	Create          = "create"
	Memberjoin      = "member.join"
	Memberleave     = "member.leave"
	Memberkick      = "member.kick"
	Msgtext         = "msg.text"
	Msgfile         = "msg.file"
	Msgedit         = "msg.edit"
	Msgdelete       = "msg.delete"
	Stateaccessmode = "state.access_mode"
	Statekeyshare   = "state.key_share"

	// events batch type
	BatchAccesses = "accesses"
)

// MembersCanSee contains all event types that can be seen by members
var MembersCanSee = []string{Create, Memberjoin, Memberleave, Memberkick, Msgtext, Msgfile, Statekeyshare}

// RequireToBuild contains all event types required to build the box
var RequireToBuild = []string{Create, Stateaccessmode, Statekeyshare}

// RequiresContent returns all events needing a content
func RequiresContent(eType string) bool {
	switch eType {
	case Accessadd, Create, Msgtext, Msgfile, Msgedit, Stateaccessmode:
		return true
	}
	return false
}
