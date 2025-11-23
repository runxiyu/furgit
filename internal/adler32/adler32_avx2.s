//go:build !noasm && amd64

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
	WORD         $0x8548; BYTE $0xf6     // TESTQ SI, SI                         // test	rsi, rsi
	JE           LBB0_1                  // <--                                  // je	.LBB0_1
	WORD         $0xf889                 // MOVL DI, AX                          // mov	eax, edi
	WORD         $0x8548; BYTE $0xd2     // TESTQ DX, DX                         // test	rdx, rdx
	JE           LBB0_2                  // <--                                  // je	.LBB0_2
	NOP                                  // (skipped)                            // push	rbp
	NOP                                  // (skipped)                            // mov	rbp, rsp
	NOP                                  // (skipped)                            // and	rsp, -8
	WORD         $0xc189                 // MOVL AX, CX                          // mov	ecx, eax
	WORD         $0xe9c1; BYTE $0x10     // SHRL $0x10, CX                       // shr	ecx, 16
	WORD         $0xb70f; BYTE $0xc0     // MOVZX AX, AX                         // movzx	eax, ax
	CMPQ         DX, $0x20               // <--                                  // cmp	rdx, 32
	JB           LBB0_17                 // <--                                  // jb	.LBB0_17
	LONG         $0x078071bf; BYTE $0x80 // MOVL $-0x7ff87f8f, DI                // mov	edi, 2147975281
	LONG         $0xc0eff9c5             // VPXOR X0, X0, X0                     // vpxor	xmm0, xmm0, xmm0
	VMOVDQA      LCPI0_0<>(SB), Y1       // <--                                  // vmovdqa	ymm1, ymmword ptr [rip + .LCPI0_0]
	VPBROADCASTW LCPI0_2<>(SB), Y2       // <--                                  // vpbroadcastw	ymm2, word ptr [rip + .LCPI0_2]
	JMP          LBB0_6                  // <--                                  // jmp	.LBB0_6

LBB0_7:
	LONG $0xf46ffdc5 // VMOVDQA Y4, Y6                       // vmovdqa	ymm6, ymm4
	LONG $0xedefd1c5 // VPXOR X5, X5, X5                     // vpxor	xmm5, xmm5, xmm5

LBB0_14:
	SUBQ  AX, DX                                // <--                                  // sub	rdx, rax
	LONG  $0xf572ddc5; BYTE $0x05               // ?                                    // vpslld	ymm4, ymm5, 5
	LONG  $0xdbfeddc5                           // VPADDD Y3, Y4, Y3                    // vpaddd	ymm3, ymm4, ymm3
	LONG  $0x397de3c4; WORD $0x01f4             // VEXTRACTI128 $0x1, Y6, X4            // vextracti128	xmm4, ymm6, 1
	LONG  $0xecc6c8c5; BYTE $0x88               // VSHUFPS $-0x78, X4, X6, X5           // vshufps	xmm5, xmm6, xmm4, 136
	LONG  $0xe470f9c5; BYTE $0x88               // VPSHUFD $-0x78, X4, X4               // vpshufd	xmm4, xmm4, 136
	LONG  $0xe4fed1c5                           // VPADDD X4, X5, X4                    // vpaddd	xmm4, xmm5, xmm4
	LONG  $0xec70f9c5; BYTE $0x55               // VPSHUFD $0x55, X4, X5                // vpshufd	xmm5, xmm4, 85
	LONG  $0xe4fed1c5                           // VPADDD X4, X5, X4                    // vpaddd	xmm4, xmm5, xmm4
	LONG  $0xe07ef9c5                           // VMOVD X4, AX                         // vmovd	eax, xmm4
	MOVQ  AX, CX                                // <--                                  // mov	rcx, rax
	IMULQ DI, CX                                // <--                                  // imul	rcx, rdi
	SHRQ  $0x2f, CX                             // <--                                  // shr	rcx, 47
	LONG  $0xfff1c969; WORD $0x0000             // IMULL $0xfff1, CX, CX                // imul	ecx, ecx, 65521
	WORD  $0xc829                               // SUBL CX, AX                          // sub	eax, ecx
	LONG  $0x397de3c4; WORD $0x01dc             // VEXTRACTI128 $0x1, Y3, X4            // vextracti128	xmm4, ymm3, 1
	LONG  $0xdbfed9c5                           // VPADDD X3, X4, X3                    // vpaddd	xmm3, xmm4, xmm3
	LONG  $0xe370f9c5; BYTE $0xee               // VPSHUFD $-0x12, X3, X4               // vpshufd	xmm4, xmm3, 238
	LONG  $0xdcfee1c5                           // VPADDD X4, X3, X3                    // vpaddd	xmm3, xmm3, xmm4
	LONG  $0xe370f9c5; BYTE $0x55               // VPSHUFD $0x55, X3, X4                // vpshufd	xmm4, xmm3, 85
	LONG  $0xdbfed9c5                           // VPADDD X3, X4, X3                    // vpaddd	xmm3, xmm4, xmm3
	LONG  $0xd97ef9c5                           // VMOVD X3, CX                         // vmovd	ecx, xmm3
	MOVQ  CX, R8                                // <--                                  // mov	r8, rcx
	IMULQ DI, R8                                // <--                                  // imul	r8, rdi
	SHRQ  $0x2f, R8                             // <--                                  // shr	r8, 47
	LONG  $0xf1c06945; WORD $0x00ff; BYTE $0x00 // IMULL $0xfff1, R8, R8                // imul	r8d, r8d, 65521
	WORD  $0x2944; BYTE $0xc1                   // SUBL R8, CX                          // sub	ecx, r8d
	CMPQ  DX, $0x1f                             // <--                                  // cmp	rdx, 31
	JBE   LBB0_15                               // <--                                  // jbe	.LBB0_15

