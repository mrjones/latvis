package latvis

import (
	"errors"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"
	"time"

	"github.com/mrjones/oauth"
)

const (
	CONSUMER_KEY         = "mrjon.es"
	CONSUMER_SECRET      = "UpS7//zXk60DkyDO8ES/xeS3"
	API_KEY              = "AIzaSyDd0W4n2lc03aPFtT0bHJAb2xkNHSduAGE"
	OUT_OF_BAND_CALLBACK = "oob"
)

// Example usage:
// connection := latitude_api.NewConnection()
// tokenSource := latitude_api.NewSimpleTokenSource(connection)
// authorizedConnection := connection.Authorize(tokenSource.GetToken("userid"))
// authorizedConnection.FetchUrl("url", nil)
//  or
// authorizedConnection.FetchRange(<start time>, <end time>);

//
// JSON Data Model of Latitude API Responses
//
type JsonRoot struct {
	Data JsonData
}

type JsonData struct {
	Kind  string
	Items []JsonItem
}

type JsonItem struct {
	Kind        string
	Latitude    float64
	Longitude   float64
	TimestampMs string
}

//
// (Unauthorized) Connection
//
type Connection struct {
	consumer *oauth.Consumer
}

func NewConnection() *Connection {
	return &Connection{consumer: NewConsumer()}
}

func NewConnectionForConsumer(consumer *oauth.Consumer) *Connection {
	return &Connection{consumer: consumer}
}

func NewConsumer() (consumer *oauth.Consumer) {
	sp := oauth.ServiceProvider{
		RequestTokenUrl: "https://www.google.com/accounts/OAuthGetRequestToken",
		AccessTokenUrl:  "https://www.google.com/accounts/OAuthGetAccessToken",
		// NOTE: The AuthorizeToken URL for latitude is different than for
		// standard Google applications.
		AuthorizeTokenUrl: "https://www.google.com/latitude/apps/OAuthAuthorizeToken",
	}

	c := oauth.NewConsumer(CONSUMER_KEY, CONSUMER_SECRET, sp)
	c.AdditionalParams["scope"] = "https://www.googleapis.com/auth/latitude"
	return c
}

func (connection *Connection) TokenRedirectUrl(callback string) (*oauth.RequestToken, string, error) {
	token, url, err := connection.consumer.GetRequestTokenAndUrl(callback)
	if err != nil {
		return nil, "", err
	}

	// The latitude API requires additional parameters
	url = url + "&domain=mrjon.es&location=all&granularity=best"
	return token, url, nil
}

func (connection *Connection) NewAccessToken() (*oauth.AccessToken, error) {
	token, url, err := connection.consumer.GetRequestTokenAndUrl(OUT_OF_BAND_CALLBACK)
	if err != nil {
		return nil, err
	}

	// The latitude API requires additional parameters
	url = url + "&domain=mrjon.es&location=all&granularity=best"

	fmt.Printf("Go to this URL: '%s'\n", url)
	fmt.Printf("Grant access, and then enter the verification code here: ")

	verificationCode := ""

	fmt.Scanln(&verificationCode)

	return connection.consumer.AuthorizeToken(token, verificationCode)
}

func (connection *Connection) ParseToken(token *oauth.RequestToken, verifier string) (*oauth.AccessToken, error) {
	return connection.consumer.AuthorizeToken(token, verifier)
}

func (connection *Connection) Authorize(token *oauth.AccessToken) *AuthorizedConnection {
	return &AuthorizedConnection{accessToken: token, consumer: connection.consumer}
}

//
// AuthorizedConnection
//

type AuthorizedConnection struct {
	accessToken *oauth.AccessToken
	consumer    *oauth.Consumer
}

func (connection *AuthorizedConnection) FetchUrl(url string, params map[string]string) (responseBody string, err error) {
	params["key"] = API_KEY
	response, err := connection.consumer.Get(url, params, connection.accessToken)

	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	responseBodyBytes, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return "", err
	}
	return string(responseBodyBytes), nil
}

func wrapError(wrapMsg string, cause error) error {
	return errors.New(wrapMsg + ": " + cause.Error())
}

