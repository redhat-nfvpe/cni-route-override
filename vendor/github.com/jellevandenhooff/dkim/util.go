package dkim

import "strings"

// CanonHeaders computes the canonical form of the underlying e-mail in
// headers-only format. The result is a valid DKIM headers-only e-mail.
func (v *VerifiedEmail) CanonHeaders() string {
	var canonHeaders []string
	for _, header := range append(v.Headers, v.Signature.canonHeader) {
		canonHeaders = append(canonHeaders, v.Signature.canon.header(header))
	}
	return strings.Join(canonHeaders, "")
}

// ExtractHeader retrieves the named header from the signed headers. Returns
// an array of headers since headers can appear multiple times.
func (v *VerifiedEmail) ExtractHeader(name string) []string {
	return extractHeaders(v.Headers, []string{name})
}
