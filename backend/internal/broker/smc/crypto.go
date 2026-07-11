package smc

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

func (c *Client) GenerateSignature(
	requestToken string,
) string {

	key := c.apiKey + requestToken

	hash := hmac.New(
		sha256.New,
		[]byte(key),
	)

	hash.Write([]byte(c.apiSecret))

	return hex.EncodeToString(
		hash.Sum(nil),
	)

}
