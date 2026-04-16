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

// standardDomainFieldOrder defines the canonical field ordering for
// auto-inference from EIP712Domain pointer fields.
// Fields are included in this order if their corresponding pointer is non-nil.
var standardDomainFieldOrder = []struct {
	eipName string
	eipType string
	present func(*EIP712Domain) bool
	encode  func(*EIP712Domain) ([32]byte, error)
}{
	{
		"name", "string",
		func(d *EIP712Domain) bool { return d.Name != nil },
		func(d *EIP712Domain) ([32]byte, error) { return EncodeString(*d.Name), nil },
	},
	{
		"version", "string",
		func(d *EIP712Domain) bool { return d.Version != nil },
		func(d *EIP712Domain) ([32]byte, error) { return EncodeString(*d.Version), nil },
	},
	{
		"chainId", "uint256",
		func(d *EIP712Domain) bool { return d.ChainID != nil },
		func(d *EIP712Domain) ([32]byte, error) { return EncodeUint256(d.ChainID) },
	},
	{
		"verifyingContract", "address",
		func(d *EIP712Domain) bool { return d.VerifyingContract != nil },
		func(d *EIP712Domain) ([32]byte, error) { return EncodeAddress(*d.VerifyingContract), nil },
	},
	{
		"salt", "bytes32",
		func(d *EIP712Domain) bool { return d.Salt != nil },
		func(d *EIP712Domain) ([32]byte, error) { return EncodeBytes32(*d.Salt), nil },
	},
	{
		"chain_name", "string",
		func(d *EIP712Domain) bool { return d.ChainName != nil },
		func(d *EIP712Domain) ([32]byte, error) { return EncodeString(*d.ChainName), nil },
	},
	{
		"contract_package_hash", "bytes32",
		func(d *EIP712Domain) bool { return d.ContractPackageHash != nil },
		func(d *EIP712Domain) ([32]byte, error) { return EncodeBytes32(*d.ContractPackageHash), nil },
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
	var encoders []func(*EIP712Domain) ([32]byte, error)

	for _, f := range standardDomainFieldOrder {
		if f.present(&domain) {
			fields = append(fields, TypedField{Name: f.eipName, Type: f.eipType})
			encoders = append(encoders, f.encode)
		}
	}

	typeStr := BuildTypeString("EIP712Domain", fields)
	typeHash := ComputeTypeHash(typeStr)

	encoded := make([]byte, 0, 32*(1+len(fields)))
	encoded = append(encoded, typeHash[:]...)

	for _, enc := range encoders {
		slot, err := enc(&domain)
		if err != nil {
			return [32]byte{}, fmt.Errorf("eip712: domain field encoding: %w", err)
		}
		encoded = append(encoded, slot[:]...)
	}
	return Keccak256(encoded), nil
}

func hashDomainWithExplicitTypes(domain EIP712Domain, domainTypes []TypedField) ([32]byte, error) {
	typeStr := BuildTypeString("EIP712Domain", domainTypes)
	typeHash := ComputeTypeHash(typeStr)

	encoded := make([]byte, 0, 32*(1+len(domainTypes)))
	encoded = append(encoded, typeHash[:]...)

	// Build a lookup from field name -> encoder from standardDomainFieldOrder
	encByName := make(map[string]func(*EIP712Domain) ([32]byte, error), len(standardDomainFieldOrder))
	for _, f := range standardDomainFieldOrder {
		encByName[f.eipName] = f.encode
	}

	for _, tf := range domainTypes {
		enc, ok := encByName[tf.Name]
		if !ok {
			return [32]byte{}, fmt.Errorf("eip712: unknown domain field %q in explicit DomainTypes", tf.Name)
		}
		slot, err := enc(&domain)
		if err != nil {
			return [32]byte{}, fmt.Errorf("eip712: encoding domain field %q: %w", tf.Name, err)
		}
		encoded = append(encoded, slot[:]...)
	}
	return Keccak256(encoded), nil
}
