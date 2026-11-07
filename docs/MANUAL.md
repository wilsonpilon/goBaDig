# Manual do goBaDig

> Este manual acompanha o projeto desde a fase de planejamento. Seções
> marcadas **(pendente)** ainda não se aplicam porque a funcionalidade
> correspondente não foi implementada — ver `OUTLINE.md` para o estado
> real de cada bloco de trabalho. Não assuma que uma seção preenchida
> aqui descreve código já existente sem conferir o `OUTLINE.md`.

## Sumário

1. [Visão geral](#1-visão-geral)
2. [Requisitos](#2-requisitos)
3. [Download](#3-download)
4. [Compilar a partir do código-fonte](#4-compilar-a-partir-do-código-fonte)
5. [Executar](#5-executar)
6. [Converter um programa Dignified](#6-converter-um-programa-dignified)
7. [Rodar em um emulador](#7-rodar-em-um-emulador)
8. [Testes](#8-testes)
9. [Integração com editores](#9-integração-com-editores)
10. [Solução de problemas](#10-solução-de-problemas)

---

## 1. Visão geral

**goBaDig** é um pré-processador de BASIC para as plataformas **MSX**,
**ZX81** e **ZX Spectrum**, reescrito em Go a partir da *Basic Dignified
Suite* original (Python, por Farique), com extensões para ZX81 e ZX
Spectrum já incorporadas. Converte código-fonte "Dignified" (sem números
de linha, com labels, macros, nomes longos, etc.) para BASIC clássico
numerado e, opcionalmente, para o binário tokenizado/empacotado em fita
de cada plataforma, pronto para rodar em um emulador.

Três formas de uso: **CLI** (linha de comando, scriptável), **TUI**
(terminal interativo, `tview`/`tcell`) e **GUI** (`Fyne`). Configuração,
palavras-chave e opções ficam em **SQLite**.

## 2. Requisitos

**(pendente de confirmação final — ver `OUTLINE.md` seção "Perguntas em
aberto")**

- Nenhuma dependência de Python/venv é necessária para o caminho
  principal (texto → BASIC numerado → binário/fita) — diferente do
  projeto original.
- Emuladores externos continuam sendo instalação manual do usuário, por
  plataforma (ver [`RESOURCES.md`](RESOURCES.md) para links, licenças e
  observações de cada um):
  - MSX: openMSX ou fMSX
  - ZX81: EightyOne (Windows), ZEsarUX ou sz81/XZ81 (Linux)
  - ZX Spectrum: EightyOne (Windows), Spectaculator (Windows) ou ZEsarUX
- Formatos de mídia/fita/disco suportados, por plataforma: MSX (`.BMX`,
  `.CAS`, `.DSK` — os dois últimos planejados); ZX81 (`.p`, `.81`,
  `.p81` — os dois últimos planejados); ZX Spectrum (`.TAP`, `.TZX`,
  `.Z80`, `.SNA`, `.DSK` +3 — os dois últimos planejados).
- Snapshots Z80/SNA do ZX Spectrum (opcional, fase tardia): dependência
  externa a definir (ver `OUTLINE.md`, Bloco 4).

## 3. Download

A partir da release `v1.1.2`, os pacotes `.zip` (um por release,
nomeados `goBaDig-vX.Y.Z.zip`) estão disponíveis nos *Releases* do
repositório no GitHub. Cada pacote contém o executável, documentação
mínima e exemplos — pronto para extrair e rodar sem instalar nada além
dos emuladores externos (ver seção 2), quando a integração de
emulador existir. Ver `CHANGELOG.md` para a lista de versões
publicadas. **Código em desenvolvimento (1.2.0, não publicado):** `--help`,
`--man`, o console `--cli` (com seleção de tema e conclusão por Tab) e a
GUI com temas. A TUI está em redesign; abrir `Setup > Install` congela a
interface na implementação atual. Não há conversor/tokenizer ainda.

## 4. Compilar a partir do código-fonte

```console
git clone <url-do-repositorio>
cd goBaDig
go build ./cmd/goBaDig
```

Requer Go 1.21+ e, para a GUI (Fyne), um compilador C (CGO habilitado —
no Windows, MinGW-w64 funciona). A TUI (`tview`/`tcell`) não precisa de CGO.

### Build automatizado (`build.ps1`)

No Windows, `build.ps1` (raiz do repositório) automatiza todo o processo
de release local: baixa as dependências (`go mod download`), roda
`go vet`/`go test`, compila e monta o pacote completo em
`dist/goBaDig-vX.Y.Z/` + o zip `dist/goBaDig-vX.Y.Z.zip`, com a versão
lida direto de `internal/version/version.go` (não precisa editar o
script a cada release).

```powershell
.\build.ps1                       # build completo (deps + testes + pacote)
.\build.ps1 -SkipTests            # pula go vet/go test
.\build.ps1 -SkipDownload         # pula go mod download (rebuild offline)
```

Ele **não** cria tag Git nem release no GitHub — isso continua sendo um
passo manual/deliberado (ver histórico em `docs/releases/`).

## 5. Executar

```console
goBaDig                        # banner curto (não entra no console)
goBaDig --help                 # lista de opções
goBaDig --man                  # referência estilo "man" (só o "about" nesta versão)
goBaDig --tui                  # abre a interface de terminal
goBaDig --gui                  # abre a interface gráfica (Fyne)
goBaDig --cli                  # abre o console interativo (prompt "goBaDig>")
```

No modo `--cli`, o prompt aceita `help`, `man`, `about`, `version`,
`theme [nome]` e `exit`/`quit` (ou `Ctrl-D` para sair). `theme` sem
argumento mostra o tema ativo e a lista; por exemplo, `theme dracula`
ou `theme Solarized Light` troca o tema persistido para o console. Em um
terminal interativo, `Tab` completa comandos e temas; pressione `Tab`
novamente para listar as alternativas (por exemplo, depois de `theme `).
A cor do prompt acompanha o tema selecionado. Comandos reais de conversão
chegam junto com o motor (Blocos 1-4 do `OUTLINE.md`).

Na TUI, há um **desktop clássico inspirado no Turbo Pascal 6, QuickBasic
e Norton Editor 2.0**: fundo azul, barra de título, menu suspenso,
espaço de trabalho e barra de status. `Setup > Install` abre uma janela
central com um campo dropdown para o tema, a descrição correspondente e
os botões `Apply` e `Cancel`. As setas navegam os menus e campos; Enter
abre/seleciona e Esc fecha/cancela. Os seis temas são independentes dos
temas da GUI e a seleção é aplicada na hora e persistida em SQLite.
Navegação: setas ou `Alt+letra` movem/abrem o menu, `Enter` confirma,
`Esc` fecha o menu suspenso. A GUI mostra uma janela com o mesmo menu
`File > Exit` / `Help > About` (nativo Fyne) e `Setup > Install`, onde
é possível escolher entre seis temas inspirados em VS Code Dark, Dracula,
Monokai, Nord, Solarized Light e Gruvbox Dark. A escolha da GUI também é
aplicada na hora e persistida em SQLite no diretório de configuração do
usuário. Nenhuma conversão de BASIC
acontece ainda em nenhuma das duas. O argumento `<arquivo> -id msx`
(conversão de fato) chega junto com os Blocos 1-4 do `OUTLINE.md`.

## 6. Converter um programa Dignified

**(pendente — depende dos Blocos 1 a 4 do `OUTLINE.md`)**

Exemplos de referência (fonte Dignified real, do projeto Python
original, úteis para entender a sintaxe enquanto o conversor Go não
existe) estão em `badig/examples/`. Os exemplos oficiais do projeto Go
ficarão em [`examples/`](../examples/README.md), um por plataforma.

## 7. Rodar em um emulador

**(pendente)**

Ver [`RESOURCES.md`](RESOURCES.md) para onde obter cada emulador e
`OUTLINE.md` (Blocos 2-4) para o estado da integração por plataforma.

## 8. Testes

**(pendente)**

O projeto Python original tem suítes de teste para ZX81
(`badig/zx81/test_tokenizer_zx81.py`) e ZX Spectrum
(`badig/zxspectrum/test_zxspectrum.py`) com vetores de bytes conhecidos —
esses testes servem de **golden-master** para validar a reescrita em Go
byte a byte. O módulo MSX não tem suíte de testes no original; vetores
próprios precisam ser gerados a partir do Python antes do port (ver
`OUTLINE.md`, Bloco 0).

## 9. Integração com editores

Suporte planejado, por ordem de prioridade (ver `OUTLINE.md`, Bloco 8):

1. **Visual Studio Code** — um único VSIX cobrindo os três dialetos
   (MSX, ZX81, ZX Spectrum), substituindo as três extensões separadas do
   projeto Python.
2. **Sublime Text** — pacotes equivalentes (syntax, build system,
   snippets), um por plataforma ou unificados, a definir.
3. **Editor próprio** (baseado no Norton Editor 2.0) — o usuário já tem
   implementações prontas em outros projetos (`msxide`: Norton Editor
   para TUI e um editor GUI com as mesmas teclas/funções do NE2;
   `paleobasic`: também tem o NE) — **fora de escopo de implementação
   por enquanto**, tratado como item futuro separado. Ver
   `docs/RESOURCES.md`.
4. **Notepad++ e UltraEdit** — avaliação de viabilidade (provavelmente
   via arquivos de realce de sintaxe/UDL próprios de cada editor),
   prioridade baixa, "se for possível".

### Sistema de ajuda e manuais de BASIC embutidos

- **CLI**: `--help` (opções) e `--man` (referência estilo `man` do
  Unix — nesta versão só o "about"; depois vai cobrir as opções do
  BaDig em detalhe **e os manuais de sintaxe BASIC do MSX, ZX81 e ZX
  Spectrum**, reaproveitando material já existente do usuário em
  outros projetos próprios, `fwMSX` e `go81` — ver `docs/RESOURCES.md`
  e `OUTLINE.md`, seção 10.7).
- **TUI e GUI**: ajuda tradicional via menu `Help > About` (implementado
  nesta versão); um visualizador do mesmo manual embutido vem depois.

**(pendente)** — nenhum pacote de editor Go-native existe ainda; os
pacotes Python (`badig/*/VSCode Extension/`, `badig/*/Sublime Package/`)
continuam a base de referência para a unificação.

## 10. Solução de problemas

**(pendente — sem build ainda para ter FAQ real)**

Problemas conhecidos do projeto Python original que o Go deve evitar por
desenho (não é algo que o usuário final deste manual precise contornar):
exit code sempre 0 em erro silencioso, launcher EightyOne travando ao
abrir Spectrum 128K automaticamente, monitoramento de erro do openMSX
não funcionando no Windows. Ver `OUTLINE.md` para detalhes e decisões.
