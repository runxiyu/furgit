//go:build !noasm && arm64

#include "textflag.h"

DATA mult_table<>+0x00(SB)/8, $0x001d001e001f0020
DATA mult_table<>+0x08(SB)/8, $0x0019001a001b001c
DATA mult_table<>+0x10(SB)/8, $0x0015001600170018
DATA mult_table<>+0x18(SB)/8, $0x0011001200130014
DATA mult_table<>+0x20(SB)/8, $0x000d000e000f0010
DATA mult_table<>+0x28(SB)/8, $0x0009000a000b000c
DATA mult_table<>+0x30(SB)/8, $0x0005000600070008
DATA mult_table<>+0x38(SB)/8, $0x0001000200030004
GLOBL mult_table<>(SB), (RODATA|NOPTR), $64

TEXT ·adler32_neon(SB), NOSPLIT, $0-36
	MOVW adler+0(FP), R0
	MOVD buf+8(FP), R1
	MOVD buf_len+16(FP), R2
	MOVD buf_cap+24(FP), R3
	NOP                     // (skipped)                            // stp	x29, x30, [sp, #-16]!
	ANDS $15, R1, R10       // <--                                  // ands	x10, x1, #0xf
	ANDW $65535, R0, R8     // <--                                  // and	w8, w0, #0xffff
	LSRW $16, R0, R9        // <--                                  // lsr	w9, w0, #16
	NOP                     // (skipped)                            // mov	x29, sp
	BEQ  LBB0_4             // <--                                  // b.eq	.LBB0_4
	ADD  $1, R1, R11        // <--                                  // add	x11, x1, #1
	MOVD R1, R12            // <--                                  // mov	x12, x1

LBB0_2:
	WORD  $0x3840158d       // MOVBU.P 1(R12), R13                  // ldrb	w13, [x12], #1
	SUB   $1, R2, R2        // <--                                  // sub	x2, x2, #1
	TST   $15, R11          // <--                                  // tst	x11, #0xf
	ADD   $1, R11, R11      // <--                                  // add	x11, x11, #1
	ADDW  R13, R8, R8       // <--                                  // add	w8, w8, w13
	ADDW  R9, R8, R9        // <--                                  // add	w9, w8, w9
	BNE   LBB0_2            // <--                                  // b.ne	.LBB0_2
	MOVW  $32881, R11       // <--                                  // mov	w11, #32881
	MOVW  $65521, R13       // <--                                  // mov	w13, #65521
	MOVKW $(32775<<16), R11 // <--                                  // movk	w11, #32775, lsl #16
	MOVW  $4294901775, R12  // <--                                  // mov	w12, #-65521
	MOVW  $65520, R14       // <--                                  // mov	w14, #65520
	SUB   R10, R1, R10      // <--                                  // sub	x10, x1, x10
	UMULL R11, R9, R11      // <--                                  // umull	x11, w9, w11
	ADDW  R12, R8, R12      // <--                                  // add	w12, w8, w12
	CMPW  R14, R8           // <--                                  // cmp	w8, w14
	ADD   $16, R10, R1      // <--                                  // add	x1, x10, #16
	LSR   $47, R11, R11     // <--                                  // lsr	x11, x11, #47
	CSELW HI, R12, R8, R8   // <--                                  // csel	w8, w12, w8, hi
	MSUBW R13, R9, R11, R9  // <--                                  // msub	w9, w11, w13, w9

LBB0_4:
	AND   $31, R2, R10                        // <--                                  // and	x10, x2, #0x1f
	CMP   $32, R2                             // <--                                  // cmp	x2, #32
	BCC   LBB0_9                              // <--                                  // b.lo	.LBB0_9
	MOVD  $mult_table<>(SB), R11              // <--                                  // adrp	x11, mult_table
	ADD   $0, R11, R11                        // <--                                  // add	x11, x11, :lo12:mult_table
	MOVW  $32881, R14                         // <--                                  // mov	w14, #32881
	MOVW  $173, R12                           // <--                                  // mov	w12, #173
	MOVD  $137438953440, R13                  // <--                                  // mov	x13, #137438953440
	MOVKW $(32775<<16), R14                   // <--                                  // movk	w14, #32775, lsl #16
	VLD1  (R11), [V0.H8, V1.H8, V2.H8, V3.H8] // <--                                  // ld1	{ v0.8h, v1.8h, v2.8h, v3.8h }, [x11]
	LSR   $5, R2, R11                         // <--                                  // lsr	x11, x2, #5
	MOVW  $65521, R15                         // <--                                  // mov	w15, #65521
	VEXT  $8, V0.B16, V0.B16, V4.B16          // <--                                  // ext	v4.16b, v0.16b, v0.16b, #8
	VEXT  $8, V1.B16, V1.B16, V5.B16          // <--                                  // ext	v5.16b, v1.16b, v1.16b, #8
	VEXT  $8, V2.B16, V2.B16, V6.B16          // <--                                  // ext	v6.16b, v2.16b, v2.16b, #8
	VEXT  $8, V3.B16, V3.B16, V7.B16          // <--                                  // ext	v7.16b, v3.16b, v3.16b, #8

