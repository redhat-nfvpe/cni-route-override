package dkim

import (
	"bytes"
	"strings"
)

// A canon is pair of canonicalization algorithms as defined in the RFC 6376.
type canon struct {
	header func(string) string
	body   func(string) string
}

// 3.4.1.  The "simple" Header Canonicalization Algorithm
//
//   The "simple" header canonicalization algorithm does not change header
//   fields in any way.  Header fields MUST be presented to the signing or
//   verification algorithm exactly as they are in the message being
//   signed or verified.  In particular, header field names MUST NOT be
//   case folded and whitespace MUST NOT be changed.
//
func simpleHeader(x string) string {
	return x
}

// 3.4.2.  The "relaxed" Header Canonicalization Algorithm
//
//    The "relaxed" header canonicalization algorithm MUST apply the
//    following steps in order:
//
//    o  Convert all header field names (not the header field values) to
//       lowercase.  For example, convert "SUBJect: AbC" to "subject: AbC".
//
//    o  Unfold all header field continuation lines as described in
//       [RFC5322]; in particular, lines with terminators embedded in
//       continued header field values (that is, CRLF sequences followed by
//       WSP) MUST be interpreted without the CRLF.  Implementations MUST
//       NOT remove the CRLF at the end of the header field value.
//
//    o  Convert all sequences of one or more WSP characters to a single SP
//       character.  WSP characters here include those before and after a
//       line folding boundary.
//
//    o  Delete all WSP characters at the end of each unfolded header field
//       value.
//
//    o  Delete any WSP characters remaining before and after the colon
//       separating the header field name from the header field value.  The
//       colon separator MUST be retained.
//
func relaxHeader(header string) string {
	// track if we have processed the ':' yet
	pastFieldName := false
	// seenWsp tracks if we are currently in whitespace sequence
	seenWsp := false
	// eatWsp tracks if the current whitespace, if any, should be killed
	eatWsp := false

	var relaxed []byte
	for _, c := range []byte(header) {
		// figure out how to print c, if at all

		if c == '\r' || c == '\n' {
			// unfold all CRLFs
			continue
		} else if c == '\t' || c == ' ' {
			if !eatWsp {
				// track that we have seen a space for later printing
				seenWsp = true
			}
		} else {
			// hit a character, so we can stop killing whitespace
			eatWsp = false

			if !pastFieldName {
				if c == ':' {
					// retroactive kill any whitespace before the ':'
					seenWsp = false
					// kill whitespace after the ':'
					eatWsp = true
					// and mark that we have passed the ':'
					pastFieldName = true
				} else if 'A' <= c && c <= 'Z' {
					// convert header name to lowercase
					c += 'a' - 'A'
				}
			}

			if seenWsp {
				// output collapsed whitespace if we have seen any
				relaxed = append(relaxed, ' ')
				seenWsp = false
			}
			// output current character
			relaxed = append(relaxed, c)
		}
	}

	if !isSignatureHeader(header) {
		// restore CRLF, but not the signature header
		relaxed = append(relaxed, []byte("\r\n")...)
	}

	return string(relaxed)
}

// 3.4.3.  The "simple" Body Canonicalization Algorithm
//
//    The "simple" body canonicalization algorithm ignores all empty lines
//    at the end of the message body.  An empty line is a line of zero
//    length after removal of the line terminator.  If there is no body or
//    no trailing CRLF on the message body, a CRLF is added.  It makes no
//    other changes to the message body.  In more formal terms, the
//    "simple" body canonicalization algorithm converts "*CRLF" at the end
//    of the body to a single "CRLF".
//
//    Note that a completely empty or missing body is canonicalized as a
//    single "CRLF"; that is, the canonicalized length will be 2 octets.
//
func simpleBody(body string) string {
	// only kill CRLFs if there are at least two; that saves us from having to
	// add a CRLF in the common case
	for strings.HasSuffix(body, "\r\n\r\n") {
		body = strings.TrimSuffix(body, "\r\n")
	}
	// if there are no CRLFs at all, add one
	if !strings.HasSuffix(body, "\r\n") {
		body += "\r\n"
	}
	return body
}

// 3.4.4.  The "relaxed" Body Canonicalization Algorithm
//
//    The "relaxed" body canonicalization algorithm MUST apply the
//    following steps (a) and (b) in order:
//
//    a.  Reduce whitespace:
//
//        *  Ignore all whitespace at the end of lines.  Implementations
//           MUST NOT remove the CRLF at the end of the line.
//
//        *  Reduce all sequences of WSP within a line to a single SP
//           character.
//
//    b.  Ignore all empty lines at the end of the message body.  "Empty
//        line" is defined in Section 3.4.3.  If the body is non-empty but
//        does not end with a CRLF, a CRLF is added.  (For email, this is
//        only possible when using extensions to SMTP or non-SMTP transport
//        mechanisms.)
//
func relaxBody(body string) string {
	var relaxed []byte
	// part a, collapse whitespace line-by-line
	for _, line := range bytes.Split([]byte(body), []byte("\r\n")) {
		// seenWsp tracks if we are currently in whitespace sequence
		seenWsp := false
		for _, c := range line {
			if c == '\t' || c == ' ' {
				// track that we have seen a space for later printing
				seenWsp = true
			} else {
				if seenWsp {
					// output collapsed whitespace if we have seen any
					relaxed = append(relaxed, ' ')
					seenWsp = false
				}
				relaxed = append(relaxed, c)
			}
		}
		// add a newline after every line (even the last line that
		// might not have had one, for b.)
		relaxed = append(relaxed, []byte("\r\n")...)
	}
	// part b, kill empty lines at end of body (but keep last line's newline)
	body = string(relaxed)
	for strings.HasSuffix(body, "\r\n\r\n") {
		body = strings.TrimSuffix(body, "\r\n")
	}
	if body == "\r\n" {
		// if there were no lines at all, then kill the last line as well
		body = ""
	}
	return body
}

var canons = map[string]*canon{
	"simple/simple":   {header: simpleHeader, body: simpleBody},
	"simple/relaxed":  {header: simpleHeader, body: relaxBody},
	"relaxed/simple":  {header: relaxHeader, body: simpleBody},
	"relaxed/relaxed": {header: relaxHeader, body: relaxBody},
	"simple":          {header: simpleHeader, body: simpleBody},
	"relaxed":         {header: relaxHeader, body: relaxBody},
}
