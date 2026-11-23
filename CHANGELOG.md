# Changelog

Todas as versões publicadas do goBaDig, da primeira à mais recente
(mais recente no topo). Cada entrada aqui tem também um documento de
release equivalente em [`docs/releases/`](docs/releases/), com o mesmo
conteúdo publicado na página de Releases do GitHub.

## Esquema de versionamento (`X.Y.Z`)

Convenção definida pelo autor do projeto, diferente de SemVer tradicional:

- **`Z`** — incrementado a **cada compilação/build** gerada.
- **`Y`** — incrementado a **cada feature nova** (incluída, solicitada
  ou testada); carrega um **nome resumido** do que foi feito (ex.:
  `2.3.0-fn-nativo`, `2.4.0-tap-tzx`). `Z` reinicia em `0` quando `Y` sobe.
- **`X`** — um **bloco grande** de mudanças: um conjunto de features
  concluídas/testadas que forma um marco maior. Recebe o **nome de um
  comando de BASIC** (de qualquer um dos três dialetos — MSX, ZX81 ou
  ZX Spectrum), ex.: `1.0.0-PRINT`, `2.0.0-GOTO`. `Y` e `Z` reiniciam em
  `0` quando `X` sobe.

> Interpretação de reset de `Y`/`Z` e formato do nome anexado à tag
> confirmados pelo usuário em 2026-11-20 (ver `OUTLINE.md`, seção 8,
> perguntas 8 e 9, e seção 10.1): `Z` reinicia em `0` quando `Y` sobe;
> `Y`/`Z` reiniciam em `0` quando `X` sobe; o nome resumido (`Y`) e o
> nome de comando BASIC (`X`) aparecem só no `CHANGELOG.md`/release
> notes — a tag Git fica só numérica (`vX.Y.Z`).

## Empacotamento de release

Cada release publicado gera um único pacote `goBaDig-vX.Y.Z.zip`
(nome do projeto + versão), contendo o executável, documentação mínima e
exemplos — ver [`dist/README.md`](dist/README.md) para a estrutura
exata do pacote.

---

## [1.22.0] — 2026-11-23 — bloco **NEW**, feature **set-command** (build 0)

**Comandos `SET`/`SETTINGS`/`DIALECT` no console CLI** — Etapa B do
painel de configuração (Etapa A, o catálogo + SQLite, foi a
v1.21.0).

### Adicionado

- `set <opção>` mostra o valor atual; `set <opção> <valor>` valida
  (contra o catálogo de `internal/options`) e grava.
- `settings` lista Basic Dignified + o dialeto atual da sessão.
- `dialect <nome>` troca qual dialeto (`msx`/`zx81`/`zxspectrum`) o
  `set` usa pra resolver uma chave que não é compartilhada.
- Todo comando do console agora casa case-insensitive (`SET
  INITIAL_LINE 1` funciona, igual ao exemplo original pedido).
- Tab-completion pra `dialect <nome>`.
- 13 testes novos.

### O que não mudou / ainda falta

Nenhum painel de configuração na TUI/GUI ainda (próximas etapas); o
dialeto da sessão não é persistido (só em memória, reseta a cada
console novo); nenhum valor configurado aqui é lido pelo pipeline de
conversão real ainda.

## [1.21.0] — 2026-11-23 — bloco **NEW**, feature **options-catalog** (build 0)

**Catálogo de opções de configuração + schema SQLite (Etapa A do
painel de configuração)** — primeira etapa de um recurso maior:
expor as ~40 opções que o `badig.py` original configura por CLI/
`.ini`/remtag (numeração de linha, cabeçalho REM, TAB, espaçamento,
tokenizer, emulador — por dialeto) como configuração persistida,
editável depois via comandos `SET` no CLI e painéis na TUI/GUI
(etapas futuras, ainda não implementadas).

### Adicionado

- Pacote novo `internal/options`: catálogo embutido com as 40
  opções, em 4 categorias (Basic Dignified geral, MSX, ZX81, ZX
  Spectrum), cada uma com tipo (`bool`/`int`/`string`/`enum`),
  default e validação.
- `internal/config.Store` ganha `SetOption`/`GetOption`/
  `ListOptions`/`EnsureOptionDefaults`, nova tabela SQLite
  `options` — mesmo padrão já usado pelo sistema de temas
  (valida antes de gravar, só semeia o default na primeira vez, nunca
  sobrescreve um valor já alterado pelo usuário).
- 18 testes novos.

### O que não mudou / ainda falta

Nenhum comando `SET` no CLI, nenhum painel na TUI/GUI ainda (próximas
etapas) — e nenhum valor configurado aqui é consumido pelo pipeline
de conversão real ainda (isso é um passo futuro separado).

## [1.20.0] — 2026-11-23 — bloco **NEW**, feature **tokenizer-wiring** (build 0)

**Ligação ao tokenizer MSX Batoken** — a primeira conversão
Dignified→binário `.bmx` de ponta a ponta de verdade, fechando o
caminho completo que faltava desde a v1.4.0.

### Adicionado

- `msx.CompileToBinary(ctx, platform) ([]byte, error)`: roda o
  pipeline textual inteiro (`core.Run`) e tokeniza o resultado
  (`Tokenize`) num só passo — `GenerateResult.Code` já tinha
  exatamente o formato que `Tokenize` espera, então a ligação não
  precisou de nenhuma mudança nos dois lados, só a função que os
  encadeia.
- 2 testes novos (a ligação em si, confirmada contra uma chamada
  equivalente feita na mão; propagação de erro de Parse).

### Validação

`CGK173.dmx` inteiro (891 linhas) compilado direto pra `.bmx` binário
bate **byte a byte** (18.815 bytes) contra o `.bmx` real gerado por
`python badig.py CGK173.dmx CGK173.amx --tk_tokenize` nesta máquina —
a primeira validação de ponta a ponta até o binário.

## [1.19.0] — 2026-11-23 — bloco **NEW**, feature **callshorthand** (build 0)

**Pass 1 MSX: separador de chamada Dignified `_` fora do fim de
linha** — última peça do caminho textual Dignified→Classic ainda não
portada.

### Adicionado

