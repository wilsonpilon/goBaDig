# goBaDig — Basic Dignified Suite em Go: plano vivo

> Documento vivo. Atualizar sempre que uma decisão for tomada, uma tarefa for
> concluída ou escopo mudar. Não apagar histórico de decisões — marcar como
> superado quando necessário. Ver também `resume.txt` (não versionado) para
> retomar o trabalho em outra máquina/IA com um resumo rápido do estado atual.

Data da primeira análise: 2026-11-05.
Estado: **fase de análise/planejamento concluída. Nenhum código Go escrito ainda.**

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
  **Ação pendente:** decidir com o usuário se esses arquivos devem ser
  resolvidos manualmente antes de servirem de referência, ou se a extração de
  informação (como fizemos aqui) já é suficiente e os arquivos Python não
  precisam mais ser "corrigidos" (já que serão substituídos pelo Go).
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
  fatal — nunca alias silencioso.
- **Transformações de dialeto**: `?`↔`PRINT` (`-cp`), remoção opcional de
  `THEN`/`GOTO` adjacentes (`-tg`, **não aplicável ao Spectrum**, onde
  `THEN 100` não é equivalente ao salto Sinclair), `TRUE`/`FALSE`→`-1`/`0`
  (lógica bit-a-bit, diferente do Spectrum que usa `1`/`0`), tradução
  Unicode→ASCII de glifos de bloco via duas tabelas paralelas de 128 chars.
- **Tokenizer MSX Batoken** (`msxbatoken.py`, 758 linhas, standalone):
  tabela de ~150 tokens (prefixo `FF` para funções estendidas), parsing
  numérico com **Microsoft Binary Format** (não IEEE-754) para float
  single/double precision — peça que exige mais cuidado numérico ao
  portar. Linha BASIC termina em `0x00` (não `0x0D` do Spectrum nem `0x76`
  do ZX81 — três formatos de linha binária incompatíveis entre si).
  Endereço de carga fixo `0x8001`.
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
**3**=opcional/tardio. Nada abaixo está concluído ainda — todo o trabalho
de código Go começa do zero.

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
- [ ] Completar `internal/core`, `internal/config`, `internal/platforms`,
  `internal/emulator` (ver seção 10.4) — esta fatia inicial só cobre
  CLI/TUI/GUI stub, não o motor. — **B**, **1**
- [ ] Escolher driver SQLite (puro-Go, ex. `modernc.org/sqlite`, para
  evitar CGO e manter build de binário único simples). — **B**, **1**
- [ ] Desenhar schema SQLite inicial: config global, config por projeto,
  tabela de keywords/tokens por dialeto, tabela de settings de emulador/
  tokenizer por plataforma. — **M**, **1**
- [ ] Gerar vetores de ouro (golden tests) do MSX rodando o Python
  original, já que não há suíte de testes MSX hoje. — **M**, **1**
- [ ] Decidir nomenclatura final de extensões de arquivo por plataforma
  no Go (reaproveitar `.DMX/.AMX/.BMX`, `.DZX/.AZX/.p`, `.DZS/.AZS/.TAP/.TZX`,
  ou propor novo esquema unificado). — **B**, **1**

### Bloco 1 — Núcleo do motor (core engine)

- [ ] Definir interfaces Go: `Description`, `Lexer`, `Parser`, `Settings`,
  `Expose`, `Tools.Interface`. — **M**, **1**
- [ ] Implementar `Token`/`Position` como structs Go. — **B**, **1**
- [ ] Escrever o scanner/lexer manual (substituindo a regex de
  maximal-munch Python). — **A**, **1**
- [ ] Implementar os 5 passes do parser + `generate()` como pipeline
  explícito sobre `*CompileContext`. — **A**, **1**
- [ ] Implementar alocador de variáveis por categoria (generaliza MSX/
  ZX81/Spectrum). — **M**, **1**
- [ ] Implementar config em camadas (SQLite global < projeto < CLI <
  remtags no código-fonte). — **M**, **1**
- [ ] Implementar `Infolog`/relatório de erro com exit codes corretos
  desde o início. — **B**, **1**
- [ ] Suporte a `INCLUDE` recursivo com tabelas de símbolo compartilhadas
  por referência explícita. — **M**, **2**