LBB0_6:
	CMP  $173, R11               // <--                                  // cmp	x11, #173
	MOVD R1, R2                  // <--                                  // mov	x2, x1
	CSEL LO, R11, R12, R16       // <--                                  // csel	x16, x11, x12, lo
	WORD $0x6f00e414             // VMOVI $0, V20.D2                     // movi	v20.2d, #0000000000000000
	MULW R16, R8, R0             // <--                                  // mul	w0, w8, w16
	ADD  R16<<5, R13, R17        // <--                                  // add	x17, x13, x16, lsl #5
	WORD $0x6f00e410             // VMOVI $0, V16.D2                     // movi	v16.2d, #0000000000000000
	AND  $137438953440, R17, R17 // <--                                  // and	x17, x17, #0x1fffffffe0
	WORD $0x6f00e412             // VMOVI $0, V18.D2                     // movi	v18.2d, #0000000000000000
	WORD $0x6f00e413             // VMOVI $0, V19.D2                     // movi	v19.2d, #0000000000000000
	WORD $0x6f00e415             // VMOVI $0, V21.D2                     // movi	v21.2d, #0000000000000000
	VMOV R0, V20.S[3]            // <--                                  // mov	v20.s[3], w0
	MOVW R16, R0                 // <--                                  // mov	w0, w16
	WORD $0x6f00e411             // VMOVI $0, V17.D2                     // movi	v17.2d, #0000000000000000

LBB0_7:
	WORD  $0xacc15857                   // FLDPQ.P 32(R2), (F23, F22)           // ldp	q23, q22, [x2], #32
	SUBSW $1, R0, R0                    // <--                                  // subs	w0, w0, #1
	VADD  V17.S4, V20.S4, V20.S4        // <--                                  // add	v20.4s, v20.4s, v17.4s
	WORD  $0x2e3712b5                   // VUADDW V23.B8, V21.H8, V21.H8        // uaddw	v21.8h, v21.8h, v23.8b
	WORD  $0x6e371273                   // VUADDW2 V23.B16, V19.H8, V19.H8      // uaddw2	v19.8h, v19.8h, v23.16b
	WORD  $0x6e202ad8                   // VUADDLP V22.B16, V24.H8              // uaddlp	v24.8h, v22.16b
	WORD  $0x2e361252                   // VUADDW V22.B8, V18.H8, V18.H8        // uaddw	v18.8h, v18.8h, v22.8b
	WORD  $0x6e361210                   // VUADDW2 V22.B16, V16.H8, V16.H8      // uaddw2	v16.8h, v16.8h, v22.16b
	WORD  $0x6e206af8                   // VUADALP V23.B16, V24.H8              // uadalp	v24.8h, v23.16b
	WORD  $0x6e606b11                   // VUADALP V24.H8, V17.S4               // uadalp	v17.4s, v24.8h
	BNE   LBB0_7                        // <--                                  // b.ne	.LBB0_7
	VSHL  $5, V20.S4, V20.S4            // <--                                  // shl	v20.4s, v20.4s, #5
	ADD   R17, R1, R17                  // <--                                  // add	x17, x1, x17
	SUBS  R16, R11, R11                 // <--                                  // subs	x11, x11, x16
	ADD   $32, R17, R1                  // <--                                  // add	x1, x17, #32
	WORD  $0x2e6082b4                   // VUMLAL V0.H4, V21.H4, V20.S4         // umlal	v20.4s, v21.4h, v0.4h
	VEXT  $8, V21.B16, V21.B16, V21.B16 // <--                                  // ext	v21.16b, v21.16b, v21.16b, #8
	WORD  $0x2e6482b4                   // VUMLAL V4.H4, V21.H4, V20.S4         // umlal	v20.4s, v21.4h, v4.4h
	VEXT  $8, V19.B16, V19.B16, V21.B16 // <--                                  // ext	v21.16b, v19.16b, v19.16b, #8
	WORD  $0x2e618274                   // VUMLAL V1.H4, V19.H4, V20.S4         // umlal	v20.4s, v19.4h, v1.4h
	VEXT  $8, V18.B16, V18.B16, V19.B16 // <--                                  // ext	v19.16b, v18.16b, v18.16b, #8
	WORD  $0x2e6582b4                   // VUMLAL V5.H4, V21.H4, V20.S4         // umlal	v20.4s, v21.4h, v5.4h
	WORD  $0x2e628254                   // VUMLAL V2.H4, V18.H4, V20.S4         // umlal	v20.4s, v18.4h, v2.4h
	WORD  $0x2e668274                   // VUMLAL V6.H4, V19.H4, V20.S4         // umlal	v20.4s, v19.4h, v6.4h
	WORD  $0x2e638214                   // VUMLAL V3.H4, V16.H4, V20.S4         // umlal	v20.4s, v16.4h, v3.4h
	VEXT  $8, V16.B16, V16.B16, V16.B16 // <--                                  // ext	v16.16b, v16.16b, v16.16b, #8
	WORD  $0x2e678214                   // VUMLAL V7.H4, V16.H4, V20.S4         // umlal	v20.4s, v16.4h, v7.4h
	WORD  $0x4eb1be30                   // VADDP V17.S4, V17.S4, V16.S4         // addp	v16.4s, v17.4s, v17.4s
	WORD  $0x4eb4be91                   // VADDP V20.S4, V20.S4, V17.S4         // addp	v17.4s, v20.4s, v20.4s
	WORD  $0x0eb1be10                   // VADDP V17.S2, V16.S2, V16.S2         // addp	v16.2s, v16.2s, v17.2s
	VMOV  V16.S[1], R0                  // <--                                  // mov	w0, v16.s[1]
	FMOVS F16, R2                       // <--                                  // fmov	w2, s16
	ADDW  R8, R2, R8                    // <--                                  // add	w8, w2, w8
	ADDW  R9, R0, R9                    // <--                                  // add	w9, w0, w9
	UMULL R14, R8, R0                   // <--                                  // umull	x0, w8, w14
	UMULL R14, R9, R2                   // <--                                  // umull	x2, w9, w14
	LSR   $47, R0, R0                   // <--                                  // lsr	x0, x0, #47
	LSR   $47, R2, R2                   // <--                                  // lsr	x2, x2, #47
	MSUBW R15, R8, R0, R8               // <--                                  // msub	w8, w0, w15, w8
	MSUBW R15, R9, R2, R9               // <--                                  // msub	w9, w2, w15, w9
	BNE   LBB0_6                        // <--                                  // b.ne	.LBB0_6

