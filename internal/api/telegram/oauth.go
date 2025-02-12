package telegram

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"sort"
	"strings"
)

func ParseDataCheckString(values url.Values) string {
	pairs := []string{}
	for key, value := range values {
		if key == "hash" {
			continue
		}

		if len(value) > 0 {
			pairs = append(pairs, fmt.Sprintf("%s=%s", key, value[0]))
		}
	}

	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i] < pairs[j]
	})

	return strings.Join(pairs, "\n")
}

func CalculateHash(values url.Values, token string) string {
	hash := values.Get("hash")
	if len(hash) == 0 {
		return "expected `hash` to be passed in url query"
	}

	h := sha256.New()
	h.Write([]byte(token))
	secret := h.Sum(nil)

	sig := hmac.New(sha256.New, secret)
	sig.Write([]byte(ParseDataCheckString(values)))

	return hex.EncodeToString(sig.Sum(nil))
}
