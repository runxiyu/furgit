//go:build amd64 && !purego

#include "textflag.h"

// func adler32AVX2(state uint32, b []byte) uint32
//	state = s2<<16 | s1
//	len(b) <= nmax enforced by Go wrapper
TEXT ·adler32AVX2(SB), NOSPLIT, $0-40
	//   state   uint32   at 0
	//   b.ptr   *byte    at 8
	//   b.len   int      at 16
	//   b.cap   int      at 24
	//   ret     uint32   at 32

	MOVL	state+0(FP), AX
	MOVQ	b_base+8(FP), SI
	MOVQ	b_len+16(FP), CX

	// s1 in R10d, s2 in R11d
	MOVL	AX, R10
	ANDL	$0xffff, R10      // s1 = low 16 bits
	SHRL	$16, AX
	MOVL	AX, R11           // s2 = high 16 bits

	// Just use the scalar tail if len < 32
	CMPQ	CX, $32
	JLT	tail

loop32:
	VMOVDQU	(SI), X0
	VMOVDQU	16(SI), X1
	ADDQ	$32, SI
	SUBQ	$32, CX

	// Split 32 bytes into uint16 lanes (four groups of eight bytes)
	VPMOVZXBW	X0, Y2
	VPMOVZXBW	X1, Y3
	VEXTRACTI128	$0, Y2, X4
	VEXTRACTI128	$1, Y2, X5
	VEXTRACTI128	$0, Y3, X6
	VEXTRACTI128	$1, Y3, X7

	// Process each group using prefix sums (only adds, no multiplies)
	// Group 0 (bytes 0..7)
	VMOVDQA	X4, X8
	VPSLLDQ	$2, X8, X9
	VPADDW	X9, X8, X8
	VPSLLDQ	$4, X8, X9
	VPADDW	X9, X8, X8
	VPSLLDQ	$8, X8, X9
	VPADDW	X9, X8, X8
	VPEXTRW	$7, X8, R8
	VPSRLDQ	$8, X8, X9
	VPADDW	X9, X8, X8
	VPSRLDQ	$4, X8, X9
	VPADDW	X9, X8, X8
	VPSRLDQ	$2, X8, X9
	VPADDW	X9, X8, X8
	VPEXTRW	$0, X8, R9
	MOVL	R10, R12
	SHLL	$3, R12
	ADDL	R12, R11
	ADDL	R9, R11
	ADDL	R8, R10

	// Group 1 (bytes 8..15)
	VMOVDQA	X5, X8
	VPSLLDQ	$2, X8, X9
	VPADDW	X9, X8, X8
	VPSLLDQ	$4, X8, X9
	VPADDW	X9, X8, X8
	VPSLLDQ	$8, X8, X9
	VPADDW	X9, X8, X8
	VPEXTRW	$7, X8, R8
	VPSRLDQ	$8, X8, X9
	VPADDW	X9, X8, X8
	VPSRLDQ	$4, X8, X9
	VPADDW	X9, X8, X8
	VPSRLDQ	$2, X8, X9
	VPADDW	X9, X8, X8
	VPEXTRW	$0, X8, R9
	MOVL	R10, R12
	SHLL	$3, R12
	ADDL	R12, R11
	ADDL	R9, R11
	ADDL	R8, R10

	// Group 2 (bytes 16..23)
	VMOVDQA	X6, X8
	VPSLLDQ	$2, X8, X9
	VPADDW	X9, X8, X8
	VPSLLDQ	$4, X8, X9
	VPADDW	X9, X8, X8
	VPSLLDQ	$8, X8, X9
	VPADDW	X9, X8, X8
	VPEXTRW	$7, X8, R8
	VPSRLDQ	$8, X8, X9
	VPADDW	X9, X8, X8
	VPSRLDQ	$4, X8, X9
	VPADDW	X9, X8, X8
	VPSRLDQ	$2, X8, X9
	VPADDW	X9, X8, X8
	VPEXTRW	$0, X8, R9
	MOVL	R10, R12
	SHLL	$3, R12
	ADDL	R12, R11
	ADDL	R9, R11
	ADDL	R8, R10

	// Group 3 (bytes 24..31)
	VMOVDQA	X7, X8
	VPSLLDQ	$2, X8, X9
	VPADDW	X9, X8, X8
	VPSLLDQ	$4, X8, X9
	VPADDW	X9, X8, X8
	VPSLLDQ	$8, X8, X9
	VPADDW	X9, X8, X8
	VPEXTRW	$7, X8, R8
	VPSRLDQ	$8, X8, X9
	VPADDW	X9, X8, X8
	VPSRLDQ	$4, X8, X9
	VPADDW	X9, X8, X8
	VPSRLDQ	$2, X8, X9
	VPADDW	X9, X8, X8
	VPEXTRW	$0, X8, R9
	MOVL	R10, R12
	SHLL	$3, R12
	ADDL	R12, R11
	ADDL	R9, R11
	ADDL	R8, R10

	CMPQ	CX, $32
	JGE	loop32

tail:
	TESTQ	CX, CX
	JEQ	done

tail_loop:
	MOVBLZX	(SI), R8
	INCQ	SI
	DECQ	CX

	ADDL	R8, R10
	ADDL	R10, R11

	TESTQ	CX, CX
	JNE	tail_loop

done:
	MOVL	$65521, R8

	// Reduce s1 %= mod
	MOVL	R10, AX
	XORL	DX, DX
	DIVL	R8
	MOVL	DX, R10

	// Reduce s2 %= mod
	MOVL	R11, AX
	XORL	DX, DX
	DIVL	R8
	MOVL	DX, R11

	MOVL	R11, AX
	SHLL	$16, AX
	ANDL	$0xffff, R10
	ORL	R10, AX

	VZEROUPPER

	MOVL	AX, ret+32(FP)
	RET
