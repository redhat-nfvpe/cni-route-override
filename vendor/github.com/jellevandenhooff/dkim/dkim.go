// Package dkim is a pure Go DKIM verification library.
//
// The main use-case for the dkim package is verifying DKIM "ownership proofs" of
// e-mail addresses. The owner of an e-mail address can sent DKIM-signed
// e-mails, and anyone in the world can verify that those e-mail were sent
// by the owner by checking the DKIM signature. The dkim package exposes an API
// to verify DKIM signatures, and to inspect the signed headers of the e-mail.
//
// As an added feature, the dkim package supports extracting and verifying just
// the signed headers of an e-mail. This makes ownership proofs much shorter.
package dkim

import (
	"bytes"
	"errors"
	"strings"
)

// VerificationMode configures the DKIM verification algorithm.
type VerificationMode int

const (
	// Complete indicates that both headers and body should be verified.
	Complete VerificationMode = iota
	// HeadersOnly indicates that only headers should be verified.
	HeadersOnly
)

// VerifiedEmail stores the verified parts of an e-mail.
type VerifiedEmail struct {
	// Information on signature
	Signature *Signature

	// Signed headers (in original form)
	Headers []string

	email *email
}

func extractHeaders(headers []string, names []string) []string {
	byName := make(map[string][]string)
	for _, header := range headers {
		nameEnd := strings.Index(header, ":")
		if nameEnd == -1 {
			nameEnd = len(header)
		}
		name := strings.ToLower(header[:nameEnd])
		byName[name] = append(byName[name], header)
	}

	var extracted []string
	for _, name := range names {
		headers := byName[name]
		if len(headers) > 0 {
			extracted = append(extracted, headers[len(headers)-1])
			byName[name] = headers[:len(headers)-1]
		}
	}

	return extracted
}

// ParseAndVerify parses the e-mail, searches for a DKIM signature, verifies
// the signature, and returns the resulting verified e-mail. Results in an
// error if there is no valid signature.
func ParseAndVerify(mail string, mode VerificationMode, dnsClient DNSClient) (*VerifiedEmail, error) {
	email := parseEmail(mail)

	var signatureHeader string
	for _, header := range email.headers {
		// we don't support DKIM-Signature headers signing other DKIM-Signature
		// headers
		if isSignatureHeader(header) {
			if signatureHeader != "" {
				return nil, errors.New("multiple DKIM headers")
			}
			signatureHeader = header
		}
	}
	if signatureHeader == "" {
		return nil, errors.New("no DKIM header found")
	}

	signature, err := parseSignature(signatureHeader)
	if err != nil {
		return nil, err
	}

	if mode == Complete {
		h := signature.algo.hasher()
		body := signature.canon.body(email.body)
		h.Write([]byte(body))

		if !bytes.Equal(signature.bodyHash, h.Sum(nil)) {
			return nil, errors.New("body hash does not match")
		}
	}

	txtRecords, err := dnsClient.LookupTxt(signature.txtRecordName())
	if err != nil {
		return nil, err
	}

	signedHeaders := extractHeaders(email.headers, signature.headerNames)

	h := signature.algo.hasher()
	for _, header := range signedHeaders {
		header = signature.canon.header(header)
		h.Write([]byte(header))
	}
	header := signature.canon.header(signature.trimmedHeader)
	h.Write([]byte(header))

	headersHash := h.Sum(nil)

	found := false
	for _, txtRecord := range txtRecords {
		pubkey := parsePubkey(txtRecord)
		if err := signature.algo.checkSig(pubkey.key, headersHash, signature.signature); err == nil {
			found = true
		}
	}
	if !found {
		return nil, errors.New("no valid DKIM signature")
	}

	return &VerifiedEmail{
		email:     email,
		Signature: signature,
		Headers:   signedHeaders,
	}, nil
}
