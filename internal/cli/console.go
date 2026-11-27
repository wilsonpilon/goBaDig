package cli

import (
	"bufio"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"golang.org/x/term"

	"github.com/wilsonpilon/goBaDig/internal/about"
	"github.com/wilsonpilon/goBaDig/internal/config"
	"github.com/wilsonpilon/goBaDig/internal/dialect"
	"github.com/wilsonpilon/goBaDig/internal/man"
	"github.com/wilsonpilon/goBaDig/internal/options"
	"github.com/wilsonpilon/goBaDig/internal/platforms/msx"
	"github.com/wilsonpilon/goBaDig/internal/theme"
	"github.com/wilsonpilon/goBaDig/internal/version"
)

const defaultCLIThemeID = "cli-vscode-dark"

// knownDialects lists the options.Category* values "dialect" can
// switch between - the ones a dialect-specific SET actually resolves
// against (options.CategoryCore is always searched too, regardless
// of the current dialect, since it is shared across all three).
var knownDialects = []string{options.CategoryMSX, options.CategoryZX81, options.CategoryZXSpectrum}

type consoleState struct {
	out         io.Writer
	store       *config.Store
	activeTheme theme.Theme
	hasTheme    bool
	// dialect is the SESSION's current dialect (SET falls back to it
	// when a key isn't one of options.CategoryCore's own) - purely
	// in-memory for this round, reset every time the console starts;
	// persisting "last dialect used" is deferred (OUTLINE.md, Fase 4).
	// Empty means "not chosen yet" - use currentDialect(), never this
	// field directly, so a consoleState built as a struct literal
	// (every existing test does this) still gets a sane default.
	dialect string
}

func (s *consoleState) currentDialect() string {
	if s.dialect == "" {
		return options.CategoryMSX
	}
	return s.dialect
}

// RunConsole implements the interactive prompt. On a real terminal it reads
// individual keys for shell-style completion; other readers remain line-based.
func RunConsole(in io.Reader, out io.Writer) (runErr error) {
	state := &consoleState{out: out}
	defer func() {
		if state.store != nil {
			if err := state.store.Close(); err != nil {
				runErr = errors.Join(runErr, fmt.Errorf("close settings database: %w", err))
			}
		}
	}()

	fmt.Fprint(out, Banner())
	fmt.Fprintln(out, `Modo console. Digite "help" para ver os comandos, "exit" para sair.`)
	file, isFile := in.(*os.File)
	if isFile && term.IsTerminal(int(file.Fd())) {
		if err := state.loadTheme(); err != nil {
			return err
		}
		return state.runTerminal(file)
	}

	scanner := bufio.NewScanner(in)
	for {
		state.writePrompt(false)
		if !scanner.Scan() {
			fmt.Fprintln(out)
			if err := scanner.Err(); err != nil {
				return fmt.Errorf("read console input: %w", err)
			}
			return nil
		}
		if state.execute(strings.TrimSpace(scanner.Text())) {
			return nil
		}
	}
}

func (s *consoleState) writePrompt(color bool) {
	if color && s.hasTheme {
		fmt.Fprintf(s.out, "%sgoBaDig>%s ", ansiColor(s.activeTheme.Palette.Primary, false), "\x1b[0m")
		return
	}
	fmt.Fprint(s.out, "goBaDig> ")
}

// openStore opens the shared settings database on first use, leaving
// it in place for every later call this console session makes -
// loadTheme and loadOptions both call this before doing their own,
// separate first-use setup (installing the theme catalog / seeding
// option defaults).
func (s *consoleState) openStore() error {
	if s.store != nil {
		return nil
	}
	store, err := config.OpenDefault()
	if err != nil {
		return err
	}
	s.store = store
	return nil
}

