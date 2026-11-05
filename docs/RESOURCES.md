# Recursos externos (resource/)

Esta página documenta tudo o que deve viver em `resource/` — pasta que
**não é versionada no Git** porque contém material de terceiros (binários,
ROMs, fontes, ferramentas). Cada item abaixo lista: o que é, para que
serve no projeto, onde baixar o original e onde deve ficar dentro de
`resource/` para que o build/testes encontrem o arquivo em outra máquina.

Mantenha esta lista atualizada sempre que um novo recurso externo for
adotado. Se um recurso deixar de ser necessário, mova-o para a seção
"Descontinuados" em vez de apagar a entrada, para preservar o histórico
de decisão.

---

## Emuladores

### EightyOne

- **O que é**: emulador Windows de Sinclair ZX80/ZX81/ZX Spectrum (e
  variantes), usado como alvo de execução do módulo ZX81 e do módulo ZX
  Spectrum.
- **Para que serve**: validar manualmente que os binários gerados
  (`.p`, `.TAP`, `.TZX`, `.Z80`, `.SNA`) carregam e rodam de verdade.
- **Onde baixar**: <https://sourceforge.net/projects/eightyone-sinclair-emulator/>
- **Onde colocar**: `resource/emulators/eightyone/EightyOne.exe` (mais
  os perfis `.ini` que o projeto gerar/usar, ex.
  `eightyone48.ini`/`eightyone128.ini`, já presentes em
  `badig/zxspectrum/*.ini` como referência).
- **Observação**: path default herdado do projeto Python é
  `C:\zeddy\EightyOne.exe` — no Go isso vira uma configuração (SQLite),
  não mais hardcoded.

### openMSX

- **O que é**: emulador multiplataforma de MSX.
- **Para que serve**: alvo de execução do módulo MSX; é o único com
  monitoramento de execução/erro em tempo real (via protocolo stdio).
- **Onde baixar**: <https://openmsx.org/>
- **Onde colocar**: `resource/emulators/openmsx/` (instalação completa,
  conforme a distribuição oficial para o SO de teste).

### Spectaculator

- **O que é**: emulador Windows de ZX Spectrum, alternativa ao EightyOne.
- **Para que serve**: segunda opção de validação manual do módulo ZX
  Spectrum (só carregamento manual verificado até agora).
- **Onde baixar**: site oficial do Spectaculator (busca desta sessão não
  retornou uma URL estável/atual — confirmar o endereço antes de linkar
  em qualquer material publicado). Licença comercial — confirmar termos
  antes de redistribuir qualquer coisa relacionada.
- **Onde colocar**: `resource/emulators/spectaculator/`.

### ZEsarUX

- **O que é**: emulador multiplataforma (Linux/Mac/Windows) de máquinas
  Sinclair — ZX80, **ZX81**, **ZX Spectrum** (48K/128K/+2/+3), QL, Z88,
  Timex TS2068, Sam Coupé e outras. Listagens de terceiros também citam
  suporte a MSX/Amstrad/ColecoVision, **mas isso não está confirmado no
  README do próprio projeto** — verificar antes de depender disso para
  o módulo MSX.
- **Para que serve**: alvo de execução adicional para ZX81 e ZX Spectrum.
  **Ponto mais importante**: tem o **ZRCP** (ZEsarUX Remote Command
  Protocol), uma API via socket TCP (porta default `10000`, habilitada
  com `--enable-remoteprotocol` ou em Settings→Debug) que permite
  controlar o emulador e inspecionar estado remotamente — é um candidato
  real para o recurso de **monitoramento de erro em tempo real mapeado
  de volta para a linha Dignified** nas plataformas Sinclair, algo que
  o EightyOne não oferece (ver `OUTLINE.md`, seção 6, item de
  monitoramento). Merece ser avaliado como alvo prioritário de
  automação para ZX81/Spectrum, não só mais um emulador manual.
- **Onde baixar**: <https://github.com/chernandezba/zesarux>
- **Onde colocar**: `resource/emulators/zesarux/`.
- **Licença**: GNU GPL v3.

### fMSX

- **O que é**: emulador de MSX por Marat Fayzullin, multiplataforma
  (Windows/Linux/Mac), um dos emuladores MSX mais antigos e usados.
