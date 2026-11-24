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
publicadas.

**`v1.2.0`** traz temas persistidos nas três interfaces (CLI/TUI/GUI),
a TUI redesenhada no visual retrô Turbo Vision/Norton Commander
(janelas com sombra, ícones Unicode) e o editor **WibEdit** integrado à
TUI (menu `Editor > Open`) — ver seções 5 e 9. **`v1.3.0`** começa o
motor de conversão propriamente dito (`internal/core`): ainda não há
nenhum comando de conversão de verdade disponível para o usuário final
— é só a base interna (tipos/interfaces) que os módulos MSX/ZX81/ZX
Spectrum vão usar (ver seção 6 e `OUTLINE.md`, Bloco 1). **`v1.4.0`**
porta o tokenizer MSX por completo (`internal/platforms/msx`); **`v1.5.0`**
porta `Description`/`Lexer` MSX (primeiro uso real do motor contra um
dialeto de verdade); **`v1.6.0`** porta `Parser` Pass 3/4 (alocação
de variáveis de 2 letras — primeira conversão Dignified→Classic de
verdade produzindo texto correto); **`v1.7.0`** porta `Parser` Pass 5
(`?`/`PRINT`, `THEN`/`GOTO`, atribuição composta, `TRUE`/`FALSE`,
Unicode); **`v1.8.0`** resolve labels (`{nome}`) e numera linhas —
primeiro programa com `GOTO`/label real compilando de ponta a ponta,
validado contra um exemplo real completo
(`badig/examples/msx/CGK173.dmx`); **`v1.9.0`** resolve loop labels
(`nome{ ... }` + `exit`); **`v1.10.0`** implementa `DEFINE` (macro de
texto com placeholder/valor padrão/chamadas aninhadas); **`v1.11.0`**
implementa `FUNC`/`RET` (função Dignified → `GOSUB` + troca de
variáveis); **`v1.12.0`** implementa `DECLARE` (reserva manual de
alias de variável); **`v1.13.0`** implementa `KEEP` (compilação
condicional via tags `#tag`); **`v1.14.0`** implementa `INCLUDE`
(mesclagem de outro arquivo Dignified); **`v1.15.0`** implementa um
`generate()` de verdade (espaçamento por palavra reservada + limite
de 256 caracteres por linha); **`v1.16.0`** implementa os
comentários exclusivos do Dignified (`##`, bloco de 3 cerquilhas,
bloco de duas aspas simples); **`v1.17.0`** implementa a fusão de
linha pelo separador clássico `:` (mantendo o próprio `:`, diferente
do `_` Dignified) + junção de strings adjacentes; **`v1.18.0`**
implementa o cabeçalho REM automático de 2 linhas e a expansão de
TAB pra espaços — fechando os dois últimos gaps conhecidos — **o
`CGK173.dmx` inteiro (891 linhas) agora bate byte a byte contra a
saída real do `badig.py` rodando em Python de verdade, sem nenhuma
diferença restante**; **`v1.19.0`** implementa o Pass 1 do separador
de chamada Dignified `_` fora do fim de linha (`_CLS` equivale a
`CALL CLS`; o nome logo depois do `_` fica fora da renomeação de
variável) — **com isso, todo o caminho textual Dignified→Classic
está fechado**; **`v1.20.0`** liga tudo isso ao tokenizer MSX Batoken
(`msx.CompileToBinary`) — **a primeira conversão Dignified→binário
`.bmx` de ponta a ponta de verdade, validada byte a byte contra o
`CGK173.dmx` inteiro tokenizado pelo `badig.py` real** — ainda sem
nenhum comando de conversão disponível pro usuário final (nenhuma das
três interfaces chama essa função ainda; ver seção 6). **`v1.21.0`**
começa o painel de configuração: `internal/options` cataloga as ~40
opções que o `badig.py` original expõe (numeração de linha, cabeçalho
REM, TAB, tokenizer, emulador — por dialeto) e `internal/config`
ganha o schema SQLite que as persiste; nas releases seguintes esse
painel ganhou comandos `SET`/`SETTINGS`/`DIALECT` no CLI (`v1.22.0`)
e os mesmos 4 painéis na TUI (`v1.23.0`) e na GUI (`v1.24.0`); a
`v1.25.0` corrigiu um bug de cor no menu suspenso da TUI; a `v1.26.0`
portou a biblioteca `.DSK` (MSX-DOS/FAT12,
`internal/platforms/msx/dsk`) usada pelo futuro gerenciador de
arquivos. **`v1.27.0`** finalmente liga tudo isso a um comando real
pro usuário: `goBaDig <arquivo.dmx>` no CLI (ou `compile
<arquivo.dmx>` no console `--cli`) converte um arquivo MSX Dignified
de verdade, gerando `.amx` e `.bmx`. **`v1.28.0`** fecha o ciclo:
`goBaDig --execute <arquivo.dmx>` (ou `execute <arquivo.dmx>` no
console) compila, empacota o resultado num disco `.dsk` de teste
(`AUTOEXEC.BAS`, roda sozinho) e abre no openMSX — editar→compilar→
testar num emulador real, tudo via CLI — ver seção 6 pro
detalhamento. TUI e GUI ainda não chamam nenhum desses comandos.
**`v1.29.0`** começa o módulo **ZX81**: só o tokenizer por enquanto
(`internal/platforms/zx81`, texto Classic ASCII já numerado →
imagem `.p` completa) — nenhum arquivo `.dzx` ZX81 compila ainda
(falta o Description/Lexer/Parser do dialeto), sem nenhum comando
novo pro usuário nesta release.

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
`theme [nome]`, `set <opção> [valor]`, `settings`, `dialect [nome]`
e `exit`/`quit` (ou `Ctrl-D` para sair) — os comandos não diferenciam
maiúsculas de minúsculas (`SET` funciona igual a `set`). `theme` sem
argumento mostra o tema ativo e a lista; por exemplo, `theme dracula`
ou `theme Solarized Light` troca o tema persistido para o console. Em um
terminal interativo, `Tab` completa comandos, temas e dialetos;
pressione `Tab` novamente para listar as alternativas (por exemplo,
depois de `theme `). A cor do prompt acompanha o tema selecionado.

