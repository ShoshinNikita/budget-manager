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

	// SkipAuth disables auth
	SkipAuth bool

	// Credentials is a list of pairs 'login:password' separated by comma.
	// Passwords must be hashed using BCrypt
	Credentials Credentials

	EnableProfiling bool
}

type Credentials map[string]string

var _ encoding.TextUnmarshaler = &Credentials{}

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
			return errors.New("login and password can't be empty")
		}

		m[login] = password
	}

	*c = m

	return nil
}
