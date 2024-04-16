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
  --l1.head 0xf6d5757b17cb460172f6704f65cd8547650af07324491fbc6740b09ebcd83439 \
  --l2.head 0x103cf407e9ed858485765c6cd9289a97d7ca24d749985fd71e5ec0938443f41f \
  --l2.outputroot 0x85ea67fc439b9be0ae6ac6a51b438a423f7a9ebd300745a5bbbf4cb5effe586d \
  --l2.claim 0x65a7f953c712dd7ce0833dca0e58ef3937e337740be72e3b4810d2889db91787 \
  --l2.blocknumber 20 \
  --datadir ./preimages \
  --log.format terminal \
  --server