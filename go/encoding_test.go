package eip712_test

import (
	"math/big"
	"testing"

	eip712 "github.com/casper-ecosystem/casper-eip-712/go"
)

// --- EncodeAddress ---

func TestEncodeAddressEthLeftPads(t *testing.T) {
	var addr [20]byte
	for i := range addr {
		addr[i] = 0x11
	}
	slot := eip712.EncodeAddress(eip712.NewEthAddress(addr))
	// First 12 bytes must be zero
	for i := 0; i < 12; i++ {
		if slot[i] != 0 {
			t.Errorf("slot[%d] = %x, want 0", i, slot[i])
		}
	}
	// Bytes 12-31 must equal addr
	for i, b := range addr {
		if slot[12+i] != b {
			t.Errorf("slot[%d] = %x, want %x", 12+i, slot[12+i], b)
		}
	}
}

func TestEncodeAddressZeroEth(t *testing.T) {
	slot := eip712.EncodeAddress(eip712.NewEthAddress([20]byte{}))
	for i, b := range slot {
		if b != 0 {
			t.Errorf("slot[%d] = %x, want 0", i, b)
		}
	}
}

func TestEncodeAddressCasperAccountHash(t *testing.T) {
	var raw [33]byte
	raw[0] = 0x00
	for i := 1; i < 33; i++ {
		raw[i] = 0x11
	}
	slot := eip712.EncodeAddress(eip712.NewCasperAddress(raw))
	expected := eip712.Keccak256(raw[:])
	if slot != expected {
		t.Errorf("Casper AccountHash slot mismatch:\ngot  %s\nwant %s",
			eip712.ToHex(slot[:]), eip712.ToHex(expected[:]))
	}
}

func TestEncodeAddressCasperPackageHash(t *testing.T) {
	var raw [33]byte
	raw[0] = 0x01
	for i := 1; i < 33; i++ {
		raw[i] = 0x11
	}
	slot := eip712.EncodeAddress(eip712.NewCasperAddress(raw))
	expected := eip712.Keccak256(raw[:])
	if slot != expected {
		t.Errorf("Casper PackageHash slot mismatch:\ngot  %s\nwant %s",
			eip712.ToHex(slot[:]), eip712.ToHex(expected[:]))
	}
}

func TestEncodeAddressCasperAccountVsPackageDiffer(t *testing.T) {
	var account, pkg [33]byte
	for i := 1; i < 33; i++ {
		account[i] = 0x42
		pkg[i] = 0x42
	}
	account[0] = 0x00
	pkg[0] = 0x01
	if eip712.EncodeAddress(eip712.NewCasperAddress(account)) == eip712.EncodeAddress(eip712.NewCasperAddress(pkg)) {
		t.Error("AccountHash and PackageHash with same payload must produce different slots")
	}
}

// --- EncodeUint256 ---

func TestEncodeUint256Zero(t *testing.T) {
	slot, err := eip712.EncodeUint256(big.NewInt(0))
	if err != nil {
		t.Fatal(err)
	}
	for i, b := range slot {
		if b != 0 {
			t.Errorf("slot[%d] = %x, want 0", i, b)
		}
	}
}

func TestEncodeUint256One(t *testing.T) {
	slot, err := eip712.EncodeUint256(big.NewInt(1))
	if err != nil {
		t.Fatal(err)
	}
	if slot[31] != 1 {
		t.Errorf("slot[31] = %x, want 1", slot[31])
	}
	for i := 0; i < 31; i++ {
		if slot[i] != 0 {
			t.Errorf("slot[%d] = %x, want 0", i, slot[i])
		}
	}
}

func TestEncodeUint256NegativeErrors(t *testing.T) {
	_, err := eip712.EncodeUint256(big.NewInt(-1))
	if err == nil {
		t.Error("EncodeUint256(-1) expected error, got nil")
	}
}

func TestEncodeUint256Overflow(t *testing.T) {
	// 2^256 exceeds 32 bytes
	val := new(big.Int).Lsh(big.NewInt(1), 256)
	_, err := eip712.EncodeUint256(val)
	if err == nil {
		t.Error("EncodeUint256(2^256) expected error, got nil")
	}
}

func TestEncodeUint256FromHexString(t *testing.T) {
	// Test via EncodeField
	types := eip712.TypeDefinitions{}
	slot, err := eip712.EncodeField("uint256", "0x000000000000000000000000000000000000000000000000000000000000002a", types)
	if err != nil {
		t.Fatal(err)
	}
	if slot[31] != 0x2a {
		t.Errorf("slot[31] = %x, want 0x2a", slot[31])
	}
}

// --- EncodeUint64 ---

func TestEncodeUint64(t *testing.T) {
	slot := eip712.EncodeUint64(42)
	if slot[31] != 42 {
		t.Errorf("slot[31] = %d, want 42", slot[31])
	}
	for i := 0; i < 31; i++ {
		if slot[i] != 0 {
			t.Errorf("slot[%d] = %x, want 0", i, slot[i])
		}
	}
}

// --- EncodeString ---

func TestEncodeStringIsKeccak(t *testing.T) {
	slot := eip712.EncodeString("hello")
	expected := eip712.Keccak256([]byte("hello"))
	if slot != expected {
		t.Errorf("EncodeString(\"hello\") mismatch:\ngot  %s\nwant %s",
			eip712.ToHex(slot[:]), eip712.ToHex(expected[:]))
	}
}

func TestEncodeStringEmpty(t *testing.T) {
	slot := eip712.EncodeString("")
	expected := eip712.Keccak256([]byte{})
	if slot != expected {
		t.Error("EncodeString(\"\") must be keccak256 of empty bytes")
	}
}

// --- EncodeBytes ---

func TestEncodeBytes(t *testing.T) {
	data := []byte{0xde, 0xad, 0xbe, 0xef}
	slot := eip712.EncodeBytes(data)
	expected := eip712.Keccak256(data)
	if slot != expected {
		t.Errorf("EncodeBytes mismatch")
	}
}

// --- EncodeBytes32 ---

func TestEncodeBytes32(t *testing.T) {
	var b [32]byte
	b[0] = 0xaa
	b[31] = 0xbb
	slot := eip712.EncodeBytes32(b)
	if slot != b {
		t.Errorf("EncodeBytes32 must return the input unchanged")
	}
}

// --- EncodeBool ---

func TestEncodeBoolTrue(t *testing.T) {
	slot := eip712.EncodeBool(true)
	if slot[31] != 1 {
		t.Errorf("EncodeBool(true)[31] = %d, want 1", slot[31])
	}
	for i := 0; i < 31; i++ {
		if slot[i] != 0 {
			t.Errorf("EncodeBool(true)[%d] = %x, want 0", i, slot[i])
		}
	}
}

func TestEncodeBoolFalse(t *testing.T) {
	slot := eip712.EncodeBool(false)
	for i, b := range slot {
		if b != 0 {
			t.Errorf("EncodeBool(false)[%d] = %x, want 0", i, b)
		}
	}
}
