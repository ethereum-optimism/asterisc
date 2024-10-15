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
  --l1.head 0xe5e936e8b01ff9084ffc91f3d6944238df0ed1bc58ee56f569c56169a4e70b52 \
  --l2.head 0x8bc9297b05324efc1f9eb48e2dcb70a85ac8ec1a25bdcbe2c4c7ea7db4b15160 \
  --l2.outputroot 0x9faa74acedf717cc0362a79beec59013cf27fce4de146d7e2d941b16b50901f3 \
  --l2.claim 0x7dbae010376c4ae02877fd146cd9b9ef3209c8a103d977d415130f725ee6b7eb \
  --l2.blocknumber 12 \
  --datadir ./test-data/preimages \
  --log.format terminal \
  --server