# dist/

Esta pasta é o **destino de build local** do goBaDig: executáveis
compilados, arquivos auxiliares (banco SQLite inicial, exemplos,
documentação mínima) e o pacote `.zip` pronto para copiar para outra
máquina e testar — sem precisar instalar Go, nem o restante do
repositório.

**Nada aqui é versionado** (ver `.gitignore`), exceto este README. O
conteúdo é gerado pelo processo de build/release e pode ser apagado e
recriado a qualquer momento.

## O que entra em um pacote de release

Convenção (ver `OUTLINE.md`, seção "Versionamento e empacotamento"):

```
dist/
  goBaDig-vX.Y.Z/
    goBaDig.exe              (ou binário por plataforma: goBaDig-linux, goBaDig-darwin, ...)
    README.md                (cópia do README do projeto)
    CHANGELOG.md
    MANUAL.md                (ou docs/MANUAL.md)
    LICENSE
    examples/                (exemplos MSX/ZX81/ZXSpectrum prontos para testar)
  goBaDig-vX.Y.Z.zip          <- artefato final, anexado ao release do GitHub
```

Nome do projeto + versão no nome do zip, conforme pedido: `goBaDig-vX.Y.Z.zip`.

Até a primeira compilação real existir, esta pasta permanece vazia
(fora deste README).