- Um `_` não imediatamente seguido de `NEWLINE` é agora reconhecido
  como o atalho de `CALL` do MSX (`_CLS` == `CALL CLS`): o
  identificador logo depois dele fica excluído da renomeação de
  variável pra alias de 2 letras — mesmo que o MESMO nome, usado de
  novo em outro lugar sem `_` na frente, continue sendo renomeado
  normalmente.
- `_` no fim de linha continua se comportando exatamente como antes
  (fusão de linha Dignified, implementada em release anterior).
- 2 testes novos de ponta a ponta no MSX.

### Validação

`CGK173.dmx` não usa esse mecanismo em nenhum lugar de verdade (todo
`_` nele está no fim de linha) — a validação 100% de ponta a ponta,
byte a byte, contra a saída real do `badig.py` continua de pé, sem
nenhuma diferença.

## [1.18.0] — 2026-11-23 — bloco **NEW**, feature **remheader** (build 0)

**Cabeçalho REM de 2 linhas + expansão de TAB** — fecha os dois
últimos gaps conhecidos contra a saída real do `badig.py`: o
`CGK173.dmx` inteiro (891 linhas) agora bate **byte a byte**, sem
nenhuma diferença restante.

### Adicionado

- Cabeçalho REM automático de 2 linhas (`"Converted with Basic
  Dignified"` / `"https://github.com/farique1/basic-dignified"`),
  controlado por `CompileSettings.RemHeader` (default `false` — sem
  wiring de CLI ainda pra expor o default real do Python, `true`).
- Expansão de TAB pra espaços em `core.LoadLines`, controlada por
  `CompileSettings.TabLength`, mirror exato de `str.expandtabs`.
- Corrigida uma suposição documentada na release anterior: o offset
  de +20 encontrado ali não era "o Python reserva 2 números de linha
  mesmo com `-rh` desligado" — era simplesmente o cabeçalho real que
  este port ainda não gerava. `-rh` desliga o cabeçalho por completo,
  sem reserva nenhuma.
- 4 testes novos no mecanismo core-level.

## [1.17.0] — 2026-11-22 — bloco **NEW**, feature **colonfusion** (build 0)

**Fusão de linha pelo separador clássico `:`** (+ junção de strings
adjacentes) — validada, pela primeira vez, 100% de ponta a ponta
contra o `CGK173.dmx` inteiro (891 linhas) e a saída real do
`badig.py` rodando num Python de verdade.

### Adicionado

- Um `:` clássico no fim de linha funde com a próxima linha,
  **mantendo o próprio `:`** (diferente do `_` Dignified, que
  desaparece inteiro): `for f=0 to 5:` + `print f:` vira uma linha
  só, `for f=0 to 5:print f:...`.
- Um `:` no início de uma linha funde com a linha anterior.
- Dois `:` adjacentes colapsam em um só.
- Duas strings literais adjacentes (mesma linha, ou separadas por uma
  quebra de linha) se concatenam em uma só.
- Corrigido: o corpo de uma função (`FUNC`/`RET`) não desperdiça mais
  uma linha em branco.
- Corrigido: uma linha em branco no início do arquivo não desperdiça
  mais um número de linha.
- Corrigido: `"?"` (o alias de `PRINT`) agora recebe espaçamento
  correto, como qualquer outra palavra reservada.
- 5 testes novos no mecanismo core-level.

### Validação

- O conteúdo gerado para o `CGK173.dmx` inteiro agora bate **byte a
  byte** contra a saída real do `badig.py`, uma vez descontadas as
  duas únicas diferenças já conhecidas e adiadas (inserção do
  cabeçalho REM de 2 linhas, expansão de TAB para espaços).

### Não mudou / ainda falta

- Inserção do cabeçalho REM automático de 2 linhas.
- Expansão de TAB para espaços.
- Pass 1 do separador `_` fora do fim de linha.

---

## [1.16.0] — 2026-11-22 — bloco **NEW**, feature **comments** (build 0)

**Comentários exclusivos do Dignified** (`##`, um bloco de 3
cerquilhas, um bloco de duas aspas simples) — primeira rodada
validada de ponta a ponta contra o `CGK173.dmx` inteiro (891 linhas)
e contra a saída real do `badig.py` rodando em Python de verdade.

### Adicionado

- `##` — comentário de 1 linha, funciona em qualquer posição,
  desaparece sem deixar rastro.
- Um bloco de 3 cerquilhas sozinho no início de uma linha — todo o
  conteúdo entre abertura e fechamento desaparece, mesmo com texto
  colado direto nos marcadores.
- Um bloco de duas aspas simples — mesmo mecanismo, mas o conteúdo
  sobrevive: cada trecho entre abertura e fechamento vira sua própria
  linha de comentário classic numerada.
- Novo `Lexer.Lex` ganha um parâmetro `atLineStart`, usado pra
  desambiguar esses marcadores dos símbolos próprios do MSX (`#`/`'`).
- 2 bugs reais de lexer corrigidos, achados testando o `CGK173.dmx`
  de ponta a ponta: `[?](x,y)` (macro embutida que o Python injeta
  automaticamente, sem o usuário nunca declarar) e um identificador
  que começa com uma palavra reservada seguida de `_` ou dígito
  (`copy_VRAM`, `color1`) sendo cortado errado.
- 7 testes novos no mecanismo core-level + 1 de ponta a ponta no MSX.

### Não mudou / ainda falta

- Um `:` clássico no fim de linha também deveria fundir com a
  próxima (confirmado comparando contra a saída real do Python) —
  ainda não implementado, afeta a numeração de linha de programas
  reais com blocos `for`/`next` escritos em várias linhas.
- Pass 1 do separador `_` fora do fim de linha.

---

## [1.15.0] — 2026-11-22 — bloco **NEW**, feature **generate** (build 0)

**`generate()` de verdade** — espaçamento por palavra reservada +
limite de 256 caracteres por linha, substituindo o "junta tudo com 1
espaço fixo" provisório desde o Bloco 1.

### Adicionado

- Palavra reservada ganha espaço antes/depois de si mesma, a menos
  que o vizinho seja um símbolo composto (`c_symb_comp`) ou já exista
  espaço ali — mesma regra do original, puramente baseada em valor,
  não em tipo de token.
- `#` é deliberadamente excluído dos símbolos compostos (igual ao
  original), então `print #1,1` ganha espaço só depois do `print`.
