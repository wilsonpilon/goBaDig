# examples/

Programas de exemplo em Basic Dignified, um conjunto por plataforma
(MSX, ZX81, ZX Spectrum), para o usuário final testar o conversor assim
que cada módulo existir. Cada exemplo é descrito aqui e também em
[`docs/MANUAL.md`](../docs/MANUAL.md#6-converter-um-programa-dignified).

**Estado atual**: pastas criadas, ainda sem exemplos Go-nativos — o
conversor ainda não existe (ver `OUTLINE.md`). Os exemplos reais serão
adaptados a partir dos exemplos do projeto Python original
(`badig/examples/<plataforma>/`), que já cobrem os casos abaixo e
continuam a referência até estes diretórios serem preenchidos.

## [`msx/`](msx/README.md)

Baseado em `badig/examples/msx/` (fonte `CGK173`, um jogo completo —
bom teste de carga real, mas extenso; pode valer criar um "hello world"
MSX mais curto além dele).

## [`zx81/`](zx81/README.md)

Baseado em `badig/examples/zx81/` (`zx81-intro`, introdução básica;
`mcode-ok`, demonstra `remx`/machine code embutido).

## [`zxspectrum/`](zxspectrum/README.md)

Baseado em `badig/examples/zxspectrum/` (`hello`, `colours`, `data`,
`functions`, `graphics`, `native`, `play128`, `streams`) — o conjunto
mais completo hoje, já com validação manual parcial documentada em
`badig/OUTLINE.md`.

## Convenção

Cada exemplo deve vir com:

- o fonte Dignified (`.Dxx`);
- um comentário/cabeçalho curto explicando o que ele demonstra;
- quando aplicável, nota de quais comandos/recursos do dialeto ele
  exercita, para servir também como checklist informal de cobertura.
