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

## [1.52.0] — 2026-11-27 — bloco **NEW**, feature **asmsx-highlight** (build 0)

Destaque de sintaxe para arquivos `.asm` (Z80/asMSX) no WibEdit, e
integração do manual do asMSX no sistema de "HELP"/`man`: novo dialeto
de sintaxe (mnemônicos e registradores Z80, diretivas do asMSX com ou
sem o `.` opcional, rótulos `nome:`/`.nome:`/`@@nome:`, as cinco formas
de comentário do asMSX e as quatro convenções numéricas da seção 2.3.1
do manual), com uma cor própria para os ~268 **endereços predefinidos
da BIOS e da área de sistema** (`CHPUT`, `HIMEM` etc. - diretivas
`.BIOS`/`.BIOSVARS`, extraídos de `resource/asMSX/src/dura.y`). Em um
`.asm` aberto, `Shift+F1` abre direto a página de um desses endereços
sob o cursor, ou o manual completo do asMSX (`man asmsx`) em qualquer
outra palavra - um atalho de um toque, em vez do índice geral. Novo
`--man asmbios`/console `man asmbios` lista todos os endereços, e `--man
<nome>` resolve um deles diretamente quando não é também um comando de
BASIC já conhecido. Ver `docs/releases/v1.52.0.md`.

---

## [1.51.0] — 2026-11-27 — bloco **NEW**, feature **asm-basic-bridge** (build 0)

Ponte entre um programa Assembly já compilado com o asMSX e o programa
MSX BASIC que vai chamá-lo: nova macro Dignified
`[INCBLOAD]("arquivo.asm")`, que expande para
`BLOAD"ARQUIVO.BIN":DEFUSR=&Hxxxx` em cada compilação, descobrindo o
endereço de entrada automaticamente a partir da diretiva `.START`/
`.ORG` do `.asm` (resolvendo um rótulo via o `.sym` que o asmsx gera).
Sem `,R` no BLOAD, de propósito: a rotina é chamada (repetidamente) via
`USR()`, não disparada uma única vez ao carregar. Nada é escrito no
`.dmx` - a expansão acontece só em memória, a cada compilação. Ver
`docs/releases/v1.51.0.md`.

---

## [1.50.0] — 2026-11-27 — bloco **NEW**, feature **asmsx-support** (build 0)

**Fase 4** (última) do plano de 4 fases (autocomplete + man/help +
ASMSX): suporte ao compilador Z80 Assembly **asMSX** dentro de
projetos MSX BASIC existentes - não uma reimplementação, só
integração: configuração de caminho, compilação via subprocess, e o
manual completo do asMSX embutido no sistema de referência das Fases
2/3.

### Adicionado

- `docs/man/asmsx.md`: cópia do manual do asMSX
  (`resource/asMSX/doc/asmsx.md`, 1148 linhas), com uma nota de
  atribuição (projeto GPLv3, originalmente de Eduardo "pitpan" A.
  Robsy Petrus, hoje mantido pelo asMSX team -
  github.com/Fubukimaru/asMSX) logo após o título, já que `resource/`
  não é distribuído com o repositório. Registrado como um tópico só
  (`"asmsx"`) via `internal/man/registry.go`'s nova
  `addMarkdownTopic` - ao contrário de `addMarkdownSource` (Fase 2),
  que divide um arquivo em vários tópicos por heading de nível 1, esta
  mantém o manual inteiro como um único tópico sob uma chave curta
  explícita, preservando o título real do manual
  ("asMSX - MSX cross assembler") só para exibição. `Index()` também
  precisou passar a usar a chave do registro (não o título de exibição
  da página) no `TOCEntry` de cada tópico, pra navegação funcionar
  igual nos dois casos.
- `internal/options/options.go`: novo campo `ASMSX_PATH` (categoria
  MSX) - caminho do executável do asMSX, sem validação de existência
  de arquivo (mesmo padrão não-especial que `EMULATOR_PATH` já segue).
  Aparece automaticamente nos painéis `Setup > MSX` da TUI/GUI, que já
  renderizam genericamente a partir do catálogo.
- `internal/platforms/msx/asmsx.go`: `ResolveAsmsxPath` (cascata -
  `ASMSX_PATH` configurado e existente, senão `PATH` do sistema -
  `EMULATOR_PATH` já modelava essa ideia, mas nenhum dos seus campos
  chegou a ser lido por código de produção até agora; este é o
  primeiro) e `CompileAsm` (mesmo padrão síncrono-com-captura-de-saída
  de `internal/editor/printer_unix.go` - único outro precedente no
  projeto, já que todo lançamento de emulador hoje é fire-and-forget).
  Compilação bem-sucedida reporta o `.bin` gerado; outros formatos de
  saída (`.rom`/`.cas`/`.wav`/`.com`) dependem de diretiva dentro do
  próprio `.asm` e não são detectados automaticamente (limitação
  conhecida de v1).
- Dispatch por extensão pra `.asm` (deliberadamente **fora** do
  `switch` fechado de `internal/dialect.go`, que é só pra dialetos
  Dignified) em três lugares: `internal/filemanager/actions.go`'s
  `compileSelected` (F2 na TUI), `internal/cli/console.go`'s
  `executeCompile` (`compile arquivo.asm` no console) e
  `internal/cli/cli.go`'s `runCompile` (`goBaDig arquivo.asm` direto).

### Verificado

- `internal/platforms/msx/asmsx_test.go`: `ResolveAsmsxPath` (caminho
  configurado existente, configurado mas inexistente, fallback pro
  PATH, nem configurado nem no PATH); `CompileAsm` testado contra uma
  ferramenta **real** (um binário Go minúsculo compilado em tempo de
  teste, já que não há `asmsx` real garantido no ambiente de CI/dev) -
  sucesso com `.bin` gerado, saída capturada em caso de erro, aviso
  quando a ferramenta termina sem erro mas não gera o `.bin`.
- `internal/man/dict_test.go`: `Lookup("asmsx")` (case-insensitive),
  `Index()` linkando pela chave curta mas exibindo o título real.
- Testes de dispatch em cada um dos três pontos de integração
  (`internal/filemanager`, `internal/cli` × 2), confirmando que
  `.asm` é roteado pro ASMSX e não cai no erro de "dialeto não
  reconhecido" do `dialect.CompileByExtension`.
- Testado de ponta a ponta nesta máquina contra um `asmsx.exe` real
  já instalado (achado via `PATH`) - confirmou a cascata de resolução
  e o encapsulamento de erro funcionando (um lock de arquivo
  específico desta máquina impediu a compilação de fato rodar, mas
  confirmou que o erro do subprocesso é capturado e reportado
  corretamente em vez de travar o programa).
- `go build`/`go vet`/`gofmt`/`go test ./...` limpos no repositório
  inteiro.

## [1.49.0] — 2026-11-27 — bloco **NEW**, feature **msx-dictionary** (build 0)

**Fase 3** do plano de 4 fases (autocomplete + man/help + ASMSX):
dicionário completo de BASIC do MSX (as 181 palavras reservadas do
livro "Linguagem BASIC MSX" + extensão MSX2+/FM, transcritas no
projeto próprio do usuário, MSXIDE) e ajuda contextual (`Shift+F1`) no
WibEdit. Conteúdo trazido e **commitado publicamente** no repositório
- decisão explícita do usuário, avisado sobre a exposição de conteúdo
de terceiros antes de prosseguir (ver `OUTLINE.md`).

### Adicionado

- `internal/man/msxdict_data.go` (gerado, não escrito à mão):
  conversão de `E:\msxide\ajuda\MsxBasicDictData.pbi` (140 verbetes,
  "Linguagem BASIC MSX", Denise Santoro Cruz, Editora Aleph/Gradiente,
  4a edição 1986) e `MsxBasic2PlusDictData.pbi` (41 verbetes, "Manual
  MSX2+ FM", ACVS Eletrônica) - um conversor Go descartável (não faz
  parte do build) parseou as chamadas `MSXDict_Add(...)`/
  `MSXDict_Add2Plus(...)` do PureBasic original (balanceando parênteses/
  aspas, resolvendo a concatenação de strings e os atalhos `MSXQ`/
  `#CRLF$`) e gerou a tabela Go literal.
- `internal/man/dict.go`: `MsxDictEntry` (título, se é função/avançado,
  origem em inglês, resumo, formato, exemplo, função, programa exemplo,
  página do livro, sistema MSX1/MSX2+), `(MsxDictEntry).Page()` (molde
  fixo NOME/FORMATO/EXEMPLO/FUNÇÃO/PROGRAMA EXEMPLO/REFERÊNCIA, mesmo
  tipo `Page` que o Markdown produz - CLI e WibEdit não precisam saber
  a diferença), `FindMsxDictEntry` (match exato case-insensitive, com
  fallback pra verbetes compostos - ex. `"RETURN"` resolve
  `"GOSUB-RETURN"`), `MsxDictIndex` (agrupado por letra inicial) e
  `AutocompleteWords` (normaliza títulos compostos/com sufixo
  `"(MSX2+)"` em palavras individuais, pro autocomplete abaixo).
  `Lookup`/`Index` (Fase 2) passam a tentar o dicionário primeiro, e o
  índice geral ganhou um link `"msx"` pra o índice do dicionário.
