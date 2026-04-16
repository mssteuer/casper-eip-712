package eip712_test

import (
	"bytes"
	"testing"

	eip712 "github.com/casper-ecosystem/casper-eip-712/go"
)

func TestToHex(t *testing.T) {
	tests := []struct {
		input []byte
		want  string
	}{
		{[]byte{}, "0x"},
		{[]byte{0x12, 0x34}, "0x1234"},
		{[]byte{0xde, 0xad, 0xbe, 0xef}, "0xdeadbeef"},
	}
	for _, tt := range tests {
		got := eip712.ToHex(tt.input)
		if got != tt.want {
			t.Errorf("ToHex(%v) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestFromHex(t *testing.T) {
	tests := []struct {
		input string
		want  []byte
	}{
		{"0x1234", []byte{0x12, 0x34}},
		{"1234", []byte{0x12, 0x34}},
		{"0xdeadbeef", []byte{0xde, 0xad, 0xbe, 0xef}},
		{"0x", []byte{}},
	}
	for _, tt := range tests {
		got, err := eip712.FromHex(tt.input)
		if err != nil {
			t.Errorf("FromHex(%q) unexpected error: %v", tt.input, err)
			continue
		}
		if !bytes.Equal(got, tt.want) {
			t.Errorf("FromHex(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestFromHexOddLength(t *testing.T) {
	_, err := eip712.FromHex("0x123")
	if err == nil {
		t.Error("FromHex(odd-length) expected error, got nil")
	}
}
