package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAttribute_IsPrivate(t *testing.T) {
	publicAsset := NewAsset("contoso.com", "18.1.2.4")
	privateAsset := NewAsset("contoso.local", "10.0.0.1")

	tests := []struct {
		name string
		attr Attribute
		want bool
	}{
		{
			name: "public https",
			attr: NewAttribute("https", "443", &publicAsset),
			want: false,
		},
		{
			name: "private https",
			attr: NewAttribute("https", "443", &privateAsset),
			want: true,
		},
		{
			name: "public port",
			attr: NewAttribute("port", "443", &publicAsset),
			want: false,
		},
		{
			name: "private port",
			attr: NewAttribute("port", "443", &privateAsset),
			want: true,
		},
	}

	for _, tc := range tests {
		asset := tc.attr.Asset()
		actual := asset.IsPrivate()
		assert.Equal(t, tc.want, actual, "test case %s failed: expected %t, got %t", tc.name, tc.want, actual)
	}
}
