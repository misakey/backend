package identity

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"gitlab.misakey.dev/misakey/backend/api/src/modules/sso/repositories/sqlboiler"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merror"
)

//
// models
//

type Notification struct {
	ID             int       `json:"id"`
	Type           string    `json:"type"`
	Details        null.JSON `json:"details"`
	CreatedAt      time.Time `json:"created_at"`
	AcknowledgedAt null.Time `json:"acknowledged_at"`

	identityID string
}

func newNotification() *Notification { return &Notification{} }

func (n Notification) toSQLBoiler() sqlboiler.IdentityNotification {
	result := sqlboiler.IdentityNotification{
		ID:             n.ID,
		IdentityID:     n.identityID,
		Type:           n.Type,
		Details:        n.Details,
		CreatedAt:      n.CreatedAt,
		AcknowledgedAt: n.AcknowledgedAt,
	}
	return result
}

func (n *Notification) fromSQLBoiler(src sqlboiler.IdentityNotification) *Notification {
	n.ID = src.ID
	n.identityID = src.IdentityID
	n.Type = src.Type
	n.Details = src.Details
	n.CreatedAt = src.CreatedAt
	n.AcknowledgedAt = src.AcknowledgedAt
	return n
}

//
// notification methods
//

func NotificationCreate(ctx context.Context, exec boil.ContextExecutor, identityID string, nType string, details null.JSON) error {
	notif := Notification{
		Type:       nType,
		Details:    details,
		CreatedAt:  time.Now(),
		identityID: identityID,
	}
	record := notif.toSQLBoiler()
	return record.Insert(ctx, exec, boil.Infer())
}

func NotificationBulkCreate(ctx context.Context, exec boil.ContextExecutor, identityIDs []string, nType string, details null.JSON) error {
	for _, identityID := range identityIDs {
		notif := Notification{
			Type:       nType,
			Details:    details,
			CreatedAt:  time.Now(),
			identityID: identityID,
		}.toSQLBoiler()
		if err := notif.Insert(ctx, exec, boil.Infer()); err != nil {
			return err
		}
	}
	return nil
}

// Count unacknowledged notifications for received identity id
func NotificationCount(ctx context.Context, exec boil.ContextExecutor, identityID string) (n int, err error) {
	mods := []qm.QueryMod{
		sqlboiler.IdentityNotificationWhere.IdentityID.EQ(identityID),
		sqlboiler.IdentityNotificationWhere.AcknowledgedAt.IsNull(),
	}

	count, err := sqlboiler.IdentityNotifications(mods...).Count(ctx, exec)
	if err != nil {
		return -1, merror.Transform(err).Describe("couting identity notifications")
	}
	return int(count), nil
}

// Returns list of notifications linked to the received identity id
// - handles pagination.
func NotificationList(
	ctx context.Context, exec boil.ContextExecutor,
	identityID string, offset, limit null.Int,
) ([]*Notification, error) {
	mods := []qm.QueryMod{
		sqlboiler.IdentityNotificationWhere.IdentityID.EQ(identityID),
		qm.OrderBy(sqlboiler.IdentityNotificationColumns.CreatedAt + " DESC"),
	}
	if offset.Valid {
		mods = append(mods, qm.Offset(offset.Int))
	}
	if limit.Valid {
		mods = append(mods, qm.Limit(limit.Int))
	}
	records, err := sqlboiler.IdentityNotifications(mods...).All(ctx, exec)
	if err != nil {
		return nil, err
	}

	notifs := make([]*Notification, len(records))
	for i, record := range records {
		notifs[i] = newNotification().fromSQLBoiler(*record)
	}
	return notifs, nil
}

func markInvitationAsUsed(ctx context.Context, exec boil.ContextExecutor, notif *sqlboiler.IdentityNotification) error {
	detailsMap := make(map[string]interface{})
	err := json.Unmarshal(notif.Details.JSON, &detailsMap)
	if err != nil {
		return err
	}

	detailsMap["used"] = true

	notif.Details.JSON, err = json.Marshal(detailsMap)
	if err != nil {
		return err
	}
	nbRowsAffected, err := notif.Update(ctx, exec, boil.Infer())
	if err != nil {
		return err
	}
	if nbRowsAffected != 1 {
		return merror.Error{
			Desc: fmt.Sprintf(`%d rows affected (expected 1)`, nbRowsAffected),
		}
	}

	return nil
}

func NotificationMarkAutoInvitationUsed(ctx context.Context, exec boil.ContextExecutor, cryptoactionID string) error {
	searchedJSONPath := fmt.Sprintf(`{"cryptoaction_id":"%s"}`, cryptoactionID)
	notifs, err := sqlboiler.IdentityNotifications(
		qm.Where(`details::jsonb @> ?`, searchedJSONPath),
	).All(ctx, exec)
	if err != nil {
		return err
	}

	for _, notif := range notifs {
		notifErr := markInvitationAsUsed(ctx, exec, notif)
		if notifErr != nil {
			log.Error().Err(notifErr).Msg(fmt.Sprintf(`marking notif %d as used`, notif.ID))
		}
	}

	return nil
}

// Set acknowledged_at to time.Now() for all unacknowledged notification of the received identity id
// // if notifIds don't belong to the identity id, it will be ignored
func NotificationAck(ctx context.Context, exec boil.ContextExecutor, identityID string, notifIDs []int) error {
	acknowledgedAt := sqlboiler.M{sqlboiler.IdentityNotificationColumns.AcknowledgedAt: null.TimeFrom(time.Now())}
	mods := []qm.QueryMod{
		sqlboiler.IdentityNotificationWhere.IdentityID.EQ(identityID),
		sqlboiler.IdentityNotificationWhere.AcknowledgedAt.IsNull(),
	}
	if len(notifIDs) > 0 {
		mods = append(mods, sqlboiler.IdentityNotificationWhere.ID.IN(notifIDs))
	}

	// NOTE: don't mind if notification is update (acknowledged)
	_, err := sqlboiler.IdentityNotifications(mods...).UpdateAll(ctx, exec, acknowledgedAt)
	if err != nil {
		return err
	}
	return nil
}
