#!/bin/bash

./asterisc run \
  --info-at '%10000000' \
  --proof-at never \
  --input ./test-data/state.bin.gz \
  --meta ./test-data/meta.json \
  -- \
  ./op-program \
  --rollup.config ./test-data/chain-artifacts/rollup.json \
  --l2.genesis ./test-data/chain-artifacts/genesis-l2.json \
  --l1.trustrpc \
  --l1.rpckind debug_geth \
  --l1.head 0xd751c9ae2912d3ab61e8ed0994538d87eb19344548b9ea4fa69cc93ba18c65cf \
  --l2.head 0x605247ac833b4df114a65df0cf5a94caf5c851e872b277956982a6cf6be9fc9e \
  --l2.outputroot 0x7021b5ba5813fa0369a5d8840c52975eaaeb217fa6e04b44a646375ef32867a9 \
  --l2.claim 0xe13d5c2d0e8264d76ed13760abf3a20d047e2b69ca22dbd354abd2e7a7cb4966 \
  --l2.blocknumber 560 \
  --datadir ./test-data/preimages \
  --log.format terminal \
  --server