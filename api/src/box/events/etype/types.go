package etype

// Event types constants
const (
	Create         = "create"
	Statelifecycle = "state.lifecycle"
	StateKeyShare  = "state.key_share"
	Memberjoin     = "member.join"
	Memberleave    = "member.leave"
	Memberkick     = "member.kick"
	Msgtext        = "msg.text"
	Msgfile        = "msg.file"
	Msgedit        = "msg.edit"
	Msgdelete      = "msg.delete"
	Accessadd      = "access.add"
	Accessrm       = "access.rm"
)

// MembersCanSee returns all event types that can be seen by members
func MembersCanSee() []string {
	return []string{
		Create,
		Statelifecycle,
		StateKeyShare,

		Msgtext,
		Msgfile,

		Memberjoin,
		Memberleave,
		Memberkick,
	}
}

// RequireToBuild returns all events required to build the box
func RequireToBuild() []string {
	return []string{
		Create,
		Statelifecycle,
		StateKeyShare,
	}
}

// RequiresContent returns all events needing a content
func RequiresContent(eType string) bool {
	switch eType {
	case
		Create,
		Statelifecycle,

		Msgtext,
		Msgfile,
		Msgedit,

		Accessadd:
		return true
	}
	return false
}