func (s *consoleState) loadTheme() error {
	if err := s.openStore(); err != nil {
		return err
	}
	if !s.hasTheme {
		if err := s.store.InstallThemes(theme.CatalogFor(theme.ModeCLI)); err != nil {
			return fmt.Errorf("install CLI themes: %w", err)
		}
	}
	item, err := s.store.EnsureActiveTheme(theme.ModeCLI, defaultCLIThemeID)
	if err != nil {
		return err
	}
	s.activeTheme = item
	s.hasTheme = true
	return nil
}

func (s *consoleState) execute(line string) bool {
	if line == "" {
		return false
	}
	command, argument, _ := strings.Cut(line, " ")
	// Commands match case-insensitively (not just "set" - every
	// command does, uniformly) since SET's own natural style, the
	// user's own worked example ("SET INITIAL_LINE 1"), spells it in
	// caps the way classic BASIC keywords are conventionally written.
	command = strings.ToLower(command)
	argument = strings.TrimSpace(argument)
	switch command {
	case "exit", "quit":
		return true
	case "help":
		fmt.Fprint(s.out, HelpText())
	case "man":
		s.executeMan(argument)
	case "about":
		fmt.Fprintln(s.out, about.Short())
	case "version":
		fmt.Fprintf(s.out, "goBaDig v%s (bloco %s, feature %s)\n",
			version.Version, version.Block, version.Feature)
	case "theme":
		s.executeTheme(argument)
	case "dialect":
		s.executeDialect(argument)
	case "set":
		s.executeSet(argument)
	case "settings":
		s.executeSettings()
	case "compile":
		s.executeCompile(argument)
	case "execute":
		s.executeExecute(argument)
	default:
		fmt.Fprintf(s.out, "comando desconhecido: %q (digite \"help\")\n", line)
	}
	return false
}

// executeMan implements the console's own "man [tópico]" - the
// interactive-mode counterpart to "goBaDig --man [tópico]" (ver
// cli.go's own runMan), both wrapping internal/man.
func (s *consoleState) executeMan(topic string) {
	if topic == "" {
		fmt.Fprintln(s.out, about.Man())
		return
	}
	page, err := man.Lookup(topic)
	if err != nil {
		fmt.Fprintln(s.out, err)
		return
	}
	for _, line := range page.PlainLines() {
		fmt.Fprintln(s.out, line)
	}
}

func (s *consoleState) executeTheme(name string) {
	if err := s.loadTheme(); err != nil {
		fmt.Fprintf(s.out, "erro ao carregar temas: %v\n", err)
		return
	}
	if name == "" {
		fmt.Fprintf(s.out, "Tema CLI ativo: %s (%s)\nTemas disponíveis: %s\n",
			s.activeTheme.Name, cliThemeAlias(s.activeTheme), strings.Join(cliThemeAliases(theme.CatalogFor(theme.ModeCLI)), ", "))
		return
	}

	items, err := s.store.ListThemes(theme.ModeCLI)
	if err != nil {
		fmt.Fprintf(s.out, "erro ao listar temas: %v\n", err)
		return
	}
	for _, item := range items {
		if normalizeThemeName(name) != normalizeThemeName(item.Name) &&
			normalizeThemeName(name) != cliThemeAlias(item) {
			continue
		}
		if err := s.store.SetActiveTheme(theme.ModeCLI, item.ID); err != nil {
			fmt.Fprintf(s.out, "erro ao ativar tema %q: %v\n", item.Name, err)
			return
		}
		s.activeTheme = item
		fmt.Fprintf(s.out, "Tema CLI ativo: %s\n", item.Name)
		return
	}
	fmt.Fprintf(s.out, "tema desconhecido: %q. Disponíveis: %s\n", name, strings.Join(cliThemeAliases(items), ", "))
}

func normalizeThemeName(value string) string {
	value = strings.ReplaceAll(value, "-", " ")
	return strings.Join(strings.Fields(strings.ToLower(strings.TrimSpace(value))), " ")
}

func cliThemeAlias(item theme.Theme) string {
	return strings.TrimPrefix(item.ID, theme.ModeCLI+"-")
}

