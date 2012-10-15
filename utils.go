package latvis

import (
	"errors"
	"net/url"
	"strconv"
	"time"

	"code.google.com/p/goauth2/oauth"
)


func appendTokenToQueryParams(token *oauth.Token, params *url.Values) {
	params.Add("access_token", token.AccessToken)
	params.Add("refresh_token", token.RefreshToken)
	params.Add("expiration_time", strconv.FormatInt(token.Expiry.Unix(), 10))
}

func parseTokenFromQueryParams(params *url.Values) (*oauth.Token, error) {
	unix, err := strconv.ParseInt(params.Get("expiration_time"), 10, 64)
	if (err != nil) {
		return nil, err
	}
	return &oauth.Token{
		AccessToken: params.Get("access_token"),
		RefreshToken: params.Get("refresh_token"),
		Expiry: time.Unix(int64(unix), 0),
	}, nil
}

func wrapError(wrapMsg string, cause error) error {
	return errors.New(wrapMsg + ": " + cause.Error())
}
