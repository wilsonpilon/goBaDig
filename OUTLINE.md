# goBaDig — Basic Dignified Suite em Go: plano vivo

> Documento vivo. Atualizar sempre que uma decisão for tomada, uma tarefa for
> concluída ou escopo mudar. Não apagar histórico de decisões — marcar como
> superado quando necessário. Ver também `resume.txt` (não versionado) para
> retomar o trabalho em outra máquina/IA com um resumo rápido do estado atual.

Data da primeira análise: 2026-11-05. Última atualização: 2026-11-21.
Estado: **v1.7.0 publicado** (GitHub Releases) — CLI/TUI/GUI com temas
persistidos, TUI no visual retrô MS-DOS (ver Bloco 6), editor WibEdit
integrado (`internal/editor`, menu Editor > Open, base do futuro
IDE), viabilidade do ZRCP confirmada (Bloco 0, não implementado
ainda) — histórico completo de cada um desses itens em `CHANGELOG.md`
e nos Blocos correspondentes abaixo, não repetido aqui.

**O motor de conversão começou a existir em 2026-11-21**: `internal/core`
(Bloco 1), **tokenizer MSX Batoken** (`internal/platforms/msx/tokenizer.go`,
13 vetores de ouro do Bloco 0 byte-exato), **`Description`/`Lexer` MSX**
(`description.go`/`lexer.go`), e **`Parser` MSX Pass 3/4/5**
(`parser.go` — alocação de variáveis de 2 letras + as 5 conversões
finais do Classic) — a **primeira conversão Dignified→Classic de
verdade já produzindo texto correto** (`counter=0` → `zz = 0`, com
`~nome` mantido por extenso, `x+=5`→`x = x + 5`, `TRUE`/`FALSE`→
`-1`/`0`), confirmada em testes de ponta a ponta completos
(`Lex`→`Parse`→`Generate`). Quatro desvios reais nas interfaces "v0"
do Bloco 1 descobertos implementando contra o MSX de verdade, todos
documentados no Bloco 1 abaixo. **Ainda falta**: Pass 1 do `Parser`
(separador de chamada Dignified `_`, rendering ainda não entendido),
`generate()` de verdade (espaçamento por palavra reservada, limite de
linha), e ligar tudo isso ao tokenizer já pronto pra fechar o caminho
completo até o binário. Ver Blocos 1-4 na seção 7 para o detalhamento.
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

**Fase 3 — Motor do pipeline + validação — PARCIAL (2026-11-21)** (o
que prova que o desenho funciona):
- [x] Driver do pipeline no core (`internal/core/pipeline.go`:
  `Lex`/`Parse`/`Generate`/`Run`) — **A**, **1**. Simplificação
  deliberada em relação ao Python: o scanner só reconhece espaço em
  branco e fim de linha; qualquer sequência de caracteres não-espaço é
  empacotada como `KindWord` e delegada pro `Lexer.Lex` da plataforma
  classificar — nenhum processamento Dignified de núcleo
  (DEFINE/INCLUDE/etc.) existe ainda, então todo pass hoje só chama o
  hook da plataforma por token, sem nenhuma lógica própria do core.
  Scanner "de verdade" por tabela/trie, com vocabulário real, continua
  sendo trabalho do Bloco 2.
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