`set <opção>` mostra o valor atual de uma das opções de configuração
do catálogo (lista completa em `OUTLINE.md`, Bloco 1/Fase 4);
`set <opção> <valor>` muda e grava
no SQLite local (ex.: `set INITIAL_LINE 1`). `settings` lista todas
as opções de Basic Dignified mais as do dialeto atual da sessão
(`dialect` sozinho mostra o dialeto atual e os disponíveis;
`dialect zx81` troca pra ZX81 — o padrão é MSX). Opções de Basic
Dignified (compartilhadas, ex. `INITIAL_LINE`/`REM_HEADER`) são
resolvidas sem precisar trocar de dialeto; uma opção específica de
dialeto (ex. `TOKENIZE`, que existe em MSX/ZX81/ZX Spectrum com
defaults diferentes) usa o dialeto atual da sessão. Nenhum valor
configurado aqui ainda é lido por um comando de conversão real — isso
chega junto com o motor (Blocos 1-4 do `OUTLINE.md`).

Na TUI, o menu **`Editor > Open`** abre o editor de texto/código
**WibEdit** (estilo WordStar/Norton Editor 2, portado do projeto
`wibworks` do mesmo autor) como uma tela cheia dentro da própria TUI —
não é um programa separado. É a base sobre a qual a futura IDE do
goBaDig será construída; ainda não está ligado a nada específico de
BASIC (sem realce de sintaxe ou abertura direta de `.bas`/fontes das 3
plataformas).

Na TUI, há um **desktop clássico inspirado no Turbo Pascal 6, QuickBasic
e Norton Editor 2.0**: fundo azul, barra de título, menu suspenso,
espaço de trabalho e barra de status. `Setup > Install` abre uma janela
central com um campo dropdown para o tema, a descrição correspondente e
os botões `Apply` e `Cancel`. As setas navegam os menus e campos; Enter
abre/seleciona e Esc fecha/cancela. Os seis temas são independentes dos
temas da GUI e a seleção é aplicada na hora e persistida em SQLite.

O menu `Setup` também tem 4 itens de configuração (desde a `v1.23.0`):
**Basic Dignified**, **MSX**, **ZX81** e **ZX Spectrum** — cada um
abre um formulário com todas as opções daquela categoria (as mesmas
do comando `set` do CLI, ver seção 5 acima): checkbox para opção
verdadeiro/falso, lista suspensa para opção de escolha fixa, campo
numérico ou de texto livre para o resto, com os valores atuais já
preenchidos. `Apply` valida e grava no mesmo SQLite do tema; `Cancel`
ou `Esc` descarta sem gravar nada. Nenhum valor configurado aqui
ainda é lido por um comando de conversão real.
Navegação: setas ou `Alt+letra` movem/abrem o menu, `Enter` confirma,
`Esc` fecha o menu suspenso. A GUI mostra uma janela com o mesmo menu
`File > Exit` / `Help > About` (nativo Fyne) e `Setup > Install`, onde
é possível escolher entre seis temas inspirados em VS Code Dark, Dracula,
Monokai, Nord, Solarized Light e Gruvbox Dark. A escolha da GUI também é
aplicada na hora e persistida em SQLite no diretório de configuração do
usuário.

