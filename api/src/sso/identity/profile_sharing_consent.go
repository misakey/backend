package identity

import (
	"context"
	"database/sql"
	"time"

	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"
	"gitlab.misakey.dev/misakey/backend/api/src/sso/repositories/sqlboiler"
)

type profileSharingConsent struct {
	id              int
	identityID      string
	informationType string
	createdAt       time.Time
	revokedAt       null.Time
}

//
// sqlboiler model helpers
//

func newProfileSharingConsent() *profileSharingConsent { return &profileSharingConsent{} }

func (ifc profileSharingConsent) toSQLBoiler() sqlboiler.IdentityProfileSharingConsent {
	result := sqlboiler.IdentityProfileSharingConsent{
		ID:              ifc.id,
		IdentityID:      ifc.identityID,
		InformationType: ifc.informationType,
		CreatedAt:       ifc.createdAt,
		RevokedAt:       ifc.revokedAt,
	}
	return result
}

func (ifc *profileSharingConsent) fromSQLBoiler(src sqlboiler.IdentityProfileSharingConsent) *profileSharingConsent {
	ifc.id = src.ID
	ifc.identityID = src.IdentityID
	ifc.informationType = src.InformationType
	ifc.createdAt = src.CreatedAt
	ifc.revokedAt = src.RevokedAt
	return ifc
}

//
// methods
//

func createProfileSharingConsent(ctx context.Context, exec boil.ContextExecutor, sharingConsent *profileSharingConsent) error {
	sharingConsent.createdAt = time.Now()

	// convert to sql model
	sqlSharingConsent := sharingConsent.toSQLBoiler()
	return sqlSharingConsent.Insert(ctx, exec, boil.Infer())
}

func revokeConsentByIdentityType(ctx context.Context, exec boil.ContextExecutor, identityID, infoType string) error {
	revocation := sqlboiler.M{sqlboiler.IdentityProfileSharingConsentColumns.RevokedAt: null.TimeFrom(time.Now())}
	mods := []qm.QueryMod{
		sqlboiler.IdentityProfileSharingConsentWhere.IdentityID.EQ(identityID),
		sqlboiler.IdentityProfileSharingConsentWhere.InformationType.EQ(infoType),
	}
	rowsAff, err := sqlboiler.IdentityProfileSharingConsents(mods...).UpdateAll(ctx, exec, revocation)
	if err != nil {
		return err
	}
	if rowsAff == 0 {
		return merror.NotFound().Describe("no profile sharing consent to revoke")
	}
	return nil
}

type profileSharingConsentFilters struct {
	identityID null.String
	revoked    null.Bool
}

func listProfileSharingConsents(ctx context.Context, exec boil.ContextExecutor, filters profileSharingConsentFilters) ([]*profileSharingConsent, error) {
	mods := []qm.QueryMod{}
	if filters.identityID.Valid {
		mods = append(mods, sqlboiler.IdentityProfileSharingConsentWhere.IdentityID.EQ(filters.identityID.String))
	}
	if filters.revoked.Valid {
		if filters.revoked.Bool {
			mods = append(mods, sqlboiler.IdentityProfileSharingConsentWhere.RevokedAt.IsNotNull())
		} else {
			mods = append(mods, sqlboiler.IdentityProfileSharingConsentWhere.RevokedAt.IsNull())
		}
	}

	consentRecords, err := sqlboiler.IdentityProfileSharingConsents(mods...).All(ctx, exec)
	consents := make([]*profileSharingConsent, len(consentRecords))
	if err == sql.ErrNoRows {
		return consents, nil
	}
	if err != nil {
		return consents, err
	}

	for i, record := range consentRecords {
		consents[i] = newProfileSharingConsent().fromSQLBoiler(*record)
	}
	return consents, nil
}