**Fase 4 — Config em camadas** (não bloqueia o motor testável das Fases
1-3):
- [ ] Estender `internal/config.Store` já existente (tabelas de config
  global/projeto, keywords/tokens por dialeto) — **M**, **2**
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
- [ ] Pass 1 MSX (separador de chamada Dignified `_` → "instrução de
  chamada") — ainda não está claro o que esse recurso deveria gerar no
  texto Classic final (`.tok` muda, `.val` não; não encontrei nenhuma
  conversão de "_" pra texto "CALL" em nenhum lugar do Python);
  precisa de mais pesquisa antes de portar. — **M**, **1**
- [ ] `generate()` de verdade (regras de espaçamento por palavra
  reservada + limite de 256 chars/linha) — hoje usa a versão mínima
  genérica do Bloco 1 (join com 1 espaço, sem limite de linha). Vai
  precisar de `Description.ReservedKeywords()`/`SymbolsWithoutSpacing()`
  novos (mirrors `c_reserved_kw`/`c_symb_comp`) e trocar o hook
  `Parser.Generate` (hoje um no-op, nunca usado de verdade) por algo
  no formato "espaço extra antes do token N dado o buffer já escrito"
  — o `generate()` original nunca troca/remove tokens, só insere
  espaços num buffer de texto, então o modelo `HookAction`/`Token` de
  hoje não é a forma certa pra isso. — **M**, **2**
- [ ] `Settings`/`Expose` MSX (`-cp`/`-tg`) — só fazem sentido quando o
  `Parser` ler essas flags. — **B**, **2**
- [ ] Portar DignifieR MSX como parser manual (sem lookahead de regex). — **A**, **2**
- [ ] Integração openMSX: lançar processo + disco virtual, **sem**
  monitoramento de erro na v1 (adiar item de risco alto). — **M**, **1**
- [ ] Monitoramento de execução em tempo real (mapeamento de erro para
  linha Dignified) — avaliar viabilidade multiplataforma em Go (pode ser
  melhor que o Python, que só funciona fora do Windows). — **A**, **3**
- [ ] Integração fMSX como backend alternativo ao openMSX. — **M**, **2**
- [ ] Empacotamento `.CAS` (MSX, captura de fita nível BIOS). — **M**, **2**
- [ ] Empacotamento `.DSK` (MSX, imagem MSX-DOS/FAT12 válida — montar
  boot sector/FAT/diretório, mais caro que `.CAS`/TAP/TZX). — **A**, **2**

### Bloco 3 — Módulo ZX81

- [ ] Portar `Description`/`Lexer`/`Parser` ZX81 (1 letra Z→A, `PRINT AT`,
  LET implícito, renumeração com rejeição de salto computado, `remx`). — **M**, **1**
- [ ] Portar tokenizer ZX81, incluindo ponto flutuante oculto (usar
  `test_tokenizer_zx81.py` como golden-master direto). — **M**, **1**
- [ ] Gerar formato `.p` (imagem de memória completa, system vars +
  display file + programa). — **M**, **1**
- [ ] Portar DignifieR ZX81. — **M**, **2**
- [ ] Integração EightyOne básica (lançar `.p`, sem seleção de modelo
  ainda, igual ao Python). — **B**, **1**
- [ ] Implementar tabela Unicode (`-tr`) para ZX81 — **nunca existiu no
  Python**, oportunidade de fazer melhor que o original. — **M**, **2**
- [ ] Validar carregamento real de `.p` em emulador — nunca foi feito nem
  no Python. — **B** (execução manual), **1**
- [ ] Suporte a `.81`/`.p81` (variantes do `.p` sem/com nome de arquivo,
  `.p81` podendo guardar múltiplos programas). — **M**, **2**
- [ ] Integração ZEsarUX e sz81/XZ81 como backends alternativos ao
  EightyOne; ZRCP já avaliado como viável para monitoramento (Bloco 0),
  falta implementar: breakpoint em `PC=0008H`, ler `E_PPC`/`ERR_NR`. — **M**, **2**

### Bloco 4 — Módulo ZX Spectrum

- [ ] Portar `Description`/`Lexer`/`Parser` Spectrum (alocação por
  categoria, LET implícito, PRINT AT, TRUE/FALSE=1/0, DATA com expressões,
  DEF FN/FN com slots ocultos de 6 bytes). — **A**, **1**
- [ ] Portar tokenizer Spectrum, incluindo encoding de número Sinclair
  (usar `math/big.Rat` para precisão, validar contra os 69 testes Python
  como golden-master). — **A**, **1**
- [ ] REMX com fill configurável. — **B**, **2**
- [ ] Empacotamento TAP e TZX (100% mecânico). — **B**, **1**
- [ ] Portar DignifieR Spectrum (AZS→DZS). — **M**, **2**
- [ ] Integração EightyOne/Spectaculator 48K básica. — **M**, **1**
- [ ] Integração ZEsarUX como backend adicional; ZRCP já avaliado como
  viável para monitoramento (Bloco 0), falta implementar: breakpoint em
  `PC=0008H`, ler `PPC`/`SUBPPC`/`ERR_NR`. — **M**, **2**
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
- [ ] **Sistema de "man" completo** (ver seção 10.7): comandos/opções do
  BaDig em detalhe + manuais de sintaxe BASIC do MSX, ZX81 e ZX
  Spectrum embutidos, navegáveis por tópico (ex. `goBaDig --man print`
  ou um modo paginado tipo `less`). Reaproveitar conteúdo de referência
  já existente do usuário (`fwMSX` para MSX, `go81` para ZX81) e a
  documentação Spectrum já extraída de `badig/documentation/ZX_SPECTRUM.md`
  — ver seção 10.7 para detalhes. — **A**, **2**
- [x] Modo de comando interativo (`--cli`, release 1.1.4): console com
  prompt `goBaDig>` aguardando comandos linha a linha (`help`, `man`,
  `about`, `version`, `exit`/`quit`, encerra também em EOF/Ctrl-D),
  além do modo "um tiro" por linha de comando. Código em
  `internal/cli/console.go`, com testes (saída e parada em EOF) e
  teste manual via pipe. Comandos reais de conversão chegam com o
  motor (Blocos 1-4). — 2026-11-05.
- [ ] Detecção automática de plataforma pela extensão do arquivo (com
  `-id`/flag equivalente como override). — **B**, **2**

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
  deste repositório). Ainda não ligado a
  nada específico de BASIC (realce de sintaxe, abertura direto de
  `.bas`/arquivos-fonte das 3 plataformas) — é a base crua do editor;
  isso fica para os próximos passos da IDE.

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
  Bloco 6. Versão GUI do mesmo editor ainda não portada.
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