- Número de linha sempre ganha espaço depois de si, incondicional.
- Linha com mais de 256 caracteres é erro fatal.
- Novo `Description.CompoundSymbols()`; `ReservedKeywords()` (já
  existia desde o `DECLARE`) é reaproveitado sem mudança.
- 5 testes novos focados em espaçamento; ~43 asserções de saída já
  existentes corrigidas para o texto real (não regressão — era o
  próprio objetivo da release).

### Não mudou / ainda falta

- Pass 1 do separador `_` fora do fim de linha — depois disso, falta
  só ligar tudo ao tokenizer MSX Batoken já pronto.

---

## [1.14.0] — 2026-11-22 — bloco **NEW**, feature **include** (build 0)

**`INCLUDE`** — mesclagem de outro arquivo Dignified.

### Adicionado

- `INCLUDE "arquivo"` carrega, lexa e processa outro arquivo (Pass
  1-3) e mescla seu conteúdo no lugar do próprio `INCLUDE`; caminhos
  relativos resolvem contra o diretório do arquivo PRINCIPAL, não do
  arquivo que inclui (mesmo em includes aninhados) — fiel ao
  original.
- Tabelas de variável (`Declares`/`HardShortVars`/`HardLongVars`) e o
  alocador de alias do MSX são automaticamente compartilhados entre
  o arquivo principal e qualquer include — zero mudança em
  `internal/platforms/msx/parser.go`.
- Novo `CompileContext.LoadFile` (I/O de disco injetável) e
  `core.LoadLines` (implementação real) — primeira vez que o `core`
  precisa ler um arquivo de verdade.
- 7 testes novos no mecanismo core-level + 1 de ponta a ponta no MSX
  lendo um arquivo real do disco.

### Não mudou / ainda falta

- Pass 1 do separador `_` fora do fim de linha, `generate()` de
  verdade (espaçamento por palavra reservada, limite de 256
  caracteres).

---

## [1.13.0] — 2026-11-22 — bloco **NEW**, feature **keep** (build 0)

**`KEEP`** — compilação condicional via tags `#tag`, usada no
`CGK173.dmx` (`keep #none` no topo, blocos `#a...#a`/`#b...#b`).

### Adicionado

- `KEEP #tag1 #tag2 ...` declara quais tags ficam no código final;
  sem `KEEP` nenhum, todo bloco com tag é removido por padrão
  (extras opt-in). `#ALL` mantém qualquer tag; `#NONE` sobrepõe
  qualquer outra tag mantida.
- `#tag` sozinho numa linha abre/fecha um bloco removido (ou mantido)
  inteiro; `#tag resto da linha` remove/mantém só aquela linha.
- Novo `core.KindToggleRem`; lookahead no `Lexer` MSX pra `#` não
  colidir com o símbolo de canal de arquivo já existente (`PRINT #1`).
- 8 testes novos no mecanismo core-level + 1 de ponta a ponta no MSX.

### Não mudou / ainda falta

- `INCLUDE`, Pass 1 do separador `_` fora do fim de linha,
  `generate()` de verdade (espaçamento por palavra reservada, limite
  de 256 caracteres).

---

## [1.12.0] — 2026-11-22 — bloco **NEW**, feature **declare** (build 0)

**`DECLARE`** — reserva manual de alias de variável, usada no
`CGK173.dmx` (`keep #none` e vizinhos).

### Adicionado

- `DECLARE nome[:curto], ...` — sem `:`, mantém o nome por extenso
  (mesmo efeito de `~nome`) ou reserva um nome já curto; com `:curto`,
  atribui o alias manualmente, pulando o alocador automático.
- Escreve direto nas 3 tabelas que o `Parser` MSX Pass 3/4 já
  consultam — **zero mudança** em `internal/platforms/msx/parser.go`
  foi necessária.
- 2 métodos novos em `core.Description`.
- 8 testes novos no mecanismo core-level + 1 de ponta a ponta
  confirmando que o Pass 3/4 já existente respeita o que `DECLARE`
  escreveu.

### Observações

- Corrigido um erro de digitação real do original (`c_varsna`'s
  `[a-zA-z]`, que por acidente de intervalo ASCII aceitava
  `[\]^_` também) — usado o padrão claramente pretendido em vez de
  replicar o typo.
- `INCLUDE`/`KEEP`, Pass 1 (`_` fora do fim de linha) e um
  `generate()` de verdade ainda faltam — ver `OUTLINE.md`, Bloco 2.

---

## [1.11.0] — 2026-11-22 — bloco **NEW**, feature **func-ret** (build 0)

**`FUNC`/`RET`** — funções Dignified, usadas em quase toda sub-rotina
de `CGK173.dmx` (`.move_cursor_files`, `.show_content`, etc.).

### Adicionado

- `func .nome(param=padrão, ...)` ... `ret retvar, ...` declara uma
  função; chamada `.nome(args)` (em qualquer lugar, antes ou depois
  da declaração no texto) expande pra `param = valor : gosub <linha>`
  seguido, se a chamada capturar retorno (`a = .nome()`), de
  `: a = retvar`.
- Atribuição redundante (`param = param`) é sempre pulada, igual o
  original.
- Parâmetro omitido na chamada fica intocado (nem o padrão é
  aplicado) — mesmo efeito de um `param=param` redundante, dado que
  o padrão real é sempre "usa o valor atual da própria variável".
- 5 métodos novos em `core.Description`.
- Primeira pass inteiramente nova (`pass2Functions`) fora do molde
  "envolve o hook da plataforma" — a captura de retorno precisa
  remover retroativamente tokens já emitidos, mesmo tipo de problema
  do `DEFINE`, só pra trás.
- 9 testes novos no mecanismo core-level + 1 de ponta a ponta
  confirmando que a renomeação de variável do MSX alcança
  parâmetro/padrão de função.

### Observações

- `INCLUDE`/`KEEP`/`DECLARE`, Pass 1 (`_` fora do fim de linha) e um
  `generate()` de verdade ainda faltam — ver `OUTLINE.md`, Bloco 2.
- Achado real replicado fielmente: a linha de declaração de `FUNC`
  desperdiça 1 número de linha (diferente de label) — confirmado que
  o `pass_3` original só remove o `NEWLINE` depois de uma declaração
  de label, não depois de `FUNC`.

