# Vetores de ouro — tokenizer ZX81

Referência (ground truth) para o tokenizer ZX81 em Go
(`internal/platforms/zx81`, Bloco 3, Etapa 1), capturada lendo
diretamente `badig/zx81/test_tokenizer_zx81.py` (classe
`TokenizerTests`, que chama `tokenize_lines` sem passar por nenhum
subprocesso) e confirmada rodando esses mesmos testes de verdade nesta
máquina (`python -m unittest zx81.test_tokenizer_zx81.TokenizerTests`,
27 testes, todos `ok`) em 2026-11-24.

**Por que este arquivo existe aqui (e não só em `badig/`)**: mesmo
motivo do `msx-tokenizer.md` — `badig/` é propositalmente ignorado
pelo Git (histórico congelado do projeto original, nunca comitado).
Esta é a cópia **versionada** dos vetores, pra não depender de
`badig/` continuar existindo neste disco.

Diferente do tokenizer MSX, o tokenizer ZX81 (`tokenize_lines`)
**já devolve a imagem `.p` completa** (system vars + programa +
display file + variáveis), não só o bytecode do programa — por isso
os vetores abaixo usam o padrão **"corpo da linha N"** (`lineBody` em
`tokenizer_test.go`, mirrors o helper `line_body` do próprio teste
Python) em vez de offset absoluto no buffer, exceto onde o próprio
teste original usa offset fixo de propósito (casos 1 e 9).

## Ponto flutuante oculto (`test_hidden_numbers_round_and_respect_the_exponent_range`)

Formato: `[0x7E][expoente][mantissa de 4 bytes, big-endian]` — expoente
com bias **128** (Sinclair, não IEEE754 nem o Microsoft Binary Format
do MSX), mantissa normalizada (bit implícito 1 removido), com correção
de overflow quando o arredondamento produz `0x80000000` exato.

| Texto | 5 bytes ocultos (hex) |
|---|---|
| `0` | `00 00 00 00 00` |
| `1` | `81 00 00 00 00` |
| `0.5` | `80 00 00 00 00` |
| `10` | `84 20 00 00 00` |
| `16514` | `8F 01 04 00 00` |
| `1.9999999999` | `82 00 00 00 00` (arredonda pra cima — exercita a correção de overflow da mantissa) |
| `3.14159265` | `82 49 0F DA 9E` |
| `1.8E38` | erro: expoente fora da faixa |
| `1E-39` | erro: expoente fora da faixa |

## Escapes hex / casos literais (`test_literal_hex_escapes_have_exact_line_bytes`)

Corpo completo da linha (tokenizado, terminador `0x76` incluído):

