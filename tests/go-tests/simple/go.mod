module simple

go 1.21

toolchain go1.21.1

require (
	github.com/ethereum-optimism/optimism v1.1.5-rc.1.0.20230922205554-a7ff5a811612
	github.com/ethereum/go-ethereum v1.12.2
)

require (
	github.com/btcsuite/btcd/btcec/v2 v2.2.0 // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.2.0 // indirect
	github.com/holiman/uint256 v1.2.3 // indirect
	golang.org/x/crypto v0.13.0 // indirect
	golang.org/x/sys v0.12.0 // indirect
)

replace github.com/ethereum/go-ethereum v1.12.2 => github.com/ethereum-optimism/op-geth v1.101200.2-rc.1.0.20230922185314-7997a6fed17c
