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
	MOVLQZX      in+0(FP), DI
	MOVQ         buf_base+8(FP), SI
	MOVQ         buf_len+16(FP), DX
	MOVQ         buf_cap+24(FP), CX
	WORD         $0x8548; BYTE $0xf6
	JE           return_one
	WORD         $0xf889
	WORD         $0x8548; BYTE $0xd2
	JE           return_result
	NOP
	NOP
	NOP
	WORD         $0xc189
	WORD         $0xe9c1; BYTE $0x10
	WORD         $0xb70f; BYTE $0xc0
	CMPQ         DX, $0x20
	JB           tail16_check
	LONG         $0x078071bf; BYTE $0x80
	LONG         $0xc0eff9c5
	VMOVDQA      LCPI0_0<>(SB), Y1
	VPBROADCASTW LCPI0_2<>(SB), Y2
	JMP          block_loop_setup

block_accum_init:
	LONG $0xf46ffdc5
	LONG $0xedefd1c5

block_reduce:
	SUBQ  AX, DX
	LONG  $0xf572ddc5; BYTE $0x05
	LONG  $0xdbfeddc5
	LONG  $0x397de3c4; WORD $0x01f4
	LONG  $0xecc6c8c5; BYTE $0x88
	LONG  $0xe470f9c5; BYTE $0x88
	LONG  $0xe4fed1c5
	LONG  $0xec70f9c5; BYTE $0x55
	LONG  $0xe4fed1c5
	LONG  $0xe07ef9c5
	MOVQ  AX, CX
	IMULQ DI, CX
	SHRQ  $0x2f, CX
	LONG  $0xfff1c969; WORD $0x0000
	WORD  $0xc829
	LONG  $0x397de3c4; WORD $0x01dc
	LONG  $0xdbfed9c5
	LONG  $0xe370f9c5; BYTE $0xee
	LONG  $0xdcfee1c5
	LONG  $0xe370f9c5; BYTE $0x55
	LONG  $0xdbfed9c5
	LONG  $0xd97ef9c5
	MOVQ  CX, R8
	IMULQ DI, R8
	SHRQ  $0x2f, R8
	LONG  $0xf1c06945; WORD $0x00ff; BYTE $0x00
	WORD  $0x2944; BYTE $0xc1
	CMPQ  DX, $0x1f
	JBE   tail_check

block_loop_setup:
	LONG $0xe06ef9c5
	LONG $0xd96ef9c5
	CMPQ DX, $0x15b0
	LONG $0x15b0b841; WORD $0x0000
	LONG $0xc2420f4c
	WORD $0x8944; BYTE $0xc0
	LONG $0x001fe025; BYTE $0x00
	JE   block_accum_init
	ADDQ $-0x20, R8
	LONG $0xedefd1c5
	LONG $0x20c0f641
	JNE  block_loop_entry
	LONG $0x2e6ffec5
	ADDQ $0x20, SI
	LEAQ -0x20(AX), CX
	LONG $0xf0f6d5c5
	LONG $0xf4fecdc5
	LONG $0x0455e2c4; BYTE $0xe9
	LONG $0xeaf5d5c5
	LONG $0xdbfed5c5
	LONG $0xec6ffdc5
	LONG $0xe66ffdc5
	CMPQ R8, $0x20
	JAE  block_loop_64
	JMP  block_reduce

block_loop_entry:
	MOVQ AX, CX
	CMPQ R8, $0x20
	JB   block_reduce

block_loop_64:
	LONG $0x366ffec5
	LONG $0x7e6ffec5; BYTE $0x20
	LONG $0xc0f64dc5
	LONG $0xc4fe3dc5
	LONG $0xecfed5c5
	LONG $0x044de2c4; BYTE $0xe1
	LONG $0xe2f5ddc5
	LONG $0xdbfeddc5
	ADDQ $0x40, SI
	LONG $0xe0f6c5c5
	LONG $0xe4febdc5
	LONG $0xedfebdc5
	LONG $0x0445e2c4; BYTE $0xf1
	LONG $0xf2f5cdc5
	LONG $0xdbfecdc5
	ADDQ $-0x40, CX
	JNE  block_loop_64
	LONG $0xf46ffdc5
	JMP  block_reduce