| # | Entrada | Corpo (hex) | O que prova |
|---|---|---|---|
| 1 | `600 REM \{D8}\{D8}\{D8}END OF GAME\{D8}\{D8}\{D8}` | `EA D8 D8 D8 2A 33 29 00 34 2B 00 2C 26 32 2A D8 D8 D8 76` | Escape hex funciona dentro de REM, misturado com texto normal. |
| 2 | `1 REM \{C9}\{00}` | `EA C9 00 76` | Qualquer byte, incluindo `0x00`, pode ser inserido literal. |
| 3 | `10 PRINT "\{F5}"` | `F5 0B F5 0B 76` | Escape hex funciona dentro de string também. |
| 4 | `20 LET \{35}OKE=6000` | `F1 35 34 30 2A 14 22 1C 1C 1C 7E 8D 3B 80 00 00 76` | Byte cru vira parte de um nome de variável (aqui, caractere gráfico + "OKE"). |
| 5 | `30 LET A=0\{7E}\{81}\{00}\{00}\{00}\{00}` | `F1 26 14 1C 7E 81 00 00 00 00 76` | Override manual do bloco oculto: `\{7E}` logo após um número desativa a geração automática do ponto flutuante — o autor fornece os 5 bytes à mão. |
| 6 | `30 LET A=1e999\{7e}\{81}\{00}\{00}\{00}\{00}` | `F1 26 14 1D 2A 25 25 25 7E 81 00 00 00 00 76` | Mesmo override, usado pra fixar um valor que SEM ele daria erro de expoente fora da faixa (`1e999`). |
| 7 | `10 PRINT "\{0b}1":\{ea}PRINT 2` | `F5 0B 0B 1D 0B 0E EA F5 1E 7E 82 00 00 00 00 76` | Escape hex pode inserir o PRÓPRIO token de uma palavra reservada (`\{EA}`=REM) sem disparar o parsing normal daquela palavra. |
| 8 | `10 PRINT \{35}123` | `F5 35 1D 1E 1F 7E 87 76 00 00 00 76` | O byte cru `\{35}` (que por coincidência decodifica pro caractere ASCII `'5'`) não interfere na detecção do número seguinte: "123" ainda é reconhecido como 1 número completo, com seu próprio bloco oculto (`7E 87 76 00 00 00` = expoente `0x87`, mantissa `0x76000000`) — a checagem de "caractere anterior não é letra" olha o texto-FONTE (`}`, não alfabético), não o byte decodificado. |
| 9 | `10 LET A=0 \{7e}\{81}\{00}\{00}\{00}\{00}` | `F1 26 14 1C 7E 00 00 00 00 00 7E 81 00 00 00 00 76` | Espaço entre o número e o escape `\{7E}` quebra a detecção do override — o `0` recebe seu PRÓPRIO bloco automático (`1C 7E 00 00 00 00`), e o `\{7E}...` que viria depois é tratado como bytes crus independentes. |

## `remx` (`test_remx_generates_machine_code_rem_lines`, `test_remx_in_numbered_and_classic_listings`)

Diretiva `remx <linha>,<count>[,<byte-ou-char>]`, bypassa completamente
o tokenizer normal — gera direto `REM` (`0xEA`) + `count` repetições do
byte + terminador `0x76`.

| Entrada | Corpo da linha gerada (hex) |
|---|---|
| `remx 1,6` (sem byte — default é o caractere `'0'`, código `0x1C`) | `EA 1C 1C 1C 1C 1C 1C 76` |
| `remx 9000,3,"A"` (byte via caractere entre aspas) | `EA 26 26 26 76` |
| `remx 0,4,118` (byte numérico direto) | `EA 76 76 76 76 76` |

Erros de validação (`test_remx_rejects_invalid_arguments`): contagem
deve ser `1..49152`, byte `0..255`, número de linha `0..9999`,
argumentos ausentes = "remx expects...".

## Estrutura da imagem `.p` (`test_program_pointers_follow_program_and_display_file`)

Para `['10 PRINT 1']`: `data_file == 0x407D + 13` (13 = tamanho do
registro da linha 10: 2+2+9), `variables == data_file + 793`, últimos
2 bytes do arquivo = `80 76`.

## Caso base completo (`test_numbered_basic_line_has_expected_zx81_tokens`)

`program[116:129]` (offset absoluto, início da área de programa) para
`['10 PRINT 1']`:

```
00 0a 09 00 f5 1d 7e 81 00 00 00 00 76
```

`00 0a` = linha 10 (big-endian) · `09 00` = tamanho 9 (little-endian)
· `f5` = `PRINT` · `1d` = caractere `'1'` · `7e 81 00 00 00 00` = bloco
oculto do número `1` · `76` = terminador de linha.

## Validação cruzada ad-hoc (não um vetor fixo, confirma o port em geral)

Rodando `tokenize_lines` do Python real e `zx81.Tokenize` do Go com as
mesmas 6 linhas (`LET`, `FOR`/`STEP`, `PRINT AT`-como-texto-clássico,
`IF`/`<>`/`GOTO`, `REM`, `STOP`) produz exatamente o mesmo buffer de
784 bytes, byte a byte — confirmado em 2026-11-24, não fixado como
teste automatizado (cobertura intencional: os vetores acima já cobrem
cada mecanismo individualmente).