func cliThemeAliases(items []theme.Theme) []string {
	aliases := make([]string, 0, len(items))
	for _, item := range items {
		if item.Mode != theme.ModeCLI {
			continue
		}
		aliases = append(aliases, cliThemeAlias(item))
	}
	return aliases
}

// executeDialect shows or switches the console session's current
// dialect - SET falls back to this category when a key isn't one of
// options.CategoryCore's own (every dialect-specific option, e.g.
// MSX's own TOKENIZE vs ZX81's own TOKENIZE, needs this to pick which
// one is meant).
func (s *consoleState) executeDialect(name string) {
	if name == "" {
		fmt.Fprintf(s.out, "Dialeto atual: %s\nDisponíveis: %s\n",
			s.currentDialect(), strings.Join(knownDialects, ", "))
		return
	}
	normalized := strings.ToLower(strings.TrimSpace(name))
	for _, d := range knownDialects {
		if d == normalized {
			s.dialect = d
			fmt.Fprintf(s.out, "Dialeto atual: %s\n", d)
			return
		}
	}
	fmt.Fprintf(s.out, "dialeto desconhecido: %q. Disponíveis: %s\n", name, strings.Join(knownDialects, ", "))
}

// loadOptions opens the shared store (if not already open) and seeds
// category's own catalog defaults - same lazy "do it on first use"
// shape as loadTheme, and just as idempotent to call again
// (EnsureOptionDefaults never reverts a value already changed).
func (s *consoleState) loadOptions(category string) error {
	if err := s.openStore(); err != nil {
		return err
	}
	return s.store.EnsureOptionDefaults(category, options.Catalog())
}

// findOptionCategory resolves which category a bare key (as typed at
// "SET <key> ...", with no category prefix - matching the user's own
// example, "SET INITIAL_LINE 1") belongs to: options.CategoryCore is
// always tried first (its keys never collide with any dialect's
// own), falling back to the session's current dialect.
func (s *consoleState) findOptionCategory(key string) (string, error) {
	if _, ok := options.Find(options.CategoryCore, key); ok {
		return options.CategoryCore, nil
	}
	dialect := s.currentDialect()
	if _, ok := options.Find(dialect, key); ok {
		return dialect, nil
	}
	return "", fmt.Errorf("opção desconhecida: %q (dialeto atual: %s - troque com \"dialect <nome>\" se a opção for de outro dialeto)", key, dialect)
}

// executeSet implements "SET <OPCAO> [VALOR]": with no value, shows
// the option's current stored value; with one, validates and
// persists it (config.Store.SetOption already validates against the
// catalog - this only needs to resolve WHICH category the bare key
// belongs to first).
func (s *consoleState) executeSet(argument string) {
	if argument == "" {
		fmt.Fprintln(s.out, `uso: set <OPCAO> [VALOR] (sem VALOR, mostra o valor atual; "settings" lista tudo)`)
		return
	}
	key, value, hasValue := strings.Cut(argument, " ")
	key = strings.TrimSpace(key)
	value = strings.TrimSpace(value)

	category, err := s.findOptionCategory(key)
	if err != nil {
		fmt.Fprintln(s.out, err)
		return
	}
	if err := s.loadOptions(category); err != nil {
		fmt.Fprintf(s.out, "erro ao carregar opções: %v\n", err)
		return
	}

	field, _ := options.Find(category, key)
	if !hasValue || value == "" {
		current, err := s.store.GetOption(category, field.Key)
		if err != nil {
			fmt.Fprintf(s.out, "erro ao ler opção: %v\n", err)
			return
		}
		fmt.Fprintf(s.out, "%s = %s\n", field.Key, current)
		return
	}
	if err := s.store.SetOption(category, field.Key, value); err != nil {
		fmt.Fprintf(s.out, "erro: %v\n", err)
		return
	}
	fmt.Fprintf(s.out, "%s = %s\n", field.Key, value)
}

