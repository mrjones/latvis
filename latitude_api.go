package latitude_api

import (
  "github.com/mrjones/oauth"
	"fmt"
	"io/ioutil"
	"json"
	"./location"
	"strconv"
	"time"
  "./tokens"
	"os"
)

const (
	CONSUMER_KEY = "mrjon.es"
	CONSUMER_SECRET = "UpS7//zXk60DkyDO8ES/xeS3"
	API_KEY = "AIzaSyDd0W4n2lc03aPFtT0bHJAb2xkNHSduAGE"
	OUT_OF_BAND_CALLBACK = "oob"
)	

// Example usage:
// connection := latitude_api.NewConnection()
// tokenSource := latitude_api.NewSimpleTokenSource(connection)
// authorizedConnection := connection.Authorize(tokenSource.GetToken("userid"))
// authorizedConnection.FetchUrl("url", nil)
//  or
// authorizedConnection.GetHistory(2011, 01);

//
// JSON Data Model of Latitude API Responses
//
type JsonRoot struct {
	Data JsonData
}

type JsonData struct {
	Kind string
	Items []JsonItem
}

type JsonItem struct {
	Kind string
	Latitude float64
	Longitude float64
	TimestampMs string
}

//
// (Unauthorized) Connection
//
type Connection struct {
	consumer *oauth.Consumer
}

func NewConnection() *Connection {
	return &Connection{consumer: NewConsumer(OUT_OF_BAND_CALLBACK)}
}

func NewConnectionForConsumer(consumer *oauth.Consumer) *Connection {
	return &Connection{consumer: consumer}  
}

func NewConsumer(callbackUrl string) (consumer *oauth.Consumer) {
	sp := oauth.ServiceProvider{
		RequestTokenUrl: "https://www.google.com/accounts/OAuthGetRequestToken",
		AccessTokenUrl: "https://www.google.com/accounts/OAuthGetAccessToken",
			// NOTE: The AuthorizeToken URL for latitude is different than for
			// standard Google applications.
		AuthorizeTokenUrl: "https://www.google.com/latitude/apps/OAuthAuthorizeToken",
	}

	c := oauth.NewConsumer(CONSUMER_KEY, CONSUMER_SECRET, sp, callbackUrl)
	c.AdditionalParams["scope"] = "https://www.googleapis.com/auth/latitude";
	return c
}

//func (connection *Connection) TokenRedirectUrl() (*oauth.UnauthorizedToken, *string, os.Error) {
//	token, url, err := connection.consumer.GetRequestTokenAndUrl()
//	if err != nil{ return nil, err }
//
/// The latitude API requires additional parameters
//	url = url + "&domain=mrjon.es&location=all&granularity=best"
//  return token, &url, nil
//}

func (connection *Connection) NewAccessToken() (*oauth.AccessToken, os.Error) {
	token, url, err := connection.consumer.GetRequestTokenAndUrl()
	if err != nil{ return nil, err }

	// The latitude API requires additional parameters
	url = url + "&domain=mrjon.es&location=all&granularity=best"

	fmt.Printf("Go to this URL: '%s'\n", url)
	fmt.Printf("Grant access, and then enter the verification code here: ")

	verificationCode := ""

	fmt.Scanln(&verificationCode)

	return connection.consumer.AuthorizeToken(token, verificationCode)
}

//func (connection *Connection) ParseToken(token string, verifier string) *oauth.AccessToken {
//  return connection.consumer.AuthorizeToken(token, verifier)
//}

func (connection *Connection) Authorize(token *oauth.AccessToken) *AuthorizedConnection {
	return &AuthorizedConnection{accessToken: token, consumer: connection.consumer}
}

//
// AuthorizedConnection
//

type AuthorizedConnection struct {
	accessToken *oauth.AccessToken
	consumer *oauth.Consumer
}

func (connection *AuthorizedConnection) FetchUrl(url string, params map[string]string) (responseBody string, err os.Error) {
  params["key"] = API_KEY
	response, err := connection.consumer.Get(url, params, connection.accessToken)

	if err != nil { return "", err }
	defer response.Body.Close()
	responseBodyBytes, err := ioutil.ReadAll(response.Body)

	if err != nil { return "", err }
	return string(responseBodyBytes), nil
}

func (conn *AuthorizedConnection) appendTimestampRange(startMs int64, endMs int64, windowSize int, history *location.History) (minTs int64, itemsReturned int, err os.Error) {
	locationHistoryUrl := "https://www.googleapis.com/latitude/v1/location"

	fmt.Printf("Time Range: %d - %d\n", startMs, endMs)

	params := map[string]string {
		"granularity": "best",
		"max-results": strconv.Itoa(windowSize),
		"min-time": strconv.Itoa64(startMs),
		"max-time": strconv.Itoa64(endMs),
	}

	body, err := conn.FetchUrl(locationHistoryUrl, params)
	if err != nil { return -1, -1, err }

	var jsonObject JsonRoot
	err = json.Unmarshal([]byte(body), &jsonObject)
	if err != nil { return -1, -1, err }

	for i := 0 ; i < len(jsonObject.Data.Items) ; i++ {
		point := &location.Coordinate{Lat: jsonObject.Data.Items[i].Longitude, Lng: jsonObject.Data.Items[i].Latitude }
		if point.Lat > -74.02 && point.Lat < -73.96 && point.Lng > 40.703 && point.Lng < 40.8 {
			history.Add(point)
		}
		minTs, err = strconv.Atoi64(jsonObject.Data.Items[i].TimestampMs)
		if err != nil { return -1, -1, err }
	}

	return minTs, len(jsonObject.Data.Items), nil
}

func (conn *AuthorizedConnection) GetHistory(year int64, month int) (*location.History, os.Error) {
	startTime := time.Time{Year: year, Month: month, Day: 1}
	endTime := time.Time{Year: year, Month: month + 1, Day: 1}
	startTimestamp := 1000* startTime.Seconds()
	endTimestamp := 1000 * endTime.Seconds()

	history := &location.History{}

	windowEnd := endTimestamp
	windowSize := 1000
	keepGoing := true

	for keepGoing {
		minTs, itemsReturned, err := conn.appendTimestampRange(startTimestamp, windowEnd, windowSize, history)
		if err != nil { return nil, err }
		fmt.Printf("Got %d items\n", itemsReturned)
		keepGoing = (itemsReturned == windowSize)
		windowEnd = minTs
	}

	return history, nil
}

//
// Various TokenSources
//

type TokenSource interface {
  GetToken(userid string) (*oauth.AccessToken, os.Error)
}

type SimpleTokenSource struct {
  connection *Connection
}

func NewSimpleTokenSource(connection *Connection) *SimpleTokenSource {
  return &SimpleTokenSource{connection: connection}
}

type CachingTokenSource struct {
  connection *Connection
  cache *tokens.Storage
}

func (source *SimpleTokenSource) GetToken(userid string) (*oauth.AccessToken, os.Error) {
  return source.connection.NewAccessToken();
}

func NewCachingTokenSource(connection *Connection, cache *tokens.Storage) *CachingTokenSource {
  return &CachingTokenSource{connection: connection, cache: cache}
}

func (source *CachingTokenSource) GetToken(userid string) (*oauth.AccessToken, os.Error) {
 	accessToken, err := source.cache.Fetch(userid)
	if err != nil{ return nil, err }
	if accessToken == nil {
		fmt.Printf("No saved token found. Generating new one")
		accessToken, err = source.connection.NewAccessToken()
		if err != nil{ return nil, err }
		err = source.cache.Store(userid, accessToken)
		if err != nil{ return nil, err }
	}
	return accessToken, nil
}
