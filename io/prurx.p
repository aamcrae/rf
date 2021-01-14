;
; Program to sample GPIO and capture
; timing between transitions
;
.origin 0
.entrypoint Start

/*
 Parameters:

u32 event    r0  Event to send when complete
u32 gpio     r1  GPIO to use
u32 data     r2  address to put data
u32 length   r3  Length of data in words
;
u32 count

*/
#define DATA (4*4)  /* Start of data */

Start:
    MOV     r8, 0
    MOV     r9, 0
    LBBO    r0, r8, 0, DATA ; Load parameters
    WBC     r31, r1       ; Wait until 0 is seen
RLoop:
; GPIO is 0
    MOV     r6, 0         ; r6 is counter
ZeroBit:
    ADD     r6, r6, 2     ; Add counter
    QBBC    ZeroBit, r31, r1 ; Loop while bit is 0
    SBBO    r6, r2, 0, 4  ; Store counter
    ADD     r2, r2, 4     ; Increment address
    ADD     r9, r9, 1
    SBBO    r9, r8, DATA, 4  ; Store current capture count
    SUB     r3, r3, 1
    QBEQ    Finish, r3, 0
; GPIO is 1
    MOV     r6, 0         ; r6 is counter
OneBit:
    ADD     r6, r6, 2     ; Add counter
    QBBS    OneBit, r31, r1 ; Loop while bit is 0
;
    SBBO    r6, r2, 0, 4  ; Store counter
    ADD     r2, r2, 4     ; Increment address
    ADD     r9, r9, 1
    SBBO    r9, r8, DATA, 4  ; Store current capture count
    SUB     r3, r3, 1
    QBNE    RLoop, r3, 0
Finish:
    OR      r31.b0, r0, 0x20
    HALT
