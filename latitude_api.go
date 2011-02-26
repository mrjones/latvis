package latitude_api

import (
	oauth "github.com/hokapoka/goauth"
	"fmt"
	"io/ioutil"
	"os"
)

type Connection struct {
	AccessToken *oauth.AccessToken
	OauthConsumer *oauth.OAuthConsumer
}

func (connection *Connection) FetchUrl(url string, params oauth.Params) (responseBody string, err os.Error) {
	response, err := connection.OauthConsumer.Get(url, params, connection.AccessToken)

	if err != nil { return "", err }
	defer response.Body.Close()
	responseBodyBytes, err := ioutil.ReadAll(response.Body)

	if err != nil { return "", err }
	return string(responseBodyBytes), nil
}

func NewAccessToken(consumer *oauth.OAuthConsumer) (token *oauth.AccessToken, err os.Error) {
	url, requestToken, err := consumer.GetRequestAuthorizationURL()
	if err != nil{ return nil, err }

	// The latitude API requires additional parameters
	url = url + "&domain=mrjon.es&location=all&granularity=best"

	fmt.Printf("Go to this URL: '%s'\n", url)
	fmt.Printf("Grant access, and then enter the verification code here: ")

	verificationCode := ""

	fmt.Scanln(&verificationCode)

	return consumer.GetAccessToken(requestToken.Token, verificationCode), nil
}

func NewConsumer() (consumer *oauth.OAuthConsumer) {
	return &oauth.OAuthConsumer{
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
}