- [ ] Caminho `-asc` (entrada já numerada, pula lexer/parser Dignified). — **B**, **1**

### Bloco 2 — Módulo MSX

- [ ] Portar `Description`/`Lexer`/`Parser` MSX (vocabulário, alocação de
  2 letras, `?`/PRINT, THEN/GOTO strip, TRUE/FALSE=-1/0, Unicode→ASCII). — **M**, **1**
- [ ] Portar tokenizer MSX Batoken, incluindo Microsoft Binary Format
  para floats (usar golden vectors gerados no Bloco 0). — **A**, **1**
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
  EightyOne (avaliar ZRCP do ZEsarUX para monitoramento). — **M**, **2**

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
- [ ] Integração ZEsarUX como backend adicional (avaliar ZRCP para
  monitoramento de erro em tempo real). — **M**, **2**
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
  `internal/version`. — 2026-11-05.
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
- [ ] Modo de comando interativo (REPL simples) além do modo "um tiro"
  por linha de comando. — **M**, **2**
- [ ] Detecção automática de plataforma pela extensão do arquivo (com
  `-id`/flag equivalente como override). — **B**, **2**

### Bloco 6 — TUI (tview/tcell)

- [x] **Stub** (release 1.1.2): menu `File > Exit` e `Help > About`,
  navegação por `Enter`/`Esc`. Código em `internal/tui`. — 2026-11-05.
- [ ] Tela de configuração (equivalente ao que hoje são os `.ini`/remtags):
  editar config SQLite global/projeto. — **M**, **2**
- [ ] Tela de pré-processamento em lote (escolher arquivos, plataforma,
  opções, rodar e ver log/erros). — **M**, **2**
- [ ] Navegação entre MSX/ZX81/Spectrum sem reiniciar o programa. — **B**, **2**
- [ ] Visualizador de manual BASIC embutido (mesmo conteúdo do `--man`
  da CLI), acessível pelo menu Help. — **M**, **2**
- [ ] Avaliar reaproveitar o **Norton Editor para TUI já pronto no
  projeto `msxide`** do usuário como base do editor de texto integrado
  (ver seção 10.7) — ainda fora de escopo de implementação, só
  registrado como opção futura. — **A**, **3**

### Bloco 7 — GUI (Fyne)

- [x] **Stub** (release 1.1.2): janela com menu `File > Exit` e
  `Help > About` (diálogo nativo Fyne). Código em `internal/gui`. — 2026-11-05.
- [ ] Janela principal com gerenciador de projetos/arquivos. — **M**, **2**
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

- [ ] Editor próprio baseado no Norton Editor 2.0 — **fora de escopo
  agora**, o usuário já tem uma base pronta em outro local; não iniciar
  sem pedido explícito. — **A**, **3**
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

## 8. Perguntas em aberto para o usuário

1. **Merge conflicts nos docs Python** (seção 2): resolver manualmente o
   `badig/` antes de seguir usando como referência, ou tratar como
   histórico "congelado" já que o Go vai substituir a documentação mesmo?
2. **Escopo do MVP**: confirma que a primeira entrega deve cobrir as 3
   plataformas em paralelo (texto Dignified → Classic → binário → fita/
   emulador básico), ou prefere um "vertical slice" completo de uma
   plataforma por vez (ex.: Spectrum primeiro, por já ser o mais avançado
   e ter mais testes) antes de generalizar o núcleo?
3. **Nomenclatura de extensões de arquivo** no Go: manter as extensões
   herdadas (`.DMX/.AMX/.BMX`, `.DZX/.AZX/.p`, `.DZS/.AZS`) ou propor um
   esquema novo mais consistente entre plataformas?
4. **Monitoramento de erro em tempo real** (recurso avançado do MSX via
   openMSX): vale o investimento de engenharia para portar/generalizar
   para as 3 plataformas, ou fica como recurso específico do MSX por ora?
   **Atualizado em 2026-11-05**: o ZEsarUX (novo emulador adicionado ao
   escopo) tem ZRCP, um protocolo de controle remoto via TCP — candidato
   real para viabilizar esse recurso também em ZX81/Spectrum, o que não
   existia como opção quando a pergunta foi escrita (só EightyOne, sem
   hook nenhum). Vale a pena investigar o ZRCP antes de decidir.
