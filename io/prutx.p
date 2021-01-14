.origin 0
.entrypoint Start

/*
    Parameters:

u32 event    r0  Event to send when complete
u32 gpio     r1  GPIO to use
u32 repeat   r2  repeat count
u32 length   r3  Length of data in words
u32 data     r4  Address of data

*/
#define DATA (5*4)  /* Start of data */

Start:
    MOV r8, 0
    LBBO r0, r8, 0, DATA ; Load parameters
RepeatLoop:
    MOV r6, r3        ; r6 is length
    MOV r5, 0         ; r5 is bit (0/1)
    MOV r7, r4        ; r7 is data address
SendLoop:
    LBBO r8, r7, 0, 4 ; Load next pulse time to r5
    ADD r7, r7, 4     ; Increment address
;
; Test current bit, and set GPIO to 0 or 1
;
    QBBC SetOff, r5.b0, 0
    SET r30, r30, r1  ; set GPIO output
    QBA Delay
SetOff:
    CLR r30, r30, r1  ; clear GPIO output
;
; Delay loop, 2 instructions.
Delay:
    SUB r8, r8, 2
    QBNE Delay, r8, 0 ; Delay loop
;
; Flip bit
;
    XOR r5.b0, r5.b0, 1
; Check length
    SUB r6, r6, 1
    QBNE SendLoop, r6, 0
;
; Repeat message
;
    SUB r2, r2, 1
    QBNE RepeatLoop, r2, 0
;
; Message repeat complete.
; Ensure GPIO is off
;
    CLR r30, r30, r1 ; clear GPIO output
    OR r31.b0, r0, 0x20
    HALT
