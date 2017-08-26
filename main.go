package main

import (
	"flag"
	"fmt"
	"github.com/Chiliec/golos-go/client"
)

const (
	rpc   = "wss://ws.golos.io"
	chain = "golos"
)

func main() {
	author, permlink := "anima", "konkurs-nochnye-ulicy-poslednie-raboty-stop-i-golosovanie"
	voter := "chiliec"
	weight := 10000
	postingKey := flag.String("postingKey", "", "posting key")
	client.Key_List = map[string]client.Keys{voter: client.Keys{postingKey, "", "", ""}}
	api := client.NewApi(rpc, chain)
	fmt.Println(api.Vote(voter, author, permlink, weight))
}
