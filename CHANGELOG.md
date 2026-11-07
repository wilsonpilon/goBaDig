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

> Esta é a interpretação registrada nesta sessão de planejamento — ver
> `OUTLINE.md`, seção "Perguntas em aberto", para confirmar/ajustar o
> comportamento exato de reset de `Y`/`Z` e o formato do nome anexado à
> tag antes do primeiro release real.

## Empacotamento de release

Cada release publicado gera um único pacote `goBaDig-vX.Y.Z.zip`
(nome do projeto + versão), contendo o executável, documentação mínima e
exemplos — ver [`dist/README.md`](dist/README.md) para a estrutura
exata do pacote.

---

## [Não lançado]

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
- Estado de trabalho: a TUI atual congela ao abrir `Setup > Install`;
  o redesign da interface e do diálogo fica pendente para a próxima
  sessão. Esta versão ainda não foi publicada.
- Primeira fatia do armazenamento SQLite com driver puro Go
  `modernc.org/sqlite`, tabelas `themes` e `active_themes` por modo.
- `build.ps1` (raiz do repo): script PowerShell que automatiza o build
  local completo — baixa dependências (`go mod download`), roda
  `go vet`/`go test`, compila e monta `dist/goBaDig-vX.Y.Z/` + o zip,
  lendo a versão direto de `internal/version/version.go`. Não cria tag
  Git nem release no GitHub (continua manual). Testado de ponta a
  ponta, inclusive com `-SkipDownload -SkipTests`.

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