LBB0_6:
	LONG $0xe06ef9c5               // VMOVD AX, X4                         // vmovd	xmm4, eax
	LONG $0xd96ef9c5               // VMOVD CX, X3                         // vmovd	xmm3, ecx
	CMPQ DX, $0x15b0               // <--                                  // cmp	rdx, 5552
	LONG $0x15b0b841; WORD $0x0000 // MOVL $0x15b0, R8                     // mov	r8d, 5552
	LONG $0xc2420f4c               // CMOVB DX, R8                         // cmovb	r8, rdx
	WORD $0x8944; BYTE $0xc0       // MOVL R8, AX                          // mov	eax, r8d
	LONG $0x001fe025; BYTE $0x00   // ANDL $0x1fe0, AX                     // and	eax, 8160
	JE   LBB0_7                    // <--                                  // je	.LBB0_7
	ADDQ $-0x20, R8                // <--                                  // add	r8, -32
	LONG $0xedefd1c5               // VPXOR X5, X5, X5                     // vpxor	xmm5, xmm5, xmm5
	LONG $0x20c0f641               // TESTL $0x20, R8                      // test	r8b, 32
	JNE  LBB0_9                    // <--                                  // jne	.LBB0_9
	LONG $0x2e6ffec5               // VMOVDQU 0(SI), Y5                    // vmovdqu	ymm5, ymmword ptr [rsi]
	ADDQ $0x20, SI                 // <--                                  // add	rsi, 32
	LEAQ -0x20(AX), CX             // <--                                  // lea	rcx, [rax - 32]
	LONG $0xf0f6d5c5               // VPSADBW Y0, Y5, Y6                   // vpsadbw	ymm6, ymm5, ymm0
	LONG $0xf4fecdc5               // VPADDD Y4, Y6, Y6                    // vpaddd	ymm6, ymm6, ymm4
	LONG $0x0455e2c4; BYTE $0xe9   // VPMADDUBSW Y1, Y5, Y5                // vpmaddubsw	ymm5, ymm5, ymm1
	LONG $0xeaf5d5c5               // VPMADDWD Y2, Y5, Y5                  // vpmaddwd	ymm5, ymm5, ymm2
	LONG $0xdbfed5c5               // VPADDD Y3, Y5, Y3                    // vpaddd	ymm3, ymm5, ymm3
	LONG $0xec6ffdc5               // VMOVDQA Y4, Y5                       // vmovdqa	ymm5, ymm4
	LONG $0xe66ffdc5               // VMOVDQA Y6, Y4                       // vmovdqa	ymm4, ymm6
	CMPQ R8, $0x20                 // <--                                  // cmp	r8, 32
	JAE  LBB0_12                   // <--                                  // jae	.LBB0_12
	JMP  LBB0_14                   // <--                                  // jmp	.LBB0_14

LBB0_9:
	MOVQ AX, CX    // <--                                  // mov	rcx, rax
	CMPQ R8, $0x20 // <--                                  // cmp	r8, 32
	JB   LBB0_14   // <--                                  // jb	.LBB0_14

