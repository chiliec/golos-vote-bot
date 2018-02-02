package helpers

import (
	"database/sql"
	"log"
	"sync"

	golosClient "github.com/asuleymanov/golos-go/client"

	configuration "github.com/GolosTools/golos-vote-bot/config"
	"github.com/GolosTools/golos-vote-bot/models"
)

func SendComment(author, permalink, text string, config configuration.Config) error {
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

func Vote(vote models.Vote, database *sql.DB, config configuration.Config) (successVotesCount int, err error) {
	credentials, err := models.GetAllActiveCredentials(database)
	if err != nil {
		return 0, err
	}
	for _, credential := range credentials {
		if config.Account != credential.UserName {
			golosClient.Key_List[credential.UserName] = golosClient.Keys{PKey: config.PostingKey}
		}
	}
	log.Printf("Голосую за пост %s/%s, загружено %d аккаунтов", vote.Author, vote.Permalink, len(credentials))
	var errors []error
	var wg sync.WaitGroup
	wg.Add(len(credentials))
	for _, credential := range credentials {
		go func(credential models.Credential) {
			defer wg.Done()
			weight := credential.Power * 100
			golos := golosClient.NewApi(config.Rpc, config.Chain)
			defer golos.Rpc.Close()
			err := golos.Vote(credential.UserName, vote.Author, vote.Permalink, weight)
			if err != nil {
				log.Println("Ошибка при голосовании: " + err.Error())
				errors = append(errors, err)
			}
		}(credential)
	}
	wg.Wait()
	successVotesCount = len(credentials) - len(errors)
	_, err = vote.Save(database)
	if err != nil {
		return successVotesCount, err
	}
	return successVotesCount, nil
}
