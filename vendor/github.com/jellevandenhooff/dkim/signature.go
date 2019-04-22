package dkim

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
)

// Signature describes a DKIM signature header
type Signature struct {
	// Signing domain
	Domain string

	canonHeader   string
	trimmedHeader string

	signature []byte
	bodyHash  []byte

	canon       *canon
	headerNames []string
	selector    string
	algo        *algo
}

func stripWhitespace(in string) string {
	var out []byte
	for _, c := range []byte(in) {
		if c != ' ' && c != '\t' && c != '\r' && c != '\n' {
			out = append(out, c)
		}
	}
	return string(out)
}

var dkimPrefix = "dkim-signature:"

func isSignatureHeader(header string) bool {
	return strings.HasPrefix(strings.ToLower(header), dkimPrefix)
}

func parseSignature(header string) (*Signature, error) {
	sig := new(Signature)

	var trimmedKVPairs []string
	var canonKVPairs []string
	for _, pair := range strings.Split(header[len(dkimPrefix):], ";") {
		idx := strings.IndexByte(pair, '=')
		if idx == -1 {
			trimmedKVPairs = append(trimmedKVPairs, pair)
			canonKVPairs = append(canonKVPairs, pair)
			continue
		}
		k, v := trimWhitespace(pair[:idx]), trimWhitespace(pair[idx+1:])

		var err error
		switch k {
		case "b":
			sig.signature, err = base64.StdEncoding.DecodeString(stripWhitespace(v))
			if err != nil {
				return nil, err
			}
		case "bh":
			sig.bodyHash, err = base64.StdEncoding.DecodeString(stripWhitespace(v))
			if err != nil {
				return nil, err
			}
		case "a":
			if a, found := algos[v]; found {
				sig.algo = a
			} else {
				return nil, errors.New("unknown algorithm")
			}
		case "c":
			if c, found := canons[v]; found {
				sig.canon = c
			} else {
				return nil, errors.New("unknown canon")
			}
		case "s":
			sig.selector = v
		case "d":
			sig.Domain = v
		case "h":
			sig.headerNames = strings.Split(v, ":")
			for i := range sig.headerNames {
				sig.headerNames[i] = strings.Trim(sig.headerNames[i], " \t\r\n")
			}
		default:
		}

		if k == "b" {
			trimmedKVPairs = append(trimmedKVPairs, pair[:idx+1])
			canonKVPairs = append(canonKVPairs, pair[:idx+1]+base64.StdEncoding.EncodeToString(sig.signature))
		} else {
			trimmedKVPairs = append(trimmedKVPairs, pair)
			canonKVPairs = append(canonKVPairs, pair)
		}
	}

	if sig.algo == nil {
		return nil, errors.New("missing algorithm")
	}
	if sig.canon == nil {
		return nil, errors.New("missing canon")
	}

	sig.trimmedHeader = header[:len(dkimPrefix)] + strings.Join(trimmedKVPairs, ";")
	sig.canonHeader = header[:len(dkimPrefix)] + strings.Join(canonKVPairs, ";")
	return sig, nil
}

func (s *Signature) txtRecordName() string {
	return fmt.Sprintf("%s._domainkey.%s.", string(s.selector), string(s.Domain))
}
