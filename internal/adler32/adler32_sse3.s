//go:build !noasm && amd64

#include "textflag.h"

DATA LCPI0_0<>+0x00(SB)/8, $0x191a1b1c1d1e1f20
DATA LCPI0_0<>+0x08(SB)/8, $0x1112131415161718
GLOBL LCPI0_0<>(SB), (RODATA|NOPTR), $16

DATA LCPI0_1<>+0x00(SB)/8, $0x0001000100010001
DATA LCPI0_1<>+0x08(SB)/8, $0x0001000100010001
GLOBL LCPI0_1<>(SB), (RODATA|NOPTR), $16

DATA LCPI0_2<>+0x00(SB)/8, $0x090a0b0c0d0e0f10
DATA LCPI0_2<>+0x08(SB)/8, $0x0102030405060708
GLOBL LCPI0_2<>(SB), (RODATA|NOPTR), $16

TEXT ·adler32_sse3(SB), NOSPLIT, $0-36
	MOVLQZX adler+0(FP), DI
	MOVQ    buf+8(FP), SI
	MOVQ    buf_len+16(FP), DX
	MOVQ    buf_cap+24(FP), CX
	NOP                         // (skipped)                            // push	rbp
	NOP                         // (skipped)                            // mov	rbp, rsp
	NOP                         // (skipped)                            // and	rsp, -8
	WORD    $0xf889             // MOVL DI, AX                          // mov	eax, edi
	LONG    $0xc8b70f44         // MOVZX AX, R9                         // movzx	r9d, ax
	WORD    $0xe8c1; BYTE $0x10 // SHRL $0x10, AX                       // shr	eax, 16
	WORD    $0xd189             // MOVL DX, CX                          // mov	ecx, edx
	WORD    $0xe183; BYTE $0x1f // ANDL $0x1f, CX                       // and	ecx, 31
	CMPQ    DX, $0x20           // <--                                  // cmp	rdx, 32
	JAE     LBB0_2              // <--                                  // jae	.LBB0_2
	WORD    $0x8944; BYTE $0xcf // MOVL R9, DI                          // mov	edi, r9d
	JMP     LBB0_6              // <--                                  // jmp	.LBB0_6

LBB0_2:
	SHRQ $0x5, DX                  // <--                                  // shr	rdx, 5
	LONG $0xc0ef0f66               // PXOR X0, X0                          // pxor	xmm0, xmm0
	MOVO LCPI0_0<>(SB), X1         // <--                                  // movdqa	xmm1, xmmword ptr [rip + .LCPI0_0]
	MOVO LCPI0_1<>(SB), X2         // <--                                  // movdqa	xmm2, xmmword ptr [rip + .LCPI0_1]
	MOVO LCPI0_2<>(SB), X3         // <--                                  // movdqa	xmm3, xmmword ptr [rip + .LCPI0_2]
	LONG $0x8071b841; WORD $0x8007 // MOVL $-0x7ff87f8f, R8                // mov	r8d, 2147975281

LBB0_3:
	CMPQ DX, $0xad                 // <--                                  // cmp	rdx, 173
	LONG $0x00adba41; WORD $0x0000 // MOVL $0xad, R10                      // mov	r10d, 173
	LONG $0xd2420f4c               // CMOVB DX, R10                        // cmovb	r10, rdx
	WORD $0x8944; BYTE $0xcf       // MOVL R9, DI                          // mov	edi, r9d
	LONG $0xfaaf0f41               // IMULL R10, DI                        // imul	edi, r10d
	LONG $0xef6e0f66               // MOVD DI, X5                          // movd	xmm5, edi
	LONG $0xe06e0f66               // MOVD AX, X4                          // movd	xmm4, eax
	WORD $0x8944; BYTE $0xd0       // MOVL R10, AX                         // mov	eax, r10d
	LONG $0xf6ef0f66               // PXOR X6, X6                          // pxor	xmm6, xmm6