- WibEdit: `Shift+F1` abre a página do dicionário pra palavra sob o
  cursor (normaliza `"?"` pra `"PRINT"`, o atalho clássico do MSX
  BASIC); fora do dialeto MSX, cai no índice geral em vez de não fazer
  nada. Autocomplete (Fase 1) passou a preferir as palavras do
  dicionário pra MSX, mantendo as tabelas de sintaxe como reforço -
  cobre palavras que só existem num verbete composto (ex. digitar
  `RET` agora sugere `RETURN`, que antes só existia dentro do verbete
  único `"GOSUB-RETURN"`).

### Corrigido no caminho (achado implementando, não planejado)

- Uma primeira versão de `(MsxDictEntry).Page()` juntava um campo
  inteiro (ex. `Função`) numa única string e quebrava linha a partir
  do zero, apagando as quebras de linha/linhas em branco que o próprio
  livro usa pra alinhar diagramas de fluxo (`GOSUB-RETURN`) e tabelas
  comparativas (`AUTO`). Corrigido quebrando primeiro pelas quebras de
  linha já existentes no texto de origem, e só reaplicando o wrap
  linha a linha quando uma linha individual excede a largura.

### Verificado

- `internal/man/dict_test.go`: contagem plausível de verbetes
  (181 = 140 MSX1 + 41 MSX2+, protege contra um arquivo gerado vazio/
  truncado passar despercebido), match exato case-insensitive, match
  composto (`RETURN`→`GOSUB-RETURN`), palavra-chave desconhecida, ordem
  fixa das seções de `Page()`, agrupamento por letra do índice,
  prioridade do dicionário em `Lookup`, pseudo-tópico `"msx"`.
- `internal/editor`: `Shift+F1` abrindo o verbete certo, normalização
  de `"?"`, aviso pra palavra desconhecida, fallback fora do MSX, F1
  puro continuando a abrir o menu de ajuda sem ser afetado.
- `go build`/`go vet`/`gofmt`/`go test ./...` limpos no repositório
  inteiro.

## [1.48.0] — 2026-11-27 — bloco **NEW**, feature **man-reference** (build 0)

**Fase 2** do plano de 4 fases anunciado na v1.47.0 (autocomplete +
man/help + ASMSX): motor genérico de "man"/referência, usado tanto na
CLI/console ("--man"/"man") quanto num visualizador novo dentro do
WibEdit. Ainda não traz o dicionário completo de BASIC do MSX (Fase 3)
nem o manual do ASMSX (Fase 4) - esta rodada entrega a infraestrutura
compartilhada e um primeiro conteúdo real: os próprios comandos do
goBaDig.

### Adicionado