LBB0_12:
	LONG $0x366ffec5             // VMOVDQU 0(SI), Y6                    // vmovdqu	ymm6, ymmword ptr [rsi]
	LONG $0x7e6ffec5; BYTE $0x20 // VMOVDQU 0x20(SI), Y7                 // vmovdqu	ymm7, ymmword ptr [rsi + 32]
	LONG $0xc0f64dc5             // VPSADBW Y0, Y6, Y8                   // vpsadbw	ymm8, ymm6, ymm0
	LONG $0xc4fe3dc5             // VPADDD Y4, Y8, Y8                    // vpaddd	ymm8, ymm8, ymm4
	LONG $0xecfed5c5             // VPADDD Y4, Y5, Y5                    // vpaddd	ymm5, ymm5, ymm4
	LONG $0x044de2c4; BYTE $0xe1 // VPMADDUBSW Y1, Y6, Y4                // vpmaddubsw	ymm4, ymm6, ymm1
	LONG $0xe2f5ddc5             // VPMADDWD Y2, Y4, Y4                  // vpmaddwd	ymm4, ymm4, ymm2
	LONG $0xdbfeddc5             // VPADDD Y3, Y4, Y3                    // vpaddd	ymm3, ymm4, ymm3
	ADDQ $0x40, SI               // <--                                  // add	rsi, 64
	LONG $0xe0f6c5c5             // VPSADBW Y0, Y7, Y4                   // vpsadbw	ymm4, ymm7, ymm0
	LONG $0xe4febdc5             // VPADDD Y4, Y8, Y4                    // vpaddd	ymm4, ymm8, ymm4
	LONG $0xedfebdc5             // VPADDD Y5, Y8, Y5                    // vpaddd	ymm5, ymm8, ymm5
	LONG $0x0445e2c4; BYTE $0xf1 // VPMADDUBSW Y1, Y7, Y6                // vpmaddubsw	ymm6, ymm7, ymm1
	LONG $0xf2f5cdc5             // VPMADDWD Y2, Y6, Y6                  // vpmaddwd	ymm6, ymm6, ymm2
	LONG $0xdbfecdc5             // VPADDD Y3, Y6, Y3                    // vpaddd	ymm3, ymm6, ymm3
	ADDQ $-0x40, CX              // <--                                  // add	rcx, -64
	JNE  LBB0_12                 // <--                                  // jne	.LBB0_12
	LONG $0xf46ffdc5             // VMOVDQA Y4, Y6                       // vmovdqa	ymm6, ymm4
	JMP  LBB0_14                 // <--                                  // jmp	.LBB0_14

LBB0_1:
	LONG $0x000001b8; BYTE $0x00 // MOVL $0x1, AX                        // mov	eax, 1

LBB0_2:
	MOVL AX, ret+32(FP) // <--
	RET                 // <--                                  // ret

LBB0_15:
	WORD $0x8548; BYTE $0xd2 // TESTQ DX, DX                         // test	rdx, rdx
	JE   LBB0_16             // <--                                  // je	.LBB0_16

