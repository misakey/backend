package events

const (
	Ecreate         = "create"
	Estatelifecycle = "state.lifecycle"
	Ememberjoin     = "member.join"
	Ememberleave    = "member.leave"
	Emsgtext        = "msg.text"
	Emsgfile        = "msg.file"
	Emsgedit        = "msg.edit"
	Emsgdelete      = "msg.delete"
	Eaccessadd      = "access.add"
	Eaccessrm       = "access.rm"
)

// Return all event types that can be seen by members
func memberReadTypes() []string {
	return []string{Ecreate, Estatelifecycle, Emsgtext, Emsgfile, Emsgedit, Emsgdelete, Ememberjoin, Ememberleave}
}

func ContentIsRequired(eType string) bool {
	return eType == Ecreate ||
		eType == Estatelifecycle ||
		eType == Emsgtext ||
		eType == Emsgfile ||
		eType == Emsgedit ||
		eType == Emsgdelete ||
		eType == Eaccessadd
}