- **Para que serve**: alvo de execução alternativo ao openMSX para o
  módulo MSX. **Relevância extra**: o formato de fita `.CAS` do MSX (ver
  seção de formatos abaixo) **se origina deste emulador**.
- **Onde baixar**: <https://fms.komkon.org/fMSX/src/fMSX.html> (página
  do autor; confirmar se ainda resolve antes de depender dela — busca
  desta sessão encontrou referências mais recentes apontando para
  `komkon.org/fms/fMSX/`).
- **Onde colocar**: `resource/emulators/fmsx/`.
- **Licença**: segundo a FAQ do próprio autor, "livremente distribuível
  para uso não comercial" — **não é GPL-compatível** (o próprio autor já
  notou isso publicamente). Não redistribuir binários do fMSX dentro
  deste repositório; só documentar onde baixar.

### XZ81 (e sz81, sucessor mantido)

- **O que é**: `xz81` é um emulador ZX81 antigo para Linux/X11, de
  Russell Marks, baseado no `xz80` de Ian Collier — citado
  explicitamente pelo usuário. Não há mais um site/pacote oficial
  estável para ele; a build mais fácil de encontrar hoje é uma versão
  modificada por Michael Minn (compila com `make`).
- **Sucessor ativo**: **sz81**, mesma linhagem (`xz80`/`z81`), é o
  projeto que continua mantido atualmente.
- **Para que serve**: alvo de execução adicional para o módulo ZX81 em
  Linux, cobrindo quem não usa EightyOne/ZEsarUX.
- **Onde baixar**: xz81 (histórico) — ver a página de Michael Minn;
  sz81 (mantido) — <https://sz81.sourceforge.net/>.
- **Onde colocar**: `resource/emulators/xz81/` (ou `resource/emulators/sz81/`
  se optarmos pelo sucessor mantido — recomendado).
- **Licença**: sz81 é GPL v2.

---

## Formatos de mídia/fita/disco suportados (planejado)

Além de TAP/TZX/Z80/SNA (ZX Spectrum, já detalhados no `OUTLINE.md`,
seção 4.3) e `.p` (ZX81, seção 4.2), o usuário pediu suporte explícito
aos formatos abaixo. Nenhum destes é gerado pelo projeto Python original
— são **escopo novo** do goBaDig (registrado também no `OUTLINE.md`,
Blocos 2-4).

### ZX81 — `.81` / `.p81`

- `.81`: dados crus como salvos pela rotina de fita da ROM ZX81, **sem**
  o nome de arquivo — mesmo conteúdo que `.p`, só sem o cabeçalho de
  nome (fonte: nota técnica de formato ZX81, ver referências).
- `.p81`: igual, mas **incluindo** o nome do arquivo; permite guardar
  múltiplos programas no mesmo arquivo.
- `.TZX` também é usado para ZX81 por alguns emuladores (o formato TZX
  tem blocos genéricos reaproveitáveis para fita Sinclair de 16 bits,
  não é exclusivo do Spectrum) — avaliar se vale suportar como alias do
  `.p`/`.81` ou tratar como formato à parte.
- Existe ferramenta de conversão `.P`→`.81` ("P-to-81") no ecossistema
  retro — útil como referência de implementação, não como dependência.

### MSX — `.CAS` / `.DSK`

- **`.CAS`**: captura dos bytes decodificados de uma fita MSX **no
  nível do BIOS** (intercepta as rotinas `TAPION`/`TAPIN`/`TAPOON`/`TAPOUT`),
  não é uma gravação de forma de onda real — por isso não reproduz
  encoding FSK nem protections de fita. **Este formato se origina do
  próprio fMSX** e é amplamente suportado por outros emuladores (ex.
  openMSX). Nenhum CAS do MSX tem relação com o `.CAS` do CoCo
  (`badig/coco/cocotocas.py`) — mesmo nome, formatos totalmente
  diferentes, já observado no `badig/OUTLINE.md` original.
- **`.DSK`**: dump bruto de setores de um disquete formatado MSX-DOS
  (FAT12), ordenado por setor→lado→trilha, **sem** cabeçalho de
  geometria — emuladores inferem o tamanho do disco pelo tamanho do
  arquivo (ex. 720KB = duas faces). Não representa proteções de cópia
  nem disquetes fora do FAT padrão. Gerar um `.DSK` válido exige montar
  uma imagem MSX-DOS completa (boot sector, FAT, diretório) — tarefa
  bem mais complexa que TAP/TZX/CAS, provável prioridade mais baixa
  dentro do Bloco 2.

