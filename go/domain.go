package eip712

import "fmt"

// CasperDomainTypes is the canonical Casper-native domain type schema.
// Pass this in TypedDataOptions.DomainTypes when using BuildDomain.
var CasperDomainTypes = []TypedField{
	{Name: "name", Type: "string"},
	{Name: "version", Type: "string"},
	{Name: "chain_name", Type: "string"},
	{Name: "contract_package_hash", Type: "bytes32"},
}

// standardDomainFieldOrder defines the canonical field ordering and default
// types for auto-inference from EIP712Domain pointer fields.
//
// The `eipType` here is only a default for the auto-infer path; when the caller
// supplies an explicit DomainTypes list, encoding is dispatched from that
// declared type (not from the name), so a caller can legitimately declare e.g.
// chainId as uint64 if their verifier expects it.
var standardDomainFieldOrder = []struct {
	eipName  string
	eipType  string
	present  func(*EIP712Domain) bool
	getValue func(*EIP712Domain) interface{} // only called when present returns true
}{
	{
		"name", "string",
		func(d *EIP712Domain) bool { return d.Name != nil },
		func(d *EIP712Domain) interface{} { return *d.Name },
	},
	{
		"version", "string",
		func(d *EIP712Domain) bool { return d.Version != nil },
		func(d *EIP712Domain) interface{} { return *d.Version },
	},
	{
		"chainId", "uint256",
		func(d *EIP712Domain) bool { return d.ChainID != nil },
		func(d *EIP712Domain) interface{} { return d.ChainID },
	},
	{
		"verifyingContract", "address",
		func(d *EIP712Domain) bool { return d.VerifyingContract != nil },
		func(d *EIP712Domain) interface{} { return *d.VerifyingContract },
	},
	{
		"salt", "bytes32",
		func(d *EIP712Domain) bool { return d.Salt != nil },
		func(d *EIP712Domain) interface{} { return *d.Salt },
	},
	{
		"chain_name", "string",
		func(d *EIP712Domain) bool { return d.ChainName != nil },
		func(d *EIP712Domain) interface{} { return *d.ChainName },
	},
	{
		"contract_package_hash", "bytes32",
		func(d *EIP712Domain) bool { return d.ContractPackageHash != nil },
		func(d *EIP712Domain) interface{} { return *d.ContractPackageHash },
	},
}

// BuildDomain constructs a Casper-native EIP712Domain.
// Use with TypedDataOptions{DomainTypes: CasperDomainTypes} when hashing.
func BuildDomain(name, version, chainName string, contractPackageHash [32]byte) EIP712Domain {
	return EIP712Domain{
		Name:                &name,
		Version:             &version,
		ChainName:           &chainName,
		ContractPackageHash: &contractPackageHash,
	}
}

// BuildDomainTypeString generates the canonical EIP712Domain type string.
//
//   - If opts is non-nil and opts.DomainTypes is set, those fields are used verbatim.
//   - Otherwise, fields are inferred from which EIP712Domain pointer fields are non-nil,
//     in the canonical order defined by standardDomainFieldOrder.
func BuildDomainTypeString(domain EIP712Domain, opts *TypedDataOptions) string {
	fields := resolveDomainFields(domain, opts)
	return BuildTypeString("EIP712Domain", fields)
}

// HashDomainSeparator computes the 32-byte domain separator hash.
//
//	keccak256(typeHash || field1Encoded || field2Encoded || ...)
func HashDomainSeparator(domain EIP712Domain, opts *TypedDataOptions) ([32]byte, error) {
	if opts != nil && opts.DomainTypes != nil {
		return hashDomainWithExplicitTypes(domain, opts.DomainTypes)
	}
	return hashDomainAutoInfer(domain)
}

func resolveDomainFields(domain EIP712Domain, opts *TypedDataOptions) []TypedField {
	if opts != nil && opts.DomainTypes != nil {
		return opts.DomainTypes
	}
	var fields []TypedField
	for _, f := range standardDomainFieldOrder {
		if f.present(&domain) {
			fields = append(fields, TypedField{Name: f.eipName, Type: f.eipType})
		}
	}
	return fields
}

func hashDomainAutoInfer(domain EIP712Domain) ([32]byte, error) {
	var fields []TypedField
	var values []interface{}

	for _, f := range standardDomainFieldOrder {
		if f.present(&domain) {
			fields = append(fields, TypedField{Name: f.eipName, Type: f.eipType})
			values = append(values, f.getValue(&domain))
		}
	}
	return encodeDomain(fields, values)
}

func hashDomainWithExplicitTypes(domain EIP712Domain, domainTypes []TypedField) ([32]byte, error) {
	// Index the canonical field table by name for O(1) lookup.
	indexByName := make(map[string]int, len(standardDomainFieldOrder))
	for i, f := range standardDomainFieldOrder {
		indexByName[f.eipName] = i
	}

	values := make([]interface{}, len(domainTypes))
	for i, tf := range domainTypes {
		idx, ok := indexByName[tf.Name]
		if !ok {
			return [32]byte{}, fmt.Errorf("eip712: unknown domain field %q in explicit DomainTypes", tf.Name)
		}
		f := standardDomainFieldOrder[idx]
		if !f.present(&domain) {
			return [32]byte{}, fmt.Errorf("eip712: domain field %q declared in DomainTypes but not set on EIP712Domain", tf.Name)
		}
		values[i] = f.getValue(&domain)
	}
	return encodeDomain(domainTypes, values)
}

// encodeDomain dispatches encoding from the declared field type (not the name),
// so a mismatch between declared type and the domain field's actual value —
// e.g. {Name: "chainId", Type: "bytes32"} against a *big.Int ChainID — is
// caught by EncodeField rather than silently producing an inconsistent digest.
func encodeDomain(fields []TypedField, values []interface{}) ([32]byte, error) {
	typeStr := BuildTypeString("EIP712Domain", fields)
	typeHash := ComputeTypeHash(typeStr)

	encoded := make([]byte, 0, 32*(1+len(fields)))
	encoded = append(encoded, typeHash[:]...)
	for i, tf := range fields {
		slot, err := EncodeField(tf.Type, values[i], nil)
		if err != nil {
			return [32]byte{}, fmt.Errorf("eip712: encoding domain field %q as %s: %w", tf.Name, tf.Type, err)
		}
		encoded = append(encoded, slot[:]...)
	}
	return Keccak256(encoded), nil
}