5. **Snapshots Z80/SNA via emulador Z80 em Go** (eliminando SkoolKit):
   confirma que isso fica para fase tardia (Bloco 10), e que a v1 do
   Spectrum entrega só TAP/TZX?
6. **Driver SQLite**: aceitar `modernc.org/sqlite` (puro Go, sem CGO, mais
   fácil de distribuir como binário único) ou preferir `mattn/go-sqlite3`
   (CGO, mais maduro mas exige toolchain C no build)?
7. **TUI e GUI**: rodam como modos do mesmo binário (`goBaDig --tui`,
   `goBaDig --gui`) ou como binários/comandos separados?
8. **Reset de `Y`/`Z` no esquema de versionamento** (seção 10.1):
   confirma que `Z` reinicia em `0` a cada novo `Y`, e `Y`/`Z` reiniciam
   em `0` a cada novo `X`? Ou os contadores devem ser cumulativos
   (nunca reiniciar)?
9. **Formato do nome na tag/release**: o nome resumido de `Y` e o nome
   de comando BASIC de `X` entram só no `CHANGELOG.md`/release notes
   (ex. "v2.3.0 — fn-nativo"), ou também compõem a tag Git literal
   (ex. `v2.3.0-fn-nativo`)?

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

**Interpretação adotada** (a confirmar): `Z` reinicia em `0` quando `Y`
sobe; `Y` e `Z` reiniciam em `0` quando `X` sobe. O nome resumido (`Y`) e
o nome de comando BASIC (`X`) são sufixos informativos na tag/changelog,
não alteram a ordenação numérica semântica de `X.Y.Z`.

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
  badig/                  projeto Python original (referência de arquitetura e golden tests)
  cmd/                    (futuro) entrypoints do binário Go — ver 10.4
  internal/               (futuro) código-fonte Go, módulos por função lógica — ver 10.4
  editors/                (futuro) pacotes de editor (VSCode, Sublime, ...)
```

`dist/` e `resource/` já existem no repositório com um `README.md`
rastreado cada um (ver seção 10.5) e o restante do conteúdo ignorado via
`.gitignore`.

### 10.4 Estrutura de diretórios do código Go (proposta, ainda não criada)

Princípio: módulos agrupados por **função lógica**, não por plataforma
misturada com infraestrutura. Proposta inicial (a refinar no Bloco 0/1):

```
cmd/
  goBaDig/                main() — CLI/TUI/GUI entrypoint único, flags decidem o modo
internal/
  core/                   Token, Position, CompileContext, interfaces Description/Lexer/Parser/Settings/Expose,
                           scanner manual, pipeline de 5 passes + generate, alocador de variável por categoria
  config/                 camada SQLite (config global/projeto, keywords/tokens por dialeto)
  platforms/
    msx/                  Description/Lexer/Parser MSX, tokenizer MSX Batoken, DignifieR MSX
    zx81/                 idem ZX81, formato .p
    zxspectrum/           idem ZX Spectrum, TAP/TZX, (futuro) Z80/SNA
  emulator/               interface EmulatorController + implementações (openMSX, EightyOne, Spectaculator)
  tui/                    telas tview/tcell
  gui/                    telas Fyne
editors/
  vscode/                 fonte do VSIX único (grammars, snippets, tasks, settings)
  sublime/                pacotes Sublime Text
```

Esta árvore **ainda não foi criada** — fica registrada aqui como plano;
será materializada no Bloco 0/1 conforme as perguntas da seção 11 forem
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
3. **Editor próprio, baseado no Norton Editor 2.0** — o usuário já tem
   bases prontas em outros projetos próprios: **`msxide`** tem um clone
   do Norton Editor tanto para **TUI** quanto uma versão **GUI** com as
   mesmas teclas/funções do NE2, e **`paleobasic`** também contém uma
   implementação do NE. **Continua fora de escopo de implementação por
   enquanto** (instrução explícita do usuário em 2026-11-05: "o Norton
   Editor fica de fora, mas deve ser mencionado") — registrado aqui como
   opção real de reaproveitamento futuro, não um projeto do zero. Não
   iniciar nenhum trabalho de integração aqui até ser pedido
   explicitamente. Ver seção 10.7 e `docs/RESOURCES.md`.
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
