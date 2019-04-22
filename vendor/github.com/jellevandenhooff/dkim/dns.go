package dkim

// A DNSClient can look up TXT records.
type DNSClient interface {
	LookupTxt(hostname string) ([]string, error)
}