LBB0_4:
	LONG  $0x3e6f0ff3                           // MOVDQU 0(SI), X7                     // movdqu	xmm7, xmmword ptr [rsi]
	LONG  $0x6f0f4466; BYTE $0xc7               // MOVDQA X7, X8                        // movdqa	xmm8, xmm7
	LONG  $0x04380f66; BYTE $0xf9               // PMADDUBSW X1, X7                     // pmaddubsw	xmm7, xmm1
	LONG  $0xfaf50f66                           // PMADDWD X2, X7                       // pmaddwd	xmm7, xmm2
	LONG  $0xfcfe0f66                           // PADDD X4, X7                         // paddd	xmm7, xmm4
	LONG  $0x666f0ff3; BYTE $0x10               // MOVDQU 0x10(SI), X4                  // movdqu	xmm4, xmmword ptr [rsi + 16]
	LONG  $0xeefe0f66                           // PADDD X6, X5                         // paddd	xmm5, xmm6
	LONG  $0xf60f4466; BYTE $0xc0               // PSADBW X0, X8                        // psadbw	xmm8, xmm0
	LONG  $0xfe0f4466; BYTE $0xc6               // PADDD X6, X8                         // paddd	xmm8, xmm6
	LONG  $0xf46f0f66                           // MOVDQA X4, X6                        // movdqa	xmm6, xmm4
	LONG  $0xf0f60f66                           // PSADBW X0, X6                        // psadbw	xmm6, xmm0
	LONG  $0xfe0f4166; BYTE $0xf0               // PADDD X8, X6                         // paddd	xmm6, xmm8
	LONG  $0x04380f66; BYTE $0xe3               // PMADDUBSW X3, X4                     // pmaddubsw	xmm4, xmm3
	LONG  $0xe2f50f66                           // PMADDWD X2, X4                       // pmaddwd	xmm4, xmm2
	LONG  $0xe7fe0f66                           // PADDD X7, X4                         // paddd	xmm4, xmm7
	ADDQ  $0x20, SI                             // <--                                  // add	rsi, 32
	WORD  $0xc8ff                               // DECL AX                              // dec	eax
	JNE   LBB0_4                                // <--                                  // jne	.LBB0_4
	LONG  $0xf5720f66; BYTE $0x05               // PSLLD $0x5, X5                       // pslld	xmm5, 5
	LONG  $0xe5fe0f66                           // PADDD X5, X4                         // paddd	xmm4, xmm5
	LONG  $0xee700f66; BYTE $0xb1               // PSHUFD $0xb1, X6, X5                 // pshufd	xmm5, xmm6, 177
	LONG  $0xeefe0f66                           // PADDD X6, X5                         // paddd	xmm5, xmm6
	LONG  $0xf5700f66; BYTE $0xee               // PSHUFD $0xee, X5, X6                 // pshufd	xmm6, xmm5, 238
	LONG  $0xf5fe0f66                           // PADDD X5, X6                         // paddd	xmm6, xmm5
	LONG  $0xf77e0f66                           // MOVD X6, DI                          // movd	edi, xmm6
	WORD  $0x0144; BYTE $0xcf                   // ADDL R9, DI                          // add	edi, r9d
	LONG  $0xec700f66; BYTE $0xb1               // PSHUFD $0xb1, X4, X5                 // pshufd	xmm5, xmm4, 177
	LONG  $0xecfe0f66                           // PADDD X4, X5                         // paddd	xmm5, xmm4
	LONG  $0xe5700f66; BYTE $0xee               // PSHUFD $0xee, X5, X4                 // pshufd	xmm4, xmm5, 238
	LONG  $0xe5fe0f66                           // PADDD X5, X4                         // paddd	xmm4, xmm5
	LONG  $0xe07e0f66                           // MOVD X4, AX                          // movd	eax, xmm4
	MOVQ  DI, R9                                // <--                                  // mov	r9, rdi
	IMULQ R8, R9                                // <--                                  // imul	r9, r8
	SHRQ  $0x2f, R9                             // <--                                  // shr	r9, 47
	LONG  $0xf1c96945; WORD $0x00ff; BYTE $0x00 // IMULL $0xfff1, R9, R9                // imul	r9d, r9d, 65521
	WORD  $0x2944; BYTE $0xcf                   // SUBL R9, DI                          // sub	edi, r9d
	MOVQ  AX, R9                                // <--                                  // mov	r9, rax
	IMULQ R8, R9                                // <--                                  // imul	r9, r8
	SHRQ  $0x2f, R9                             // <--                                  // shr	r9, 47
	LONG  $0xf1c96945; WORD $0x00ff; BYTE $0x00 // IMULL $0xfff1, R9, R9                // imul	r9d, r9d, 65521
	WORD  $0x2944; BYTE $0xc8                   // SUBL R9, AX                          // sub	eax, r9d
	WORD  $0x8941; BYTE $0xf9                   // MOVL DI, R9                          // mov	r9d, edi
	SUBQ  R10, DX                               // <--                                  // sub	rdx, r10
	JNE   LBB0_3                                // <--                                  // jne	.LBB0_3

