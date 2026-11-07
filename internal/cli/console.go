package cli

import (
	"bufio"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode"

	"golang.org/x/term"

	"github.com/wilsonpilon/goBaDig/internal/about"
	"github.com/wilsonpilon/goBaDig/internal/config"
	"github.com/wilsonpilon/goBaDig/internal/theme"
	"github.com/wilsonpilon/goBaDig/internal/version"
)

const defaultCLIThemeID = "cli-vscode-dark"

type consoleState struct {
	out         io.Writer
	store       *config.Store
	activeTheme theme.Theme
	hasTheme    bool
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

func (s *consoleState) loadTheme() error {
	if s.store == nil {
		store, err := config.OpenDefault()
		if err != nil {
			return err
		}
		s.store = store
		if err := store.InstallThemes(theme.CatalogFor(theme.ModeCLI)); err != nil {
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
	argument = strings.TrimSpace(argument)
	switch command {
	case "exit", "quit":
		return true
	case "help":
		fmt.Fprint(s.out, HelpText())
	case "man":
		fmt.Fprintln(s.out, about.Man())
	case "about":
		fmt.Fprintln(s.out, about.Short())
	case "version":
		fmt.Fprintf(s.out, "goBaDig v%s (bloco %s, feature %s)\n",
			version.Version, version.Block, version.Feature)
	case "theme":
		s.executeTheme(argument)
	default:
		fmt.Fprintf(s.out, "comando desconhecido: %q (digite \"help\")\n", line)
	}
	return false
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