### ZX Spectrum — `.DSK` (+3, opcional/tardio)

- Disco de 3" do Spectrum +3 — fora do recorte 48K/128K atual; mantido
  como item opcional de baixa prioridade (ver `OUTLINE.md`, Bloco 10).

Referências consultadas nesta sessão (2026-11-05): wiki
["Emulation related file formats"](https://msx.org/wiki/Emulation_related_file_formats)
do msx.org (CAS/DSK), nota técnica
["O80 and P81 Format"](https://k1.spdns.de/Develop/Projects/zasm/Info/O80%20and%20P81%20Format.txt)
(formato `.81`/`.p81` do ZX81).

---

## Imagens de ROM

### ROM ZX81 (8K)

- **O que é**: imagem da ROM original do ZX81 (ou clone livre
  compatível), necessária apenas se/quando construirmos um emulador Z80
  próprio em Go para destokenizar binários ou validar execução (Bloco 10,
  tardio — ver `OUTLINE.md`).
- **Onde baixar**: confirmar fonte redistribuível antes de usar (ROMs
  originais Sinclair têm direitos autorais; preferir sempre que possível
  uma ROM aberta/clone compatível já usada por projetos de emulação).
- **Onde colocar**: `resource/roms/zx81-rom.bin`.

### ROM ZX Spectrum 48K

- **O que é**: imagem da ROM original de 48K, usada pelo SkoolKit
  (Python) hoje para simular a rotina real de LOAD; seria necessária
  também para um futuro emulador Z80 em Go (snapshots Z80/SNA nativos).
- **Onde baixar**: idem nota acima sobre direitos autorais — confirmar
  fonte antes de redistribuir qualquer coisa.
- **Onde colocar**: `resource/roms/zxspectrum-48k-rom.bin`.

---

## Fontes (para editores/temas)

### MSX Screen 0 New.ttf / CoCo Lower Clean.ttf

- **O que é**: fontes TrueType que simulam a tela original do MSX/CoCo,
  usadas pelos pacotes Sublime Text originais (`badig/msx/Sublime
  Package/.../extras/`, `badig/coco/Sublime Package/.../extras/`).
- **Para que serve**: eventual tema/preview fiel ao charset original em
  editores (baixa prioridade, Bloco 10).
- **Onde baixar**: já existem em `badig/msx/Sublime Package/.../extras/`
  e `badig/coco/Sublime Package/.../extras/` — copiar de lá, não é um
  download externo.
- **Onde colocar**: `resource/fonts/msx-screen0.ttf`.

---

## Ferramentas de conversão / dependências opcionais

### SkoolKit (Python)

- **O que é**: pacote Python usado hoje (versão Python do projeto) para
  gerar snapshots Z80/SNA do ZX Spectrum, simulando a rotina de LOAD da
  ROM 48K real.
- **Para que serve no Go**: não é mais um requisito do caminho principal.
  Só relevante se decidirmos manter uma ponte temporária Python↔Go para
  snapshots antes de termos um emulador Z80 nativo (ver `OUTLINE.md`,
  Bloco 4, item Z80/SNA — fase tardia).
- **Onde baixar**: `pip install skoolkit` (PyPI), versão `>=8.10,<11`.
- **Onde colocar**: não se aplica a `resource/` diretamente (é um
  pacote Python instalado via `pip`, não um arquivo solto) — documentar
  aqui só para registrar a dependência e a decisão de adiá-la.

---

## Material do projeto Python original (`badig/`)

O diretório `badig/` (projeto Python original + extensões ZX81/ZXSpectrum
do usuário) existe **só localmente, em disco, como referência** de
arquitetura, especificação de dialeto e oráculo de testes (golden
vectors) — não é um "recurso externo" no sentido desta página, mas
também **não é parte do histórico do Git deste repositório**.

**Decisão confirmada pelo usuário em 2026-11-05**: `badig/` nunca deve
ser commitado — serve só de referência durante o port, nada além disso.
Está listado em `.gitignore` para evitar que um `git add` amplo o inclua
por engano. Quem for reproduzir o ambiente em outra máquina precisa
copiar essa pasta manualmente (ela não vem do `git clone`).

---

## Descontinuados

_(nenhum item ainda)_