O menu `Setup` da GUI também tem os mesmos 4 itens de configuração da
TUI (desde a `v1.24.0`): **Basic Dignified**, **MSX**, **ZX81** e
**ZX Spectrum** — cada um abre um formulário com as opções daquela
categoria (checkbox/lista suspensa/campo de texto conforme o tipo),
valores atuais já preenchidos, botões `Apply`/`Cancel`. Diferente da
TUI, campos numéricos validam em tempo real: `Apply` fica desabilitado
enquanto o valor estiver fora do intervalo permitido. O ciclo
completo de conversão (compilar e testar num emulador real) já
existe de verdade no CLI (`goBaDig <arquivo.dmx>`/`--execute
<arquivo.dmx>`, ou `compile`/`execute <arquivo.dmx>` no console, ver
seção 6), mas ainda não em nenhuma das duas interfaces gráficas —
TUI e GUI ainda não chamam esses comandos, e nenhum valor configurado
nos painéis (CLI, TUI ou GUI) ainda é lido por eles (usam sempre os
defaults do `badig.py` original).

## 6. Converter um programa Dignified

**Ciclo completo funciona de verdade pro MSX desde a `v1.28.0`, só
pelo CLI** (TUI/GUI ainda não chamam esses comandos — ver seção 5).
Histórico de como
o motor chegou até aqui: a partir da `v1.3.0` existe um primeiro
núcleo do motor em Go (`internal/core`), a
partir da `v1.4.0` o tokenizer MSX (`internal/platforms/msx`) já
converte texto Classic ASCII numerado em binário `.bmx`, e a partir da
`v1.5.0` o `Description`/`Lexer` MSX já classificam texto Dignified
real em tokens, a partir da `v1.6.0` o `Parser` MSX já aloca
variáveis de 2 letras e gera texto Classic de verdade pra isso
(`counter=0` → `zz = 0`), a partir da `v1.7.0` as conversões finais
do Pass 5 também funcionam (`x+=5` → `x = x + 5`,
`TRUE`/`FALSE` → `-1`/`0`, etc.), a partir da `v1.8.0` labels
(`{nome}`) resolvem pra números de linha de verdade (`goto {dest}` →
`goto 20`), a partir da `v1.9.0` loop labels também (`nome{ ... }`
vira um `GOTO` implícito de volta, `exit` salta pra fora), e a partir
da `v1.10.0` `DEFINE` expande macros de texto (`[nome](arg)` →
conteúdo registrado, com placeholder/valor padrão), a partir da
`v1.11.0` `FUNC`/`RET` também (`.func(5)` → `param=5:gosub <linha>`),
a partir da `v1.12.0` `DECLARE` reserva alias manual de variável
(`declare x:ab` → `x` sempre vira `ab`), a partir da `v1.13.0` `KEEP`
compila condicionalmente blocos marcados com `#tag` (sem `keep #a`
declarado, um bloco `#a...#a` some; com `keep #a`, fica), a partir da
`v1.14.0` `INCLUDE` mescla outro arquivo Dignified no meio do stream
(caminho relativo resolve contra o arquivo principal, não o que
inclui), a partir da `v1.15.0` o `generate()` já é de verdade
(espaçamento por palavra reservada — `print #1,1` ganha espaço só
depois do `print` — e limite de 256 caracteres por linha), e a
partir da `v1.16.0` os comentários exclusivos do Dignified
funcionam (`##` some em qualquer posição; um bloco de 3 cerquilhas
some inteiro; um bloco de duas aspas simples vira N linhas de
comentário classic reais), e a partir da `v1.17.0` um `:` clássico no
fim de linha também funde com a próxima (mantendo o próprio `:`,
diferente do `_`), e a partir da `v1.18.0` um cabeçalho REM
automático de 2 linhas é inserido e TAB é expandido pra espaço, e a
partir da `v1.19.0` um `_` fora do fim de linha vira o atalho de
`CALL` do MSX (`_CLS` equivale a `CALL CLS`, excluindo o nome
seguinte da renomeação de variável) — com isso todo o caminho textual
Dignified→Classic fechado —, e a partir da `v1.20.0` o resultado já
é tokenizado de verdade (`msx.CompileToBinary`, ligando ao tokenizer
MSX Batoken pronto desde a `v1.4.0`): o `CGK173.dmx` inteiro
compilado direto pra `.bmx` binário bate byte a byte contra o `.bmx`
real do `badig.py`. A partir da `v1.27.0` isso já tem um comando real
(`compilar`, ver "Uso" abaixo), e a partir da `v1.28.0` um segundo
comando (`execute`) fecha o ciclo completo, empacotando o resultado
num disco `.dsk` de teste e abrindo no openMSX de verdade.

