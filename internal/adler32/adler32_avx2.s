//go:build !purego && amd64

#include "textflag.h"

DATA LCPI0_0<>+0x00(SB)/8, $0x191a1b1c1d1e1f20
DATA LCPI0_0<>+0x08(SB)/8, $0x1112131415161718
DATA LCPI0_0<>+0x10(SB)/8, $0x090a0b0c0d0e0f10
DATA LCPI0_0<>+0x18(SB)/8, $0x0102030405060708
GLOBL LCPI0_0<>(SB), (RODATA|NOPTR), $32

DATA LCPI0_1<>+0x00(SB)/8, $0x0001000100010001
DATA LCPI0_1<>+0x08(SB)/8, $0x0001000100010001
DATA LCPI0_1<>+0x10(SB)/8, $0x0001000100010001
DATA LCPI0_1<>+0x18(SB)/8, $0x0001000100010001
GLOBL LCPI0_1<>(SB), (RODATA|NOPTR), $32

DATA LCPI0_2<>+0x00(SB)/2, $0x0001
GLOBL LCPI0_2<>(SB), (RODATA|NOPTR), $2

TEXT ·adler32_avx2(SB), NOSPLIT, $0-36
	MOVLQZX       in+0(FP), DI
	MOVQ          buf_base+8(FP), SI
	MOVQ          buf_len+16(FP), DX
	MOVQ          buf_cap+24(FP), CX
	TESTQ         SI, SI
	JE            return_one
	MOVL          DI, AX
	TESTQ         DX, DX
	JE            return_current
	MOVL          AX, CX
	SHRL          $0x10, CX
	MOVWLZX       AX, AX
	CMPQ          DX, $0x20
	JB            scalar_unrolled16
	MOVL          $2147975281, DI
	VPXOR         X0, X0, X0
	VMOVDQA       LCPI0_0<>(SB), Y1
	VPBROADCASTW  LCPI0_2<>(SB), Y2
	JMP           vector_outer

vector_tail_init:
	VMOVDQA       Y4, Y6
	VPXOR         X5, X5, X5

vector_reduce_finalize_chunk:
	SUBQ          AX, DX
	VPSLLD        $0x05, Y5, Y4
	VPADDD        Y3, Y4, Y3
	VEXTRACTI128  $0x1, Y6, X4
	VSHUFPS       $0x88, X4, X6, X5
	VPSHUFD       $0x88, X4, X4
	VPADDD        X4, X5, X4
	VPSHUFD       $0x55, X4, X5
	VPADDD        X4, X5, X4
	VMOVD         X4, AX
	MOVQ          AX, CX
	IMULQ         DI, CX
	SHRQ          $0x2f, CX
	IMULL         $0xfff1, CX
	SUBL          CX, AX
	VEXTRACTI128  $0x1, Y3, X4
	VPADDD        X3, X4, X3
	VPSHUFD       $0xee, X3, X4
	VPADDD        X4, X3, X3
	VPSHUFD       $0x55, X3, X4
	VPADDD        X3, X4, X3
	VMOVD         X3, CX
	MOVQ          CX, R8
	IMULQ         DI, R8
	SHRQ          $0x2f, R8
	IMULL         $0xfff1, R8
	SUBL          R8, CX
	CMPQ          DX, $0x1f
	JBE           scalar_entry

vector_outer:
	VMOVD         AX, X4
	VMOVD         CX, X3
	CMPQ          DX, $0x15b0
	MOVL          $0x15b0, R8
	CMOVQCS       DX, R8
	MOVL          R8, AX
	ANDL          $0x1fe0, AX
	JE            vector_tail_init
	ADDQ          $-0x20, R8
	VPXOR         X5, X5, X5
	TESTL         $0x20, R8
	JNE           vector_block32_check
	VMOVDQU       0(SI), Y5
	ADDQ          $0x20, SI
	LEAQ          -0x20(AX), CX
	VPSADBW       Y0, Y5, Y6
	VPADDD        Y4, Y6, Y6
	VPMADDUBSW    Y1, Y5, Y5
	VPMADDWD      Y2, Y5, Y5
	VPADDD        Y3, Y5, Y3
	VMOVDQA       Y4, Y5
	VMOVDQA       Y6, Y4
	CMPQ          R8, $0x20
	JAE           vector_block64_loop
	JMP           vector_reduce_finalize_chunk

vector_block32_check:
	MOVQ          AX, CX
	CMPQ          R8, $0x20
	JB            vector_reduce_finalize_chunk