LBB0_9:
	CBZ  R10, LBB0_15  // <--                                  // cbz	x10, .LBB0_15
	CMP  $16, R10      // <--                                  // cmp	x10, #16
	BCC  LBB0_13       // <--                                  // b.lo	.LBB0_13
	WORD $0x3940002b   // MOVBU (R1), R11                      // ldrb	w11, [x1]
	SUBS $16, R10, R10 // <--                                  // subs	x10, x10, #16
	WORD $0x3940042c   // MOVBU 1(R1), R12                     // ldrb	w12, [x1, #1]
	WORD $0x3940082d   // MOVBU 2(R1), R13                     // ldrb	w13, [x1, #2]
	ADDW R11, R8, R8   // <--                                  // add	w8, w8, w11
	WORD $0x39400c2b   // MOVBU 3(R1), R11                     // ldrb	w11, [x1, #3]
	ADDW R9, R8, R9    // <--                                  // add	w9, w8, w9
	ADDW R12, R8, R8   // <--                                  // add	w8, w8, w12
	WORD $0x3940102c   // MOVBU 4(R1), R12                     // ldrb	w12, [x1, #4]
	ADDW R8, R9, R9    // <--                                  // add	w9, w9, w8
	ADDW R13, R8, R8   // <--                                  // add	w8, w8, w13
	WORD $0x3940142d   // MOVBU 5(R1), R13                     // ldrb	w13, [x1, #5]
	ADDW R8, R9, R9    // <--                                  // add	w9, w9, w8
	ADDW R11, R8, R8   // <--                                  // add	w8, w8, w11
	WORD $0x3940182b   // MOVBU 6(R1), R11                     // ldrb	w11, [x1, #6]
	ADDW R8, R9, R9    // <--                                  // add	w9, w9, w8
	ADDW R12, R8, R8   // <--                                  // add	w8, w8, w12
	WORD $0x39401c2c   // MOVBU 7(R1), R12                     // ldrb	w12, [x1, #7]
	ADDW R8, R9, R9    // <--                                  // add	w9, w9, w8
	ADDW R13, R8, R8   // <--                                  // add	w8, w8, w13
	ADDW R8, R9, R9    // <--                                  // add	w9, w9, w8
	ADDW R11, R8, R8   // <--                                  // add	w8, w8, w11
	WORD $0x3940202b   // MOVBU 8(R1), R11                     // ldrb	w11, [x1, #8]
	ADDW R8, R9, R9    // <--                                  // add	w9, w9, w8
	ADDW R12, R8, R8   // <--                                  // add	w8, w8, w12
	WORD $0x3940242c   // MOVBU 9(R1), R12                     // ldrb	w12, [x1, #9]
	ADDW R8, R9, R9    // <--                                  // add	w9, w9, w8
	WORD $0x3940382d   // MOVBU 14(R1), R13                    // ldrb	w13, [x1, #14]
	ADDW R11, R8, R8   // <--                                  // add	w8, w8, w11
	WORD $0x3940282b   // MOVBU 10(R1), R11                    // ldrb	w11, [x1, #10]
	ADDW R8, R9, R9    // <--                                  // add	w9, w9, w8
	ADDW R12, R8, R8   // <--                                  // add	w8, w8, w12
	WORD $0x39402c2c   // MOVBU 11(R1), R12                    // ldrb	w12, [x1, #11]
	ADDW R8, R9, R9    // <--                                  // add	w9, w9, w8
	ADDW R11, R8, R8   // <--                                  // add	w8, w8, w11
	WORD $0x3940302b   // MOVBU 12(R1), R11                    // ldrb	w11, [x1, #12]
	ADDW R8, R9, R9    // <--                                  // add	w9, w9, w8
	ADDW R12, R8, R8   // <--                                  // add	w8, w8, w12
	WORD $0x3940342c   // MOVBU 13(R1), R12                    // ldrb	w12, [x1, #13]
	ADDW R8, R9, R9    // <--                                  // add	w9, w9, w8
	ADDW R11, R8, R8   // <--                                  // add	w8, w8, w11
	WORD $0x39403c2b   // MOVBU 15(R1), R11                    // ldrb	w11, [x1, #15]
	ADDW R8, R9, R9    // <--                                  // add	w9, w9, w8
	ADDW R12, R8, R8   // <--                                  // add	w8, w8, w12
	ADDW R8, R9, R9    // <--                                  // add	w9, w9, w8
	ADDW R13, R8, R8   // <--                                  // add	w8, w8, w13
	ADDW R8, R9, R9    // <--                                  // add	w9, w9, w8
	ADDW R11, R8, R8   // <--                                  // add	w8, w8, w11
	ADDW R8, R9, R9    // <--                                  // add	w9, w9, w8
	BEQ  LBB0_14       // <--                                  // b.eq	.LBB0_14
	ADD  $16, R1, R1   // <--                                  // add	x1, x1, #16

