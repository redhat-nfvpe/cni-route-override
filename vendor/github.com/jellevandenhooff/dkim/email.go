package dkim

type email struct {
	headers []string
	body    string
}

func readHeader(data string) (header string, rem string) {
	seenLf := false
	for i, c := range []byte(data) {
		if seenLf {
			if c != ' ' && c != '\t' {
				return data[:i], data[i:]
			}
		}
		seenLf = c == '\n'
	}
	return data, ""
}

func parseEmail(mail string) *email {
	var headers []string

	for len(mail) > 0 {
		var header string
		header, mail = readHeader(mail)
		if header == "\r\n" {
			break
		}

		headers = append(headers, header)
	}

	return &email{
		headers: headers,
		body:    mail,
	}
}
