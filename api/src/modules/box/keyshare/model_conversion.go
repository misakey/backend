package keyshare

import (
	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/repositories/sqlboiler"
)

type KeyShare struct {
	InvitationHash string `json:"invitation_hash"`
	Share          string `json:"share"`
}

func (src KeyShare) toSqlBoiler() *sqlboiler.KeyShare {
	dest := sqlboiler.KeyShare{
		InvitationHash: src.InvitationHash,
		Share:          src.Share,
	}
	return &dest
}

func fromSqlBoiler(src *sqlboiler.KeyShare) KeyShare {
	dest := KeyShare{
		InvitationHash: src.InvitationHash,
		Share:          src.Share,
	}
	return dest
}
