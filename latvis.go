package main

import (
	"github.com/mrjones/latvis/latitude"
	"github.com/mrjones/latvis/server"

	"flag"
	"fmt"
	"log"
)

func GetApiHistorySource() *latitude.AuthorizedConnection {
	connection := latitude.NewConnection()
	tokenStore := latitude.NewTokenStorage("tokens.txt")

  tokenSource := latitude.NewCachingTokenSource(connection, tokenStore);

	fmt.Println("User to generate map for:")
	var user string
	fmt.Scanln(&user)

	accessToken, err := tokenSource.GetToken(user);
	if err != nil{ log.Fatal(err) }
	return connection.Authorize(accessToken);
}

func main() {
	flag.Parse()

  server.Serve()     
}
