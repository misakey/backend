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

func ContentIsRequired(eType string) bool {
	return eType == Ecreate ||
		eType == Estatelifecycle ||
		eType == Emsgtext ||
		eType == Emsgfile ||
		eType == Emsgedit ||
		eType == Emsgdelete ||
		eType == Eaccessadd
}
