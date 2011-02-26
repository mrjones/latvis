package latitude_api

import (
	oauth "github.com/hokapoka/goauth"
	"fmt"
	"io/ioutil"
	"os"
)

type Connection struct {
	consumer *oauth.OAuthConsumer
}

type AuthorizedConnection struct {
	accessToken *oauth.AccessToken
	consumer *oauth.OAuthConsumer
}

const (
	CONSUMER_KEY = "mrjon.es"
	CONSUMER_SECRET = "UpS7//zXk60DkyDO8ES/xeS3"
	API_KEY = "AIzaSyDd0W4n2lc03aPFtT0bHJAb2xkNHSduAGE"
	OUT_OF_BAND_CALLBACK = "oob"
)	

func NewConnection() *Connection {
	return &Connection{consumer: newConsumer()}
}

func newConsumer() (consumer *oauth.OAuthConsumer) {
	return &oauth.OAuthConsumer{
	Service:"google",
	RequestTokenURL:"https://www.google.com/accounts/OAuthGetRequestToken",
	AccessTokenURL:"https://www.google.com/accounts/OAuthGetAccessToken",
		// NOTE: The AuthorizeToken URL for latitude is different than for
		// standard Google applications.
	AuthorizationURL:"https://www.google.com/latitude/apps/OAuthAuthorizeToken",
	ConsumerKey:CONSUMER_KEY,
	ConsumerSecret:CONSUMER_SECRET,
	CallBackURL:OUT_OF_BAND_CALLBACK,
	AdditionalParams:oauth.Params{
			&oauth.Pair{ Key:"scope", Value:"https://www.googleapis.com/auth/latitude"},
		},
	}
}

func (connection *Connection) NewAccessToken() (token *oauth.AccessToken, err os.Error) {
	url, requestToken, err := connection.consumer.GetRequestAuthorizationURL()
	if err != nil{ return nil, err }

	// The latitude API requires additional parameters
	url = url + "&domain=mrjon.es&location=all&granularity=best"

	fmt.Printf("Go to this URL: '%s'\n", url)
	fmt.Printf("Grant access, and then enter the verification code here: ")

	verificationCode := ""

	fmt.Scanln(&verificationCode)

	return connection.consumer.GetAccessToken(requestToken.Token, verificationCode), nil
}

func (connection *Connection) Authorize(token *oauth.AccessToken) *AuthorizedConnection {
	return &AuthorizedConnection{accessToken: token, consumer: connection.consumer}
}

func (connection *AuthorizedConnection) FetchUrl(url string, params oauth.Params) (responseBody string, err os.Error) {
	response, err := connection.consumer.Get(url, params, connection.accessToken)

	params.Add(&oauth.Pair{Key:"key", Value: API_KEY})

	if err != nil { return "", err }
	defer response.Body.Close()
	responseBodyBytes, err := ioutil.ReadAll(response.Body)

	if err != nil { return "", err }
	return string(responseBodyBytes), nil
}