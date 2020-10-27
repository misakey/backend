package cache

import "fmt"

func GetSenderBoxesKey(senderID string) string {
	return fmt.Sprintf("cache:user_%s:boxIDs", senderID)
}

func GetBoxMembersKey(boxID string) string {
	return fmt.Sprintf("cache:box_%s:membersIDs", boxID)
}

func GetEventCountKey(identityID, boxID string) string {
	return fmt.Sprintf("eventCounts:user_%s:%s", identityID, boxID)
}

func GetToNotifyKey(identityID, boxID string) string {
	return fmt.Sprintf("toNotify:user_%s:%s", identityID, boxID)
}

func GetAllToNotifyKeysForIdentity(identityID string) string {
	return fmt.Sprintf("toNotify:user_%s:*", identityID)
}

func GetAllToNotifyKeys() string {
	return "toNotify:*"
}

func GetEventCountKeys(identityID string) string {
	return fmt.Sprintf("eventCounts:user_%s:*", identityID)
}
