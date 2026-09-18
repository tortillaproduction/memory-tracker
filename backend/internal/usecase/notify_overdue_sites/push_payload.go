package notify_overdue_sites

import (
	"encoding/json"
	"time"
)

// pushSiteView はService Worker(frontend/src/sw.ts)が読むJSONの1サイト分。
// メールの「サイト名+ステータスの一覧」と「サイト名を記載したCTAボタン」という
// 構成をそのままWeb Push通知に踏襲するため、フィールドはemailSiteViewと対応する。
type pushSiteView struct {
	Name        string `json:"name"`
	StatusLabel string `json:"statusLabel"`
	URL         string `json:"url"`
}

type pushPayload struct {
	Sites []pushSiteView `json:"sites"`
}

// buildPushPayload はユーザーの期限切れサイト一覧からプッシュ通知のペイロードを組み立てる。
// URLには必ずCheckinURL(ワンタイムトークン付き/go/{siteId}リンク)を使うこと。
// SiteURLを直接使うと/go/を経由せずチェックインが記録されない。
func buildPushPayload(sites []*SiteRow, now time.Time) ([]byte, error) {
	payload := pushPayload{Sites: make([]pushSiteView, 0, len(sites))}
	for _, s := range sites {
		payload.Sites = append(payload.Sites, pushSiteView{
			Name:        s.SiteName,
			StatusLabel: statusLabel(s, now),
			URL:         s.CheckinURL,
		})
	}
	return json.Marshal(payload)
}
