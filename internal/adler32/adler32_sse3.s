//go:build !purego && amd64

#include "textflag.h"

DATA weights_17_32<>+0x00(SB)/8, $0x191a1b1c1d1e1f20
DATA weights_17_32<>+0x08(SB)/8, $0x1112131415161718
GLOBL weights_17_32<>(SB), (RODATA|NOPTR), $16

DATA ones_u16<>+0x00(SB)/8, $0x0001000100010001
DATA ones_u16<>+0x08(SB)/8, $0x0001000100010001
GLOBL ones_u16<>(SB), (RODATA|NOPTR), $16

DATA weights_1_16<>+0x00(SB)/8, $0x090a0b0c0d0e0f10
DATA weights_1_16<>+0x08(SB)/8, $0x0102030405060708
GLOBL weights_1_16<>(SB), (RODATA|NOPTR), $16

TEXT ·adler32_sse3(SB), NOSPLIT, $0-36
	MOVLQZX in+0(FP), DI
	MOVQ    buf_base+8(FP), SI
	MOVQ    buf_len+16(FP), DX
	MOVQ    buf_cap+24(FP), CX
	NOP
	NOP
	NOP
	WORD    $0xf889
	LONG    $0xc8b70f44
	WORD    $0xe8c1; BYTE $0x10
	WORD    $0xd189
	WORD    $0xe183; BYTE $0x1f
	CMPQ    DX, $0x20
	JAE     block_loop_setup
	WORD    $0x8944; BYTE $0xcf
	JMP     tail_entry

block_loop_setup:
	SHRQ $0x5, DX
	LONG $0xc0ef0f66
	MOVO weights_17_32<>(SB), X1
	MOVO ones_u16<>(SB), X2
	MOVO weights_1_16<>(SB), X3
	LONG $0x8071b841; WORD $0x8007

block_outer_loop:
	CMPQ DX, $0xad
	LONG $0x00adba41; WORD $0x0000
	LONG $0xd2420f4c
	WORD $0x8944; BYTE $0xcf
	LONG $0xfaaf0f41
	LONG $0xef6e0f66
	LONG $0xe06e0f66
	WORD $0x8944; BYTE $0xd0
	LONG $0xf6ef0f66

block_inner_loop:
	LONG  $0x3e6f0ff3
	LONG  $0x6f0f4466; BYTE $0xc7
	LONG  $0x04380f66; BYTE $0xf9
	LONG  $0xfaf50f66
	LONG  $0xfcfe0f66
	LONG  $0x666f0ff3; BYTE $0x10
	LONG  $0xeefe0f66
	LONG  $0xf60f4466; BYTE $0xc0
	LONG  $0xfe0f4466; BYTE $0xc6
	LONG  $0xf46f0f66
	LONG  $0xf0f60f66
	LONG  $0xfe0f4166; BYTE $0xf0
	LONG  $0x04380f66; BYTE $0xe3
	LONG  $0xe2f50f66
	LONG  $0xe7fe0f66
	ADDQ  $0x20, SI
	WORD  $0xc8ff
	JNE   block_inner_loop
	LONG  $0xf5720f66; BYTE $0x05
	LONG  $0xe5fe0f66
	LONG  $0xee700f66; BYTE $0xb1
	LONG  $0xeefe0f66
	LONG  $0xf5700f66; BYTE $0xee
	LONG  $0xf5fe0f66
	LONG  $0xf77e0f66
	WORD  $0x0144; BYTE $0xcf
	LONG  $0xec700f66; BYTE $0xb1
	LONG  $0xecfe0f66
	LONG  $0xe5700f66; BYTE $0xee
	LONG  $0xe5fe0f66
	LONG  $0xe07e0f66
	MOVQ  DI, R9
	IMULQ R8, R9
	SHRQ  $0x2f, R9
	LONG  $0xf1c96945; WORD $0x00ff; BYTE $0x00
	WORD  $0x2944; BYTE $0xcf
	MOVQ  AX, R9
	IMULQ R8, R9
	SHRQ  $0x2f, R9
	LONG  $0xf1c96945; WORD $0x00ff; BYTE $0x00
	WORD  $0x2944; BYTE $0xc8
	WORD  $0x8941; BYTE $0xf9
	SUBQ  R10, DX
	JNE   block_outer_loop

