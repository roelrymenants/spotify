// Copyright 2014, 2015 Zac Bergquist
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package spotify

import (
	"errors"
	"net/http"
	"os"

	"golang.org/x/oauth2"
)

const (
	// AuthURL is the URL to Spotify Accounts Service's OAuth2 endpoint.
	AuthURL = "https://accounts.spotify.com/authorize"
	// TokenURL is the URL to the Spotify Accounts Service's OAuth2
	// token endpoint.
	TokenURL = "https://accounts.spotify.com/api/token"
)

// Scopes let you specify exactly which types of data your application wants to access.
// The set of scopes you pass in your authentication request determines what access the
// permissions the user is asked to grant.
const (
	// ScopePlaylistReadPrivate seeks permission to read
	// a user's private playlists.
	ScopePlaylistReadPrivate = "playlist-read-private"
	// ScopePlaylistModifyPublic seeks write access
	// to a user's public playlists.
	ScopePlaylistModifyPublic = "playlist-modify-public"
	// ScopePlaylistModifyPrivate seeks write access to
	// a user's private playlists.
	ScopePlaylistModifyPrivate = "playlist-modify-private"
	// ScopePlaylistReadCollaborative seeks permission to
	// access a user's collaborative playlists.
	ScopePlaylistReadCollaborative = "playlist-read-collaborative"
	// ScopeUserFollowModify seeks write/delete access to
	// the list of artists and other users that a user follows.
	ScopeUserFollowModify = "user-follow-modify"
	// ScopeUserFollowRead seeks read access to the list of
	// artists and other users that a user follows.
	ScopeUserFollowRead = "user-follow-read"
	// ScopeUserLibraryModify seeks write/delete acess to a
	// user's "Your Music" library.
	ScopeUserLibraryModify = "user-library-modify"
	// ScopeUserLibraryRead seeks read access to a user's "Your Music" library.
	ScopeUserLibraryRead = "user-library-read"
	// ScopeUserReadPrivate seeks read access to a user's
	// subsription details (type of user account).
	ScopeUserReadPrivate = "user-read-private"
	// ScopeUserReadEmail seeks read access to a user's email address.
	ScopeUserReadEmail = "user-read-email"
	// ScopeUserReadBirthdate seeks read access to a user's birthdate.
	ScopeUserReadBirthdate = "user-read-birthdate"
)

// Authenticator provides convenience functions for implementing the OAuth2 flow.
// You should always use `NewAuthenticator` to make them.
//
// Example:
//
//     a := spotify.NewAuthenticator(redirectURL, spotify.ScopeUserLibaryRead, spotify.ScopeUserFollowRead)
//     // direct user to Spotify to log in
//     http.Redirect(w, r, a.AuthURL("state-string"), http.StatusFound)
//
//     // then, in redirect handler:
//     token, err := a.Token(state, r)
//     client := a.NewClient(token)
//
type Authenticator struct {
	config *oauth2.Config
}

// NewAuthenticator creates an authenticator which is used to implement the
// OAuth2 authorization flow.  The redirectURL must exactly match one of the
// URLs specified in your Spotify developer account.
//
// By default, NewAuthenticator pulls your client ID and secret key from the
// SPOTIFY_ID and SPOTIFY_SECRET environment variables.  If you'd like to provide
// them from some other source, you can call `SetAuthInfo(id, key)` on the
// returned authenticator.
func NewAuthenticator(redirectURL string, scopes ...string) Authenticator {
	cfg := &oauth2.Config{
		ClientID:     os.Getenv("SPOTIFY_ID"),
		ClientSecret: os.Getenv("SPOTIFY_SECRET"),
		RedirectURL:  redirectURL,
		Scopes:       scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  AuthURL,
			TokenURL: TokenURL,
		},
	}
	return Authenticator{
		config: cfg,
	}
}

// SetAuthInfo overwrites the client ID and secret key used by the authenticator.
// You can use this if you don't want to store this information in environment variables.
func (a *Authenticator) SetAuthInfo(clientID, secretKey string) {
	a.config.ClientID = clientID
	a.config.ClientSecret = secretKey
}

// AuthURL returns a URL to the the Spotify Accounts Service's OAuth2 endpoint.
//
// State is a token to protect the user from CSRF attacks.  You should pass the
// same state to `Token`, where it will be validated.  For more info, refer to
// http://tools.ietf.org/html/rfc6749#section-10.12.
func (a Authenticator) AuthURL(state string) string {
	return a.config.AuthCodeURL(state)
}

// Token pulls an authorization code from an HTTP request and attempts to exchange
// it for an access token.  The standard use case is to call Token from the handler
// that handles requests to your application's redirect URL.
func (a Authenticator) Token(state string, r *http.Request) (*oauth2.Token, error) {
	values := r.URL.Query()
	if e := values.Get("error"); e != "" {
		return nil, errors.New("spotify: auth failed - " + e)
	}
	code := values.Get("code")
	if code == "" {
		return nil, errors.New("spotify: didn't get access code")
	}
	actualState := values.Get("state")
	if actualState != state {
		return nil, errors.New("spotify: redirect state parameter doesn't match")
	}
	return a.config.Exchange(oauth2.NoContext, code)
}

// Exchange is like Token, except it allows you to manually specify the access
// code instead of pulling it out of an HTTP request.
func (a Authenticator) Exchange(code string) (*oauth2.Token, error) {
	return a.config.Exchange(oauth2.NoContext, code)
}

// NewClient creates a Client that will use the specified access token for its API requests.
func (a Authenticator) NewClient(token *oauth2.Token) Client {
	return Client{
		http: a.config.Client(oauth2.NoContext, token),
	}
}
