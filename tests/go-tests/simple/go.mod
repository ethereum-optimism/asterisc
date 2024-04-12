module simple

go 1.21

toolchain go1.21.1

require (
	github.com/ethereum-optimism/optimism v1.7.3
	github.com/ethereum/go-ethereum v1.13.11
)

require (
	github.com/btcsuite/btcd/btcec/v2 v2.3.3 // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.3.0 // indirect
	github.com/holiman/uint256 v1.2.4 // indirect
	golang.org/x/crypto v0.22.0 // indirect
	golang.org/x/sys v0.19.0 // indirect
)

replace github.com/ethereum/go-ethereum v1.13.11 => github.com/ethereum-optimism/op-geth v1.101311.0
