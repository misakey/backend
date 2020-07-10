package domain

type BackupKeyShare struct {
	AccountID      string `json:"account_id"`
	Share          string `json:"share"`
	OtherShareHash string `json:"other_share_hash"`
}
