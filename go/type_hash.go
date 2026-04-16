package eip712

// ComputeTypeHash returns keccak256 of the UTF-8 encoded type string.
// Input format: "TypeName(field1Type field1Name,...)"
func ComputeTypeHash(typeString string) [32]byte {
	return Keccak256([]byte(typeString))
}