---

## [1.10.0] — 2026-11-22 — bloco **NEW**, feature **define** (build 0)

**`DEFINE`** — macro de texto Dignified, usada de verdade em
`CGK173.dmx` (`[up]`, `[var0](&h20)`, `[pauseKEY]`, etc.).

### Adicionado

- `DEFINE [nome][conteúdo], [nome2][conteúdo2], ...` no início de
  linha; chamada `[nome]`/`[nome](arg)` em qualquer outro lugar,
  expandindo pro conteúdo registrado.
- Placeholder opcional (`[0]` dentro do conteúdo) com valor padrão
  quando a chamada não dá argumento explícito.
- Chamadas aninhadas recursivas (`[a]([b])`).
- A linha inteira do `DEFINE` desaparece sem deixar rastro (nem uma
  linha em branco).
- Primeira pass do `core` com cursor explícito em vez do modelo
  `passFunc`/`runPass` (necessário porque conteúdo/argumento têm
  comprimento variável e recursivo).
- 7 testes novos no mecanismo core-level + 1 de ponta a ponta
  confirmando que convive com label e renomeação de variável.

### Observações

- `FUNC`/`RET`/`INCLUDE`/`KEEP`/`DECLARE`, Pass 1 (`_` fora do fim de
  linha) e um `generate()` de verdade ainda faltam — ver
  `OUTLINE.md`, Bloco 2.

---

## [1.9.1] — 2026-11-22 — bloco **NEW**, feature **loop-labels** (build 1)

**Correção de bug**: o padrão de identificador Dignified compartilhado
(`internal/core`) era restrito a ASCII minúsculo, cortando nomes
camelCase/mixedCase (`gameName$`, `fill_VRAM`, `pauseKEY`) no primeiro
caractere maiúsculo. O motor original é case-insensitive por completo
(`re.I` no regex combinado) — corrigido com a flag `(?i)` no mesmo
padrão. Bug latente desde as rodadas de variável/label, nunca
disparado porque nenhum teste usava nomes mistos; achado pesquisando
`DEFINE` e confirmado em uso real no `CGK173.dmx`.

---

## [1.9.0] — 2026-11-22 — bloco **NEW**, feature **loop-labels** (build 0)

**Loop labels** (`nome{ ... }` como laço fechado + `exit`) — usado de
verdade em `CGK173.dmx` (`keys{ }`, `read_keyboard_main{ }` etc.).

### Adicionado

- `nome{` abre um laço fechado (mesma declaração de label simples,
  só a detecção difere); `}` solto fecha, emitindo um `GOTO`
  implícito de volta pro início (com separador `:` só quando não está
  sozinho na linha); `exit` dentro do laço salta pra depois do
  fechamento.
- `core.Description` ganhou `StatementSeparator()`/
  `LoopBackInstruction()`.
- Erro fatal pra `}` sem abertura, laço não fechado, `exit` fora de
  qualquer laço, ou `exit` sem linha real pra pousar.
- 11 testes novos no mecanismo core-level + 1 de ponta a ponta
  confirmando que convive com a renomeação de variável do MSX.

### Observações

- `DEFINE`/`FUNC`/`RET`/`INCLUDE`/`KEEP`/`DECLARE`, Pass 1 (`_` fora
  do fim de linha) e um `generate()` de verdade ainda faltam — ver
  `OUTLINE.md`, Bloco 2.

---

## [1.8.0] — 2026-11-21 — bloco **NEW**, feature **label-numbering** (build 0)

**Resolução de labels e numeração de linha** — a peça que faltava
pra qualquer programa Dignified real (mesmo o mais simples, com um
`GOTO {algumlugar}`) compilar de ponta a ponta. Validada contra um
exemplo real completo (`CGK173.dmx`, "Change Graph Kit") que o
usuário apontou.

### Adicionado

- `{nome}` no início de linha é declaração; em qualquer outro lugar é
  referência de salto; `{@}` aponta pra linha atual. Label nunca
  gasta linha própria — "cola" na próxima linha com conteúdo real.
- Linhas em branco somem sem gastar número; `_` no fim de linha funde
  com a linha seguinte numa só linha numerada.
- Erro fatal ao saltar pra um label não declarado.
- **`internal/core` passou a fazer trabalho Dignified de verdade
  antes de delegar pra plataforma** — a primeira vez que isso
  acontece (`pass2Labels`/`pass3LineHygiene`/`pass4Numbering`), não
  só documentado como intenção futura.
- 10 testes novos (6 no mecanismo core-level, 1 de ponta a ponta
  confirmando que convive com a renomeação de variável do MSX).

### Observações

- Loop labels (`nome{ }`/`exit`), inserção de header, relatório de
  labels (`-lbr`), `DEFINE`/`FUNC`/`RET`/`INCLUDE`/`KEEP`/`DECLARE` e
  o Pass 1 (`_` fora do fim de linha) ainda faltam — ver
  `OUTLINE.md`, Bloco 2.

---

## [1.7.0] — 2026-11-21 — bloco **NEW**, feature **msx-parser-pass5** (build 0)

**`Parser` MSX Pass 5** — as 5 conversões finais específicas do
Classic, fechando as conversões de token-para-token do `Parser`.

### Adicionado

- `?`/`PRINT` convertível nos dois sentidos, strip de `THEN` antes de
  `GOTO` (ou `GOTO` depois de `THEN`/`ELSE`), expansão de atribuição
  composta/incremento/decremento (`x+=5`→`x = x + 5`, `x++`→
  `x = x + 1`), `TRUE`/`FALSE`→`-1`/`0`, tradução Unicode→ASCII de
  caracteres literais — tudo fiel a `pass_5` do motor original.
- **Quarto desvio na interface `core.Parser`**: expandir 1 token em
  vários (a atribuição composta) exigiu trocar o payload de `Replace`
  de `Token` para `[]Token`. `core.Lex()` ganhou reconhecimento de
  `TRUE`/`FALSE` (`KindDignifiedOperator`).
- 8 testes novos.

### Observações

- Pass 1 (separador de chamada Dignified `_`) e um `generate()` de
  verdade (espaçamento por palavra reservada, limite de 256
  chars/linha) ainda faltam — ver `OUTLINE.md`, Bloco 2.
