; Simple arena allocator
; System V AMD64 ABI (Linux)
; void arena_init(void *buffer, size_t size);
; void *arena_alloc(size_t size);
; void arena_reset(void);

global arena_init
global arena_alloc
global arena_reset

section .data
    arena_base    dq 0
    arena_current dq 0
    arena_end     dq 0

section .text

; void arena_init(void *buffer, size_t size)
; rdi = buffer
; rsi = size
arena_init:
    mov [arena_base], rdi
    mov [arena_current], rdi
    lea rax, [rdi + rsi]
    mov [arena_end], rax
    ret

; void* arena_alloc(size_t size)
; rdi = size
arena_alloc:
    mov rax, [arena_current]     ; rax = current
    mov rcx, rax                 ; save old pointer (return value)
    add rax, rdi                 ; new_current = current + size

    cmp rax, [arena_end]
    ja  .fail                    ; if beyond end → return NULL

    mov [arena_current], rax     ; commit allocation
    mov rax, rcx                 ; return old pointer
    ret

.fail:
    xor rax, rax                 ; return NULL
    ret

; void arena_reset(void)
arena_reset:
    mov rax, [arena_base]
    mov [arena_current], rax
    ret

; how to compilate:
; do an arena.h
; declare in C with a defined buffer

; compile example
; nasm -f elf64 arena.asm -o arena.o
; gcc main.c arena.o -o test
; ./test