vector_block64_loop:
	VMOVDQU       0(SI), Y6
	VMOVDQU       0x20(SI), Y7
	VPSADBW       Y0, Y6, Y8
	VPADDD        Y4, Y8, Y8
	VPADDD        Y4, Y5, Y5
	VPMADDUBSW    Y1, Y6, Y4
	VPMADDWD      Y2, Y4, Y4
	VPADDD        Y3, Y4, Y3
	ADDQ          $0x40, SI
	VPSADBW       Y0, Y7, Y4
	VPADDD        Y4, Y8, Y4
	VPADDD        Y5, Y8, Y5
	VPMADDUBSW    Y1, Y7, Y6
	VPMADDWD      Y2, Y6, Y6
	VPADDD        Y3, Y6, Y3
	ADDQ          $-0x40, CX
	JNE           vector_block64_loop
	VMOVDQA       Y4, Y6
	JMP           vector_reduce_finalize_chunk

return_one:
	MOVL          $0x1, AX

return_current:
	MOVL          AX, ret+32(FP)
	RET

scalar_entry:
	TESTQ         DX, DX
	JE            return_final

scalar_unrolled16:
	CMPQ          DX, $0x10
	JB            scalar_byte_prelude
	MOVBLZX       0(SI), DI
	ADDL          DI, AX
	ADDL          AX, CX
	MOVBLZX       0x1(SI), DI
	ADDL          AX, DI
	ADDL          DI, CX
	MOVBLZX       0x2(SI), AX
	ADDL          DI, AX
	ADDL          AX, CX
	MOVBLZX       0x3(SI), DI
	ADDL          AX, DI
	ADDL          DI, CX
	MOVBLZX       0x4(SI), AX
	ADDL          DI, AX
	ADDL          AX, CX
	MOVBLZX       0x5(SI), DI
	ADDL          AX, DI
	ADDL          DI, CX
	MOVBLZX       0x6(SI), AX
	ADDL          DI, AX
	ADDL          AX, CX
	MOVBLZX       0x7(SI), DI
	ADDL          AX, DI
	ADDL          DI, CX
	MOVBLZX       0x8(SI), AX
	ADDL          DI, AX
	ADDL          AX, CX
	MOVBLZX       0x9(SI), DI
	ADDL          AX, DI
	ADDL          DI, CX
	MOVBLZX       0xa(SI), AX
	ADDL          DI, AX
	ADDL          AX, CX
	MOVBLZX       0xb(SI), DI
	ADDL          AX, DI
	ADDL          DI, CX
	MOVBLZX       0xc(SI), AX
	ADDL          DI, AX
	ADDL          AX, CX
	MOVBLZX       0xd(SI), DI
	ADDL          AX, DI
	ADDL          DI, CX
	MOVBLZX       0xe(SI), R8
	ADDL          DI, R8
	ADDL          R8, CX
	MOVBLZX       0xf(SI), AX
	ADDL          R8, AX
	ADDL          AX, CX
	ADDQ          $-0x10, DX
	JE            scalar_finalize
	ADDQ          $0x10, SI

scalar_byte_prelude:
	LEAQ          -0x1(DX), DI
	MOVQ          DX, R9
	ANDQ          $0x3, R9
	JE            scalar_dword_prelude
	XORL          R8, R8

scalar_byte_prelude_loop:
	MOVBLZX       0(SI)(R8*1), R10
	ADDL          R10, AX
	ADDL          AX, CX
	INCQ          R8
	CMPQ          R9, R8
	JNE           scalar_byte_prelude_loop
	ADDQ          R8, SI
	SUBQ          R8, DX

scalar_dword_prelude:
	CMPQ          DI, $0x3
	JB            scalar_finalize
	XORL          DI, DI

scalar_dword_loop:
	MOVBLZX       0(SI)(DI*1), R8
	ADDL          AX, R8
	ADDL          R8, CX
	MOVBLZX       0x1(SI)(DI*1), AX
	ADDL          R8, AX
	ADDL          AX, CX
	MOVBLZX       0x2(SI)(DI*1), R8
	ADDL          AX, R8
	ADDL          R8, CX
	MOVBLZX       0x3(SI)(DI*1), AX
	ADDL          R8, AX
	ADDL          AX, CX
	ADDQ          $0x4, DI
	CMPQ          DX, DI
	JNE           scalar_dword_loop

scalar_finalize:
	LEAL          -0xfff1(AX), DX
	CMPL          AX, $0xfff1
	CMOVLCS       AX, DX
	MOVL          CX, AX
	MOVL          $2147975281, SI
	IMULQ         AX, SI
	SHRQ          $0x2f, SI
	MOVL          SI, AX
	IMULL         $0xfff1, AX
	SUBL          AX, CX
	SHLL          $0x10, CX
	ORL           DX, CX
	MOVL          CX, AX
	VZEROUPPER
	MOVL          AX, ret+32(FP)
	RET

return_final:
	SHLL          $0x10, CX
	ORL           CX, AX
	VZEROUPPER
	MOVL          AX, ret+32(FP)
	RET
