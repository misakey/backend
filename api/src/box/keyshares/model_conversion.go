package keyshares

import (
	"gitlab.misakey.dev/misakey/backend/api/src/box/repositories/sqlboiler"
)

type BoxKeyShare struct {
	OtherShareHash string `json:"other_share_hash"`
	Share          string `json:"share"`
	BoxID          string `json:"box_id"`
	creatorID      string
}

func (src BoxKeyShare) toSQLBoiler() *sqlboiler.BoxKeyShare {
	dest := sqlboiler.BoxKeyShare{
		OtherShareHash: src.OtherShareHash,
		Share:          src.Share,
		BoxID:          src.BoxID,
		CreatorID:      src.creatorID,
	}
	return &dest
}

func fromSQLBoiler(src *sqlboiler.BoxKeyShare) BoxKeyShare {
	dest := BoxKeyShare{
		OtherShareHash: src.OtherShareHash,
		Share:          src.Share,
		BoxID:          src.BoxID,
		creatorID:      src.CreatorID,
	}
	return dest
}