LBB0_6:
	WORD $0x8548; BYTE $0xc9     // TESTQ CX, CX                         // test	rcx, rcx
	JE   LBB0_18                 // <--                                  // je	.LBB0_18
	CMPL CX, $0x10               // <--                                  // cmp	ecx, 16
	JB   LBB0_10                 // <--                                  // jb	.LBB0_10
	WORD $0xb60f; BYTE $0x16     // MOVZX 0(SI), DX                      // movzx	edx, byte ptr [rsi]
	WORD $0xd701                 // ADDL DX, DI                          // add	edi, edx
	WORD $0xf801                 // ADDL DI, AX                          // add	eax, edi
	LONG $0x0156b60f             // MOVZX 0x1(SI), DX                    // movzx	edx, byte ptr [rsi + 1]
	WORD $0xfa01                 // ADDL DI, DX                          // add	edx, edi
	WORD $0xd001                 // ADDL DX, AX                          // add	eax, edx
	LONG $0x027eb60f             // MOVZX 0x2(SI), DI                    // movzx	edi, byte ptr [rsi + 2]
	WORD $0xd701                 // ADDL DX, DI                          // add	edi, edx
	WORD $0xf801                 // ADDL DI, AX                          // add	eax, edi
	LONG $0x0356b60f             // MOVZX 0x3(SI), DX                    // movzx	edx, byte ptr [rsi + 3]
	WORD $0xfa01                 // ADDL DI, DX                          // add	edx, edi
	WORD $0xd001                 // ADDL DX, AX                          // add	eax, edx
	LONG $0x047eb60f             // MOVZX 0x4(SI), DI                    // movzx	edi, byte ptr [rsi + 4]
	WORD $0xd701                 // ADDL DX, DI                          // add	edi, edx
	WORD $0xf801                 // ADDL DI, AX                          // add	eax, edi
	LONG $0x0556b60f             // MOVZX 0x5(SI), DX                    // movzx	edx, byte ptr [rsi + 5]
	WORD $0xfa01                 // ADDL DI, DX                          // add	edx, edi
	WORD $0xd001                 // ADDL DX, AX                          // add	eax, edx
	LONG $0x067eb60f             // MOVZX 0x6(SI), DI                    // movzx	edi, byte ptr [rsi + 6]
	WORD $0xd701                 // ADDL DX, DI                          // add	edi, edx
	WORD $0xf801                 // ADDL DI, AX                          // add	eax, edi
	LONG $0x0756b60f             // MOVZX 0x7(SI), DX                    // movzx	edx, byte ptr [rsi + 7]
	WORD $0xfa01                 // ADDL DI, DX                          // add	edx, edi
	WORD $0xd001                 // ADDL DX, AX                          // add	eax, edx
	LONG $0x087eb60f             // MOVZX 0x8(SI), DI                    // movzx	edi, byte ptr [rsi + 8]
	WORD $0xd701                 // ADDL DX, DI                          // add	edi, edx
	WORD $0xf801                 // ADDL DI, AX                          // add	eax, edi
	LONG $0x0956b60f             // MOVZX 0x9(SI), DX                    // movzx	edx, byte ptr [rsi + 9]
	WORD $0xfa01                 // ADDL DI, DX                          // add	edx, edi
	WORD $0xd001                 // ADDL DX, AX                          // add	eax, edx
	LONG $0x0a7eb60f             // MOVZX 0xa(SI), DI                    // movzx	edi, byte ptr [rsi + 10]
	WORD $0xd701                 // ADDL DX, DI                          // add	edi, edx
	WORD $0xf801                 // ADDL DI, AX                          // add	eax, edi
	LONG $0x0b56b60f             // MOVZX 0xb(SI), DX                    // movzx	edx, byte ptr [rsi + 11]
	WORD $0xfa01                 // ADDL DI, DX                          // add	edx, edi
	WORD $0xd001                 // ADDL DX, AX                          // add	eax, edx
	LONG $0x0c7eb60f             // MOVZX 0xc(SI), DI                    // movzx	edi, byte ptr [rsi + 12]
	WORD $0xd701                 // ADDL DX, DI                          // add	edi, edx
	WORD $0xf801                 // ADDL DI, AX                          // add	eax, edi
	LONG $0x0d56b60f             // MOVZX 0xd(SI), DX                    // movzx	edx, byte ptr [rsi + 13]
	WORD $0xfa01                 // ADDL DI, DX                          // add	edx, edi
	WORD $0xd001                 // ADDL DX, AX                          // add	eax, edx
	LONG $0x46b60f44; BYTE $0x0e // MOVZX 0xe(SI), R8                    // movzx	r8d, byte ptr [rsi + 14]
	WORD $0x0141; BYTE $0xd0     // ADDL DX, R8                          // add	r8d, edx
	WORD $0x0144; BYTE $0xc0     // ADDL R8, AX                          // add	eax, r8d
	LONG $0x0f7eb60f             // MOVZX 0xf(SI), DI                    // movzx	edi, byte ptr [rsi + 15]
	WORD $0x0144; BYTE $0xc7     // ADDL R8, DI                          // add	edi, r8d
	WORD $0xf801                 // ADDL DI, AX                          // add	eax, edi
	ADDQ $-0x10, CX              // <--                                  // add	rcx, -16
	JE   LBB0_17                 // <--                                  // je	.LBB0_17
	ADDQ $0x10, SI               // <--                                  // add	rsi, 16

