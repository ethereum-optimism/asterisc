import asyncio
import json
import os

import eth_abi
import requests
import websockets

L1_WS_ENDPOINT = "ws://localhost:8546"
L1_HTTP_ENDPOINT = "http://localhost:8545"
L2_HTTP_ENDPOINT = "http://localhost:9545"
# event DisputeGameCreated(address indexed disputeProxy, GameType indexed gameType, Claim indexed rootClaim);
# event DisputeGameCreated(address indexed disputeProxy, uint32 indexed gameType, bytes32 indexed rootClaim);
DISPUTE_GAME_CREATED_TOPIC = (
    "0x5b565efe82411da98814f356d0e7bcb8f0219b8d970307c5afb4a6903a8b2e35"
)
CREATE_TX_ABI_TYPES = ["uint32", "bytes32", "bytes"]

current_dir = os.path.dirname(os.path.abspath(__file__))
project_root = os.path.dirname(os.path.dirname(current_dir))
optimism_root = os.path.join(project_root, "rvsol/lib/optimism")

with open(os.path.join(optimism_root, ".devnet/addresses.json"), "r") as f:
    addrs = json.load(f)

logs = []
l2_head = ""


async def subscribe_logs():
    async with websockets.connect(L1_WS_ENDPOINT) as websocket:
        subscription_request = {
            "jsonrpc": "2.0",
            "method": "eth_subscribe",
            "params": [
                "logs",
                {
                    "address": addrs["DisputeGameFactoryProxy"],
                    "topics": [DISPUTE_GAME_CREATED_TOPIC],
                },
            ],
            "id": 1,
        }

        await websocket.send(json.dumps(subscription_request))

        print("Waiting DisputeGameCreated logs...")
        while True:
            message = await websocket.recv()
            res = json.loads(message)
            if "params" in res:
                event_result = res["params"]["result"]
                l1_tx_request = {
                    "jsonrpc": "2.0",
                    "method": "eth_getTransactionByHash",
                    "params": [event_result["transactionHash"]],
                    "id": 1,
                }
                try:
                    res = requests.post(L1_HTTP_ENDPOINT, json=l1_tx_request).json()
                    calldata = bytes.fromhex(res["result"]["input"][10:])
                    params = eth_abi.decode(CREATE_TX_ABI_TYPES, calldata)
                    l2_block_number = int.from_bytes(params[-1], byteorder="big")
                except Exception as e:
                    raise Exception(f"Failed to fetch L2 block number: {e}")

                logs.append(
                    {
                        "outputRoot": event_result["topics"][3],
                        "l1BlockHash": event_result["blockHash"],
                        "l1BlockNumber": int(event_result["blockNumber"], base=16),
                        "l2BlockNumber": l2_block_number,
                    }
                )
            if len(logs) == 2:
                break

    l2_block_reqeust = {
        "jsonrpc": "2.0",
        "method": "eth_getBlockByNumber",
        "params": [hex(logs[0]["l2BlockNumber"]), False],
        "id": 1,
    }
    res = requests.post(L2_HTTP_ENDPOINT, json=l2_block_reqeust).json()

    global l2_head
    l2_head = res["result"]["hash"]


asyncio.run(subscribe_logs())

local_cmd = f'''#!/bin/bash

./asterisc run \\
  --info-at '%10000000' \\
  --proof-at never \\
  --input ./test-data/state.json \\
  --meta ./test-data/meta.json \\
  -- \\
  ./op-program \\
  --rollup.config ./test-data/chain-artifacts/rollup.json \\
  --l2.genesis ./test-data/chain-artifacts/genesis-l2.json \\
  --l1.trustrpc \\
  --l1.rpckind debug_geth \\
  --l1.head {logs[1]["l1BlockHash"]} \\
  --l2.head {l2_head} \\
  --l2.outputroot {logs[0]["outputRoot"]} \\
  --l2.claim {logs[1]["outputRoot"]} \\
  --l2.blocknumber {logs[1]["l2BlockNumber"]} \\
  --datadir ./test-data/preimages \\
  --log.format terminal \\
  --server'''

capture_cmd = local_cmd + " --l1 http://127.0.0.1:8545 --l2 http://127.0.0.1:9545"

# Script to capture preimages from the local devnet
with open(os.path.join(current_dir, "capture_cmd.sh"), "w") as f:
    f.write(capture_cmd)
os.chmod("capture_cmd.sh", 0o755)

# Script to run op-program in offline mode
with open(os.path.join(current_dir, "local_cmd.sh"), "w") as f:
    f.write(local_cmd)
os.chmod("local_cmd.sh", 0o755)
