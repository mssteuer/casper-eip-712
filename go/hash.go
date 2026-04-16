package eip712

import "fmt"

// HashStruct computes the EIP-712 struct hash for a typed message.
//
//	keccak256(typeHash || encodeData(field1) || encodeData(field2) || ...)
func HashStruct(primaryType string, types TypeDefinitions, message map[string]interface{}) ([32]byte, error) {
	fields, ok := types[primaryType]
	if !ok {
		return [32]byte{}, fmt.Errorf("eip712: type %q not found in TypeDefinitions", primaryType)
	}

	typeStr, err := BuildCanonicalTypeString(primaryType, types)
	if err != nil {
		return [32]byte{}, err
	}
	typeHash := ComputeTypeHash(typeStr)

	encoded := make([]byte, 0, 32*(1+len(fields)))
	encoded = append(encoded, typeHash[:]...)

	for _, field := range fields {
		value := message[field.Name]
		slot, err := EncodeField(field.Type, value, types)
		if err != nil {
			return [32]byte{}, fmt.Errorf("eip712: encoding field %q (%s): %w", field.Name, field.Type, err)
		}
		encoded = append(encoded, slot[:]...)
	}

	return Keccak256(encoded), nil
}

// HashTypedData computes the full EIP-712 digest.
//
//	keccak256(0x19 || 0x01 || domainSeparator || structHash)
func HashTypedData(
	domain EIP712Domain,
	types TypeDefinitions,
	primaryType string,
	message map[string]interface{},
	opts *TypedDataOptions,
) ([32]byte, error) {
	domainSep, err := HashDomainSeparator(domain, opts)
	if err != nil {
		return [32]byte{}, fmt.Errorf("eip712: domain separator: %w", err)
	}

	structHash, err := HashStruct(primaryType, types, message)
	if err != nil {
		return [32]byte{}, fmt.Errorf("eip712: struct hash: %w", err)
	}

	return assembleDigest(domainSep, structHash), nil
}

// HashTypedDataRaw is the low-level variant accepting pre-computed inputs.
// typeHash is the 32-byte type hash; encodedStruct is the ABI-encoded struct
// fields WITHOUT the leading typeHash (HashStruct prepends it internally).
func HashTypedDataRaw(
	domain EIP712Domain,
	typeHash [32]byte,
	encodedStruct []byte,
	opts *TypedDataOptions,
) ([32]byte, error) {
	domainSep, err := HashDomainSeparator(domain, opts)
	if err != nil {
		return [32]byte{}, fmt.Errorf("eip712: domain separator: %w", err)
	}

	// Struct hash = keccak256(typeHash || encodedStruct)
	raw := make([]byte, 0, 32+len(encodedStruct))
	raw = append(raw, typeHash[:]...)
	raw = append(raw, encodedStruct...)
	structHash := Keccak256(raw)

	return assembleDigest(domainSep, structHash), nil
}

// assembleDigest builds the EIP-191 encoded digest.
// Format: keccak256(0x19 || 0x01 || domainSeparator || structHash)
func assembleDigest(domainSep, structHash [32]byte) [32]byte {
	buf := make([]byte, 66)
	buf[0] = 0x19
	buf[1] = 0x01
	copy(buf[2:34], domainSep[:])
	copy(buf[34:66], structHash[:])
	return Keccak256(buf)
}
