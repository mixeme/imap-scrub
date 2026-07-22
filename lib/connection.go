package lib

import (
	"fmt"
	"os"

	"github.com/emersion/go-imap/client"
)

// Connect returns a *client.Client
func Connect() *client.Client {
	imapServer := fmt.Sprintf("%s:%d", Config.Host, *Config.Port)
	var c *client.Client
	var err error
	if *Config.SSL {
		c, err = client.DialTLS(imapServer, nil)
		if err != nil {
			Log.ErrorF("%v", err)
			os.Exit(2)
		}
	} else {
		c, err = client.Dial(imapServer)
		if err != nil {
			Log.ErrorF("%v", err)
			os.Exit(2)
		}
	}

	return c
}

// Authenticate logs a client in using the mechanism selected by the config.
// For OAuth2 the caller supplies an access token from OAuthAccessToken() so a
// single token can be shared between the reader and writer connections.
func Authenticate(c *client.Client, accessToken string) error {
	if Config.UseOAuth2() {
		return c.Authenticate(&Xoauth2Client{Username: Config.User, Token: accessToken})
	}

	return c.Login(Config.User, Config.Pass)
}
