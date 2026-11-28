package edit

// z80Mnemonics are every official Z80 opcode mnemonic plus the
// commonly-used undocumented ones (the IX/IY half-register shift/
// rotate/bit forms - SLL, and the undocumented IN F,(C) spelled OUT
// (C),0/IN (C) here are omitted since they aren't separate mnemonics)
// - distilled from asMSX's own manual (resource/asMSX/doc/asmsx.md,
// "supports all official Z80 instructions" / "all known unofficial
// instructions"), not from any Sublime/VSCode grammar (none exists in
// this repo for Z80 - see OUTLINE.md's Fase 6 entry).
var z80Mnemonics = wordSet(
	"ADC", "ADD", "AND", "BIT", "CALL", "CCF", "CP", "CPD", "CPDR", "CPI",
	"CPIR", "CPL", "DAA", "DEC", "DI", "DJNZ", "EI", "EX", "EXX", "HALT",
	"IM", "IN", "INC", "IND", "INDR", "INI", "INIR", "JP", "JR", "LD",
	"LDD", "LDDR", "LDI", "LDIR", "NEG", "NOP", "OR", "OTDR", "OTIR",
	"OUT", "OUTD", "OUTI", "POP", "PUSH", "RES", "RET", "RETI", "RETN",
	"RL", "RLA", "RLC", "RLCA", "RLD", "RR", "RRA", "RRC", "RRCA", "RRD",
	"RST", "SBC", "SCF", "SET", "SLA", "SLL", "SRA", "SRL", "SUB", "XOR",
)

// z80Registers are the Z80's register names, including the IX/IY
// 8-bit halves (an undocumented but universally supported form).
var z80Registers = wordSet(
	"A", "B", "C", "D", "E", "F", "H", "L", "I", "R",
	"IX", "IY", "IXH", "IXL", "IYH", "IYL", "SP", "PC",
	"AF", "BC", "DE", "HL",
)

// asmsxDirectives are asMSX's own assembler directives and data-
// definition pseudo-ops (resource/asMSX/doc/asmsx.md sections 2.4,
// 2.5, 2.7 and 2.8.1), listed WITHOUT their optional leading "."
// (asmsx's own docs: "Remember, you can use them without first point
// character") - tokenizeLine strips a leading dot before this lookup,
// so both ".ORG" and "ORG" resolve here. Conditional-assembly
// (IF/ELSE/ENDIF/IFDEF) and macro (MACRO/ENDM/REPT/ENDR) keywords are
// included even though asMSX itself never requires a dot on those.
var asmsxDirectives = wordSet(
	"ZILOG", "ORG", "PAGE", "EQU", "BIOS", "BIOSVARS", "ROM", "MEGAROM",
	"BASIC", "MSXDOS", "START", "SEARCH", "SUBPAGE", "SELECT", "PHASE",
	"DEPHASE", "CALLBIOS", "CALLDOS", "SIZE", "INCLUDE", "INCBIN",
	"RANDOM", "DEBUG", "BREAK", "BREAKPOINT", "PRINTTEXT", "PRINTSTRING",
	"PRINT", "PRINTDEC", "PRINTHEX", "PRINTFIX", "CAS", "CASSETTE",
	"WAV", "FILENAME", "SINCLAIR", "BYTE", "WORD",
	"DB", "DEFB", "DT", "DEFT", "DW", "DEFW", "DS", "DEFS",
	"IF", "ELSE", "ENDIF", "IFDEF",
	"MACRO", "ENDM", "REPT", "ENDR",
)

// asmsxNumericFunctions are the numeric-expression functions and
// constants asMSX's own expression evaluator supports (asmsx.md
// section 2.3.2) - highlighted as functions, same visual class as a
// BASIC dialect's built-in functions.
var asmsxNumericFunctions = wordSet(
	"SIN", "COS", "TAN", "ACOS", "ASIN", "ATAN", "SQR", "SQRT", "EXP",
	"POW", "LOG", "LN", "ABS", "PI", "FIX", "FIXMUL", "FIXDIV", "INT",
)