- `ConvertPrint`/`StripThenGoto`/`Translate` só são configuráveis via
  código por ora; os flags de CLI (`-cp`/`-tg`) ainda não existem.

---

## [1.6.0] — 2026-11-21 — bloco **NEW**, feature **msx-parser-variables** (build 0)

**Primeira conversão Dignified→Classic de verdade produzindo texto
correto**: `Parser` MSX Pass 3 e Pass 4 — a alocação de variáveis de
2 letras, o coração do `Parser`.

### Adicionado

- Novo `internal/platforms/msx/parser.go`: aloca aliases de 2 letras
  pra variáveis longas, descendente de `zz` até `aa`, fiel a
  `get_hard_variable`/`process_variable` do motor original. Nome já
  curto fica inalterado (e reserva esse nome pros outros); `~nome`
  fica mantido por extenso (o `~` desaparece do texto gerado); erro
  fatal ao esgotar os 676 aliases ou usar `~` incorretamente.
- 7 testes novos, incluindo um de ponta a ponta completo
  (`Lex`→`Parse`→`Generate`) confirmando que `counter=0` vira `zz = 0`
  de verdade.
- **Terceiro desvio na interface `core.Parser`** (Bloco 1): precisa
  olhar pro token anterior pra detectar o símbolo `~` antes de um
  identificador — cada `PassN` agora recebe a lista inteira de
  entrada da pass mais o índice do token atual, não só o token
  isolado. `core.Lex()` ganhou reconhecimento dos 6 símbolos
  Dignified compartilhados (`~ [ ] { } @`). `core.SymbolTables` ficou
  precisa (`Declares`/`HardShortVars`/`HardLongVars`).
  `core.VariableAllocator.Allocate` ganhou um predicado `excluded`.

### Observações

- Não portados (não-fatais no original): os avisos de colisão por
  prefixo em `get_hard_variable`.
- Ainda faltam Pass 1 (separador de chamada Dignified), Pass 5
  (`?`/PRINT, `THEN`/`GOTO`, `TRUE`/`FALSE`, tradução Unicode) e um
  `generate()` de verdade (espaçamento, limite de linha) — ver
  `OUTLINE.md`, Bloco 2.

---

## [1.5.0] — 2026-11-21 — bloco **NEW**, feature **msx-lexer** (build 0)

**`Description` e `Lexer` MSX portados de verdade** — o primeiro uso
real das interfaces "v0" do motor (`internal/core`, Bloco 1), que até
aqui só tinham sido validadas contra uma plataforma fictícia.

### Adicionado

- Novo `internal/platforms/msx/description.go`: todo o vocabulário do
  dialeto MSX (instruções, funções, operadores, símbolos, símbolos
  compostos, palavras-chave de salto) transcrito de `badig_msx.py`.
- Novo `internal/platforms/msx/lexer.go`: classifica texto Dignified
  MSX em tokens — palavras-chave (case-insensitive), operadores,
  símbolos simples e compostos (`+=`, `++`, etc.), números (decimal,
  `&H`/`&O`/`&B`), comentários (`REM`/`'`), `DATA`. 8 testes novos,
  incluindo um fallback tolerante para caractere não reconhecido (não
  aborta a conversão).
- **Corrigida a interface `core.Lexer`** (Bloco 1): a v0 dava à
  plataforma uma "palavra" pré-cortada por espaço em branco — não
  funciona para números com notação científica nem símbolos compostos
  sem espaço (`x+=5`). Agora a plataforma recebe a **linha inteira
  restante** e devolve quantos caracteres consumiu mais zero ou mais
  tokens (mesmo padrão já usado no tokenizer Batoken). `core.Lex()`
  ganhou fallbacks genéricos compartilhados entre plataformas
  (identificador Dignified, string entre aspas), usados só quando a
  plataforma não reconhece nada na posição atual.

### Observações

- Ainda não há nenhuma conversão Dignified→Classic completa — falta o
  `Parser` MSX (5 passes, inclusive a alocação de 2 letras por
  variável) para fechar o caminho de ponta a ponta. Ver `OUTLINE.md`,
  Bloco 2.
- Fora de escopo deste release (nem bug, só adiado): bloco de
  comentário `''` multi-linha, semântica literal dos itens do `DATA`,
  tradução Unicode→ASCII (isso é trabalho do `Parser`, Pass 5).

---

## [1.4.0] — 2026-11-21 — bloco **NEW**, feature **msx-tokenizer** (build 0)

**Primeira conversão de BASIC de verdade: o tokenizer MSX Batoken
portado por completo.** Texto ASCII numerado → binário `.bmx`,
standalone (não depende do `internal/core` do Bloco 1 — o tokenizer
original também não depende do motor Dignified/`Classic`).

### Adicionado

- Novo pacote `internal/platforms/msx` (`tokenizer.go` +
  `tokenizer_test.go`): tabela de ~150 tokens transcrita **em ordem**
  (não por hash — a ordem de varredura do original importa de
  verdade, ex. `ERROR` antes de `ERR`), valores de tamanho variável
  (1-3 bytes), codificação de inteiros em 3 faixas (0-9 byte dedicado,
  10-255 prefixo `0x0f`, 256-32767 prefixo `0x1c` — acima disso
  auto-promove pra float, igual o MSX real), alvo de `THEN`/`GOTO`/
  etc. como valor bruto de 2 bytes com marcador `0x0e` (sem resolução
  de endereço), `&H`/`&O` calculando o valor real vs. `&B` guardando
  os dígitos como ASCII literal (discrepância real do MSX, portada
  fiel), scan de identificador que re-checa a tabela de tokens a cada
  caractere (gotcha classico onde `SCORE` quebra em `SC`+`OR`+`E`).
- **Corrigida uma suposição errada do `OUTLINE.md`**: o parsing
  numérico não usa Microsoft Binary Format nem nenhum formato binário
  de float — é puramente decimal (um byte de expoente + dígitos
  decimais empacotados dois por byte, já que cada dígito é um nibble
  hex válido). Ver `OUTLINE.md`, Bloco 2, para a explicação completa.
