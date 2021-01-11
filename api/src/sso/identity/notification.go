package identity

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v7"
	"github.com/rs/zerolog/log"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"gitlab.misakey.dev/misakey/backend/api/src/box/realtime"
	"gitlab.misakey.dev/misakey/backend/api/src/sdk/merr"
	"gitlab.misakey.dev/misakey/backend/api/src/sso/repositories/sqlboiler"
)

// Notification ...
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

// NotificationCreate ...
func NotificationCreate(ctx context.Context, exec boil.ContextExecutor, redConn *redis.Client, identityID string, nType string, details null.JSON) error {
	notif := Notification{
		Type:       nType,
		Details:    details,
		CreatedAt:  time.Now(),
		identityID: identityID,
	}
	record := notif.toSQLBoiler()
	if err := record.Insert(ctx, exec, boil.Infer()); err != nil {
		return err
	}

	// send notification in websocket
	notif.fromSQLBoiler(record)
	notifWS := realtime.Update{
		Type:   "notification",
		Object: notif,
	}

	realtime.SendUpdate(ctx, redConn, identityID, &notifWS)

	return nil

}

// NotificationBulkCreate ...
func NotificationBulkCreate(ctx context.Context, exec boil.ContextExecutor, redConn *redis.Client, identityIDs []string, nType string, details null.JSON) error {
	for _, identityID := range identityIDs {
		notif := Notification{
			Type:       nType,
			Details:    details,
			CreatedAt:  time.Now(),
			identityID: identityID,
		}
		record := notif.toSQLBoiler()
		if err := record.Insert(ctx, exec, boil.Infer()); err != nil {
			return err
		}
		// send notification in websocket
		notif.fromSQLBoiler(record)
		notifWS := realtime.Update{
			Type:   "notification",
			Object: notif,
		}

		realtime.SendUpdate(ctx, redConn, identityID, &notifWS)

	}
	return nil
}

// NotificationCount unacknowledged notifications for received identity id
func NotificationCount(ctx context.Context, exec boil.ContextExecutor, identityID string) (n int, err error) {
	mods := []qm.QueryMod{
		sqlboiler.IdentityNotificationWhere.IdentityID.EQ(identityID),
		sqlboiler.IdentityNotificationWhere.AcknowledgedAt.IsNull(),
	}

	count, err := sqlboiler.IdentityNotifications(mods...).Count(ctx, exec)
	if err != nil {
		return -1, merr.From(err).Desc("couting identity notifications")
	}
	return int(count), nil
}

// NotificationList returns list of notifications linked to the received identity id
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
		return merr.Internal().Descf("%d rows affected (expected 1)", nbRowsAffected)
	}

	return nil
}

// NotificationMarkAutoInvitationUsed ...
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

// NotificationAck acknowledged_at to time.Now() for all unacknowledged notification of the received identity id
// if notifIds don't belong to the identity id, it will be ignored
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
