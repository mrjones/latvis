package latvis

import (
	"errors"
	"encoding/json"
	"fmt"
	"io/ioutil"


	"net/http"
	"net/url"
	"strconv"
	"time"

	"code.google.com/p/goauth2/oauth"
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

func AppendTokenToQueryParams(token *oauth.Token, params *url.Values) {
	if params == nil {
		panic("nil map")
	}
	params.Add("access_token", token.AccessToken)
	params.Add("refresh_token", token.RefreshToken)
	params.Add("expiration_time", strconv.FormatInt(token.Expiry.Unix(), 10))
}

// errors?
func ParseTokenFromQueryParams(params *url.Values) *oauth.Token {
	// TODO(mrjones): handle erros
	unix, _ := strconv.ParseInt(params.Get("expiration_time"), 10, 64)
	fmt.Println("Reconstructing token from: " + params.Encode())
	return &oauth.Token{
		AccessToken: params.Get("access_token"),
		RefreshToken: params.Get("refresh_token"),
		Expiry: time.Unix(int64(unix), 0),
	}
}

var inited = false
var configHolder = &oauth.Config{}

func NewOauthConfig(callbackUrl string) *oauth.Config {
	return &oauth.Config {
//		ClientId:     "202917186305.apps.googleusercontent.com",
//		ClientSecret: "misioP3p+wNjgnaN9Z1QzXZR",
		ClientId:     "202917186305-0l82gmi2lg74nc1v62r364ec3e2240u9.apps.googleusercontent.com",
		ClientSecret: "s-DSmW16VVC6tW-9BSctdML5",
		Scope:        "https://www.googleapis.com/auth/latitude.all.best",
		AuthURL:      "https://accounts.google.com/o/oauth2/auth",
		TokenURL:     "https://accounts.google.com/o/oauth2/token",
	  RedirectURL:  callbackUrl,
	}
}

// TODO(mrjones) errors
func OauthClientFromVerificationCode(code string) (*oauth.Token,*http.Client,error) {
	transport := &oauth.Transport{Config: configHolder}
	fmt.Println("CODE: " + code)
	token, err := transport.Exchange(code)
	if err != nil {
		return nil, nil, err
	}
	return token, transport.Client(), nil
}

func OauthClientFromToken(token *oauth.Token) *http.Client {
	transport := &oauth.Transport{Config: configHolder}
	transport.Token = token
	return transport.Client()
}


//func (connection *Connection) NewAccessToken() (*oauth.AccessToken, error) {
//	token, url, err := connection.consumer.GetRequestTokenAndUrl(OUT_OF_BAND_CALLBACK)
//	if err != nil {
//		return nil, err
//	}
//
//	// The latitude API requires additional parameters
//	url = url + "&domain=mrjon.es&location=all&granularity=best"
//
//	fmt.Printf("Go to this URL: '%s'\n", url)
//	fmt.Printf("Grant access, and then enter the verification code here: ")
//
//	verificationCode := ""
//
//	fmt.Scanln(&verificationCode)
//
//	return connection.consumer.AuthorizeToken(token, verificationCode)
//}
//
//func (connection *Connection) ParseToken(token *oauth.RequestToken, verifier string) (*oauth.AccessToken, error) {
//	return connection.consumer.AuthorizeToken(token, verifier)
//}
//
//func (connection *Connection) Authorize(token *oauth.AccessToken) *AuthorizedConnection {
//	return &AuthorizedConnection{accessToken: token, consumer: connection.consumer}
//}

//
// AuthorizedConnection
//

type AuthorizedConnection struct {
	Client *http.Client
}

func (connection *AuthorizedConnection) FetchUrl(url string, params url.Values) (responseBody string, err error) {
	params.Set("key", API_KEY)

//	query := ""
//	delim := ""
//	for k,v := range(params) {
//		query = query + delim + k + "=" + v
//		delim = "&"
//	}


	response, err := connection.Client.Get(url + "?" + params.Encode())

//	request, err := http.NewRequest("GET", url + "?key=" + API_KEY, nil)
//	if err != nil {
//		return "", err
//	}
//	response, err := connection.Client.Do(request)

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

func (conn *AuthorizedConnection) appendTimestampRange(startMs int64, endMs int64, windowSize int, history *History) (minTs int64, maxTs int64, itemsReturned int, err error) {
	locationHistoryUrl := "https://www.googleapis.com/latitude/v1/location"

	fmt.Printf("Time Range: %d - %d\n", startMs, endMs)

	params := make(url.Values)
	params.Set("granularity", "best")
	params.Set("max-results", strconv.Itoa(windowSize))
	params.Set("min-time", strconv.FormatInt(startMs, 10))
	params.Set("max-time", strconv.FormatInt(endMs, 10))

	body, err := conn.FetchUrl(locationHistoryUrl, params)
	if err != nil {
		return -1, -1, -1, wrapError("FetchUrl error / "+locationHistoryUrl, err)
	}

	var jsonObject JsonRoot
	err = json.Unmarshal([]byte(body), &jsonObject)
	if err != nil {
		return -1, -1, -1, wrapError("JSON Error", err)
	}

//	fmt.Println("JSON: ", body)

	minTs = -1
	maxTs = -1

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
			ts, err := strconv.ParseInt(jsonObject.Data.Items[i].TimestampMs, 10, 64)
			if minTs == -1 || ts < minTs {
				minTs = ts
			}
			if maxTs == -1 || ts > maxTs {
				maxTs = ts
			}
			if err != nil {
				return -1, -1, -1, wrapError(
					"Atoi Error / "+jsonObject.Data.Items[i].TimestampMs, err)
			}
		}
	}

	return minTs, maxTs, len(jsonObject.Data.Items), nil
}

