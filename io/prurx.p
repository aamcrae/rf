;
; Program to sample GPIO and capture
; timing between transitions
;
.origin 0
.entrypoint Start

/*
 Parameters:

u32 event    r0  Event to send when buffer is ready
u32 gpio     r1  GPIO to use
u32 count    r2  Count of buffers
u32 size    r3  Buffer size
u32 addr     ... Addresses of buffers
...
*/
#define BUFS (4*4)  /* Start of buffer addresses */
;
; Register use
; r4 - Number of buffers counter
; r5 - Current buffer address
; r6 - timer counter
; r7 - current buffer size
; r8 - zero
; r9 - Address of list of buffers
Start:
    MOV     r8, 0
    LBBO    r0, r8, 0, BUFS ; Load parameters
    WBC     r31, r1       ; Wait until 0 is seen
    MOV     r6, 14          ; r6 is counter
StartLoop:
; GPIO is 0
    MOV     r9, BUFS        ; Load start of buffer addresses
    MOV     r4, r2          ; Count of buffer addresses
NextBuf:
    LBBO    r5, r9, 0, 4    ; Load next buffer address
    ADD     r9, r9, 4       ; Increment pointer to addresses
    MOV     r7, r3          ; Reload buffer size
RLoop:
ZeroBit:
    ADD     r6, r6, 2       ; Add counter
    QBBC    ZeroBit, r31, r1 ; Loop while bit is 0
    SBBO    r6, r5, 0, 4    ; Store counter
; GPIO is 1
    MOV     r6, 8           ; r6 is counter
OneBit:
    ADD     r6, r6, 2       ; Add counter
    QBBS    OneBit, r31, r1 ; Loop while bit is 0
;
    SBBO    r6, r5, 4, 4    ; Store counter
    ADD     r5, r5, 8       ; Increment address
    SUB     r7, r7, 2       ; Decrement buffer size
    MOV     r6, 7          ; r6 is counter
    QBNE    RLoop, r7, 0
; Buffer is finished, signal event.
    SBBO    r5, r9, 0, 4    ; Store counter
    ADD     r9, r9, 4       ; Increment pointer to addresses
    OR      r31.b0, r0, 0x20
    MOV     r6, 17          ; r6 is counter
    SUB     r4, r4, 1
    QBNE    NextBuf, r4, 0
    QBA     StartLoop
