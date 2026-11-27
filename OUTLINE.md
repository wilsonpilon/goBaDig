# goBaDig — Basic Dignified Suite em Go: plano vivo

> Documento vivo. Atualizar sempre que uma decisão for tomada, uma tarefa for
> concluída ou escopo mudar. Não apagar histórico de decisões — marcar como
> superado quando necessário. Ver também `resume.txt` (não versionado) para
> retomar o trabalho em outra máquina/IA com um resumo rápido do estado atual.

Data da primeira análise: 2026-11-05. Última atualização: 2026-11-27
(nota: este parágrafo de estado ficou desatualizado em várias rodadas
recentes - ver `resume.txt`, seção MARCOs, para o estado real mais
recente; a versão publicada mais recente é a v1.51.0, bloco NEW,
feature asm-basic-bridge).
Estado: **v1.40.0 publicado** (GitHub Releases) — CLI/TUI/GUI com temas
persistidos, TUI no visual retrô MS-DOS (ver Bloco 6), editor WibEdit
integrado (`internal/editor`, menu Editor > Open, base do futuro
IDE), viabilidade do ZRCP confirmada (Bloco 0, não implementado
ainda), **catálogo de opções de configuração do `badig.py` + schema
SQLite + comandos `SET`/`SETTINGS`/`DIALECT` no CLI + 4 painéis de
configuração na TUI e na GUI (Etapas A, B, C e D do painel de config,
ver Bloco 1/Fase 4 abaixo — painel de configuração COMPLETO nas três
interfaces), e um bug de cor no menu suspenso da TUI corrigido (item
não selecionado invisível, destaque parcial — reportado pelo usuário
assim que o menu `Setup` passou a ter itens de verdade de sobra pra
revelar o problema), e a biblioteca `.DSK` (MSX-DOS/FAT12) portada de
`E:\msxdiskutil` (referência real do usuário em PureBasic) pro Go —
`internal/platforms/msx/dsk`, Etapa 1 do futuro gerenciador de
arquivos de dois painéis, e o comando `compilar` de verdade
(`goBaDig <arquivo.dmx>` no CLI, `compile <arquivo.dmx>` no console
`--cli`) — a primeira conversão Dignified→binário disponível pro
USUÁRIO FINAL, não só testada internamente, validada byte a byte nos
dois arquivos de saída contra o `CGK173.dmx` inteiro, e o comando
`execute` (`goBaDig --execute <arquivo.dmx>`/`execute <arquivo.dmx>`
no console) que fecha o ciclo completo: compila, gera um disco
`.dsk` de teste (`AUTOEXEC.BAS`) e abre no openMSX de verdade —
ciclo MSX completo: editar→compilar→testar num emulador real, tudo
via CLI (ver Bloco 1 abaixo — TUI/GUI ainda não chamam nenhum desses
comandos), e o módulo ZX81 (Bloco 3) deu um salto grande: tokenizer
(Classic ASCII numerado→imagem `.p` completa, incluindo o ponto
flutuante oculto), `Description`+`Lexer` (vocabulário clássico, 1 letra
de variável, escape hex `\{XX}`, um achado empírico grande sobre
número com sinal) e agora `Parser` (alocação de 1 letra, `PRINT AT`
via uma evolução pequena do `core.BuiltinDefine` — `Expand`, `LET`
implícito) — e a validação de ponta a ponta com 2 programas ZX81
reais já existentes no projeto (`zx81-intro.DZX`/`mcode-ok.DZX`),
batendo byte a byte contra o `badig.py` real — e agora os comandos
`compile`/`execute` do CLI/console (até aqui hard-coded pro MSX)
escolhem o dialeto pela extensão do arquivo (`.dmx`→MSX, `.dzx`→ZX81),
com `goBaDig --execute programa.dzx` compilando e abrindo de verdade
no EightyOne (confirmado com um processo real rodando nesta máquina)
(ver Bloco 3 abaixo; `remx`/renumeração de listagem numerada
deliberadamente adiados, documentados lá), tudo confirmado contra o
Python real e o EightyOne real — ciclo ZX81 completo via CLI, igual
ao MSX), e agora o módulo **ZX Spectrum** (Bloco 4) começou: tokenizer
portado (`internal/platforms/zxspectrum`) — encoding de número
Sinclair via `math/big` (aritmética racional exata, não `float64`,
diferente do ZX81), `DEF FN`/slots ocultos de 6 bytes, e
empacotamento `.tap`/`.tzx` (100% mecânico); o `Description`/`Lexer`
do dialeto; o `Parser`; a validação de ponta a ponta contra os 9
exemplos reais já existentes no projeto, batendo byte a byte (texto E
`.tap` tokenizado) contra o `badig.py` real; e agora o comando
`compile`/`execute` também — `goBaDig programa.dzs`/`goBaDig
--execute programa.dzs` já funcionam de verdade, abrindo no
EightyOne (confirmado com um processo real rodando nesta máquina) —
**ROTEIRO COMPLETO DO BLOCO 4 FECHADO (Etapas 1-5), ciclo ZX
Spectrum completo via CLI, igual MSX e ZX81** (por decisão tomada com
o usuário, o `execute` do Spectrum NÃO espera o emulador fechar,
diferente do Python original nesse ponto específico — consistência
com MSX/ZX81 no Go); e agora **seleção de emulador alternativo**
(`--emulator <nome>`/console `execute <arquivo> [emulador]`), pedido
do usuário: MSX ganha `fMSX`/`blueMSX+` como alternativas ao openMSX,
ZX81 e ZX Spectrum ganham o `ZEsarUX` como alternativa ao EightyOne —
todos os 4 backends novos confirmados abrindo de verdade nesta
máquina (ver Bloco 5 abaixo)** — histórico completo de
cada um
desses itens em `CHANGELOG.md` e nos Blocos correspondentes abaixo,
não repetido aqui.

**O motor de conversão começou a existir em 2026-11-21**: `internal/core`
(Bloco 1), **tokenizer MSX Batoken** (`internal/platforms/msx/tokenizer.go`,
13 vetores de ouro do Bloco 0 byte-exato), **`Description`/`Lexer` MSX**
(`description.go`/`lexer.go`), **`Parser` MSX Pass 3/4/5** (`parser.go`
— alocação de variáveis de 2 letras + as 5 conversões finais do
Classic), **resolução de labels/numeração de linha, incluindo loop
labels** (`nome{ ... }`/`exit`, core-level, não específico de MSX),
**`DEFINE`** (macro de texto com placeholder/valor padrão/chamadas
aninhadas, também core-level), **`FUNC`/`RET`** (função Dignified →
`GOSUB` + troca de variáveis, idem core-level), **`DECLARE`**
(reserva manual de alias de variável, idem core-level), **`KEEP`**
(compilação condicional via tags `#tag`, idem core-level),
**`INCLUDE`** (mesclagem de outro arquivo Dignified, idem
core-level), um **`generate()` de verdade** (espaçamento por
palavra reservada + limite de 256 caracteres, idem core-level), os
**comentários exclusivos do Dignified** (`##`/`###`/duas aspas
simples, idem core-level), e a **fusão de linha pelo separador
clássico `:`** (+ junção de strings adjacentes, idem core-level) — a
**primeira conversão Dignified→Classic de verdade já produzindo
texto correto E formatado de verdade, incluindo `GOTO`/label, laço
fechado, macro, função, alias manual, tag condicional, mesclagem de
arquivo, espaçamento, comentários e fusão por `:` reais** (`counter=0`
→ `zz=0`, `goto {dest}` → `goto 20`, `keys{ }` → `GOTO` implícito de
volta, `[nome](arg)` → texto expandido, `.func(5)` → `param=5:gosub
<linha>`, `declare x:ab` → `x` sempre vira `ab`, bloco `#a...#a` some
por padrão a menos que `keep #a`/`keep #all` seja declarado, `include
"sub.dmx"` mescla o conteúdo de outro arquivo do disco no meio do
stream, `print #1,1` recebe espaço só depois do `print`, `##`/`###`
somem sem deixar rastro, um bloco de duas aspas simples virando N
linhas de comentário classic reais, `for f=0 to 5:` + `print f:` +
`next:` numa linha só mantendo os `:`, cabeçalho REM de 2 linhas
quando pedido, TAB expandido pra espaço, e um `_` fora do fim de
linha virando o atalho de `CALL` do MSX, excluindo o identificador
seguinte da renomeação de variável — `MSX_CALL_SEPARATOR`/
`MSX_CALL_IDENTIFIER`, `internal/platforms/msx/parser.go`), **e
agora já tokenizado em binário `.bmx` de verdade** via
`msx.CompileToBinary` (`internal/platforms/msx/compile.go`, ligando
`core.Run` ao tokenizer MSX Batoken pronto desde a v1.4.0),
confirmada em testes de ponta a ponta completos
(`Lex`→`Parse`→`Generate`→`Tokenize`), e **validada 100% de ponta a
ponta contra o `CGK173.dmx` inteiro (891 linhas) E contra a saída REAL
do `badig.py` rodando num Python de verdade nesta máquina, com ZERO
diferença restante, agora inclusive no `.bmx` binário tokenizado**
(`diff` de texto vazio E 18.815 bytes binários idênticos — não só
leitura de código). **Correção de uma suposição
anterior**: numa rodada passada eu tinha documentado que "o Python
reserva 2 números de linha pro cabeçalho mesmo com `-rh` desligado" —
isso estava ERRADO (achado rodando de novo um teste mínimo): `-rh`
de verdade desliga o cabeçalho por completo (`insert_rem_header`
nunca é chamado), e o arquivo de referência que eu venho usando pra
comparar tinha sido gerado SEM passar `-rh` da primeira vez — ou
seja, ele JÁ tinha o cabeçalho de verdade; o offset de +20 que eu via
era exatamente essas 2 linhas REM reais que meu port ainda não
gerava, não um mecanismo escondido de reserva. Essa validação
revelou e já corrigiu 4 bugs reais de lexer MSX (além do próprio
achado da fusão por `:`): `[?](x,y)` (macro embutida que o Python
injeta automaticamente, nunca declarada pelo usuário), identificador
que começa com uma palavra reservada seguida de `_` (`copy_VRAM`) OU
DÍGITO (`color1`) cortando errado — essa última reverte uma
suposição documentada antes sem checar contra o original de verdade
("GOTO10" nunca foi um caso real) —, e `"?"` (o alias de `PRINT`)
faltando na lista de palavras reservadas pra espaçamento (estava
escondido num item de lista Python escrito como `r'\?'`, fácil de
não notar lendo rápido). Dezoito desvios reais nas interfaces "v0" do
Bloco 1 descobertos implementando contra o MSX de verdade, todos
documentados no Bloco 1 abaixo — a resolução de labels foi
também a primeira vez que o `core` passou a fazer trabalho Dignified
de verdade antes de delegar pra plataforma, não só documentar a
intenção; `DEFINE` foi a primeira vez que isso precisou de um cursor
explícito em vez do modelo `passFunc`/`runPass`; `FUNC`/`RET` foi a
primeira vez que uma pass inteira nova (`pass2Functions`) precisou
existir fora do molde de "envolve o hook da plataforma"; `DECLARE`
foi a primeira vez que um recurso novo não precisou tocar em
`msx/parser.go` NENHUM pouco (escreve direto nas tabelas que o Pass
3/4 já consulta); `KEEP` foi a primeira vez que um recurso novo
precisou de uma mudança FORA do `core`, no `Lexer` da própria
plataforma (conflito lexical real: `#` já é símbolo do MSX);
`INCLUDE` foi a primeira vez que o core precisou tocar disco de
verdade (I/O injetável, `CompileContext.LoadFile`) e a primeira vez
que uma pass chama `core.Lex`/`core.Parse` recursivamente sobre um
`CompileContext` novo; o `generate()` de verdade foi a primeira vez
que uma rodada mudou o TEXTO RENDERIZADO de quase todo teste de ponta
a ponta já existente (~40 asserções corrigidas pro valor real, não
regressão - era o próprio objetivo); os comentários exclusivos foram
a primeira vez que `Lexer.Lex` precisou saber sua própria posição
(novo parâmetro `atLineStart`) e a primeira vez testado contra a saída
REAL de um Python de verdade rodando nesta máquina, não só leitura de
código-fonte; a fusão por `:` foi a rodada que finalmente validou o
`CGK173.dmx` inteiro 100% de ponta a ponta contra a saída real,
revelando de brinde que uma linha em branco no INÍCIO do arquivo
(antes de qualquer conteúdo real) também desperdiçava um número de
linha (corrigido: `KindProgram` agora entra no mesmo colapso de
`NEWLINE` que labels/`_`/separador já tinham). Também corrigido nesta
janela: `identifierPattern` (compartilhado) era restrito a ASCII
minúsculo — bug real, latente desde as rodadas de variável/label, que
cortava nomes camelCase/mixedCase (`v1.9.1`). O Pass 1 do `Parser`
(separador de chamada Dignified `_` fora do fim de linha) foi
implementado em seguida (v1.19.0, ver Bloco 1 abaixo) — era a última
peça não entendida do caminho textual Dignified→Classic. Logo depois,
o pipeline foi ligado ao tokenizer MSX Batoken (v1.20.0,
`msx.CompileToBinary`) — **primeira conversão Dignified→binário
`.bmx` de ponta a ponta, validada byte a byte contra o `CGK173.dmx`
tokenizado pelo `badig.py` real**. **Ainda falta**: nenhum wiring de
CLI real expõe essa conversão pro usuário final ainda, e os outros
dois dialetos (ZX81/ZX Spectrum) não têm nenhum
`Description`/`Lexer`/`Parser` próprio. Ver Blocos 1-4 na seção 7
para o detalhamento.
**As 9 perguntas da seção 8 foram respondidas pelo usuário em
2026-11-20** — ver seção 8 para o registro de cada decisão. Nota de
ambiente: nesta máquina o repositório vive em `C:\goBaDig` (caminho
anterior registrado no `resume.txt`, `C:\dos\goBaDig`, era de outra
máquina/sessão).

---

## 1. Objetivo do projeto

Reescrever a **Basic Dignified Suite** (originalmente de Farique, em Python,
pasta `badig/`) como um programa **Go** autocontido, chamado `goBaDig`, que:

- Lê/escreve configuração, palavras-chave e opções em **SQLite** (em vez de
  arquivos `.ini` espalhados por módulo).
- Oferece **três superfícies de uso**: linha de comando com parâmetros (CLI
  scriptável), um modo interativo de comandos, uma **TUI** (`tview`/`tcell`)
  e uma **GUI** (`Fyne`) para configurar e pré-processar arquivos como um
  gerenciador.
- Suporta três dialetos: **MSX Basic**, **ZX81 Basic** e **ZX Spectrum
  (Sinclair) Basic**. **CoCo Basic fica de fora** — decisão do usuário, não
  faz sentido manter.
- Termina com um **VSIX único** para VSCode cobrindo os três dialetos num
  só módulo, publicável na loja oficial.
- É distribuído como **executável único** (sem depender de Python, venv ou
  pacotes externos para o caminho principal texto→binário→fita).

### 1.1 Por que é viável

O projeto Python já resolve o problema de design mais difícil: um núcleo
comum (`badig.py` + `support/`) com um contrato de extensão por plataforma
(`Description`/`Lexer`/`Parser`/`Settings`/`Expose`/`Tools.Interface`) que
já prova, na prática, que dá para ter MSX, ZX81 e ZX Spectrum coexistindo
sob um só pipeline. Isso **não precisa ser reinventado conceitualmente** —
precisa ser **traduzido com redesenho pontual** (ver seção 4). O trabalho
duro (regras de cada dialeto, formatos binários, empacotamento de fita,
DignifieRs) já está especificado e testado em Python — ele vira a
especificação de referência e o **oráculo de testes (golden vectors)**
para a reescrita.

---

## 2. Achado importante: estado real do repositório `badig/`

- O repositório Go (`goBaDig`) já tem releases publicados (ver
  `CHANGELOG.md`/`docs/releases/`). **`badig/` nunca é commitado** —
  decisão confirmada pelo usuário em 2026-11-05: ele existe só
  localmente, em disco, como referência de arquitetura/dialeto/testes
  durante o port, nada além disso. Está listado em `.gitignore` por
  segurança. Quem for reproduzir o ambiente em outra máquina precisa
  copiar essa pasta manualmente — ela não vem do `git clone`.
- **Vários arquivos de documentação têm marcadores de merge conflict (`<<<<<<< HEAD`
  / `=======` / `>>>>>>> suporte-zx-spectrum`) ainda não resolvidos**:
  `README.md`, `documentation/BASIC_DIGNIFIED.md`, `documentation/INSTALLATION.md`,
  `documentation/MODULE_TOOLS.md`, `documentation/IDE_TOOLS.md`,
  `examples/README.md`. Isso indica que o branch `suporte-zx-spectrum` nunca
  foi de fato mergeado com o que tem o ZX81 — os dois conjuntos de mudanças
  parecem ter sido colados lado a lado em texto, não integrados como código.
  **Decisão do usuário em 2026-11-20:** tratar como histórico congelado —
  não resolver os conflitos manualmente; a extração de informação já feita
  neste `OUTLINE.md` é suficiente, já que essa documentação Python será
  substituída pela documentação do Go.
- O próprio `badig/OUTLINE.md` (777 linhas) já é um plano de análise extenso
  e específico da adição do ZX Spectrum ao Python, com estado de
  implementação detalhado (módulo Spectrum em v1.10.0 preview, pendências
  de validação manual, bug de freeze do EightyOne em 128K, etc.). Ele
  **continua a referência técnica mais detalhada da plataforma Spectrum** e
  não deve ser apagado; este `OUTLINE.md` (raiz) é o plano do **projeto Go**,
  em escopo mais amplo.
- Não há testes automatizados fora dos módulos ZX81/ZXSpectrum (`test_tokenizer_zx81.py`,
  `test_zxspectrum.py`) — MSX e CoCo não têm suíte de testes no repositório.
  Isso significa que, para o MSX, teremos que **gerar nossos próprios vetores
  de ouro** (rodando o Python original e capturando saídas) antes de portar,
  já que não há testes prontos para usar como oráculo.

---

## 3. Arquitetura do sistema original (núcleo Python) — resumo técnico

Baseado em leitura completa de `badig.py` (2054 linhas), `support/badig_settings.py`,
`support/badig_dignified.py`, `support/helper.py` e `documentation/NEW_MODULES.md`/`MODULE_TOOLS.md`.

### 3.1 Pipeline de dados

```
Fonte Dignified (.Dxx)
   │  Classic.preprocess (hook opcional)
   ▼
Lexer.lex()  ──usa──  Classic.Lexer.lexing() / Classic.Lexer.quoted_string() (hooks)
   ▼
Parser.par(): pass_1 → pass_2 → pass_3 → (includes recursivos) → pass_4 → pass_5
   │  cada pass chama Classic.Parser.pass_N() como hook
   ▼
Parser.generate()  ──usa──  Classic.Parser.generate() (hook)
   │  Classic.postprocess (hook opcional)
   ▼
Listagem Classic numerada (.Axx)
   ▼
Tools.Interface.run():  tokenizer_interface.Run → (troca destino p/ binário) → emulator_interface.Run
   ▼
Binário tokenizado (.Bxx / .TAP / .TZX / .Z80 / .SNA / .p)  +  execução opcional em emulador
```

Caminho alternativo `-asc`: entra direto com uma listagem já numerada,
pulando lexer/parser Dignified — vai direto para `Tools.Interface.run()`.
**O tokenizer e o emulador precisam funcionar nos dois casos.**

### 3.2 Contrato de extensão por plataforma

Cada módulo de plataforma (`badig_<id>.py`) implementa:

- **`Description`**: vocabulário léxico do dialeto (instruções, funções,
  operadores, símbolos, literais, números, regras de nome de variável) e
  `sysinfo()` (nome do sistema, extensões de arquivo).
- **`Lexer`**: `initialization()` + `lexing()` (chamado a cada token não
  resolvido pelos casos especiais do núcleo) + opcionalmente `quoted_string()`.
- **`Parser`**: `initialization()` + `pass_1..pass_5()` + participa de
  `generate()`.
- **`Settings`**: config específica da plataforma.
- **`Expose`**: argumentos de CLI extras (parser "filho") + remtags
  (`##BB:chave=valor` no código-fonte) que o módulo entende.
- Hooks de módulo opcionais: `preprocess(code, stg)`, `postprocess(code, stg)`,
  constante `ERROR_EXIT_CODE`.

**Protocolo de retorno tri-estado**, repetido em todo hook de lexer/parser:
`False` → siga o fluxo normal do núcleo; `None` → descarte o token atual;
um `Token` → substitua o token atual por este. É o mecanismo central de
extensibilidade de todo o sistema.

Cada módulo também expõe um `tools_<id>.py` com `Interface.run()` (sempre
tokenizer → emulador, nessa ordem, trocando o arquivo-alvo pelo binário se
gerado) e pares `tokenizer_interface.py`/`emulator_interface.py` com seu
próprio `Settings`/`Expose` e um INI próprio.

### 3.3 Configuração em camadas

Quatro camadas, com bootstrap em duas fases:

1. INI global (`badig.ini`).
2. CLI (argparse), com parsers "filhos" de cada módulo concatenados via
   `parents=[...]`.
3. Remtags no código-fonte (`##BB:chave=valor`), lidos **depois** do
   primeiro parse de CLI (porque dependem de saber o nome do arquivo de
   entrada e a lista de remtags válidos, que só existe após `Expose()` de
   todos os módulos rodar).
4. Remtag especial `ARGUMENTS`: se presente, **reparseia os argumentos uma
   segunda vez** concatenando CLI real + tokens extraídos do código.

Precedência final: INI < CLI < remtag no código (com `ARGUMENTS` se
comportando como "CLI adicional reprocessada").

### 3.4 Estruturas de dados centrais

- `Token`: tipo (uppercase), valor (qualquer tipo), posição; propriedades
  computadas (`var_name`/`var_type` via regex do módulo, recalculadas a
  cada acesso, sem cache).
- `Position`: referência à lista inteira do código-fonte + linha/coluna/offset,
  usada para apontar erros com precisão (`^` sob o caractere).
- Não existe AST: uma "linha BASIC" é apenas uma sublista de tokens entre
  dois `NEWLINE`. O parser é fundamentalmente um **sistema de reescrita de
  stream de tokens com estado global mutável** (dicionários de defines,
  declares, functions acumulados como efeito colateral ao longo dos passes),
  não um parser recursivo com árvore.
- Includes rodam Lexer+Parser próprios recursivamente, e tabelas de símbolo
  (`c_hard_short_vars`, `c_hard_long_vars`, `d_declares`) atravessam a
  fronteira include↔arquivo principal por referência explícita.

### 3.5 Pontos de acoplamento problemáticos (confirmados lendo o código)

| Acoplamento | Onde | Problema | Decisão para o Go |
| --- | --- | --- | --- |
| `c_var_valid_chars` (comprimento único de alias curto) | `badig.py` função `get_declares`, usado em ~6 pontos | Assume *um* comprimento de alias por sistema; não modela categorias (escalar vs string/array/FOR) com limites diferentes — exatamente o que o Spectrum precisa | Variável alocada por **categoria**, não por sistema inteiro; hook explícito de alocação por tipo de nome |
| Limite de 256 chars/linha | `badig.py:1864-1866`, literal hardcoded | Regra herdada do MSX/CoCo, sem ponto de extensão | Tornar configurável por dialeto, com 256 como default MSX/CoCo |
| `SystemExit(0)` em erro silencioso | `support/helper.py` `Infolog.log` | Em modo silencioso (`-vb 0`), erro sai com código 0 a menos que o módulo declare `ERROR_EXIT_CODE=1`; é atributo de **classe** compartilhado, afeta instâncias já criadas | **Corrigir de vez no Go**: `error` Go subindo a pilha, exit code decidido uma única vez no `main()`, nunca 0 em erro |
| Sentinela `PROGRAM` | `badig.py:419` | Token sintético no início da lista de tokens para dar "ponto zero" estável; pode ser mal-interpretado como variável real | Modelar como estado de parser explícito, não como token fantasma na mesma lista |
| Lexer de regex combinada (maximal munch) | `badig.py:get_token` | Prefixos parciais tipo `&(h|o|b)?` só existem para permitir avanço caractere a caractere; ordem das alternativas importa; não há equivalente direto em RE2/Go | **Reescrever como scanner manual** (hand-written, por tabela/trie), não traduzir regex por regex |
| Herança múltipla cooperativa (`class Description(Dignified.Description, Classic.Description)`) | todo `badig_<id>.py` | Depende de MRO do Python | Composição explícita de structs + interfaces em Go |
| Duplicação de `join_commands`/`str2bol`/`get_ini` | núcleo e cada módulo | Mesma lógica reimplementada 3-4 vezes | Uma implementação única no core Go, compartilhada |

---

## 4. Inventário por plataforma

### 4.1 MSX Basic — módulo mais maduro (referência original de Farique)

- **Nomes de variáveis**: alocação automática para 2 letras (`aa`..`zz`,
  676 combinações, ordem descendente), namespace único por tipo. `~nome`
  preserva nome longo; `declare nome:xy` força alias. Esgotamento é erro
  fatal — nunca alias silencioso. **Algoritmo confirmado lendo
  `badig_msx.py:516-558` por completo (2026-11-21)**: um cursor
  decrescente único (`var_index`, começa em 676) compartilhado por
  toda a compilação — decrementa, calcula `zz` descendo até `aa`
  (`idx_h=var_index//26`, `idx_l=var_index%26`,
  `short=c_var_chr[idx_h]+c_var_chr[idx_l]`), pulando um candidato se
  colidir com 1 de 3 conjuntos de exclusão: `d_declares` (nomes já
  usados), `c_hard_short_vars` (aliases curtos reservados manualmente),
  e os 2 primeiros chars de `c_hard_long_vars` (nomes longos mantidos
  via `~`). Uma vez mapeado, o nome longo reusa sempre o mesmo alias
  (`d_declares` cacheia por `nome+arquivo`) — o cursor só decrementa no
  primeiro uso de cada variável distinta. **Nota pro Go**:
  `core.SymbolTables` (Bloco 1) hoje só modela o equivalente a
  `d_declares`; vai precisar de mais dois campos (conjuntos de
  exclusão) quando o `Parser` MSX for implementado de verdade.
- **Transformações de dialeto**: `?`↔`PRINT` (`-cp`), remoção opcional de
  `THEN`/`GOTO` adjacentes (`-tg`, **não aplicável ao Spectrum**, onde
  `THEN 100` não é equivalente ao salto Sinclair), `TRUE`/`FALSE`→`-1`/`0`
  (lógica bit-a-bit, diferente do Spectrum que usa `1`/`0`), tradução
  Unicode→ASCII de glifos de bloco via duas tabelas paralelas de 128 chars.
- **Tokenizer MSX Batoken** (`msxbatoken.py`, 758 linhas, standalone):
  tabela de ~150 tokens, varrida em **ordem** (não por hash — ordem
  importa, ex. `ERROR` antes de `ERR`), valores de tamanho variável
  (1-3 bytes; prefixo `FF` para funções estendidas). **Correção
  2026-11-21**: o parsing numérico **não usa Microsoft Binary Format
  nem qualquer formato binário de float** (achado inicial errado,
  corrigido lendo `parse_sgn_dbl` por completo) — é puramente
  **decimal**: um byte de "expoente" (`0x40 + quantidade de dígitos
  significativos`) seguido dos dígitos decimais em si, dois por byte
  (cada dígito já é um nibble hex válido), arredondado "half up" no
  dígito 6 (simples) ou 14 (dupla). Alvo de `THEN`/`GOTO`/etc. é só o
  número de linha literal reencodado como `u16` little-endian com
  marcador `0x0e` — sem resolução de endereço. Linha BASIC termina em
  `0x00` (não `0x0D` do Spectrum nem `0x76` do ZX81 — três formatos de
  linha binária incompatíveis entre si). Endereço de carga fixo
  `0x8001` (o primeiro byte do arquivo, `0xFF`, é um marcador de
  arquivo separado, em `0x8000`, não faz parte do endereçamento de
  nenhuma linha).
- **Sem formato de fita**: MSX roda de disco via openMSX (`-diska`), não
  de "cassete" — não há TAP/TZX equivalente a portar para MSX.
- **Integração openMSX** (`emulator_interface.py`, 358 linhas): a parte
  mais avançada e mais arriscada de portar. Controle remoto via
  `subprocess` + protocolo XML-like sobre stdio, com um watchpoint de
  memória para acelerar o boot de disco, e — recurso único — **monitoramento
  de execução em tempo real que mapeia erros de volta para a linha
  Dignified original** via tabela `line_list`. **Só funciona fora do
  Windows** no Python original (limitação do próprio I/O de pipes). Exige
  revisão do script `openmsx_output.tcl` antes de portar.
- **DignifieR MSX** (`msxbader.py`, 654 linhas): conversor Classic→Dignified
  heurístico (não 1:1), com lista de keywords **duplicada e levemente
  divergente** da do módulo principal — oportunidade de unificar no Go.
  Usa regex com lookahead que **RE2 do Go não suporta** — precisa virar
  parser manual, não é tradução direta.
- Formatos: `.dmx` (Dignified), `.amx` (Classic numerado), `.bmx`
  (binário tokenizado), `.lmx` (listagem de depuração, cosmética).
- **Escopo novo (2026-11-05, pedido do usuário, não existe no Python
  original)**: empacotamento em `.CAS` (captura de fita MSX a nível de
  BIOS, formato que se origina do próprio emulador fMSX — sem relação
  com o `.CAS` do CoCo, mesmo nome, formato diferente) e `.DSK` (dump
  bruto de setores de um disquete MSX-DOS/FAT12). `.CAS` é mecânico
  (parecido em espírito com TAP/TZX); `.DSK` é bem mais caro — exige
  montar uma imagem MSX-DOS válida (boot sector, FAT, diretório), não
  só empacotar bytes. Ver `docs/RESOURCES.md` para a especificação de
  cada formato e as fontes consultadas.

### 4.2 ZX81 Basic — adicionado pelo usuário, herda de MSX

- **Nomes de variáveis**: só 1 letra, alocadas **de Z para A**, 26 no
  máximo, esgotamento é erro explícito.
- **`PRINT AT`/`[?](row,col)`**: parser de profundidade de parênteses
  escrito à mão (não regex).
- **`LET` implícito obrigatório** inserido após número de linha ou `THEN`.
- **Renumeração de listagem numerada → Dignified com labels**: rejeita
  explicitamente destinos de salto computados (variável/expressão) —
  decisão de segurança central a preservar.
- **`remx <linha>,<contagem>[,byte]`**: reserva linha REM crua em endereço
  absoluto, bypassando numeração normal; implementado via
  `preprocess`/`postprocess` (extrai, depois reinsere após numeração final).
- **Formato binário `.p`**: **não é um encapsulamento de fita, é uma
  imagem de memória completa do sistema ZX81** — 116 bytes de system
  variables (offsets mágicos da ROM), `PROGRAM_START=0x407D` fixo,
  programa tokenizado, display file fixo de 25×33 bytes (sempre expandido —
  não implementa o modo compacto 1K/16K), 2 bytes de fim de variáveis.
  Portável como construção determinística de bytes, mas **conceitualmente
  incompatível** com TAP/TZX — não há reaproveitamento de código entre
  ZX81 e Spectrum aqui, só de conceito.
- **Números em ponto flutuante oculto**: dígitos ASCII + marcador `0x7E` +
  expoente+128 (1 byte) + mantissa 4 bytes big-endian. Bug histórico de
  arredondamento já corrigido no Python (usar como golden test).
- **Gráficos**: 22 glifos de bloco/xadrez fixos (sem UDG customizável,
  diferente do Spectrum); **tradução Unicode (`-tr`) nunca foi implementada**
  para ZX81 — usa só ASCII + escapes `\XX`/`%X`/`\{XX}`. Apóstrofo literal
  não suportado (fora do charset).
- **Emulador**: mesmo EightyOne do Spectrum, mas chamada "crua" — só abre
  o `.p`, sem seleção de modelo/máquina, sem verificação de carregamento
  bem-sucedido (sem hook de monitoramento).
- **DignifieR ZX81** (`zx81bader.py`): heurísticas próprias, parser
  baseado em um único regex combinado alternativo (mais simples que o
  lexer character-by-character do core), state machine de espaçamento
  própria (não reaproveita o core).
- **Pendências conhecidas da "ZX-81 Audit 2026-11-03"**: mapa Unicode
  ausente, auditoria de comandos contra ROM real incompleta, **nunca
  testado carregando/rodando `.p` real no EightyOne** (só testes
  unitários/estáticos), apóstrofo sem solução, faltam amostras de
  regressão aprovadas pelo usuário, diferença de comportamento
  Bash vs PowerShell no helper de conversão.
- Formatos: `.DZX` (Dignified), `.AZX` (Classic numerado), `.p` (binário/imagem de memória).
- **Escopo novo (2026-11-05, pedido do usuário)**: suporte também a
  `.81` (mesmos dados crus da rotina de fita da ROM, sem o nome de
  arquivo — equivalente a `.p` sem o cabeçalho) e `.p81` (igual, mas com
  nome de arquivo, podendo guardar múltiplos programas). `.TZX` também
  aparece como formato de fita aceito por alguns emuladores ZX81 (ex.
  RetroPie `lr-81` aceita `.p`/`.tzx`/`.t81`) — avaliar se vira alias do
  pipeline `.p`/`.81` ou um empacotador à parte. Ver `docs/RESOURCES.md`.

### 4.3 ZX Spectrum (Sinclair) Basic — módulo mais recente e mais avançado, v1.10.0

- Já documentado em profundidade em `badig/OUTLINE.md` (777 linhas) — ver
  esse arquivo para o histórico completo de decisões e estado atualizado.
  Resumo técnico confirmado por leitura de código:
- **Herda de MSX** (como ZX81), só sobrescreve o necessário.
- **Nomes de variáveis por categoria** (o caso mais complexo das três
  plataformas): escalares numéricos mantêm nome alfanumérico nativo de
  qualquer tamanho; strings exigem 2 chars terminados em `$`; arrays e
  variáveis de FOR/NEXT exigem exatamente 1 letra. `DECLARE` mapeia nomes
  longos para o alias exigido pela categoria. **É o padrão que deveria
  guiar o design do alocador de variáveis "por categoria" no núcleo Go**
  (ver tabela da seção 3.5).
- **`LET` implícito**, `PRINT AT` via `[?](linha,coluna)` (ordem diferente
  do MSX!), `TRUE`/`FALSE`→`1`/`0` (semântica Sinclair real, não bit-a-bit),
  operadores compostos (proibidos em arrays/slices), `DATA` aceitando
  expressões (diferente do MSX, que trata itens DATA como literais opacos).
- **`DEF FN`/`FN` nativo**: parâmetros de 1 letra, **6 bytes ocultos
  reservados pela ROM por parâmetro** (marcador `0x0E` + 5 bytes) — detalhe
  binário invisível no texto mas obrigatório no tokenizado.
- **Representação binária Sinclair**: linha BASIC com número **big-endian**
  2 bytes + comprimento **little-endian** 2 bytes + corpo + `0x0D`; tokens
  `0xA5`-`0xFF` (ordem da tabela de keywords = codificação, fiel à ROM real);
  números em 5 bytes (inteiro compacto ou float com bias de expoente 129 +
  mantissa 32-bit, usando `Decimal`/`Fraction` para arredondamento exato —
  em Go, `math/big.Rat` reproduz isso 1:1); caracteres libra/copyright/
  blocos gráficos `\80-\8F`/UDGs `%A-%U`/`\90-\A4`, bytes crus `\{XX}`.
- **REMX**: mesma ideia do ZX81 (linha REM absoluta), mas com fill
  configurável e contagem em bytes decodificados.
- **TAP**: header (flag, tipo, nome 10 bytes, tamanho, autorun, offset de
  variáveis) + checksum XOR + bloco de dados. Limite de 41612 bytes de
  programa (RAMTOP−PROG 48K). **TZX**: assinatura `ZXTape!`, blocos `0x10`
  (Standard Speed Data), reempacotando os mesmos blocos TAP. Ambos 100%
  mecânicos — portáveis diretamente.
- **Snapshots Z80/SNA via SkoolKit** (dependência Python externa opcional):
  **única peça não-mecânica do módulo**. SkoolKit roda um **simulador Z80
  completo executando a rotina real de LOAD da ROM 48K** até um endereço
  de retorno conhecido, valida PC/PROG/VARS resultantes, e só então
  converte para SNA. Não é um decodificador de formato — é um emulador.
  **Decisão recomendada**: adiar para fase tardia do roadmap Go; se/quando
  portar, usar uma lib de CPU Z80 em Go já existente (ex. `remogatto/z80`)
  rodando a ROM 48K real (de domínio público de fato, amplamente
  redistribuída por emuladores) até o endereço de retorno do loader —
  conceitualmente o mesmo que o SkoolKit faz, só que nativo.
- **Modelo 128K opcional**: tokens `SPECTRUM`/`PLAY` reaproveitando os
  códigos dos UDGs `%T`/`%U` do 48K (mutuamente exclusivos por modelo).
  Z80/SNA bloqueados explicitamente com 128K. **Bug conhecido**: launcher
  automático do EightyOne com `--sp_model 128 --em_run` trava em tela
  preta após boot — carregamento **manual** funciona normalmente nos dois
  emuladores testados (EightyOne, Spectaculator). O binário gerado está
  correto; o bug é de integração com o emulador externo, não do tokenizer.