func (conn *AuthorizedConnection) appendTimestampRange(startMs int64, endMs int64, windowSize int, history *History) (minTs int64, itemsReturned int, err error) {
	locationHistoryUrl := "https://www.googleapis.com/latitude/v1/location"

	fmt.Printf("Time Range: %d - %d\n", startMs, endMs)

	params := map[string]string{
		"granularity": "best",
		"max-results": strconv.Itoa(windowSize),
		"min-time":    strconv.FormatInt(startMs, 10),
		"max-time":    strconv.FormatInt(endMs, 10),
	}

	body, err := conn.FetchUrl(locationHistoryUrl, params)
	if err != nil {
		return -1, -1, wrapError("FetchUrl error / "+locationHistoryUrl, err)
	}

	var jsonObject JsonRoot
	err = json.Unmarshal([]byte(body), &jsonObject)
	if err != nil {
		return -1, -1, wrapError("JSON Error", err)
	}

	for i := 0; i < len(jsonObject.Data.Items); i++ {
		point := &Coordinate{
			Lat: jsonObject.Data.Items[i].Latitude,
			Lng: jsonObject.Data.Items[i].Longitude}
		history.Add(point)
		if jsonObject.Data.Items[i].TimestampMs == "" {
			data, err := json.Marshal(jsonObject.Data.Items[i])
			if err != nil {
				fmt.Println("Can't even error properly: " + err.Error())
			}
			fmt.Println("Bad history item: " + string(data))
		} else {
			minTs, err = strconv.ParseInt(jsonObject.Data.Items[i].TimestampMs, 10, 64)
			if err != nil {
				return -1, -1, wrapError(
					"Atoi Error / "+jsonObject.Data.Items[i].TimestampMs, err)
			}
		}
	}

	return minTs, len(jsonObject.Data.Items), nil
}

func (conn *AuthorizedConnection) fetchMilliRange(startTimestamp, endTimestamp int64, history *History) {
	windowEnd := endTimestamp
	windowSize := 1000
	keepGoing := true

	for keepGoing {
		minTs, itemsReturned, err := conn.appendTimestampRange(startTimestamp, windowEnd, windowSize, history)
		if err != nil {
			panic(err)
		}
		fmt.Printf("Got %d items\n", itemsReturned)
		keepGoing = (itemsReturned == windowSize)
		windowEnd = minTs
	}
}

func (conn *AuthorizedConnection) FetchRange(start, end time.Time) (*History, error) {
	history := &History{}

	startTimestamp := 1000 * start.Unix()
	endTimestamp := 1000 * end.Unix()

	parallelism := 20
	shardMillis := float64(endTimestamp-startTimestamp) / float64(parallelism)

	fmt.Printf("Fetching from %d to %d\n", startTimestamp, endTimestamp)

	for i := 0; i < parallelism; i++ {
		shardStart := startTimestamp + int64(float64(i)*shardMillis)
		shardEnd := startTimestamp + int64(float64(i+1)*shardMillis)
		conn.fetchMilliRange(shardStart, shardEnd, history)
	}

	return history, nil
}

//
// Various TokenSources
//

type TokenSource interface {
	GetToken(userid string) (*oauth.AccessToken, error)
}

type SimpleTokenSource struct {
	connection *Connection
}

func NewSimpleTokenSource(connection *Connection) *SimpleTokenSource {
	return &SimpleTokenSource{connection: connection}
}

type CachingTokenSource struct {
	connection *Connection
	cache      *Storage
}

func (source *SimpleTokenSource) GetToken(userid string) (*oauth.AccessToken, error) {
	return source.connection.NewAccessToken()
}

func NewCachingTokenSource(connection *Connection, cache *Storage) *CachingTokenSource {
	return &CachingTokenSource{connection: connection, cache: cache}
}

func (source *CachingTokenSource) GetToken(userid string) (*oauth.AccessToken, error) {
	accessToken, err := source.cache.Fetch(userid)
	if err != nil {
		return nil, err
	}
	if accessToken == nil {
		fmt.Printf("No saved token found. Generating new one")
		accessToken, err = source.connection.NewAccessToken()
		if err != nil {
			return nil, err
		}
		err = source.cache.Store(userid, accessToken)
		if err != nil {
			return nil, err
		}
	}
	return accessToken, nil
}