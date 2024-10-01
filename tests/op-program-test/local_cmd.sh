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
  --l1.head 0x32230088dbbc84c096d91e7fe67df0e18451f47ec8e5170911a4aa77e7d8bfc5 \
  --l2.head 0x4550a5cdaae0fb257a4034af8c2b6c20263319b56672545d6917c0785972a72e \
  --l2.outputroot 0x8f6100bd0f05fefed78e7d94b9f7397ce98a0f9fe515ae25e2a315a9fa0520a4 \
  --l2.claim 0xe508b8b3fe23fc4d3372c77a5896d6ea1177bf12ef8ff2a1ba6d22ac4b9447da \
  --l2.blocknumber 12 \
  --datadir ./test-data/preimages \
  --log.format terminal \
  --server