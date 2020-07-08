package keyshare

import (
	"gitlab.misakey.dev/misakey/backend/api/src/modules/box/repositories/sqlboiler"
)

type KeyShare struct {
	InvitationHash string `json:"invitation_hash"`
	Share          string `json:"share"`
	BoxID          string `json:"box_id"`
	creatorID      string
}

func (src KeyShare) toSQLBoiler() *sqlboiler.KeyShare {
	dest := sqlboiler.KeyShare{
		InvitationHash: src.InvitationHash,
		Share:          src.Share,
		BoxID:          src.BoxID,
		CreatorID:      src.creatorID,
	}
	return &dest
}

func fromSQLBoiler(src *sqlboiler.KeyShare) KeyShare {
	dest := KeyShare{
		InvitationHash: src.InvitationHash,
		Share:          src.Share,
		BoxID:          src.BoxID,
		creatorID:      src.CreatorID,
	}
	return dest
}
