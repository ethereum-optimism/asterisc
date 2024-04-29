#!/bin/bash

./asterisc run \
  --info-at '%10000000' \
  --proof-at never \
  --input ./state.json \
  -- \
  ./op-program \
  --rollup.config ./chain-artifacts/rollup.json \
  --l2.genesis ./chain-artifacts/genesis-l2.json \
  --l1.trustrpc \
  --l1.rpckind debug_geth \
  --l1.head 0xce8f502bb58452ab4597292b354949016f65b255030d044f15c2290dcbf15b96 \
  --l2.head 0x721f4eb085f68959c0eb7df26cc7b6ec55873e8c8f0c987c09731a2d0fe9d774 \
  --l2.outputroot 0xa4d328501f15ad607dd796aede9d71c2124b7d6985a9f960a0333b2939e538a2 \
  --l2.claim 0x821c4841a99b02a501d2a6803f19737ee47039acbeb54b073f23bb3389bf075a \
  --l2.blocknumber 20 \
  --datadir ./preimages \
  --log.format terminal \
  --server