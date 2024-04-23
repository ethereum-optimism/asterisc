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
  --l1.head 0xdbfb0caff9c3c3d75ed48b7ff7a6eeb26fdb5329e669e9e0d700a7a05c32db74 \
  --l2.head 0xdb19a102d2f4e2778ceb1121843aa9a47e7c68ceaff9407d02f1519e0716ec8d \
  --l2.outputroot 0x53983aa145265affe0675b43cda64baffdb97fea5dfb7a7144386b80430d27de \
  --l2.claim 0x7a35b8c93e63f12fb6078afff0ebf8b2ab41328419ec413ad010b22a74444ef1 \
  --l2.blocknumber 20 \
  --datadir ./preimages \
  --log.format terminal \
  --server