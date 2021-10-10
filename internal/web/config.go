package web

import (
	"encoding"
	"errors"
	"strings"
)

type Config struct { //nolint:maligned
	Port int

	// UseEmbed defines whether server should use embedded templates and static files
	UseEmbed bool

	// EnableProfiling can be used to enable pprof handlers
	EnableProfiling bool

	// Auth contains auth configuration
	Auth AuthConfig
}

type AuthConfig struct {
	// Disable disables auth
	Disable bool

	// Type defines the auth type. Available options: 'basic' and 'totp'
	Type string

	// BasicAuthCreds is a list of pairs 'login:password' separated by comma.
	// Passwords must be hashed using BCrypt
	BasicAuthCreds Credentials

	// TOTPAuthSecrets is a list of pairs 'login:secret' separated by comma
	TOTPAuthSecrets Credentials
}

type Credentials map[string]string

var _ encoding.TextUnmarshaler = (*Credentials)(nil)

func (c *Credentials) UnmarshalText(text []byte) error {
	m := make(Credentials)

	pairs := strings.Split(string(text), ",")
	for _, pair := range pairs {
		split := strings.Split(pair, ":")
		if len(split) != 2 {
			return errors.New("invalid credential pair")
		}

		login := split[0]
		password := split[1]
		if login == "" || password == "" {
			return errors.New("credentials can't be empty")
		}

		m[login] = password
	}

	*c = m

	return nil
}

func (c Credentials) Get(username string) (secret string, ok bool) {
	secret, ok = c[username]
	return secret, ok
}
