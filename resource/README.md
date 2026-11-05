# resource/

Esta pasta guarda material de terceiros que o projeto **estuda, manipula
ou reempacota**, mas que **não temos o direito (ou não faz sentido) de
redistribuir dentro do repositório Git** — binários de emulador, imagens
de ROM, fontes, conversores externos, etc.

**Nada aqui é versionado** (ver `.gitignore`), exceto este README.

A lista completa de recursos esperados aqui — o que é, para que serve,
onde baixar o original e como recolocá-lo nesta pasta em outra máquina —
fica em **[`docs/RESOURCES.md`](../docs/RESOURCES.md)**. Mantenha esse
arquivo atualizado sempre que um novo recurso externo for adotado pelo
projeto.

Estrutura sugerida (preenchida conforme os recursos forem usados):

```
resource/
  emulators/
    eightyone/       (EightyOne.exe + perfis .ini de referência)
    openmsx/
    spectaculator/
  roms/
    zx81-rom.bin
    zxspectrum-48k-rom.bin
  fonts/
    msx-screen0.ttf
    coco-lower-clean.ttf
  skoolkit/           (opcional, só se mantivermos o caminho Python/SkoolKit
                        como ponte temporária para snapshots Z80/SNA)
```
