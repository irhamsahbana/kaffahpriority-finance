package security

import (
	"codebase-app/internal/infrastructure/config"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"strconv"
	"time"
)

type SignedURL struct {
	Link      string `json:"link"`
	Expires   int64  `json:"expires"`
	Signature string `json:"signature"`
}

func GenerateSignedURL(link string, expiration time.Duration) (l SignedURL, err error) {
	var (
		key            = []byte(config.Envs.Guard.JwtPrivateKey)
		expirationTime = time.Now().UTC().Add(expiration).Unix()
		data           = fmt.Sprintf("%s%d", link, expirationTime)
	)

	// Create a new HMAC by defining the hash type and the key (as byte array)
	h := hmac.New(sha256.New, key)
	h.Write([]byte(data))
	signature := hex.EncodeToString(h.Sum(nil))

	// Add the expiration time and signature to the URL
	u, _ := url.Parse(link)
	q := u.Query()
	q.Set("expires", strconv.FormatInt(expirationTime, 10))
	q.Set("signature", signature)
	u.RawQuery = q.Encode()

	l.Link = u.String()
	l.Expires = expirationTime
	l.Signature = signature

	return l, nil
}