LBB0_10:
	LEAQ -0x1(CX), DX // <--                                  // lea	rdx, [rcx - 1]
	MOVQ CX, R9       // <--                                  // mov	r9, rcx
	ANDQ $0x3, R9     // <--                                  // and	r9, 3
	JE   LBB0_14      // <--                                  // je	.LBB0_14
	XORL R8, R8       // <--                                  // xor	r8d, r8d

LBB0_12:
	LONG $0x14b60f46; BYTE $0x06 // MOVZX 0(SI)(R8*1), R10               // movzx	r10d, byte ptr [rsi + r8]
	WORD $0x0144; BYTE $0xd7     // ADDL R10, DI                         // add	edi, r10d
	WORD $0xf801                 // ADDL DI, AX                          // add	eax, edi
	INCQ R8                      // <--                                  // inc	r8
	CMPQ R9, R8                  // <--                                  // cmp	r9, r8
	JNE  LBB0_12                 // <--                                  // jne	.LBB0_12
	ADDQ R8, SI                  // <--                                  // add	rsi, r8
	SUBQ R8, CX                  // <--                                  // sub	rcx, r8

LBB0_14:
	CMPQ DX, $0x3 // <--                                  // cmp	rdx, 3
	JB   LBB0_17  // <--                                  // jb	.LBB0_17
	XORL DX, DX   // <--                                  // xor	edx, edx

LBB0_16:
	LONG $0x04b60f44; BYTE $0x16   // MOVZX 0(SI)(DX*1), R8                // movzx	r8d, byte ptr [rsi + rdx]
	WORD $0x0141; BYTE $0xf8       // ADDL DI, R8                          // add	r8d, edi
	WORD $0x0144; BYTE $0xc0       // ADDL R8, AX                          // add	eax, r8d
	LONG $0x167cb60f; BYTE $0x01   // MOVZX 0x1(SI)(DX*1), DI              // movzx	edi, byte ptr [rsi + rdx + 1]
	WORD $0x0144; BYTE $0xc7       // ADDL R8, DI                          // add	edi, r8d
	WORD $0xf801                   // ADDL DI, AX                          // add	eax, edi
	LONG $0x44b60f44; WORD $0x0216 // MOVZX 0x2(SI)(DX*1), R8              // movzx	r8d, byte ptr [rsi + rdx + 2]
	WORD $0x0141; BYTE $0xf8       // ADDL DI, R8                          // add	r8d, edi
	WORD $0x0144; BYTE $0xc0       // ADDL R8, AX                          // add	eax, r8d
	LONG $0x167cb60f; BYTE $0x03   // MOVZX 0x3(SI)(DX*1), DI              // movzx	edi, byte ptr [rsi + rdx + 3]
	WORD $0x0144; BYTE $0xc7       // ADDL R8, DI                          // add	edi, r8d
	WORD $0xf801                   // ADDL DI, AX                          // add	eax, edi
	ADDQ $0x4, DX                  // <--                                  // add	rdx, 4
	CMPQ CX, DX                    // <--                                  // cmp	rcx, rdx
	JNE  LBB0_16                   // <--                                  // jne	.LBB0_16

LBB0_17:
	LONG  $0x000f8f8d; WORD $0xffff // LEAL -0xfff1(DI), CX                 // lea	ecx, [rdi - 65521]
	CMPL  DI, $0xfff1               // <--                                  // cmp	edi, 65521
	WORD  $0x420f; BYTE $0xcf       // CMOVB DI, CX                         // cmovb	ecx, edi
	WORD  $0xc289                   // MOVL AX, DX                          // mov	edx, eax
	LONG  $0x078071be; BYTE $0x80   // MOVL $-0x7ff87f8f, SI                // mov	esi, 2147975281
	IMULQ DX, SI                    // <--                                  // imul	rsi, rdx
	SHRQ  $0x2f, SI                 // <--                                  // shr	rsi, 47
	LONG  $0xfff1d669; WORD $0x0000 // IMULL $0xfff1, SI, DX                // imul	edx, esi, 65521
	WORD  $0xd029                   // SUBL DX, AX                          // sub	eax, edx
	WORD  $0xcf89                   // MOVL CX, DI                          // mov	edi, ecx

LBB0_18:
	WORD $0xe0c1; BYTE $0x10 // SHLL $0x10, AX                       // shl	eax, 16
	WORD $0xf809             // ORL DI, AX                           // or	eax, edi
	NOP                      // (skipped)                            // mov	rsp, rbp
	NOP                      // (skipped)                            // pop	rbp
	MOVL AX, ret+32(FP)      // <--
	RET                      // <--                                  // ret