- Os 13 vetores de ouro do tokenizer (gerados no Bloco 0) portados
  1:1 como testes Go, todos passando byte-exato; validado ainda
  contra 10 casos extras (DATA, CALL, `&H`/`&O`/`&B`, variável com
  palavra-chave embutida, `ON...GOTO` com múltiplos alvos, inteiro
  grande sem sufixo, atribuição, sufixo `%`) comparando diretamente
  com o Python original, byte a byte.

### Observações

- Simplificações deliberadas (nenhuma afeta os vetores de ouro):
  notação científica (`1E+10`) não suportada ainda (erro explícito);
  sem espaço embutido dentro de um literal numérico; literais maiores
  que cabem num `int64` (~18-19 dígitos) retornam erro em vez de
  aritmética de precisão arbitrária. Um bug latente do Python no
  tratamento de `AS` com 2 dígitos (`ord()` numa string de 2
  caracteres, sempre quebrava) foi corrigido, não reproduzido.
- Ainda não há nenhuma conversão Dignified→Classic de verdade
  (`Description`/`Lexer`/`Parser` MSX) — esta release só tokeniza
  texto Classic ASCII já numerado. Ligar isso ao motor Dignified do
  Bloco 1 é o próximo passo do Bloco 2.

---

## [1.3.0] — 2026-11-21 — bloco **NEW**, feature **core-engine** (build 0)

**Primeiro código do motor de conversão.** Até aqui, tudo em goBaDig
era infraestrutura/interface (CLI/TUI/GUI, temas, editor) — nenhuma
linha de BASIC foi convertida ainda. Esta release começa o núcleo do
motor propriamente dito (Bloco 1 do `OUTLINE.md`): ainda não há
nenhum dialeto real (MSX/ZX81/ZX Spectrum) usando isso, e nenhum
comando de conversão está disponível para o usuário final.

### Adicionado

- Novo pacote `internal/core`: `Token`/`Position`/`TokenKind` (`Value`
  é `any`, não `string` — o Pass 4 do motor original reaproveita o
  mesmo campo pra guardar um número de linha resolvido), `HookAction`
  (substitui o protocolo tri-estado `False`/`None`/`Token` do Python
  original por um enum explícito `Continue`/`Skip`/`Replace`),
  diagnóstico/erro (substitui o `Infolog` original — sem o bug do
  exit code silencioso compartilhado como atributo de classe),
  `CompileContext`/`SymbolTables` (estado explícito por parâmetro, não
  atributos dinâmicos de instância), alocador de variáveis por
  categoria genérico e descendente (generaliza MSX/ZX81/ZX Spectrum).
- Interfaces "v0" `Description`/`Lexer`/`Parser`/`Settings` e o struct
  `Platform` que agrupa as 4 (equivalente Go ao trio dinamicamente
  importado `Classic`/`Tools`/`Dignified` do motor original - Go não
  tem esse truque, então a escolha de dialeto é um valor explícito).
- Driver do pipeline (`Lex`→`Parse`→`Generate`→`Run`), validado de
  ponta a ponta contra uma plataforma fictícia mínima (não é MSX de
  verdade - só reconhece `PRINT`/inteiros) antes de qualquer dialeto
  real depender do contrato. `Parse` já respeita o mesmo gate do motor
  original: passes 4/5 só rodam no arquivo principal, nunca em
  arquivos trazidos por `INCLUDE`.

### Observações

- Ainda faltam, documentados como próximas fases em `OUTLINE.md`
  (Bloco 1): `Expose`/`Tools.Interface` (só importam com um CLI real
  de flags, Bloco 5), `INCLUDE` recursivo de verdade, o caminho `-asc`,
  e a resolução de fato das 4 camadas de config (SQLite global <
  projeto < CLI < remtags) - por ora `CompileSettings` é só um
  placeholder.
- O scanner de hoje só reconhece espaço em branco e fim de linha;
  qualquer vocabulário real de dialeto (palavras-chave, números,
  strings) é trabalho do Bloco 2, quando o módulo MSX for de fato
  portado.

---

## [1.2.0] — 2026-11-21 — bloco **NEW**, feature **themes** (build 0)

**Temas persistidos nas três interfaces, TUI redesenhada no visual
retrô Turbo Vision/Norton Commander (com o bug de "congelamento" real
do `Setup > Install` encontrado e corrigido), e o editor WibEdit
integrado à TUI como base da futura IDE.** Acumula tudo que ficou sem
release desde a `v1.1.4`.

### Adicionado

- Catálogo inicial de seis temas para a GUI (VS Code Dark, Dracula,
  Monokai, Nord, Solarized Light e Gruvbox Dark), selecionáveis em
  `Setup > Install`; a seleção é aplicada imediatamente e persistida
  separadamente do catálogo no SQLite do usuário.
- Catálogos e preferências independentes dos mesmos seis temas para a TUI
  (`Setup > Install`) e para a CLI (`theme <nome>`), ambos persistidos no
  SQLite do usuário.
- TUI reorganizada como desktop clássico de terminal, com barra de título,
  fundo azul, menu, área de trabalho e barra de status; `Setup > Install`
  abre um diálogo central com dropdown, descrição e botões Apply/Cancel.
- Conclusão interativa de comandos e temas com `Tab` no console; um
  segundo `Tab` lista as opções disponíveis.
- TUI modernizada no visual retrô estilo MS-DOS anos 1990 (Turbo Vision/
  Turbo Pascal, Norton Commander): janelas com borda dupla, caixa de
  sistema `[■]` e sombra projetada (`internal/tui/chrome.go`); ícones
  Unicode nos menus e itens (`▭ File`, `⚙ Setup`, `ℹ Help`, `✕ Exit`,
  `◐ Install`, `★ About`); desktop com textura pontilhada (`▒`) e
  banner central com régua decorativa. Testado por simulação de
  terminal (`internal/tui/freeze_diagnosis_test.go`).
