package main

import (
	oauth "github.com/hokapoka/goauth"
//	oauth "./oauth"
	"fmt"
	"io/ioutil"
	"log"
)

var googleConsumer *oauth.OAuthConsumer
var accessToken *oauth.AccessToken

func main() {
	googleConsumer = &oauth.OAuthConsumer{
	Service:"google",
	RequestTokenURL:"https://www.google.com/accounts/OAuthGetRequestToken",
	AccessTokenURL:"https://www.google.com/accounts/OAuthGetAccessToken",
//	AuthorizationURL:"https://www.google.com/accounts/OAuthAuthorizeToken",
	AuthorizationURL:"https://www.google.com/latitude/apps/OAuthAuthorizeToken",
	ConsumerKey:"mrjon.es",
	ConsumerSecret:"UpS7//zXk60DkyDO8ES/xeS3",
	CallBackURL:"oob",
	AdditionalParams:oauth.Params{
			&oauth.Pair{ Key:"scope", Value:"https://www.googleapis.com/auth/latitude"},
		},
	}

	url, requestToken, err := googleConsumer.GetRequestAuthorizationURL()
	if err != nil{
		log.Exit(err)
	}
	url = url + "&domain=mrjon.es&location=all&granularity=best"

	fmt.Printf("Go to this URL: '%s'\n", url)
	fmt.Printf("Grant access, and then enter the verification code here: ")

	verificationCode := ""

	fmt.Scanln(&verificationCode)

	fmt.Printf("Verifier: '%s'\n", verificationCode)
	fmt.Printf("GetAccessToken(%s, %s)\n", requestToken.Token, verificationCode)
	accessToken := googleConsumer.GetAccessToken(requestToken.Token, verificationCode)
	
	if accessToken == nil { log.Exit("ERROR") }
	fmt.Printf("Access token ID: '%s'\n", accessToken.Id)
	fmt.Printf("Access token Token: '%s'\n", accessToken.Token)
	fmt.Printf("Access token Secret: '%s'\n", accessToken.Secret)
	fmt.Printf("Access token UserRef: '%s'\n", accessToken.UserRef)
	fmt.Printf("Access token Verifier: '%s'\n", accessToken.Verifier)
	fmt.Printf("Access token Service: '%s'\n", accessToken.Service)


//		"https://www.googleapis.com/latitude/v1/currentLocation",
//		"https://www.googleapis.com/latitude/v1/location?key=" + googleApiKey + "&min-time=1293840000&max-time=1293926400&max-results=10",

	params := oauth.Params{
		&oauth.Pair{Key:"key", Value:"AIzaSyDd0W4n2lc03aPFtT0bHJAb2xkNHSduAGE"},
	}


	response, err := googleConsumer.Get(
		"https://www.googleapis.com/latitude/v1/currentLocation",
		params,
		accessToken)

	if err != nil{
		fmt.Println("ERROR!")
		log.Exit(err)
	}

	fmt.Println(response.Status + "\n")
	for key, value := range response.Header {
		fmt.Printf("%s = %s\n", key, value)
	}
	body, err := ioutil.ReadAll(response.Body)
	if err != nil{
		fmt.Println("ERROR!")
		log.Exit(err)
	}
	response.Body.Close()
	fmt.Println(string(body))
}
