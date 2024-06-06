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
  --l1.head 0x2ba0c10c6a4dcdaa9983000be6c8304af390c04f35f1a10f5c87aa0657ad0fc3 \
  --l2.head 0x46899c0edbcee5d0b431091b0e4bdf534427e8410f84fa369a4e076cade68bd2 \
  --l2.outputroot 0xd41a99d72a07dd3fa23ad0564133bb761ee5b1423f0d72c1c44cfdbfe55706ee \
  --l2.claim 0x14a6cf9782f56c0ffa9f2af76540fdfec93f2c67a3790d85201d6cad229cafec \
  --l2.blocknumber 20 \
  --datadir ./preimages \
  --log.format terminal \
  --server