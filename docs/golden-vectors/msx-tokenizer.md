# Vetores de ouro — tokenizer MSX (Batoken)

Referência (ground truth) para testar o futuro tokenizer MSX em Go
(Bloco 2), capturada rodando o tokenizer Python original
(`badig/msx/msxbatoken/msxbatoken.py`, via `badig.py -id msx -asc
--tk_tokenize -rh`) em 2026-11-21. Ver `OUTLINE.md`, Bloco 0, para o
contexto completo da investigação.

**Por que este arquivo existe aqui (e não só em `badig/`)**: todo o
diretório `badig/` é propositalmente ignorado pelo Git (`/badig/` no
`.gitignore` — decisão do usuário em 2026-11-05, histórico congelado,
nunca comitado). O arquivo de teste Python que gerou estes vetores
(`badig/msx/test_tokenizer_msx.py`) existe só localmente nesta máquina
e **não é rastreado** — se `badig/` for apagado/re-clonado do
[projeto original](https://github.com/farique1/basic-dignified), esse
arquivo de teste some junto. Este documento é a cópia **versionada**
dos mesmos vetores, para não depender de `badig/` continuar existindo
neste disco.

`-asc` pula o lexer/parser Dignified por completo — cada vetor testa
só o tokenizer Batoken (texto ASCII já numerado → binário), não a
conversão Dignified→Classic.

| # | Entrada (ASCII numerado) | Saída (hex) | O que isso prova |
|---|---|---|---|
| 1 | `10 PRINT 1` | `ff 09 80 0a 00 91 20 12 00 00 00` | Caso base: ponteiro de próxima linha, número de linha, token, terminador. |
| 2 | `10 ? 1` | (idêntico ao #1) | `?` e `PRINT` tokenizam para o mesmo byte `0x91` — são o mesmo token na ROM, não dois aliases distintos. |
| 3 | `10 PRINT 0` | `... 91 20 11 ...` | Inteiro `0` tem token dedicado de 1 byte: `0x11`. |
| 4 | `10 PRINT 9` | `... 91 20 1a ...` | Inteiro `9` tem token dedicado de 1 byte: `0x1a` (= `0x11 + 9`). Faixa 0-9 inteira usa esse esquema. |
| 5 | `10 PRINT 255` | `... 91 20 0f ff ...` | Inteiro 10-255 usa prefixo de tipo `0x0f` + 1 byte de valor — esquema diferente do caso 0-9. |
| 6 | `10 PRINT 32767` | `... 91 20 1c ff 7f ...` | Inteiro 256-32767 usa prefixo `0x1c` + 2 bytes little-endian — terceiro esquema distinto. |
| 7 | `10 PRINT -1` | `... 91 20 f2 12 ...` | Menos unário é só o operador `-` (`0xf2`) seguido da constante positiva `1` (`0x12`) — **não existe** token de "literal negativo" dedicado. |
| 8 | `10 PRINT 3.14` | `... 91 20 1d 41 31 40 00 ...` | Float precisão simples: prefixo `0x1d` + 4 bytes **Microsoft Binary Format** (não IEEE-754). |
| 9 | `10 PRINT 3.14#` | `... 91 20 1f 41 31 40 00 00 00 00 00 ...` | Sufixo `#` força precisão dupla: prefixo `0x1f` + 8 bytes MBF (os 4 bytes extras ficam zerados aqui porque `3.14` não tem mais dígitos significativos para preencher — não é bug do vetor). |
| 10 | `10 PRINT "HI"` | `... 91 20 22 48 49 22 ...` | String literal é só ASCII puro entre dois bytes `0x22` (`"`) — sem escape, sem prefixo de tamanho. |
| 11 | `10 REM hello` | `... 8f 20 68 65 6c 6c 6f ...` | `REM` por extenso é 1 token (`0x8f`). |
| 12 | `10 ' hello` | `... 3a 8f e6 20 68 65 6c 6c 6f ...` | O atalho `'` **não** é o mesmo token de `REM` — expande para `:` (`0x3a`) + `REM` (`0x8f`) + marcador `0xe6`. Diferente do caso 11. |
| 13 | `10 PRINT 1:PRINT 2` | `... 91 20 12 3a 91 20 13 ...` | `:` (`0x3a`) separa múltiplos statements na mesma linha, byte a byte literal. |
| 14 | `10 PRINT PEEK(0)` | `... 91 20 ff 97 28 11 29 ...` | Função além da tabela base de ~150 tokens usa prefixo `0xff` + código próprio (aqui, `PEEK` = `ff 97`) — 2 bytes em vez de 1. |
| 15 | `10 IF A=1 THEN 20` + `20 PRINT 1` (2 linhas) | linha 1: `... da 20 0e 14 00`; linha 2 começa em `00 19 80 14 00 91 20 12 00` | O alvo de `THEN`/`GOTO` é um valor **bruto de 2 bytes**, sem nenhum prefixo de tipo — diferente de uma constante numérica normal (casos 3-6). Testado com um alvo real (linha 20 existe) para não fixar um artefato de referência pendurada. |

Arquivo-fonte dos testes (não versionado, só local):
`badig/msx/test_tokenizer_msx.py` — rodar com
`python -m unittest discover -s msx -p "test_*.py"` a partir de
`badig/` para reproduzir/revalidar contra o Python original.
