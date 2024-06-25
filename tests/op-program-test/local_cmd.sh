#!/bin/bash

./asterisc run \
  --info-at '%10000000' \
  --proof-at never \
  --input ./test-data/state.json \
  --meta ./test-data/meta.json \
  -- \
  ./op-program \
  --rollup.config ./test-data/chain-artifacts/rollup.json \
  --l2.genesis ./test-data/chain-artifacts/genesis-l2.json \
  --l1.trustrpc \
  --l1.rpckind debug_geth \
  --l1.head 0xde3fdc5ac10696af68240e1b503b2adc9546fad649142b0cb8bc77eee0dd4bf8 \
  --l2.head 0xeebb4bfeb240e7eb927277eb3055a64e640764b612d3074582ce815e0cc5f2ef \
  --l2.outputroot 0x628b7f5289f8d67203b36f14474eb616a5a5b7a97d9636d125eb142ffe463d12 \
  --l2.claim 0x6546c2b1d660ac6b1e8a2e2bc68d4b992a0508e3ff481d0b68988dd9423a47f4 \
  --l2.blocknumber 13 \
  --datadir ./test-data/preimages \
  --log.format terminal \
  --server