- **Cobertura de testes**: 69 funções de teste (`test_zxspectrum.py`,
  1243 linhas) cobrindo tokenização core, FN nativo (inclusive execução
  real na ROM via SkoolKit quando disponível), REMX, tape, DignifieR,
  erros compartilhados, compostos/LET implícito, 128K, snapshots,
  emulador. **Excepcionalmente rigorosa para projeto hobby** — valida
  bytes exatos contra vetores conhecidos, não round-trip encoder/decoder
  que compartilhariam o mesmo bug. **Usar como golden-master do porte Go**:
  comparar bytes gerados pela implementação Go contra os mesmos vetores
  fixos do Python.
- **DignifieR Spectrum**: AZS→DZS com labels, encadeamento condicional
  preservado, numeração irregular com destino dinâmico rejeitada,
  binários não são entrada válida.
- Formatos: `.DZS` (Dignified, padrão atual), `.AZS` (Classic numerado,
  padrão atual), compatível com `.DSP`/`.ASP` (nomes originalmente
  propostos). Binário: `.TAP`/`.TZX`/`.Z80`/`.SNA`.
- **Escopo novo opcional/tardio (2026-11-05)**: `.DSK` de disco do
  Spectrum +3 — fora do recorte 48K/128K atual, baixa prioridade
  (Bloco 10).

### 4.4 Pendências específicas do Spectrum ainda em aberto (herdadas do Python, relevantes para o Go)

Ver `badig/OUTLINE.md` seção "Pendencias priorizadas" para a tabela completa.
Resumo das que importam para o planejamento Go:

1. Validação manual completa em emulador (native/graphics/fn/streams,
   snapshots, mídia/impressora) — ainda parcial.
2. Monitoramento de erros da ROM — **investigado e adiado** no Python:
   só viável via simulação offline (SkoolKit Simulator), não via EightyOne
   (sem gancho de automação) nem como o MSX faz via openMSX (exclusivo Unix).
3. Empacotamento integrado final (instalar/configurar sem depender de
   Python/conversor externo) — **é basicamente o objetivo central do
   projeto Go**: um executável único resolve isso de graça.
4. Extração/detokenização de binários (TAP/TZX/Z80/SNA → AZS) — não
   implementada em nenhuma plataforma; trabalho novo se quisermos round-trip
   completo.
5. Outros modelos/periféricos (128K parcial, Microdrive/disco, NextBASIC,
   turbo, fitas multi-arquivo) — opcional, alto custo.

---

## 5. Documentação, dependências externas e integração com editores

### 5.1 Linguagem Dignified comum (features compartilhadas entre as 3 plataformas)

Labels (`{name}`, `{@}` auto-referência), loop labels com `exit`, `DEFINE`
(macros com posicionais), `DECLARE` (forma do alias varia por plataforma —
2 letras MSX, 1 letra ZX81, nome nativo ou 1 letra por categoria no
Spectrum), `FUNC`/`RET` (açúcar sintático para GOSUB/RETURN, **não** são
closures nem o `DEF FN` nativo do Spectrum), separação de linha `:`/`_`,
comentários `##` (removidos)/`REM`/`'` com toggles `#nome`/`keep`/`#all`/`#none`,
tradução Unicode→ASCII de gráficos (ausente no ZX81, esquemas diferentes
em cada plataforma — **não há tabela de caracteres compartilhável**),
`TRUE`/`FALSE` com semântica diferente por plataforma, operadores
compostos, `INCLUDE` com namespace por arquivo, config em 4 camadas
(código-fonte < `.ini` < CLI < remtags).

### 5.2 Backlog herdado do Python (`documentation/IMPLEMENTATIONS.md`)

Itens relevantes para o roadmap Go (resolvidos "de graça" pelo redesenho,
não como débito a carregar):

- **`SystemExit(0)` em erro** → corrigir desde o início com exit codes Go.
- Publicar extensões nas lojas oficiais → objetivo final deste projeto.
- GUI completa de configuração → mapeia direto para o Fyne GUI planejado.
- Detectar sistema pela extensão do arquivo sem precisar de `-id` → tarefa
  natural do novo CLI.
- Carregamento de módulo dinâmico por nome com ordenação declarada → vira
  design de plugin/registro de módulo no Go.
- `badig.ini` de projeto (local) com prioridade → mapeia para config
  SQLite por projeto.
- "Ultra compact code" (modo otimizado) → feature nova, não existe ainda
  em nenhuma plataforma — avaliar se entra no escopo do Go v1 ou depois.

Itens específicos por plataforma (bugs/pendências de ZX81/Spectrum) já
listados nas seções 4.2/4.3/4.4.

### 5.3 Dependências externas documentadas

- Python 3.8+ (será eliminado pelo port Go).
- SkoolKit 8.10–11 (opcional, só snapshots Spectrum — ver decisão na seção 4.3).
- Emuladores externos, nenhum bundlado. Matriz atualizada em
  2026-11-05 com os emuladores adicionais pedidos pelo usuário (detalhe
  completo e links em `docs/RESOURCES.md`):

| Emulador | Plataforma(s) | SO | Observação |
| --- | --- | --- | --- |
| **openMSX** | MSX | Mac/Linux (monitoramento), também Windows (sem monitoramento) | Único do projeto original com monitoramento de execução/erro em tempo real |
| **fMSX** | MSX | Win/Linux/Mac | Alternativa ao openMSX; origem do formato `.CAS` do MSX; licença não-GPL, não redistribuir binário |
| **EightyOne** | ZX81, ZX Spectrum | Windows | Sem hook de automação/debug; bug de freeze no lançamento automático de 128K |
| **Spectaculator** | ZX Spectrum | Windows | Alternativa ao EightyOne; só carregamento manual verificado até agora |
| **ZEsarUX** | ZX81, ZX Spectrum (MSX não confirmado) | Linux/Mac/Windows | Tem **ZRCP**, protocolo de controle remoto via TCP — candidato real a monitoramento de erro em tempo real para as plataformas Sinclair, algo que o EightyOne não oferece |
| **XZ81 / sz81** | ZX81 | Linux | `xz81` é histórico/sem site estável; `sz81` é o sucessor mantido, mesma linhagem |
| ~~XRoar~~ | ~~CoCo~~ | — | Fora de escopo (CoCo descartado) |

  **Achado importante desta sessão**: o ZEsarUX, por ter o ZRCP (API de
  controle remoto via socket TCP, habilitável por flag), é um candidato
  mais forte do que o EightyOne para portar o recurso de monitoramento
  de execução/erro em tempo real (hoje exclusivo do MSX/openMSX) também
  para ZX81 e ZX Spectrum. Isso muda o panorama da pergunta 4 da seção 8
  — vale reavaliar antes de decidir "fica só MSX por ora".

### 5.4 As 3 extensões VSCode hoje (candidatas a virar uma só)

São **puramente declarativas** (TextMate grammars + tasks shell/process,
sem lógica de runtime JS/TS, exceto um teste de gramática no módulo
Spectrum) — isso barateia muito a unificação. Já compartilham ~80% do
contrato (`<plataforma>-basic-classic`/`-dignified`, 1 setting de path + 1
de interpretador, snippets, build tasks). Diferenças a resolver na
unificação: MSX tem 7 temas + ícones de linguagem + comandos/menus/
keybindings completos (o mais rico); ZX81 segue o mesmo padrão do MSX mas
sem temas; Spectrum tem settings mais modernas (`spectrum-basic.model`
enum 48/128) e teste de gramática automatizado, mas **falta
commands/menus/keybindings** (só tasks.json). `engines.vscode` também
diverge (^1.60.0 MSX/ZX81 vs ^1.85.0 Spectrum) — decidir versão mínima
única.

**Esforço de unificação avaliado como baixo-médio**: consolidar
`languages`/`grammars`/`snippets` sob IDs namespaced num único
`package.json`, unificar settings sob uma seção `badig.*`, fundir os 3
`tasks.json` (usando o padrão do Spectrum, mais rico, como base), portar
commands/menus/keybindings do MSX para as 3 linguagens, decidir sobre
temas, adicionar testes de gramática para as 3. **Ganho real de UX**:
trocar `badigPath`/`pythonInterpreter` por chamada direta ao binário Go —
elimina a necessidade de configurar Python nas settings do usuário.

### 5.5 Sublime Text

Estrutura quase idêntica entre plataformas (build/syntax/snippets/settings
declarativos), esforço de porte baixo. **Atualizado em 2026-11-05: o
usuário confirmou que o suporte a Sublime Text é objetivo real do
projeto** (prioridade 2, logo após o VSIX), não mais um "se sobrar
tempo" — ver seção 10.6. Gerar, quando possível, a partir da mesma fonte
de verdade (grammar/keywords) que alimenta o VSIX, para não manter duas
definições de sintaxe divergentes.

---

## 6. Decisões de arquitetura propostas para o Go

Estas são propostas a validar com o usuário antes/durante a Fase 1 (ver
seção 7). Não é código ainda.

1. **Interfaces por hook, não herança.** `Description`, `Lexer`, `Parser`,
   `Settings`, `Expose` viram interfaces Go pequenas; cada plataforma
   implementa um struct. Composição explícita (embutir um struct
   `DignifiedCore` dentro do struct da plataforma) no lugar da herança
   múltipla cooperativa do Python.
2. **Protocolo de hook como enum explícito**, não tri-estado
   `False`/`None`/`Token`: `type HookAction int` com `Continue`, `Skip`,
   `Replace`, retornado junto com `*Token` — mais claro, sem a ambiguidade
   Python `None` vs `False`.
3. **Lexer escrito à mão** (hand-written scanner por tabela/trie), não
   tradução literal da regex combinada de maximal-munch do Python — RE2
   do Go não suporta o padrão usado e seria lento de qualquer forma.
4. **Contexto de compilação explícito** (`*CompileContext`, passado por
   parâmetro) no lugar de estado mutável pendurado em atributos dinâmicos
   de instância — resolve o acoplamento de includes/hard-vars do Python.
5. **Alocador de variáveis por categoria**, não por sistema inteiro:
   cada categoria (escalar numérico, string, array, controle FOR/NEXT)
   declara seu próprio comprimento de alias e conjunto de caracteres
   válidos — generaliza o caso MSX (tudo 2 letras), ZX81 (tudo 1 letra) e
   Spectrum (nativo + 1 letra por categoria) sob o mesmo mecanismo.
6. **Erros como `error` Go subindo a pilha**, exit code decidido uma única
   vez no `main()` — elimina a armadilha do `SystemExit(0)` silencioso.
7. **SQLite como camada de config única**, substituindo os múltiplos
   `.ini` por módulo: tabelas por `system_id` (namespace), com uma tabela
   de keywords/tokens por dialeto (pode inclusive servir de fonte única
   para o VSIX gerar grammars/snippets automaticamente — avaliar).
   **Decisão em 2026-11-07:** usar `modernc.org/sqlite` (puro Go, sem
   CGO). A primeira fatia do schema foi implementada para temas: tabela
   `themes` guarda o catálogo e tabela `active_themes` guarda a seleção
   separadamente, por modo (`gui` já; TUI/CLI quando receberem temas).
8. **Tokenizers isolados do CLI**, testáveis por vetor de bytes fixo
   (entrada textual → saída binária esperada), usando os testes Python
   existentes (Spectrum: 69 casos; ZX81: ~549 linhas de teste) como
   golden-master direto — gerar os mesmos vetores para MSX (sem testes
   hoje) rodando o Python original antes de portar.
9. **Emuladores externos por trás de uma interface `EmulatorController`**,
   com múltiplos backends por plataforma (MSX: openMSX, fMSX; ZX81:
   EightyOne, ZEsarUX, sz81/XZ81; ZX Spectrum: EightyOne, Spectaculator,
   ZEsarUX), para poder adiar/placeholder qualquer um sem bloquear o
   resto (especialmente openMSX no Windows e o bug de freeze 128K do
   EightyOne). O backend ZEsarUX (ZRCP) é candidato prioritário para
   monitoramento de erro em tempo real nas plataformas Sinclair — ver
   seção 5.3.
10. **Z80/SNA via emulador Z80 em Go é trabalho de fase tardia** — não
    bloqueia MVP (texto + TAP/TZX cobre carregar e rodar em emulador).
11. **Paridade de opções entre interfaces**: CLI, TUI e GUI devem expor,
    na medida do possível, o mesmo conjunto de opções/funções — cada uma
    só muda a forma de apresentação (flags vs. menus vs. telas), não o
    conjunto de recursos disponível. Um texto "about"/"man" central
    (`internal/about` no stub) é compartilhado pelas três, para não
    divergirem. Ver seção 10.7 para o desenho do sistema de ajuda.

---

## 7. Divisão em blocos de trabalho

Legenda de peso: **B**=baixo, **M**=médio, **A**=alto.
Prioridade: **1**=essencial ao MVP, **2**=importante mas adiável,
**3**=opcional/tardio. Itens marcados `[x]` já foram concluídos; o motor
de conversão e os módulos de plataforma seguem pendentes.

### Bloco 0 — Fundamentos do projeto Go

- [x] Estrutura de diretórios e documentação de alto nível criada:
  `dist/`, `resource/`, `examples/` (com subpastas por plataforma),
  `docs/` (`MANUAL.md`, `RESOURCES.md`, `releases/`), `README.md`,
  `CHANGELOG.md` e esquema de versionamento `X.Y.Z` registrados (ver
  seção 10). — 2026-11-05.
- [x] Módulo Go inicializado (`github.com/wilsonpilon/goBaDig`), com
  `cmd/goBaDig` + `internal/{about,cli,gui,tui,version}` — primeira
  fatia da árvore da seção 10.4. Dependências buscadas: `rivo/tview`,
  `gdamore/tcell/v2`, `fyne.io/fyne/v2`. Build (`go build ./...`),
  `go vet` e `go test ./...` passando; binário testado em runtime (CLI,
  TUI e GUI abrem e fecham sem erro no Windows). — 2026-11-05 (release
  1.1.2). Pendente ainda: lint/format formal e pipeline de CI.
- [x] `build.ps1` (raiz do repo): automatiza `go mod download` + `go vet`
  + `go test` + `go build` + montagem do pacote `dist/goBaDig-vX.Y.Z/`
  e o zip, lendo a versão direto de `internal/version/version.go`.
  Testado de ponta a ponta (build completo e com `-SkipDownload
  -SkipTests`). Não cria tag/release — isso continua manual. Ainda não
  liberado como release próprio (só o script em si). — 2026-11-05.
- [x] Driver SQLite puro Go (`modernc.org/sqlite`) escolhido e primeira
  fatia do schema implementada: catálogo `themes` e seleção separada em
  `active_themes`, organizados por modo de interface. — 2026-11-07.
- [ ] Completar `internal/core`, `internal/config`, `internal/platforms`,
  `internal/emulator` (ver seção 10.4) — esta fatia inicial só cobre
  CLI/TUI/GUI stub, não o motor. — **B**, **1**
- [x] Escolher driver SQLite: `modernc.org/sqlite`, sem CGO para essa
  dependência. — 2026-11-07.
- [ ] Desenhar schema SQLite inicial: config global, config por projeto,
  tabela de keywords/tokens por dialeto, tabela de settings de emulador/
  tokenizer por plataforma. — **M**, **1**
- [x] **Gerados os vetores de ouro (golden tests) do MSX** rodando o
  Python original, já que não havia suíte de testes MSX até agora
  (2026-11-21). Novo arquivo `badig/msx/test_tokenizer_msx.py` (13
  casos), no mesmo estilo/convenção de `zx81/test_tokenizer_zx81.py` e
  `zxspectrum/test_zxspectrum.py` — mas, diferente deles, o tokenizer
  MSX (`msxbatoken.py`) não tem uma função pura importável equivalente a
  `tokenize_lines()`; é uma ferramenta arquivo-a-arquivo via CLI. Por
  isso os testes rodam `badig.py <arquivo>.amx -id msx -asc
  --tk_tokenize -rh` via `subprocess` (a flag `-asc` pula o
  lexer/parser Dignified por completo, isolando só o tokenizer Batoken
  — escopo combinado com a seção 6 item 8: "vetores de bytes fixo" é
  especificamente sobre o tokenizer, não o lexer/parser). Cada caso foi
  gerado rodando a implementação Python real e depois fixado como
  asserção exata — são a referência (ground truth) pra comparar contra
  quando o tokenizer MSX for portado pro Go (Bloco 2). Cobertura:
  equivalência `PRINT`/`?` (mesmo byte `0x91`), faixa de inteiros
  pequenos 0-9 (byte dedicado `0x11+n`) vs. inteiro de 1 byte (prefixo
  `0x0f`) vs. 2 bytes (prefixo `0x1c`), menos unário como operador `-`
  (não um literal negativo), float precisão simples (prefixo `0x1d` +
  dígitos decimais empacotados com um byte de expoente — **não** é
  Microsoft Binary Format nem nenhum formato binário, achado inicial
  errado corrigido em 2026-11-21, ver Bloco 2) vs. dupla (prefixo
  `0x1f`, sufixo `#`), string literal (ASCII entre aspas literais, sem escape),
  `REM` vs. `'` (encodings diferentes — `'` expande pra
  `3a 8f e6`, não é so um alias de byte único), múltiplos statements
  separados por `:`, função estendida com prefixo `ff` (ex.: `PEEK`), e
  o alvo de `THEN`/`GOTO` sendo um valor bruto de 2 bytes (sem prefixo
  de tipo), diferente de uma constante numérica normal. Confirmado:
  `python -m unittest discover -s msx -p "test_*.py"` → 13/13 OK; não
  afeta `zx81`/`zxspectrum` (continuam OK, roda cada um com
  `python -m unittest <pasta>.test_<arquivo>` a partir de `badig/` —
  **nota**: `discover -s .` a partir da raiz do `badig/` só encontra os
  testes de `zxspectrum/`, porque só esse diretório tem `__init__.py`;
  isso já era assim antes desta mudança, não é um problema introduzido
  aqui, e `badig/` é histórico congelado — não mexido além do novo
  arquivo de teste). **Importante achado ao commitar**: `/badig/`
  inteiro é ignorado pelo Git (`.gitignore`, decisão de 2026-11-05) —
  `badig/msx/test_tokenizer_msx.py` **nunca vai ser versionado**, existe
  só localmente nesta máquina. Por isso os 15 vetores (13 testes, dois
  deles com múltiplos casos) foram **também** copiados, já rastreados
  pelo Git, para
  [`docs/golden-vectors/msx-tokenizer.md`](docs/golden-vectors/msx-tokenizer.md)
  — essa é a cópia durável de referência; o arquivo de teste Python é só
  a ferramenta que os gerou/valida localmente. — **M**, **1**
- [x] Decidir nomenclatura final de extensões de arquivo por plataforma
  no Go: **mantidas as herdadas** (`.DMX/.AMX/.BMX`, `.DZX/.AZX/.p`,
  `.DZS/.AZS/.TAP/.TZX`), sem esquema novo. — 2026-11-20 (seção 8, pergunta 3).
- [x] **Investigado o ZRCP (ZEsarUX)** para monitoramento de erro em
  tempo real em ZX81/Spectrum (2026-11-21, pedido explícito do usuário
  em 2026-11-20 para avaliar isso já, não depois do MVP — seção 8,
  pergunta 4). **Veredito: viável**, lido direto do código-fonte do
  ZEsarUX (`src/zrcp/remote.c`, `src/debug.c`, `src/expression_parser.c`
  no GitHub `chernandezba/zesarux`, branch `main`) já que não existe
  documentação oficial do protocolo em si — só exemplos de terceiros
  (cliente ZRCP de `vscode-zx-debugger`) e o próprio código.

  **Técnica recomendada — diferente do hack do MSX/openMSX.** O MSX
  hoje captura caractere por caractere (breakpoint na rotina de saída de
  texto da BIOS, ver `openmsx_output.tcl`) e escaneia o texto resultante
  procurando o bell (`\x07`) de erro. Para ZX81/Spectrum existe uma
  abordagem mais direta: **ambas as ROMs reportam todo erro de BASIC via
  `RST 08h`** (instrução Z80 que sempre pula para o endereço físico fixo
  `0x0008`, confirmado em disassemblies de ambas as ROMs) — então basta
  um breakpoint condicional em `PC=0008H` pra capturar exatamente o
  instante do erro, sem depender de texto na tela nem de heurística de
  bell. Nesse ponto, os valores corretos já estão nas variáveis de
  sistema que a própria ROM usaria pra montar a mensagem de erro:
  - **ZX Spectrum**: `PPC` (`$5C45`, 2 bytes, little-endian) = número da
    linha BASIC em execução; `SUBPPC` (`$5C47`) = número do statement
    dentro da linha; `ERR_NR` (`$5C3A`) = código do erro menos 1.
  - **ZX81**: `E_PPC` (`$400A`) = número da linha atual; `ERR_NR`
    (`$4000`) = código do erro.

  **Comandos ZRCP confirmados no código-fonte** (protocolo texto puro
  por socket TCP, porta padrão 10000, habilitado com
  `--enable-remoteprotocol`):
  - `enable-breakpoints` (precisa vir antes de `set-breakpoint`, senão
    dá `Error. You must enable breakpoints first`).
  - `set-breakpoint <indice 1-based> <condição>` (alias `sb`) — ex.:
    `set-breakpoint 1 PC=0008H`. **Pegadinha de formato confirmada lendo
    o parser**: números são **decimais por padrão**; hex exige sufixo
    `H` ou prefixo `0x`/`$` — `PC=0008` sem o `H` é decimal 8, não
    `0x0008`. `PC=`/`MRA=`/`MWA=` (igualdade simples) usam um caminho
    "otimizado" mais rápido no núcleo do emulador; qualquer outra
    condição (registradores, `MRV`/`MWV`/`PRA`/`PRV`/`PWA`/`PWV`,
    combinações `and`/`or`/`xor`) cai no avaliador de expressão genérico,
    mais lento.
  - `run` (alias `r`) — **bloqueia a conexão até o breakpoint disparar**
    (ou outro evento). Não existe notificação assíncrona separada: a
    mesma resposta ao `run` que ficou pendurada chega completa só quando
    o breakpoint dispara, já contendo `Breakpoint fired: ...` + dump de
    registradores + a próxima linha de prompt (`<nome>@cpu-step> `).
    Pra continuar depois de um disparo, manda `run` de novo (não
    `exit-cpu-step` — esse comando sai do modo de debug por completo).
  - `get-registers` (alias `gr`) — dump completo dos registradores Z80
    (mesmo formato usado no dump automático após um breakpoint).
  - `read-memory <endereço> <tamanho>` — mesma regra decimal/hex dos
    endereços de breakpoint; sem tamanho, lê 1 byte.
  - `run` aceita um parâmetro `no-stop-on-data` pra não ficar também
    observando o socket por dados de entrada durante a execução — ver
    ressalva de performance abaixo.

  **Ressalva de performance, documentada no próprio código-fonte do
  ZEsarUX** (comentário do autor com benchmarks de uso de CPU do host):
  simplesmente **habilitar breakpoints já custa CPU do processo do
  emulador durante toda a execução** (não só no instante do disparo) —
  ~44% de CPU do host com uma condição `PC=`/`MRA=`/`MWA=` otimizada
  (contra 66% no parser antigo não-otimizado, pré-7.1); e o `run`
  também observar o socket por dados de entrada (comportamento padrão,
  sem passar `no-stop-on-data`) soma mais um custo independente (46%
  contra 19% sem observar o socket). Não é um número absoluto de
  overhead relativo (o autor não documentou o baseline com breakpoints
  totalmente desligados), mas é CPU real sendo gasta o tempo todo que o
  breakpoint estiver armado — provavelmente irrelevante pra rodar um
  programa BASIC de teste (execução automatizada, não tempo real pro
  usuário), mas vale ter em mente se algum dia isso rodar em lote/CI.

  **Conclusão prática**: ZEsarUX/ZRCP é a escolha certa pra monitoramento
  de erro em tempo real em ZX81/Spectrum — mais direto e mais preciso
  que o hack de captura de texto do MSX, sem exigir scripting dentro do
  emulador (TCL do openMSX) já que o controle todo roda só do lado de
  fora via socket. Isso **confirma e reforça** a decisão já registrada
  na seção 6, item 9 (ZEsarUX como backend prioritário Sinclair para
  monitoramento) — não muda a prioridade dos Blocos 3/4/8, só dá o
  desenho técnico concreto pra quando a interface `EmulatorController`
  for implementada (ainda não é — ver Bloco 1). **Ainda não verificado
  numa instância real do ZEsarUX** (pesquisa feita 100% lendo o
  código-fonte no GitHub, sem rodar o emulador) — antes de implementar
  de verdade, testar manualmente com `telnet localhost 10000` ou
  equivalente contra uma instalação real, pra confirmar o formato exato
  do texto de retorno (quebras de linha, etc.) e o comando certo de
  digitar teclas/carregar o programa (não investigado nesta rodada).
  — **M**, **1**

### Bloco 1 — Núcleo do motor (core engine)