LBB0_17:
	CMPQ DX, $0x10               // <--                                  // cmp	rdx, 16
	JB   LBB0_20                 // <--                                  // jb	.LBB0_20
	WORD $0xb60f; BYTE $0x3e     // MOVZX 0(SI), DI                      // movzx	edi, byte ptr [rsi]
	WORD $0xf801                 // ADDL DI, AX                          // add	eax, edi
	WORD $0xc101                 // ADDL AX, CX                          // add	ecx, eax
	LONG $0x017eb60f             // MOVZX 0x1(SI), DI                    // movzx	edi, byte ptr [rsi + 1]
	WORD $0xc701                 // ADDL AX, DI                          // add	edi, eax
	WORD $0xf901                 // ADDL DI, CX                          // add	ecx, edi
	LONG $0x0246b60f             // MOVZX 0x2(SI), AX                    // movzx	eax, byte ptr [rsi + 2]
	WORD $0xf801                 // ADDL DI, AX                          // add	eax, edi
	WORD $0xc101                 // ADDL AX, CX                          // add	ecx, eax
	LONG $0x037eb60f             // MOVZX 0x3(SI), DI                    // movzx	edi, byte ptr [rsi + 3]
	WORD $0xc701                 // ADDL AX, DI                          // add	edi, eax
	WORD $0xf901                 // ADDL DI, CX                          // add	ecx, edi
	LONG $0x0446b60f             // MOVZX 0x4(SI), AX                    // movzx	eax, byte ptr [rsi + 4]
	WORD $0xf801                 // ADDL DI, AX                          // add	eax, edi
	WORD $0xc101                 // ADDL AX, CX                          // add	ecx, eax
	LONG $0x057eb60f             // MOVZX 0x5(SI), DI                    // movzx	edi, byte ptr [rsi + 5]
	WORD $0xc701                 // ADDL AX, DI                          // add	edi, eax
	WORD $0xf901                 // ADDL DI, CX                          // add	ecx, edi
	LONG $0x0646b60f             // MOVZX 0x6(SI), AX                    // movzx	eax, byte ptr [rsi + 6]
	WORD $0xf801                 // ADDL DI, AX                          // add	eax, edi
	WORD $0xc101                 // ADDL AX, CX                          // add	ecx, eax
	LONG $0x077eb60f             // MOVZX 0x7(SI), DI                    // movzx	edi, byte ptr [rsi + 7]
	WORD $0xc701                 // ADDL AX, DI                          // add	edi, eax
	WORD $0xf901                 // ADDL DI, CX                          // add	ecx, edi
	LONG $0x0846b60f             // MOVZX 0x8(SI), AX                    // movzx	eax, byte ptr [rsi + 8]
	WORD $0xf801                 // ADDL DI, AX                          // add	eax, edi
	WORD $0xc101                 // ADDL AX, CX                          // add	ecx, eax
	LONG $0x097eb60f             // MOVZX 0x9(SI), DI                    // movzx	edi, byte ptr [rsi + 9]
	WORD $0xc701                 // ADDL AX, DI                          // add	edi, eax
	WORD $0xf901                 // ADDL DI, CX                          // add	ecx, edi
	LONG $0x0a46b60f             // MOVZX 0xa(SI), AX                    // movzx	eax, byte ptr [rsi + 10]
	WORD $0xf801                 // ADDL DI, AX                          // add	eax, edi
	WORD $0xc101                 // ADDL AX, CX                          // add	ecx, eax
	LONG $0x0b7eb60f             // MOVZX 0xb(SI), DI                    // movzx	edi, byte ptr [rsi + 11]
	WORD $0xc701                 // ADDL AX, DI                          // add	edi, eax
	WORD $0xf901                 // ADDL DI, CX                          // add	ecx, edi
	LONG $0x0c46b60f             // MOVZX 0xc(SI), AX                    // movzx	eax, byte ptr [rsi + 12]
	WORD $0xf801                 // ADDL DI, AX                          // add	eax, edi
	WORD $0xc101                 // ADDL AX, CX                          // add	ecx, eax
	LONG $0x0d7eb60f             // MOVZX 0xd(SI), DI                    // movzx	edi, byte ptr [rsi + 13]
	WORD $0xc701                 // ADDL AX, DI                          // add	edi, eax
	WORD $0xf901                 // ADDL DI, CX                          // add	ecx, edi
	LONG $0x46b60f44; BYTE $0x0e // MOVZX 0xe(SI), R8                    // movzx	r8d, byte ptr [rsi + 14]
	WORD $0x0141; BYTE $0xf8     // ADDL DI, R8                          // add	r8d, edi
	WORD $0x0144; BYTE $0xc1     // ADDL R8, CX                          // add	ecx, r8d
	LONG $0x0f46b60f             // MOVZX 0xf(SI), AX                    // movzx	eax, byte ptr [rsi + 15]
	WORD $0x0144; BYTE $0xc0     // ADDL R8, AX                          // add	eax, r8d
	WORD $0xc101                 // ADDL AX, CX                          // add	ecx, eax
	ADDQ $-0x10, DX              // <--                                  // add	rdx, -16
	JE   LBB0_27                 // <--                                  // je	.LBB0_27
	ADDQ $0x10, SI               // <--                                  // add	rsi, 16

LBB0_20:
	LEAQ -0x1(DX), DI // <--                                  // lea	rdi, [rdx - 1]
	MOVQ DX, R9       // <--                                  // mov	r9, rdx
	ANDQ $0x3, R9     // <--                                  // and	r9, 3
	JE   LBB0_24      // <--                                  // je	.LBB0_24
	XORL R8, R8       // <--                                  // xor	r8d, r8d