- `internal/man`: modelo compartilhado de "página de manual" (`Page`/
  `Line`/`TOCEntry`) e um parser de Markdown escrito à mão
  (`ParseMarkdown`) - inspirado conceitualmente no
  `BuildMarkdownBufferFromText` do MSXIDE (projeto próprio do usuário,
  estudado nesta mesma rodada da Fase 1), mas implementado do zero em
  Go. Suporta headings `#`.."######", blocos de código cercados
  (` ``` `), citações (`>`), listas (`-`/`*`/`N.`) e parágrafos com
  quebra de linha automática; tabelas e ênfase inline ficam como texto
  literal nesta versão.
- `docs/man/badig.md` + `docs/man/embed.go` (pacote `docsman`): a
  referência dos próprios comandos do goBaDig (`compile`, `execute`,
  `set`, `settings`, `dialect`, `theme`, `version`, `man`) no formato
  NOME/SINOPSE/DESCRIÇÃO, um `# tópico` por comando - escrita a partir
  do comportamento real atual. Fica em `docs/` (não `internal/`)
  porque `go:embed` não alcança diretórios pai - o conteúdo versionado
  continua visível em `docs/` como `docs/MANUAL.md`, mas o
  `//go:embed` precisa estar ao lado do `.md`.
- `internal/man/registry.go`: `Lookup(tópico)`/`Index()`, dividindo
  `docs/man/badig.md` pelos headings de nível 1 - cada um se torna um
  tópico independente, embora o arquivo inteiro se leia como um único
  documento Markdown normal no GitHub/editor. Pronto para a Fase 3
  acrescentar o dicionário MSX e a Fase 4 o manual do ASMSX do mesmo
  jeito.
- CLI: `--man [tópico]` (antes só mostrava o "about" do projeto) e o
  console ganhou "man [tópico]" de verdade (antes ignorava o
  argumento) - ambos com autocompletar Tab dos tópicos disponíveis
  (`internal/cli/completion.go`).
- WibEdit: visualizador novo em tela cheia (`internal/editor/
  view_man.go`), aberto por um menu "Manual" próprio (tecla M) -
  deliberadamente **fora** do menu "F1=Help" e da família
  `neMenus`/`neHelpMenu` existente, que é uma reprodução fiel de um
  editor de referência específico (NE2) e não deveria ganhar itens que
  não fazem parte dela. Navegação real: `Up`/`Down`/`PgUp`/`PgDn`/
  `Home`/`End` rolam, `Enter` numa linha de índice abre o tópico
  (empilhando histórico), `Esc` volta um nível ou fecha quando não há
  mais histórico. Modal enquanto aberto - nenhuma outra tecla do
  editor chega ao documento.

### Verificado

- `internal/man`: parser de Markdown contra cada elemento suportado,
  quebra de linha de parágrafos longos, `Lookup`/`Index` (incluindo
  case-insensitive e tópico desconhecido).
- `internal/cli`: `--man`/`man` sem tópico, com tópico válido, com
  tópico inválido (console e `--man`), autocompletar de tópicos.
- `internal/editor`: abertura do visualizador, navegação Up/Down,
  Enter abrindo um tópico ligado e Esc voltando/fechando, e que o
  visualizador é mesmo modal (uma tecla qualquer não alcança o
  documento).
- `go build`/`go vet`/`gofmt`/`go test ./...` limpos no repositório
  inteiro.

## [1.47.0] — 2026-11-27 — bloco **NEW**, feature **editor-autocomplete** (build 0)

**Autocomplete no WibEdit**, primeira metade do próximo passo já
anunciado na v1.46.0 ("depois vamos usar o autocomplete e o help
sensível ao contexto"). Pedido do usuário antes de implementar: dar uma
olhada no MSXIDE (projeto próprio dele, FreeBASIC) pra estudar como ele
já resolve help/autocomplete/man - confirmado que a arquitetura é
reaproveitável como conceito (popup por prefixo, mesma fonte de dados
do destaque de sintaxe/dicionário), mas não há código Go a portar (é
FreeBASIC). Decisão de escopo: usar só as tabelas de palavra-chave/
função já existentes (`internal/editor/syntax_keywords.go`), sem
dicionário externo nesta rodada - isso fica para uma rodada futura.

### Adicionado

- `internal/editor/autocomplete.go`: popup de autocomplete por
  prefixo, ancorado na palavra sendo digitada. Reusa as mesmas tabelas
  `msxKeywords`/`msxFunctions`/`zx81Keywords`/`zx81Functions`/
  `zxSpectrumKeywords`/`zxSpectrumFunctions` do destaque de sintaxe
  (v1.46.0) como única fonte de candidatos - nenhum dado novo. `Tab`/
  `Enter` completa a palavra, `Up`/`Down` navega entre as opções,
  `Esc` fecha sem completar; qualquer outra tecla (espaço, pontuação,
  as setas esquerda/direita, Backspace) cai no fluxo normal do editor
  e a lista se atualiza ou fecha sozinha a partir do novo estado do
  cursor/texto.
- `Editor.Capture` (`internal/editor/editor.go`) ganhou um terceiro
  modo de interceptação de tecla que não existia antes: nem "passa
  tudo" (documento normal) nem "bloqueia tudo" (menu/diálogo/outline
  já existentes) - só as cinco teclas acima são roubadas enquanto o
  popup está aberto, o resto sempre cai pro fluxo padrão.

### Corrigido no caminho (achado implementando, não planejado)

- Uma primeira versão do popup usava o mesmo mecanismo do menu
  suspenso (`editorDropdown`): uma página nova em `tview.Pages`. Isso
  quebrava a digitação por completo depois do popup abrir uma vez -
  causa raiz: `tview.Pages.AddPage` **rouba o foco** pra página recém-
  adicionada sempre que o container já tem foco
  (`rivo/tview@v0.42.0/pages.go:88-90`), e `*tview.Box` não tem
  `InputHandler` próprio, então toda tecla seguinte era engolida em
  silêncio. Corrigido abandonando `Pages` pro popup: `Editor` ganhou
  um método `Draw` próprio que desenha o popup diretamente por cima de
  `Pages.Draw`, sem nunca criar uma página nem tocar foco - o foco
  nunca saiu do documento, exatamente a intenção original.

### Verificado

- `internal/editor/autocomplete_test.go`: popup abre com o candidato
  certo (MSX/ZX81/ZX Spectrum), `Tab` confirma e substitui o prefixo,
  `Esc` fecha sem alterar o texto, espaço fecha sozinho (sai da
  palavra), setas navegam entre várias opções com wrap-around, nunca
  abre num documento sem dialeto reconhecido.
- `go build`/`go vet`/`gofmt`/`go test ./...` limpos no repositório
  inteiro.

## [1.46.0] — 2026-11-27 — bloco **NEW**, feature **editor-syntax-highlight** (build 0)

**Destaque de sintaxe no WibEdit**, pedido explícito do usuário:
"Vamos implementar este destaque de sintaxe no Norton Editor que está
dentro do GoBaDig... inicialmente vamos usar apenas o destaque de
sintaxe, mas depois vamos usar o autocomplete e o help sensível ao
contexto". Baseado no material de referência já existente pros
dialetos (`badig/<dialeto>/Sublime Package` e
`badig/<dialeto>/VSCode Extension`), não no lexer da compilação -
aquele é construído pra um passo único, batch, sobre um programa
inteiro e correto; isso aqui roda a cada tecla digitada, sobre texto
possivelmente incompleto ou inválido no meio da edição.

### Adicionado

- `internal/editor/syntax.go`/`syntax_keywords.go`: um tokenizador por
  linha (`tokenizeLine`), independente do pipeline de compilação, com
  tabelas de palavras-chave/funções próprias por dialeto (MSX, ZX81,
  ZX Spectrum), destiladas do material Sublime/VSCode já existente no
  projeto. Seis categorias: palavra-chave, função, comentário, string,
  número, label (`{nome}`) — mais simples que as gramáticas completas
  originais (sem distinguir operadores, sem detecção de "invalid
  input", sem tratar construções Dignified como uma categoria à
  parte) de propósito: o valor visual está majoritariamente nessas
  seis, pelo menor risco de implementação.
- Dialeto escolhido pela extensão do arquivo aberto (`.dmx`/`.amx` =
  MSX, `.dzx`/`.azx` = ZX81, `.dzs`/`.azs` = ZX Spectrum) — tanto o
  fonte Dignified quanto a listagem clássica numerada abrem com
  destaque.
- Seis cores novas (`internal/editor/editor.go`), com o fundo sempre
  sincronizado com `contentStyle` (`options_display.go`), pra
  continuar combinando com qualquer um dos 5 perfis de cor já
  existentes do WibEdit — esse sistema de cores é independente do
  `internal/theme`, igual já documentado.
- Ligado no próprio laço de desenho do editor
  (`internal/editor/document_view.go`'s `Draw`): tokeniza só as linhas
  realmente visíveis na tela a cada redesenho, não o documento
  inteiro — o custo fica proporcional à altura do terminal, não ao
  tamanho do arquivo.

### Limitação conhecida, documentada de propósito

- O tokenizador não guarda estado entre linhas: comentários de bloco
  de várias linhas (apóstrofo duplo abrindo/fechando, ou
  `###...###`) não são rastreados como abertos/fechados — uma
  palavra em inglês dentro de um desses blocos que por acaso também é
  uma palavra-chave BASIC (`ON`, `TO`, `FOR`, `AND`...) ainda aparece
  destacada, mesmo o bloco inteiro sendo um comentário. Rastrear isso
  direito pediria varrer o documento inteiro (ou um índice por
  documento de "quais linhas estão dentro de um bloco não fechado",
  invalidado a cada edição) — fora de escopo nesta primeira rodada,
  registrado como um possível próximo passo.

### Verificado

- `internal/editor/syntax_test.go`: tokenizador testado
  isoladamente pros três dialetos (palavras-chave, funções, strings,
  números, hex MSX, labels, comentários `'`/`REM`/`##`, variantes de
  `USR`), mais um teste de ponta a ponta
  (`TestOpeningADmxFileActuallyRendersSyntaxColors`) que abre um
  `.dmx` de verdade via `Editor.OpenFile` (o mesmo caminho que o `F4`
  do gerenciador de arquivos usa) e confirma que a palavra-chave e a
  string realmente aparecem com cores diferentes na tela simulada —
  confirmado capaz de pegar uma regressão desligando o gancho de
  propósito.
- Verificação manual contra um programa MSX real
  (`badig/examples/msx/CGK173.dmx`): palavras-chave, números e
  comentários destacados corretamente ao longo do arquivo inteiro.
- `go build`/`go vet`/`gofmt`/`go test ./...` limpos no repositório
  inteiro.

## [1.45.0] — 2026-11-27 — bloco **NEW**, feature **fm-project-reload** (build 0)

**Correção de bug**, reportada diretamente pelo usuário: "ao criar um
projeto novo, às vezes dá erro e não mostra os arquivos no file
manager, mas entrando e saindo do programa o diretório está lá".

### Corrigido

- **Causa raiz**: `config.Store.CreateProject` roda `os.MkdirAll`
  ANTES de inserir no banco — uma falha no passo do banco (o caso
  mais comum é um nome de projeto duplicado, já que `name` é a chave
  primária dessa tabela e persiste entre reinícios no `settings.db`
  real do usuário, então reusar um nome de uma sessão anterior também
  falha aqui) ainda assim deixa um diretório de verdade, recém-criado,
  no disco. O `createProject` do gerenciador de arquivos retornava
  assim que `store.CreateProject` falhava, sem nunca recarregar o
  painel — esse diretório ficava invisível na sessão atual até o
  programa inteiro ser reiniciado (o que lista o sistema de arquivos
  de verdade, revelando que ele estava lá o tempo todo).
- Corrigido recarregando o painel ativo sempre, mesmo quando o passo
  do banco falha.

### Verificado

- Novo `internal/filemanager/filemanager_test.go`
  (`TestCreateProjectStillShowsTheDirectoryWhenTheDatabaseStepFails`)
  — reproduz com um nome duplicado (a forma mais simples e real de
  disparar a falha no passo do banco); confirmado capaz de pegar o
  bug revertendo o fix de propósito.
- `go build`/`go vet`/`gofmt`/`go test ./...` limpos no repositório
  inteiro.

## [1.44.0] — 2026-11-27 — bloco **NEW**, feature **fm-status-reset** (build 0)

**Correção de bug**, reportada diretamente pelo usuário: "a linha de
status ainda não está sendo atualizada a cada opção usada, algumas
alteram a barra de status e a mesma não volta".

### Corrigido

- **Causa raiz**: `F2` (compilar), `F5` (copiar), `F7` (criar
  diretório) e outras ações deixavam uma mensagem de status "presa"
  na tela — o único lugar que reconduzia a barra de volta pro texto
  de atalhos era `Enter`/`Backspace` (navegar diretório). Navegação
  por setas dentro de um painel nunca passa por
  `FileManager.Capture`: `Up`/`Down`/`PageUp`/`PageDown` não estão no
  seu `switch`, então o próprio `InputHandler` do `tview.Table`
  processa essas teclas direto — uma mensagem de ação ficava na tela
  indefinidamente assim que o usuário só movia a seleção pelo
  teclado, sem usar `Enter`/`Backspace`.
- Corrigido com um `SetSelectionChangedFunc` em cada painel (ligado
  em `NewFileManager`), que reconduz a barra pro texto de atalhos
  sempre que a seleção muda — incluindo as mudanças de seleção que
  as próprias ações já disparam via `reload()`/`load()`: a chamada
  final de `setStatus` de cada ação sempre roda DEPOIS do seu próprio
  `reload`/`load`, então sobrescreve esse reset de volta pra mensagem
  certa da ação.

### Verificado

- Novo `internal/filemanager/filemanager_test.go`
  (`TestSelectingADifferentRowResetsTheStatusBar`) — confirmado capaz
  de pegar o bug revertendo o fix de propósito.
- `go build`/`go vet`/`gofmt`/`go test ./...` limpos no repositório
  inteiro.

## [1.43.0] — 2026-11-27 — bloco **NEW**, feature **rem-header-branding** (build 0)

**Cabeçalho REM aponta pro goBaDig, não mais pro projeto Python
original**, pedido explícito do usuário: "a linhas 10 e 20 com o rem
indicando o projeto precisa mudar a primeira linha deve ser REM
Converted by goBaDig 2026 e a segunda linha deve mostra o link do
gobadig no github".

### Alterado

- `internal/core/pipeline.go`'s `remHeaderLine1`/`remHeaderLine2`
  (inseridas nas linhas 10/20 de todo programa convertido quando
  `RemHeader` está ativo - o padrão) deixam de ser as mesmas do
  `support/badig_settings.py` original ("Converted with Basic
  Dignified" / link pro `farique1/basic-dignified`) e passam a ser
  **"Converted by goBaDig 2026"** e
  **`https://github.com/wilsonpilon/goBaDig`** — goBaDig é um port
  independente, não um wrapper do projeto original, então o cabeçalho
  agora reflete isso. Compartilhado entre as três plataformas (MSX,
  ZX81, ZX Spectrum); cada uma usa seu próprio caractere de comentário
  (`'` no MSX, `REM` em ZX81/ZX Spectrum) — confirmado rodando o CLI
  de verdade nos três dialetos.

### Verificado

- `internal/core/pipeline_test.go` e os testes `compile_test.go` das
  três plataformas atualizados pro texto novo.
- Dois testes de validação pessoal, não rastreados pelo git
  (`internal/platforms/msx/zz_scratch_cgk173_test.go`), comparavam
  byte a byte contra a saída real do `badig.py` original — agora
  pulados (`t.Skip`) de propósito nesse ponto específico: essa saída
  de referência é de ANTES dessa mudança de marca, e a divergência no
  cabeçalho é esperada e desejada a partir de agora.
- `goBaDig <arquivo>` rodado de verdade nos três dialetos
  (`.dmx`/`.dzx`/`.dzs`), conferindo o `.amx`/`.azx`/`.azs` resultante
  à mão.
- `go build`/`go vet`/`gofmt`/`go test ./...` limpos no repositório
  inteiro.

## [1.42.1] — 2026-11-26 — bloco **NEW**, feature **fm-control-center** (build 1)

**Duas correções**, reportadas diretamente pelo usuário depois de usar
a v1.42.0.

### Corrigido

- **Barra de status "bagunçada" ao voltar do editor**: "ao ir para o
  editor, voltando a barra de status fica bagunçada, não mostrando
  mais as teclas de atalho". Causa raiz: fechar o editor sempre
  trocava pra página `"app"` (o desktop) incondicionalmente, mesmo
  quando o editor tinha sido aberto a partir do gerenciador de
  arquivos (`F3`/`F4`) - o usuário via a barra de status do DESKTOP
  (`F10 Menu...`), diferente da barra do gerenciador (`Tab Painel...`),
  e percebeu isso como "a barra quebrou". Corrigido com
  `editorReturnPage`, uma variável que guarda de qual página o editor
  foi aberto (`"app"` pelo menu `Editor > Open`, `"filemanager"` pelo
  `F3`/`F4`) e que `closeEditor` agora usa pra voltar pro lugar certo.
- **Projeto novo organizado em pasta própria**: "ao criar projeto,
  crie uma pasta dentro do diretório corrente para organizar os
  arquivos". O campo "Pasta" do diálogo `Ctrl+N` passa a ser a pasta
  **pai** (onde criar o projeto), não mais a pasta final do projeto -
  `createProject` agora sempre cria uma nova subpasta com o nome do
  projeto dentro dela, em vez de jogar os arquivos direto na pasta que
  o usuário estava navegando.

### Verificado

- `internal/tui/new_file_test.go`
  (`TestClosingTheEditorReturnsToTheFileManagerItWasOpenedFrom`) -
  usa `F4` (editar um arquivo já existente), não `F3`: `F3` passa por
  um diálogo `tview.Form` cuja contagem exata de `Enter`s até
  submeter não é uma constante confiável (confirmado durante o
  desenvolvimento - uma sequência de `Enter`s, fixa ou com espera
  ativa, às vezes "passa do ponto" depois que o diálogo fecha,
  caindo dentro do editor já focado e digitando uma quebra de linha
  em vez de clicar em nada).
- `internal/filemanager/filemanager_test.go`
  (`TestCreateProjectCreatesItsOwnSubfolderInsideTheParent`) - confirma
  que a pasta pai fica limpa (só a subpasta do projeto aparece nela).
- `go build`/`go vet`/`gofmt`/`go test ./...` limpos no repositório
  inteiro.

## [1.42.0] — 2026-11-26 — bloco **NEW**, feature **fm-control-center** (build 0)

**O gerenciador de arquivos vira o centro de controle do programa**,
pedido explícito do usuário depois de usar a v1.41.3: "vamos fazer o
File->File Manager e não ter mais o Files. O File Manager agora vai
ser o centro de controle total do programa".

### Adicionado

- **Menu consolidado**: o menu de topo "Files" separado some; o
  gerenciador agora abre por **`File > File Manager`** (primeiro item,
  antes de "Exit").
- **`F3` Criar Arquivo**: um arquivo vazio no diretório atual do
  painel ativo, nomeado pelo usuário (a extensão escolhe o dialeto -
  `.dmx`/`.dzx`/`.dzs` pra um fonte que o próprio gerenciador já
  compila/executa, ou qualquer outra coisa pra um arquivo de texto
  comum). Recusa sobrescrever um arquivo já existente. Aberto na hora
  no WibEdit já embutido.
- **`Ctrl+N` Criar Projeto, revisado**: o arquivo inicial deixa de ser
  opcional (checkbox removida) - todo projeto novo já nasce com um
  ponto de entrada. Deixa de se chamar `<nome do projeto><ext>` e
  passa a se chamar sempre **`main<ext>`** (`main.dmx`/`main.dzx`/
  `main.dzs`), nome fixo e previsível independente de como o projeto
  se chama. Também aberto na hora no WibEdit.
- **Arquivos de interesse ganham cor forte + ícone próprio** num
  painel, pra reconhecer de relance qual é qual: fonte BADIG
  Dignified (`.dmx`/`.dzx`/`.dzs`, ícone `◆`), listagem ASCII clássica
  (`.amx`/`.azx`/`.azs`, ícone `≡`), artefato compilado/empacotado
  (`.bmx`/`.p`/`.tap`/`.tzx`/`.dsk`, ícone `▣`) - cada categoria usa
  uma cor do tema ativo que nenhuma outra parte da TUI já usava
  (`Hover`/`Error`), pra garantir contraste forte em qualquer tema.
- Fecha o ciclo completo pedido pelo usuário inteiramente pelo
  gerenciador de arquivos: **criar → editar → salvar/sair → compilar
  (`F2`) → executar (`F9`)**.

### Verificado

- `internal/tui/new_file_test.go`
  (`TestF3CreatesAFileAndOpensItInTheEditor`) - dirige o diálogo de
  verdade por teclas simuladas (não uma chamada Go direta, que todo
  outro teste do gerenciador usa e que pula o `tview.Form` por
  inteiro); confirma um detalhe real do `tview.Form` registrado no
  próprio teste: `Enter` com um `InputField` em foco não envia o
  formulário na hora - são 3 `Enter`s (campo → foco no botão → clique)
  pra concluir, não 1.
- `internal/filemanager/panel_test.go`
  (`TestFileCategoriesGetTheirOwnColorAndIcon`),
  `internal/filemanager/filemanager_test.go` (testes novos de
  `createFile`, `createProject` com `main<ext>` e abertura automática
  no editor).
- `go build`/`go vet`/`gofmt`/`go test ./...` limpos no repositório
  inteiro.

## [1.41.3] — 2026-11-26 — bloco **NEW**, feature **file-manager** (build 3)

**Correção de bug real**, reportada pelo usuário com uma captura de
tela depois da v1.41.2 (`images/erro-filemanager.png`): o nome de
quase todo diretório no gerenciador de arquivos aparecia em branco —
só a data da linha ficava visível. As duas correções anteriores
(v1.41.1/v1.41.2) miraram num problema de COR porque o primeiro
relato ("preto com preto") parecia um problema de contraste — mas o
problema nunca foi cor nenhuma.

### Corrigido

- **Causa raiz real**: o `tview` trata qualquer trecho `[...]` como
  uma tentativa de tag de estilo `[cor]`/`["região"]` sempre que o
  caractere logo depois do `[` é letra, dígito, `#`, `:` ou `-` —
  casado e CONSUMIDO (zero caracteres visíveis), quer a "cor" dentro
  resolva pra algo real ou não. O formato `"[" + nome + "]"` usado
  pelos diretórios (estilo Norton Commander, igual o Midnight
  Commander/`mc`) colidia direto com isso: `"[..]"` sobrevivia por
  acidente (o caractere depois do `[` é `.`, que não é válido como
  início de tag), mas `"[cmd]"`, `"[docs]"`, `"[internal]"` — qualquer
  diretório real começando com letra ou dígito — desaparecia por
  completo. Confirmado isolando uma cópia local do `rivo/tview` com
  instrumentação (mesma técnica já usada uma vez neste projeto, pro
  bug do `SetBackgroundColor` em `Grid`): `printWithStyle` realmente
  retornava zero caracteres impressos pra `"[aaa]"`, e todos os 4 pra
  `"[..]"`.
- Corrigido com `tview.Escape` (a própria ferramenta documentada do
  `tview` pra esse caso exato) em `internal/filemanager/panel.go`'s
  novo `entryLabel` (nome do arquivo/diretório) e no título do painel
  (caminho do diretório), e também na barra de status (mensagens de
  erro frequentemente citam o caminho que falhou) e na lista de
  projetos salvos (nome/pasta do projeto, digitados pelo usuário).

### Verificado

- Novo `internal/filemanager/panel_test.go`
  (`TestMultipleDirectoryEntriesAreActuallyVisibleOnScreen`) —
  renderiza uma tela simulada de verdade com diretórios de nomes
  comuns (`cmd`, `docs`, `internal`, `aaa`, `resource`) e confirma que
  o rótulo completo de cada um aparece na tela; confirmado capaz de
  reproduzir exatamente a captura de tela do usuário ao reverter o
  fix de propósito.
- `go build`/`go vet`/`gofmt`/`go test ./...` limpos no repositório
  inteiro.

## [1.41.2] — 2026-11-26 — bloco **NEW**, feature **file-manager** (build 2)

**Correção de bug**, reportada diretamente pelo usuário depois da
v1.41.1: "não mudou nada, as cores ficaram diferente, mas agora
aparece texto cinza com fundo cinza" — a v1.41.1 corrigiu só a linha
SELECIONADA; todas as outras linhas continuavam com o mesmo problema.

### Corrigido

- **Causa raiz real**: `tview.NewTableCell` grava `cell.Style` com um
  valor concreto já no construtor (branco sobre preto, o padrão
  global do `tview`), e `TableCell.SetTextColor` só altera o
  componente de **primeiro plano** desse estilo (`c.Style =
  c.Style.Foreground(cor)`, só quando `cell.Style != tcell.
  StyleDefault` — o que é sempre o caso, vindo do construtor). O
  fundo fica exatamente como o construtor deixou — preto, fixo do
  `tview` — a menos que `SetBackgroundColor` *também* seja chamado.
  A correção da v1.41.1 só chamava `SetTextColor` por célula: o texto
  pegava a cor certa do tema, mas ficava sentado numa caixa preta que
  o tema nunca escolheu - "cinza sobre cinza" quando o tema ativo tem
  um texto em tom acinzentado (confirmado nesta máquina: o tema
  realmente ativo, "Solarized Light", tem `Foreground = #586E75`).
  Corrigido construindo o estilo inteiro (texto E fundo) com
  `SetStyle` de uma vez, em vez de `SetTextColor`/`SetBackgroundColor`
  separados.

### Verificado

- Novo `internal/filemanager/panel_test.go`
  (`TestRenderUsesTheConfiguredColorsNotTviewDefaults`) - lê o campo
  `Style` da própria `TableCell` direto depois de `render()`, não uma
  tela simulada renderizada; foi exatamente essa inspeção direta que
  achou e confirmou a causa raiz real (uma tela simulada não distingue
  "o padrão do `tview` coincidiu com o que eu queria" de "minha cor
  foi realmente aplicada").
- `internal/tui/filemanager_colors_test.go` reforçado pra também
  checar o fundo (não só o texto) de uma linha NÃO selecionada,
  contra o tema realmente ativo e persistido nesta máquina (não um
  tema fixo assumido).
- `go build`/`go vet`/`gofmt`/`go test ./...` limpos no repositório
  inteiro.

## [1.41.1] — 2026-11-26 — bloco **NEW**, feature **file-manager** (build 1)

**Correção de bug + visual**, reportada diretamente pelo usuário: o
gerenciador de arquivos tinha um esquema de cores fixo (amarelo sobre
preto, "parecendo um programa de 1979"), diferente do resto do
programa, e pior - **alguns diretórios apareciam com letra preta
sobre fundo preto**, invisíveis (navegáveis, mas não visíveis).

### Corrigido

- **Causa raiz do texto invisível**: nenhum `table.SetSelectedStyle`
  era configurado nos painéis (`tview.Table`); sem isso,
  `tview.Table` cai no fallback de "inverter" a cor da linha
  selecionada combinando o texto da célula com o `BackgroundColor`
  dela (nunca definido explicitamente) - combinação sem garantia
  nenhuma de contraste. Corrigido chamando `SetSelectedStyle`
  explicitamente com um par de cores garantido (fundo = cor de
  seleção do tema, texto = branco fixo — mesma escolha já comprovada
  nos dropdowns de menu da própria TUI).
- **Visual**: o gerenciador de arquivos agora usa o tema ativo da TUI
  (`internal/theme`) em vez de um esquema fixo próprio - painéis,
  diálogos e a barra de atalhos acompanham `Setup > Install` em tempo
  real (novo `FileManager.SetColors`, chamado pelo mesmo botão
  "Apply" que já reaplica o tema no resto da TUI). Novo
  `internal/filemanager.Colors` (com `DefaultColors()` pra testes sem
  tema) mapeado em `internal/tui/tui.go`'s `filemanagerColorsFor`.

### Verificado

- Novo teste de regressão `internal/tui/filemanager_colors_test.go`
  (`TestFileManagerEntriesAreNotBlackOnBlack`), mesma técnica de
  varredura de tela inteira já usada em
  `TestSetupMenuUnselectedItemsAreNotBlackOnBlack` - confirmado
  capaz de pegar a classe de bug (testado removendo o
  `SetSelectedStyle` de propósito e revertendo).
- `go build`/`go vet`/`gofmt`/`go test ./...` limpos no repositório
  inteiro.

## [1.41.0] — 2026-11-26 — bloco **NEW**, feature **file-manager** (build 0)

**Gerenciador de arquivos de dois painéis na TUI** — pedido do
usuário: "já que não temos os gerenciadores de arquivo ainda, podemos
seguir por este caminho e assim chamar o compilador do gerenciador,
criar projetos, etc.". VSIX/Sublime (Bloco 8) ficam deliberadamente
pro final — o WibEdit já embutido cobre edição.

### Adicionado

- `internal/filemanager` (novo pacote): componente de tela cheia estilo
  Norton Commander, dois painéis (`tview.Table`, colunas nome/tamanho/
  data), navegação local (`Tab` troca painel, `Enter`/`Backspace`
  entra/sai de diretório).
- Atalhos: `F2` compila (`dialect.CompileByExtension`), `F9` executa
  no emulador padrão da plataforma (`dialect.RunInEmulatorByExtension`),
  `F4` abre o arquivo selecionado no WibEdit já embutido (novo
  `editor.Editor.OpenFile`), `F5` copia e `F6` move pro painel oposto
  (`F6`/mover funciona em diretórios também, via `os.Rename`; `F5`/
  copiar só em arquivos nesta rodada — copiar diretório pediria
  recursão, fora de escopo), `F7` cria diretório, `F8` apaga (arquivo
  ou diretório vazio — `os.Remove` já rejeita diretório não-vazio,
  sem precisar de checagem extra), `Ctrl+N` cria um projeto novo,
  `Ctrl+P` lista projetos salvos e pula o painel ativo pra pasta
  escolhida.
- `internal/config/projects.go`: tabela `projects` (nome, pasta,
  dialeto, data de criação) no mesmo `*Store` SQLite de temas/opções.
  `CreateProject`/`ListProjects`. Projeto = nome + pasta (criada se
  não existir) + dialeto + checkbox opcional pra já gerar um arquivo-
  fonte inicial vazio (`.dmx`/`.dzx`/`.dzs`) — confirmado com o
  usuário; configurações por projeto (linha inicial, passo, emulador
  padrão etc.) ficam fora desta rodada, assim como editar/apagar um
  projeto já salvo.
- Novo item de menu "Files" na TUI (`internal/tui/tui.go`), mesmo
  padrão de 3 pontos já usado pro Editor/WibEdit (`AddPage`+
  `SetInputCapture`).
- `internal/chrome` (novo pacote): os ícones/sombra/moldura de janela
  de `internal/tui/chrome.go` (`attachShadow`/`windowChrome`) extraídos
  pra um pacote compartilhado, exportado — `internal/filemanager`
  precisa do mesmo visual retro, e não podia importar `internal/tui`
  sem criar um ciclo (`internal/tui` importa `internal/filemanager`
  pra embutir a página, igual já faz com o editor). `internal/tui/
  chrome.go` manteve só o que é específico dele (o `desktop`, o fundo
  tramado da área de trabalho).

### Decidido com o usuário

- TUI primeiro, GUI depois — mesma decisão já tomada pro editor.
- Montagem de `.DSK` num painel e operações local↔`.DSK` ficam pra
  depois — o `compile`/`execute` do MSX já não precisam disso (o
  `BuildTestDisk` interno cuida do disco de teste).

### Verificado

- `internal/config/projects_test.go`, `internal/filemanager/
  filemanager_test.go` (navegação, troca de painel, compilar/executar
  com `PATH` vazio, copiar/mover/criar diretório/criar projeto), mais
  um teste de menu na TUI com `tcell.SimulationScreen`
  (`TestFilesMenuOpensAndClosesTheFileManager`) confirmando que Files
  > Open abre o gerenciador e Escape fecha de volta pro desktop.
- `go build`/`go vet`/`gofmt`/`go test ./...` limpos no repositório
  inteiro.

## [1.40.0] — 2026-11-26 — bloco **NEW**, feature **dialect-package** (build 0)

**Refatoração interna, sem mudança de comportamento pro usuário** —
extrai o dispatch de dialeto (`compileByExtension`/
`runInEmulatorByExtension`/`emulatorBackends`/`resolveBackend`), até
aqui privado ao pacote `internal/cli`, pro novo pacote exportado
`internal/dialect` (`CompileByExtension`/`RunInEmulatorByExtension`/
`EmulatorBackends`/`ResolveBackend`). Preparação direta para o
gerenciador de arquivos de dois painéis (próxima release): a TUI
precisa chamar esse mesmo dispatch sem passar a depender de
`internal/cli`, que também carrega parsing de flags de linha de
comando — isso sim específico da CLI, sem nada a ver com a TUI.

### Alterado

- `internal/cli/dialect.go`/`dialect_test.go` → `internal/dialect/
  dialect.go`/`dialect_test.go`, com tudo exportado.
- `internal/cli/cli.go`/`console.go` passam a chamar
  `dialect.CompileByExtension`/`dialect.RunInEmulatorByExtension`.

### Verificado

- Mesma cobertura de testes, migrada sem alteração de comportamento.
- `go build`/`go vet`/`gofmt`/`go test ./...` limpos no repositório
  inteiro.

## [1.39.0] — 2026-11-26 — bloco **NEW**, feature **emulator-backends** (build 0)

**Seleção de emulador alternativo** — pedido do usuário depois de
perceber que MSX, ZX81 e ZX Spectrum agora têm mais de um emulador
instalado nesta máquina. Nova flag `--emulator <nome>` no CLI,
`execute <arquivo> [emulador]` no console.

### Adicionado

- `internal/cli/dialect.go`: `emulatorBackends` (lista de nomes
  válidos por extensão, primeiro = padrão) + `resolveBackend`.
- MSX: `msx.RunInFMSX`/`msx.RunInBlueMSX`, alternativas ao `openmsx`
  (padrão).
- ZX81: `zx81.RunInZEsarUX`, alternativa ao `eightyone` (padrão).
- ZX Spectrum: `zxspectrum.RunInZEsarUX`, alternativa ao `eightyone`
  (padrão) — bem mais simples de integrar que o EightyOne: `--machine
  48k --tape <arquivo>`, sem nenhum perfil INI.
- ZEsarUX deliberadamente fora do MSX (só emula MSX1 — openMSX/fMSX/
  blueMSX+ cobrem MSX2/MSX2+ também).

### Corrigido

- `--emulator` ANTES de `--execute`/do nome do arquivo (exatamente a
  forma documentada no `--help`) fazia `executeFileArg` confundir o
  nome do emulador com o arquivo a executar, já que sempre tomava
  literalmente o próximo token depois de `--execute`. Corrigido com
  `stripEmulatorFlag`, que remove a flag `--emulator`+valor de `args`
  antes de qualquer outra detecção de argumento posicional rodar.

### Verificado

- Os 4 backends novos (fMSX, blueMSX+, ZEsarUX×2) confirmados abrindo
  de verdade nesta máquina, incluindo através da flag `--emulator`
  (ambas as ordens: antes e depois do nome do arquivo) e do comando
  `execute` do console.
- `go build`/`go vet`/`gofmt`/`go test ./...` limpos no repositório
  inteiro.

## [1.38.0] — 2026-11-26 — bloco **NEW**, feature **zxspectrum-execute** (build 0)

**Comando `compile`/`execute` pro ZX Spectrum** (Bloco 4, Etapa 5 de 5
— roteiro completo) — os comandos já existentes (MSX/ZX81) agora
também escolhem ZX Spectrum pela extensão do arquivo: `goBaDig
programa.dzs`/`goBaDig --execute programa.dzs` funcionam de verdade.
**Ciclo ZX Spectrum completo via CLI, igual MSX e ZX81.**

### Adicionado

- `internal/cli/dialect.go`: `.dzs`→ZX Spectrum nos dois dispatchers
  (`compileByExtension`/`runInEmulatorByExtension`).
- `internal/platforms/zxspectrum/compile.go` (`CompileFile`,
  `DefaultCompileSettings`) e `run.go` (`RunInEmulator`) —
  `RunInEmulator` precisa de um perfil INI do EightyOne (dois
  argumentos posicionais, `[ini, fita]`, diferente de MSX/ZX81) —
  embutido como string Go, uma cópia temporária por chamada.
- Testes automatizados novos (`compile_test.go`/`run_test.go` no
  pacote `zxspectrum`, mais os pares `.dzs` dos testes já existentes
  em `internal/cli`).

### Decidido com o usuário

- O `emulator_interface.py` original do Spectrum espera o EightyOne
  fechar antes de retornar; MSX/ZX81 no Go não esperam. Escolhida a
  **consistência entre as três plataformas no Go** (não esperar), não
  a fidelidade ao Python nesse ponto específico.

### Validado

- Um processo `EightyOne.exe` real, lançado de verdade por `goBaDig
  --execute`, confirmado rodando nesta máquina.

### O que não mudou / ainda falta

- `REMX` e o modelo 128K na integração de emulador continuam fora de
  escopo (documentado no `OUTLINE.md`, Bloco 4).
- TUI/GUI ainda não chamam nenhum desses comandos.

## [1.37.0] — 2026-11-26 — bloco **NEW**, feature **zxspectrum-validation** (build 0)

**Validação de ponta a ponta ZX Spectrum** (Bloco 4, Etapa 4 de 5) —
os 9 exemplos reais já existentes no projeto
(`badig/examples/zxspectrum/*.DZS`) rodados pelo pipeline inteiro
(Lex→Parse→Generate→Tokenize→MakeTap) batem byte a byte, tanto no
`.AZS` quanto no `.tap`, contra a saída real do `badig.py`.

### Corrigido

- Maiúscula incondicional de palavra reservada e a normalização
  `GOTO`/`GOSUB`→`GO TO`/`GO SUB` (Pass5) checavam `tok.Kind` em vez
  do valor do token — os tokens sintéticos que o `pass2Functions`/
  `pass2Labels` compartilhados do core emitem (`gosub`/`return`/
  `goto`, para `FUNC`/`RET` e labels de loop) nunca carregam o Kind
  deste pacote, então ficavam sem forçar.
- `LET` implícito e a validação de atribuição composta checavam
  `prev.Kind == KindSeparator`, mas o `":"` sintético que
  `pass2Functions` insere entre os passos de uma chamada é
  `core.KindSynthetic` — ambos corrigidos pra checar por VALOR
  (mesmo princípio do `tk.uval in c_reserved_kw` do Python original,
  nunca por "tipo de token").

### Adicionado

- 3 testes de exemplo real comitados (`TestEndToEndRealHelloExample`/
  `...FunctionsExample`/`...FnExample`).

### Verificado

- Os 9 exemplos reais (`hello`/`data`/`fn`/`colours`/`functions`/
  `native`/`streams`/`graphics`/`play128`) batem byte a byte contra o
  `badig.py` real, texto e `.tap` tokenizado.
- `go build`/`go vet`/`gofmt`/`go test ./...` limpos no repositório
  inteiro.

### O que não mudou / ainda falta

- `REMX` continua fora de escopo.
- Nenhum comando `compile`/`execute` pro Spectrum ainda — última
  Etapa do roteiro, ver `OUTLINE.md`, Bloco 4.

## [1.36.0] — 2026-11-26 — bloco **NEW**, feature **zxspectrum-parser** (build 0)

**Parser ZX Spectrum** (Bloco 4, Etapa 3 de 5) — fecha o motor
Lex→Parse→Generate: o primeiro arquivo `.dzs` real já compila de
ponta a ponta. Transcrito de `badig_zxspectrum.py`'s
`Parser(MSXParser)`.

### Adicionado

- `zxspectrum.Parser`/`zxspectrum.NewParser` (`core.Parser`): `LET`
  implícito (linha nova, `:` ou `THEN`), `PRINT AT` (macro simples,
  sem `Expand`), atribuição composta (`+=`/`-=`/`*=`/`/=`/`^=`/`++`/
  `--`, RHS completo entre parênteses), `TRUE`/`FALSE`→`1`/`0`,
  normalização `GOTO`/`GOSUB`→`GO TO`/`GO SUB`, maiúscula
  incondicional de palavra reservada, limite de linha 9999.
- 23 testes automatizados novos.

### Verificado

- Validação extensiva rodando o `badig.py` real (não só os testes
  automatizados): todos os casos de atribuição composta, `PRINT AT`
  com 1 argumento, `GOTO`/`GOSUB`, maiúscula de palavra reservada,
  colisão com palavra reservada da ROM, e as 4 mensagens de erro de
  atribuição composta batem exato.
- `go build`/`go vet`/`gofmt`/`go test ./...` limpos no repositório
  inteiro.

### O que não mudou / ainda falta

- `REMX` continua fora de escopo (mesma decisão do ZX81).
- Nenhum comando `compile`/`execute` pro Spectrum ainda — próxima
  Etapa.
- Roteiro completo das próximas etapas documentado em `OUTLINE.md`,
  Bloco 4.

## [1.35.0] — 2026-11-26 — bloco **NEW**, feature **zxspectrum-lexer** (build 0)

**Description/Lexer ZX Spectrum** (Bloco 4, Etapa 2 de 5) — vocabulário
Dignified (`instructions`/`jumps`/`operators`/`symbols`) e o `Lexer`
transcritos de `badig_zxspectrum.py`'s `Description`/`Lexer(MSXLexer)`.
Ainda sem `Parser` (nenhum `.dzs` compila de ponta a ponta ainda).

### Adicionado

- `zxspectrum.Description` (`core.Description`).
- `zxspectrum.Lexer` (`core.Lexer`).
- 15 testes automatizados novos.

### Verificado

- Confirmado empiricamente (rodando o `badig.py` real) antes de
  travar cada ponto ambíguo: significado do apóstrofo (`'`, separador
  de `PRINT`, não comentário), comportamento de `GO TO`/`GO SUB` com
  espaço variável, e o espaçamento duplo real de `REM` + conteúdo.
- `go build`/`go vet`/`gofmt`/`go test ./...` limpos no repositório
  inteiro.

### O que não mudou / ainda falta

- Nenhum `Parser` Spectrum ainda — nenhum `.dzs` compila de ponta a
  ponta.
- Nenhum comando `compile`/`execute` pro Spectrum ainda.

## [1.34.0] — 2026-11-26 — bloco **NEW**, feature **zxspectrum-tokenizer** (build 0)

**Tokenizer ZX Spectrum** (Bloco 4, Etapa 1 de 5) — primeira peça do
módulo ZX Spectrum: `internal/platforms/zxspectrum` converte texto
Classic ASCII já numerado em bytes tokenizados, igual ao que o
tokenizer MSX/ZX81 fizeram antes de qualquer Description/Lexer/Parser.
Ainda sem Description/Lexer/Parser Spectrum (nenhum `.dzs` Dignified
ainda compila) — ver Etapas 2-5 no roteiro, `OUTLINE.md`, Bloco 4.

### Adicionado

- `zxspectrum.Tokenize(source, model string) ([]byte, error)`
  (`model` é `"48"` ou `"128"`).
- `zxspectrum.EncodeNumber`/`EncodeText`, `FnDefinition`,
  `ValidateFnCall`, `ValidateIdentifier` (exportadas; usadas também
  pelo futuro Parser, Etapa 3).
- `zxspectrum.MakeTap`/`MakeTzx` (`tape.go`) — empacotamento de fita,
  100% mecânico.
- 19 testes automatizados, portados dos vetores de ouro reais do
  próprio `test_zxspectrum.py` do projeto Python original
  (`TokenizerTests`/`TapeTests`).

### Verificado

- Todos os 19 testes Go passam, confirmados byte a byte contra os
  vetores reais do `test_zxspectrum.py` (ponto flutuante Sinclair,
  slots ocultos de `DEF FN`, escapes gráficos/UDG, `.tap`/`.tzx`).
- Ponto flutuante Sinclair codificado com `math/big` (aritmética
  racional exata), não `float64` — diferente do ZX81, que usa
  `float64`/IEEE754 porque o Python original também usa `float64` ali;
  o Spectrum real usa `decimal.Decimal`+`fractions.Fraction`.

### O que não mudou / ainda falta

- Nenhum Description/Lexer/Parser Spectrum ainda — nenhum arquivo
  `.dzs` Dignified compila de verdade.
- `REMX` deliberadamente fora de escopo nesta Etapa (mesma decisão já
  tomada pro ZX81, Bloco 3) — vive só no nível do Parser, não do
  tokenizer, pro Spectrum.
- Nenhum comando `compile`/`execute` pro Spectrum ainda.
- Roteiro completo das próximas etapas documentado em `OUTLINE.md`,
  Bloco 4.

## [1.33.0] — 2026-11-24 — bloco **NEW**, feature **zx81-execute** (build 0)

**Comando `compile`/`execute` pro ZX81** (Bloco 3, Etapa 5 de 5 —
roteiro completo) — os comandos já existentes (até aqui hard-coded
pro MSX) agora escolhem o dialeto pela extensão do arquivo:
`goBaDig programa.dzx`/`goBaDig --execute programa.dzx` funcionam
igual aos seus equivalentes `.dmx`. **Ciclo ZX81 completo via CLI,
igual ao MSX.**

### Adicionado

- `internal/cli/dialect.go`: `compileByExtension`/
  `runInEmulatorByExtension`, despachando `.dmx`→MSX, `.dzx`→ZX81.
- `msx.CompileFile`/`msx.RunInEmulator` (e os equivalentes do console)
  deixam de ser chamados direto por `cli.go`/`console.go` — só pelo
  dispatch acima.
- `internal/platforms/zx81/compile.go` (`CompileFile`,
  `DefaultCompileSettings`) e `run.go` (`RunInEmulator`) — sem nenhum
  passo de "empacotar num disco" (diferente do MSX): o `.p` já é tudo
  que o EightyOne precisa.
- 11 testes automatizados novos.

### Validado

- Um processo `EightyOne.exe` real, lançado de verdade por
  `goBaDig --execute`, confirmado rodando nesta máquina.

### O que não mudou / ainda falta

- `remx` e a renumeração de listagem numerada continuam fora de
  escopo (documentado desde a Etapa 3).
- TUI/GUI ainda não chamam nenhum desses comandos.

## [1.32.0] — 2026-11-24 — bloco **NEW**, feature **zx81-validation** (build 0)

**Validação de ponta a ponta ZX81** (Bloco 3, Etapa 4 de 5) — os 2
exemplos reais já existentes no projeto (`badig/examples/zx81/
zx81-intro.DZX` e `mcode-ok.DZX`) rodados pelo pipeline inteiro
(`Lex→Parse→Generate`) batem **byte a byte** contra a saída real do
`badig.py`.

### Adicionado

- 2 testes de ponta a ponta comitados (conteúdo embutido, sem
  depender de arquivo em disco) + 2 vetores de ouro adicionais do
  `test_tokenizer_zx81.py` que dependiam do `Parser` (agora existe):
  as duas formas de escape de aspas sobrevivem intactas até o
  `Generate()`; `DECLARE` com alias explícito interage corretamente
  com a alocação automática de outra variável.

## [1.31.0] — 2026-11-24 — bloco **NEW**, feature **zx81-parser** (build 0)

**Parser ZX81** (Bloco 3, Etapa 3 de 5) — fecha o Dignified→Classic
pro ZX81: o **primeiro arquivo `.dzx` real já compila de ponta a
ponta** (Lex→Parse→Generate). Falta só ligar `compile`/`execute` (já
tem o tokenizer pronto) e validar no emulador.

### Adicionado

- `internal/platforms/zx81/parser.go`: alocação de variável de 1
  letra (reusa `core.NewDescendingAllocator()` sem mudança), `LET`
  implícito, `PRINT AT` (`[?](row,col)` → `print at row,col;`).
- `core.BuiltinDefine` ganha um campo novo, `Expand` (opcional) —
  deixa um `DEFINE` embutido substituir sua expansão inteira em vez
  de só preencher um placeholder, usado pelo `PRINT AT` do ZX81 pra
  separar e defaultar duas coordenadas independentemente.
- 12 testes automatizados de ponta a ponta.

### Achados

- `~` (Dignified, "mantenha o nome longo") é sempre erro fatal pro
  ZX81 — BASIC clássico do ZX81 não consegue representar um nome de
  variável com mais de 1 caractere, diferente do MSX.
- Toda a validação (alocação, `PRINT AT`, `LET` implícito, os erros
  de `~`) foi confirmada rodando o `badig.py` real primeiro, teste Go
  depois.

### Decisão de escopo

- `remx` e a renumeração de listagem numerada ficam de fora desta
  etapa — são hooks que envolvem o pipeline inteiro no Python
  (`preprocess`/`postprocess`), não passes do `Parser`, e exigiriam
  uma evolução de arquitetura maior. Ver `OUTLINE.md`, Bloco 3.

## [1.30.0] — 2026-11-24 — bloco **NEW**, feature **zx81-lexer** (build 0)

**Description + Lexer ZX81** (Bloco 3, Etapa 2 de 5) —
`internal/platforms/zx81` ganha o vocabulário clássico do dialeto
(31 instruções, funções, operadores, símbolos) e o scanner que
classifica texto ZX81 Dignified real. Ainda sem Parser — nenhum
`.dzx` compila de ponta a ponta ainda.

### Adicionado

- `description.go`: `core.Description` completo pro ZX81 (1 letra de
  variável, `CommentChar()="REM"`, etc.).
- `lexer.go`: `core.Lexer` completo — palavras-chave, operadores,
  símbolos (incluindo os compostos `<>`/`>=`/`<=`/`**`), escape hex
  `\{XX}` com Kind próprio, string com `\"` e `""` coexistindo,
  `REM`, `?` como atalho de `PRINT`.
- 11 testes automatizados, incluindo 9 casos documentando um
  comportamento real bem peculiar de número com sinal (confirmado
  rodando o `badig.py` real com `-lex -prr`): um sinal só funde com
  um número quando imediatamente seguido de um dígito puro — nunca
  atravessa um ponto decimal nem um expoente.

### Achado

- O ZX81 real não tem o atalho `''` de bloco de comentário clássico
  que o MSX tem — confirmado lendo o código-fonte das duas
  plataformas; o mecanismo genérico do `core` já lida com isso sem
  nenhum código extra.

## [1.29.0] — 2026-11-24 — bloco **NEW**, feature **zx81-tokenizer** (build 0)

**Tokenizer ZX81** (Bloco 3, Etapa 1 de 5) — primeira peça do módulo
ZX81: `internal/platforms/zx81` converte texto Classic ASCII ZX81 já
numerado numa imagem `.p` completa (system vars + programa + display
file + variáveis), incluindo o "ponto flutuante oculto" que o ZX81
esconde depois de cada número tokenizado. Ainda sem
Description/Lexer/Parser ZX81 (nenhum `.dzx` Dignified ainda
compila) — ver Etapas 2-5 no roteiro, `OUTLINE.md`, Bloco 3.

### Adicionado

- `zx81.Tokenize(lines []string) ([]byte, error)`.
- `remx <linha>,<count>[,<byte>]` (diretiva de reserva de bytes crus
  dentro de uma linha REM).
- `docs/golden-vectors/zx81-tokenizer.md`.
- 27 testes automatizados, portados dos vetores de ouro reais do
  próprio `test_tokenizer_zx81.py` do projeto Python original.

### Verificado

- Os 27 testes do `test_tokenizer_zx81.py` real (`TokenizerTests`)
  continuam passando nesta máquina.
- Validação cruzada ad-hoc: `tokenize_lines` (Python) e
  `zx81.Tokenize` (Go) produzem o mesmo buffer byte a byte pra uma
  entrada nova, fora dos vetores já fixados.

## [1.28.0] — 2026-11-24 — bloco **NEW**, feature **execute-command** (build 0)

**Comando `execute`** — fecha o ciclo MSX completo via CLI: editar →
compilar → testar num emulador real. `goBaDig --execute
<arquivo.dmx>` (ou `execute <arquivo.dmx>` no console `--cli`)
compila, gera um disco `.dsk` de teste (o programa como
`AUTOEXEC.BAS`, pra rodar sozinho sem digitar nada) e abre no
openMSX.

### Adicionado

- `msx.BuildTestDisk(dmxPath)`: compila e empacota o resultado num
  disco `.dsk` novo.
- `msx.RunInEmulator(diskPath)`: abre o openMSX (precisa estar no
  `PATH`) com o disco montado.
- `goBaDig --execute <arquivo.dmx>` e `execute <arquivo.dmx>` no
  console.
- Validado com um processo `openmsx.exe` real, lançado de verdade
  pelos dois pontos de entrada.
- 12 testes novos.

### O que não mudou / ainda falta

Nenhum protocolo de monitoramento de erro em tempo real (mapear erro
pra linha Dignified); nenhuma opção de máquina/extensão do emulador
configurável ainda (usa os defaults do openMSX); os arquivos reais de
sistema MSX-DOS (`MSXDOS.SYS`/`COMMAND.COM`) não são embutidos no
disco de propósito (são software de terceiros sob copyright) — o
disco gerado mostra um breve "Boot error" antes de cair pro
`AUTOEXEC.BAS`, não é um bug; TUI/GUI ainda não chamam esse comando.

## [1.27.0] — 2026-11-24 — bloco **NEW**, feature **compile-command** (build 0)

**Comando `compilar` de verdade** — a primeira conversão
Dignified→binário disponível pro usuário final (não só testada
internamente): `goBaDig <arquivo.dmx>` no CLI, `compile
<arquivo.dmx>` no console `--cli`.

### Adicionado

- `msx.CompileFile(dmxPath, settings)`: converte um `.dmx` de
  verdade, escrevendo `<base>.amx` (listagem Classic, CRLF — mesma
  convenção do `badig.py` real no Windows) e `<base>.bmx` (binário
  tokenizado) ao lado do arquivo de entrada.
- `msx.DefaultCompileSettings()`: os defaults reais do `badig.py`
  (`LineStart`/`LineStep=10`, `RemHeader=true`, `TabLength=4`).
- `goBaDig <arquivo.dmx>` (CLI, sem nenhuma opção) e `compile
  <arquivo.dmx>` (console) chamam os dois.
- Validado byte a byte, nos dois arquivos de saída, contra o
  `CGK173.dmx` inteiro (891 linhas) convertido pelo `badig.py` real.
- 9 testes novos.

### O que não mudou / ainda falta

Nenhuma flag ainda lê as opções persistidas pelos painéis de
configuração (usa sempre os defaults); nenhum caminho de saída
customizável; TUI/GUI ainda não chamam esse comando; a visão maior
(copiar pra pasta, gerar disco `.DSK`, testar num emulador) fica
pra uma rodada futura.

## [1.26.0] — 2026-11-23 — bloco **NEW**, feature **dsk-library** (build 0)

**Biblioteca `.DSK` (MSX-DOS/FAT12) portada** — Etapa 1 do
gerenciador de arquivos de dois painéis (estilo Norton Commander/PC
Tools) pedido pelo usuário, cujo uso principal é criar/editar
imagens de disquete MSX.

### Adicionado

- Pacote novo `internal/platforms/msx/dsk`: `Create`/`Open`/`Close`/
  `ListFiles`/`AddFile`/`ExtractFile`/`DeleteFile`, mais
  `ConvertToFAT11`/`MatchesFAT11` (nome 8.3 + wildcard). Porta direta
  de `E:\msxdiskutil\MSXDisk.pbi` — a reescrita em PureBasic, do
  próprio usuário, do utilitário original em C (`rddsk`/`wrdsk`,
  Arnold Metselaar).
- Validado byte a byte contra um disco `.DSK` real (extração
  idêntica aos arquivos de referência soltos).
- 14 testes novos.

### O que não mudou / ainda falta

Nenhuma navegação/UI de dois painéis ainda (próximas etapas); nenhum
valor gerado aqui é lido pelo fluxo de conversão/empacotamento real
ainda.

## [1.25.0] — 2026-11-23 — bloco **NEW**, feature **tui-menu-colors** (build 0)

**Correção de cor no menu suspenso da TUI** — reportada pelo usuário
assim que o menu `Setup` passou a ter itens de verdade de sobra (4
painéis novos da Etapa C) pra revelar o problema.

### Corrigido

- Item não selecionado de um menu suspenso aberto renderizava com
  texto preto em fundo preto, invisível em todos os temas.
- O destaque do item selecionado parava no fim do próprio texto, em
  vez de preencher a largura inteira do menu.
- 2 testes novos de ponta a ponta (tela de terminal simulada),
  confirmados contra o bug real antes da correção.

## [1.24.0] — 2026-11-23 — bloco **NEW**, feature **gui-options-panels** (build 0)

**4 painéis de configuração na GUI** — Etapa D, a última do painel de
configuração: agora as três interfaces (CLI, TUI, GUI) têm o mesmo
conjunto de opções editável.

### Adicionado

- Menu **Setup** da GUI ganha os mesmos 4 itens da TUI: **Basic
  Dignified**, **MSX**, **ZX81**, **ZX Spectrum** — cada um abre um
  formulário (`dialog.ShowForm` do Fyne) com todas as opções daquela
  categoria, valores atuais pré-preenchidos, Apply/Cancel.
- Validação em tempo real para opções numéricas (o botão Apply fica
  desabilitado enquanto o valor estiver fora do intervalo permitido).
- 6 testes novos.

### O que não mudou / ainda falta

Nenhum valor configurado neste painel (em qualquer das três
interfaces) ainda é lido por um comando de conversão real — isso é
consumo futuro pelo pipeline, um passo separado.

## [1.23.0] — 2026-11-23 — bloco **NEW**, feature **tui-options-panels** (build 0)

**4 painéis de configuração na TUI** — Etapa C do painel de
configuração, sobre o catálogo (v1.21.0) e os comandos `SET`/
`SETTINGS`/`DIALECT` do CLI (v1.22.0).

### Adicionado

- Menu **Setup** ganha 4 itens novos: **Basic Dignified**, **MSX**,
  **ZX81**, **ZX Spectrum** — cada um abre um formulário com todas as
  opções daquela categoria (checkbox pra booleano, dropdown pra enum,
  campo numérico pra inteiro, campo de texto livre pro resto),
  valores atuais pré-preenchidos, botões Apply/Cancel.
- 1 teste novo de ponta a ponta (tela de terminal simulada).

### O que não mudou / ainda falta

- O painel equivalente na GUI ainda não existe (Etapa D).
- Nenhum valor configurado aqui é lido pelo pipeline de conversão
  real ainda.

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
