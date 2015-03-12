package main

import (
	"strings"
	"testing"
)

func TestRealStack(t *testing.T) {
	stacks, err := ParseStack(strings.NewReader(exampleStack), nil)
	if err != nil {
		t.Fatal(err)
	}
	if stacks[0].RIP != 0x366b1 {
		t.Fatalf("Expected %x first stack RIP, have %x", 0x366b1, stacks[0].RIP)
	}
	for _, s := range stacks {
		t.Logf("RIP %x", s.RIP)
	}
	if len(stacks) != 11 {
		t.Fatalf("Expected %d stacks, have %d", 11, len(stacks))
	}
}
func TestSimpleStack(t *testing.T) {
	stacks, err := ParseStack(strings.NewReader(`CPU:2 RIP: 98f32c1
CPU:2 RIP: 366b2
ffffe8ff0c344f38
ffffe8ff0c342f30`), nil)
	if err != nil {
		t.Fatal(err)
	}
	if stacks[0].RIP != 0x366b2 {
		t.Fatalf("Expected %x first stack RIP, have %x", 0x366b2, stacks[0].RIP)
	}
	if len(stacks) != 1 {
		t.Fatalf("Expected %d stacks, have %d", 12, len(stacks))
	}
	if stacks[0].Addr[0] != 0xffffe8ff0c344f38 || stacks[0].Addr[1] != 0xffffe8ff0c342f30 {
		t.Fatalf("Unexpected addresses %x %x", stacks[0].Addr[0], stacks[0].Addr[1])
	}
}

