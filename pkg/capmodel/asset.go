package capmodel

// NewIPAsset creates an IP capability model from an IP address string.
// Both dns and name are set to the IP address value.
func NewIPAsset(ip string) IP {
	return IP{DNS: ip}
}
