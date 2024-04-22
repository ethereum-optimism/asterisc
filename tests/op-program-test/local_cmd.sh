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
  --l1.head 0x722330dbad7965f62d1ca187f1f4468fadc57669613b9db17321f2514882ff45 \
  --l2.head 0x47e53c1c90ff2e52f3b7eb5121d655499ce714f65d6de3856df122c762466323 \
  --l2.outputroot 0xd6db4d9b8e1fb700a8c8f36c01dabea14713d99989ffcca12074bbd564071515 \
  --l2.claim 0xa84525e49f8c8ef78a50ed6c9185933257b26628d478ddd3cd63ceaa79ddd01b \
  --l2.blocknumber 30 \
  --datadir ./preimages \
  --log.format terminal \
  --server