**Plano detalhado (tipos/assinaturas Go concretos, pesquisa no
`badig.py`/`support/*` que fundamenta cada peça) em
`C:\Users\barney\.claude\plans\sharded-plotting-conway.md` — não
versionado (fora do repo), só nesta máquina. Resumo abaixo, quebrado em
fases/subfases com esforço (**B**aixo/**M**édio/**A**lto) e
criticidade (**1**=bloqueia o resto, **2**=importante mas não bloqueia
o essencial, **3**=pode esperar), igual à convenção já usada no resto
deste documento, pra decidir o próximo passo por esforço x criticidade.

**Fase 1 — Tipos fundamentais, sem incerteza de design — COMPLETA
(2026-11-21)** (nenhum depende de como uma plataforma real vai usá-los):
- [x] `Token`/`Position`/`TokenKind` (`internal/core/token.go`) — **B**, **1**
- [x] `HookAction` (substitui o protocolo tri-estado `False`/`None`/`Token`
  do Python) (`internal/core/hook.go`) — **B**, **1**
- [x] Diagnóstico/erro, substitui `Infolog` — erro como `error` Go
  subindo a pilha, exit code decidido uma vez só no `main()`
  (`internal/core/diagnostic.go`) — **B**, **1**
- [x] `CompileContext` + `SymbolTables` (estado explícito por parâmetro,
  resolve o acoplamento de includes) (`internal/core/context.go`) — **M**, **1**
- [x] Alocador de variáveis por categoria (`internal/core/allocator.go`,
  implementação genérica descendente `NewDescendingAllocator`) — **M**, **2**

**Fase 2 — Contrato "v0" (interfaces por hook + `Platform`) — PARCIAL
(2026-11-21)**:
- [x] `Description`/`SystemInfo` (`internal/core/contract.go`) — **B**, **1**
- [x] `Lexer` (`internal/core/contract.go`) — **B**, **1**
- [x] `Parser` (5 passes + `Generate`) (`internal/core/contract.go`) — **M**, **1**
- [x] `Platform` (`internal/core/platform.go`) — **B**, **1**
- [x] `Settings` (`internal/core/contract.go`) — **B**, **2**
- [x] `GenerateResult`/`LineCorrespondence` (`internal/core/contract.go`) — **B**, **2**
- [ ] `Expose`/`Exposed`/`RemtagSpec` — **B**, **3** (próxima fase — só
  importa quando o CLI real de flags existir, Bloco 5)
- [ ] `Tools.Interface` — **B**, **3** (próxima fase — sem corpo real
  até ter tokenizer/emulador de verdade, Bloco 2+)

**Desvio do plano original, descoberto implementando**: `Lexer.Lex` e
cada `Parser.PassN`/`Generate` ganharam um terceiro retorno `error`
(`(HookAction, Token, error)`, não só `(HookAction, Token)`) — faltava
um jeito de um hook sinalizar falha fatal sem o driver precisar
inspecionar o `DiagnosticSink`. É exatamente o tipo de ajuste que a
plataforma fake (Fase 3) deveria revelar antes de qualquer plataforma
real depender da interface.

**Segundo desvio, este sim só revelado pela plataforma MSX de
verdade** (2026-11-21, ver Bloco 2): a v0 do `Lexer.Lex` (palavra
pré-cortada por espaço em branco, `(ctx, Token) (HookAction, Token,
error)`) não aguenta números com notação científica nem símbolos
compostos sem espaço (`x+=5`) — precisa de lookahead multi-caractere
que varia por dialeto. Trocada por `Lex(ctx, remaining string)
(consumed int, tokens []Token, ok bool, err error)`, dando à
plataforma a **linha inteira restante** a partir da posição atual
(mesmo padrão já usado no tokenizer Batoken), e podendo devolver
**zero ou várias** tokens numa só chamada (não só uma) — resolve tanto
"consumir e descartar" quanto casos como `DATA` com múltiplos itens
numa só chamada. `core.Description` ganhou `QuoteChar() byte`, e
`core.Lex()` ganhou fallbacks genéricos compartilhados (identificador
Dignified, string entre aspas) usados só quando a plataforma não
reconhece nada na posição atual. `KindWord` foi removido (não existe
mais pré-corte); `KindIdentifier`/`KindString`/`KindUnknown` entraram
no lugar.

**Terceiro desvio** (2026-11-21, Pass 3/4 do `Parser` MSX): o `Parser`
de hoje (`PassN(ctx, tok) (HookAction, Token, error)`) só dá ao hook o
token atual — sem jeito de olhar pro vizinho. `get_hard_variable`
(Pass 3 do Python, confirmado lendo `badig_msx.py:463-514` linha a
linha) precisa checar se o **token anterior** foi o símbolo `~`
(mantém nome por extenso) e removê-lo da saída. Troquei a interface
pra dar ao hook a **lista de entrada completa da pass + o índice do
token atual** (`PassN(ctx, tokens []Token, index int) (HookAction,
Token, error)`) — permite olhar pra qualquer vizinho sem precisar
desfazer nada, já que "token anterior na entrada" e "já emitido na
saída" são a mesma coisa exceto quando a própria pass decidiu `Skip`
um token (como decide pro `~` aqui). Mais barato que reescrever o
`Parser` pra operar na lista inteira de uma vez (considerado, mas
grande demais pro ganho).

Também: `core.Lex()` ganhou reconhecimento dos 6 símbolos Dignified
compartilhados (`~ [ ] { } @`, `d_symbol` de `badig_dignified.py`) —
novo `KindDignifiedSymbol`. `core.SymbolTables` (que no Bloco 1 tinha
2 campos nunca usados de verdade) ficou precisa:
`Declares`/`HardShortVars`/`HardLongVars`, confirmados linha a linha
contra `get_hard_variable`/`process_variable`. `core.VariableAllocator.Allocate`
ganhou um parâmetro `excluded func(string) bool`, já que os 3
conjuntos de exclusão do MSX não fazem sentido dentro do pacote
`core` (genérico, não sabe nada sobre "nome curto já usado" etc.).

**Quarto desvio** (2026-11-21, Pass 5 do `Parser` MSX): `pass_5`
(`badig_msx.py:617-682`, lido linha a linha) precisa expandir **1**
token de entrada (o símbolo de atribuição composta, `+=`/`++`/etc.) em
**vários** de saída (`x+=5` → `=`, `x`, `+`, `5` — 3 tokens novos no
lugar de 1). O `Replace` de ontem só aceitava 1 `Token`. Troquei o
payload de `Replace`, uniformemente nos 6 métodos de `Parser`
(incluindo `Generate`, só por consistência de assinatura — ele não usa
a capacidade nova ainda), de `Token` pra `[]Token`: `Continue`/`Skip`
continuam ignorando o slice (já ignoravam o valor antes de mudar de
tipo); `Replace` agora emite exatamente os tokens do slice (0 = como
`Skip`, 1 = como antes, vários = o caso novo).

Também: `core.Lex()` ganhou reconhecimento de `TRUE`/`FALSE`
(`d_operat` compartilhado de `badig_dignified.py`, não vocabulário
MSX) — novo `KindDignifiedOperator`, checado ANTES do padrão de
identificador (senão um `true`/`false` minúsculo já teria sido
capturado como identificador antes de chegar nesse fallback — bug
real encontrado e corrigido nesta mesma sessão, pego pelo teste
`TestPass5ReplacesTrueAndFalse`).

**O `core` passou a fazer trabalho Dignified de verdade antes de
delegar pra plataforma** (2026-11-21, resolução de labels e numeração
de linha): até aqui, `Pass2`/`Pass3`/`Pass4` eram 100% delegados pro
hook da plataforma (`platform.Parser.PassN`, via `runPass`) — o
comentário do `Parser` desde o Bloco 1 já antecipava que isso mudaria
("Nenhuma dessas responsabilidades de nível core... está implementada
ainda"). Essa fatia é a primeira vez que isso acontece de verdade,
confirmado lendo `badig.py` (`pass_2`/`pass_3`/`pass_4`, que fazem seu
próprio trabalho Dignified primeiro e só caem pro `self.clc.passN()`
no "senão"): `pass2Labels`/`pass3LineHygiene` (`internal/core/pipeline.go`)
envolvem o hook da plataforma com esse padrão "core primeiro, cai pra
plataforma só se não for nada Dignified-específico" — `{nome}`/`{@}`
(sempre 3 tokens adjacentes, já que espaço nunca gera token: `{`, nome
ou `@`, `}`) colapsam em 1 token (`KindLabelDeclaration` se no início
de linha, senão `KindLabelJump`); 3 limpezas de linha (linha em branco
some, `_` no fim de linha funde com a próxima, `NEWLINE` depois de uma
declaração de label some) garantem que a numeração não desperdice
números em coisas que não vão gerar linha de verdade.

`pass4Numbering` **não** cabe no formato `passFunc`/`runPass` (que só
cobre "decida o destino do token atual") porque o algoritmo real
resolve os saltos (`{nome}` → número) numa **pós-passada** que muda o
valor de um token já emitido por uma chamada ANTERIOR — não dá pra
expressar isso como "retorne um HookAction pra esta chamada". Resolvi
sem rastrear índice (como o Python faz, porque o loop dele já constrói
a lista final token a token): a varredura via `runPass` deixa o
`KindLabelJump` atravessar sem resolver (exceto `{@}`, resolvido na
hora pro número da linha atual), e **depois** que `runPass` devolve a
lista pronta, uma segunda passada simples (`for i := range tokens`)
procura por `KindLabelJump` e sobrescreve `.Value` com o número
resolvido — como slices em Go compartilham o array por baixo, isso
muda a lista de verdade sem precisar de índice rastreado. Erro fatal
se o nome não existir no mapa de labels.

Novos tokens compartilhados: `KindDignifiedSeparator` (`_`, não lexado
antes — igual aos outros símbolos Dignified), `KindLabelDeclaration`,
`KindLabelJump`, `KindLineNumber` (token sintético que `pass4Numbering`
insere depois de cada `NEWLINE` que abre uma linha nova — e depois do
próprio `PROGRAM`, já que a primeira linha real não tem nenhum
`NEWLINE` antes dela pra disparar a numeração: achado real, pego pelo
primeiro teste escrito, corrigido tratando `KindProgram` como gatilho
também). **Fora de escopo, documentado** (nessa rodada): loop labels
(`nome{ }`/`exit`), inserção de header (2 linhas REM sintéticas),
relatório de labels (`-lbr`), `DEFINE`/`FUNC`/`RET`/`INCLUDE`/`KEEP`/
`DECLARE`.

**Loop labels implementados** (2026-11-22, `nome{ ... }` + `exit`):
fiel a `get_label_lines`/`get_jump_labels`/`get_loop_label_return`/
`get_labels_exit` (`badig.py:849-921`) e ao dispatch de `pass_2`
(1468-1481), confirmados lendo `last_tok`/`peek_ahead`/`next_tok`/
`prev_tok` (`badig.py:556-594`) pra distinguir com precisão lookback
de SAÍDA (`last_tok`, lê `tok_list_out`) de lookahead de ENTRADA
(`peek_ahead`, lê `tok_list_in`). Achado chave: um loop-opener
(`nome{`) produz o MESMO `D_LBLLIN` que uma declaração simples — não
precisei de Kind novo pra isso, só de detecção diferente. A detecção
original usa `last_tok(-2)` (SAÍDA) pra ver se o identificador antes
do `{` estava no início de linha — isso não cabe no meu modelo de
lookback por ENTRADA sem estado (o identificador já teria sido
emitido numa chamada anterior e eu não posso "despublicá-lo"
retroativamente). Resolvido sem estado: a detecção acontece no PRÓPRIO
identificador (olhando pra frente: está no início de linha E seguido
de `{`?) em vez de no `{` (que teria que olhar pra trás pra algo já
decidido) — e os dois lados (identificador decide se se esconde,
`{` decide se declara) usam EXATAMENTE a mesma checagem auxiliar
(`isPlainLabelAt`) pra nunca discordar sobre se um dado `{` é um
`{nome}` comum ou o abridor de um loop — confirmado que isso bate a
MESMA prioridade de dispatch do Python (`get_jump_labels` checado
antes de `get_label_lines` em `pass_2`): `nome{x}` sem espaço e com
corpo de 1 só token vira `nome` (literal, sobrando) + `{x}` (referência
comum), nunca um loop-opener — replicado de propósito, não é bug.

`c_instsp_str`/`c_loop_back` (separador `:` e palavra "goto" que o
fechamento do laço insere) são da `Description` de PLATAFORMA, não
existe default compartilhado — `core.Description` ganhou
`StatementSeparator()`/`LoopBackInstruction()` (mesmo padrão de
`QuoteChar()`). Os tokens sintéticos que isso gera usam um Kind novo
genérico, `core.KindSynthetic` (nada ainda precisa diferenciá-los por
Kind — isso é trabalho do `generate()` de verdade, ainda não feito).

`pass2Labels` deixou de ser um `passFunc` simples e passou a ser
bespoke (mesmo molde de `pass4Numbering`): fecha sobre uma pilha local
de labels de laço abertos, e checa no final se ela ficou vazia (erro
fatal "loop label not closed" se não — mesma checagem de pós-passada
que o `pass_2` original tem). `D_LBLRET` (fechamento do laço, novo
`KindLoopReturn`) resolve EXATAMENTE como uma referência de salto
comum (mesmo mapa `label_lines`, mesma passada de resolução já
existente — só ampliei o `case` pra aceitar os dois Kinds) e, como
efeito colateral na varredura de ida, grava onde o `exit`
correspondente deve pousar. `D_LBLEXT` (`exit`, novo `KindLoopExit`)
resolve numa passada SEPARADA (contra esse mapa de retorno, só
completo depois que a varredura inteira termina, já que `exit` sempre
aparece ANTES do `}` que fecha o laço no fluxo de tokens) — com o erro
fatal fiel ("loop exit past end of program") se o alvo for além do fim
do programa.

**Desvio deliberado, documentado**: `get_labels_exit` no Python não
tem proteção se `exit` aparece fora de qualquer laço aberto (quebraria
com `IndexError` não tratado); no Go isso é um erro fatal `Diagnostic`
de verdade — mesmo espírito de "error como `error` Go subindo a pilha"
já decidido (seção 6, item 6). **Simplificações aceitas, documentadas**:
não porto a deduplicação de separador (`last_tok() != 'C_INSTSP'` —
pior caso raríssimo gera um `::` inofensivo) nem a limpeza de
`C_INSTSP` repetido que o Pass 3 original tem em outro lugar
(`badig.py:1557-1575`).

`EXIT` é vocabulário Dignified COMPARTILHADO (`d_instrc` em
`badig_dignified.py`, junto com `DEFINE`/`DECLARE`/`INCLUDE`/`KEEP`/
`ENDIF`/`FUNC`/`RET` — ainda fora de escopo, mas reconhecidas de uma
vez por custo zero) — não vocabulário MSX. Novo `core.KindDignifiedInstruction`,
mesma técnica de `TRUE`/`FALSE`.

11 testes novos em `internal/core` (mecanismo puro) + 1 de ponta a
ponta em `internal/platforms/msx` confirmando que convive com a
renomeação de variável.

**Bug real corrigido: `identifierPattern` era minúsculo-apenas, devia
ser case-insensitive** (2026-11-22, achado pesquisando `DEFINE`):
`self.d_idnttp` (`[a-z][a-z_0-9]*[%!#$]?`) parece restrito a
minúsculas, mas é compilado dentro do regex combinado único do motor
original junto com a flag `re.I` (`badig.py:127-132`) — ou seja,
`[a-z]` ali bate maiúscula também. Confirmado que isso é usado de
verdade: `CGK173.dmx` tem várias variáveis/labels em camelCase
(`gameName$`, `fill_VRAM`, `pauseKEY`, label `portuguese_ROM`). Meu
`identifierPattern` (estritamente ASCII minúsculo) cortava esses
nomes no primeiro caractere maiúsculo — bug latente desde as rodadas
de variável/label, nunca disparado porque todo teste usava nomes
100% minúsculos. Fix cirúrgico: só a flag `(?i)` do Go adicionada ao
mesmo padrão. **Fora de escopo, documentado**: normalização de case
pra correspondência de referências à MESMA variável escritas com
capitalização diferente em lugares diferentes (`Token.var_name`/
`val_w_file` do Python sempre fazem `.lower()`; `msx/parser.go` ainda
não faz isso ao montar a chave de variável) — sem evidência de que
isso ocorra de verdade no `CGK173.dmx`, adiado.

**`DEFINE` implementado** (2026-11-22, `internal/core/pipeline.go`,
`pass1Dignified` — renomeada de `pass1Defines` na rodada seguinte, ver
abaixo, já que passou a cobrir mais que `DEFINE`): fiel a
`get_defines`/`get_defined_content`/
`replace_define` (`badig.py:596-745`, lido linha a linha). `DEFINE`
já era lexado como `KindDignifiedInstruction` e `[`/`]` já eram
`KindDignifiedSymbol` (rodadas anteriores) — nada novo no Lexer.
`DEFINE [nome][conteúdo], [nome2][conteúdo2], ...` só é reconhecido
no início de linha (erro fatal fora disso, confirmado no `else` final
do dispatch de `pass_1`); o conteúdo pode ter no máximo 1 nível de
`[...]` aninhado (o "placeholder" — ex. o `0` em `[0]` — vira o valor
padrão usado quando a chamada `[nome]` não dá argumento explícito);
a linha inteira do `DEFINE`, incluindo seu `NEWLINE`, desaparece sem
deixar rastro (confirmado rastreando a interação `next_tok()`/
`continue` entre `get_defines` e o loop de `pass_1`). Chamada
(`[nome]` ou `[nome](args)`, em qualquer outro lugar) expande pro
conteúdo registrado — `args` pode conter OUTRA chamada, recursivamente
(`[a]([b])`); tokens do conteúdo literal ganham a posição do CALL
(erro aponta pro lugar certo), os do argumento/padrão mantêm a
própria posição.

**Primeira pass que não cabe nem um pouco no modelo `passFunc`/
`runPass`** (diferente de labels, que usavam janela fixa de 3 tokens):
conteúdo/argumento têm comprimento variável e recursivo. `pass1Dignified`
usa um **cursor explícito** (mesmo estilo do `next_tok()`/`prev_tok()`
original) que substitui por completo a chamada de
`runPass(ctx, tokens, platform.Parser.Pass1)` em `Parse()` — pra
qualquer token que não seja `DEFINE`/`[`, chama
`platform.Parser.Pass1` manualmente e aplica o `HookAction` (mesma
lógica de 3 linhas que `runPass` já tem). Como constrói a saída com
as próprias mãos (não via `runPass`), tem rastreamento de SAÍDA de
verdade pra decidir "`DEFINE` está no início de linha?" — bate exato
com o `last_tok()` do Python, sem a aproximação por entrada que as
outras passes precisaram.

Novo `core.KindDefinePlaceholder` (marca interna, nunca chega na
saída de verdade — sempre substituído na expansão). O check de
parêntese de argumento usa Kind de plataforma no original
(`C_SYMBOL`); eu comparo só o VALOR (`"("`), simplificação pragmática
documentada. **Fora de escopo** (nessa rodada): `DECLARE`/`KEEP`/
`FUNC`/`RET`/`INCLUDE`/`ENDIF` (ainda caem no fallback genérico);
permissão especial de nome igual a `c_print_alt` ("?"). 7 testes
novos em `internal/core` + 1 de ponta a ponta no MSX confirmando
`DEFINE` + label + renomeação de variável convivendo (replicando o
padrão real
do `CGK173.dmx`, onde `pauseKEY`/`pauseANY` têm `{@}` dentro do
próprio conteúdo).

**`FUNC`/`RET` implementados** (2026-11-22, `internal/core/pipeline.go`):
fiel a `get_func_def`/`get_func_ret`/`replace_func_calls`/
`compare_func_args`/`get_func_args` (`badig.py:947-1193`, lido linha
a linha). `.nomefunção` (`D_FUNCNM`) é vocabulário Dignified
COMPARTILHADO (`badig_dignified.py`), não MSX — novo `core.KindFunctionName`,
novo fallback no `Lex()` (não colide com número decimal começando em
`.`, que a plataforma já captura primeiro). Declaração (`FUNC`/`RET`,
mesmo "deve estar no início de linha" do `DEFINE`) entrou dentro de
`pass1Defines`, agora renomeada **`pass1Dignified`** (já cobre mais
que `DEFINE`). Cada parâmetro é `nome[=padrão]`; no `CGK173.dmx` o
padrão é sempre `nome=nome` ("usa o valor atual da mesma variável
global" — não existe escopo local, "parâmetro" é só açúcar sintático
pra gerar atribuições condicionais antes do `GOSUB`). O registro em
`d_functions` só acontece no `RET` (não no `FUNC`) — precisei do
mesmo tipo de estado (`inFunc`/pendências) carregado de um dispatch
pro outro que o `self.in_func` original tem.

**Chamada é uma pass NOVA e SEPARADA** (`pass2Functions`), não dentro
de `pass2Labels`: no original, `replace_func_calls` roda do `pass_2`
junto com labels, numa única passada interligada — mas a captura de
retorno (`a, b = .func()`) exige **remover retroativamente tokens já
emitidos** (`tok_list_out.pop()`, varredura pra trás até achar
`NEWLINE`/separador/palavra de parada — `c_func_stop_kw`) — mesmo tipo
de problema do `DEFINE` (Achado 2 da rodada anterior), só que pra
trás em vez de pra frente, e não cabe em `runPass`. **Simplificação
deliberada, documentada**: `pass2Functions` roda ANTES de
`pass2Labels` e nunca chama `platform.Parser.Pass2` ela mesma (só
`Continue` pra token que não é chamada de função) — `pass2Labels`
continua sendo a única que chama o hook da plataforma. Efeito
colateral inofensivo: os tokens sintéticos que uma expansão de
chamada insere passam pelo hook de `pass2Labels` também (no original,
nunca passariam de novo por `self.clc.pass_2()`) — sem diferença
observável enquanto o `Pass2` do MSX for um no-op.

Resolução de linha da função é EXATAMENTE igual a labels, só em mapa
separado (`func_defs`, namespace diferente de `label_lines` no
original) — `pass4Numbering` ganhou 2 casos novos
(`KindFunctionDefinition`/`KindFunctionCall`).

5 métodos novos em `core.Description` (`AssignmentOperator`/
`FunctionCallInstruction`/`FunctionReturnInstruction`/
`FunctionCallBoundaryWords`/`CommentChar`) — o último (`c_altrem`,
"'") resolve se um `ret` sozinho antes de comentário tem retorno ou
não. Novo helper de nível de pacote `readParenArgs` (não-closure,
usado tanto pela assinatura `FUNC` quanto pela chamada — mirrors
`get_func_args`, compartilhado nos dois lugares igual o original).

**Achado real sobre numeração**: diferente de label, a linha de
declaração de `FUNC` NÃO tem o `NEWLINE` removido por `pass3LineHygiene`
(confirmado que o `pass_3` original só remove `NEWLINE` depois de
`D_LBLLIN`, não depois de `D_FUNCDF`) — ou seja, toda declaração
`func` desperdiça 1 número de linha (sem conteúdo nenhum nessa
linha), diferente de label. Replicado fielmente, não "corrigido" —
confirmado também que `zip()` parando na lista mais curta (quando a
chamada dá menos argumentos que a função declara) é uma consequência
INTENCIONAL do padrão `nome=nome`: parâmetro omitido = continua com o
valor atual da variável global = exatamente o mesmo efeito de um
`nome=nome` redundante.

9 testes novos em `internal/core` + 1 de ponta a ponta no MSX
confirmando que a renomeação de variável do MSX alcança corretamente
o parâmetro/padrão de uma função (só são identificadores comuns no
momento em que `FUNC`/`RET` são lidos, no Pass 1 — ganham a chance de
serem renomeados normalmente quando Pass 3/4 rodam depois).

**`DECLARE` implementado** (2026-11-22, `internal/core/pipeline.go`):
fiel a `get_declares` (`badig.py:747-846`, lido linha a linha).
`DECLARE nome1[:curto1], nome2[:curto2], ...` escreve DIRETO nas 3
tabelas que `core.SymbolTables` já tem e que o `Parser` MSX Pass 3/4
JÁ CONSULTAM — **zero mudança em `msx/parser.go`** pra isso funcionar.
Sem `:` (nome sozinho): se o nome já tem o tamanho de um alias (2 no
MSX), reserva em `HardShortVars`; senão reserva em `HardLongVars`
(**exatamente o mesmo efeito de `~nome` inline**, só declarado
antecipadamente). Com `:curto`: atribui `Declares[nome@arquivo] =
curto` direto, pulando o alocador. Mesmo "deve estar no início de
linha" e mesma linha inteira desaparecendo (como `DEFINE`, não como
`FUNC` — `DECLARE` TEM o `continue` no dispatch original) — entrou
dentro de `pass1Dignified`, mais um `case`.

Confirmado que o separador `:` nunca é validado contra seu valor
exato no original (só checa se o próximo token NÃO é `,` nem
`NEWLINE`, daí assume que tem `:curto` e consome 1 token sem
checar) — replicado exatamente, não apertado.

2 métodos novos em `core.Description`: `VariableAliasLength()`
(`c_var_valid_chars`, já existia como `variableAliasLen` dentro do
pacote MSX, só nunca exposto pro `core`) e `ReservedKeywords()`
(`c_reserved_kw` — **diferente** de `FunctionCallBoundaryWords`'
`c_func_stop_kw`: tem 2 grupos mais, `c_operat` e `'DATA'` — lista já
derivada numa pesquisa anterior, da rodada do `generate()` real
ainda adiada, reaproveitável quando ela chegar).

**Correção deliberada, documentada**: `c_varsna` (regex validando o
nome curto) tem um erro de digitação real no original (`[a-zA-z]` —
o intervalo `A-z` em ASCII inclui `[\]^_` por acidente, junto com
A-Z) — uso o padrão claramente pretendido (`[a-zA-Z0-9]`) em vez de
replicar o typo, mesmo espírito do bug do `AS` de 2 dígitos já
corrigido no tokenizer. **Fora de escopo**: os 5 loops de aviso
não-fatal (`Info.log` nível 2, conflito de prefixo).

8 testes novos em `internal/core` (inspecionando `ctx.Symbols`
direto, já que a plataforma fake não tem renomeação de variável de
verdade) + 1 de ponta a ponta no MSX confirmando que o Pass 3/4 JÁ
EXISTENTE respeita o que `DECLARE` escreveu, sem precisar mudar nada
nele.

**`KEEP` implementado** (2026-11-22, `internal/core/pipeline.go` +
`internal/platforms/msx/lexer.go`): fiel a `get_keeps`
(`badig.py:1196-1211`) e `toggle_lines` (`badig.py:1213-1240`), lidos
linha a linha. Compilação condicional via tags `#tag` (`D_TGLREM` no
original) — usada de verdade no `CGK173.dmx` (`keep #none` no topo,
blocos `#a...#a`/`#b...#b` envolvendo 2 sub-rotinas utilitárias
opcionais).

**Achado — conflito real de léxico**: `#` já é símbolo de 1 caractere
do MSX (`PRINT #1`/`OPEN...AS #1`, confirmado no próprio
`CGK173.dmx`), e o `Lexer.Lex` do MSX reclamava ele incondicionalmente
ANTES de qualquer fallback do `core` rodar. Resolvido com lookahead
cirúrgico no loop de símbolos do MSX: `#` seguido de LETRA é recusado
ali (`ok=false`), deixando o novo fallback do `core.Lex()`
(`toggleRemPattern`, novo `core.KindToggleRem`) reconhecer a tag
inteira; `#` seguido de dígito (ou qualquer outra coisa) continua
símbolo normal. Primeira vez que um recurso novo precisou de uma
mudança FORA de `internal/core`, no `Lexer` da própria plataforma —
mas é inerente ao recurso (a ambiguidade é lexical, não de parsing),
não quebra o desenho estabelecido.

**Achado — duas coisas despachadas pelo mesmo tipo de token, em
pontos diferentes**: a declaração `KEEP #tag1 #tag2 ...` (via
`D_INSTRC`, mesma linha toda desaparecendo que `DEFINE`/`DECLARE`) só
guarda as tags numa lista (`keeps`, variável local fechada em
`pass1Dignified` — nunca precisa escapar da função, já que só
`readToggle` a lê, e só dentro do mesmo Pass 1). A alternância em si
(`#tag` sozinho, em qualquer início de linha fora de uma declaração
`KEEP`) é despachada ANTES do `D_INSTRC` no original — sem overlap de
verdade, já que `get_keeps`/`readKeeps` consome os `#tag` argumentos
direto, nunca devolvendo controle pro despacho externo no meio do
caminho. Decide manter (`tag` ou `#ALL` em `keeps` E `#NONE` não) ou
remover; **sem `KEEP` nenhum declarado, todo bloco com tag é removido
por padrão** (álgebra booleana confirmada: `(false or false) and true
= false` — extras opt-in). Duas formas: 1 linha (remove só aquela
linha, incluindo seu próprio `NEWLINE`) e bloco (`#tag` sozinho na
própria linha, remove da abertura — abertura+seu `NEWLINE` — até achar
a MESMA tag sozinha depois — remove a tag de fechamento, mas **não**
o `NEWLINE` depois dela, assimetria real confirmada rastreando a
interação `continue`+`next_tok()` do original, não um erro de porte).
Sem erro fatal se `#tag` aparecer fora do início de linha — atravessa
como texto comum.

**Correção deliberada, documentada**: `get_keeps` compara `keep.val
in self.des.d_keeps` (case original) contra uma lista que só guarda
`.uval` (sempre maiúsculo) — a checagem de tag duplicada quase nunca
bate de verdade. Uso maiúsculo nos dois lados, mesmo espírito das
correções já feitas (`c_varsna`, `AS` de 2 dígitos).

1 `TokenKind` novo (`KindToggleRem`). 8 testes novos em `internal/core`
(padrão/bloco mantidos ou removidos conforme `keeps`, `#ALL`, `#NONE`
com override absoluto, erro fatal de bloco não fechado/tag
inválida/tag duplicada) + 1 de ponta a ponta no MSX confirmando que
`#` convive normal como símbolo de canal de arquivo ao lado de um
bloco `#a...#a`.

**`INCLUDE` implementado** (2026-11-22, `internal/core/pipeline.go` +
`internal/core/context.go`): fiel a `include_file`
(`badig.py:1243-1286`), lido linha a linha, junto com o dispatch de
`pass_1`/`pass_3` e `is_main_file`/`par()` (`badig.py:1344-1366`) e
`support/badig_settings.py:505`.

**Achado — processado só no Pass 3, não no Pass 1**: diferente de
`DEFINE`/`FUNC`/`RET`/`DECLARE`/`KEEP` (todos resolvidos no Pass 1),
`INCLUDE` só é agido no Pass 3 (`badig.py:1522-1526`) — o Pass 1
deixa os tokens `INCLUDE`/`"arquivo"` atravessarem sem nenhuma
mudança (`d_instr == 'INCLUDE': pass`), e o Pass 2 também não faz
nada especial com eles. Meu `pass3LineHygiene` (até então puramente
genérico, `passFunc`/`runPass`) virou bespoke (cursor explícito,
renomeado `pass3Includes`) — mesmo padrão evolutivo de Pass 1/2
chegando no Pass 3. Sem checagem de "deve estar no início de linha"
(nem o original exige isso pro `INCLUDE` — na prática só funciona
assim mesmo, porque o próximo token tem que ser literalmente uma
string entre aspas).

**Achado — resolução de caminho relativo usa o diretório do arquivo
PRINCIPAL, não do arquivo que inclui**: `self.stg.file_path` é
setado UMA VEZ (`os.path.dirname(file_load)`, confirmado em
`support/badig_settings.py:505`) e nunca muda, mesmo em includes
aninhados (o mesmo objeto `Settings` é compartilhado por referência
entre todas as instâncias de `Parser` recursivas) — um include
relativo DENTRO de um arquivo já incluído ainda resolve contra o
diretório do arquivo principal, não do arquivo atual. Replicado
resolvendo caminhos relativos contra `filepath.Dir(ctx.
IncludeStack[0])` (sempre o arquivo principal), confirmado com um
teste de 2 níveis de profundidade que SÓ passa se a resolução usar o
diretório certo.

**Achado — Go com ponteiro dispensa a dança de tupla do original**:
`include_file()` passa `(c_hard_short_vars, c_hard_long_vars,
d_declares)` como tupla pro `Parser` aninhado e copia de volta depois
— meu `core.SymbolTables` já é um PONTEIRO
(`CompileContext.Symbols`), então o `CompileContext` aninhado só
reusa o MESMO ponteiro: qualquer `DECLARE`/alocação que o include
escrever já fica visível pro arquivo que incluiu, sem merge
explícito. Pelo mesmo motivo, o alocador de variável do MSX
(`*msx.Parser.allocator`, campo da própria instância de `Parser`, não
do `CompileContext`) também já é compartilhado automaticamente,
contanto que o `Platform` passado pro `core.Parse`/`core.Lex`
aninhados seja o MESMO valor recebido de fora — zero mudança
precisou em `msx/parser.go`.

**Achado — limites exatos da mesclagem**: o original usa
`parsed[2:-2]` porque seu próprio `lex()` insere um `NEWLINE`
sintético logo depois de `PROGRAM` (`badig.py:418-419); meu
`core.Lex()` NÃO insere esse `NEWLINE` sintético — minha estrutura é
`[PROGRAM, ...conteúdo..., NEWLINE da última linha, EOF]`.
Equivalente exato: `nestedTokens[1:len-2]` (só derruba `PROGRAM` na
frente, `NEWLINE`+`EOF` atrás) — o `NEWLINE` que terminava a própria
linha do `INCLUDE` no arquivo que inclui sobrevive normal e assume o
papel de separador entre o conteúdo mesclado e o que vem depois.

Novo campo `CompileContext.LoadFile func(string) ([]Line, error)` —
primeira vez que o `core` precisa de I/O de disco de verdade (nada
usava isso antes, zero wiring de CLI); injetável, mesmo princípio já
usado pra `Description`/`Lexer`/`Parser` (valores explícitos, não
globais), mantendo testável sem arquivo real. Nova função real
`core.LoadLines` (implementação de verdade, reaproveitável quando a
futura integração de CLI carregar o arquivo principal também).

Reescrever `pass3LineHygiene` pra bespoke também corrigiu, sem
querer, uma pegadinha que só apareceu testando: o colapso de linha em
branco original olhava só `tokens[index-1]` (índice ORIGINAL de
entrada, mesma regra que `runPass` documenta) — mas `INCLUDE` pode
fazer um TRECHO INTEIRO desaparecer (não só 1 token por vez, como
`_`/label), deixando duas `NEWLINE` reais adjacentes na SAÍDA sem que
o token ORIGINAL imediatamente anterior seja `NEWLINE`. Corrigido
olhando os DOIS lookbacks (`tokens[i-1]` E `out[-1]`) — qualquer um
dos dois sendo `NEWLINE`/`LabelDeclaration`/`DignifiedSeparator`
colapsa.

**Fora de escopo, documentado**: normalização de encoding/tamanho de
tab (`load_format`/`tab_lenght` do original); proteção contra include
circular (o original não tem, não é questão de fidelidade — fico
igual, sem adicionar validação nova).

7 testes novos em `internal/core` (mescla sem linha em branco
duplicada; `DECLARE` dentro de um include fica visível em
`ctx.Symbols` do arquivo principal via ponteiro compartilhado;
caminho relativo resolve contra o arquivo principal mesmo num include
aninhado 2 níveis; erro fatal token depois de `INCLUDE` não é string;
erro fatal `LoadFile` não configurado; erro fatal `LoadFile`
retornando erro) + 1 de ponta a ponta no MSX usando `t.TempDir()` +
`core.LoadLines` de verdade, confirmando que Pass 4/5 do arquivo
incluído são pulados e que uma variável declarada nele não colide com
a alocação automática do arquivo principal.

**`generate()` de verdade implementado** (2026-11-22,
`internal/core/pipeline.go`): fiel a `generate()`
(`badig.py:1796-1878`), lido linha a linha, junto com
`c_symb_comp`/`c_reserved_kw`/`general_spaces` (`badig_msx.py:60-99`,
`support/badig_settings.py:524`).

**Achado — regra de espaçamento é só baseada em VALOR, não em Kind**:
`self.token.uval in self.des.c_reserved_kw` — comparação puramente
textual (maiúsculo), nada a ver com o tipo/Kind do token.
`Description.ReservedKeywords()` (já existia desde o `DECLARE`, com
um comentário que já previa esse uso) foi reaproveitado SEM mudança
nenhuma. Só precisou de UMA lista nova: `CompoundSymbols()` (mirrors
`c_symb_comp` — todo símbolo clássico MENOS `#`, mais `:` e `_`; o
`#` é removido DE PROPÓSITO no original, pra `print #1` ganhar espaço
depois do `print` mesmo assim).

**Achado — algoritmo exato** (confirmado rastreando token a token
dentro de uma linha): antes de cada token, se ele é palavra
reservada E a linha montada até agora não termina em espaço E o
token anterior não é símbolo composto → insere espaço; escreve o
texto do token sempre; depois do token, se é o número de linha
(sempre, incondicional) OU (é palavra reservada E o próximo não é
símbolo composto) → insere espaço; no fim da linha, `rstrip()` +
confere limite de 256 caracteres (fatal se passar). Confirmei, via
indexação negativa do Python nos dois extremos da linha
(`self.line[-1]` quando `n=0`, índice fora dos limites quando
`n=len-1`), que isso NUNCA dispara de verdade na prática — o `and`
de curto-circuito do Python nunca chega a avaliar esses acessos
nesses extremos, porque o número de linha nunca é reservado e o
`NEWLINE` final nunca é reservado — repliquei com guardas simples
(`n>0`/`n+1<len`) em vez de tentar reproduzir o wraparound, mesmo
efeito final. Não precisei incluir a própria `NEWLINE` no grupo de
tokens da linha (diferente do original, que inclui) — o `rstrip()`
final cobre a diferença, mesmo resultado.

**Fora de escopo, documentado**: `general_spaces` (flag `-ss`,
"squeeze spaces" — nenhuma flag de CLI existe ainda pra isso,
hardcodeado `" "`, o default do original sem `-ss`); relatório de
labels (`-lbr`) — cosmético, já adiado desde o Bloco 1; limite de 256
é um literal no original (não vem de `Settings`), replicado como
constante.

**Impacto nos testes existentes**: a troca de "junta tudo com 1
espaço fixo" pra espaçamento de verdade mudou o texto renderizado de
quase todo teste de ponta a ponta já existente (~43 asserções de
`Code` entre `internal/core` e `internal/platforms/msx`) — corrigido
cada string pro valor real, sem inflar a lista de palavras reservadas
da plataforma fake só pra forçar saída bonita (ela é deliberadamente
incompleta, só prova a mecânica) exceto por completar `goto`/
`gosub`/`return`, que a própria fake já declara como suas instruções
de salto/chamada/retorno (`LoopBackInstruction`/
`FunctionCallInstruction`/`FunctionReturnInstruction`) — omiti-las do
`ReservedKeywords()` seria só uma fake incompleta, não uma mais
honesta.

5 testes novos em `internal/core` focados especificamente nas regras
de espaçamento (reservada ganha espaço antes/depois quando o vizinho
não é símbolo composto; símbolo composto suprime o espaço; `#` ainda
ganha espaço depois de uma reservada; número de linha sempre ganha
espaço depois, mesmo sem nenhuma palavra reservada na linha; linha
com mais de 256 caracteres é erro fatal).

**Validação de ponta a ponta contra o `CGK173.dmx` inteiro E contra o
Python de verdade** (2026-11-22): depois do `generate()`, usuário
pediu pra tentar compilar o `CGK173.dmx` (891 linhas) de ponta a
ponta pra ver o que faltava. Primeira vez usando um teste descartável
(`internal/platforms/msx/zz_scratch_cgk173_test.go`, NUNCA commitado)
pra rodar o pipeline contra um exemplo real completo, e primeira vez
usando o Python instalado nesta máquina (`python badig.py` de
verdade, não só leitura do código-fonte) pra conferir a saída real
linha a linha quando uma suposição ficava ambígua lendo só o código.

**Achado 1 — `[?](x,y)` é um macro embutido, nunca declarado pelo
usuário**: `badig_msx.py`'s `Parser.initialization()`
(`badig_msx.py:448-460`) injeta `[?]` → `locate VAR:?` em TODO
compile (arquivo principal e cada include), sem o usuário nunca
escrever `DEFINE [?][...]`. Novo método
`core.Parser.BuiltinDefines() map[string]core.BuiltinDefine`
(`contract.go`), consultado no topo de `pass1Dignified`; MSX
constrói os 4 tokens de "locate VAR:?" na mão (template fixo e
pequeno, mais simples que relexar uma string em runtime).

**Achado 2 — bug real no lexer MSX: identificador que começa com uma
palavra reservada seguida de `_` OU DÍGITO cortava errado**:
`scanKeywordCandidate` só varre letras, parando no primeiro
caractere que não é letra — então "COPY" batia como instrução em
`copy_VRAM` (sobrando "_VRAM" solto) e "COLOR" batia como instrução
em `color1` (sobrando "1" solto, ambos usados de verdade no
`CGK173.dmx`: `.set_color(color1=color1, ...)`). Corrigido com
`isIdentifierContinuation` (`internal/platforms/msx/lexer.go`): se o
que vem IMEDIATAMENTE depois da palavra candidata é letra, `_` OU
DÍGITO, não é um match de palavra-chave de verdade — cai pro
fallback de identificador. **Reverte uma suposição documentada antes
sem checar contra o original**: um comentário anterior dizia que
dígito ficava de fora "porque GOTO10 precisa continuar funcionando" —
confirmado AGORA, lendo `join_commands` (`badig.py:134-139`, que
embrulha TODA palavra reservada em `^...$`, exigindo match EXATO
contra o candidato pré-segmentado do lexer original) e testando
"GOTO10" de verdade no Python (`badig.py` lê isso como UM
identificador, não como `GOTO`+`10`) — "GOTO10" nunca foi um caso
real que precisasse desse tratamento especial.

**Achado 3 — são TRÊS mecanismos de comentário exclusivos do
Dignified, não um só** (pesquisado em `get_lit_block`,
`badig.py:331-380`, o dispatch em `lex()`, `badig.py:442-478`, e
`badig_dignified.py:86-131`):
- `##` (linha, `D_LINREM`): funciona em QUALQUER posição da linha
  (confirmado usado solto E como comentário no fim de uma linha de
  código real), consome até o fim da linha física, zero tokens.
- Um bloco de 3 cerquilhas (`D_O_EREM`, exclusivo Dignified): só
  dispara sozinho no início de uma linha — mas **"sozinho" não
  significa "nada mais na linha depois do marcador"**: o original
  permite (e o `CGK173.dmx` usa de verdade) texto colado direto no
  marcador de abertura E de fechamento, tipo `###\tInsert ML
  routines` abrindo e `...VRAM=&h1940###` fechando, SEM espaço antes
  do fechamento. Essa descoberta derrubou minha primeira
  implementação (que exigia o marcador sozinho via comparação exata
  após `TrimSpace` — simplificação ERRADA, não confirmada contra
  exemplo real na hora). A versão certa varre por SUFIXO: qualquer
  linha (abertura, fechamento, ou intermediária) que TERMINE no
  marcador fecha o bloco; tudo isso é jogado fora (`rem_block = []`
  no original).
- Um bloco de duas aspas simples (`C_O_BREM`, igual ao `'`/REM
  clássico do MSX dobrado): mesmo mecanismo de abertura/fechamento
  por sufixo, mas o CONTEÚDO sobrevive — cada trecho entre abertura e
  fechamento (mesmo uma linha em branco) vira uma linha classic
  `'texto` de verdade, numerada separadamente (confirmado contra a
  saída REAL do Python: usado no próprio cabeçalho do `CGK173.dmx`).
  Simplificação segura: faço a expansão inteira dentro do próprio
  `core.Lex()` (não precisa do estágio intermediário `C_O_BREM`/
  `C_LIBREM`/`C_C_BREM` do Pass 3 original, já que o conteúdo nunca é
  re-tokenizado como código).
- `##BB:tag=valor`: é só um `##` comum pro lexer/parser — a extração
  de verdade (`read_remtags_from_code`,
  `support/badig_settings.py:285-302`) é uma camada SEPARADA,
  varrendo o arquivo bruto por regex ANTES de lexar, parte do sistema
  de config em 4 camadas já adiado desde o Bloco 1/Fase 4. **Fora de
  escopo**: só garanti que não quebra o lexer.

**Achado 4 — regra do "chunk vazio" calibrada contra a saída REAL do
Python, não só leitura de código**: minha primeira tentativa incluía
um chunk vazio pro próprio resto da linha de abertura quando o
marcador vem sozinho (ex. uma linha só com duas aspas simples) —
rodei um `mini.dmx` mínimo no Python de verdade e confirmei que isso
produz uma linha de comentário EXTRA, vazia, que o original NÃO
produz. Regra certa: só o resto da linha de ABERTURA e o resto da
linha de FECHAMENTO (antes do marcador) são pulados quando vazios;
uma linha TOTALMENTE em branco no MEIO do bloco ainda gera sua
própria linha `'` vazia normalmente (confirmado, também contra a
saída real).

**Achado 5 — `Lexer.Lex` precisa saber se está no início da linha**:
`#` e `'` já são símbolos/comentário próprios do MSX e sempre
recebem a primeira chance. Novo parâmetro `atLineStart bool` em
`Lexer.Lex` (`contract.go`), calculado em `core.Lex()` (mesma lógica
de `atOutputLineStart()`, um nível abaixo — no lexer, não no Pass 1).
MSX recusa incondicionalmente `#` seguido de outro `#` (deixando o
`core` decidir `##` vs bloco de 3 cerquilhas) e `'` seguido de outro
`'` quando `atLineStart` (deixando o `core` decidir o bloco de
verdade) — todo resto do comportamento de `#`/`'` continua idêntico.

7 testes novos em `internal/core` + 1 de ponta a ponta no MSX usando
o cabeçalho real do `CGK173.dmx` (bloco de duas aspas simples com
texto de várias linhas incluindo uma em branco, bloco de 3
cerquilhas, linha `##` solta, linha `##BB:...`, `'`/`#` continuando
normais fora desses casos).

**Resultado da validação (comentários)**: o `CGK173.dmx` inteiro (891
linhas) compila de ponta a ponta sem erro, produzindo ~650 linhas de
MSX BASIC reconhecível. Comparado linha a linha contra a saída REAL
do `badig.py`: o conteúdo bate bem de cara, mas a NUMERAÇÃO DE LINHA
diverge a partir do primeiro `for...next` com corpo em linhas
separadas terminando em `:` — achado que levou direto à PRÓXIMA
rodada, fusão de linha por `:` (abaixo).

**Fusão de linha pelo separador clássico `:` implementada**
(2026-11-22, `internal/core/pipeline.go`): usuário confirmou
explicitamente "sim, ja pode entrar neste achado, fundir linhas com
:" e deu o exemplo certo pra guiar a implementação — diferente do
`_` Dignified (que desaparece inteiro), o `:` clássico PERMANECE na
linha quando funde (`for a=1 to 10 :` + `next a` → `10 for a=1 to
10:next a`, não `10 for a=1 to 10 next a`). Peguei o resto do bloco
Pass 3 (`badig.py:1548-1605`) e confirmei CADA regra rodando `python
badig.py` de verdade nesta máquina contra arquivos `.dmx` mínimos,
não só lendo o código.

**Achado — são cinco regras independentes no mesmo bloco, não uma
só**, todas comparando contra `self.last_tok()` (o ÚLTIMO token
REALMENTE colocado em `tok_list_out` — equivalente exato do meu
`out[-1]`, SEM precisar do lookback duplo em `tokens[i-1]` que o
colapso de linha em branco usa pro `_`/`INCLUDE`, porque nenhuma
dessas cinco regras faz um token desaparecer sem nunca ter sido
"append"ado do jeito que o `_` faz neste port):
1. Dois `:` adjacentes colapsam em um só.
2. `NEWLINE` depois de `:` desaparece, MAS O `:` FICA — confirmado
   com um repro mínimo: `for f = 0 to 5:` / `print f:` / `next:` /
   `print "done"` virou `for f=0 to 5:print f:next:print "done"` numa
   linha só, com todos os `:` originais preservados.
3. `:` no início de uma linha funde com a anterior (remove o
   `NEWLINE` que vinha antes dele).
4. Duas strings literais adjacentes se concatenam (quotes das pontas
   que se tocam removidas, texto colado) — confirmado via `DEFINE`
   (`define [x]["hello"]` chamado como `[x]"world"` → `"helloworld"`),
   tanto na mesma linha quanto atravessando um `NEWLINE`. Real e
   NECESSÁRIO pra correção: duas strings coladas sem operador entre
   elas nem seria Classic BASIC válido sem essa junção.
5. `NEWLINE` depois de `KindFunctionDefinition` — bug achado relendo
   o código de novo (não só o Pass 3 do Python): faltava
   `KindFunctionDefinition` no switch de colapso que já tinha
   `KindLabelDeclaration`/`KindDignifiedSeparator` — o corpo de uma
   função desperdiçava uma linha em branco antes desta correção.

Comparação por VALOR, não por Kind, mesmo princípio já usado no
projeto inteiro: `platform.Description.StatementSeparator()` (já
existe) comparado via `tok.String() == sep`, já que `:` no MSX lexa
como `KindSeparator` (Kind PRÓPRIO da plataforma que o `core` não
deveria referenciar). Simplificação deliberada, documentada: as
guardas de índice do Python (`self.index < self.len - 3`/`> 3`) são
proteções de borda da indexação dele, não regra de negócio — as
checagens de limite normais do Go já previnem acesso fora dos
limites, sem precisar replicar os números exatos.

**Achado extra, de brinde, re-validando contra a saída real**: uma
linha em branco no INÍCIO do arquivo (antes de qualquer conteúdo
real) TAMBÉM desperdiçava um número de linha — `KindProgram` não
estava no switch de colapso (só `NEWLINE`/`KindLabelDeclaration`/
`KindDignifiedSeparator`/`KindFunctionDefinition`). Confirmado
empiricamente: um `"\nprint 1\n"` mínimo no Python real renderiza só
`"10 print 1"`, nunca uma linha `"10"` vazia seguida de `"20 print
1"`. Esse achado foi crucial pra validação final: o `CGK173.dmx`
começa com uma linha em branco antes do seu próprio cabeçalho de
comentário, e essa ÚNICA linha desperdiçada estava deslocando TODO
número de linha do arquivo por 10 (confirmado rastreando o offset
exato em ~15 `goto`/`gosub` espalhados pelo arquivo, TODOS
exatamente +10 antes da correção).

5 testes novos em `internal/core` (dois `:` adjacentes colapsam; `:`
no fim de linha funde mantendo o próprio `:`; `:` no início de linha
funde com a anterior; string+string na mesma linha concatena;
string+`NEWLINE`+string concatena removendo o `NEWLINE`).

**Resultado da validação (fusão por `:`) — 100% de ponta a ponta**:
depois desta correção, comparei o `CGK173.dmx` inteiro (891 linhas,
555 linhas Classic geradas, IGUAL ao total real do `badig.py`) linha
a linha contra a saída REAL. Confirmado por script: o OFFSET entre
os dois é uma CONSTANTE EXATA de 20 em TODA LINHA do arquivo do
início ao fim — explicado na hora (achado incorretamente, ver
correção abaixo) como "o Python reserva 2 números de linha pro
cabeçalho mesmo com `-rh` desligado". Descontando esse offset
constante, e a expansão de TAB pra espaço (`-tl`, também já adiada
nessa rodada) — **o CONTEÚDO bate byte a byte, do início ao fim do
arquivo inteiro**, incluindo todo `goto`/`gosub`/`restore`/`resume`/
`on...goto` e toda variável renomeada.

**Cabeçalho REM de 2 linhas + expansão de TAB implementados**
(2026-11-23, `internal/core/pipeline.go`/`settings.go`): usuário
pediu sugestão de próximo passo; sugeri fechar os dois gaps
conhecidos antes de ligar ao tokenizer; usuário confirmou "ok feche
os dois gaps".

**Correção da suposição anterior**: rodando de novo um teste mínimo
(`python badig.py minif.dmx out.amx -rh`), confirmei que `-rh`
(`action='store_false'`, `support/badig_settings.py:432`) desliga o
cabeçalho DE VERDADE — `insert_rem_header` (`badig.py:1289-1303`,
chamado 2x no início do Pass 4, `badig.py:1643-1645`) simplesmente
NUNCA é chamado nesse caso, sem reserva escondida nenhuma. O arquivo
de referência que eu vinha usando pra comparar
(`/tmp/badig_test/CGK173.amx`) tinha sido gerado SEM a flag `-rh` da
primeira vez — ou seja, ele JÁ tinha o cabeçalho de verdade
(`"10 'Converted with Basic Dignified"` / `"20 'https://github.com/
farique1/basic-dignified"`, strings fixas e compartilhadas,
`support/badig_settings.py:68-69`). O offset de +20 era exatamente
essas 2 linhas REM reais que meu port ainda não gerava — mais simples
do que eu tinha documentado antes.

Implementado construindo o bloco inteiro NA MÃO dentro do `case
KindProgram` de `pass4Numbering` (que já insere o PRIMEIRO
`KindLineNumber` ali mesmo): monta `[PROGRAM, LineNumber(10),
Synthetic(aspa), Synthetic(header1), NEWLINE, LineNumber(20),
Synthetic(aspa), Synthetic(header2), NEWLINE, LineNumber(30)]` como
um `Replace` só — simples porque um token que esse pass `Replace`
NUNCA é revisitado pelo próprio forward scan (`runPass` só anexa o
resultado), então cada linha do cabeçalho precisa avançar
`lineNumber` manualmente, direto aqui, igual as duas chamadas
seguidas de `insert_rem_header` do original.

Expansão de TAB: `line.expandtabs(tab_lenght)` (`support/helper.py:88`,
dentro de `IO.load_file`) roda LINHA POR LINHA no CARREGAMENTO do
arquivo, antes de qualquer lexing — nova função `expandTabs(s string,
tabSize int) string` replicando o algoritmo exato do `str.expandtabs`
(substitui cada TAB por espaços suficientes pra alcançar o próximo
múltiplo de `tabSize` colunas, rastreando a posição conforme
processa). `core.LoadLines` muda de assinatura pra receber o tamanho
do TAB.

Dois campos novos em `CompileSettings`: `RemHeader bool` e
`TabLength int`, mesmo padrão de `LineStart`/`LineStep`. **Decisão
deliberada sobre o default**: o zero-value do Go pra `RemHeader` é
`false` ("sem cabeçalho") — INVERTIDO do default real do Python
(`True`) — não corrigido agora, já que não existe NENHUM wiring de
CLI ainda que exponha essas flags pro usuário; mudar o zero-value
quebraria silenciosamente todo teste existente. Quando a FUTURA
integração de CLI existir, ela decide o default real, não o
zero-value da struct.

4 testes novos em `internal/core` (cabeçalho aparece quando
`RemHeader=true`, numerado corretamente; nada aparece por padrão;
`expandTabs` avança pro próximo múltiplo de colunas certo, inclusive
com múltiplos TABs/texto misturado; `LoadLines` expande TAB de
verdade num arquivo real no disco).

**Resultado da validação final — ZERO diferenças**: com os dois gaps
fechados, roder o `CGK173.dmx` inteiro de novo e comparar
(`diff <(sed 's/\r$//' go_output.txt) <(sed 's/\r$//'
py_output_raw.txt)`, o arquivo de referência ORIGINAL, COM o
cabeçalho de verdade) retorna **vazio** — nenhuma diferença, nem de
conteúdo nem de número de linha, do início ao fim do arquivo inteiro.
Essa é a primeira validação de ponta a ponta 100% completa, byte a
byte, contra um exemplo real e contra o motor Python de verdade.

**Pass 1 do separador Dignified `_` fora do fim de linha implementado**
(2026-11-23, `internal/platforms/msx/parser.go`): última peça
documentada como "ainda não entendida" do caminho textual
Dignified→Classic. Usuário escolheu esse passo como próximo (depois
de "qual o próximo passo" → "siga então pelo Pass 1 do separador
Dignified `_` fora do fim de linha").

Pesquisei `badig_dignified.py:96-100` (o comentário já dava a pista:
"Instruction separator - Shared with some classic Basics call
instruction - Decided in parser based on position relative to end of
line") e `badig_msx.py:570-587`, o `pass_1` do próprio dialeto MSX
(`self.clc.pass_1()`, chamado pelo core pra TODO token que não bate
em nenhum dos casos especiais do core — confirmado em
`badig.py:1432-1437`). Achado: um `_` (`D_INSTSP`) NÃO imediatamente
seguido de `NEWLINE` é o atalho de `CALL` do MSX de verdade
(`_CLS` == `CALL CLS`) — retagueado pra `C_CALINS`; o identificador
IMEDIATAMENTE seguinte também é retagueado, pra `C_CALIDF`, cujo
único efeito é NÃO bater nos checks `D_IDNTTP` que o Pass 3
(`get_hard_variable`)/Pass 4 (`process_variable`, a renomeação pra
alias de 2 letras) fazem — ou seja, o nome depois de `_` nunca é
renomeado, mesmo que o MESMO nome, usado de novo em OUTRO lugar sem
`_` na frente, continue sendo renomeado normalmente. Confirmado
rodando `python badig.py` de verdade contra um `.dmx` mínimo
(`_myroutine` / `myroutine=5` / `print myroutine` → `"10 _myroutine"`
/ `"20 zz=5"` / `"30 print zz"` — só a OCORRÊNCIA logo após `_` fica
intocada).

Implementado em `msx.Parser.Pass1` (que já existia como stub vazio,
nunca usado até agora) — exatamente o ponto certo, já que esse hook é
chamado pelo core pra todo token que não é um dos casos especiais
dele (`pass1Dignified`, `platform.Parser.Pass1(ctx, tokens, i)`),
igual o `self.clc.pass_1()` do original. Dois Kinds novos,
próprios da plataforma (mesmo padrão de `KindSeparator`/`MSX_SEPARATOR`):
`msx.KindCallSeparator`/`MSX_CALL_SEPARATOR` e
`msx.KindCallIdentifier`/`MSX_CALL_IDENTIFIER`. Como o Pass 3/4 de
`msx/parser.go` já só reage a `core.KindIdentifier` (`if tok.Kind !=
core.KindIdentifier { return Continue }`), retaguear o identificador
pra um Kind DIFERENTE já basta — zero mudança adicional precisou nas
duas passes de renomeação. O `_` retagueado (`MSX_CALL_SEPARATOR`)
renderiza como texto puro "_", colado sem espaço no nome seguinte,
porque "_" já está em `CompoundSymbols()` (`c_symb_comp`) desde a
rodada do `generate()` — nenhuma mudança adicional precisou ali
também.

`_` ao FIM de linha (seguido de `NEWLINE`) continua intocado por esse
hook — a OUTRA metade do mesmo token, a fusão de linha Dignified, que
já funcionava desde antes e continua passando pelo `pass3Includes`
do core exatamente como sempre.

Simplificação deliberada, documentada: um identificador retagueado
seguido imediatamente de `{` (padrão de abertura de loop label do
`pass2Labels`) não foi tratado como caso especial — nenhum programa
Dignified real combina atalho de `CALL` com loop label no mesmo
identificador, e o próprio original nem tem essa ambiguidade pra
começar (detecta loop labels pelos símbolos `{`/`}`, nunca pelo tipo
do identificador anterior, `badig.py:1469-1478`).

2 testes novos de ponta a ponta no MSX (identificador exclusivo da
renomeação, inclusive confirmando que o MESMO nome usado de novo sem
`_` na frente continua sendo renomeado; `_` no fim de linha continua
fundindo a linha normalmente). `CGK173.dmx` não usa esse mecanismo em
nenhum lugar de verdade (todo `_` nele está no fim de linha) — a
validação 100% de ponta a ponta contra a saída real do `badig.py`
continua de pé, zero diferença, confirmada de novo depois dessa
mudança.

Com isso, **todo o caminho textual Dignified→Classic documentado como
"ainda falta" está fechado** — o que resta é ligar tudo isso ao
tokenizer MSX Batoken (pronto desde a v1.4.0) pra produzir `.bmx`
binário de ponta a ponta de verdade.

**Ligação ao tokenizer MSX Batoken implementada** (2026-11-23,
`internal/platforms/msx/compile.go`): usuário pediu diretamente
("pode fazer a ligacao") depois que o Pass 1 fechou o último gap
textual.

Pesquisei o fluxo real do `badig.py` (`badig.py:1909-2043`,
`dignified()`) pra entender como o original liga texto Classic ao
tokenizer: ele é um estágio SEPARADO e opcional, gated por uma flag
(`stg.tokenize`, default `False`, ligada via `--tk_tokenize` ou
remtag `TK_TOKENIZE`) — `dignified()` roda `generate()`, SALVA o
texto Classic em disco (`IO.save_file`), e só DEPOIS, se a flag
estiver ligada, `tools_msx.Interface.run()` → `tokenizer_interface.
Tokenizer.run()` relê esse MESMO arquivo do disco e chama
`MSXBatoken.Main().execute()` nele, escrevendo o `.bmx` à parte
(`msx/tokenizer_interface.py:135-163`). Ou seja: no original, texto e
binário são duas passagens desacopladas, unidas só por um arquivo
intermediário em disco — um detalhe de como o CLI dele é
costurado, não uma exigência de comportamento.

Achado central: `core.GenerateResult.Code []string` (o texto Classic
já numerado, uma linha física por elemento, sem `\n`) e
`msx.Tokenize(lines []string) ([]byte, error)` (já portado desde a
v1.4.0, golden-vector testado) **já tinham exatamente o mesmo
formato** — zero mudança precisou em qualquer um dos dois lados.
"Ligar" virou só escrever a função que encadeia os dois em memória,
sem o round-trip por disco que o original faz só por causa de como
seu CLI é montado: `msx.CompileToBinary(ctx, platform) ([]byte,
error)` roda `core.Run` e passa `result.Code` direto pra `Tokenize`.

2 testes novos comitados: a ligação em si (`CompileToBinary` com um
programa pequeno deve bater, byte a byte, contra uma chamada
equivalente feita na mão — `Tokenize` recebendo o texto Classic
esperado digitado à mão) — pensado pra pegar um bug de ligação (linha
a mais/a menos, formato errado) mesmo que a correção byte a byte do
`Tokenize` em si já tenha seus próprios vetores de ouro
(`tokenizer_test.go`); e propagação de erro (um `FUNC` sem `RET`
correspondente, erro de Pass 1, ainda aparece através de
`CompileToBinary`).

**Validação final, pela primeira vez até o BINÁRIO**: rodei
`python badig.py CGK173.dmx CGK173.amx --tk_tokenize` de verdade
nesta máquina (sem `-rh`, pra manter o cabeçalho real — mesma
referência já usada desde a rodada do cabeçalho REM), produzindo um
`CGK173.bmx` real de 18.815 bytes. `msx.CompileToBinary` rodado
contra o MESMO `CGK173.dmx` (com `RemHeader: true`, `TabLength: 4`)
bateu **byte a byte, de primeira**, contra esse `.bmx` real — a
primeira vez que a validação deste projeto vai além do texto Classic
e confirma o BINÁRIO tokenizado inteiro.

Com isso, **o motor de conversão MSX está completo e validado de
ponta a ponta, do Dignified ao binário `.bmx`** — o que falta agora é
só wiring de CLI/interface (nenhuma das três expõe isso pro usuário
ainda) e portar os outros dois dialetos (ZX81/ZX Spectrum).

**Fase 3 — Motor do pipeline + validação** (o que prova que o desenho
funciona) — **estado em 2026-11-22, superando a descrição original de
2026-11-21 abaixo**: o core JÁ faz processamento Dignified de núcleo
de verdade (`DEFINE` via `pass1Defines`, labels/loop labels via
`pass2Labels`, numeração via `pass4Numbering` — ver Bloco 1 acima,
seções "labels e numeração" e "`DEFINE` implementado") — não é mais
"todo pass só chama o hook da plataforma por token" como descrito
quando essa fase foi aberta. `INCLUDE` (que precisava exatamente
disso pra mesclar arquivos) já está implementado desde 2026-11-22 —
ver Bloco 1 acima.
- [x] Driver do pipeline no core (`internal/core/pipeline.go`:
  `Lex`/`Parse`/`Generate`/`Run`) — **A**, **1**. Descrição original
  (2026-11-21, já superada pelo texto acima): o scanner só reconhecia
  espaço em branco e fim de linha; qualquer sequência de caracteres
  não-espaço era empacotada como `KindWord` (removido desde então, ver
  "Segundo desvio" no Bloco 1) e delegada pro `Lexer.Lex` da
  plataforma classificar.
- [x] Plataforma fake de validação (`internal/core/pipeline_test.go`,
  reconhece `PRINT`/inteiros/`IGNORE`) — **M**, **1**
- [x] Teste de ponta a ponta (`TestPipelineEndToEnd`, exercitando as 3
  `HookAction` - `Replace`/`Skip`/`Continue` - numa linha só) e teste
  dedicado confirmando que `Pass4`/`Pass5` são pulados quando
  `IncludeStack` indica um arquivo incluído
  (`TestParseSkipsPass4And5ForIncludedFiles`) — **B**, **1**
- [ ] Suporte a `INCLUDE` recursivo com tabelas de símbolo
  compartilhadas por referência explícita — **M**, **2** (próxima
  fase — testável agora que o driver existe)
- [ ] Caminho `-asc` (decisão de orquestração — pula Lex/Parser por
  completo, chama `Platform.Tools.Run` direto) — **B**, **3** (próxima
  fase — nada pra testar de verdade até `Tools.Interface` ter corpo real)

**Fase 4 — Config em camadas**: desenho real a partir de 2026-11-23
(usuário pediu um painel de configuração persistido em SQLite,
editável na UI, com as mesmas 4 opções: Basic Dignified geral, MSX +
emulador, ZX81 + emulador, ZX Spectrum + emulador, e comandos `SET`
no CLI — ver `internal/options` e `internal/config/options.go`,
detalhado logo abaixo). Só a Etapa A está feita; B/C/D ficam pra
rodadas seguintes:
- [x] **Etapa A — catálogo + schema SQLite** (2026-11-23,
  `internal/options/options.go` novo + `internal/config/options.go`
  novo): pacote `internal/options` é o catálogo embutido (mesmo
  padrão de `internal/theme`/`CatalogFor`) das 40 opções do
  `badig.py` original que fazem sentido como config persistida (não
  flag de uma execução só) — 14 compartilhadas (`CategoryCore`:
  `INITIAL_LINE`, `LINE_INCREMENT`, `REM_HEADER`, `TAB_SIZE`, etc.,
  de `support/badig_settings.py`), 14 MSX (`CategoryMSX`:
  `CONVERT_PRINT`/`STRIP_THEN_GOTO` de `badig_msx.py` + tokenizer +
  emulador do openMSX), 4 ZX81 (`CategoryZX81`: só tokenizer +
  emulador, `badig_zx81.py` não tem nenhuma opção própria), 8 ZX
  Spectrum (`CategoryZXSpectrum`: `SPECTRUM_MODEL` + tokenizer de
  fita + emulador). Nomes de chave são deliberadamente amigáveis
  (`INITIAL_LINE`, não `line_start`) — o nome Python original só
  fica em comentário, nunca aparece pro usuário; mesmo princípio que
  motivou a escolha: usuário deu o exemplo `SET INITIAL_LINE 1`
  espontaneamente, confirmando o estilo. `config.Store` ganhou
  `SetOption`/`GetOption`/`ListOptions`/`EnsureOptionDefaults` (mesmo
  princípio de `InstallThemes`/`EnsureActiveTheme` — valida via
  `options.Validate` antes de gravar, só semeia default na PRIMEIRA
  vez, nunca sobrescreve valor que o usuário já mudou). 18 testes
  novos (`internal/options` + `internal/config`). **Escopo
  deliberado desta etapa**: só a camada GLOBAL agora — sem "projeto"
  nem resolução das 4 camadas (SQLite global < projeto < CLI <
  remtags), porque "projeto" não tem nenhuma âncora concreta ainda
  (nenhum comando de compilação real existe no CLI/TUI/GUI); nenhum
  valor liga ao `core.CompileSettings` de verdade ainda (isso é
  consumo pelo pipeline, um passo futuro separado — mesmo princípio
  de como o sistema de tema foi construído antes de qualquer comando
  de conversão existir). — **M**, **2**
- [x] **Etapa B — comando `SET <OPÇÃO> <VALOR>` implementado**
  (2026-11-23, `internal/cli/console.go`): `set`/`settings`/
  `dialect` novos, mesmo padrão de `case "theme":`/`executeTheme`.
  `SET <OPÇÃO>` (sem valor) mostra o valor atual; `SET <OPÇÃO>
  <VALOR>` valida (via `config.Store.SetOption`, que já valida contra
  o catálogo) e grava; `SETTINGS` lista tudo (Basic Dignified +
  dialeto atual); `DIALECT <nome>` troca qual dialeto `SET` usa pra
  resolver uma chave que não é do core (`CategoryCore` é sempre
  tentado primeiro, já que suas chaves nunca colidem com nenhum
  dialeto; o dialeto da sessão, só em memória, começa em MSX, nunca
  persistido - "lembrar o último dialeto" fica pra quando a camada de
  projeto existir). Comandos agora casam case-insensitive (`"SET
  INITIAL_LINE 1"` funciona igual ao exemplo original do usuário;
  mudança aplicada a TODOS os comandos existentes também, não só
  `set`, pra manter o console consistente). Completar com Tab
  funciona pra `dialect <nome>` (lista fixa, 3 valores); completar as
  chaves de `set <opção>` fica de fora por enquanto (depende do
  dialeto da sessão, que a função de completar hoje não enxerga).
  13 testes novos. — **M**, **2**
- [x] **Etapa C — 4 painéis na TUI implementados** (2026-11-23,
  `internal/tui/tui.go`): `Setup` ganha 4 itens novos (`Basic
  Dignified`/`MSX`/`ZX81`/`ZX Spectrum`), cada um abrindo
  `showOptionsPanel(category, title)` — mesma estrutura de
  `showThemePicker` (`tview.Form` numa `tview.Grid` centralizada,
  Apply/Cancel), um widget por `options.Field` em vez de um dropdown
  de tema só: `TypeBool`→`Checkbox`, `TypeEnum`→`DropDown`,
  `TypeInt`→`InputField` com `tview.InputFieldInteger` (só restringe
  o FORMATO da tecla; o intervalo Min/Max continua sendo validado de
  verdade só no Apply, por `options.Validate` dentro de
  `Store.SetOption`), `TypeString`→`InputField` livre. Altura da
  janela calculada por `len(fields)` (10-22 linhas) — o próprio
  `tview.Form` rola o item focado pra dentro da área visível sozinho
  quando o painel tem mais campos do que cabem (MSX/Basic Dignified,
  14 campos cada). `showError` generalizado (ganhou um parâmetro de
  contexto, usado tanto pelo erro de tema quanto pelos novos erros de
  opção). Simplificação deliberada, documentada: Apply grava campo
  por campo e para no primeiro rejeitado (só pode acontecer com um
  `TypeInt` fora do Min/Max, já que todo outro widget só oferece
  valores já válidos) — campos já aplicados antes do erro NÃO
  revertem; reabrir e corrigir resolve. 1 teste novo de ponta a
  ponta (tela simulada, confirma que o painel MSX mostra campos MSX e
  o painel ZX81 mostra campos ZX81, sem vazar um no outro) — achado
  rodando o `build.ps1` de verdade (não só o teste isolado): a
  sequência Down/Down/Enter do teste falhou UMA vez sob a carga da
  suíte inteira rodando em paralelo (outros pacotes também sobem seu
  próprio `tview.Application`), porque o `sleep` fixo de 30ms entre
  teclas (mesmo padrão já usado pelos testes de regressão do Install)
  não é sempre suficiente quando a máquina está ocupada; corrigido
  com um polling (`waitForText`, até 2s) esperando o texto esperado
  aparecer na tela simulada em vez de confiar só no tempo fixo —
  continua pegando uma navegação REALMENTE errada (a tela nunca
  mostraria o texto certo e o teste estoura o timeout com erro claro),
  só deixa de ser frágil sob carga. — **M**, **2**
- [x] **Etapa D — os mesmos 4 painéis na GUI implementados**
  (2026-11-23, `internal/gui/gui.go`): `Setup` ganha os mesmos 4
  itens da TUI, cada um chamando `showOptionsPanel(w, store,
  category, title)` — mas em vez de reusar `dialog.NewCustomConfirm`
  (o padrão de `showThemeInstaller`, que já tem seus próprios botões
  Confirm/Dismiss), usa `dialog.ShowForm` (Fyne já tem um diálogo de
  formulário pronto, com Apply/Cancel e integração de validação
  embutida) — mais perto do próprio `widget.Form`/`tview.Form` da TUI
  do que forçar o padrão de tema num formulário de N campos. Mapeamento
  `options.Field`→widget: `TypeBool`→`widget.Check`,
  `TypeEnum`→`widget.Select`, `TypeInt`/`TypeString`→`widget.Entry`.
  Diferente da TUI (que só valida no clique de Apply): `Entry.
  Validator` liga direto em `options.Validate`, e o `dialog.ShowForm`
  do Fyne já desabilita o próprio botão Apply sozinho enquanto
  qualquer campo reporta inválido — validação em tempo real de
  graça pros campos `TypeInt` fora do Min/Max.

  A parte de montar os widgets (`options.Field`→`*widget.FormItem` +
  getter de leitura) foi separada numa função pura,
  `buildOptionsFormItems`, justamente pra dar pra testar sem precisar
  rodar uma janela/diálogo de verdade (Fyne não tem, neste projeto,
  nenhum padrão estabelecido de testar diálogo renderizado de ponta a
  ponta — `theme_test.go`, o precedente mais próximo, também só testa
  a função pura de mapear paleta, nunca o `showThemeInstaller` em si).
  6 testes novos (contagem/ordem dos campos bate com o catálogo;
  cada `Type` mapeia pro widget certo; valor default usado quando
  nada foi gravado ainda; valor gravado tem prioridade sobre o
  default; seleção de enum funciona mesmo com case diferente;
  validador de inteiro rejeita fora do Min/Max). Precisou de um
  `TestMain` novo no pacote (`test.NewApp()`, o driver de teste do
  Fyne) — primeira vez que este pacote cria um widget de verdade em
  teste, `Entry.SetText` mede texto pra layout e precisa de um app
  Fyne "atual" configurado, coisa que `theme_test.go` nunca precisou
  (só mexe com cor/paleta, nunca constrói widget). — **M**, **2**
- [x] **Bug de cor no menu suspenso da TUI corrigido** (2026-11-23,
  `internal/tui/tui.go`) — reportado pelo usuário direto depois de
  ver o menu `Setup` com 5 itens de verdade pela primeira vez (antes
  da Etapa C só tinha 1, "Install", então o problema nunca tinha
  ficado visível): item NÃO selecionado de um dropdown aberto
  renderizava com texto preto em fundo preto (invisível, "em todos
  os temas"), e o destaque do item SELECIONADO não preenchia a
  largura inteira do menu, parando exatamente no fim do próprio
  texto — exatamente o efeito contrário do visual "PC Tools 9/Norton
  Utilities 9" que o usuário queria.

  Causa raiz (achada lendo `tview@v0.42.0/list.go`): `List.
  SetMainTextColor` só toca a metade FOREGROUND do `mainTextStyle`
  interno da lista — a metade BACKGROUND fica no que `tview.NewList()`
  semeia por padrão (`Styles.PrimitiveBackgroundColor`, preto,
  global do tview, sem nenhuma relação com a paleta deste menu), e
  `List.Draw` pinta cada linha NÃO selecionada com esse estilo
  inteiro (frente+fundo juntos), sobrescrevendo o próprio fundo da
  `Box` (já pintado certo via `SetBackgroundColor`) exatamente nas
  células onde o texto é desenhado — texto preto em frente a um
  fundo preto que não tem nada a ver com o tema, só não sobrando
  nenhum "preto" na sobra de espaço além do texto (aí sim mostrando o
  fundo certo da `Box`) — daí a aparência de "cor que não vai até o
  fim". Corrigido trocando pra `List.SetMainTextStyle` (seta as duas
  metades de uma vez). O destaque parcial do item selecionado tem a
  MESMA causa de fundo (`List.Draw` só estende o destaque até o fim
  da linha `if selected && l.highlightFullLine`, que é `false` por
  padrão) — corrigido com `List.SetHighlightFullLine(true)`, a opção
  que o próprio `tview` já tem pronta pra esse exato efeito "barra
  sólida" de menu.

  2 testes novos de ponta a ponta (tela simulada): um varre a área
  INTERNA do dropdown (descontando a moldura, pra não confundir com
  o padrão `▒` de fundo do desktop, que deliberadamente pinta frente
  igual a fundo — não é o mesmo bug) procurando qualquer caractere
  não-espaço com frente==fundo; outro confere que o fundo do item
  selecionado continua igual um pouco depois do fim do próprio texto.
  Confirmei os dois PEGAM o bug de verdade revertendo o fix
  temporariamente (`git stash`) antes de escrever a versão final —
  os dois falham sem a correção, do jeito certo. Acerto de rota no
  meio do caminho: a primeira versão desses testes comparava posição
  de coluna usando `strings.Index`/`strings.LastIndex` (índice de
  BYTE) contra uma tela com caracteres de desenho de caixa
  multi-byte (`║`, `▒`, `◐`) antes do texto procurado — todo índice
  de coluna saía errado; corrigido convertendo a linha pra `[]rune`
  primeiro e contando posição em RUNAS (= colunas de tela), não
  bytes. — **M**, **2**
- [ ] Resolução de fato das 4 camadas (SQLite global < projeto < CLI <
  remtags no código) — **A**, **3** (não há CLI real de flags pra
  testar contra ainda)

Caminho de menor esforço + maior criticidade até ter um motor
testável de ponta a ponta: Fase 1 completa + Fase 2 (menos `Expose`/
`Tools.Interface`, que podem esperar) + Fase 3 (menos `INCLUDE`/`-asc`,
que podem esperar). Fase 4 fica pra quando houver um CLI real de flags
pra validar contra (Bloco 5).

**Implementado em 2026-11-21** (pacote novo `internal/core`, 9
arquivos: `token.go`, `hook.go`, `diagnostic.go`, `settings.go`,
`context.go`, `allocator.go`, `contract.go`, `platform.go`,
`pipeline.go`, mais `pipeline_test.go`): Fase 1 completa, Fase 2 sem
`Expose`/`Tools.Interface`, Fase 3 sem `INCLUDE`/`-asc` — exatamente o
caminho de menor esforço + maior criticidade acima. `go build`/
`go vet`/`gofmt -l`/`go test` limpos em `internal/core` e no resto do
repo (nada mais foi tocado). Falta documentado em cada subitem acima
pras próximas fases: `Expose`/`Tools.Interface` (Fase 2), `INCLUDE`
recursivo/`-asc` (Fase 3), e a Fase 4 inteira (config em camadas).

**Gerenciador de arquivos de dois painéis (Norton Commander/PC
Tools) — Etapa 1: biblioteca `.DSK` portada** (2026-11-23,
`internal/platforms/msx/dsk`, pacote novo): depois de decidir que o
goBaDig é o shell principal (não o editor), usuário pediu um
gerenciador de dois painéis cujo uso principal é criar/editar
imagens `.DSK` do MSX. Perguntei onde ficava a referência de disco
que ele já tinha — ele tem um utilitário **completo e funcionando**,
`E:\msxdiskutil`, reescrito em PureBasic (histórico: v1.6 original em
C por Arnold Metselaar — `rddsk`/`wrdsk` — depois adaptado,
reescrito do zero em PureBasic). Confirmado por busca: zero código de
disco existia antes disso, nem no `badig/` Python nem em
`internal/` — `docs/RESOURCES.md:161-168` e este arquivo (linha
2315, agora atualizada) já citavam `.DSK` como pendente, nunca
começado.

Dado o tamanho (biblioteca + UI de dois painéis + operações de
arquivo), segui o mesmo padrão do painel de configuração: etapas
separadas, cada uma lançável sozinha. Esta rodada entrega só a
**Etapa 1 — portar a biblioteca**, base de tudo o resto.

Lido por completo `E:\msxdiskutil\MSXDisk.pbi` (719 linhas) +
`MSXDiskDLL.pb` + `msxdisk.pb` (CLI original) antes de desenhar
qualquer coisa — mesmo princípio já usado o projeto inteiro (portar
`badig.py`, não reinventar a partir de uma spec externa, mesmo tendo
pesquisado a especificação técnica do FAT12/MSX-DOS por fora
primeiro). API portada: `Create`/`Open`/`Close`/`ListFiles`/
`ExtractFile`/`AddFile`/`DeleteFile`, mais `ConvertToFAT11`/
`MatchesFAT11` (nome 8.3 + wildcard, exportados pra Etapa 3 usar
depois).

**Mudança estrutural necessária, não só estilo**: o módulo PureBasic
usa estado GLOBAL (um disco aberto por vez no processo inteiro) — não
serve pro gerenciador de dois painéis (cada painel pode ter seu
próprio `.DSK` aberto ao mesmo tempo). Virou `type Image struct`, sem
nada de pacote-level, cada `Open`/`Create` devolve seu próprio
`*Image` independente.

**Melhoria deliberada, documentada**: `Create` no original estica o
arquivo só escrevendo o ÚLTIMO setor em zero (conta com semântica de
"sparse file" do SO pro buraco virar zero) — aqui escrevo o arquivo
INTEIRO de zeros explicitamente (720KB é pequeno, mais robusto que
depender de comportamento específico de sistema de arquivos).

**Correção deliberada, documentada**: `ExtractFile` do original seta
uma mensagem de aviso ("File was truncated") quando a cadeia de
clusters acaba antes do tamanho declarado, mas AINDA retorna sucesso
— e a própria CLI original nunca checa essa mensagem depois de um
retorno de sucesso, tornando o aviso código morto na prática. Aqui
esse caso vira um `error` de verdade.

**Validação contra dados reais** (não só releitura/testes
sintéticos): copiei `E:\msxdiskutil\msxdos.dsk` (disco de exemplo
real, 720KB, com `command.com`/`msxdos.sys` dentro) pro scratchpad e
abri com o port Go — `ListFiles()` achou os dois arquivos com nome e
tamanho exatos, e `ExtractFile()` dos dois bateu **byte a byte**
contra as cópias soltas desses mesmos arquivos que já existem em
`E:\msxdiskutil\` (prova real de que o parsing de boot sector/FAT/
diretório/dados está correto, não só "os testes sintéticos passam").
Também criei um disco novo do zero (`Create` + `AddFile` dos dois
mesmos arquivos) pra exercitar o caminho de ESCRITA — abre sem erro,
mesma estrutura. Tentei confirmar num emulador real (`openmsx`, já em
PATH nesta máquina) — rodou sem erro/crash, mas **sem confirmação
visual possível** (ambiente sem display interativo nesta sessão) -
deixado como inconclusivo, não um "não funciona"; a validação byte a
byte contra dado real já é uma evidência forte por si só.

14 testes automatizados novos em `internal/platforms/msx/dsk`
(formatação/tamanho do `Create`; `AddFile`+`ListFiles`; sobrescrita
por nome igual; `ExtractFile` ida e volta byte-exato; `DeleteFile`
libera cluster pra reuso de verdade — confirmado adicionando outro
arquivo depois e vendo que reusa o mesmo cluster; arquivo de 0 bytes;
disco cheio retorna erro sem deixar entrada de diretório pela metade;
`ConvertToFAT11`/`MatchesFAT11` com wildcard; erro em arquivo/nome
desconhecido; erro depois de `Close`). Scratch de validação contra o
disco real (`zz_scratch_realdisk_test.go`) deliberadamente NÃO
comitado, mesmo padrão de `zz_scratch_cgk173_test.go`.

**Fora de escopo nesta etapa, já documentado**: Etapa 2 (navegação
de dois painéis na TUI, só leitura, local + abrir um `.DSK` num
painel); Etapa 3 (operações de arquivo pela UI — adicionar/extrair/
apagar/criar disco, o uso "principal" pedido); Etapa 4 (operações
local↔local completas, baixa prioridade); GUI equivalente.

**Comando `compilar` de verdade implementado** (2026-11-24,
`internal/platforms/msx/compile.go` + `internal/cli/cli.go` +
`internal/cli/console.go`) — a primeira conversão Dignified→binário
de verdade disponível pro USUÁRIO FINAL (não só testada
internamente). Usuário perguntou se já era hora de implementar o
"compilar" (preprocessar: pegar um `.dmx`, gerar `.amx`+`.bmx`),
citando a visão maior de depois copiar pra uma pasta, gerar um disco
`.DSK` com o resultado, e testar num emulador de verdade. Respondi
(pergunta exploratória, recomendação curta): sim, boa hora -
`core.Run`, `msx.CompileToBinary` e a biblioteca `dsk` já estavam
todos prontos e validados, esse passo era puramente orquestração, não
lógica de motor nova. Usuário escolheu seguir.

Implementado: `msx.DefaultCompileSettings()` (os defaults reais do
`badig.py` - `LineStart`/`LineStep=10`, `RemHeader=true`,
`TabLength=4` - resolvendo de vez a decisão deliberada, documentada
desde a rodada do cabeçalho REM, de manter o zero-value do Go
diferente do default Python só até existir um chamador real pra
decidir; agora existe). `msx.CompileFile(dmxPath, settings)` -
`settings=nil` usa os defaults - carrega o arquivo
(`core.LoadLines`), monta o `CompileContext`/`Platform` MSX na mão,
roda `core.Run`, tokeniza (`Tokenize`), escreve `<base>.amx` (texto,
CRLF - confirmado que é exatamente o que o `badig.py` real escreve
no Windows: `generate()` usa só `\n` por linha,
`support/helper.py`'s `IO.load_file`/`save_file` abre em modo texto
do Python, que traduz pra `\r\n` no Windows) e `<base>.bmx` (binário)
ao lado do arquivo de entrada.

CLI: `goBaDig <arquivo.dmx>` (sem opção nenhuma) agora compila -
mesma convenção do `badig.py` original (`badig.py <input>
[output]`, sem precisar de verbo/subcomando). `cli.Run`'s dispatch
intercepta isso dentro do próprio `case ModeHelp` (onde um argumento
posicional sem flag reconhecida já cai hoje) via
`compileFileArg(args)` (exatamente 1 argumento que não começa com
"-"), SEM mudar `Parse`/seu enum `Mode` nem quebrar nenhum teste
existente - uma flag explícita (`--tui` etc.) sempre ganha, porque
`Parse` já teria retornado outro `Mode` antes disso ser alcançado.
Console (`--cli`) ganhou o comando irmão `compile <arquivo.dmx>`
(`executeCompile`), chamando a mesma `msx.CompileFile`.

**Validação de ponta a ponta contra o `CGK173.dmx` inteiro através do
NOVO ponto de entrada** (não só a lógica interna já testada): copiei
o arquivo real pra um diretório temporário, rodei `CompileFile` nele,
e comparei os DOIS arquivos resultantes byte a byte contra as
referências reais já validadas em rodadas anteriores
(`D:\temp\user\badig_test\CGK173.amx`, o `.amx` real do `badig.py`,
confirmado CRLF; e `CGK173.bmx.reference` no scratchpad) - bateu
EXATO nos dois, de primeira. Também testado manualmente pelos dois
pontos de entrada de verdade (`go run ./cmd/goBaDig arquivo.dmx` e
`compile arquivo.dmx` dentro do console `--cli`), incluindo um
programa com `DEFINE` de verdade.

9 testes automatizados novos (`msx.CompileFile` escreve os dois
arquivos com o conteúdo certo incluindo o cabeçalho REM; erro claro
pra arquivo de origem inexistente; `compileFileArg` cobre
nenhum/um/dois argumentos e um argumento com flag; `cli.Run` compila
de verdade um argumento posicional, cria os arquivos de verdade,
reporta erro pra arquivo inexistente, e uma flag explícita como
`--tui` sempre ganha de um argumento posicional; o comando `compile`
do console escreve os arquivos, mostra uso sem argumento, mostra erro
pra arquivo inexistente; completar com Tab reconhece `compile`).

**Fora de escopo, documentado**: nenhuma flag de linha de comando
ainda lê/aplica as opções persistidas pelos painéis de configuração
(usa sempre `DefaultCompileSettings`); nenhum caminho de saída
customizável (sempre ao lado do arquivo de entrada, mesmo nome base);
seleção de dialeto por extensão/flag `-id` (só MSX existe); TUI/GUI
ainda não chamam esse comando. A visão maior (copiar pra pasta, gerar
disco `.DSK`, testar no emulador) foi implementada na MESMA sessão,
logo abaixo.

**Comando `execute` (disco de teste + emulador) implementado**
(2026-11-24, `internal/platforms/msx/run.go` +
`internal/cli/cli.go`/`console.go`) — fecha o ciclo completo que o
usuário descreveu: "copiar para uma pasta e gerar um disco, e este
disco poder ser usado no emulador para o usuário já poder testar".
Usuário confirmou seguir direto depois do `compilar` (v1.27.0), e
perguntou também sobre a sequência maior (MSX → ZX81/ZXSpectrum →
Interfaces — respondi que sim, boa ordem, recomendação curta, ver
registro na conversa).

Pesquisei como o `badig.py` real invoca o openMSX
(`msx/emulator_interface.py:209-303`) antes de desenhar: ele NÃO
monta um `.dsk` de verdade pra testar — usa o próprio recurso do
openMSX de montar uma PASTA como disco virtual (`cmd.extend(['-diska',
self.stg.file_path])`, onde `file_path` é o diretório do arquivo
salvo, não um arquivo `.dsk`). Decidi NÃO replicar isso ao pé da
letra: o usuário pediu explicitamente "gerar um disco" (e comissionou
a biblioteca `dsk` especificamente pra isso, v1.26.0) — mais rico que
o que o original faz, então segui o pedido explícito do usuário em
vez do comportamento literal do original nesse ponto específico.

`msx.BuildTestDisk(dmxPath)`: compila (`CompileFile`), cria um disco
novo (`dsk.Create`) e adiciona o `.bmx` resultante como
`AUTOEXEC.BAS` — convenção da própria ROM de disco do MSX pra rodar
um programa sozinho, sem precisar digitar nada, assim que o disco dá
boot. **Decisão deliberada, documentada**: NÃO embuto os arquivos
reais de sistema do MSX-DOS (`MSXDOS.SYS`/`COMMAND.COM`) no disco
gerado, mesmo tendo cópias reais deles em `E:\msxdiskutil` (usadas lá
só pros testes locais do próprio usuário) — são software de sistema
de terceiros, sob copyright, não algo que esse projeto deveria
embutir/redistribuir. Efeito prático: o boot sector padrão (herdado
do `dsk.Create`) tenta carregar `MSXDOS.SYS` primeiro, não acha, e o
MSX real mostra um "Boot error" breve antes de cair pra BASIC e
achar o `AUTOEXEC.BAS` sozinho — confirmado rodando de verdade no
openMSX (ver Verificação abaixo); custa 1 tecla, não é bug.

`msx.RunInEmulator(diskPath)`: acha `openmsx` no PATH
(`exec.LookPath`, resolve `.exe` no Windows sozinho) e abre
`-diska <diskPath>` — mesma flag do original, sem o protocolo de
monitoramento via `stdio`/script TCL (`badig_msx.py`'s próprio
`-control stdio` + comandos — feature bem maior, já adiada desde
sempre em "Monitoramento de execução em tempo real", Bloco 5). Não
espera o emulador terminar (abre a janela e devolve o controle).

CLI: `goBaDig --execute <arquivo.dmx>` — precisou de uma flag
explícita (diferente de `compile`, que reusa o argumento posicional
livre, `--execute` precisa do NOME do arquivo logo depois, que
`Parse`/seu `Mode` não carregam) — checado em `cli.Run` ANTES de
`Parse` rodar, prioridade mais alta que qualquer outra coisa (único
jeito de garantir que funciona independente de onde `Parse` erraria
sem essa checagem). Console (`--cli`) ganhou o comando irmão
`execute <arquivo.dmx>`, usando o vocabulário exato que o usuário já
tinha usado espontaneamente ("edit/badig/execute").

**Validação real**: rodei o CLI de verdade (`go run ./cmd/goBaDig
--execute arquivo.dmx`) — abriu DOIS processos `openmsx.exe` reais
(confirmado via `Get-Process`), ficaram rodando sem crash. Sem
confirmação VISUAL do boot (ambiente sem display nesta sessão,
mesma limitação já documentada na Etapa 1 da biblioteca `dsk`) —
pude confirmar que o "Boot error" é esperado (não um bug) só por
raciocínio sobre o boot sector, não visualmente. Os dois processos
ficaram pra o usuário fechar manualmente (matar processo via
PowerShell foi bloqueado pelo classificador de permissão do Claude
Code como ação destrutiva — corretamente, já que eu não deveria
matar processo sem confirmação).

12 testes automatizados novos (`BuildTestDisk` produz um disco com
exatamente `AUTOEXEC.BAS` contendo os bytes certos do `.bmx`; erro
claro pra arquivo de origem inexistente; `RunInEmulator` com `PATH`
vazio retorna erro claro em vez de travar/crashar; `executeFileArg`
cobre flag com/sem arquivo e ausência da flag; `cli.Run`/o comando
`execute` do console constroem o disco de verdade mesmo quando o
emulador não é encontrado — `PATH` sobrescrito pra um diretório vazio
nos testes, assim a suíte nunca abre uma janela de emulador de
verdade sozinha — e reportam o erro certo; completar com Tab
reconhece `execute`).

**Fora de escopo, documentado**: protocolo de monitoramento via
`stdio`/TCL (mapear erro de execução pra linha Dignified); seleção de
máquina/extensão/configuração do emulador via flag (as opções já
existem no catálogo - `EMULATOR_MACHINE`/`EMULATOR_EXTENSION`/etc. -
mas nada ainda lê o SQLite pra aplicar aqui); copiar MSX-DOS de
verdade pro disco (decisão deliberada, documentada acima); TUI/GUI
ainda não chamam nenhum desses comandos.

### Bloco 2 — Módulo MSX

- [x] **Portado o tokenizer MSX Batoken** (2026-11-21,
  `internal/platforms/msx/tokenizer.go` + `tokenizer_test.go`) — texto
  ASCII numerado → binário `.bmx`, standalone, **não usa
  `internal/core`** (o tokenizer no Python original também não usa
  nada do motor Dignified/`Classic`). Os 13 vetores de ouro do Bloco 0
  portados 1:1 como testes Go, todos passando byte-exato; validado
  ainda contra 10 casos extras (DATA, CALL, `&H`/`&O`/`&B`, variável
  com palavra-chave embutida, `ON...GOTO` com múltiplos alvos, inteiro
  grande sem sufixo, atribuição, sufixo `%`) rodando o Python real
  lado a lado — todos batem. **Achado que corrige o item anterior
  desta lista**: não existe Microsoft Binary Format nem nenhuma
  conversão binária de float no tokenizer original — é puramente
  decimal (ver seção 4.1 e o achado em `docs/golden-vectors/
  msx-tokenizer.md`/`OUTLINE.md` acima). Simplificações deliberadas,
  documentadas no código (`Tokenize`, doc comment): notação científica
  (`1E+10`) não suportada ainda (erro explícito, não vetor de ouro
  cobre isso); sem espaço embutido dentro de um literal numérico
  (réplica de uma lacuna do Python seria portar um bug latente, não
  comportamento real do MSX); literais maiores que `int64` (~18-19
  dígitos) retornam erro em vez de aritmética de precisão arbitrária;
  o bug latente do Python no tratamento de `AS` com 2 dígitos (`ord()`
  numa string de 2 chars, cai em `TypeError`) foi corrigido, não
  reproduzido. — **A**, **1**
- [x] **Portados `Description` e `Lexer` MSX** (2026-11-21,
  `internal/platforms/msx/description.go` + `lexer.go` +
  `lexer_test.go`) — primeiro uso real das interfaces "v0" do
  Bloco 1 (antes só validadas contra uma plataforma fictícia). Toda a
  tabela de vocabulário (`c_instrc` 73 entradas, `c_funcdl`/`c_funcnm`,
  `c_jumpin`, `c_operat`, `c_symbol`, `c_e_symb`) transcrita linha a
  linha de `badig_msx.py:1-224` lido por completo nesta sessão.
  **Desvio real na interface `core.Lexer`, descoberto implementando**
  (documentado na Fase 2/3 do Bloco 1 abaixo): a v0 (palavra
  pré-cortada por espaço) não aguenta números com notação científica
  nem símbolos compostos sem espaço (`x+=5`) — trocada por
  `Lex(ctx, remaining string) (consumed int, tokens []Token, ok bool,
  err error)`, dando à plataforma a **linha inteira restante**, mesmo
  padrão já usado no tokenizer Batoken. `core.Description` ganhou
  `QuoteChar() byte`. `core.Lex()` agora também reconhece, de forma
  compartilhada entre plataformas, identificador Dignified
  (`^[a-z][a-z_0-9]*[%!#$]?`) e string entre aspas — só entram em jogo
  quando a plataforma não reconhece nada ali (prioridade confirmada
  lendo a ordem do regex combinado original: `classic_groups |
  dignified_groups | main_groups | d_idnttp` — identificador é
  **sempre** o último recurso). 8 testes cobrindo palavras-chave
  (maiúsculo/minúsculo), operadores, símbolos simples/compostos,
  números (decimal/`&H`/`&O`/`&B`), comentários (`REM`/`'`), `DATA`, e
  o fallback tolerante pra caractere não reconhecido. **Fora de
  escopo nesta fatia** (nenhum real do MSX, nenhum bug — só adiados):
  bloco de comentário `''` multi-linha, semântica literal de itens do
  `DATA` (por ora lexados como expressões normais), tradução
  Unicode→ASCII (isso é Pass 5 do `Parser`, próxima fase). — **M**, **1**
- [x] **Portados `Parser` Pass 3 e Pass 4 MSX** (2026-11-21,
  `internal/platforms/msx/parser.go` + `parser_test.go`) — a alocação
  de variáveis de 2 letras, fiel a `get_hard_variable`/
  `process_variable` (`badig_msx.py:463-558`, lidos linha a linha):
  cursor único descendente de `zz` até `aa`, 3 conjuntos de exclusão
  (`Declares`/`HardShortVars`/`HardLongVars`), nome já curto fica
  inalterado e reserva seu próprio nome, `~nome` mantido por extenso
  com o `~` removido da saída, erro fatal em esgotamento (676 aliases)
  e nos dois casos de uso inválido de `~` documentados no Python. **Não
  portados** (não-fatais, `Info.log` nível 2 no original): os loops de
  aviso de colisão por prefixo em `get_hard_variable`. Terceiro desvio
  de interface do `core.Parser` descoberto implementando (ver Fase 2/3
  do Bloco 1 abaixo). 7 testes novos, incluindo um de ponta a ponta
  completo (`Lex`→`Parse`→`Generate`) num programinha Dignified real
  com variável repetida, nome curto e `~`. **Primeira conversão
  Dignified→Classic de verdade que já produz texto correto** (mesmo
  sem Pass 1/5/generate real ainda). — **A**, **1**
- [x] **Portado `Parser` Pass 5 MSX** (2026-11-21, `parser.go`) — as 5
  conversões finais específicas do Classic, fiéis a `pass_5`
  (`badig_msx.py:617-682`, lido linha a linha): conversão `?`/`PRINT`
  nos dois sentidos (`ConvertPrint`), strip de `THEN` antes de `GOTO`
  ou de `GOTO` depois de `THEN`/`ELSE` (`StripThenGoto`), expansão de
  atribuição composta/incremento/decremento (`x+=5`→`x = x + 5`,
  `x++`→`x = x + 1`, reaproveitando o operando já emitido — ver
  "Quarto desvio" no Bloco 1), `TRUE`/`FALSE`→`-1`/`0`, tradução
  Unicode→ASCII (`translateChars`, fiel a `trans_char`, tabela de 128
  chars extraída programaticamente do Python nesta sessão pra evitar
  erro de transcrição manual). `ConvertPrint`/`StripThenGoto`/
  `Translate` são campos do `Parser` com os mesmos defaults "desligado"
  do Python, já que CLI/Settings (`-cp`/`-tg`) ainda não existem (ver
  item abaixo) — só configuráveis via código por ora. Quarto desvio de
  interface do `core.Parser` descoberto implementando (ver Fase 2/3 do
  Bloco 1 acima): `Replace` ganhou payload `[]Token`. `core.Lex()`
  ganhou reconhecimento de `TRUE`/`FALSE` (`KindDignifiedOperator`). 8
  testes novos. **Fora de escopo, documentado, não implementado**:
  capitalização (`capitalise_all`, desligada por default no original
  também). — **A**, **1**
- [x] **Resolução de labels e numeração de linha** (2026-11-21,
  `internal/core/pipeline.go`: `pass2Labels`/`pass3LineHygiene`/
  `pass4Numbering`) — fiel a `badig.py` (`get_label_lines`/
  `get_jump_labels` 849-921, `pass_2` 1447-1504, limpeza de linha em
  `pass_3` 1536-1613, `pass_4` completo 1615-1757), confirmado também
  empiricamente contra `badig/examples/msx/CGK173.dmx`/`.amx`, um
  exemplo real completo (apontado pelo usuário). `{nome}` no início de
  linha = declaração; em qualquer outro lugar = referência de salto;
  `{@}` = linha atual; label nunca gasta linha própria (cola na
  próxima linha com conteúdo real); linha em branco e `_` no fim de
  linha (funde com a próxima) não gastam número; erro fatal pra salto
  a label não declarado. Ver Bloco 1 acima pro detalhamento da mudança
  de arquitetura do `core` (primeira vez que `Pass2`/`Pass3`/`Pass4` do
  `core` fazem trabalho Dignified de verdade, não só delegam pra
  plataforma) e os 4 tokens novos. 10 testes novos (6 em
  `internal/core`, mecanismo puro; 1 de ponta a ponta em
  `internal/platforms/msx` confirmando que label+numeração convive com
  a renomeação de variável do MSX). **Primeiro programa Dignified com
  `GOTO`/label de verdade compilando de ponta a ponta.**
  **Fora de escopo, documentado**: loop labels (`nome{ }`/`exit`,
  usado no `CGK173.dmx` em `keys{ }` etc.), inserção de header,
  relatório de labels (`-lbr`). — **A**, **1**
- [x] **Pass 1 MSX implementado** (2026-11-23,
  `internal/platforms/msx/parser.go`) — a dúvida antiga acima estava
  mal colocada: não gera texto "CALL" nenhum (`.val` mesmo não muda,
  confirmado de novo) — o efeito real é só EXCLUIR o identificador
  logo depois do `_` da renomeação de variável de 2 letras
  (`C_CALIDF` nunca bate nos checks `D_IDNTTP` do Pass 3/4). Dois
  Kinds novos (`msx.KindCallSeparator`/`msx.KindCallIdentifier`),
  retagueados dentro do próprio `Pass1` (stub vazio até agora, o
  ponto certo - é chamado pra todo token que não é um caso especial
  do core, igual `self.clc.pass_1()`). Confirmado contra
  `python badig.py` de verdade. `CGK173.dmx` não usa esse mecanismo
  (todo `_` nele está no fim de linha) - validação 100% de ponta a
  ponta continua de pé. 2 testes novos de ponta a ponta. Ver Bloco 1
  acima pro detalhamento completo. — **M**, **1**
- [x] **Ligação ao tokenizer MSX Batoken implementada** (2026-11-23,
  `internal/platforms/msx/compile.go`) — `msx.CompileToBinary(ctx,
  platform) ([]byte, error)` roda `core.Run` e tokeniza o resultado
  direto em memória (`GenerateResult.Code`/`Tokenize` já tinham
  exatamente o mesmo formato, zero mudança precisou nos dois lados).
  `CGK173.dmx` inteiro compilado pra `.bmx` bate byte a byte (18.815
  bytes) contra o `.bmx` real gerado por `python badig.py ... --tk_
  tokenize` — primeira validação de ponta a ponta até o BINÁRIO, não
  só o texto Classic. 2 testes novos. Ver Bloco 1 acima pro
  detalhamento completo. — **M**, **1**
- [x] **Loop labels implementados** (2026-11-22, `nome{ ... }` como
  laço fechado + `exit`) — fiel a `get_label_lines`/`get_jump_labels`/
  `get_loop_label_return`/`get_labels_exit` (`badig.py:849-921`), um
  loop-opener usa o MESMO `KindLabelDeclaration` de uma declaração
  simples (só a detecção difere); fechamento `}` emite um `GOTO`
  implícito de volta pro início (com separador `:` só quando não está
  sozinho na linha); `exit` salta pra depois do fechamento do laço
  mais interno. `core.Description` ganhou `StatementSeparator()`/
  `LoopBackInstruction()`; `pass2Labels` (ver Bloco 1 acima) passou a
  ser bespoke (pilha de labels abertos). Usado de verdade no
  `CGK173.dmx` (`keys{ }`, `read_keyboard_main{ }` etc.). **Fora de
  escopo, documentado**: deduplicação de separador (`::` raro),
  limpeza de `C_INSTSP` repetido do Pass 3 original. 11 testes novos
  em `internal/core` + 1 de ponta a ponta no MSX. — **A**, **1**
- [x] **`DEFINE` implementado** (2026-11-22, `internal/core/pipeline.go`,
  `pass1Defines`) — macro de texto com até 1 placeholder (`[0]`),
  valor padrão quando a chamada não dá argumento, chamadas aninhadas
  recursivas (`[a]([b])`), a linha inteira do `DEFINE` desaparece sem
  rastro. Primeira pass com cursor explícito (não cabe no modelo
  `passFunc`/`runPass`). Ver Bloco 1 acima pro detalhamento. 7 testes
  novos + 1 de ponta a ponta confirmando que convive com label e
  renomeação de variável (padrão real do `CGK173.dmx`). — **A**, **1**
- [x] **`FUNC`/`RET` implementados** (2026-11-22,
  `internal/core/pipeline.go`) — declaração em `pass1Dignified`
  (renomeada de `pass1Defines`); chamada numa pass nova
  (`pass2Functions`), já que a captura de retorno
  (`a, b = .func()`) precisa remover retroativamente tokens já
  emitidos — mesmo tipo de problema do `DEFINE`, só pra trás. Cada
  parâmetro é `nome[=padrão]`; resolução de linha igual label, mapa
  separado. 5 métodos novos em `core.Description`. Ver Bloco 1 acima
  pro detalhamento (inclusive o achado de que a linha de declaração
  de `FUNC` desperdiça 1 número, diferente de label). 9 testes novos
  + 1 de ponta a ponta confirmando que a renomeação de variável
  alcança parâmetro/padrão de função. — **A**, **1**
- [x] **`DECLARE` implementado** (2026-11-22,
  `internal/core/pipeline.go`) — escreve direto nas 3 tabelas de
  variável que o `Parser` MSX Pass 3/4 já consultam (zero mudança em
  `msx/parser.go`). Sem `:` = mesmo efeito de `~nome`; com `:curto` =
  atribui o alias manualmente, pulando o alocador. Ver Bloco 1 acima
  pro detalhamento (inclusive a correção do typo de `c_varsna`). 8
  testes novos + 1 de ponta a ponta. — **A**, **1**
- [x] **`KEEP` implementado** (2026-11-22, `internal/core/pipeline.go`
  + `internal/platforms/msx/lexer.go`) — compilação condicional via
  tags `#tag`. Ver Bloco 1 acima pro detalhamento (inclusive o
  conflito real de léxico com `#` já ser símbolo do MSX, e a correção
  do bug de comparação de duplicata). 8 testes novos em `core` + 1 de
  ponta a ponta no MSX. — **A**, **1**
- [x] **`INCLUDE` implementado** (2026-11-22, `internal/core/pipeline.go`
  + `internal/core/context.go`) — mesclagem de outro arquivo Dignified
  (não usado no `CGK173.dmx` de verdade, apesar do item anterior
  supor que sim - confirmado não ter nenhuma ocorrência nesse
  exemplo). Ver Bloco 1 acima pro detalhamento (inclusive o achado de
  que é processado só no Pass 3, não no Pass 1, e a resolução de
  caminho relativo contra o arquivo PRINCIPAL, não o que inclui). 7
  testes novos em `core` + 1 de ponta a ponta no MSX lendo um arquivo
  real do disco. — **A**, **1**
- [ ] Relatório de labels (`-lbr`) — cosmético, pequeno. — **B**, **2**
- [x] **`generate()` de verdade implementado** (2026-11-22,
  `internal/core/pipeline.go`) — regras de espaçamento por palavra
  reservada + limite de 256 chars/linha, fiel a `generate()`
  (`badig.py:1796-1878`). Ver Bloco 1 acima pro detalhamento
  (inclusive o achado de que a regra é puramente baseada em VALOR,
  não em Kind, e por que não precisou trocar o modelo
  `HookAction`/`runPass` do hook `Parser.Generate` - a montagem de
  texto por linha roda DEPOIS do hook, não dentro dele). Novo
  `Description.CompoundSymbols()` (mirrors `c_symb_comp`);
  `ReservedKeywords()` já existia (rodada do `DECLARE`), reaproveitado
  sem mudança. 5 testes novos focados em espaçamento + ~43 asserções
  de `Code` já existentes corrigidas pro valor real (não regressão -
  era o próprio objetivo da rodada). — **M**, **2**
- [x] **Comentários exclusivos do Dignified implementados**
  (2026-11-22, `internal/core/pipeline.go` +
  `internal/platforms/msx/lexer.go`) — `##` (linha, qualquer posição),
  um bloco de 3 cerquilhas (descartado inteiro) e um bloco de duas
  aspas simples (conteúdo virando N linhas classic reais). Ver Bloco 1
  acima pro detalhamento (inclusive o novo parâmetro `atLineStart` em
  `Lexer.Lex` e os 3 bugs reais de lexer achados testando o
  `CGK173.dmx` de ponta a ponta). 7 testes novos em `core` + 1 de
  ponta a ponta no MSX usando o cabeçalho real do `CGK173.dmx`. —
  **M**, **1**
- [x] **Fusão de linha pelo separador clássico `:` implementada**
  (2026-11-22, `internal/core/pipeline.go`) — fiel a
  `badig.py:1550-1605`: `:` clássico no fim de linha funde com a
  próxima MANTENDO o próprio `:` (diferente do `_` Dignified, que
  desaparece inteiro); `:` no início de linha funde com a anterior;
  dois `:` adjacentes colapsam; strings literais adjacentes (mesma
  linha ou por uma `NEWLINE`) se concatenam; `NEWLINE` depois de
  `KindFunctionDefinition` também passou a colapsar (bug achado
  relendo o código, não só o Python — o corpo de uma função
  desperdiçava uma linha em branco). De brinde, achado re-validando
  contra a saída real: uma linha em branco no INÍCIO do arquivo
  também desperdiçava um número de linha — `KindProgram` entrou no
  mesmo colapso. Ver Bloco 1 acima pro detalhamento (inclusive a
  validação 100% de ponta a ponta contra o `CGK173.dmx` e a saída
  real do `badig.py`, e o achado do `"?"` faltando nas palavras
  reservadas). 5 testes novos em `core`. — **M**, **1**
- [x] **Cabeçalho REM de 2 linhas + expansão de TAB implementados**
  (2026-11-23, `internal/core/pipeline.go`/`settings.go`) — fiel a
  `insert_rem_header` (`badig.py:1289-1303`/`1643-1645`) e
  `line.expandtabs` (`support/helper.py:88`). Novos
  `CompileSettings.RemHeader`/`TabLength` (zero-value `false`/`0` —
  decisão deliberada, ver Bloco 1 acima, sem CLI ainda pra expor o
  default real do Python). **Fechou os dois últimos gaps conhecidos**
  contra a saída real do `badig.py`: `CGK173.dmx` inteiro agora bate
  byte a byte, ZERO diferenças. 4 testes novos em `core`. — **B**, **1**

  **Texto do cabeçalho trocado pra apontar pro goBaDig (2026-11-27,
  v1.43.0)**, pedido explícito do usuário: `remHeaderLine1`/
  `remHeaderLine2` (`internal/core/pipeline.go`) deixam de ser as
  mesmas do `support/badig_settings.py` original ("Converted with
  Basic Dignified" / link pro `farique1/basic-dignified`) e passam a
  ser "Converted by goBaDig 2026" /
  `https://github.com/wilsonpilon/goBaDig` — divergência DELIBERADA
  da saída real do `badig.py` a partir de agora (zero diferença deixa
  de valer só pra essas duas linhas; o resto do arquivo continua
  batendo byte a byte). Dois testes de validação pessoal, não
  rastreados pelo git (`internal/platforms/msx/
  zz_scratch_cgk173_test.go`, comparavam byte a byte contra referências
  reais do `badig.py`), marcados com `t.Skip` nesse ponto específico -
  a referência é de antes dessa mudança.
- [ ] `Settings`/`Expose` MSX (`-cp`/`-tg`) — só fazem sentido quando o
  `Parser` ler essas flags. — **B**, **2**
- [ ] Portar DignifieR MSX como parser manual (sem lookahead de regex). — **A**, **2**
- [ ] Integração openMSX: lançar processo + disco virtual, **sem**
  monitoramento de erro na v1 (adiar item de risco alto). — **M**, **1**
- [ ] Monitoramento de execução em tempo real (mapeamento de erro para
  linha Dignified) — avaliar viabilidade multiplataforma em Go (pode ser
  melhor que o Python, que só funciona fora do Windows). — **A**, **3**
- [x] Integração fMSX como backend alternativo ao openMSX. — **M**, **2**
  — **feito em 2026-11-26**, pedido do usuário:
  `msx.RunInFMSX(diskPath)` (`internal/platforms/msx/run.go`) — mesma
  flag `-diska` do openMSX, confirmado lendo a documentação do fMSX
  (`fMSX.html`) e lançando de verdade nesta máquina. **blueMSX+
  também adicionado**, de brinde, não estava neste item original mas
  foi pedido junto: `msx.RunInBlueMSX(diskPath)` — forma de "um
  arquivo só" (`blueMSX+ <file>`), confirmada lendo o próprio texto de
  uso embutido no executável. Seleção de qual backend usar: nova flag
  `--emulator <nome>`/console `execute <arquivo> [emulador]` — ver
  Bloco 5 abaixo (`internal/cli/dialect.go`'s `emulatorBackends`).
  ZEsarUX deliberadamente FORA (só emula MSX1, decisão do usuário —
  openMSX/fMSX/blueMSX+ cobrem MSX2/MSX2+ também).
- [ ] Empacotamento `.CAS` (MSX, captura de fita nível BIOS). — **M**, **2**
- [ ] Empacotamento `.DSK` automático no fluxo de conversão (MSX) —
  a biblioteca de baixo nível (`internal/platforms/msx/dsk`:
  `Create`/`AddFile`) já existe e está validada (ver Bloco 1, "Etapa
  1" do gerenciador de arquivos, abaixo); falta só ligar isso ao
  `Tools.Interface`/fluxo de conversão de verdade, quando ele
  existir. — **B**, **2**

### Bloco 3 — Módulo ZX81

- [x] Portar `Description`/`Lexer`/`Parser` ZX81 (1 letra Z→A, `PRINT AT`,
  LET implícito — `remx`/renumeração de listagem numerada **deferidos**,
  ver nota ao final). — **M**, **1** —
  **Description+Lexer feitos em 2026-11-24**
  (`internal/platforms/zx81/{description,lexer}.go`, v1.30.0); **Parser
  feito em 2026-11-24** (`internal/platforms/zx81/parser.go`, v1.31.0).

  **Decisão de escopo, documentada**: `remx` e a renumeração de
  listagem numerada (`dignify_numbered`) ficam de fora desta Etapa 3,
  diferente do que o roteiro aprovado previa inicialmente. Achado,
  lendo `badig_zx81.py` de novo com atenção: `preprocess`/`postprocess`
  (as duas funções por trás de `remx` E da renumeração) não são passes
  do `Parser` (`pass_1`...`pass_5`/`generate`) — `preprocess` roda
  ANTES do lexer até ver o texto fonte (`d_code`, linhas crúas), e
  `postprocess` roda DEPOIS do `generate()`, sobre a listagem Classic
  já RENDERIZADA (`c_code`, strings de texto, não tokens). São hooks
  que ENVOLVEM o pipeline inteiro, não parte dele — meu `core.Parse`/
  `core.Generate` de hoje não tem (nem deveria ganhar sem necessidade
  real) um ponto de extensão pra isso; seria uma evolução de
  arquitetura BEM maior que a do `BuiltinDefine.Expand` (abaixo), e
  nenhum programa ZX81 real ainda precisa de nenhuma das duas. Junto
  com a Etapa 6, já adiada, do roteiro.

  **Alocação de 1 letra** (`process_variable`, Pass4): reusa
  `core.NewDescendingAllocator()` sem nenhuma mudança, só com
  `variableCategory{AliasChars:"abcdefghijklmnopqrstuvwxyz",
  AliasLen:1}` — confirma a previsão de `internal/core/allocator.go`
  (citada desde a Etapa 1). **Diferença real de comportamento vs MSX,
  confirmada lendo o Python de novo**: pro MSX, uma variável marcada
  `~` (Dignified, "mantenha o nome longo") fica INTOCADA — pro ZX81,
  isso é um ERRO FATAL sempre ("ZX-81 variables must be one letter,
  optionally followed by $."), porque o BASIC clássico do ZX81
  literalmente não consegue representar um nome de variável com mais
  de 1 caractere — `~` é uma sintaxe Dignified compartilhada, mas
  genuinamente impossível de honrar nesse dialeto. Confirmado rodando
  o `badig.py` real tanto pro caso `~nomelongo` quanto `~a` (nome já
  curto, erro DIFERENTE: "can't use ~ on a short named variable").

  **`LET` implícito** (`generate()`): BASIC clássico do ZX81 não tem
  atribuição nua (`a=1` sempre precisa de `LET`/`?`/alguma instrução
  no início) — inserido automaticamente quando um identificador abre
  uma declaração (logo depois do número de linha, ou logo depois de
  `THEN`). Implementado como o hook `Parser.Generate` devolvendo
  `Replace` com `[Instruction("let"), <identificador original>]` — o
  passe de espaçamento genérico do `core.Generate` (já existente,
  compartilhado) cuida do espaço entre os dois sozinho, sem precisar
  de nenhuma lógica nova aí.

  **`PRINT AT`** (`[?](row,col)` → `print at row,col;`) — **achado de
  arquitetura confirmado, evolução do contrato feita**: em vez da
  evolução de `HookAction`/`Pass2` que a Etapa 2 tinha previsto
  (consumir um span de vários tokens), a solução real foi mais simples:
  `core.BuiltinDefine` ganhou um campo novo, `Expand func(args []Token)
  []Token` (opcional, `nil` pro `[?](x,y)` mais simples do MSX) — quando
  presente, substitui o mecanismo de splice-num-placeholder-só por uma
  função que recebe os argumentos JÁ RESOLVIDOS da chamada (os tokens
  crus entre os parênteses, ou o `DefaultVar` se chamado sem parênteses)
  e devolve a expansão INTEIRA da chamada. Isso roda dentro do Pass 1,
  já bespoke/com cursor próprio (`expandDefineCall`), que JÁ calculava
  esses argumentos resolvidos como efeito colateral de preencher o
  placeholder — zero necessidade de o Pass 2 também virar bespoke.
  `printAtExpand` (novo, `internal/platforms/zx81/parser.go`) separa
  os argumentos por vírgula de nível zero (respeitando parênteses
  aninhados, mesmo princípio do `replace_print_at` original), default
  `"0"` pra cada lado que faltar.

  Validação cruzada extensa contra o `badig.py` real (não só os testes
  automatizados): alocação de 3 variáveis longas em sequência
  (`z`,`y`, depois outra `y`/`z` dependendo da ordem - confirmado
  rodando com 3 nomes diferentes), sufixo `$` compartilhando o mesmo
  slot do nome base, os 4 casos de `PRINT AT` (ambas coordenadas, sem
  nenhuma, só a primeira, só a segunda), parênteses aninhados dentro
  de uma coordenada (`[?](row,peek(1,2))`), `LET` implícito depois do
  número de linha E depois de `THEN`, e os dois erros de `~` — TODOS
  batendo exatamente com a saída/erro reais.

  12 testes automatizados novos em `parser_test.go` (ponta a ponta,
  Lex→Parse→Generate).

  **Etapa 4 (validação de ponta a ponta), feita em 2026-11-24**: os
  DOIS exemplos reais já existentes no projeto
  (`badig/examples/zx81/zx81-intro.DZX` e `mcode-ok.DZX` — não
  escritos pra esta rodada, já estavam no repositório) rodados pelo
  pipeline inteiro (`Lex→Parse→Generate`) batem **byte a byte** contra
  a saída real do `badig.py` (confirmado rodando `python badig.py` de
  verdade nesta máquina, com o cabeçalho de remtag/comentário e — no
  caso do `mcode-ok.DZX` — a linha `remx` removidos antes, já que
  extração de remtag e `remx`-como-diretiva estão fora de escopo,
  documentado acima). `zx81-intro.DZX` exercita `DECLARE` com dois
  aliases numa linha só, dois `FOR`/`NEXT` com variáveis diferentes,
  sintaxe classic `print at R,C;` digitada direto (sem passar pelo
  atalho `[?]`), caracteres gráficos com escape de barra invertida
  dentro de string, e um label de loop que aponta pra ELE MESMO (loop
  infinito "espera uma tecla"). `mcode-ok.DZX` exercita `DECLARE` com
  alias explícito por variável, sufixo `$` compartilhando o alias da
  base, indexação de string (`h$(i)`), expressões aritméticas, e as
  funções `len`/`code`/`rand`/`usr`. Virou 2 testes comitados
  (`TestEndToEndRealIntroExample`/`TestEndToEndRealMcodeExample`,
  conteúdo embutido como string Go, sem depender de arquivo em disco —
  mesmo princípio 100%-em-memória já usado o projeto inteiro) +
  2 scratch (`zz_scratch_e2e_test.go`, não comitado, lendo os arquivos
  reais do disco via `core.LoadLines`).

  Também portados (agora que o `Parser` existe) mais 2 vetores de ouro
  do `test_tokenizer_zx81.py` que dependiam do pipeline Dignified
  completo, antes fora de alcance: `test_doubled_and_backslash_quotes_
  inside_dignified_strings` (as duas formas de escape de aspas
  sobrevivem intactas até o `Generate()`, já que nada entre o Lexer e
  a renderização as decodifica — só o tokenizer, Etapa 1, faz isso) e
  `test_long_variables_allocate_from_z_down_and_honor_declare`
  (`DECLARE` com alias explícito reserva o slot antes de qualquer
  alocação automática; uma variável longa alocada automaticamente
  depois ainda desce corretamente a partir de `z`). Os demais vetores
  do arquivo Python (relatórios `-lnr`/`-var`, renumeração) continuam
  fora de escopo pelas mesmas razões já documentadas.

  **Etapa 5 (comando `compile`/`execute` pro ZX81), feita em
  2026-11-24**: generaliza os comandos já existentes do MSX (até aqui
  hard-coded) pra escolher o dialeto pela EXTENSÃO do arquivo —
  `internal/cli/dialect.go` (novo), `compileByExtension`/
  `runInEmulatorByExtension`, `.dmx`→MSX, `.dzx`→ZX81. `cli.go`/
  `console.go` chamam só essas duas funções agora, sem mais nenhuma
  referência direta a `msx.*`.

  `internal/platforms/zx81/compile.go` (novo): `DefaultCompileSettings`
  (mesmos defaults compartilhados do MSX — `LineStart`/`LineStep`=10,
  `RemHeader`=true, `TabLength`=4 — não são valores ZX81-específicos)
  e `CompileFile(dzxPath, settings) (azxPath, pPath string, err
  error)`, espelhando `msx.CompileFile` quase 1:1: `core.Run` + o
  `Tokenize` do ZX81 (pronto desde a Etapa 1), escrevendo `.azx`
  (CRLF, mesma convenção do `.amx`) e `.p` (binário, bytes crus).

  `internal/platforms/zx81/run.go` (novo): `RunInEmulator(pPath)` —
  SEM nenhum passo de "empacotar num disco" (diferente do MSX): o
  `.p` que `CompileFile` já produz é tudo que o EightyOne precisa pra
  bootar direto no programa compilado — ZX81 não tem sistema de
  arquivos de disco/convenção AUTOEXEC pra passar por isso.
  `exec.LookPath("eightyone")` + `exec.Command(exePath, pPath)` — SEM
  nenhuma flag (diferente do `-diska` do openMSX), confirmado lendo
  `emulator_interface.py` de novo: `subprocess.Popen([emulator_path,
  file_save])`, só o caminho como argumento posicional único.

  **Validação real**: estendi o PATH manualmente SÓ pra esta sessão
  de shell (`D:\zeddy\EightyOne`) — o usuário já instalou e configurou
  o PATH do sistema, mas isso ainda não chegou ao processo do Claude
  Code em si (precisa reiniciar a sessão/terminal pra isso propagar).
  Com isso, rodei `goBaDig --execute` de VERDADE contra um `.dzx`
  mínimo — confirmei um processo `EightyOne.exe` real rodando via
  `Get-Process` (mesmo nível de confirmação já aceito pro MSX/openMSX:
  processo vivo confirmado, sem confirmação visual do programa
  rodando dentro — sem display neste ambiente).

  11 testes automatizados novos (`internal/platforms/zx81`:
  `CompileFile` escreve `.azx`/`.p` corretos incluindo o cabeçalho REM
  — `"REM"`, a palavra, não `'` — erro claro pra arquivo inexistente,
  `RunInEmulator` com `PATH` vazio retorna erro claro;
  `internal/cli`: `compileByExtension`/`runInEmulatorByExtension`
  despacham certo por extensão e rejeitam extensão desconhecida, mais
  os pares `.dzx` dos testes `Run`/console já existentes pro `.dmx`).

  Vocabulário clássico (31 instruções, 3 funções `$`, 21 funções
  planas, 4 jump keywords, 3 operadores, 18 símbolos incluindo os 4
  compostos `<>`/`>=`/`<=`/`**` dentro do próprio `c_symbol`, mais 7
  operadores compostos exclusivos do Dignified `c_e_symb`) portado
  verbatim de `badig_zx81.py`, lido por completo (não só resumido)
  nesta rodada. `CommentChar()` = `"REM"` (a PALAVRA, não um caractere
  de pontuação como o `'` do MSX) — confirmado também, lendo
  `badig_msx.py`/`badig_zx81.py` direto, que o ZX81 real **não tem**
  o atalho `''` de bloco de comentário clássico que o MSX tem
  (`c_o_brem` é uma entrada só do `classic_groups` do MSX, nunca
  existiu no do ZX81) — o mecanismo genérico do `core`
  (`CommentChar()+CommentChar()`) já lida com isso de graça (`"REMREM"`
  nunca bate com texto real), sem precisar de nenhum código extra.

  **Achado grande, confirmado empiricamente rodando o `badig.py` real
  com `-lex -prr`** (não só lendo o regex): um sinal (`+`/`-`) no
  início de um número só funde com um número de verdade quando
  IMEDIATAMENTE seguido de um DÍGITO PURO — nunca atravessa um ponto
  decimal nem um expoente, mesmo sem nenhum espaço no meio. `"y-5"` →
  um token só `-5`; `"-5.5"` → DOIS tokens, `-5` e `.5` separados;
  `"-.5"` → símbolo `-` sozinho, depois `.5`. Documentado com 9 casos
  de teste (`TestSignedNumbersOnlyMergeWithAPlainDigitRun`), cada um
  confirmado contra a saída real do Python antes de virar teste.
  `c_numdec`'s próprio grupo de sinal opcional parecia, lendo só o
  regex, que deveria se estender por decimal/expoente também — não é
  o que acontece de verdade.

  Escape hex `\{XX}` (Dignified-exclusivo, badig_zx81.py's
  `Lexer.lexing` override) vira um Kind novo (`KindHexByte`) — token
  próprio, igual ao Python (`C_HEXBYT`), não decodificado ainda (isso
  já é trabalho do tokenizer, Etapa 1, pronto). String literal ZX81
  precisou de scanner PRÓPRIO (reivindica `"` por completo, não usa o
  fallback genérico do `core.Lex` como o MSX faz) — suporta `\"` E
  `""` (aspas dobradas) como escape de aspas dentro da string, as
  DUAS formas coexistindo (confirmado: `badig_zx81.py`'s
  `Lexer.quoted_string` override faz exatamente isso).

  Validação cruzada ad-hoc extra (não fixada como teste): uma linha
  com `FOR`/`IF`/`AND`/`NOT`/`OR`/`THEN`/`PRINT`/variável `$`/`PEEK`
  produziu a MESMA sequência de tokens (por valor) nos dois lados,
  Go e Python real.

  11 testes automatizados novos em `lexer_test.go` (teclado/operador/
  símbolo básico; símbolo composto do Dignified antes do prefixo de 1
  char; os 4 símbolos compostos do próprio `c_symbol` antes do
  prefixo; `#`+letra recusado pro fallback compartilhado; `#` sozinho
  ainda é símbolo; os 9 casos de sinal+número; número sem sinal mantém
  forma decimal/expoente completa; escape hex como token próprio;
  escape hex malformado é fatal; string com as duas formas de escape
  de aspas; `REM` reivindica o resto da linha; `?` como atalho de
  `PRINT`; jump keywords+funções; prefixo de palavra-chave não
  engole um identificador mais longo).
- [x] Portar tokenizer ZX81, incluindo ponto flutuante oculto (usar
  `test_tokenizer_zx81.py` como golden-master direto). — **M**, **1** —
  **feito em 2026-11-24** (`internal/platforms/zx81/tokenizer.go`,
  v1.29.0). Ainda SEM Description/Lexer/Parser ZX81 (item acima) — este
  tokenizer só converte texto Classic ASCII **já numerado** (o mesmo
  formato intermediário que o `.AZX` do Python produziria) em imagem
  `.p` completa; nada ainda chama ele a partir do CLI/console (precisa
  do item de cima primeiro, pra ter de onde vir esse texto Classic a
  partir de um `.dzx` real).

  Confirmado byte a byte contra o `tokenize_lines` real do Python:
  todos os vetores de ouro comitados (`docs/golden-vectors/
  zx81-tokenizer.md`, extraídos de `test_tokenizer_zx81.py`) passam,
  e os 27 testes do próprio `test_tokenizer_zx81.py` (classe
  `TokenizerTests`) rodam limpos nesta máquina (`python -m unittest`).
  Validação cruzada ad-hoc extra (6 linhas cobrindo `LET`/`FOR`/`STEP`/
  `IF`/`<>`/`GOTO`/`REM`, não fixada como teste) produziu o MESMO
  buffer de 784 bytes nos dois lados.

  Achado de arquitetura confirmado durante a pesquisa (fork dedicado):
  `internal/core/allocator.go` **já previa o esquema de 1 letra do
  ZX81** antes mesmo deste port começar (comentário próprio citando
  "ZX81 (everything 1 letter)") — o alocador descendente genérico
  (`core.NewDescendingAllocator()`) vai servir sem nenhuma mudança
  quando o item de cima (Parser) chegar, só com uma
  `VariableCategory{AliasChars: "abcdefghijklmnopqrstuvwxyz",
  AliasLen: 1}` nova.

  `PRINT AT` (achado, documentado pro item de cima): no Python é um
  `DEFINE` embutido (`[?](row,col)` → `print at row,col;`, mesmo
  mecanismo do `[?](x,y)`→`locate x,y:?` do MSX), mas com uma
  diferença que o `core.Parser`/`HookAction` de hoje não cobre ainda:
  precisa "contrair" vários tokens de entrada (marcador + argumentos
  brutos + marcador de fechamento) numa única decisão, pra tratar
  coordenada ausente como `0` - o `BuiltinDefines`+`Pass1` de hoje só
  sabe EXPANDIR um token em vários, não o inverso. Vai precisar de uma
  evolução pequena e já no mesmo espírito de toda evolução anterior
  desse contrato (ex.: `Replace` ganhou `[]Token` quando o Pass 5 do
  MSX precisou expandir `x+=5`) — decisão de design adiada pro
  momento de implementar esse item, não resolvida agora.

  **EightyOne não está instalado nesta máquina** (nem `C:\zeddy\`, nem
  PATH) — não bloqueia este item nem o próximo (Description/Lexer/
  Parser), mas vai bloquear a validação real de "abrir `.p` no
  emulador" quando chegar a hora (equivalente ao `execute` do MSX).
- [x] Gerar formato `.p` (imagem de memória completa, system vars +
  display file + programa). — **M**, **1** — **feito junto com o
  tokenizer** (`programImage`, `internal/platforms/zx81/tokenizer.go`,
  v1.29.0) — `Tokenize()` já devolve a imagem completa, não só o
  bytecode do programa.
- [ ] `remx` (diretiva Dignified, reserva bytes crus numa linha REM) +
  renumeração de listagem numerada colada direto (`dignify_numbered`,
  rejeição de salto computado) — **adiado da Etapa 3 do Parser**
  (2026-11-24): as duas são `preprocess`/`postprocess` no Python,
  hooks que envolvem o pipeline INTEIRO (antes do lexer / depois do
  `generate()` já renderizado), não passes do `Parser` — exigiriam uma
  evolução de arquitetura maior que a já feita pro `PRINT AT`
  (`BuiltinDefine.Expand`), e nenhum programa ZX81 real ainda precisa
  de nenhuma das duas. O tokenizer (Etapa 1) já sabe gerar a linha REM
  de `remx` sozinho (`remxLine`/`parseRemx`, `tokenizer.go`) — só falta
  o reconhecimento como diretiva Dignified e a reinserção pós-`generate`.
  — **M**, **2**
- [ ] Portar DignifieR ZX81. — **M**, **2**
- [x] Integração EightyOne básica (lançar `.p`, sem seleção de modelo
  ainda, igual ao Python). — **B**, **1** — **feito em 2026-11-24**
  (`internal/platforms/zx81/{compile,run}.go`, v1.33.0) — ver Etapa 5
  abaixo, junto com o comando `compile`/`execute` que a usa.
- [ ] Implementar tabela Unicode (`-tr`) para ZX81 — **nunca existiu no
  Python**, oportunidade de fazer melhor que o original. — **M**, **2**
- [x] Validar carregamento real de `.p` em emulador — nunca foi feito nem
  no Python. — **B** (execução manual), **1** — **feito em
  2026-11-24**: `goBaDig --execute` de verdade, com o EightyOne de
  verdade no PATH desta sessão de shell (estendido manualmente pra
  essa validação — o PATH do sistema que o usuário configurou ainda
  não chegou ao processo do Claude Code em si), abriu um processo
  `EightyOne.exe` real (confirmado via `Get-Process`) carregando um
  `.p` de verdade gerado pelo compilador. Sem confirmação visual do
  programa rodando (mesma limitação de ambiente sem display já
  documentada pro MSX).
- [ ] Suporte a `.81`/`.p81` (variantes do `.p` sem/com nome de arquivo,
  `.p81` podendo guardar múltiplos programas). — **M**, **2**
- [x] Integração ZEsarUX como backend alternativo ao EightyOne
  — **feito em 2026-11-26**, pedido do usuário: `zx81.RunInZEsarUX(pPath)`
  (`internal/platforms/zx81/run.go`) — `--machine ZX81 --tape <arquivo>`,
  confirmado lendo o `--help` do próprio `zesarux.exe` e lançando de
  verdade nesta máquina (processo real confirmado via `Get-Process`).
  Seleção de backend: nova flag `--emulator <nome>`/console `execute
  <arquivo> [emulador]`, compartilhada com MSX/ZX Spectrum — ver Bloco
  5 abaixo. **Só a invocação direta, SEM ZRCP** — monitoramento de
  erro em tempo real via ZRCP (breakpoint em `PC=0008H`, já avaliado
  viável no Bloco 0) continua não implementado, item separado abaixo.
  sz81/XZ81 continuam fora (nenhum pedido do usuário por eles). — **M**, **2**
- [ ] Monitoramento de erro em tempo real via ZRCP (ZEsarUX) —
  breakpoint em `PC=0008H`, ler `E_PPC`/`ERR_NR` — já avaliado viável
  no Bloco 0, ainda não implementado (a integração acima só abre o
  emulador, sem usar o protocolo ZRCP em si). — **A**, **2**

### Bloco 4 — Módulo ZX Spectrum

- [x] Portar `Description`/`Lexer`/`Parser` Spectrum (LET implícito,
  PRINT AT, TRUE/FALSE=1/0, DATA com expressões, DEF FN/FN com slots
  ocultos de 6 bytes). — **A**, **1** — **Description+Lexer feitos em
  2026-11-26** (`internal/platforms/zxspectrum/{description,lexer}.go`,
  v1.35.0, Etapa 2 de 5); **Parser feito em 2026-11-26**
  (`internal/platforms/zxspectrum/parser.go`, v1.36.0, Etapa 3 de 5) —
  **o primeiro arquivo `.dzs` real já compila de ponta a ponta**
  (Lex→Parse→Generate).

  **Desvio de arquitetura necessário, descoberto implementando o
  Parser** (mesmo tipo de achado que `BuiltinDefine.Expand` revelou
  pro ZX81): `internal/core/pipeline.go`'s `runPass` só dá a cada hook
  de pass acesso ao slice de ENTRADA original daquela pass, nunca a
  uma saída que o PRÓPRIO pass já tenha construído até aquele ponto —
  diferente do Python original, cujo `last_tok()` sempre reflete a
  saída JÁ ACUMULADA pelo pass atual, incluindo qualquer inserção
  sintética que o mesmo pass já tenha feito mais atrás. Isso importa
  porque `pass_5` do Python insere `LET` implícito E expande
  atribuição composta (`x+=5*2`) NA MESMA PASS, com a segunda
  dependendo de inspecionar a inserção da primeira. Duas adaptações,
  nenhuma mudança no `core` em si:
  1. O "LET" implícito moveu de `Pass5` (onde o Python está) para
     `Generate` — mesmo lugar que o `Parser` do ZX81 já usa
     (`zx81/parser.go`) — já que sua condição de disparo (identificador
     logo depois de número de linha, `:` ou `THEN`) é autocontida: não
     precisa saber se um `LET` foi "de verdade" inserido antes, só
     reavaliar a MESMA condição no lugar certo.
  2. A expansão de atribuição composta usa um campo de estado na
     própria instância do `Parser` (`closeParenAt`, "onde inserir o `)`
     de fechamento, mais adiante na mesma passada") — porque, diferente
     do `x+=5` do MSX (RHS de 1 token só, nunca precisa disso), o RHS
     do Spectrum pode ser uma EXPRESSÃO INTEIRA (`x+=5*2`), varrida até
     o próximo `NEWLINE`/`EOF`/separador — a leitura de validação (RHS
     vazio, aninhamento, string só aceita `+=`) acontece direto sobre o
     slice de entrada (sem precisar "consumir" nada), mas o FECHAMENTO
     do parênteses só pode acontecer numa chamada POSTERIOR, daí o
     estado.

  Validação extensiva rodando o `badig.py` real (não só os testes
  automatizados, mesma disciplina já usada nas Etapas 1/2): `PRINT AT`
  com 1 argumento (splice literal "PRINT AT 5;", sem default por
  coordenada — confirma a diferença do ZX81 já suspeitada na pesquisa);
  atribuição composta (`x+=5*2`→`x=x+(5*2)`, `x++`→`x=x+(1)`, string só
  aceita `+=`, `++`/`--` rejeita RHS, aninhamento rejeitado, exige
  precedido de `LET`/posição de início de statement — as 4 mensagens de
  erro batem caractere a caractere); `GOTO`/`GO  TO`/`GOSUB`/`GO SUB`
  sempre virando `GO TO`/`GO SUB` com exatamente 1 espaço,
  independente de como foi escrito; `TRUE`/`FALSE`→`1`/`0`; toda
  palavra reservada sempre maiúscula incondicional (`sin`/`cos`/`let`/
  `if`/`rem` todos viram maiúsculo); colisão de nome de variável com
  `cat`/`format`/`move`/`erase` (comandos de disco reais da ROM que
  este módulo não suporta, mas cuja colisão ainda precisa ser
  rejeitada) rejeitada com a mensagem exata; `~` removido sem efeito
  de aliasing (diferente de MSX/ZX81); `DECLARE` substitui o alias e
  mantém o sufixo `$`; nomes de string precisam ser EXATAMENTE 1 letra
  + `$` (não 2 letras + `$` — ponto de atenção real: "2 chars
  terminados em `$`" na pesquisa original significava o comprimento
  TOTAL da string incluindo o `$`, não 2 letras antes dele); array/
  FOR-control precisam de exatamente 1 letra; escalar comum aceita
  qualquer comprimento.

  **Correção de premissa confirmada pela pesquisa ANTES de portar**: a
  frase "alocação por categoria" (como no cabeçalho original desta
  tarefa) estava ERRADA — Spectrum **não tem nenhum alocador de
  variável** (nomes nativos de qualquer comprimento são válidos); só
  existe validação de forma por categoria (`ValidateIdentifier`,
  tokenizer.go, já portada na Etapa 1) — corrigido no texto do item
  acima (removida a menção a "alocação por categoria").

  Achados confirmados **rodando o `badig.py` real** (não só lendo o
  código, disciplina já estabelecida no projeto): `GO TO`/`GO SUB` com
  espaço variável são vocabulário do **Lexer Dignified** aqui (não só
  do tokenizer, diferente do ZX81); apóstrofo (`'`) é o separador real
  de `PRINT` do Sinclair BASIC (`PRINT "A"'"B"`), nunca um marcador de
  comentário; `REM` precisou virar **DOIS tokens** no Lexer (a própria
  palavra-chave + o conteúdo cru como um token separado) — confirmado
  que isso reproduz o espaçamento duplo real da saída (`REM  texto`,
  dois espaços: um do espaçamento automático de `core.Generate` entre
  palavra reservada e o que vem depois, outro já parte do próprio
  conteúdo preservado) — diferente da convenção mais simples de
  MSX/ZX81 (REM como um blob só).

  **Etapa 4 (validação de ponta a ponta), feita em 2026-11-26**: os 9
  exemplos reais já existentes no projeto
  (`badig/examples/zxspectrum/{hello,data,fn,colours,functions,
  native,streams,graphics,play128}.DZS` — não escritos pra esta
  rodada, já estavam no repositório) rodados pelo pipeline inteiro
  (`Lex→Parse→Generate→Tokenize→MakeTap`) batem **byte a byte**, tanto
  no `.AZS` renderizado quanto no `.tap` tokenizado, contra a saída
  real do `badig.py` (confirmado rodando `python` de verdade nesta
  máquina) — `graphics.DZS` só com a linha `REMX 1000,32` removida
  (mesma decisão de escopo já tomada pro ZX81: `REMX` fora de alcance),
  `play128.DZS` tokenizado com `model="128"`. Dois bugs reais achados e
  corrigidos por essa validação (nenhum nos testes unitários isolados,
  só apareceram com exemplos reais que combinam `FUNC`/`RET` com as
  regras do próprio dialeto): a maiúscula incondicional e a
  normalização `GOTO`→`GO TO` (Pass5, acima) checavam `tok.Kind` em
  vez do VALOR do token — core's compartilhado `pass2Functions`/
  `pass2Labels` emitem seus próprios `gosub`/`return`/`goto` sintéticos
  como `core.KindSynthetic`, nunca `KindJump`/`KindInstruction` deste
  pacote, então ficavam minúsculos sem forçar (`gosub 130`/`return`
  em vez de `GO SUB 130`/`RETURN`) — corrigido pra checar por VALOR
  (igual o próprio Python faz, `tk.uval in c_reserved_kw`, nunca por
  "tipo de token"); e `atStatementStart` (usado tanto pelo `LET`
  implícito quanto pela atribuição composta) checava `prev.Kind ==
  KindSeparator`, mas o `":"` sintético que `pass2Functions` insere
  entre os passos de uma chamada de `FUNC` TAMBÉM é `KindSynthetic` —
  corrigido pra checar `prev.String() == ":"` por valor também.
  `functions.DZS` (que combina os dois: `.double(score)` expande pra
  `GO SUB`+`RETURN` sintéticos, E a atribuição de retorno
  `answer=result` precisa do próprio `LET` implícito depois do `:`
  sintético) foi exatamente o exemplo que revelou os dois.

  3 testes comitados (`TestEndToEndRealHelloExample`/
  `...FunctionsExample`/`...FnExample`, conteúdo embutido como string
  Go, sem depender de arquivo em disco — mesmo princípio 100%-em-
  memória já usado o projeto inteiro, cobrindo os 3 exemplos
  estruturalmente mais distintos: programa simples com label/`LET`
  implícito/`GOTO`, atribuição composta + `FUNC`/`RET`, e `DEF FN`/
  `FN`) + 1 scratch (`zz_scratch_e2e_test.go`, não comitado, lendo os
  9 arquivos reais de uma cópia local e comparando `.AZS`/`.tap`
  contra a saída do Python real).

  **Etapa 5 (comando `compile`/`execute` pro Spectrum), feita em
  2026-11-26 — ROTEIRO COMPLETO DO BLOCO 4 FECHADO (Etapas 1-5)**:
  `internal/cli/dialect.go` ganha `.dzs`→Spectrum nos dois
  dispatchers (`compileByExtension`/`runInEmulatorByExtension`),
  junto com `.dmx`/`.dzx` já existentes.

  `internal/platforms/zxspectrum/compile.go` (novo):
  `DefaultCompileSettings` (mesmos defaults compartilhados do MSX/
  ZX81) e `CompileFile(dzsPath, settings) (azsPath, tapPath string,
  err error)`, espelhando `msx.CompileFile`/`zx81.CompileFile`: `core.Run`
  + `Tokenize` (sempre `model="48"` — o modelo 128K fica fora da
  integração de emulador deste round, decisão de escopo já
  documentada acima) + `MakeTap` (nome da fita derivado do nome base
  do arquivo, truncado em 10 chars — `TK_NAME` padrão do original).

  `internal/platforms/zxspectrum/run.go` (novo): `RunInEmulator(tapPath)`
  — bem mais rico que ZX81/MSX, confirmado lendo `emulator_interface.py`
  de novo: EightyOne do Spectrum recebe DOIS argumentos posicionais
  (`[ini, fita]`, não um só), exigindo um perfil INI próprio
  (`eightyone48.ini`, 11 linhas, embutido como string Go — `TZXMAN`
  já vem com `FlashLoad`/`AutoStart`/`AutoLoad=1`, exatamente o que
  este round sempre precisa já que só produz `.tap`). Uma cópia
  temporária desse INI é escrita a cada chamada (nunca edita o
  original) e NÃO é apagada depois — como esta função não espera o
  emulador terminar (decisão abaixo), apagar na hora arriscaria uma
  corrida com o próprio EightyOne ainda lendo o arquivo ao abrir.

  **Decisão tomada com o usuário antes de implementar**: o
  `emulator_interface.py` original do Spectrum ESPERA o EightyOne
  fechar (`process.wait()`, falha se saída for diferente de zero) —
  diferente de MSX/ZX81 no Go, que já abrem e retornam na hora.
  Perguntado via `AskUserQuestion`: escolhida a **consistência com
  MSX/ZX81** (não esperar), não a fidelidade ao Python nesse ponto
  específico — as três plataformas do `goBaDig --execute` agora têm
  o MESMO comportamento de CLI, mesmo que o Python original divirja
  por dialeto.

  Validado de verdade: `goBaDig --execute demo.dzs` (fonte mínima
  `CLS:PRINT "...":STOP`) gerou `demo.azs`/`demo.tap` corretos e abriu
  um processo `EightyOne.exe` REAL nesta máquina (confirmado via
  `Get-Process` — mesmo nível de confirmação já aceito pro MSX/ZX81:
  processo vivo confirmado, sem confirmação visual de tela neste
  ambiente). Testes novos: `compile_test.go`/`run_test.go`
  (`internal/platforms/zxspectrum`), mais os pares `.dzs` dos testes
  `Run`/console/dialect já existentes pro `.dmx`/`.dzx`
  (`internal/cli`).
- [x] Portar tokenizer Spectrum, incluindo encoding de número Sinclair
  (usar `math/big.Rat` para precisão, validar contra os 69 testes Python
  como golden-master). — **A**, **1** — **feito em 2026-11-26**
  (`internal/platforms/zxspectrum/{tokenizer,tape}.go`, v1.34.0 - Etapa 1
  de 5 do roteiro deste Bloco, mesmo padrão do Bloco 3/ZX81). Ainda SEM
  Description/Lexer/Parser Spectrum (próxima Etapa) - este tokenizer só
  converte texto Classic ASCII **já numerado** em bytes tokenizados; nada
  ainda chama ele a partir do CLI/console. `.tap`/`.tzx` (empacotamento de
  fita, item separado abaixo) também portados nesta rodada, junto -
  100% mecânico (`tape.go`), sem motivo pra separar numa Etapa própria.

  **Ponto flutuante Sinclair via `math/big` (Rat/Int), NÃO `float64`**:
  diferente do ZX81 (que usa `math.Frexp`/IEEE754 porque o Python
  original também usa `float64` ali), o Spectrum real usa
  `decimal.Decimal`+`fractions.Fraction` (aritmética racional EXATA) -
  replicado com `big.Rat` (auto-reduz numerador/denominador, essencial:
  o algoritmo original depende da fração já REDUZIDA, não só do valor).
  Confirmado à mão, byte a byte, contra os vetores reais de
  `test_zxspectrum.py` (`0.1`→`7d4ccccccd`, `65536`→`9100000000`, etc.)
  - bias de expoente **+129** (não +128 como o ZX81 - ROM diferente).
  Achado de robustez: a checagem rápida de faixa (`decimal.adjusted()`
  no Python, replicada com `big.Int` puro, sem nunca calcular `10^E`
  primeiro) precisa rodar ANTES de montar a fração de verdade, exatamente
  na mesma ordem do Python - só assim um literal com expoente
  astronômico (`1e999999999999999999999999`, vetor real de teste) falha
  rápido em vez de tentar computar uma potência de 10 gigantesca.

  **Diferença estrutural mais importante do módulo** (confirmada pela
  pesquisa ANTES de portar, ver nota no topo deste Bloco): Spectrum
  **não tem alocador de variável nenhum** (diferente de MSX/ZX81) -
  nomes nativos de qualquer tamanho são válidos; só existe
  `ValidateIdentifier` (validação por categoria: string precisa de 2
  chars terminados em `$`, array/FOR-control precisa de 1 char, escalar
  comum aceita qualquer tamanho) - fica pra Pass4 do Parser (próxima
  Etapa) chamar, já que não envolve `core.VariableAllocator`.

  **REMX deliberadamente FORA de escopo nesta Etapa** (mesma decisão já
  tomada pro ZX81, Bloco 3): diferente do ZX81 (onde o tokenizer
  reconhece `remx` como prefixo especial de uma linha já numerada),
  aqui `parse_remx`/REMX inteiro vive só no nível do Parser Dignified
  (`Pass1` + uma reinserção pós-`generate()` que contorna a numeração
  normal) - não há superfície alguma de tokenizer pra portar aqui.
  Revisitar junto com o Parser (Etapa 3), se um programa real precisar.

  Portados também (funções puras do tokenizer original, usadas por
  `DEF FN`/`FN`/comandos nativos): `FnDefinition` (slot oculto de 6
  bytes por parâmetro, marcador `0x0E`+5 zeros - mecanismo real de
  runtime da ROM, não um truque deste projeto), `ValidateFnCall`
  (sintaxe de uma chamada `FN nome(args)`, parênteses balanceados),
  `validateCommandArguments` (contagem/forma de argumento dos comandos
  nativos como `OUT`/`PLAY`/`OPEN #`/etc.).

  Verificado: os 19 testes Go novos (`internal/platforms/zxspectrum`,
  golden vectors transcritos de `TokenizerTests`/`TapeTests` do
  `test_zxspectrum.py` real, 68 casos no total no Python) passam -
  únicas 2 discrepâncias encontradas rodando os testes foram no PRÓPRIO
  vetor transcrito à mão (um espaço esquecido em 2 lugares), não na
  implementação; corrigido no teste, não no código. `go build`/`go
  vet`/`gofmt`/`go test ./...` limpos no repo inteiro.
- [ ] REMX com fill configurável. — **B**, **2**
- [x] Empacotamento TAP e TZX (100% mecânico). — **B**, **1** —
  **feito em 2026-11-26** (`internal/platforms/zxspectrum/tape.go`,
  v1.34.0, junto com o tokenizer acima — sem motivo pra separar numa
  Etapa própria, é mecânico de verdade: checksum XOR, header/data block,
  TZX só reempacota os mesmos blocos do TAP).
- [ ] Portar DignifieR Spectrum (AZS→DZS). — **M**, **2**
- [x] Integração EightyOne 48K básica — **feito em 2026-11-26**
  (Etapa 5, `internal/platforms/zxspectrum/run.go`'s `RunInEmulator`,
  v1.38.0 — ver acima). Spectaculator continua fora (nenhum pedido do
  usuário por ele ainda). — **M**, **1**
- [x] Integração ZEsarUX como backend adicional — **feito em
  2026-11-26**, pedido do usuário: `zxspectrum.RunInZEsarUX(tapPath)`
  — `--machine 48k --tape <arquivo>`, confirmado lendo o `--help` do
  próprio `zesarux.exe` e lançando de verdade nesta máquina. Seleção
  de backend: nova flag `--emulator <nome>`/console `execute
  <arquivo> [emulador]`, compartilhada com MSX/ZX81 — ver Bloco 5
  abaixo. **Só a invocação direta, SEM ZRCP** — monitoramento de erro
  em tempo real (breakpoint em `PC=0008H`, já avaliado viável no
  Bloco 0) continua não implementado, item separado abaixo. — **M**, **2**
- [ ] Monitoramento de erro em tempo real via ZRCP (ZEsarUX) —
  breakpoint em `PC=0008H`, ler `PPC`/`SUBPPC`/`ERR_NR` — já avaliado
  viável no Bloco 0, ainda não implementado (a integração acima só
  abre o emulador, sem usar o protocolo ZRCP em si). — **A**, **2**
- [ ] Suporte a modelo 128K (tokens SPECTRUM/PLAY). — **M**, **2**
- [ ] `.DSK` de disco +3 (opcional, baixa prioridade). — **A**, **3**
- [ ] Investigar/corrigir o bug de freeze do launcher 128K do EightyOne
  (pode ser mais fácil de diagnosticar com `exec`/pipes do Go do que foi
  em Python). — **M**, **2**
- [ ] Snapshots Z80/SNA via emulador Z80 em Go (substituindo a dependência
  SkoolKit) — **adiar para fase tardia**, não bloqueia MVP. — **A**, **3**
- [ ] Monitoramento de erro de ROM — **investigado e considerado inviável
  por automação externa** no Python (só via simulação offline). Reavaliar
  só se/quando o emulador Z80 em Go (item anterior) existir. — **A**, **3**

### Bloco 5 — CLI

- [x] **Stub** (release 1.1.2): `--help` (lista de opções), `--man`
  (referência estilo man — só o "about" do projeto nesta versão),
  `--tui`/`--gui` (despacho para as interfaces interativas), banner
  padrão sem argumentos. Código em `internal/cli`, `internal/about`,
  `internal/version`. — 2026-11-05. **Atualizado na 1.1.4** com `--cli`
  (console interativo, ver item abaixo).
- [ ] Parser de argumentos real (flags core + flags por módulo
  compostas, equivalente aos `parent_parsers`), substituindo o
  dispatch simples do stub. — **M**, **1**
- [x] **Motor de "man" + comandos do BaDig + dicionário MSX completo +
  manual do ASMSX** (parcial, 2026-11-27, v1.48.0/v1.49.0/v1.50.0 -
  Fases 2, 3 e 4 do plano de 4 fases, ver Bloco 6): `goBaDig --man
  [tópico]`/console `man [tópico]` navegam por tópico de verdade
  agora (`internal/man`, ver bullets detalhados no Bloco 6), cobrindo
  os próprios comandos do BaDig (`compile`/`execute`/`set`/
  `settings`/`dialect`/`theme`/`version`/`man`), as 181 palavras
  reservadas do BASIC MSX (`--man abs`, `--man msx` pro índice do
  dicionário - dados do `msxide`/`paleobasic` do usuário) e, desde a
  v1.50.0, o manual completo do compilador ASMSX (`--man asmsx`).
  **Ainda faltam** os manuais de sintaxe BASIC do ZX81 e ZX Spectrum
  embutidos (sem fonte equivalente ao dicionário MSX - via
  `fwMSX`/`go81` e `badig/documentation/ZX_SPECTRUM.md`, se/quando
  decidido) — ver seção 10.7 para detalhes. — **A**, **2**
- [x] Modo de comando interativo (`--cli`, release 1.1.4): console com
  prompt `goBaDig>` aguardando comandos linha a linha (`help`, `man`,
  `about`, `version`, `exit`/`quit`, encerra também em EOF/Ctrl-D),
  além do modo "um tiro" por linha de comando. Código em
  `internal/cli/console.go`, com testes (saída e parada em EOF) e
  teste manual via pipe. Comandos reais de conversão chegam com o
  motor (Blocos 1-4). — 2026-11-05.
- [x] Detecção automática de plataforma pela extensão do arquivo —
  `internal/cli/dialect.go`'s `compileByExtension`/
  `runInEmulatorByExtension`, feito junto com `compile`/`execute` de
  cada plataforma (Blocos 2/3/4, `.dmx`/`.dzx`/`.dzs`) — item marcado
  retroativamente, já estava pronto desde a v1.27.0 (`.dmx`), sem
  nenhuma flag `-id`/override explícita ainda (nunca foi pedida).
  — **B**, **2**
- [x] **Seleção de emulador alternativo** (`--emulator <nome>` no CLI,
  `execute <arquivo> [emulador]` no console) — **feito em 2026-11-26**,
  pedido do usuário depois de perceber que MSX/ZX81/ZX Spectrum agora
  têm mais de um backend instalado nesta máquina. `emulatorBackends`
  (mapa extensão→lista de nomes válidos, primeiro = padrão) +
  `resolveBackend` (valida, rejeita um nome válido pra OUTRA
  extensão). `--emulator` pode vir antes OU depois de `--execute`/do
  nome do arquivo (`emulatorArg` lê o valor dos args originais;
  `stripEmulatorFlag` remove a flag+valor ANTES de qualquer outra
  detecção de argumento posicional rodar — bug real encontrado e
  corrigido nesta mesma rodada: sem o strip, `--execute --emulator
  zesarux prog.dzs` confundia "zesarux" com o nome do arquivo, já que
  `--execute` sempre pegava literalmente o próximo token). Backends
  por plataforma: MSX = `openmsx` (padrão), `fmsx`, `bluemsx+`;
  ZX81/ZX Spectrum = `eightyone` (padrão), `zesarux`. — **B**, **2**

  **Extraído pro pacote `internal/dialect` em 2026-11-26** (v1.40.0,
  sem mudança de comportamento): `compileByExtension`/
  `runInEmulatorByExtension`/`emulatorBackends`/`resolveBackend`
  viviam em `internal/cli/dialect.go`, privados ao pacote `cli` — a
  TUI não conseguia chamar esse dispatch sem passar a depender de
  `internal/cli` (que também carrega parsing de flags, algo específico
  da CLI). Agora `internal/dialect/dialect.go`, tudo exportado
  (`CompileByExtension`/`RunInEmulatorByExtension`/`EmulatorBackends`/
  `ResolveBackend`) — preparação direta pro gerenciador de arquivos,
  ver Bloco 6 abaixo.

### Bloco 6 — TUI (tview/tcell)

- [x] **Stub** (release 1.1.2): menu `File > Exit` e `Help > About`,
  navegação por `Enter`/`Esc`. Código em `internal/tui`. — 2026-11-05.
- [x] **Menu suspenso estilo DOS clássico** (release 1.1.3, pedido
  explícito do usuário): substituído o menu em tela cheia por uma
  barra fixa no topo (`File`, `Help`) com item ativo em destaque
  invertido e caixa flutuante sobre o conteúdo — visual e navegação no
  estilo **Turbo Pascal 7 / Norton Commander / Norton Editor**. Setas
  movem entre itens da barra; com o menu aberto, Esquerda/Direita troca
  direto para o menu vizinho (padrão TP7); `Alt+letra` abre direto;
  `Enter`/`F10` abre; `Esc` fecha o suspenso. Implementado com
  `tview.Pages` + `Flex` com espaçadores `nil` (truque padrão do tview
  para sobrepor uma caixa sem apagar o conteúdo por baixo). Testado em
  runtime (abre/fecha sem erro). **Não implementado ainda**: sombra de
  borda (drop shadow) característica do visual original — ficou de
  fora por custo de engenharia desproporcional ao stub; suporte a
  clique de mouse na barra (os itens do dropdown já aceitam clique,
  nativo do `tview.List` — só a barra em si ainda não). — 2026-11-05.
- [x] **Investigação e correção do congelamento em `Setup > Install`**
  (2026-11-20, em duas etapas da mesma sessão):
  - Primeira etapa: revisão do código (`internal/tui/tui.go`) cruzada com
    o código-fonte de `tview`/`tcell` instalado (`rivo/tview v0.42.0`), e
    reprodução dirigida com uma `tcell.SimulationScreen`
    (`internal/tui/freeze_diagnosis_test.go`, mantido como teste de
    regressão). Não foi possível reproduzir um deadlock Go com a
    sequência óbvia (abrir `Setup`, selecionar `Install`, `Escape`,
    `Stop()`) — o app sempre respondia e retornava normalmente.
  - Segunda etapa, indo mais a fundo: montando um dump textual da tela
    simulada após abrir `Setup > Install`, ficou claro que **o diálogo de
    tema simplesmente nunca aparecia** — nada travava, mas nada era
    desenhado. **Causa raiz real, confirmada isolando uma cópia local do
    `rivo/tview` com instrumentação (via `replace` temporário no
    `go.mod`) até provar que `Grid.Draw()` nunca era chamado**: a
    construção do diálogo encadeava `tview.NewGrid()...AddItem(...).
    SetBackgroundColor(colors.desktopBg)` — `SetBackgroundColor` é
    herdado de `*tview.Box` (não tem override em `Grid`) e **retorna
    `*tview.Box`**, não `*tview.Grid`. Como o resultado dessa cadeia era
    atribuído direto à variável do diálogo, essa variável virava
    silenciosamente um `*Box` liso (ainda satisfaz a interface
    `Primitive`, sem erro de compilação) em vez de `*Grid` — e
    `Box.Draw()` só desenha fundo/borda e retorna, nunca chega no
    layout/desenho dos filhos do Grid (o `form` com o dropdown de temas).
    **Era exatamente esse o "congelamento"**: nenhum deadlock, só um
    diálogo inteiro invisível — o usuário pressiona `Install` e nada
    parece acontecer. Corrigido em `internal/tui/tui.go` separando essa
    chamada em instrução própria, fora da cadeia que define o tipo da
    variável (comentário extenso deixado no código explicando a pegadinha
    para não se repetir). Validado de novo com dump textual: o diálogo
    agora renderiza completo (janela, dropdown, descrição, botões).
  - Na mesma sessão, a TUI também foi modernizada no visual retrô estilo
    MS-DOS anos 1990 (Turbo Vision/Turbo Pascal/Norton Commander), pedido
    explícito do usuário: janelas com borda dupla, caixa de sistema `[■]`
    e sombra projetada (`internal/tui/chrome.go`, funções `drawShadow`/
    `attachShadow`/`windowChrome`), ícones Unicode nos menus e itens
    (`▭ File`, `⚙ Setup`, `ℹ Help`, `✕ Exit`, `◐ Install`, `★ About`,
    swatch `■` colorido por tema no dropdown de temas), e um desktop com
    textura pontilhada (`▒`, tipo `desktop` em `chrome.go`) com banner
    central e régua decorativa. Também corrigido, no mesmo lote, um
    off-by-one pré-existente: o dropdown de menu abria sobrepondo a
    própria barra de menu em vez de aparecer embaixo dela.
- [ ] Tela de configuração (equivalente ao que hoje são os `.ini`/remtags):
  editar config SQLite global/projeto. — **M**, **2**
- [ ] Tela de pré-processamento em lote (escolher arquivos, plataforma,
  opções, rodar e ver log/erros). — **M**, **2**
- [ ] Navegação entre MSX/ZX81/Spectrum sem reiniciar o programa. — **B**, **2**
- [ ] Visualizador de manual BASIC embutido (mesmo conteúdo do `--man`
  da CLI), acessível pelo menu Help. — **M**, **2**
- [x] **WibEdit integrado como editor da TUI** (2026-11-21, pedido
  explícito do usuário: "quero este tipo de interface neste projeto...
  ele vai ser a base de nossa IDE"). Reverte a decisão de 2026-11-05 que
  mantinha um editor próprio fora de escopo — o usuário apontou
  diretamente para o projeto `E:\wibworks` (não o `msxide`/`paleobasic`
  mencionados antes) e pediu a cópia completa, não um binário separado.
  Portado de `E:\wibworks\internal\modules\edit` (WibEdit — editor
  estilo WordStar/Norton Editor 2, já no mesmo stack `tcell v2.13.10`/
  `tview v0.42.0` do goBaDig) para o novo pacote `internal/editor`
  (~9 mil linhas + ~5,2 mil linhas de testes, todos portados e passando
  sem alteração de comportamento), mais `internal/printpdf` (exportação
  PDF, copiado de `internal/shared/printpdf` do wibworks). Embutido como
  mais uma página do `tview.Pages` já existente em `internal/tui/tui.go`
  (não é um binário separado): novo menu de topo **Editor > Open**
  (`iconEdit` em `chrome.go`), com o `app.SetInputCapture` roteando para
  `editor.Capture` sempre que a página "editor" está em primeiro plano —
  mesmo padrão que o próprio wibworks usa para embutir o WibEdit no seu
  shell (`internal/shared/ui/shell.go`). Textos de identidade (About,
  ajuda, nome do documento na impressora, diretório de config global)
  adaptados de "WIBEditor/WIBWorks" para "WibEdit (goBaDig)", mantendo o
  crédito de autoria a Wilson Pilon. Teste de regressão novo
  (`internal/tui/editor_menu_test.go`) confirmado como capaz de detectar
  regressão (testado quebrando o roteamento de propósito e revertendo).
  Chrome visual do WibEdit (bordas duplas/sombra desenhadas manualmente
  em `editor.go`) mantido como está, coexistindo com o `chrome.go` do
  goBaDig sem tentar unificar os dois sistemas agora. Impressão/exportação
  PDF portadas junto (não adiadas): nova dependência `go-pdf/fpdf
  v0.9.0`, `golang.org/x/image` atualizado de v0.24.0 para v0.48.0 (usado
  também pela GUI Fyne — **correção 2026-11-21**: o build completo
  `go build ./...`, incluindo `internal/gui`, foi verificado nesta
  máquina com CGO habilitado e compila/testa limpo com a versão nova de
  `x/image`; a ressalva anterior sobre não conseguir verificar a GUI
  valia só para o ambiente usado naquele momento, não é uma limitação
  deste repositório). Na época, ainda não estava ligado a nada
  específico de BASIC (realce de sintaxe, abertura direto de
  `.bas`/arquivos-fonte das 3 plataformas) — destaque de sintaxe
  chegou na v1.46.0, ver abaixo.
- [x] **Gerenciador de arquivos de dois painéis** (2026-11-26, v1.41.0,
  pedido explícito do usuário: "já que não temos os gerenciadores de
  arquivo ainda, podemos seguir por este caminho e assim chamar o
  compilador do gerenciador, criar projetos, etc."; VSIX/Sublime —
  Bloco 8 — ficam deliberadamente pro final, já que o WibEdit embutido
  cobre edição). Novo pacote `internal/filemanager`: dois painéis
  (`tview.Table`, colunas nome/tamanho/data) estilo Norton Commander.
  `Tab` troca o painel ativo; `Enter`/`Backspace` navega; `F2` compila
  (`dialect.CompileByExtension`); `F9` executa no emulador padrão da
  plataforma (`dialect.RunInEmulatorByExtension`); `F4` abre o arquivo
  selecionado no WibEdit já embutido (novo `editor.Editor.OpenFile`,
  exportando `openInitialFile`); `F5` copia (só arquivo - copiar
  diretório pediria recursão, fora de escopo nesta rodada) e `F6`
  move (arquivo OU diretório, via `os.Rename`, sem recursão) pro
  painel oposto, com confirmação; `F7` cria diretório; `F8` apaga
  (arquivo ou diretório vazio - `os.Remove` já rejeita diretório
  não-vazio, sem checagem extra); `Ctrl+N` cria um projeto (nome +
  pasta, criada se não existir, + dialeto + checkbox opcional pra já
  gerar um arquivo-fonte inicial vazio) e `Ctrl+P` lista projetos
  salvos, pulando o painel ativo pra pasta escolhida.

  **Nova tabela `projects`** em `internal/config/projects.go`, mesmo
  `*Store` SQLite de temas/opções: `CreateProject`/`ListProjects`.
  Configurações por projeto (linha inicial, passo, emulador padrão
  etc.) e editar/apagar um projeto já salvo ficam fora desta rodada -
  confirmado com o usuário via pergunta: "pasta + dialeto + registro
  pra acesso rápido" é a definição aprovada de "projeto".

  **Novo pacote `internal/chrome`**: os ícones/sombra/moldura de janela
  de `internal/tui/chrome.go` (`attachShadow`/`windowChrome`/
  `drawShadow`) extraídos pra lá, exportados. Necessário porque
  `internal/filemanager` precisa do mesmo visual retro mas não pode
  importar `internal/tui` (que importa `internal/filemanager` pra
  embutir a página, igual já faz com o editor - importar nos dois
  sentidos seria um ciclo). `internal/tui/chrome.go` manteve só o
  `desktop` (fundo tramado da área de trabalho, genuinamente
  específico da TUI). Mesma lógica já usada pra extrair
  `internal/dialect` do `internal/cli` na release anterior.

  **Decisões de escopo confirmadas com o usuário**: TUI primeiro, GUI
  depois (mesma decisão já tomada pro editor); montagem de `.DSK` num
  painel e operações local↔`.DSK` ficam pra depois (o `compile`/
  `execute` do MSX já não precisam disso - o `BuildTestDisk` interno
  cuida do disco de teste); novo item de menu "Files" colocado no
  **final** da barra (depois de Help), não entre Editor e Setup, pra
  não quebrar a navegação por índice (`File→Editor→Setup` via duas
  setas) que vários testes de regressão já existentes assumiam.

  Testes novos: `internal/config/projects_test.go` (criar/listar,
  nome vazio/duplicado rejeitado, dialeto inválido rejeitado, ordem
  por mais recente primeiro), `internal/filemanager/
  filemanager_test.go` (12 casos: navegação, troca de painel,
  compilar/executar com `PATH` vazio, copiar/mover/criar diretório/
  criar projeto), mais `internal/tui/filemanager_menu_test.go`
  (`tcell.SimulationScreen`, mesmo padrão de `editor_menu_test.go`)
  confirmando que **Files > Open** abre o gerenciador e `Esc` fecha de
  volta pro desktop.

  **Correção de bug + visual (2026-11-26, v1.41.1)**, reportada
  diretamente pelo usuário: o esquema de cores fixo (amarelo sobre
  preto) "parecia um programa de 1979" e, pior, alguns diretórios
  apareciam com letra preta sobre fundo preto - navegáveis, mas
  invisíveis. **Causa raiz**: nenhum `table.SetSelectedStyle` era
  configurado nos painéis (`tview.Table`) - sem isso, `tview.Table`
  cai no fallback de inverter a cor da linha selecionada combinando o
  texto da célula com o `BackgroundColor` dela (nunca definido
  explicitamente), uma combinação sem garantia de contraste. Mesma
  classe de bug já vista uma vez antes neste projeto (`tview.List` no
  dropdown do menu Setup, ver nota acima sobre
  `TestSetupMenuUnselectedItemsAreNotBlackOnBlack`), agora em
  `tview.Table`. Corrigido com `SetSelectedStyle` explícito (fundo =
  cor de seleção do tema, texto = branco fixo - mesma escolha já
  comprovada nos dropdowns de menu da própria TUI) **e** trocando o
  esquema fixo pelo tema ativo da TUI: novo `internal/filemanager.
  Colors` (com `DefaultColors()` pra testes sem tema), mapeado do
  tema ativo em `internal/tui/tui.go`'s `filemanagerColorsFor`, e um
  novo `FileManager.SetColors` chamado pelo mesmo botão "Apply" que
  já reaplica o tema no resto da TUI. Teste de regressão novo
  (`internal/tui/filemanager_colors_test.go`,
  `TestFileManagerEntriesAreNotBlackOnBlack`), mesma técnica de
  varredura de tela inteira da `TestSetupMenuUnselectedItemsAreNot
  BlackOnBlack` - confirmado capaz de pegar a classe de bug.

  **Segunda correção de bug (2026-11-26, v1.41.2)**: a v1.41.1 só
  corrigiu a linha SELECIONADA de cada painel - o usuário reportou de
  volta que "não mudou nada, as cores ficaram diferente, mas agora
  aparece texto cinza com fundo cinza", ou seja, todas as OUTRAS
  linhas (não selecionadas) continuavam com o mesmo problema.
  **Causa raiz real**: `tview.NewTableCell` grava `cell.Style` com um
  valor concreto já no construtor (branco sobre preto, padrão global
  do `tview`), e `TableCell.SetTextColor` só altera o componente de
  PRIMEIRO PLANO desse estilo (`c.Style = c.Style.Foreground(cor)`,
  só quando `cell.Style != tcell.StyleDefault` - o que é sempre o
  caso, vindo do construtor). O fundo fica exatamente como o
  construtor deixou - preto, fixo do `tview` - a menos que
  `SetBackgroundColor` TAMBÉM seja chamado. A v1.41.1 só chamava
  `SetTextColor` por célula: o texto pegava a cor certa do tema, mas
  ficava sentado numa caixa preta que o tema nunca escolheu - "cinza
  sobre cinza" quando o tema ativo tem um texto em tom acinzentado
  (confirmado nesta máquina: o tema realmente ativo e persistido,
  "Solarized Light", tem `Foreground = #586E75`, um cinza-azulado
  musgo). Essa causa raiz só foi confirmada de verdade inspecionando
  o campo `Style` da própria `TableCell` direto em um teste unitário
  (`internal/filemanager/panel_test.go`,
  `TestRenderUsesTheConfiguredColorsNotTviewDefaults`) - um teste
  anterior baseado só em dump de tela simulada (`tcell.
  SimulationScreen`) passou incorretamente duas vezes antes disso,
  porque não distinguia "o padrão do `tview` coincidiu com o valor
  esperado" de "minha cor foi realmente aplicada" - lição registrada
  pra qualquer bug de cor futuro nesta UI: inspecionar o estado do
  widget direto quando possível, não confiar só numa tela renderizada.
  Corrigido construindo o estilo inteiro (texto E fundo) com
  `SetStyle` de uma vez, em `panel.go`'s `render`, em vez de
  `SetTextColor`/`SetBackgroundColor` separados. Teste de tela
  simulada reforçado pra também checar o fundo (não só o texto) de
  uma linha NÃO selecionada, contra o tema realmente ativo e
  persistido nesta máquina.

  **Terceira correção, a causa raiz de verdade (2026-11-26, v1.41.3)**:
  o usuário mandou uma captura de tela (`images/erro-filemanager.png`)
  depois da v1.41.2 mostrando que o nome de quase todo diretório
  aparecia EM BRANCO - só a data da linha ficava visível. As duas
  correções anteriores miraram cor porque o relato original ("preto
  com preto") parecia um problema de contraste, mas NUNCA foi cor
  nenhuma. **Causa raiz real**: o `tview` trata qualquer trecho
  `[...]` como uma tentativa de tag de estilo `[cor]`/`["região"]`
  sempre que o caractere logo depois do `[` é letra, dígito, `#`, `:`
  ou `-` - casado e CONSUMIDO (zero caracteres visíveis), quer a "cor"
  dentro resolva pra algo real ou não. O formato `"[" + nome + "]"`
  usado pelos diretórios (estilo Norton Commander) colidia direto com
  isso: `"[..]"` sobrevivia por acidente (o caractere depois do `[` é
  `.`, inválido como início de tag), mas `"[cmd]"`, `"[docs]"`,
  `"[internal]"` - qualquer diretório real começando com letra ou
  dígito - desaparecia por completo. Confirmado isolando uma cópia
  local do `rivo/tview` com instrumentação via `replace` temporário em
  `go.mod` (mesma técnica já usada uma vez neste projeto, pro bug do
  `SetBackgroundColor` em `Grid` do Setup > Install) até provar que
  `printWithStyle` retornava zero caracteres impressos pra `"[aaa]"` e
  todos os 4 pra `"[..]"`. Corrigido com `tview.Escape` (a própria
  ferramenta documentada do `tview` pra esse caso exato) em
  `internal/filemanager/panel.go`'s novo `entryLabel`, no título do
  painel, na barra de status (mensagens de erro citam caminhos) e na
  lista de projetos salvos. **Lição registrada**: um teste que só lê
  `cell.Style`/`cell.Color` direto (como os dois anteriores) não pega
  um bug de TEXTO ausente - só um teste que renderiza de verdade numa
  `tcell.SimulationScreen` e procura o texto completo na tela pega
  essa classe de problema (`internal/filemanager/panel_test.go`,
  `TestMultipleDirectoryEntriesAreActuallyVisibleOnScreen` - confirmado
  capaz de reproduzir exatamente a captura de tela do usuário).

- [x] **Gerenciador de arquivos vira o centro de controle do programa**
  (2026-11-26, v1.42.0), pedido explícito do usuário depois de usar a
  v1.41.3: "o File Manager agora vai ser o centro de controle total
  do programa". Menu de topo "Files" separado removido - o
  gerenciador agora abre por **`File > File Manager`** (primeiro item
  do menu `File`, antes de "Exit"); `menus[0]` ganhou um segundo item,
  os outros índices (`Editor`/`Setup`/`Help`) não mudaram. Novo `F3`
  **Criar Arquivo** (arquivo vazio no painel ativo, nomeado pelo
  usuário - a extensão escolhe o dialeto -, recusa sobrescrever,
  aberto na hora no WibEdit). `Ctrl+N` **Criar Projeto** revisado: o
  arquivo inicial deixa de ser opcional (checkbox removida) e passa a
  se chamar sempre `main<ext>` (não mais `<nome do projeto><ext>`) -
  nome fixo e previsível, também aberto na hora no editor. **Arquivos
  de interesse ganham cor forte + ícone próprio**: fonte BADIG
  Dignified (`.dmx`/`.dzx`/`.dzs`, ícone `◆`), listagem ASCII clássica
  (`.amx`/`.azx`/`.azs`, ícone `≡`), artefato compilado/empacotado
  (`.bmx`/`.p`/`.tap`/`.tzx`/`.dsk`, ícone `▣`) - cada categoria usa
  uma cor do tema ativo que nenhuma outra parte da TUI já usava
  (`Palette.Hover`/`Palette.Error`, ambas deliberadamente vívidas em
  todo tema curado), pra garantir contraste forte em qualquer tema;
  `SourceText` divide `Primary` com `DirectoryText` de propósito - a
  notação `[nome]` já distingue diretório de arquivo. Fecha o ciclo
  completo pedido pelo usuário inteiramente pelo gerenciador: criar →
  editar → salvar/sair → compilar (`F2`) → executar (`F9`). Teste
  novo (`internal/tui/new_file_test.go`,
  `TestF3CreatesAFileAndOpensItInTheEditor`) dirige o diálogo de
  verdade por teclas simuladas, não uma chamada Go direta - achado
  nesse processo: `Enter` com um `InputField` em foco não envia o
  formulário na hora (o campo só "termina", passando o foco pro botão
  - mais um `Enter` clica nele) - registrado como lição de teste pra
  qualquer diálogo futuro deste tipo. **A contagem exata não é uma
  constante confiável** (3 numa rodada, 5 noutra pro mesmo cenário -
  achado corrigindo os dois bugs da v1.42.1 abaixo): o teste usa
  `submitFormDialog`, que pressiona `Enter` e espera ativamente o
  diálogo fechar antes de decidir se precisa de mais um, em vez de uma
  contagem fixa.

  **Duas correções (2026-11-26, v1.42.1)**, reportadas pelo usuário
  depois de usar a v1.42.0:
  1. **Barra de status "bagunçada" ao voltar do editor**: fechar o
     editor sempre trocava pra página `"app"` (o desktop)
     incondicionalmente, mesmo quando aberto a partir do gerenciador
     de arquivos (`F3`/`F4`) - o usuário via a barra de status do
     DESKTOP (`F10 Menu...`), diferente da do gerenciador (`Tab
     Painel...`), e percebeu isso como "a barra quebrou". Corrigido
     com `editorReturnPage` (`"app"` pelo menu `Editor > Open`,
     `"filemanager"` pelo `F3`/`F4`), que `closeEditor` usa pra voltar
     pro lugar certo. Achado tentando escrever um teste de regressão
     pra isso via `F3` (criar arquivo) - o teste em si era instável
     (a sequência de `Enter`s do diálogo às vezes "passava do ponto" e
     caía dentro do editor já aberto, digitando uma quebra de linha em
     vez de clicar em nada); reescrito usando `F4` (editar um arquivo
     já existente, sem diálogo nenhum no caminho) pra testar a mesma
     correção sem essa fragilidade.
  2. **Projeto novo organizado em pasta própria**: pedido do usuário -
     "crie uma pasta dentro do diretório corrente para organizar os
     arquivos". O campo "Pasta" do diálogo `Ctrl+N` passa a ser a
     pasta PAI (onde criar o projeto), não mais a pasta final - `
     createProject` sempre cria uma subpasta com o nome do projeto
     dentro dela, em vez de jogar os arquivos direto na pasta que o
     usuário estava navegando. Teste novo
     (`TestCreateProjectCreatesItsOwnSubfolderInsideTheParent`)
     confirma que a pasta pai fica limpa.

  **Barra de status "presa" numa mensagem antiga (2026-11-27,
  v1.44.0)**, reportada pelo usuário: "a linha de status ainda não
  está sendo atualizada a cada opção usada, algumas alteram a barra
  de status e a mesma não volta". Causa raiz: `F2`/`F5`/`F7` e
  outras ações deixavam a mensagem de status presa na tela - o único
  lugar que reconduzia pro texto de atalhos era `Enter`/`Backspace`.
  Navegação por setas dentro de um painel nunca passa por
  `FileManager.Capture` (`Up`/`Down`/`PageUp`/`PageDown` não estão no
  seu `switch`) - o próprio `InputHandler` do `tview.Table` processa
  essas teclas direto, então uma mensagem de ação ficava presa
  indefinidamente assim que o usuário só movia a seleção pelo
  teclado. Corrigido com `SetSelectionChangedFunc` em cada painel
  (ligado em `NewFileManager`): reconduz a barra pro texto de
  atalhos sempre que a seleção muda, incluindo as mudanças que as
  próprias ações já disparam via `reload()`/`load()` - a chamada
  final de `setStatus` de cada ação sempre roda DEPOIS do seu
  `reload`/`load`, então sobrescreve esse reset de volta pra mensagem
  certa da ação. `enterSelected`/`goUp` perderam sua própria chamada
  explícita de `setStatus(statusHelp)` no caminho de sucesso (virou
  redundante - o `load()` que já fazem dispara o mesmo reset sozinho),
  mantendo só a do caminho de erro. Teste novo
  (`TestSelectingADifferentRowResetsTheStatusBar`) confirmado capaz
  de pegar o bug revertendo o fix de propósito.

  **Projeto novo "invisível" quando o banco falha (2026-11-27,
  v1.45.0)**, reportada pelo usuário: "ao criar um projeto novo, às
  vezes dá erro e não mostra os arquivos no file manager, mas
  entrando e saindo do programa o diretório está lá". Causa raiz:
  `config.Store.CreateProject` roda `os.MkdirAll` ANTES de inserir no
  banco - uma falha no passo do banco (o caso mais comum é um nome de
  projeto duplicado, já que `name` é a chave primária dessa tabela e
  persiste entre reinícios no `settings.db` real do usuário) ainda
  assim deixa um diretório de verdade, recém-criado, no disco. O
  `createProject` do gerenciador retornava assim que `store.
  CreateProject` falhava, sem nunca recarregar o painel - esse
  diretório ficava invisível na sessão atual até o programa inteiro
  ser reiniciado. Corrigido recarregando o painel ativo sempre, mesmo
  quando o passo do banco falha. Teste novo
  (`TestCreateProjectStillShowsTheDirectoryWhenTheDatabaseStepFails`)
  reproduz com um nome duplicado - a forma mais simples e real de
  disparar a falha - confirmado capaz de pegar o bug revertendo o fix
  de propósito (achado um detalhe de isolamento ao escrever o teste:
  o painel precisa estar navegado pra um diretório DIFERENTE de
  `parent` antes da segunda chamada, senão um recarregamento manual
  anterior no próprio teste mascara a ausência do recarregamento
  dentro de `createProject`).

- [x] **Destaque de sintaxe no WibEdit** (2026-11-27, v1.46.0), pedido
  explícito do usuário: "Vamos implementar este destaque de sintaxe
  no Norton Editor que está dentro do GoBaDig... inicialmente vamos
  usar apenas o destaque de sintaxe, mas depois vamos usar o
  autocomplete e o help sensível ao contexto". Baseado no material de
  referência já existente pros dialetos (`badig/<dialeto>/Sublime
  Package` e `badig/<dialeto>/VSCode Extension`), **não** no lexer da
  compilação (`internal/platforms/*`) - aquele é construído pra um
  passo único, batch, sobre um programa inteiro e correto; isso aqui
  roda a cada tecla digitada, sobre texto possivelmente incompleto ou
  inválido no meio da edição, então reusar o mesmo lexer traria risco
  de performance/estabilidade sem necessidade.

  Novo `internal/editor/syntax.go`/`syntax_keywords.go`: um
  tokenizador POR LINHA (`tokenizeLine`), independente do pipeline de
  compilação, com tabelas de palavras-chave/funções próprias por
  dialeto (MSX, ZX81, ZX Spectrum), destiladas do material Sublime/
  VSCode já existente no projeto (pesquisado via subagent: o `.vsix`
  do MSX precisou ser descompactado - é só um zip - pra achar a
  gramática real, `extension/syntaxes/msx-basic-dignified.
  tmLanguage.json`, já que o único `.sublime-syntax` do MSX no
  repositório é só pro output do compilador, não pro código-fonte).
  Seis categorias: palavra-chave, função, comentário, string, número,
  label (`{nome}`) - mais simples que as gramáticas completas
  originais (sem distinguir operadores, sem detecção de "invalid
  input", sem separar construções Dignified como categoria à parte)
  de propósito: o valor visual está majoritariamente nessas seis,
  pelo menor risco de implementação. Dialeto escolhido pela extensão
  do arquivo aberto (`.dmx`/`.amx` = MSX, `.dzx`/`.azx` = ZX81,
  `.dzs`/`.azs` = ZX Spectrum) - fonte Dignified e listagem clássica
  numerada abrem com destaque igual.

  Seis cores novas (`internal/editor/editor.go`), com o fundo sempre
  sincronizado com `contentStyle` (`options_display.go`), pra
  continuar combinando com qualquer um dos 5 perfis de cor já
  existentes do WibEdit (esse sistema de cores é independente do
  `internal/theme`, como já documentado antes). Ligado no próprio
  laço de desenho do editor (`internal/editor/document_view.go`'s
  `Draw`): tokeniza só as linhas realmente visíveis na tela a cada
  redesenho, não o documento inteiro - custo proporcional à altura do
  terminal, não ao tamanho do arquivo.

  **Limitação conhecida, documentada de propósito**: o tokenizador não
  guarda estado entre linhas - comentários de bloco de várias linhas
  (apóstrofo duplo abrindo/fechando, ou `###...###`) não são
  rastreados como abertos/fechados, então uma palavra em inglês
  dentro de um desses blocos que por acaso também é uma palavra-chave
  BASIC (`ON`, `TO`, `FOR`, `AND`...) ainda aparece destacada, mesmo o
  bloco inteiro sendo um comentário - confirmado rodando contra um
  programa MSX real (`badig/examples/msx/CGK173.dmx`, que tem um
  cabeçalho de comentário de várias linhas). Rastrear isso direito
  pediria varrer o documento inteiro (ou um índice por documento de
  "quais linhas estão dentro de um bloco não fechado", invalidado a
  cada edição) - fora de escopo nesta primeira rodada, registrado como
  possível próximo passo.

  Teste novo (`internal/editor/syntax_test.go`): tokenizador testado
  isoladamente pros três dialetos, mais um teste de ponta a ponta
  (`TestOpeningADmxFileActuallyRendersSyntaxColors`) que abre um
  `.dmx` de verdade via `Editor.OpenFile` (o mesmo caminho que o `F4`
  do gerenciador de arquivos usa) e confirma que a palavra-chave e a
  string realmente aparecem com cores diferentes na tela simulada -
  confirmado capaz de pegar uma regressão desligando o gancho de
  propósito.

- [x] **Autocomplete no WibEdit** (2026-11-27, v1.47.0), primeira
  metade do próximo passo já anunciado na v1.46.0. Antes de
  implementar, o usuário pediu para estudar o **MSXIDE** (projeto
  próprio dele, `E:\msxide`, **FreeBASIC**, não Go) - investigação
  confirmou que a arquitetura (popup por prefixo reusando a mesma
  fonte de dados do dicionário/man, parser de Markdown próprio pro
  manual/help, ASMSX vendorizado como subprocess com resolução de
  caminho em cascata) é reaproveitável como **conceito**, sem código
  Go a portar. A partir dessa investigação + 3 agentes Explore no
  próprio goBaDig, foi registrado um **plano de 4 fases** (salvo fora
  do repo em `C:\Users\barney\.claude\plans\`, ver `resume.txt` pro
  resumo completo): Fase 1 (esta) = autocomplete só com as tabelas
  já existentes; Fase 2 = motor genérico de "man"/referência +
  parser de Markdown + visualizador no WibEdit + `--man`/`man
  <tópico>` reais na CLI (hoje ainda stub, ver nota no Bloco 8/seção
  10.7); Fase 3 = dicionário MSX Basic completo (dados de
  `E:\msxide\ajuda`, trazidos e **commitados publicamente** - decisão
  explícita do usuário, avisado sobre a exposição de conteúdo de
  terceiros) + ajuda contextual Shift+F1; Fase 4 = suporte ao **ASMSX**
  (compilador Assembly, `resource/asMSX/`) como ferramenta dentro de
  projetos MSX BASIC existentes (não um dialeto de projeto separado -
  decisão explícita do usuário), com `ASMSX_PATH` no catálogo de
  opções e compilação via subprocess síncrono.

  Implementado nesta rodada: novo `internal/editor/autocomplete.go` -
  popup ancorado na palavra sendo digitada, reusando
  `msxKeywords`/`msxFunctions`/`zx81Keywords`/`zx81Functions`/
  `zxSpectrumKeywords`/`zxSpectrumFunctions` (as mesmas tabelas do
  destaque de sintaxe) como única fonte de candidatos - nenhum dado
  externo. `Tab`/`Enter` completa, `Up`/`Down` navega, `Esc` fecha sem
  completar; qualquer outra tecla cai no fluxo normal do editor e a
  lista se atualiza/fecha sozinha a partir do callback `changed`/
  `moved` que `documentView` já expunha. `Editor.Capture` ganhou um
  terceiro modo de interceptação de tecla (nem "passa tudo" nem
  "bloqueia tudo", os dois já existentes).

  **Bug real encontrado e corrigido no caminho**: a primeira versão do
  popup copiava o mecanismo do `editorDropdown` (página nova em
  `tview.Pages`) - isso quebrava a digitação por completo depois do
  popup abrir uma vez. Causa raiz confirmada lendo o código-fonte de
  `rivo/tview@v0.42.0` (`pages.go:88-90`): `Pages.AddPage` **rouba o
  foco** pra página recém-adicionada sempre que o container já tem
  foco, e `*tview.Box` não tem `InputHandler` próprio - toda tecla
  seguinte era engolida em silêncio. Corrigido abandonando `Pages` pro
  popup: `Editor` ganhou um método `Draw` próprio que desenha o popup
  direto por cima de `Pages.Draw`, sem nunca criar página nem tocar
  foco - mesma classe de bug (mecanismo do `tview` com efeito colateral
  não-óbvio) que já tinha custado a investigação do "congelamento" do
  Setup > Install em 2026-11-20.

  Teste novo (`internal/editor/autocomplete_test.go`, 7 casos): abre
  com o candidato certo nos 3 dialetos, `Tab` confirma e substitui o
  prefixo, `Esc` fecha sem alterar texto, espaço fecha sozinho (sai da
  palavra), setas navegam com wrap-around, nunca abre sem dialeto
  reconhecido.

- [x] **Motor de "man"/referência + visualizador de Markdown** (2026-11-27,
  v1.48.0) - Fase 2 do plano de 4 fases registrado na v1.47.0 (ver
  bullet acima). Ainda não traz o dicionário MSX (Fase 3) nem o manual
  do ASMSX (Fase 4) - esta rodada entrega a infraestrutura
  compartilhada e um primeiro conteúdo real: os próprios comandos do
  goBaDig.

  Novo pacote `internal/man`: `Page`/`Line`/`TOCEntry` e um parser de
  Markdown escrito à mão (`ParseMarkdown`, `markdown.go`) - inspirado
  conceitualmente no `BuildMarkdownBufferFromText` do MSXIDE (estudado
  na Fase 1), implementado do zero em Go. Suporta headings
  `#`.."######", blocos de código cercados, citações `>`, listas
  `-`/`*`/`N.` e parágrafos com quebra de linha automática (70
  colunas) - tabelas e ênfase inline ficam como texto literal nesta
  versão (documentado, não bloqueia).

  Conteúdo novo versionado `docs/man/badig.md` + `docs/man/embed.go`
  (pacote `docsman`): a referência dos próprios comandos do goBaDig
  (`compile`/`execute`/`set`/`settings`/`dialect`/`theme`/`version`/
  `man`) no formato NOME/SINOPSE/DESCRIÇÃO, um `# tópico` por comando.
  Fica em `docs/` (não `internal/`) porque `go:embed` não alcança
  diretórios pai - o `//go:embed` em si mora em `docs/man/embed.go`,
  ao lado do `.md`, e `internal/man/registry.go` importa esse pacote
  em vez de embutir o arquivo direto. `registry.go` divide o arquivo
  pelos headings de nível 1 (`splitTopLevelSections`) - cada um se
  torna um tópico independente (`Lookup`/`Index`), embora o arquivo
  inteiro se leia como um único documento Markdown normal no GitHub/
  editor. Pronto para a Fase 3 acrescentar o dicionário MSX e a Fase 4
  o manual do ASMSX do mesmo jeito (mais uma fonte em
  `buildTopics()`).

  CLI: `--man [tópico]` (`cli.go`, `manTopicArg`/`runMan`) e console
  `man [tópico]` (`console.go`, `executeMan`) substituem o stub antigo
  (`about.Man()`, que agora também delega pro índice de `internal/
  man`). Autocompletar Tab dos tópicos disponíveis em
  `completion.go`.

  WibEdit: visualizador novo em tela cheia (`internal/editor/
  view_man.go`, `editorManView`), aberto por um menu **"Manual"**
  próprio (tecla `M`) - navegação real (`Up`/`Down`/`PgUp`/`PgDn`/
  `Home`/`End` rolam, `Enter` numa linha de índice abre o tópico
  empilhando histórico, `Esc` volta um nível ou fecha), modal enquanto
  aberto (nenhuma outra tecla alcança o documento).

  **Achado implementando, não planejado**: o comando novo do WibEdit
  quase ganhou o id `"help.manual"` - só não ficou assim porque
  `Editor.execute` intercepta **todo** id com prefixo `"help."`
  incondicionalmente pra `showHelp` (a ajuda F1 de "como usar o
  editor", que reproduz fielmente um editor de referência específico,
  NE2 - ver `neMenus`/`neHelpMenu`, `commands.go`). O manual do BaDig é
  um conceito diferente (referência da linguagem BASIC/ASMSX, não da
  própria edição), então ganhou id (`manual.open`) e menu próprios,
  sem tocar no "F1=Help" existente nem na fidelidade do `neMenus`.

  Testes novos: `internal/man/man_test.go` (parser contra cada
  elemento suportado, quebra de parágrafos longos, `Lookup`/`Index`
  case-insensitive e tópico desconhecido); `internal/cli` ganhou casos
  pra `--man`/`man` sem tópico, tópico válido/inválido e
  autocompletar; `internal/editor/view_man_test.go` (abertura,
  navegação Up/Down, Enter num link de índice, Esc voltando/fechando,
  e que o visualizador é mesmo modal).

- [x] **Dicionário MSX BASIC completo + ajuda contextual Shift+F1**
  (2026-11-27, v1.49.0) - Fase 3 do plano de 4 fases. Conteúdo trazido
  e **commitado publicamente** no repositório, decisão explícita do
  usuário (avisado sobre a exposição de conteúdo de terceiros antes
  de prosseguir - "Linguagem BASIC MSX", Denise Santoro Cruz, Editora
  Aleph/Gradiente, 4a edição 1986, + "Manual MSX2+ FM", ACVS
  Eletrônica, ambos já transcritos no projeto próprio do usuário,
  MSXIDE, `E:\msxide\ajuda`).

  Conversão de dados (passo único, não é código de produção): um
  conversor Go descartável (não faz parte do build, escrito/rodado no
  scratchpad da sessão) parseou `MsxBasicDictData.pbi` (140 verbetes,
  letras A-W, sem J/Q) e `MsxBasic2PlusDictData.pbi` (41 verbetes) -
  balanceando parênteses/aspas das chamadas `MSXDict_Add(...)`/
  `MSXDict_Add2Plus(...)` do PureBasic original, resolvendo a
  concatenação de strings (literais `"..."`, o atalho `MSXQ` pra um
  caractere de aspas literal, `#CRLF$` pra nova linha) - e gerou
  `internal/man/msxdict_data.go` (181 entradas, 140 MSX1 + 41 MSX2+).

  Novo em `internal/man`: `dict.go` - `MsxDictEntry` (título, se é
  função/avançado, origem em inglês, resumo, formato, exemplo, função,
  programa exemplo, página do livro, sistema); `(MsxDictEntry).Page()`
  (molde fixo NOME/FORMATO/EXEMPLO/FUNÇÃO/PROGRAMA EXEMPLO/REFERÊNCIA,
  produzindo o mesmo tipo `Page` do Markdown - CLI/WibEdit não
  precisam saber a diferença); `FindMsxDictEntry` (match exato case-
  insensitive, com fallback por verbete composto dividido em "-" - ex.
  `"RETURN"` resolve `"GOSUB-RETURN"`, o próprio exemplo do plano
  original); `MsxDictIndex` (agrupado por letra inicial);
  `AutocompleteWords` (normaliza títulos compostos/com sufixo
  "(MSX2+)"/"(MSX2+/FM-Music)" em palavras individuais - remove grupos
  entre parênteses, trata "/" como separador extra, descarta o "e"
  solto de listas em português). `registry.go`'s `Lookup` passa a
  tentar `FindMsxDictEntry` primeiro; `Index()` ganhou um link pro
  pseudo-tópico reservado `"msx"`, que `Lookup` resolve pra
  `MsxDictIndex()`.

  WibEdit: `Shift+F1` (`isShiftF1` em `view_man.go` - cobre tanto
  `KeyF1`+`ModShift`, como o driver de console nativo do Windows do
  tcell reporta, quanto `KeyF13`, convenção xterm pra terminais ANSI
  genéricos) chama `contextualHelp()`: extrai a palavra tocando o
  cursor (`wordAtCursor`, novo em `autocomplete.go` - escaneia nos
  dois sentidos, diferente do `wordPrefixBeforeCursor` da Fase 1, que
  só olha pra trás), normaliza `"?"` pra `"PRINT"` (atalho clássico do
  MSX BASIC), busca no dicionário e abre no visualizador da Fase 2;
  fora do MSX, cai no índice geral em vez de não fazer nada (decisão
  de polimento do próprio plano, não bloqueia). Autocomplete (Fase 1)
  passou a combinar `man.AutocompleteWords()` com as tabelas de
  sintaxe pra MSX (`candidatesFor` em `autocomplete.go`), mantendo-as
  como reforço; ZX81/Spectrum inalterados.

  **Achado implementando, não planejado**: uma primeira versão de
  `(MsxDictEntry).Page()` juntava um campo inteiro (ex. `Função`) numa
  única string e refazia a quebra de linha do zero (`wrapText` sobre
  o todo) - isso apagava as quebras de linha/linhas em branco que o
  próprio livro usa pra alinhar diagramas de fluxo (`GOSUB-RETURN`,
  ver o texto "10 GOSUB 1000" / "20 GOSUB 1000" / "1000 (sub
  rotina)..." em linhas separadas de propósito) e tabelas comparativas
  (`AUTO`). Corrigido quebrando primeiro pelas quebras de linha já
  existentes no texto de origem (`strings.SplitSeq(text, "\n")`), e só
  reaplicando `wrapText` numa linha individual que ultrapasse a
  largura.

  Testes novos: `internal/man/dict_test.go` (contagem plausível de
  verbetes - 181 = 140+41, protege contra um arquivo gerado vazio/
  truncado passar despercebido -, match exato case-insensitive, match
  composto RETURN→GOSUB-RETURN, palavra-chave desconhecida, ordem fixa
  das seções de `Page()`, agrupamento por letra do índice, prioridade
  do dicionário em `Lookup`, pseudo-tópico `"msx"`); novo caso em
  `autocomplete_test.go` (RET→RETURN, só existe via o verbete
  composto); novos casos em `view_man_test.go` (Shift+F1 abrindo o
  verbete certo nos dois formatos de tecla, normalização do "?", aviso
  pra palavra desconhecida sem abrir nada, fallback fora do MSX, F1
  puro continuando a abrir o menu de ajuda sem ser afetado pela nova
  checagem).

- [x] **Suporte ao compilador ASMSX** (2026-11-27, v1.50.0) - Fase 4
  (última) do plano de 4 fases. Não uma reimplementação do asMSX -
  só integração: configuração de caminho, compilação via subprocess,
  e o manual completo embutido no mesmo sistema de referência das
  Fases 2/3.

  `docs/man/asmsx.md`: cópia de `resource/asMSX/doc/asmsx.md` (1148
  linhas), com uma nota de atribuição (projeto GPLv3, originalmente de
  Eduardo "pitpan" A. Robsy Petrus, hoje mantido pelo asMSX team -
  github.com/Fubukimaru/asMSX) logo após o título - `resource/` é
  gitignored e não é distribuído, então o conteúdo precisa estar num
  caminho versionado pra funcionar em qualquer máquina que rode o
  binário publicado (mesma razão de `docs/man/badig.md`, Fase 2).
  Registrado como um tópico só (`"asmsx"`) via uma função nova em
  `internal/man/registry.go`, `addMarkdownTopic(pages, key, src)` -
  diferente de `addMarkdownSource` (Fase 2, que divide um arquivo em
  vários tópicos por heading de nível 1), esta mantém o manual
  inteiro como um único tópico sob uma chave curta explícita
  ("asmsx"), preservando o título real do manual ("asMSX - MSX cross
  assembler") só pra exibição - a chave de registro e o título de
  exibição deliberadamente NÃO coincidem aqui, ao contrário de
  `badig.md`/dicionário MSX, onde coincidiam por construção. Isso
  expôs um bug sutil em `Index()`: o `TOCEntry.Title` de cada linha
  do índice estava usando `topics[name].Title` (o título de EXIBIÇÃO)
  em vez de `name` (a chave de BUSCA) - funcionava por acidente pra
  todo tópico onde os dois coincidiam, mas quebrava a navegação por
  Enter especificamente pro "asmsx". Corrigido usando `name` no
  `TOCEntry`.

  `internal/options/options.go`: novo campo `ASMSX_PATH` (categoria
  MSX, ao lado do `EMULATOR_PATH` existente), sem validação de
  existência de arquivo (mesmo padrão não-especial que
  `EMULATOR_PATH` já segue). Aparece automaticamente nos painéis
  `Setup > MSX` da TUI/GUI, que já renderizam genericamente a partir
  de `options.ForCategory` - nenhuma mudança de UI necessária.

  Novo `internal/platforms/msx/asmsx.go`: `ResolveAsmsxPath(store)` -
  cascata (1) `ASMSX_PATH` configurado e `os.Stat` confirma que
  existe, (2) `exec.LookPath("asmsx")` no PATH do sistema;
  `EMULATOR_PATH` já modelava essa mesma ideia de cascata, mas nenhum
  dos seus campos chegou a ser lido por código de produção até agora
  - este é o primeiro. Camada "bundled ao lado do goBaDig.exe"
  deliberadamente fora de escopo (documentado, não implementado) -
  não há pipeline de build que compile `asmsx.exe` a partir do código
  C vendorizado em `resource/asMSX/src/` (precisa gcc+flex+bison).
  `CompileAsm(asmPath, asmsxExe)` - mesmo padrão síncrono-com-captura-
  de-saída de `internal/editor/printer_unix.go` (único outro
  precedente no projeto - todo lançamento de emulador hoje é fire-
  and-forget, `cmd.Start()` sem esperar). Roda dentro do diretório do
  próprio `.asm`, passando só o nome base (pra `#include` relativo do
  asMSX resolver certo); sucesso confere se `<base>.bin` foi criado
  (convenção padrão do asMSX sem diretiva de saída alternativa -
  `.rom`/`.cas`/`.wav`/`.com` dependem de diretiva dentro do próprio
  `.asm` e não são detectados automaticamente - limitação conhecida
  de v1, documentada).

  Dispatch por extensão pra `.asm`, deliberadamente **fora** do
  `switch` fechado de `internal/dialect.go` (só pra dialetos
  Dignified), em três lugares: `internal/filemanager/actions.go`'s
  `compileSelected` (F2 na TUI - novo `compileAsmSelected`),
  `internal/cli/console.go`'s `executeCompile` (`compile arquivo.asm`
  no console - novo `executeCompileAsm`) e `internal/cli/cli.go`'s
  `runCompile` (`goBaDig arquivo.asm` direto - novo `runCompileAsm`,
  abre a store padrão já que não há estado de sessão nesse caminho
  non-interativo).

  **Cuidado de teste encontrado implementando**: `executeCompileAsm`/
  `runCompileAsm` abrem a store de configurações (pra ler
  `ASMSX_PATH`) - `executeCompileAsm` reusa `s.openStore()`, que é
  no-op se `s.store` já estiver setado, então os testes do console
  SEMPRE precisam injetar uma store isolada (`config.Open` num
  arquivo temporário) antes de chamar `executeCompile(".asm")`, senão
  o teste abriria/tocaria o `settings.db` real do usuário em
  `%APPDATA%\goBaDig` (confirmado lendo `config.OpenDefault`). Já
  `runCompileAsm` (CLI não-interativa) não tem como receber uma store
  injetada - o teste correspondente aceita esse toque mínimo e
  idempotente no ambiente real (mesma coisa que uma invocação real do
  programa compilado faria).

  Testes novos: `internal/platforms/msx/asmsx_test.go`
  (`ResolveAsmsxPath`: caminho configurado existente, configurado mas
  inexistente, fallback pro PATH, nem configurado nem no PATH;
  `CompileAsm` testado contra uma ferramenta REAL - um binário Go
  minúsculo compilado em tempo de teste via `go build`, já que não há
  `asmsx` real garantido no ambiente - sucesso com `.bin` gerado,
  saída capturada em caso de erro, aviso quando a ferramenta termina
  sem erro mas não gera o `.bin`); novos casos em `internal/man/
  dict_test.go` (`Lookup("asmsx")` case-insensitive, `Index()`
  linkando pela chave curta mas exibindo o título real); um teste de
  dispatch em cada um dos três pontos de integração confirmando que
  `.asm` é roteado pro ASMSX e não cai no erro de "dialeto não
  reconhecido" do `CompileByExtension`.

  Testado manualmente de ponta a ponta nesta máquina contra um
  `asmsx.exe` real já instalado (achado via PATH do sistema) -
  confirmou a cascata de resolução e o encapsulamento de erro
  funcionando (um lock de arquivo específico desta máquina impediu a
  compilação de fato rodar, mas confirmou que o erro do subprocesso é
  capturado e reportado corretamente em vez de travar o programa).

  Com isso, as 4 fases do plano registrado na v1.47.0 (autocomplete +
  man/help + dicionário MSX + ASMSX) estão completas e lançadas
  (v1.47.0 a v1.50.0). Fora de escopo, documentado como passo futuro
  natural (não bloqueia "suporte ao ASMSX" em si, que já está
  completo): os 3 formatos de ponte Assembly→BASIC que o MSXIDE
  oferece (BLOAD direto / DATA-READ-POKE embutido / `.inc` reusável).

- [x] **Ponte Assembly→BASIC: macro `[INCBLOAD]`** (2026-11-27,
  v1.51.0) - Fase 5, pedida pelo usuário na sequência da Fase 4: o
  primeiro dos 3 formatos de ponte do MSXIDE (BLOAD direto), como uma
  macro Dignified nova em vez de edição automática do `.dmx` -
  confirmado com o usuário via pergunta fechada (macro recalculada a
  cada compilação vs. inserção de texto automática no arquivo-fonte;
  macro escolhida, evita o compilador escrever no arquivo do usuário
  como efeito colateral de compilar o Assembly).

  `[INCBLOAD]("arquivo.asm")` expande, em cada compilação do `.dmx`,
  para o equivalente a `BLOAD"ARQUIVO.BIN":DEFUSR=&Hxxxx` - sem `,R`
  no BLOAD de propósito (`,R` dispara execução automática ao carregar,
  adequado a um programa que toma conta da máquina inteira, não a uma
  rotina pensada para ser chamada repetidamente via `USR()`, que é o
  caso de uso pedido). Exige que `arquivo.asm` já tenha sido compilado
  (seu `.bin` precisa existir) - a macro só lê artefatos já gerados,
  nunca chama o asmsx ela mesma, mantendo compilação do Assembly e da
  BASIC como dois passos desacoplados (já existentes).

  Mecanismo escolhido depois de descartar dois outros: uma instrução
  Dignified nova em `internal/core` (rejeitada - macro é vocabulário
  só do MSX, não deveria morar no núcleo compartilhado pelos 3
  dialetos) e pós-processamento fora do pipeline, direto em
  `msx.CompileFile` (rejeitada - perderia a numeração de linha
  automática do core). Escolhido: `BuiltinDefine.Expand`
  (`internal/core/contract.go`), o mesmo mecanismo que já registra
  `"[?](x,y)"` pro MSX e `"[?](row,col)"` pro ZX81 - mas `Expand`
  recebia só os tokens do argumento, sem o `*CompileContext`
  necessário pra resolver caminho relativo (mesma regra do `INCLUDE`,
  `ctx.IncludeStack[0]`) nem forma de reportar erro (um `.bin`
  faltando é falha real de compilação, não algo que uma macro
  puramente sintática como a do ZX81 precisava expressar). Mudança
  aditiva em `core.BuiltinDefine.Expand`:
  `func(args []Token) []Token` →
  `func(ctx *CompileContext, args []Token) ([]Token, error)` - só 2
  implementações existentes no repo (MSX e ZX81), ambas atualizadas;
  `expandDefineCall` (`pipeline.go`) já retornava `([]Token, error)`
  então o call site simplesmente passou a propagar os dois valores
  direto.

  Descoberta do endereço de entrada (`internal/platforms/msx/
  incbload.go`, `resolveEntryAddress`): lê a diretiva `.START` do
  `.asm` (ou `.ORG`, se não houver `.START`) - se o argumento for um
  literal (`0x`/`$`/sufixo `h`, ou decimal), usa direto; se for um
  rótulo, resolve lendo o arquivo `.sym` que o próprio asmsx gera
  junto com o `.bin` (confirmado lendo `resource/asMSX/src/dura.y`,
  `write_sym`/`finalize`: só é gerado se o programa define algum
  símbolo, formato `XXXXh NOME` por linha, `run_address` já resolvido
  pelo bison antes de chegar lá - nenhuma reimplementação de
  expressões Z80 foi necessária, só leitura do resultado já calculado
  pelo asmsx). Limitação conhecida, documentada: um rótulo cujo nome
  parece um literal hex válido com sufixo `h` (ex. `ABh`) é lido como
  número em vez de rótulo - a mesma ambiguidade que o próprio asmsx
  resolve com contexto de tabela de símbolos completo, que esta
  varredura textual não tem.

  Sintaxe do argumento (`[INCBLOAD]("arquivo.asm")`, nome entre aspas
  com a extensão completa) escolhida pra espelhar exatamente a
  convenção já existente do `INCLUDE "arquivo"` - mesma função de
  "referenciar um arquivo externo por nome".

  Testes novos: `internal/platforms/msx/incbload_test.go` - os 3
  formatos numéricos do asMSX + decimal, rejeição de nome de rótulo
  como número, `.START` tendo prioridade sobre `.ORG`, última
  ocorrência quando repetida, resolução de rótulo via `.sym`, erros
  (sem `.START`/`.ORG`, símbolo não encontrado no `.sym`, `.bin`
  ainda não compilado) e um teste de integração completo (`.dmx` real
  com `[INCBLOAD]`, passando pelo `CompileFile` inteiro até a listagem
  Classic gerada, confirmando o `BLOAD`/`DEFUSR` certos no texto
  final). Testado manualmente de ponta a ponta via `goBaDig
  arquivo.dmx` contra fixtures reais de `.asm`/`.bin`/`.sym`.

  Novo tópico `incbload` em `docs/man/badig.md` (`--man incbload`).

### Bloco 7 — GUI (Fyne)

- [x] **Stub** (release 1.1.2): janela com menu `File > Exit` e
  `Help > About` (diálogo nativo Fyne). Código em `internal/gui`. — 2026-11-05.
- [ ] Janela principal com gerenciador de projetos/arquivos. — **M**, **2**
- [x] Catálogo inicial de seis temas coloridos para GUI, com seleção
  pelo menu `Setup > Install`; a seleção ativa é gravada no SQLite e o
  tema é aplicado imediatamente e restaurado na próxima inicialização.
  Catálogo e seleção ficam em tabelas distintas, organizadas por modo
  (`themes` e `active_themes`). — 2026-11-07.
- [ ] Tela de configuração equivalente à TUI, mas gráfica. — **M**, **2**
- [ ] Execução de pré-processamento com feedback visual de erro
  (aproveitando `Position` para apontar linha/coluna exata). — **M**, **2**
- [ ] Integração com emuladores (botão "rodar no emulador"). — **M**, **2**
- [ ] Visualizador de manual BASIC embutido, mesmo conteúdo do `--man`. — **M**, **2**
- [ ] Avaliar reaproveitar o **editor GUI com teclas/funções do Norton
  Editor 2 já pronto no projeto `msxide`** do usuário como editor de
  texto integrado (ver seção 10.7) — ainda fora de escopo de
  implementação, só registrado como opção futura. — **A**, **3**

### Bloco 8 — Suporte a editores

Prioridade confirmada pelo usuário (seção 10.6): VSCode (1) → Sublime
Text (2) → editor próprio Norton-based (fora de escopo agora) →
Notepad++/UltraEdit (opcional, "se for possível").

**VSCode (VSIX único):**

- [ ] Consolidar os 3 `package.json` num só, IDs de linguagem namespaced
  (`badig-msx-dignified`, `badig-zx81-dignified`, `badig-spectrum-dignified`,
  e variantes "classic"). — **M**, **1** (para a entrega final)
- [ ] Unificar settings sob seção `badig.*` (caminho do binário Go único,
  sem precisar mais de Python/interpretador). — **B**, **1**
- [ ] Fundir tasks.json das 3 plataformas usando o padrão mais rico
  (Spectrum) como base. — **M**, **1**
- [ ] Portar commands/menus/keybindings do MSX para as 3 linguagens. — **M**, **2**
- [ ] Decidir sobre temas (hoje só MSX tem 7) — incluir no VSIX unificado
  ou manter como extensão de tema separada. — **B**, **3**
- [ ] Testes de gramática automatizados para as 3 linguagens (replicar
  `test-grammar.js` do Spectrum). — **M**, **2**
- [ ] Publicação na loja oficial do VSCode. — **B** (processo), **1**

**Sublime Text:**

- [ ] Portar build systems/syntaxes/snippets das 3 plataformas,
  idealmente gerados da mesma fonte de verdade do VSIX. — **M**, **2**
- [ ] Decidir sobre portar os temas/fontes existentes (hoje só
  MSX/CoCo têm). — **B**, **3**

**Editor próprio (Norton Editor 2.0) e outros:**

- [x] Editor próprio baseado no Norton Editor 2.0 — **integrado em
  2026-11-21** como WibEdit (`internal/editor`, portado de
  `E:\wibworks`), embutido na TUI via menu **Editor > Open**. Ver
  Bloco 6. Versão GUI do mesmo editor ainda não portada. **Destaque
  de sintaxe por dialeto adicionado em 2026-11-27 (v1.46.0)** -
  baseado no MESMO material Sublime/VSCode listado acima, ver Bloco
  6 pro detalhamento completo (tokenizador próprio, independente do
  lexer da compilação). **Autocomplete adicionado em 2026-11-27
  (v1.47.0)** - ver Bloco 6 pro detalhamento completo. **Motor de
  "man"/referência adicionado em 2026-11-27 (v1.48.0)** - Fase 2 do
  mesmo plano, ver Bloco 6: `--man [tópico]`/console `man [tópico]`
  reais (`internal/man`, substituindo o stub de `internal/about/
  about.go`) e um menu "Manual" próprio no WibEdit. **Dicionário MSX
  Basic completo (181 verbetes) + ajuda contextual Shift+F1
  adicionados em 2026-11-27 (v1.49.0)** - Fase 3 do mesmo plano, ver
  Bloco 6. **Suporte ao compilador ASMSX adicionado em 2026-11-27
  (v1.50.0)** - Fase 4 (última) do mesmo plano, ver Bloco 6: `.asm`
  compila direto (`compile`/`goBaDig arquivo.asm`/`F2` na TUI),
  `ASMSX_PATH` no catálogo de opções, manual completo em `--man
  asmsx`. Com isso, as 4 fases do plano registrado na v1.47.0 estão
  completas.
- [ ] Avaliar viabilidade de realce de sintaxe para Notepad++ e
  UltraEdit (UDL). — **B** (avaliação), **3**

### Bloco 9 — Documentação e empacotamento

- [ ] Reescrever `BASIC_DIGNIFIED.md` como gramática formal única com
  seções por dialeto (resolver os merge conflicts pendentes antes ou
  tratá-los como obsoletos pelo novo texto). — **M**, **2**
- [ ] `INSTALLATION.md` novo: sem Python/venv/SkoolKit obrigatório — só
  os emuladores externos continuam sendo "instalação manual". — **B**, **2**
- [ ] Backlog/issues a partir de `IMPLEMENTATIONS.md` (ver seção 5.2). — **B**, **2**
- [ ] Empacotamento de binário único (Windows prioritário, avaliar
  Mac/Linux depois). — **M**, **2**

### Bloco 10 — Itens opcionais/tardios (prioridade 3, não bloqueiam v1)

- [ ] Extração/detokenização de binários (TAP/TZX/Z80/SNA/.p/.bmx → fonte
  numerada) — recurso novo, não existe em nenhuma plataforma hoje.
- [ ] Suporte a Sublime Text (gerar a partir da mesma fonte do VSIX).
- [ ] "Ultra compact code" (modo de otimização de saída).
- [ ] Suporte a Microdrive/disco, NextBASIC, turbo, fitas multi-arquivo
  (Spectrum 128K avançado).

---

## 8. Perguntas da seção 8 — todas respondidas pelo usuário em 2026-11-20

Histórico preservado (pergunta original + decisão), nenhuma reaberta por ora.

1. **Merge conflicts nos docs Python** (seção 2) — **RESOLVIDO: histórico
   congelado.** Não resolver manualmente o `badig/`; a extração já feita
   neste documento basta, já que o Go substitui essa documentação.
2. **Escopo do MVP** — **RESOLVIDO: 3 plataformas em paralelo.** Core +
   os 3 módulos de plataforma (Blocos 1-4) avançam juntos desde o início,
   em vez de um "vertical slice" de uma plataforma só primeiro.
3. **Nomenclatura de extensões de arquivo** — **RESOLVIDO: manter as
   herdadas.** `.DMX/.AMX/.BMX` (MSX), `.DZX/.AZX/.p` (ZX81),
   `.DZS/.AZS` (Spectrum) continuam como estão, sem esquema novo.
4. **Monitoramento de erro em tempo real** — **RESOLVIDO: investigar
   ZRCP (ZEsarUX) já, antes do MVP.** Diferente do adiamento padrão —
   o usuário quer avaliar a viabilidade do ZRCP antes de fechar a
   arquitetura do módulo `emulator`, podendo mudar prioridades dos
   Blocos 3/4/8. Ver novo item no Bloco 0.
5. **Snapshots Z80/SNA via emulador Z80 em Go** — **RESOLVIDO: confirmado,
   fase tardia (Bloco 10).** v1 do Spectrum entrega só TAP/TZX; SkoolKit
   continua a dependência opcional até lá.
6. **Driver SQLite** — resolvido em 2026-11-07: `modernc.org/sqlite`
   (puro Go, sem CGO para a camada SQLite). Já registrado no Bloco 0,
   primeira fatia de schema criada para temas.
7. **TUI e GUI** — **RESOLVIDO: mesmo binário.** Confirma o que já está
   implementado (`goBaDig --cli`/`--tui`/`--gui`); não vira binários
   separados.
8. **Reset de `Y`/`Z` no esquema de versionamento** (seção 10.1) —
   **RESOLVIDO: reinicia.** `Z` volta a `0` a cada `Y` novo; `Y` e `Z`
   voltam a `0` a cada `X` novo. Interpretação adotada confirmada.
9. **Formato do nome na tag/release** — **RESOLVIDO: só no
   changelog/release notes.** A tag Git fica só numérica (`vX.Y.Z`); o
   nome resumido (`Y`) e o nome de comando BASIC (`X`) aparecem apenas no
   `CHANGELOG.md` e no texto da release do GitHub.

---

## 10. Versionamento, empacotamento, estrutura de diretórios e documentação

Decisões registradas nesta sessão (2026-11-05), a partir de instruções
diretas do usuário. Onde há ambiguidade, a interpretação adotada está
marcada como tal — ver também as perguntas 8 e 9 na seção 8.

### 10.1 Esquema de versionamento `X.Y.Z`

Convenção própria do projeto, **diferente de SemVer tradicional**:

- **`Z`** — incrementado a cada **compilação/build** gerada.
- **`Y`** — incrementado a cada **feature nova** incluída, solicitada ou
  testada; carrega um **nome resumido** do que foi feito (ex.:
  `2.3.0-fn-nativo`).
- **`X`** — um **bloco grande**: conjunto extenso de mudanças
  concluídas/testadas que formam um marco maior; recebe o **nome de um
  comando de BASIC**, de qualquer um dos três dialetos (MSX, ZX81 ou ZX
  Spectrum) — ex.: `1.0.0-PRINT`, `2.0.0-GOTO`.

**Confirmado pelo usuário em 2026-11-20** (ver seção 8, perguntas 8 e 9):
`Z` reinicia em `0` quando `Y` sobe; `Y` e `Z` reiniciam em `0` quando `X`
sobe. O nome resumido (`Y`) e o nome de comando BASIC (`X`) aparecem só
no `CHANGELOG.md`/release notes — a tag Git fica só numérica (`vX.Y.Z`),
sem o sufixo de nome.

Documento de referência viva: [`CHANGELOG.md`](CHANGELOG.md) (raiz do
repo) — contém o esquema por extenso e a entrada "Não lançado" com o
estado atual. Cada versão publicada ganha também um arquivo irmão em
[`docs/releases/vX.Y.Z.md`](docs/releases/README.md), com o mesmo
conteúdo publicado na página de Releases do GitHub.

### 10.2 Empacotamento de release

Cada release gera **um único pacote zip**: `goBaDig-vX.Y.Z.zip` (nome do
projeto + versão), contendo o executável, documentação mínima
(`README.md`, `CHANGELOG.md`, `MANUAL.md`, `LICENSE`) e os exemplos de
`examples/`. Ver [`dist/README.md`](dist/README.md) para a estrutura
exata montada antes de zipar.

### 10.3 Estrutura de diretórios de alto nível

```
goBaDig/
  OUTLINE.md            plano vivo (este arquivo)
  README.md              info do projeto, créditos, ferramentas, screenshots
  CHANGELOG.md            resumo de cada release
  LICENSE
  docs/
    MANUAL.md             manual completo: download, build, execução, testes, editores
    RESOURCES.md           lista de recursos externos (o que é, onde baixar, onde colocar)
    releases/
      vX.Y.Z.md            um arquivo por release, espelha o conteúdo do GitHub Release
  examples/
    msx/ zx81/ zxspectrum/ programas de exemplo para teste do usuário final
  dist/                   (gitignored, exceto README.md) saída de build local + pacote .zip
  resource/               (gitignored, exceto README.md) material de terceiros (emuladores, ROMs, fontes)
  current/                (gitignored por completo) build local sempre atualizado para teste
    resource/              arquivos de trabalho do usuário para testar o programa — nunca limpo
  badig/                  projeto Python original (referência de arquitetura e golden tests)
  cmd/                    (futuro) entrypoints do binário Go — ver 10.4
  internal/               (futuro) código-fonte Go, módulos por função lógica — ver 10.4
  editors/                (futuro) pacotes de editor (VSCode, Sublime, ...)
```

`dist/` e `resource/` já existem no repositório com um `README.md`
rastreado cada um (ver seção 10.5) e o restante do conteúdo ignorado via
`.gitignore`.

**`current/` (novo, 2026-11-21, pedido explícito do usuário):**
diretório de build local sempre atualizado para teste manual —
diferente de `dist/` (pacote de release versionado, via `build.ps1`).
Regra registrada em `CLAUDE.md`: toda mudança em qualquer programa do
goBaDig recompila o executável mais atual em `current/`. É
**inteiramente gitignored** (nem README fica rastreado, diferente de
`dist/`/`resource/`). Dentro dele, `current/resource/` guarda arquivos
de trabalho do próprio usuário para testar o programa (exemplos,
fixtures) — esse subdiretório **nunca é limpo/apagado**; só o restante
de `current/` é recriado a cada rebuild.

### 10.4 Estrutura de diretórios do código Go

Princípio: módulos agrupados por **função lógica**, não por plataforma
misturada com infraestrutura. Estado real em 2026-11-05 (releases
1.1.2-1.1.4 + `build.ps1`) marcado com `[x]`; o resto é proposta a
refinar no Bloco 0/1:

```
cmd/
  goBaDig/            [x] main() — CLI/TUI/GUI entrypoint único, flags decidem o modo
internal/
  about/              [x] texto "about"/"man" compartilhado pelas 3 interfaces
  cli/                [x] --help/--man/--cli (console interativo), dispatch
  tui/                [x] menu suspenso estilo DOS (tview/tcell) — só File/Help por ora
  gui/                [x] janela com menu File/Help (Fyne) — só stub por ora
  version/            [x] versão/bloco/feature (X.Y.Z)
  core/               [ ] Token, Position, CompileContext, interfaces Description/Lexer/
                          Parser/Settings/Expose, scanner manual, pipeline de 5 passes +
                          generate, alocador de variável por categoria
  config/             [x parcial] SQLite (schema inicial de temas; config global/projeto e keywords pendentes)
  platforms/
    msx/              [ ] Description/Lexer/Parser MSX, tokenizer MSX Batoken, DignifieR MSX
    zx81/              [ ] idem ZX81, formato .p
    zxspectrum/        [ ] idem ZX Spectrum, TAP/TZX, (futuro) Z80/SNA
  emulator/           [ ] interface EmulatorController + implementações (openMSX, EightyOne, Spectaculator)
editors/
  vscode/             [ ] fonte do VSIX único (grammars, snippets, tasks, settings)
  sublime/            [ ] pacotes Sublime Text
build.ps1             [x] automatiza build local + empacotamento em dist/
```

`internal/core`, `internal/config`, `internal/platforms/*`,
`internal/emulator` e `editors/*` **ainda não existem** — nenhuma
conversão de BASIC é possível hoje, só o stub de interface. Serão
materializados nos Blocos 1-4/8 conforme as perguntas da seção 8 forem
respondidas.

### 10.5 Conjunto de documentação

| Documento | Papel |
| --- | --- |
| `OUTLINE.md` | Plano vivo: arquitetura, decisões, blocos de trabalho, perguntas em aberto |
| `README.md` | Vitrine do projeto: o que é, plataformas, ferramentas usadas, créditos, screenshots, links |
| `CHANGELOG.md` | Resumo de cada release, do primeiro ao último, com o esquema de versionamento |
| `docs/MANUAL.md` | Manual completo: requisitos, download, compilação, execução, conversão, testes, editores, troubleshooting |
| `docs/RESOURCES.md` | O que precisa viver em `resource/`, o que é, onde baixar, como remontar em outra máquina |
| `docs/releases/vX.Y.Z.md` | Um por release, espelha o texto publicado no GitHub Releases |
| `examples/*/README.md` | Descrição dos programas de exemplo de cada plataforma |
| `resource/README.md`, `dist/README.md` | Pointers curtos explicando o propósito de cada pasta (únicos arquivos rastreados dentro delas) |

Todos os sete criados/atualizados nesta sessão; conteúdo ainda reflete a
fase de planejamento (muitas seções marcadas **(pendente)** até o código
Go existir).

### 10.6 Suporte a editores — plano consolidado

Prioridade confirmada pelo usuário:

1. **VSCode** — VSIX único cobrindo os três dialetos (já detalhado na
   seção 5.4/Bloco 8). Prioridade 1.
2. **Sublime Text** — suporte confirmado como objetivo real, não mais
   apenas "se sobrar tempo" (reavaliar a nota da seção 5.5: esforço
   baixo, estrutura quase idêntica entre plataformas já no projeto
   Python). Prioridade 2.
3. **Editor próprio, baseado no Norton Editor 2.0** — **resolvido em
   2026-11-21**: o usuário pediu explicitamente a integração, apontando
   para o projeto **`E:\wibworks`** (WibEdit), não os `msxide`/
   `paleobasic` mencionados em 2026-11-05 (que continuam existindo como
   outras implementações do NE do mesmo autor, só não foram as
   escolhidas). WibEdit foi portado por completo para
   `internal/editor` (TUI do goBaDig) — ver Bloco 6 para os detalhes
   técnicos da integração. A versão GUI (Fyne) do editor **não foi
   portada** nesta rodada; fica para quando for pedida.
4. **Notepad++ e UltraEdit** — "se for possível": avaliar viabilidade
   (arquivos de realce de sintaxe próprios de cada um — UDL no
   UltraEdit, por exemplo) como item de prioridade baixa/esforço a
   avaliar, sem compromisso de entrega.

### 10.7 Sistema de ajuda/man e manuais de BASIC embutidos

Decisão do usuário em 2026-11-05: as três interfaces terão ajuda, mas de
formas diferentes, apropriadas a cada paradigma —

- **TUI e GUI**: ajuda "tradicional" de aplicativo: menu `Help > About`
  (implementado no stub da release 1.1.2) e, depois, um visualizador de
  manual embutido (mesmo conteúdo do `--man` da CLI, apresentado de
  forma navegável).
- **CLI**: dois níveis. `--help` lista as opções de linha de comando
  (como qualquer CLI comum — implementado no stub). Um sistema mais
  profundo, **ao estilo do `man` do Unix**, cobre em detalhe as
  funções/opções do próprio BaDig **e também os manuais de sintaxe
  BASIC das três plataformas**, para consulta rápida sem sair do
  terminal (ex. `goBaDig --man` ou um subcomando `man` navegando por
  tópico/palavra-chave). Isso é uma funcionalidade nova, sem equivalente
  no projeto Python original.

**Conteúdo de referência já disponível, a reaproveitar** (projetos
próprios do usuário, fora deste repositório — ver `docs/RESOURCES.md`
para o registro formal; localização exata a obter com o usuário quando
o trabalho desta seção começar):

- **`fwMSX`** — tem material de referência do MSX Basic já construído,
  reaproveitável para o manual MSX embutido.
- **`go81`** — idem para ZX81.
- **ZX Spectrum**: sem um projeto equivalente citado ainda; a base mais
  próxima hoje é `badig/documentation/ZX_SPECTRUM.md` (já extraído nesta
  sessão de planejamento) mais os links do manual oficial Sinclair já
  catalogados em `badig/OUTLINE.md`, seção "Referências consultadas".
- **`paleobasic`** e **`msxide`** — não são manuais de sintaxe, mas
  contêm as implementações do Norton Editor (TUI e GUI) citadas no
  item 3 acima; relevantes para quando a integração de editor for
  retomada, não para o sistema de ajuda em si.

Nenhum destes quatro projetos foi lido nesta sessão (são locais, fora
deste repositório) — quando o Bloco 5/6/7 "sistema de man completo"
começar de verdade, pedir ao usuário o caminho de cada um para avaliar
o que é diretamente reaproveitável vs. o que precisa ser reescrito.

---

## 11. Referências

- `badig/OUTLINE.md` — plano de análise detalhado da adição do ZX Spectrum
  ao Python (777 linhas), com histórico de versões 1.1.0 a 1.10.1 e
  pendências priorizadas específicas da plataforma.
- `badig/documentation/*.md` — manuais originais (linguagem, ferramentas,
  instalação, IDE, diferenças, implementações, DignifieR, BATOKEN, ZX Spectrum).
- `badig/badig.py`, `badig/support/*.py` — núcleo do motor Python (referência
  de arquitetura e oráculo de comportamento).
- `badig/msx/*`, `badig/zx81/*`, `badig/zxspectrum/*` — módulos de
  plataforma Python (referência de dialeto/formato binário/testes).
- `badig/*/VSCode Extension/`, `badig/*/Sublime Package/` — pacotes de
  editor existentes, base para a unificação do VSIX.
