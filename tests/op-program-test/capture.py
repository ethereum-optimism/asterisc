import asyncio
import websockets
import requests
import json
import os

L1_WS_ENDPOINT = "ws://localhost:8546"
L2_HTTP_ENDPOINT = "http://localhost:9545"
OUTPUT_PROPOSED_TOPIC = "0xa7aaf2512769da4e444e3de247be2564225c2e7a8f74cfe528e46e17d24868e2"

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
            "params": ["logs", {"address": addrs["L2OutputOracleProxy"], "topics": [OUTPUT_PROPOSED_TOPIC]}],
            "id": 1
        }

        await websocket.send(json.dumps(subscription_request))

        print("Waiting OutputProposed logs...")
        while True:
            message = await websocket.recv()
            res = json.loads(message)
            if "params" in res:
                result = res["params"]["result"]
                logs.append({
                    "outputRoot": result["topics"][1],
                    "l2BlockNumber": int(result["topics"][3], base=16),
                    "l1BlockNumber": int(result["blockNumber"], base=16),
                    "l1BlockHash": result["blockHash"]
                })
            if len(logs) == 2:
                break

    l2_block_reqeust = {
        "jsonrpc": "2.0",
        "method": "eth_getBlockByNumber",
        "params": [hex(logs[0]["l2BlockNumber"]), False],
        "id": 1
    }
    res = requests.post(L2_HTTP_ENDPOINT, json=l2_block_reqeust).json()

    global l2_head
    l2_head = res["result"]["hash"]

asyncio.run(subscribe_logs())

local_cmd = f'''#!/bin/bash

./asterisc run \\
  --info-at '%10000000' \\
  --proof-at never \\
  --input ./state.json \\
  -- \\
  ./op-program \\
  --rollup.config ./chain-artifacts/rollup.json \\
  --l2.genesis ./chain-artifacts/genesis-l2.json \\
  --l1.trustrpc \\
  --l1.rpckind debug_geth \\
  --l1.head {logs[1]["l1BlockHash"]} \\
  --l2.head {l2_head} \\
  --l2.outputroot {logs[0]["outputRoot"]} \\
  --l2.claim {logs[1]["outputRoot"]} \\
  --l2.blocknumber {logs[1]["l2BlockNumber"]} \\
  --datadir ./preimages \\
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
