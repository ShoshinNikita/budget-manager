package totp

//nolint:gosec
import (
	"crypto/hmac"
	"crypto/sha1"
	"crypto/subtle"
	"encoding/base32"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"net/url"
	"strconv"
	"time"
)

// Google Authenticator and other popular apps don't support custom pass length and interval.
// So, hardcode them.
const (
	PassLength = 6
	Interval   = 30 * time.Second
)

type Password string

func (p Password) Equal(p1 string) bool {
	return subtle.ConstantTimeCompare([]byte(p), []byte(p1)) == 1
}

// Generate generates a time-based one-time password. It panics if secret is empty
func Generate(secret string) Password {
	if secret == "" {
		panic("secret can't be empty")
	}

	totp := generateTOTP(secret, time.Now(), Interval)

	// We can cast to int32 because the max pass length is limited
	mod := int32(math.Pow10(PassLength))

	return Password(fmt.Sprintf("%0*d", PassLength, totp%mod))
}

// generateTOTP generates a time-based one-time password
func generateTOTP(secret string, now time.Time, interval time.Duration) int32 {
	counter := now.Unix() / int64(interval.Seconds())

	return generateHOTP(secret, uint64(counter))
}

// generateHOTP generates a HMAC-based one-time password.
//
// Algorithm: https://en.wikipedia.org/wiki/HMAC-based_one-time_password#Algorithm
func generateHOTP(secret string, counter uint64) int32 {
	// Counter must be in big-endian
	c := make([]byte, 8)
	binary.BigEndian.PutUint64(c, counter)

	h := hmac.New(sha1.New, []byte(secret))
	if _, err := h.Write(c); err != nil {
		panic(err)
	}
	mac := h.Sum(nil)

	// Index is the last 4 bits
	index := int(mac[len(mac)-1]) & 0b1111

	// Extract 31 bits
	return 0 |
		int32(mac[index+0])&0b01111111<<24 |
		int32(mac[index+1])&0b11111111<<16 |
		int32(mac[index+2])&0b11111111<<8 |
		int32(mac[index+3])&0b11111111
}

// FormatURL returns the otpauth:// url that can be used to generate a QR code. Username is optional.
//
// URL format: https://docs.yubico.com/yesdk/users-manual/application-oath/uri-string-format.html
func FormatURL(secret, issuer, username string) (string, error) {
	if secret == "" {
		return "", errors.New("secret can't be empty")
	}
	if issuer == "" {
		return "", errors.New("issuer can't be empty")
	}

	query := url.Values{
		"secret": []string{base32.StdEncoding.EncodeToString([]byte(secret))},
		"digits": []string{strconv.Itoa(PassLength)},
		"period": []string{strconv.Itoa(int(Interval.Seconds()))},
		"issuer": []string{issuer},
	}

	label := issuer
	if username != "" {
		label += ":" + username
	}
	u := &url.URL{
		Scheme:   "otpauth",
		Host:     "totp",
		Path:     "/" + label,
		RawQuery: query.Encode(),
	}
	return u.String(), nil
}