// executeSettings lists every option (and its current value) in
// options.CategoryCore plus the session's current dialect - the two
// categories executeSet itself resolves a bare key against.
func (s *consoleState) executeSettings() {
	dialect := s.currentDialect()
	fmt.Fprintln(s.out, "Basic Dignified:")
	if err := s.printOptions(options.CategoryCore); err != nil {
		fmt.Fprintf(s.out, "  erro: %v\n", err)
	}
	fmt.Fprintf(s.out, "%s:\n", dialect)
	if err := s.printOptions(dialect); err != nil {
		fmt.Fprintf(s.out, "  erro: %v\n", err)
	}
}

func (s *consoleState) printOptions(category string) error {
	if err := s.loadOptions(category); err != nil {
		return err
	}
	values, err := s.store.ListOptions(category)
	if err != nil {
		return err
	}
	for _, field := range options.ForCategory(category) {
		fmt.Fprintf(s.out, "  %s = %s\n", field.Key, values[field.Key])
	}
	return nil
}

// executeCompile implements the console's own "compile
// <arquivo.dmx|arquivo.dzx|arquivo.dzs|arquivo.asm>" command - the
// interactive-mode counterpart to "goBaDig <arquivo.dmx>" at the
// non-interactive command line (ver cli.go's own runCompile), both
// wrapping the same dialect.CompileByExtension dispatch
// (internal/dialect) for the three Dignified dialects, plus an ASMSX
// branch (Fase 4 of the autocomplete+man/help+ASMSX plan, ver
// OUTLINE.md) for ".asm" - not part of dialect.CompileByExtension's
// own switch, since ASMSX is an MSX-specific tool, not a Dignified
// dialect.
func (s *consoleState) executeCompile(path string) {
	if path == "" {
		fmt.Fprintln(s.out, "uso: compile <arquivo.dmx|arquivo.dzx|arquivo.dzs|arquivo.asm>")
		return
	}
	if strings.EqualFold(filepath.Ext(path), ".asm") {
		s.executeCompileAsm(path)
		return
	}
	classicPath, binaryPath, err := dialect.CompileByExtension(path)
	if err != nil {
		fmt.Fprintf(s.out, "erro: %v\n", err)
		return
	}
	fmt.Fprintf(s.out, "Convertido: %s\n  -> %s\n  -> %s\n", path, classicPath, binaryPath)
}

// executeCompileAsm resolves the asMSX executable (internal/options'
// ASMSX_PATH, falling back to PATH - ver internal/platforms/msx's own
// ResolveAsmsxPath) and compiles path with it, opening the settings
// store first if this console session hasn't already (same lazy
// "do it on first use" shape as loadOptions/loadTheme).
func (s *consoleState) executeCompileAsm(path string) {
	if err := s.openStore(); err != nil {
		fmt.Fprintf(s.out, "erro ao abrir configurações: %v\n", err)
		return
	}
	asmsxExe, err := msx.ResolveAsmsxPath(s.store)
	if err != nil {
		fmt.Fprintf(s.out, "erro: %v\n", err)
		return
	}
	binPath, output, err := msx.CompileAsm(path, asmsxExe)
	if err != nil {
		fmt.Fprintf(s.out, "erro: %v\n", err)
		return
	}
	if output != "" {
		fmt.Fprintln(s.out, output)
	}
	fmt.Fprintf(s.out, "Compilado: %s\n  -> %s\n", path, binPath)
}

