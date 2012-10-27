package latvis

import (
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
	// TODO(mrjones): flag/config file?
	API_KEY              = "AIzaSyDd0W4n2lc03aPFtT0bHJAb2xkNHSduAGE"
	// TODO(mrjones): flag/config file?
	CLIENT_ID            = "202917186305-0l82gmi2lg74nc1v62r364ec3e2240u9.apps.googleusercontent.com"
	// TODO(mrjones): flag/config file?
	CLIENT_SECRET        = "s-DSmW16VVC6tW-9BSctdML5"
	LOCATION_HISTORY_URL = "https://www.googleapis.com/latitude/v1/location"
	OUT_OF_BAND_CALLBACK = "oob"
	MAX_RESULTS          = 1000
)

// ======================================
// ========== DATA FETCH API ============
// ======================================

// TODO(mrjones): document object lifetime
type Authorizer interface {
	StartAuthorize(applicationStats string) string
	FinishAuthorize(verificationCode string) (DataStream, error)
}

// TODO(mrjones): remove callback url?
func GetAuthorizer(callbackUrl string) Authorizer {
	return &AuthorizerImpl{oauthConfig: NewOauthConfig(callbackUrl)}
}

type DataStream interface {
	FetchRange(start, end time.Time) (*History, error)
}

// ======================================
// ========== IMPLEMENTATION ============
// ======================================

type AuthorizerImpl struct {
	oauthConfig *oauth.Config
}

func (auth *AuthorizerImpl) StartAuthorize(applicationState string) string {
	return auth.oauthConfig.AuthCodeURL(applicationState)
}

func (auth *AuthorizerImpl) FinishAuthorize(verificationCode string) (DataStream, error) {
	// TODO(mrjones): remove reference to configHolder
	transport := &oauth.Transport{Config: configHolder}

//  TODO(mrjones): ok to remove?
//	_, err := transport.Exchange(verificationCode)
//	if err != nil {
//		return nil, err
//	}

	return &DataStreamImpl{client: &ApiClient{Client: transport.Client()}}, nil
}


// OLD STUFF ============================


// Simple ApiClient supports raw (authenticated) HTTP requests to the
// latitude API.
type ApiClientInterface interface {
	FetchUrl(url string, params url.Values) (responseBody string, err error)
}

// TODO(mrjones): gross
var inited = false
var configHolder = &oauth.Config{}

func NewOauthConfig(callbackUrl string) *oauth.Config {
	return &oauth.Config {
		ClientId:     CLIENT_ID,
		ClientSecret: CLIENT_SECRET,
		Scope:        "https://www.googleapis.com/auth/latitude.all.best",
		AuthURL:      "https://accounts.google.com/o/oauth2/auth",
		TokenURL:     "https://accounts.google.com/o/oauth2/token",
	  RedirectURL:  callbackUrl,
	}
}



// DataStream implementation
// Layer on top of ApiClient to support latvis-specific history fetching
// from the latitude API.

type DataStreamImpl struct {
	client ApiClientInterface
}


// JSON Data Model of Latitude API Responses
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

func (stream *DataStreamImpl) fetchJsonForRange(startMs int64, endMs int64) (*JsonRoot, error) {
	fmt.Printf("fetchJsonForRange: %d - %d\n", startMs, endMs)
	params := make(url.Values)
	params.Set("granularity", "best")
	params.Set("max-results", strconv.Itoa(MAX_RESULTS))
	params.Set("min-time", strconv.FormatInt(startMs, 10))
	params.Set("max-time", strconv.FormatInt(endMs, 10))

	body, err := stream.client.FetchUrl(LOCATION_HISTORY_URL, params)
	if err != nil {
		return nil, wrapError("fetchJsonForRange error / "+LOCATION_HISTORY_URL, err)
	}
//	fmt.Println("JSON: ", body)

	var jsonObject JsonRoot
	err = json.Unmarshal([]byte(body), &jsonObject)
	if err != nil {
		return nil, wrapError("JSON Error", err)
	}
	return &jsonObject, nil
}

func (stream *DataStreamImpl) parseJson(jsonObject *JsonRoot, out *History) (startMs int64, endMs int64, count int, err error) {
	minTs := int64(-1)
	maxTs := int64(-1)

	for i := 0; i < len(jsonObject.Data.Items); i++ {
		point := &Coordinate{
			Lat: jsonObject.Data.Items[i].Latitude,
			Lng: jsonObject.Data.Items[i].Longitude}
		out.Add(point)
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

func (stream *DataStreamImpl) FetchRange(start, end time.Time) (*History, error) {
	history := &History{}

	startTs := 1000 * start.Unix()
	endTs := 1000 * end.Unix()
	keepGoing := true

	// The Latitude API returns points at the end of the time range we ask for.
	// So we iteratively shrink our window, excluding the time range covered by
	// the data recieved so far, until we no longer get any new data.
	for keepGoing {
		json, err := stream.fetchJsonForRange(startTs, endTs)
		if err != nil {
			return nil, err
		}

		// TODO(mrjones): verify that we're getting data at the end of the
		// window as we expect.
		minTs, _, itemsReturned, err := stream.parseJson(json, history)

		if err != nil {
			return nil, err
		}
		keepGoing = (itemsReturned > 0)
		// Make sure we exclude everything we've seen: ask for the min, minus 1ms
		endTs = minTs - 1  
	}
	return history,nil
}

// ApiClientInterface implementation
type ApiClient struct {
	Client *http.Client
}

func (conn *ApiClient) FetchUrl(url string, params url.Values) (responseBody string, err error) {
	params.Set("key", API_KEY)

	response, err := conn.Client.Get(url + "?" + params.Encode())

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
