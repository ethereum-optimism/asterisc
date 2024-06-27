## Asterisc Stage 1.4 Deployment Information

### Sepolia

Deployer Address(EOA): `0x7DaE43a953C40d371C4bE9963FdD250398a4915A`

**Deployments**
| Contract                   | Address                                      |
|--------------------------- |--------------------------------------------- |
| DisputeGameFactory (proxy) | [`0x1307f44beDCdCaeA0716abb2016c1792ed310f46`](https://sepolia.etherscan.io/address/0x1307f44beDCdCaeA0716abb2016c1792ed310f46#code) |
| DisputeGameFactory (impl)  | [`0x872507195C86Da21A9db7b9eD2589d21f20CB61E`](https://sepolia.etherscan.io/address/0x872507195C86Da21A9db7b9eD2589d21f20CB61E#code) |
| AnchorStateRegistry (proxy) | [`0x1a31b297355a30752D5E2A4AC3e9DC47D02485c2`](https://sepolia.etherscan.io/address/0x1a31b297355a30752D5E2A4AC3e9DC47D02485c2#code) |
| AnchorStateRegistry (impl)  | [`0xB93884E2C21a78d036f213eA4f77Eb7e94147065`](https://sepolia.etherscan.io/address/0xB93884E2C21a78d036f213eA4f77Eb7e94147065#code) |
| RISCV VM                    | [`0xfD70c0391eC0308445c2BE83F85689aF72946332`](https://sepolia.etherscan.io/address/0xfD70c0391eC0308445c2BE83F85689aF72946332#code) |
| FaultDisputeGame (impl)    | [`0x71838a0d3b3226842166619338ab3218c62bb4a1`](https://sepolia.etherscan.io/address/0x71838a0d3b3226842166619338ab3218c62bb4a1#code) |
| ProxyAdmin    | [`0x2F4613Aa09634CD07f39b3ee91FB9b1ca57b94C1`](https://sepolia.etherscan.io/address/0x2F4613Aa09634CD07f39b3ee91FB9b1ca57b94C1#code) |

**Version**
- FPVM release: https://github.com/ethereum-optimism/asterisc/releases/tag/v1.0.0
- Monorepo commit hash: [`457f33f4fdda9373dcf2839619ebf67182ee5057`](https://github.com/ethereum-optimism/optimism/tree/457f33f4fdda9373dcf2839619ebf67182ee5057)

**Configuration**
- Fault Proof Program: op-program v1.1.0
- Absolute prestate hash:  `0x0338dc64405def7e3b9ce8f5076b422a846d831832617d227f13baf219cb5406`
- Max game depth: `73`
    - Supports an instruction trace up to `2 ** 73` instructions long.
- Max game duration: `302,400 seconds` (84 hours)
- All other configs at [deploy-config/op-sepolia.json](deploy-config/op-sepolia.json)
