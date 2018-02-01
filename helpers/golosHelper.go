package helpers

import (
	"log"
	"sync"

	golosClient "github.com/asuleymanov/golos-go/client"

	"database/sql"
	configuration "github.com/GolosTools/golos-vote-bot/config"
	"github.com/GolosTools/golos-vote-bot/models"
)

var config configuration.Config

func init() {
	c, err := GetConfig()
	if err != nil {
		log.Panic(err.Error())
	}
	config = c
}

func SendComment(author, permalink, text string) error {
	golos := golosClient.NewApi(config.Rpc, config.Chain)
	defer golos.Rpc.Close()
	vote := golosClient.PC_Vote{Weight: 100 * 100}
	options := golosClient.PC_Options{Percent: 50}
	err := golos.Comment(
		config.Account,
		author,
		permalink,
		text,
		&vote,
		&options)
	return err
}

func Vote(author, permalink string, database *sql.DB) (successVotesCount int) {
	credentials, err := models.GetAllActiveCredentials(database)
	if err != nil {
		log.Println("Не смогли извлечь ключи из базы")
		return
	}
	for _, credential := range credentials {
		if config.Account != credential.UserName {
			golosClient.Key_List[credential.UserName] = golosClient.Keys{PKey: config.PostingKey}
		}
	}
	log.Printf("Голосую за пост %s/%s, загружено %d аккаунтов", author, permalink, len(credentials))
	var errors []error
	var wg sync.WaitGroup
	wg.Add(len(credentials))
	for _, credential := range credentials {
		go func(credential models.Credential) {
			defer wg.Done()
			weight := credential.Power * 100
			golos := golosClient.NewApi(config.Rpc, config.Chain)
			defer golos.Rpc.Close()
			err := golos.Vote(credential.UserName, author, permalink, weight)
			if err != nil {
				log.Println("Ошибка при голосовании: " + err.Error())
				errors = append(errors, err)
			}
		}(credential)
	}
	wg.Wait()
	successVotesCount = len(credentials) - len(errors)
	return successVotesCount
}
