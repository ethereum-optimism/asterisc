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
  --l1.head 0x52a2f6b4e022b51f00c262926f43383a216a957c3c3b9463b15866663da8bf44 \
  --l2.head 0x4f0fe0525f90bf9012628a1dd1213b24751862005aa78d6b417897c56f88e440 \
  --l2.outputroot 0x6035a4f175d1e45fded771eec8306357ccabe37de26abcdb51711395543305a4 \
  --l2.claim 0x2d2db457102d783867103b7e7e6e1dce9e5bd6465f44ecd5e870c721b93be429 \
  --l2.blocknumber 13 \
  --datadir ./preimages \
  --log.format terminal \
  --server