tail_entry:
	WORD $0x8548; BYTE $0xc9
	JE   return_result
	CMPL CX, $0x10
	JB   tail_bytes_setup
	WORD $0xb60f; BYTE $0x16
	WORD $0xd701
	WORD $0xf801
	LONG $0x0156b60f
	WORD $0xfa01
	WORD $0xd001
	LONG $0x027eb60f
	WORD $0xd701
	WORD $0xf801
	LONG $0x0356b60f
	WORD $0xfa01
	WORD $0xd001
	LONG $0x047eb60f
	WORD $0xd701
	WORD $0xf801
	LONG $0x0556b60f
	WORD $0xfa01
	WORD $0xd001
	LONG $0x067eb60f
	WORD $0xd701
	WORD $0xf801
	LONG $0x0756b60f
	WORD $0xfa01
	WORD $0xd001
	LONG $0x087eb60f
	WORD $0xd701
	WORD $0xf801
	LONG $0x0956b60f
	WORD $0xfa01
	WORD $0xd001
	LONG $0x0a7eb60f
	WORD $0xd701
	WORD $0xf801
	LONG $0x0b56b60f
	WORD $0xfa01
	WORD $0xd001
	LONG $0x0c7eb60f
	WORD $0xd701
	WORD $0xf801
	LONG $0x0d56b60f
	WORD $0xfa01
	WORD $0xd001
	LONG $0x46b60f44; BYTE $0x0e
	WORD $0x0141; BYTE $0xd0
	WORD $0x0144; BYTE $0xc0
	LONG $0x0f7eb60f
	WORD $0x0144; BYTE $0xc7
	WORD $0xf801
	ADDQ $-0x10, CX
	JE   final_reduce
	ADDQ $0x10, SI

tail_bytes_setup:
	LEAQ -0x1(CX), DX
	MOVQ CX, R9
	ANDQ $0x3, R9
	JE   tail_dword_setup
	XORL R8, R8

tail_byte_loop:
	LONG $0x14b60f46; BYTE $0x06
	WORD $0x0144; BYTE $0xd7
	WORD $0xf801
	INCQ R8
	CMPQ R9, R8
	JNE  tail_byte_loop
	ADDQ R8, SI
	SUBQ R8, CX

tail_dword_setup:
	CMPQ DX, $0x3
	JB   final_reduce
	XORL DX, DX

tail_dword_loop:
	LONG $0x04b60f44; BYTE $0x16
	WORD $0x0141; BYTE $0xf8
	WORD $0x0144; BYTE $0xc0
	LONG $0x167cb60f; BYTE $0x01
	WORD $0x0144; BYTE $0xc7
	WORD $0xf801
	LONG $0x44b60f44; WORD $0x0216
	WORD $0x0141; BYTE $0xf8
	WORD $0x0144; BYTE $0xc0
	LONG $0x167cb60f; BYTE $0x03
	WORD $0x0144; BYTE $0xc7
	WORD $0xf801
	ADDQ $0x4, DX
	CMPQ CX, DX
	JNE  tail_dword_loop

final_reduce:
	LONG  $0x000f8f8d; WORD $0xffff
	CMPL  DI, $0xfff1
	WORD  $0x420f; BYTE $0xcf
	WORD  $0xc289
	LONG  $0x078071be; BYTE $0x80
	IMULQ DX, SI
	SHRQ  $0x2f, SI
	LONG  $0xfff1d669; WORD $0x0000
	WORD  $0xd029
	WORD  $0xcf89

return_result:
	WORD $0xe0c1; BYTE $0x10
	WORD $0xf809
	NOP
	NOP
	MOVL AX, ret+32(FP)
	RET