LBB0_13:
	WORD $0x3840142b  // MOVBU.P 1(R1), R11                   // ldrb	w11, [x1], #1
	SUBS $1, R10, R10 // <--                                  // subs	x10, x10, #1
	ADDW R11, R8, R8  // <--                                  // add	w8, w8, w11
	ADDW R9, R8, R9   // <--                                  // add	w9, w8, w9
	BNE  LBB0_13      // <--                                  // b.ne	.LBB0_13

LBB0_14:
	MOVW  $32881, R10       // <--                                  // mov	w10, #32881
	MOVW  $65521, R12       // <--                                  // mov	w12, #65521
	MOVKW $(32775<<16), R10 // <--                                  // movk	w10, #32775, lsl #16
	MOVW  $4294901775, R11  // <--                                  // mov	w11, #-65521
	MOVW  $65520, R13       // <--                                  // mov	w13, #65520
	ADDW  R11, R8, R11      // <--                                  // add	w11, w8, w11
	UMULL R10, R9, R10      // <--                                  // umull	x10, w9, w10
	CMPW  R13, R8           // <--                                  // cmp	w8, w13
	CSELW HI, R11, R8, R8   // <--                                  // csel	w8, w11, w8, hi
	LSR   $47, R10, R10     // <--                                  // lsr	x10, x10, #47
	MSUBW R12, R9, R10, R9  // <--                                  // msub	w9, w10, w12, w9

LBB0_15:
	ORRW R9<<16, R8, R0 // <--                                  // orr	w0, w8, w9, lsl #16
	NOP                 // (skipped)                            // ldp	x29, x30, [sp], #16
	MOVW R0, ret+32(FP) // <--
	RET                 // <--                                  // ret