### Uso

```console
goBaDig programa.dmx
```

Sem nenhuma opção antes do nome do arquivo, converte: gera
`programa.amx` (listagem Classic numerada, mesma convenção de fim de
linha do `badig.py` original) e `programa.bmx` (binário tokenizado),
os dois no mesmo diretório de `programa.dmx`. No console (`--cli`),
o mesmo comando é `compile programa.dmx`. Os dois caminhos chamam a
mesma função (`msx.CompileFile`), que já usa os mesmos defaults do
`badig.py` original (numeração a partir de 10, passo 10, cabeçalho
REM de 2 linhas, TAB expandido em 4 colunas) — ainda não há nenhuma
flag pra mudar isso a partir da linha de comando (as opções
persistidas pelos painéis de configuração — seção 5 — ainda não são
lidas por este comando).

```console
goBaDig --execute programa.dmx
```

Compila, empacota o resultado num disco `.dsk` novo (o programa
compilado gravado como `AUTOEXEC.BAS` — convenção da própria ROM de
disco do MSX pra rodar um programa sozinho assim que o disco dá
boot, sem precisar digitar nada) e abre o openMSX (precisa estar no
`PATH`) com esse disco montado, pra testar o programa na hora. No
console, o mesmo comando é `execute programa.dmx`. Os dois chamam
`msx.BuildTestDisk` + `msx.RunInEmulator`, reaproveitando a
biblioteca `.DSK` (`internal/platforms/msx/dsk`, seção acima) e o
comando `compilar`. Os arquivos reais de sistema MSX-DOS
(`MSXDOS.SYS`/`COMMAND.COM`) não são incluídos no disco de propósito
(são software de terceiros sob copyright) — o disco mostra um breve
"Boot error" antes de cair pro `AUTOEXEC.BAS` sozinho; não é um bug,
é só apertar uma tecla. Nenhum protocolo de monitoramento de erro em
tempo real ainda (mapear erro de execução pra linha Dignified de
origem) — o emulador abre numa janela normal, pro usuário acompanhar
e interagir diretamente.

Erro de conversão (arquivo não encontrado, sintaxe Dignified
inválida, etc.) imprime a mensagem na saída de erro e sai com código
diferente de zero (CLI) ou mostra a mensagem no próprio console
(`--cli`) — nenhum arquivo parcial é deixado pela metade nesse caso.

Só MSX por enquanto (ZX81/ZX Spectrum ainda não têm
`Description`/`Lexer`/`Parser` próprio implementado). TUI e GUI ainda
não têm nenhum desses comandos — próxima etapa natural depois de
ZX81/ZX Spectrum terem seu próprio motor de conversão.

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
byte a byte. O módulo MSX não tinha suíte de testes no original; os
vetores próprios já foram gerados rodando o tokenizer Python real
(`badig/msx/test_tokenizer_msx.py`, 13 casos — não versionado, já que
`badig/` inteiro é histórico local) e copiados, versionados, em
[`docs/golden-vectors/msx-tokenizer.md`](golden-vectors/msx-tokenizer.md)
(ver `OUTLINE.md`, Bloco 0). O tokenizer MSX em Go
(`internal/platforms/msx`) já passa todos os 13 byte-exato (ver
`go test ./internal/platforms/msx/...` e `OUTLINE.md`, Bloco 2).

O primeiro pacote do próprio motor em Go (`internal/core`: tipos
`Token`/`Position`, o contrato `Description`/`Lexer`/`Parser`/
`Settings`, e o pipeline `Lex`→`Parse`→`Generate`) já tem testes
automatizados validando o contrato contra uma plataforma fictícia
mínima — ver `go test ./internal/core/...` e `OUTLINE.md`, Bloco 1.

## 9. Integração com editores

Suporte planejado, por ordem de prioridade (ver `OUTLINE.md`, Bloco 8):

1. **Visual Studio Code** — um único VSIX cobrindo os três dialetos
   (MSX, ZX81, ZX Spectrum), substituindo as três extensões separadas do
   projeto Python.
2. **Sublime Text** — pacotes equivalentes (syntax, build system,
   snippets), um por plataforma ou unificados, a definir.
3. **Editor próprio** (baseado no Norton Editor 2.0) — **integrado a
   partir da `v1.2.0`**: o editor **WibEdit**, portado do projeto
   `wibworks` do mesmo autor, está embutido na TUI (`Editor > Open`,
   ver seção 5). A versão GUI do mesmo editor (também disponível em
   `wibworks`) ainda não foi portada. `msxide`/`paleobasic` (outras
   implementações do NE do usuário) continuam só como referência, não
   foram as escolhidas para esta integração. Ver `docs/RESOURCES.md`.
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