LBB0_22:
	LONG $0x14b60f46; BYTE $0x06 // MOVZX 0(SI)(R8*1), R10               // movzx	r10d, byte ptr [rsi + r8]
	WORD $0x0144; BYTE $0xd0     // ADDL R10, AX                         // add	eax, r10d
	WORD $0xc101                 // ADDL AX, CX                          // add	ecx, eax
	INCQ R8                      // <--                                  // inc	r8
	CMPQ R9, R8                  // <--                                  // cmp	r9, r8
	JNE  LBB0_22                 // <--                                  // jne	.LBB0_22
	ADDQ R8, SI                  // <--                                  // add	rsi, r8
	SUBQ R8, DX                  // <--                                  // sub	rdx, r8

LBB0_24:
	CMPQ DI, $0x3 // <--                                  // cmp	rdi, 3
	JB   LBB0_27  // <--                                  // jb	.LBB0_27
	XORL DI, DI   // <--                                  // xor	edi, edi

LBB0_26:
	LONG $0x04b60f44; BYTE $0x3e   // MOVZX 0(SI)(DI*1), R8                // movzx	r8d, byte ptr [rsi + rdi]
	WORD $0x0141; BYTE $0xc0       // ADDL AX, R8                          // add	r8d, eax
	WORD $0x0144; BYTE $0xc1       // ADDL R8, CX                          // add	ecx, r8d
	LONG $0x3e44b60f; BYTE $0x01   // MOVZX 0x1(SI)(DI*1), AX              // movzx	eax, byte ptr [rsi + rdi + 1]
	WORD $0x0144; BYTE $0xc0       // ADDL R8, AX                          // add	eax, r8d
	WORD $0xc101                   // ADDL AX, CX                          // add	ecx, eax
	LONG $0x44b60f44; WORD $0x023e // MOVZX 0x2(SI)(DI*1), R8              // movzx	r8d, byte ptr [rsi + rdi + 2]
	WORD $0x0141; BYTE $0xc0       // ADDL AX, R8                          // add	r8d, eax
	WORD $0x0144; BYTE $0xc1       // ADDL R8, CX                          // add	ecx, r8d
	LONG $0x3e44b60f; BYTE $0x03   // MOVZX 0x3(SI)(DI*1), AX              // movzx	eax, byte ptr [rsi + rdi + 3]
	WORD $0x0144; BYTE $0xc0       // ADDL R8, AX                          // add	eax, r8d
	WORD $0xc101                   // ADDL AX, CX                          // add	ecx, eax
	ADDQ $0x4, DI                  // <--                                  // add	rdi, 4
	CMPQ DX, DI                    // <--                                  // cmp	rdx, rdi
	JNE  LBB0_26                   // <--                                  // jne	.LBB0_26

LBB0_27:
	LONG  $0x000f908d; WORD $0xffff // LEAL -0xfff1(AX), DX                 // lea	edx, [rax - 65521]
	CMPL  AX, $0xfff1               // <--                                  // cmp	eax, 65521
	WORD  $0x420f; BYTE $0xd0       // CMOVB AX, DX                         // cmovb	edx, eax
	WORD  $0xc889                   // MOVL CX, AX                          // mov	eax, ecx
	LONG  $0x078071be; BYTE $0x80   // MOVL $-0x7ff87f8f, SI                // mov	esi, 2147975281
	IMULQ AX, SI                    // <--                                  // imul	rsi, rax
	SHRQ  $0x2f, SI                 // <--                                  // shr	rsi, 47
	LONG  $0xfff1c669; WORD $0x0000 // IMULL $0xfff1, SI, AX                // imul	eax, esi, 65521
	WORD  $0xc129                   // SUBL AX, CX                          // sub	ecx, eax
	WORD  $0xe1c1; BYTE $0x10       // SHLL $0x10, CX                       // shl	ecx, 16
	WORD  $0xd109                   // ORL DX, CX                           // or	ecx, edx
	WORD  $0xc889                   // MOVL CX, AX                          // mov	eax, ecx
	NOP                             // (skipped)                            // mov	rsp, rbp
	NOP                             // (skipped)                            // pop	rbp
	VZEROUPPER                      // <--                                  // vzeroupper
	MOVL  AX, ret+32(FP)            // <--
	RET                             // <--                                  // ret

LBB0_16:
	WORD $0xe1c1; BYTE $0x10 // SHLL $0x10, CX                       // shl	ecx, 16
	WORD $0xc809             // ORL CX, AX                           // or	eax, ecx
	NOP                      // (skipped)                            // mov	rsp, rbp
	NOP                      // (skipped)                            // pop	rbp
	VZEROUPPER               // <--                                  // vzeroupper
	MOVL AX, ret+32(FP)      // <--
	RET                      // <--                                  // ret
