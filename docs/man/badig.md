# compile

## SINOPSE

```
goBaDig <arquivo.dmx|arquivo.dzx|arquivo.dzs>
compile <arquivo.dmx|arquivo.dzx|arquivo.dzs>   (modo console)
```

## DESCRIÇÃO

Converte um arquivo Dignified (MSX ".dmx", ZX81 ".dzx" ou ZX Spectrum
".dzs") na listagem Classic numerada e no binário tokenizado do
dialeto correspondente, no mesmo diretório do arquivo de entrada:
".amx"/".bmx" para MSX, ".azx"/".p" para ZX81, ".azs"/".tap" para ZX
Spectrum. O dialeto é escolhido pela extensão do arquivo, não por uma
opção separada.

# execute

## SINOPSE

```
goBaDig --execute [--emulator <nome>] <arquivo.dmx|arquivo.dzx|arquivo.dzs>
execute <arquivo.dmx|arquivo.dzx|arquivo.dzs> [emulador]   (modo console)
```

## DESCRIÇÃO

Compila o arquivo (como "compile") e abre o resultado num emulador
real, pronto pra testar: MSX gera um disco ".dsk" de teste (o programa
compilado como AUTOEXEC.BAS) e abre no openMSX; ZX81 gera o ".p" direto
e abre no EightyOne; ZX Spectrum gera o ".tap" direto e abre também no
EightyOne. O emulador precisa estar disponível no PATH do sistema.

"--emulator"/o segundo argumento de "execute" escolhe outro emulador
em vez do padrão de cada plataforma: MSX aceita "openmsx" (padrão),
"fmsx" ou "bluemsx+"; ZX81 e ZX Spectrum aceitam "eightyone" (padrão)
ou "zesarux".

# set

## SINOPSE

```
set <OPCAO> [VALOR]
```

## DESCRIÇÃO

Sem VALOR, mostra o valor atual da opção. Com VALOR, valida e grava o
novo valor. A opção é procurada primeiro nas opções comuns (Basic
Dignified), depois nas opções do dialeto atual - troque o dialeto
atual com "dialect <nome>" se a opção pertencer a outro dialeto.

# settings

## SINOPSE

```
settings
```

## DESCRIÇÃO

Lista todas as opções (e seus valores atuais) das opções comuns (Basic
Dignified) e do dialeto atual.

# dialect

## SINOPSE

```
dialect [nome]
```

## DESCRIÇÃO

Sem argumento, mostra o dialeto atual da sessão de console e a lista
dos disponíveis. Com um nome (msx, zx81 ou zxspectrum), troca o
dialeto atual - usado por "set"/"settings" para saber em qual conjunto
de opções procurar uma chave que não seja comum aos três.

# theme

## SINOPSE

```
theme [nome]
```

## DESCRIÇÃO

Sem argumento, mostra o tema ativo do console e a lista dos
disponíveis. Com um nome, troca o tema ativo.

# version

## SINOPSE

```
version
```

## DESCRIÇÃO

Mostra a versão do goBaDig (bloco, feature e build, no esquema X.Y.Z
próprio do projeto - ver CHANGELOG.md, não é SemVer).

# man

## SINOPSE

```
goBaDig --man [tópico]
man [tópico]   (modo console)
```

## DESCRIÇÃO

Sem tópico, mostra o índice de tópicos disponíveis. Com um tópico,
mostra a página de referência correspondente - os próprios comandos do
goBaDig (como este documento), o dicionário de BASIC do MSX e o manual
do compilador Assembly ASMSX.

# incbload

## SINOPSE

```
[INCBLOAD]("arquivo.asm")
```

## DESCRIÇÃO

Ponte entre um programa Assembly (compilado com o asMSX, ver
"man asmsx") e o programa MSX BASIC que o chama. Só existe em projetos
MSX (".dmx") e exige que "arquivo.asm" já tenha sido compilado antes
(comando "compile"/tecla F2) - ela nunca compila o Assembly por conta
própria, só lê o que já foi gerado.

Em cada compilação do ".dmx", "[INCBLOAD]("arquivo.asm")" é expandido
para o equivalente a:

```
BLOAD"ARQUIVO.BIN":DEFUSR=&Hxxxx
```

onde "ARQUIVO.BIN" é o binário que o asMSX gerou e "&Hxxxx" é o
endereço de entrada da rotina, lido automaticamente da diretiva
".START" do ".asm" (ou de ".ORG", se não houver ".START") - incluindo
o caso em que ".START"/".ORG" apontam pra um rótulo em vez de um
número, resolvido pelo arquivo ".sym" que o próprio asMSX gera junto
com o ".bin". Sem o sufixo ",R" do BLOAD, de propósito: ele dispara a
execução automaticamente ao carregar, o que serve pra um programa que
toma conta da máquina inteira, não pra uma rotina pensada para ser
chamada (inclusive repetidas vezes) via USR() - depois do BLOAD, o
programa BASIC chama a rotina normalmente com "x=USR(0)" ou
"print USR(0)", onde for preciso.

Como nada é escrito no ".dmx" - a expansão acontece só em memória, a
cada compilação - mudar o ".ORG"/".START" do ".asm" e recompilar os
dois (".asm" primeiro, depois o ".dmx") já basta pra atualizar o
endereço automaticamente, sem editar nada à mão.

Limitação conhecida: um rótulo cujo nome parece um número hexadecimal
válido com sufixo "h" (ex. um rótulo chamado "ABh") pode ser lido como
literal numérico em vez de rótulo - evite nomes assim numa linha
".START"/".ORG" usada com INCBLOAD.