// executeExecute implements the console's own "execute
// <arquivo.dmx|arquivo.dzx|arquivo.dzs> [emulador]" - the user's own
// "edit/badig/execute" workflow vocabulary: compile, build whatever
// ready-to-run artifact its own dialect needs and open it in a real
// emulator. Counterpart to cli.go's own runExecute (--execute
// --emulator), both wrapping dialect.RunInEmulatorByExtension
// (internal/dialect). The optional second token picks an explicit
// backend (internal/dialect's own EmulatorBackends) instead of each
// dialect's own default - added 2026-11-26 at the user's own request;
// argument was already the WHOLE rest of the line past "execute " (ver
// execute's own strings.Cut), so this does one more split on the
// first remaining space, same convention.
func (s *consoleState) executeExecute(argument string) {
	path, backend, _ := strings.Cut(argument, " ")
	backend = strings.TrimSpace(backend)
	if path == "" {
		fmt.Fprintln(s.out, "uso: execute <arquivo.dmx|arquivo.dzx|arquivo.dzs> [emulador]")
		return
	}
	artifactPath, err := dialect.RunInEmulatorByExtension(path, backend)
	if err != nil {
		fmt.Fprintf(s.out, "erro: %v\n", err)
		return
	}
	fmt.Fprintf(s.out, "Gerado: %s\nAbrindo no emulador...\n", artifactPath)
}

func ansiColor(colorHex string, background bool) string {
	if len(colorHex) != 7 || colorHex[0] != '#' {
		return ""
	}
	rgb, err := hex.DecodeString(colorHex[1:])
	if err != nil || len(rgb) != 3 {
		return ""
	}
	code := 38
	if background {
		code = 48
	}
	return fmt.Sprintf("\x1b[%d;2;%d;%d;%dm", code, rgb[0], rgb[1], rgb[2])
}

func (s *consoleState) runTerminal(file *os.File) (runErr error) {
	state, err := term.MakeRaw(int(file.Fd()))
	if err != nil {
		return fmt.Errorf("enable terminal input mode: %w", err)
	}
	defer func() {
		if err := term.Restore(int(file.Fd()), state); err != nil {
			runErr = errors.Join(runErr, fmt.Errorf("restore terminal input mode: %w", err))
		}
	}()

	reader := bufio.NewReader(file)
	colorPrompt := false
	if outputFile, ok := s.out.(*os.File); ok {
		colorPrompt = term.IsTerminal(int(outputFile.Fd()))
	}
	line := make([]rune, 0, 80)
	var completion completionState

	redraw := func() {
		fmt.Fprint(s.out, "\r\x1b[2K")
		s.writePrompt(colorPrompt)
		fmt.Fprint(s.out, string(line))
	}
	s.writePrompt(colorPrompt)
	for {
		r, _, readErr := reader.ReadRune()
		if readErr != nil {
			fmt.Fprintln(s.out)
			if errors.Is(readErr, io.EOF) {
				return nil
			}
			return fmt.Errorf("read terminal input: %w", readErr)
		}
		switch r {
		case 0x1b:
			// Ignore terminal navigation sequences (for example, arrow keys).
			if next, _, err := reader.ReadRune(); err == nil && next == '[' {
				for {
					last, _, err := reader.ReadRune()
					if err != nil || (last >= '@' && last <= '~') {
						break
					}
				}
			}
		case '\r', '\n':
			fmt.Fprintln(s.out)
			if s.execute(strings.TrimSpace(string(line))) {
				return nil
			}
			line = line[:0]
			completion = completionState{}
			s.writePrompt(colorPrompt)
		case '\t':
			current := string(line)
			completed, candidates, show := completion.tab(current)
			if show {
				fmt.Fprintln(s.out)
				fmt.Fprintln(s.out, strings.Join(candidates, "  "))
				redraw()
				continue
			}
			if len(candidates) == 0 {
				fmt.Fprint(s.out, "\a")
				continue
			}
			line = []rune(completed)
			redraw()
		case '\b', 0x7f:
			if len(line) > 0 {
				line = line[:len(line)-1]
				redraw()
			}
			completion = completionState{}
		case 0x03:
			fmt.Fprintln(s.out, "^C")
			line = line[:0]
			completion = completionState{}
			s.writePrompt(colorPrompt)
		case 0x04:
			if len(line) == 0 {
				fmt.Fprintln(s.out)
				return nil
			}
		default:
			if unicode.IsPrint(r) {
				line = append(line, r)
				fmt.Fprint(s.out, string(r))
				completion = completionState{}
			}
		}
	}
}
