package dkim

import (
	"encoding/base64"
	"strings"
)

type pubkey struct {
	key []byte
}

func trimWhitespace(in string) string {
	return strings.Trim(in, "\r\n\t ")
}

func parsePubkey(txtRecord string) *pubkey {
	pubkey := new(pubkey)

	for _, pair := range strings.Split(txtRecord, ";") {
		idx := strings.IndexByte(pair, '=')
		if idx == -1 {
			continue
		}
		k, v := trimWhitespace(pair[:idx]), trimWhitespace(pair[idx+1:])

		switch k {
		case "p":
			pubkey.key, _ = base64.StdEncoding.DecodeString(stripWhitespace(v))
		}
	}

	return pubkey
}
