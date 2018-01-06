package helpers

import "testing"

func TestGetInstantViewLink(t *testing.T) {
	author := "some-author"
	permalink := "some-permalink"
	instantViewLink := GetInstantViewLink(author, permalink)
	expectedLink := "https://t.me/iv?url=https://goldvoice.club/" + "@" + author + "/" + permalink + "&rhash=70f46c6616076d"
	if expectedLink != instantViewLink {
		t.Fatal("Неожиданная ссылка")
	}
}