var exampleStack = `CPU:2 RIP: 810c9340
CPU:2 RIP: f6c0
CPU:2 RIP: 3ae2008d
CPU:2 RIP: f97dbbfa
CPU:2 RIP: 10865
CPU:2 RIP: 988e487
CPU:2 RIP: 9925bcb
CPU:2 RIP: 954900d
CPU:2 RIP: 810af3de
CPU:2 RIP: 98e10a0
CPU:2 RIP: 9995d84
CPU:2 RIP: 99953b5
CPU:2 RIP: d65f03bd
CPU:2 RIP: 9907000
CPU:2 RIP: 98f32c1
CPU:2 RIP: 366b1
ffffe8ff0c344f38
ffffe8ff0c342f30
ffffe8ff00411a00
ffffe8ff0c0f1000
ffffe8ff00dcb000
ffffe8ff00dcb000
ffffe8ff0000f48f
ffffe8ff0c182db8
ffffe8ff0c0f1000
ffffe8ff0001aa20
ffffe8ff00000003
ffffe8ff00e0cf38
ffffe8ff0c2ac570
ffffe8ff00e0cf48
ffffe8ff0006a115
ffffe8ff00e0ce70
ffffe8ff00e0ced8
ffffe8ff0c0f1000
ffffe8ff0c0f1000
ffffe8ff00411a00
ffffe8ff00030969
ffffe8ff00e0cf28
ffffe8ff0003594d
ffffe8ff0c0f1000
ffffe8ff09a06300
ffffe8ff0040c000
ffffe8ff00dcb000
ffffe8ff0003dba5
ffffe8ff0040c000
ffffe8ff0c0f1000
ffffe8ff0040c000
ffffe8ff0c0f1000
ffffe8ff0c0f1000
ffffe8ff00069016
ffffe8ff0c0f1000
ffffe8ff00001a55
CPU:2 RIP: fe47
CPU:2 RIP: 35a22
ffffe8ff0003a803
ffffe8ff000000b5
ffffe8ff00dcb000
ffffe8ff00e02e28
ffffe8ff00e02e08
ffffe8ff0c0f1000
ffffe8ff00e02f50
ffffe8ff0000a0bc
ffffe8ff00000671
ffffe8ff00e02f78
ffffe8ff099c2661
CPU:2 RIP: 914100b
CPU:2 RIP: 990728d
CPU:2 RIP: 1084d
CPU:2 RIP: fe40
CPU:2 RIP: 9a3b34d
CPU:2 RIP: 996e803
CPU:2 RIP: 814ab8eb
ffffe8ff000022cd
CPU:2 RIP: d6a37e9d
CPU:2 RIP: 940dc30
CPU:2 RIP: 98f32c1
CPU:2 RIP: 9634fb0
CPU:2 RIP: fbce
CPU:2 RIP: 5d4ab
ffffe8ff0c0f1000
ffffe8ff00e02e58
ffffe8ff0c0f1000
ffffe8ff0006338d
ffffe8ff00dcb000
ffffe8ff0c0f00b5
ffffe8ff00e02d00
ffffe8ff0c0f1000
ffffe8ff00cef740
ffffe8ff00ac99c0
ffffe8ff0c0f1000
ffffe8ff00e02de0
ffffe8ff0c0f1000
ffffe8ff00e02f50
ffffe8ff00e02de0
ffffe8ff00e02e58
ffffe8ff0c0f1000
ffffe8ff0002400b
ffffe8ff00e02e58
ffffe8ff0c0f1000
ffffe8ff0c0f1000
ffffe8ff000248fd
ffffe8ff0c1036a0
ffffe8ff0c0f1000
ffffe8ff00e02f50
ffffe8ff0000a4bd
ffffe8ff00000671
ffffe8ff00e02f78
ffffe8ff098f7633
CPU:2 RIP: 9abf075
CPU:2 RIP: 96661d0
CPU:2 RIP: fe8f
CPU:2 RIP: 10865
CPU:2 RIP: 972722d
CPU:2 RIP: 9f79
ffffe8ff00000671
ffffe8ff00e02f78
ffffe8ff0942d745
CPU:2 RIP: 9a7cb40
CPU:2 RIP: 1ae7
CPU:2 RIP: 11c3f
ffffe8ff0c344f38
ffffe8ff0c342f30
ffffe8ff00411a00
ffffe8ff00e02e58
ffffe8ff0c0f1000
ffffe8ff00011d2e
ffffe8ff0c0f1000
ffffe8ff00e02e58
ffffe8ff0c0f1000
ffffe8ff00e02e58
ffffe8ff00e02e78
ffffe8ff00024389
ffffe8ff0c0f1000
ffffe8ff0c0f1000
ffffe8ff00e02e78
ffffe8ff00e02e58
ffffe8ff0002473e
ffffe8ff00e02e58
ffffe8ff00e02e58
ffffe8ff00e02e58
ffffe8ff0c0f1000
ffffe8ff0005d4d1
ffffe8ff0c0f1000
ffffe8ff00e02e58
ffffe8ff0c0f1000
ffffe8ff0006338d
ffffe8ff00dcb000
ffffe8ff0c0f00b5
ffffe8ff00e02d00
ffffe8ff0c0f1000
ffffe8ff00cef740
ffffe8ff00ac99c0
ffffe8ff0c0f1000
ffffe8ff00e02de0
ffffe8ff0c0f1000
ffffe8ff00e02f50
ffffe8ff00e02de0
ffffe8ff00e02e58
ffffe8ff0c0f1000
ffffe8ff0002400b
ffffe8ff00e02e58
ffffe8ff0c0f1000
ffffe8ff0c0f1000
ffffe8ff000248fd
ffffe8ff0c0f1000
ffffe8ff00e02f50
ffffe8ff0000a4bd
ffffe8ff00000671
ffffe8ff00e02f78
ffffe8ff098f7db6
CPU:2 RIP: 98edc19
CPU:2 RIP: 19a5e
ffffe8ff00000003
ffffe8ff00e0cf38
ffffe8ff0c2ac570
ffffe8ff00e0cf48
ffffe8ff0006a115
ffffe8ff00e0ce70
ffffe8ff00e0ced8
ffffe8ff0c0f1000
ffffe8ff0c0f1000
ffffe8ff00411a00
ffffe8ff00030969
ffffe8ff00e0cf28
ffffe8ff0003594d
ffffe8ff0c0f1000
ffffe8ff09a06300
ffffe8ff0040c000
ffffe8ff00dcb000
ffffe8ff0003dba5
ffffe8ff0040c000
ffffe8ff0c0f1000
ffffe8ff0040c000
ffffe8ff0c0f1000
ffffe8ff0c0f1000
ffffe8ff00069016
ffffe8ff00001a55
CPU:2 RIP: 664eb
ffffe8ff00dcb000
ffffe8ff0006277e
ffffe8ff0c0f1000
ffffe8ff00e0ce08
ffffe8ff0c0f1000
ffffe8ff00064071
ffffe8ff00003df7
ffffe8ff00e0ce08
ffffe8ff0c0f1000
ffffe8ff00011d2e
ffffe8ff0c0f1000
ffffe8ff00e0ce08
ffffe8ff00e0ce40
ffffe8ff00024389
ffffe8ff0c0f1000
ffffe8ff0c182db8
ffffe8ff00010be0
ffffe8ff00e0ce08
ffffe8ff0c0f1000
ffffe8ff0c344f38
ffffe8ff0c342f30
ffffe8ff00411a00
ffffe8ff0c0f1000
ffffe8ff00dcb000
ffffe8ff0c0f1000
ffffe8ff0c0e6590
ffffe8ff00dcb000
ffffe8ff00019bce
ffffe8ff00000003
ffffe8ff00e0cf38
ffffe8ff0c2ac570
ffffe8ff00e0cf48
ffffe8ff0006a115
ffffe8ff00e0ce70
ffffe8ff00e0ced8
ffffe8ff0c0f1000
ffffe8ff0c0f1000
ffffe8ff00411a00
ffffe8ff00030969
ffffe8ff00e0cf28
ffffe8ff0003594d
ffffe8ff0c0f1000
ffffe8ff093f43b0
ffffe8ff0040c000
ffffe8ff00dcb000
ffffe8ff0003dba5
ffffe8ff0040c000
ffffe8ff0c0f1000
ffffe8ff0040c000
ffffe8ff0c0f1000
ffffe8ff00061cd5
ffffe8ff00069016
ffffe8ff00001a55
CPU:2 RIP: 9a7c048
CPU:2 RIP: 99223ef
CPU:2 RIP: 7abfbd68
CPU:2 RIP: d6a3aced
CPU:2 RIP: 962f014
CPU:2 RIP: 6b35d
ffffe8ff0c0f1000
ffffe8ff0006234d
ffffe8ff0c182db8
ffffe8ff0c0f1000
ffffe8ff0001b6cd
ffffe8ff00000003
ffffe8ff00e0cf38
ffffe8ff0c2ac570
ffffe8ff00e0cf48
ffffe8ff0006a115
ffffe8ff00e0ce70
ffffe8ff00e0ced8
ffffe8ff0c0f1000
ffffe8ff0c0f1000
ffffe8ff00411a00
ffffe8ff00030969
ffffe8ff00e0cf28
ffffe8ff0003594d
ffffe8ff0c0f1000
ffffe8ff09a06300
ffffe8ff0040c000
ffffe8ff00dcb000
ffffe8ff0003dba5
ffffe8ff0040c000
ffffe8ff0c0f1000
ffffe8ff0040c000
ffffe8ff0c0f1000
ffffe8ff0c0f1000
ffffe8ff00069016
ffffe8ff0c0f1000
ffffe8ff00001a55
CPU:2 RIP: d6a298c0
CPU:2 RIP: d6a2e1fc
CPU:2 RIP: 98e12a5
CPU:2 RIP: 6235c
ffffe8ff0c182db8
ffffe8ff0c0f1000
ffffe8ff0001b6cd
ffffe8ff00000003
ffffe8ff00e0cf38
ffffe8ff0c2ac570
ffffe8ff00e0cf48
ffffe8ff0006a115
ffffe8ff00e0ce70
ffffe8ff00e0ced8
ffffe8ff0c0f1000
ffffe8ff0c0f1000
ffffe8ff00411a00
ffffe8ff00030969
ffffe8ff00e0cf28
ffffe8ff0003594d
ffffe8ff0c0f1000
ffffe8ff09a06300
ffffe8ff0040c000
ffffe8ff00dcb000
ffffe8ff0003dba5
ffffe8ff0040c000
ffffe8ff0c0f1000
ffffe8ff0040c000
ffffe8ff0c0f1000
ffffe8ff0c182db8
ffffe8ff00069016
ffffe8ff0c0f1000
ffffe8ff00001a55
CPU:2 RIP: 916f008
CPU:2 RIP: d5c31021
CPU:2 RIP: 10854
CPU:2 RIP: 90c1ae0
CPU:2 RIP: fe40
CPU:2 RIP: 90a90d7
CPU:2 RIP: 90beda7
CPU:2 RIP: 814ae1c0
CPU:2 RIP: 93eb900
CPU:2 RIP: 4a8f5fc0
CPU:2 RIP: 8107c6ea
CPU:2 RIP: 62761
ffffe8ff00e0ce08
ffffe8ff0c0f1000
ffffe8ff00063374
ffffe8ff00003df7
ffffe8ff00e0ce00
ffffe8ff00011d00
ffffe8ff0c0f1000
ffffe8ff00024389
ffffe8ff0008cfdc
ffffe8ff00e0ce08
ffffe8ff0c0f1000
ffffe8ff0c344f38
ffffe8ff0c342f30
ffffe8ff00411a00
ffffe8ff0c0f1000
ffffe8ff00dcb000
ffffe8ff0c0f1000
ffffe8ff0bc04b80
ffffe8ff00dcb000
ffffe8ff00019bce
ffffe8ff00000001
ffffe8ff00e0cf38
ffffe8ff0c2bb590
ffffe8ff00e0cf48
ffffe8ff0006a115
ffffe8ff00e0ce70
ffffe8ff00e0ced8
ffffe8ff0c1036a0
ffffe8ff0c0f1000
ffffe8ff0c0f1000
ffffe8ff00411a00
ffffe8ff00030969
ffffe8ff00e0cf28
ffffe8ff0003594d
ffffe8ff0c0f1000
ffffe8ff099fe300
ffffe8ff0040c000
ffffe8ff00dcb000
ffffe8ff0003dba5
ffffe8ff0040c000
ffffe8ff0c0f1000
ffffe8ff0040c000
ffffe8ff0c0f1000
ffffe8ff0c0f1000
ffffe8ff00069016
ffffe8ff00001a55
CPU:2 RIP: 8101167c
CPU:2 RIP: 5c8859
CPU:2 RIP: d3cad688
CPU:2 RIP: d367590f
CPU:2 RIP: d3d0c31e
CPU:2 RIP: d3d084be
CPU:2 RIP: 93a10c0
CPU:2 RIP: 81048ba4`