return_one:
	LONG $0x000001b8; BYTE $0x00

return_result:
	MOVL AX, ret+32(FP)
	RET

tail_check:
	WORD $0x8548; BYTE $0xd2
	JE   return_no_tail

tail16_check:
	CMPQ DX, $0x10
	JB   tail_bytes_setup
	WORD $0xb60f; BYTE $0x3e
	WORD $0xf801
	WORD $0xc101
	LONG $0x017eb60f
	WORD $0xc701
	WORD $0xf901
	LONG $0x0246b60f
	WORD $0xf801
	WORD $0xc101
	LONG $0x037eb60f
	WORD $0xc701
	WORD $0xf901
	LONG $0x0446b60f
	WORD $0xf801
	WORD $0xc101
	LONG $0x057eb60f
	WORD $0xc701
	WORD $0xf901
	LONG $0x0646b60f
	WORD $0xf801
	WORD $0xc101
	LONG $0x077eb60f
	WORD $0xc701
	WORD $0xf901
	LONG $0x0846b60f
	WORD $0xf801
	WORD $0xc101
	LONG $0x097eb60f
	WORD $0xc701
	WORD $0xf901
	LONG $0x0a46b60f
	WORD $0xf801
	WORD $0xc101
	LONG $0x0b7eb60f
	WORD $0xc701
	WORD $0xf901
	LONG $0x0c46b60f
	WORD $0xf801
	WORD $0xc101
	LONG $0x0d7eb60f
	WORD $0xc701
	WORD $0xf901
	LONG $0x46b60f44; BYTE $0x0e
	WORD $0x0141; BYTE $0xf8
	WORD $0x0144; BYTE $0xc1
	LONG $0x0f46b60f
	WORD $0x0144; BYTE $0xc0
	WORD $0xc101
	ADDQ $-0x10, DX
	JE   final_reduce
	ADDQ $0x10, SI

tail_bytes_setup:
	LEAQ -0x1(DX), DI
	MOVQ DX, R9
	ANDQ $0x3, R9
	JE   tail_dword_setup
	XORL R8, R8

tail_byte_loop:
	LONG $0x14b60f46; BYTE $0x06
	WORD $0x0144; BYTE $0xd0
	WORD $0xc101
	INCQ R8
	CMPQ R9, R8
	JNE  tail_byte_loop
	ADDQ R8, SI
	SUBQ R8, DX

tail_dword_setup:
	CMPQ DI, $0x3
	JB   final_reduce
	XORL DI, DI

tail_dword_loop:
	LONG $0x04b60f44; BYTE $0x3e
	WORD $0x0141; BYTE $0xc0
	WORD $0x0144; BYTE $0xc1
	LONG $0x3e44b60f; BYTE $0x01
	WORD $0x0144; BYTE $0xc0
	WORD $0xc101
	LONG $0x44b60f44; WORD $0x023e
	WORD $0x0141; BYTE $0xc0
	WORD $0x0144; BYTE $0xc1
	LONG $0x3e44b60f; BYTE $0x03
	WORD $0x0144; BYTE $0xc0
	WORD $0xc101
	ADDQ $0x4, DI
	CMPQ DX, DI
	JNE  tail_dword_loop

final_reduce:
	LONG  $0x000f908d; WORD $0xffff
	CMPL  AX, $0xfff1
	WORD  $0x420f; BYTE $0xd0
	WORD  $0xc889
	LONG  $0x078071be; BYTE $0x80
	IMULQ AX, SI
	SHRQ  $0x2f, SI
	LONG  $0xfff1c669; WORD $0x0000
	WORD  $0xc129
	WORD  $0xe1c1; BYTE $0x10
	WORD  $0xd109
	WORD  $0xc889
	NOP
	NOP
	VZEROUPPER
	MOVL  AX, ret+32(FP)
	RET

return_no_tail:
	WORD $0xe1c1; BYTE $0x10
	WORD $0xc809
	NOP
	NOP
	VZEROUPPER
	MOVL AX, ret+32(FP)
	RET
