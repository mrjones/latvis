package main

import (
	oauth "github.com/hokapoka/goauth"
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
		// NOTE: The AuthorizeToken URL for latitude is different than for
		// standard Google applications.
	AuthorizationURL:"https://www.google.com/latitude/apps/OAuthAuthorizeToken",
	ConsumerKey:"mrjon.es",
	ConsumerSecret:"UpS7//zXk60DkyDO8ES/xeS3",
	CallBackURL:"oob",
	AdditionalParams:oauth.Params{
			&oauth.Pair{ Key:"scope", Value:"https://www.googleapis.com/auth/latitude"},
		},
	}

	url, requestToken, err := googleConsumer.GetRequestAuthorizationURL()
	if err != nil{ log.Exit(err) }

	// The latitude API requires additional parameters
	url = url + "&domain=mrjon.es&location=all&granularity=best"

	fmt.Printf("Go to this URL: '%s'\n", url)
	fmt.Printf("Grant access, and then enter the verification code here: ")

	verificationCode := ""

	fmt.Scanln(&verificationCode)

	accessToken := googleConsumer.GetAccessToken(requestToken.Token, verificationCode)

	params := oauth.Params{
		&oauth.Pair{Key:"key", Value:"AIzaSyDd0W4n2lc03aPFtT0bHJAb2xkNHSduAGE"},
		&oauth.Pair{Key:"granularity", Value:"best"},
		&oauth.Pair{Key:"max-results", Value:"1"},
	}

	locationHistoryUrl := "https://www.googleapis.com/latitude/v1/location"

	response, err := googleConsumer.Get(
		locationHistoryUrl,
		params,
		accessToken)

	if err != nil{ log.Exit(err) }

	fmt.Println(response.Status + "\n")
	for key, value := range response.Header {
		fmt.Printf("%s = %s\n", key, value)
	}
	body, err := ioutil.ReadAll(response.Body)

	if err != nil{ log.Exit(err) }

	response.Body.Close()

	fmt.Println(string(body))
}