func (conn *AuthorizedConnection) fetchMilliRange(startTimestamp, endTimestamp int64, history *History) {
	windowEnd := endTimestamp
	windowSize := 1000
	keepGoing := true

	// The Latitude API returns points at the end of the time range we ask for.
	// So we iteratively shrink our window, excluding the time range covered by
	// the data recieved so far, until we no longer get any new data.
	for keepGoing {
		minTs, maxTs, itemsReturned, err := conn.appendTimestampRange(startTimestamp, windowEnd, windowSize, history)
		if err != nil {
			panic(err)
		}
		fmt.Printf("Got %d items in range: %d - %d \n", itemsReturned, minTs, maxTs)
		keepGoing = (itemsReturned > 0)
		// Make sure we exclude everything we've seen: ask for the min, minus 1ms
		windowEnd = minTs - 1  
	}
}

func (conn *AuthorizedConnection) FetchRange(start, end time.Time) (*History, error) {
	history := &History{}

	startTimestamp := 1000 * start.Unix()
	endTimestamp := 1000 * end.Unix()

	parallelism := 1
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

//type TokenSource interface {
//	GetToken(userid string) (*oauth.AccessToken, error)
//}
//
//type SimpleTokenSource struct {
//	connection *Connection
//}
//
//func NewSimpleTokenSource(connection *Connection) *SimpleTokenSource {
//	return &SimpleTokenSource{connection: connection}
//}
//
//type CachingTokenSource struct {
//	connection *Connection
//	cache      *Storage
//}
//
//func (source *SimpleTokenSource) GetToken(userid string) (*oauth.AccessToken, error) {
//	return source.connection.NewAccessToken()
//}
//
//func NewCachingTokenSource(connection *Connection, cache *Storage) *CachingTokenSource {
//	return &CachingTokenSource{connection: connection, cache: cache}
//}
//
//func (source *CachingTokenSource) GetToken(userid string) (*oauth.AccessToken, error) {
//	accessToken, err := source.cache.Fetch(userid)
//	if err != nil {
//		return nil, err
//	}
//	if accessToken == nil {
//		fmt.Printf("No saved token found. Generating new one")
//		accessToken, err = source.connection.NewAccessToken()
//		if err != nil {
//			return nil, err
//		}
//		err = source.cache.Store(userid, accessToken)
//		if err != nil {
//			return nil, err
//		}
//	}
//	return accessToken, nil
//}
