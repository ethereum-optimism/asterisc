# Benchmarking Radix Trie

Asterisc moved away from its hashmap-based memory structure to a radix-trie based memory structure.

This was done in order to:

1. Improve client diversity by differentiating the implementation from cannon
2. Improve runtime performance

In radix trie, the branching factor and the depth of the trie is critical. Depending on the sparsity of the dataset, we must adjust the radix trie to best suit the program it runs.

- Larger radix branching factors can lead to less levels and larger node sizes, which can lead to less pointer indirection and depth traversal while the larger node sizes leads to more memory footprint. Larger radix node also require more computation to generate a merkle root of a single node.
- Smaller radix branching factors lead to more levels and smaller node sizes, which have contrary impact compared to above.

There are two methods we used to benchmark the change to radix trie.

Multiple variants of radix trie were tested, with different branching factors.

Here’s the list of asterisc implemented with different configs:

| Variant | Radix type |
| --- | --- |
| Asterisc v1.0.0 | non-radix |
| Radix 1 | [4,4,4,4,4,4,4,8,8,8] - 10 levels |
| Radix 2 | [8,8,8,8,8,8,4] - 7 levels |
| Radix 3 | [4,4,4,4,4,4,4,4,4,4,4,4,4] - 13 levels |
| Radix 4 | [8,8,8,8,8,4,4,4] - 8 levels |
| Radix 5 | [16,16,6,6,4,4] - 6 levels |
| Radix 6 | [16,16,6,6,4,2,2] - 7 levels |
| Radix 7 | [10,10,10,10,12] - 5 levels |

## Benchmark Unit Test

New benchmark suite is added, which measures the latency of the following operations;

- Memory read / write to random addresses
- Memory read / write to contiguous address
- Memory write to sparse memory addresses
- Memory write to dense memory addresses
- Merkle proof generation
- Merkle root calculation

For the above cases, each asterisc implementation had the following results:

|  | Asteris v1.0.0 | Radix 1 | Radix 2 | Radix 3 | Radix 4 | Radix 5 | Radix 6 | Radix 7 |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| RandomReadWrite | 17.9n | 15.96 | 15.62 | 16.58 | 15.82 | 16.2 | 18.98 | 15.89 |
| SequentialReadWrite | 5.68n | 4.386 | 4.214 | 4.177 | 4.242 | 4.335 | 4.573 | 4.238 |
| SparseMemoryUsage | 4.964µ | 5.845 | 6.317 | 5.187 | 5.93 | 4.954 | 8.265 | 24.567 |
| DenseMemoryUsage | 11.73n | 9.094 | 9.649 | 10.11 | 10.12 | 10.4 | 10.12 | 10.12 |
| MerkleProof | 1.97µ | 1.441 | 1.464 | 1.611 | 1.737 | 1.604 | 1.737 | 1.98 |
| MerkleRoot | 6.129n | 4.536 | 4.52 | 4.509 | 4.648 | 4.623 | 4.746 | 4.928 |

Above statistics are based on  `sec/op` . Most of the results show that radix-based implementation is faster than the previous hashmap-based memory, except for few outliers.

Note that this does not account for memory usage such as `B/op` and `allocs/op`. As explained above, each initialization of radix-trie node allocates more memory than a hashmap would, leading to usually larger memory footprint.

## Full op-program run

For a more realistic performance of asterisc, we need to run it against the real chain data by running it as a VM client of op-program.

Tests were done on Asterisc running with Kona, for op-sepolia at [block#17484899](https://sepolia-optimism.etherscan.io/block/17484899)

|  | Average | Min | Max | % from v1.0.0 |
| --- | --- | --- | --- | --- |
| Asterisc v1.0.0 | 112.759 | 109.345 | 116.045 | 0.00% |
| Radix 1 | 110.349 | 109.418 | 112.149 | -2.14% |
| Radix 2 | 109.9 | 107.526 | 111.398 | -2.54% |
| Radix 3 | 110.589 | 107.814 | 113.544 | -1.92% |
| Radix 4 | 106.902 | 103.71 | 110.453 | -5.19% |
| Radix 5 | 106.605 | 104.469 | 109.754 | -5.46% |
| Radix 6 | 109.137 | 106.764 | 111.819 | -3.21% |
| Radix 7 | 111.163 | 110.392 | 111.634 | -1.42% |
| Radix 4 w/ pgo | 98.742 | 97.055 | 101.035 | -12.43% |

As you can see above, radix 4/5 had the best results compared to original asterisc implementation, with more than 5% improvement in op-program run duration.

After applying [pgo(profile-guided optimization)](https://go.dev/doc/pgo) on radix 4, we can observe over 12% improvement in speed.

## Visualizing address allocation pattern

In this radix-trie implementation, only the memory addresses that are actually allocatd are initialized as radix trie. Therefore, we can look at the overall memory allocation pattern to see how we can optimize the radix branching factor.

| Radix level (4bit each) | Allocations |
| --- | --- |
| 1 | 2 |
| 2 | 2 |
| 3 | 2 |
| 4 | 2 |
| 5 | 2 |
| 6 | 2 |
| 7 | 2 |
| 8 | 2 |
| 9 | 2 |
| 10 | 3 |
| 11 | 33 |
| 12 | 502 |
| 13 | 7982 |

Above graph is allocation count during a full op-program run, where the full address space(52 bits) are split into 13 nodes(4 bits each).

We can observe that the memory allocation is very sparse in the upper parts of the memory address, while it is heavily dense in the lower part of the memory address.

With only couple of allocation for 36bit-upper memory region, we could generalize that most of the op-program runs are confined to lower memory address regions.

## Conclusion

Based on above observations, and our goal of improving runtime performance, we decided on using `radix 5 (16, 16, 6, 6, 4, 4)`

Usually, sparse region would utilize smaller branching factor for memory optimization. However, since our goal is faster performance, utilizing larger levels at upper memory region and reducing trie traversal depth.

- use larger branching factors at the upper address level to reduce the trie traversal depth
- use smaller branching factors at the lower address level to reduce computation for each node.

In addition, we can apply pgo as mentioned above. To apply pgo to asterisc builds, we can run asterisc with cpu pprof enabled, and ship asterisc with `default.pgo` in the build path. This way, whenever the user builds Asterisc, pgo will be enabled by default, leading to addition 5+% improvement in speed.
