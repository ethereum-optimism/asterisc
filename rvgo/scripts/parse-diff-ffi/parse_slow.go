package main

import (
	"flag"
	"fmt"

	"github.com/holiman/uint256"

	"github.com/ethereum-optimism/asterisc/rvgo/slow"
)

func main() {
	function := flag.String("fuzz", "ParseTypeI", "fuzz function")
	input := flag.Int64("number", 0, "input to parse")
	flag.Parse()

	number := *input

	//u256_input := slow.ShortToU256(uint16(number))

	toU64 := slow.U64(*uint256.NewInt(uint64(number)))

	switch *function {
	case "ParseTypeI":
		{
			resultU64 := slow.ParseImmTypeI(toU64) // type should be U64
			h := fmt.Sprintf("%064x", slow.Val(resultU64))
			fmt.Print(h)
		}
	case "ParseTypeS":
		{
			resultU64 := slow.ParseImmTypeS(toU64) // type should be U64
			h := fmt.Sprintf("%064x", slow.Val(resultU64))
			fmt.Print(h)
		}
	case "ParseTypeB":
		{
			resultU64 := slow.ParseImmTypeB(toU64) // type should be U64
			h := fmt.Sprintf("%064x", slow.Val(resultU64))
			fmt.Print(h)
		}
	case "ParseTypeU":
		{
			resultU64 := slow.ParseImmTypeU(toU64) // type should be U64
			h := fmt.Sprintf("%064x", slow.Val(resultU64))
			fmt.Print(h)
		}
	case "ParseTypeJ":
		{
			resultU64 := slow.ParseImmTypeJ(toU64) // type should be U64
			h := fmt.Sprintf("%064x", slow.Val(resultU64))
			fmt.Print(h)
		}
	case "ParseOpcode":
		{
			resultU64 := slow.ParseOpcode(toU64) // type should be U64
			h := fmt.Sprintf("%064x", slow.Val(resultU64))
			fmt.Print(h)
		}
	case "ParseRd":
		{
			resultU64 := slow.ParseRd(toU64) // type should be U64
			h := fmt.Sprintf("%064x", slow.Val(resultU64))
			fmt.Print(h)
		}
	case "ParseFunct3":
		{
			resultU64 := slow.ParseFunct3(toU64) // type should be U64
			h := fmt.Sprintf("%064x", slow.Val(resultU64))
			fmt.Print(h)
		}
	case "ParseRs1":
		{
			resultU64 := slow.ParseRs1(toU64) // type should be U64
			h := fmt.Sprintf("%064x", slow.Val(resultU64))
			fmt.Print(h)
		}
	case "ParseRs2":
		{
			resultU64 := slow.ParseRs2(toU64) // type should be U64
			h := fmt.Sprintf("%064x", slow.Val(resultU64))
			fmt.Print(h)
		}
	case "ParseFunct7":
		{
			resultU64 := slow.ParseFunct7(toU64) // type should be U64
			h := fmt.Sprintf("%064x", slow.Val(resultU64))
			fmt.Print(h)
		}
	default:
		{
			panic("unknown input")
		}
	}
}