- **Corrigido o congelamento real de `Setup > Install`** (reportado como
  pendente nesta mesma seção até 2026-11-20): a causa nunca foi um
  deadlock — era o diálogo de temas inteiro ficando invisível. A
  chamada `.SetBackgroundColor(...)` encadeada direto no construtor do
  `tview.Grid` retorna `*tview.Box` (método herdado, sem override em
  `Grid`), então a variável do diálogo virava um `*Box` liso em vez de
  `*Grid` — `Box.Draw()` roda e sai antes de posicionar/desenhar
  qualquer filho, sem erro nem log algum. Corrigido em
  `internal/tui/tui.go` separando essa chamada em instrução própria.
  Achado e confirmado isolando uma cópia local do `rivo/tview` com
  instrumentação (`go.mod` `replace` temporário) até provar que
  `Grid.Draw()` nunca era chamado. Novo teste de regressão
  (`internal/tui/install_dialog_visible_test.go`) lê o conteúdo real da
  tela simulada e falha se o diálogo voltar a ficar invisível; o teste
  de deadlock antigo (`freeze_diagnosis_test.go`) não detectava esse
  bug, já que a aplicação nunca parou de responder.
- **Editor WibEdit integrado à TUI** como base do futuro editor/IDE do
  goBaDig: novo pacote `internal/editor` (portado de
  `E:\wibworks\internal\modules\edit`, mesmo autor — editor estilo
  WordStar/Norton Editor 2, ~9 mil linhas + ~5,2 mil linhas de testes,
  já no mesmo stack `tcell`/`tview` do goBaDig) e `internal/printpdf`
  (exportação PDF). Acessível pelo novo menu **Editor > Open**
  (`internal/tui/tui.go`), embutido como mais uma página do
  `tview.Pages` existente — não é um binário separado. Identidade visual
  (diálogo About, ajuda, nome enviado à impressora) adaptada de
  "WIBEditor/WIBWorks" para "WibEdit (goBaDig)", mantendo o crédito de
  autoria. Novo teste de regressão
  (`internal/tui/editor_menu_test.go`) confirma que o menu abre o editor
  e que fechá-lo retorna à área de trabalho do goBaDig. Traz suporte a
  impressão e exportação PDF (nova dependência `go-pdf/fpdf`;
  `golang.org/x/image` atualizado de v0.24.0 para v0.48.0).
- Primeira fatia do armazenamento SQLite com driver puro Go
  `modernc.org/sqlite`, tabelas `themes` e `active_themes` por modo.
- `build.ps1` (raiz do repo): script PowerShell que automatiza o build
  local completo — baixa dependências (`go mod download`), roda
  `go vet`/`go test`, compila e monta `dist/goBaDig-vX.Y.Z/` + o zip,
  lendo a versão direto de `internal/version/version.go`. Não cria tag
  Git nem release no GitHub (continua manual). Testado de ponta a
  ponta, inclusive com `-SkipDownload -SkipTests`.
- `current/` (raiz do repo, gitignored por completo): convenção nova —
  toda mudança em qualquer programa do goBaDig recompila o executável
  mais atual ali, para teste manual rápido, diferente do pacote de
  release versionado em `dist/`. `current/resource/` guarda arquivos de
  trabalho do usuário e nunca é limpo. Regra documentada em `CLAUDE.md`.
- Investigada e confirmada a viabilidade do **ZRCP** (protocolo de
  controle remoto do emulador ZEsarUX) para monitoramento de erro em
  tempo real em ZX81/ZX Spectrum, lendo o código-fonte do ZEsarUX
  (não há documentação oficial do protocolo). Técnica recomendada:
  breakpoint em `PC=0008H` (onde ambas as ROMs Sinclair reportam erro
  via `RST 08h`), lendo depois `PPC`/`SUBPPC`/`ERR_NR` (Spectrum) ou
  `E_PPC`/`ERR_NR` (ZX81). Ainda não implementado — ver `OUTLINE.md`,
  Bloco 0, e `docs/RESOURCES.md`.
- Gerados os vetores de ouro do tokenizer MSX (`badig/msx/
  test_tokenizer_msx.py`, 13 casos, não versionado — `badig/` é
  histórico congelado), copiados também para
  [`docs/golden-vectors/msx-tokenizer.md`](docs/golden-vectors/msx-tokenizer.md)
  (versionado): MSX não tinha suíte de testes no projeto original,
  diferente de ZX81/ZX Spectrum.

---

## [1.1.4] — 2026-11-05 — bloco **NEW**, feature **outline** (build 4)

**Modo console interativo** (`--cli`): um prompt de verdade, que fica
aguardando comandos do usuário — faltava isso para fechar a paridade
entre CLI/TUI/GUI (as três agora têm um modo interativo dedicado:
`--cli`, `--tui`, `--gui`).

### Adicionado

- `goBaDig --cli` abre um console com prompt `goBaDig> `, que lê
  comandos linha a linha até `exit`/`quit` ou fim de entrada (Ctrl-D).
- Comandos do stub: `help`, `man`, `about`, `version`; comando
  desconhecido avisa e continua no prompt (não derruba o console).
- Testes automatizados (saída esperada, parada limpa em EOF) e teste
  manual via pipe de comandos.

### Observações

- Comandos reais de conversão (abrir arquivo, escolher plataforma,
  tokenizar, rodar emulador) chegam junto com o motor (Blocos 1-4 do
  `OUTLINE.md`) — por ora o console só tem os comandos de
  ajuda/identidade que já existiam como flags.

---

## [1.1.3] — 2026-11-05 — bloco **NEW**, feature **outline** (build 3)

**Menu da TUI reescrito** a pedido do usuário: em vez do menu em tela
cheia da 1.1.2, agora é um **menu suspenso clássico de MS-DOS dos anos
1990**, no estilo Turbo Pascal 7 / Norton Commander / Norton Editor.

### Adicionado

- Barra de menu fixa no topo (`File`, `Help`), item ativo destacado com
  cores invertidas (branco sobre azul), paleta clássica DOS (barra
  clara, "desktop" azul).
- Caixa suspensa (dropdown) desenhada **sobre** o conteúdo, posicionada
  diretamente abaixo do item da barra — não substitui a tela como antes.
- Navegação: setas Esquerda/Direita movem entre itens da barra; com o
  menu aberto, Esquerda/Direita troca direto para o dropdown vizinho
  (padrão clássico Turbo Pascal); Cima/Baixo navegam os itens do
  dropdown; `Enter`/`Down`/`F10` abrem; `Esc` fecha o suspenso;
  `Alt+letra` abre o menu direto pela primeira letra (`Alt+F`, `Alt+H`).
- Diálogo "About" recolorido para combinar com a paleta clássica.

