package keyshares

import (
	"github.com/volatiletech/null/v8"
	"gitlab.misakey.dev/misakey/backend/api/src/box/repositories/sqlboiler"
)

type BoxKeyShare struct {
	OtherShareHash              string      `json:"other_share_hash"`
	Share                       string      `json:"share"`
	BoxID                       string      `json:"box_id"`
	EncryptedInvitationKeyShare null.String `json:"-"`
	creatorID                   string
}

func (src BoxKeyShare) toSQLBoiler() *sqlboiler.BoxKeyShare {
	dest := sqlboiler.BoxKeyShare{
		OtherShareHash:              src.OtherShareHash,
		Share:                       src.Share,
		BoxID:                       src.BoxID,
		EncryptedInvitationKeyShare: src.EncryptedInvitationKeyShare,
		CreatorID:                   src.creatorID,
	}
	return &dest
}

func fromSQLBoiler(src *sqlboiler.BoxKeyShare) BoxKeyShare {
	dest := BoxKeyShare{
		OtherShareHash:              src.OtherShareHash,
		Share:                       src.Share,
		BoxID:                       src.BoxID,
		EncryptedInvitationKeyShare: src.EncryptedInvitationKeyShare,
		creatorID:                   src.CreatorID,
	}
	return dest
}
