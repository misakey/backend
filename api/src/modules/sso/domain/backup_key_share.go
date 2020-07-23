package domain

type BackupKeyShare struct {
	AccountID      string `json:"account_id"`
	SaltBase64     string `json:"salt_base64"`
	Share          string `json:"share"`
	OtherShareHash string `json:"other_share_hash"`
}
