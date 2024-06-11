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
  --l1.head 0xd08fb3fa90d28261be12ad0863bd3268baf8629066901bd5b75a2a49f0c034a0 \
  --l2.head 0xee0ccbbd6958805c7db525504eb3c5da3d8b467553fd6431996f72cef9397164 \
  --l2.outputroot 0x4c746d0a08fffc6652dbb8800daddb058dd10f5a145c6fa651d301d9b72a20d8 \
  --l2.claim 0xde14084b8fc8cfbffc5038648e0b7da3514faa716b9d1dc0dc040bd48bc17c35 \
  --l2.blocknumber 8 \
  --datadir ./preimages \
  --log.format terminal \
  --server