### Observações

- Sombra de borda (drop shadow) do visual original do Turbo
  Pascal/Norton Commander **não foi implementada** — custo de
  engenharia desproporcional para um stub; pode voltar depois como
  polimento visual.
- Clique de mouse funciona nos itens do dropdown (nativo do
  `tview.List`), mas ainda não na barra superior em si.
- GUI (Fyne) não muda nesta release — continua com o menu nativo do
  sistema operacional, como antes.

---

## [1.1.2] — 2026-11-05 — bloco **NEW**, feature **outline** (build 2)

**Primeiro código Go do projeto.** Programa STUB real: compila, roda e
tem CLI (`--help`/`--man`), TUI e GUI mínimas, com paridade de menu
entre as três interfaces. Ainda sem conversor/tokenizer de BASIC.

### Adicionado

- Módulo Go inicializado (`github.com/wilsonpilon/goBaDig`), com
  `cmd/goBaDig` e `internal/{about,cli,gui,tui,version}`.
- **CLI**: `--help` (lista de opções), `--man` (referência estilo
  `man`, só o "about" nesta versão), banner padrão sem argumentos.
- **TUI** (`tview`/`tcell`): menu `File > Exit` e `Help > About`.
- **GUI** (`Fyne`): janela com menu `File > Exit` e `Help > About`
  (diálogo nativo).
- Texto "about" centralizado em `internal/about`, compartilhado pelas
  três interfaces, para não divergirem (ver `OUTLINE.md`, seção 6,
  item "Paridade de opções entre interfaces").
- Testes unitários do dispatch de modo da CLI (`internal/cli`).
- Build, `go vet` e `go test` passando; binário testado em runtime no
  Windows (CLI, TUI e GUI abrem e fecham sem erro).
- `docs/RESOURCES.md`: registrados os projetos próprios do usuário
  `fwMSX`, `go81` (conteúdo para os futuros manuais de BASIC
  embutidos), `paleobasic` e `msxide` (implementações prontas do
  Norton Editor — TUI e GUI — para quando o editor próprio for
  retomado; localizações a confirmar com o usuário).
- `OUTLINE.md`, seção 10.7: desenho do sistema de ajuda — `--help` +
  `--man` ao estilo Unix na CLI (cobrindo, no futuro, as opções do
  BaDig e os manuais de sintaxe BASIC do MSX/ZX81/ZX Spectrum), e
  `Help > About` tradicional na TUI/GUI.
- Confirmado: `badig/` (projeto Python original) **nunca será
  commitado** — decisão do usuário, serve só de referência local
  durante o port. Adicionado a `.gitignore`.

### Observações

- Esta release já distribui um executável real (Windows) no pacote
  `goBaDig-v1.1.2.zip`, mas é um stub: nenhuma conversão de BASIC
  acontece ainda.

---

## [1.1.1] — 2026-11-05 — bloco **NEW**, feature **outline**

> Nomes de bloco (`X`) e feature (`Y`) propostos nesta sessão como
> primeira aplicação do esquema de versionamento acima — `NEW` por ser
> o comando BASIC (presente nos três dialetos) que limpa a memória e
> começa um programa do zero, metáfora para este ser o ponto de partida
> do projeto; `outline` por esta release conter só o plano e a estrutura,
> nenhum código funcional. Renomear livremente se preferir outros nomes.

**Release inicial do projeto — "nossa partida".** Nenhuma linha de
código Go existe ainda; esta versão marca o fim da fase de análise e o
início formal do repositório, com toda a arquitetura, o escopo e a
divisão de trabalho documentados.

### Adicionado

- Análise técnica completa do projeto Python original (`badig/`):
  núcleo do motor (`badig.py`/`support/`), módulos MSX/ZX81/ZX Spectrum,
  documentação original e as três extensões VSCode existentes.
- `OUTLINE.md` — plano vivo de arquitetura, decisões e divisão do
  trabalho em 11 blocos (pesos/prioridades), com perguntas em aberto
  registradas para o usuário decidir.
- Estrutura de diretórios do projeto: `dist/`, `resource/` (não
  versionados, exceto um `README.md` cada), `examples/` (com subpastas
  por plataforma), `docs/` (`MANUAL.md`, `RESOURCES.md`, `releases/`).
- Documentação nova: `README.md` (vitrine do projeto), `CHANGELOG.md`
  (este arquivo, com o esquema de versionamento), `docs/MANUAL.md`
  (manual de uso, em boa parte "(pendente)" até existir código),
  `docs/RESOURCES.md` (recursos externos: emuladores, ROMs, fontes).
- Esquema de versionamento `X.Y.Z` definido (ver seção acima).
- Escopo de emuladores ampliado: além de openMSX, EightyOne e
  Spectaculator, o projeto passa a visar também **fMSX** (MSX),
  **ZEsarUX** (ZX81/ZX Spectrum, com protocolo de controle remoto ZRCP)
  e **sz81/XZ81** (ZX81, Linux).
- Escopo de formatos de mídia ampliado: além de TAP/TZX/Z80/SNA
  (Spectrum) e `.p` (ZX81), o projeto passa a visar `.CAS`/`.DSK` (MSX)
  e `.81`/`.p81` (ZX81); `.DSK` +3 do Spectrum fica como item opcional
  de baixa prioridade.
- Suporte a editores priorizado: VSCode (1), Sublime Text (2). Um
  editor próprio baseado no Norton Editor 2.0 é mencionado como visão de
  longo prazo, mas **fica fora de escopo por enquanto**. Notepad++ e
  UltraEdit ficam como alvos opcionais ("se for possível").

### Observações

- Vários arquivos de documentação do projeto Python original
  (`badig/README.md`, `badig/documentation/*.md`) têm marcadores de
  merge conflict não resolvidos entre os branches ZX81 e ZX Spectrum —
  registrado como pendência em `OUTLINE.md`.
- `badig/` (o projeto Python original) permanece **fora do controle de
  versão neste repositório Go** por enquanto — decisão deliberadamente
  não tomada nesta release, para evitar importar binários (`.exe`,
  `.pyc`, `.vsix`) para o histórico do Git sem confirmação explícita.
  Ver `docs/RESOURCES.md`, seção "Material do projeto Python original".
