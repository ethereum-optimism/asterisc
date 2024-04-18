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
  --l1.head 0x8190a9b94dc936005a939ec2b4d16b6e8e93452a91ad36003e4115cab78cd1f6 \
  --l2.head 0xa8fbd5da3f0d1fb132d481bae72850abbb623c314120f72bdc79d65d3729dd08 \
  --l2.outputroot 0xe21c0dc811e8de2357f48377c1269348814afee8735cfb44f94774b65fee1987 \
  --l2.claim 0x871310050613a9ece416c30bf2ecaa3f16acdcca533e12f858f29bfc4b4f29f4 \
  --l2.blocknumber 280 \
  --datadir ./preimages \
  --log.format terminal \
  --server