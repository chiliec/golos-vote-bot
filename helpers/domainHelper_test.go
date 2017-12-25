package helpers

import (
	"fmt"
	"testing"

	"github.com/GolosTools/golos-vote-bot/config"
)

func TestGetDomainRegexp(t *testing.T) {
	var configuration config.Config
	err := config.LoadConfiguration("../config.json", &configuration)
	if err != nil {
		t.Error(err)
	}
	regexp, err := GetDomainRegexp(configuration.Domains)
	texts := []string{
		"https://mapala.net/ru/liora137/diekabr-skii-nizhnii-riedkiie-luchiki-solntsa/",
		"https://goldvoice.club/@golos.loto/5x36-golos-lottery-1273-7757/",
		"https://golos.io/ru--art/@amalinavia/otkrytka-k-novomu-godu-pryanichnyi-domik",
	}
	for _, text := range texts {
		if !regexp.MatchString(text) {
			t.Error("Не подошла ссылка " + text)
		}
	}

	matched := regexp.FindStringSubmatch(texts[0])
	author, permalink := matched[1], matched[2]
	if author != "liora137" || permalink != "diekabr-skii-nizhnii-riedkiie-luchiki-solntsa" {
		t.Error(fmt.Sprintf("Неожиданный автор %s и ссылка %s", author, permalink))
	}

	matched = regexp.FindStringSubmatch(texts[1])
	author, permalink = matched[1], matched[2]
	if author != "golos.loto" || permalink != "5x36-golos-lottery-1273-7757" {
		t.Error(fmt.Sprintf("Неожиданный автор %s и ссылка %s", author, permalink))
	}

	matched = regexp.FindStringSubmatch(texts[2])
	author, permalink = matched[1], matched[2]
	if author != "amalinavia" || permalink != "otkrytka-k-novomu-godu-pryanichnyi-domik" {
		t.Error(fmt.Sprintf("Неожиданный автор %s и ссылка %s", author, permalink))
	}
}
