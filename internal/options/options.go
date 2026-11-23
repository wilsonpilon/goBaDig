// Package options is the catalog of every badig.py configuration knob
// (CLI flag / .ini key / remtag in the original engine -
// support/badig_settings.py plus each dialect's own
// badig_*.py/tokenizer_interface.py/emulator_interface.py) that this
// project persists as a user-editable setting, instead of a
// per-invocation flag. See OUTLINE.md's "Fase 4 - Config em camadas"
// for the layering this is the first piece of (SQLite global for
// now; project-scope and remtag resolution are later layers, deferred
// until a real compile command exists to anchor "project" against).
//
// Mirrors internal/theme's own split: this package is the
// embedded-in-the-binary CATALOG (what options exist, their type and
// default); internal/config.Store is the SQLite-backed persistence
// layer that reads/writes actual values, validating against this
// catalog the same way it already validates theme.Theme via
// Theme.Validate before writing.
package options

import (
	"fmt"
	"strconv"
	"strings"
)

// Type is the kind of value a Field holds. Every value is still
// stored and passed around as a plain string (same representation
// SQLite's TEXT column holds) - Type only drives validation and,
// later, how a TUI/GUI form renders the field's widget.
type Type string

const (
	TypeBool   Type = "bool"
	TypeInt    Type = "int"
	TypeString Type = "string"
	TypeEnum   Type = "enum"
)

// Field describes one configurable option: which category (panel) it
// belongs to, its key (the name used in "SET <key> <value>" and
// shown in the TUI/GUI forms - deliberately NOT the original engine's
// own Python field name, which only appears in this file's comments
// for traceability), its Type, and its Default value.
type Field struct {
	Category    string
	Key         string
	Type        Type
	Default     string
	Description string

	// Enum lists the allowed values, case-insensitively matched -
	// only meaningful when Type is TypeEnum.
	Enum []string
	// Min/Max bound a TypeInt field; nil means unbounded on that side.
	Min, Max *int
}

// Find looks up one Field by category and key, matching Key
// case-insensitively (a human typing "SET initial_line 1" at the CLI
// should work the same as "SET INITIAL_LINE 1") - Category is matched
// exactly, since it is never user-typed yet (no CLI/TUI/GUI wiring
// selects a category in this round; it is always passed as one of
// the package's own literal category constants by calling code).
func Find(category, key string) (Field, bool) {
	for _, f := range Catalog() {
		if f.Category == category && strings.EqualFold(f.Key, key) {
			return f, true
		}
	}
	return Field{}, false
}

// ForCategory returns every Field belonging to one category, in
// Catalog's own order - the list a TUI/GUI panel renders as its form.
func ForCategory(category string) []Field {
	var out []Field
	for _, f := range Catalog() {
		if f.Category == category {
			out = append(out, f)
		}
	}
	return out
}

// Validate reports whether raw is a legal value for the named option,
// looking it up via Find first. Mirrors the shape of theme.Theme's
// own Validate method: config.Store calls this before writing, the
// same way it already calls Theme.Validate before writing a theme.
func Validate(category, key, raw string) error {
	field, ok := Find(category, key)
	if !ok {
		return fmt.Errorf("unknown option %s.%s", category, key)
	}
	switch field.Type {
	case TypeBool:
		if raw != "true" && raw != "false" {
			return fmt.Errorf("option %s.%s must be true or false, got %q", category, field.Key, raw)
		}
	case TypeInt:
		n, err := strconv.Atoi(raw)
		if err != nil {
			return fmt.Errorf("option %s.%s must be an integer, got %q", category, field.Key, raw)
		}
		if field.Min != nil && n < *field.Min {
			return fmt.Errorf("option %s.%s must be >= %d, got %d", category, field.Key, *field.Min, n)
		}
		if field.Max != nil && n > *field.Max {
			return fmt.Errorf("option %s.%s must be <= %d, got %d", category, field.Key, *field.Max, n)
		}
	case TypeEnum:
		for _, allowed := range field.Enum {
			if strings.EqualFold(allowed, raw) {
				return nil
			}
		}
		return fmt.Errorf("option %s.%s must be one of %v, got %q", category, field.Key, field.Enum, raw)
	case TypeString:
		// Any string is legal.
	}
	return nil
}

// Categories, matching the four panels the user asked for: Basic
// Dignified (shared, dialect-agnostic), and one per dialect (each
// bundling that dialect's own Settings plus its tokenizer and
// emulator interface options together, same as the user's own
// request: "outra com os parametros especificos de MSX e do
// emulador do MSX", etc.).
const (
	CategoryCore       = "core"
	CategoryMSX        = "msx"
	CategoryZX81       = "zx81"
	CategoryZXSpectrum = "zxspectrum"
)

// Catalog is every option this project currently models, grouped by
// category in the order panels should render them. Field.Key names
// are deliberately friendly/renamed (ALL_CAPS_WITH_UNDERSCORES,
// matching the user's own "SET INITIAL_LINE 1" example) rather than
// the original engine's own snake_case Python field names - each
// entry's doc comment below names that original field for
// traceability only; it is never user-facing.
//
// Deliberately NOT modeled here (documented, not a persistent
// setting): write_ini_file/-ini (one-shot CLI action, not state to
// persist), -asc (a per-invocation mode flag), input/output file
// paths (per-invocation, not a setting), ARGUMENTS/HELP remtags (a
// source-code-level mechanism, a different config layer entirely -
// see this package's own doc comment), and system_id/dialect
// selection (implied by which category/panel is being edited, not a
// setting within one).
func Catalog() []Field {
	return []Field{
		// --- core (badig_settings.py, shared across every dialect) ---
		{Category: CategoryCore, Key: "INITIAL_LINE", Type: TypeInt, Default: "10",
			Description: "Primeira linha numerada do Classic gerado.", Min: new(0)}, // line_start
		{Category: CategoryCore, Key: "LINE_INCREMENT", Type: TypeInt, Default: "10",
			Description: "Passo entre números de linha consecutivos.", Min: new(1)}, // line_step
		{Category: CategoryCore, Key: "TAB_SIZE", Type: TypeInt, Default: "4",
			Description: "Quantas colunas um TAB avança ao ser expandido em espaços.", Min: new(1)}, // tab_lenght
		{Category: CategoryCore, Key: "REM_HEADER", Type: TypeBool, Default: "true",
			Description: "Insere o cabeçalho REM de 2 linhas no início do programa."}, // rem_header
		{Category: CategoryCore, Key: "STRIP_SPACES", Type: TypeBool, Default: "false",
			Description: "Remove todo espaço opcional da listagem Classic gerada."}, // strip_spaces
		{Category: CategoryCore, Key: "CAPITALIZE_ALL", Type: TypeBool, Default: "false",
			Description: "Capitaliza toda palavra reservada na saída."}, // capitalise_all
		{Category: CategoryCore, Key: "TRANSLATE_UNICODE", Type: TypeBool, Default: "false",
			Description: "Traduz caracteres Unicode gráficos pro equivalente ASCII/nativo."}, // translate
		{Category: CategoryCore, Key: "VERBOSE_LEVEL", Type: TypeInt, Default: "3",
			Description: "Nível de detalhe das mensagens de conversão (0 silencioso - 4 detalhado).",
			Min:         new(0), Max: new(4)}, // verbose_level
		{Category: CategoryCore, Key: "PRINT_REPORTS_TO_CONSOLE", Type: TypeBool, Default: "false",
			Description: "Imprime os relatórios no console em vez de salvar em arquivo."}, // print_report
		{Category: CategoryCore, Key: "LABEL_REPORT", Type: TypeBool, Default: "false",
			Description: "Mostra o nome original de cada label como comentário no código convertido."}, // label_report
		{Category: CategoryCore, Key: "LINE_REPORT", Type: TypeBool, Default: "false",
			Description: "Gera o relatório de correspondência entre linha Dignified e linha Classic."}, // line_report
		{Category: CategoryCore, Key: "VARIABLE_REPORT", Type: TypeBool, Default: "false",
			Description: "Gera o relatório de substituição de nome de variável."}, // var_report
		{Category: CategoryCore, Key: "LEXER_REPORT", Type: TypeBool, Default: "false",
			Description: "Grava a saída bruta do lexer pra depuração."}, // lexer_report
		{Category: CategoryCore, Key: "PARSER_REPORT", Type: TypeBool, Default: "false",
			Description: "Grava a saída bruta do parser pra depuração."}, // parser_report

		// --- msx (msx/badig_msx.py + tokenizer_interface.py + emulator_interface.py) ---
		{Category: CategoryMSX, Key: "CONVERT_PRINT", Type: TypeEnum, Default: "NONE",
			Description: "Converte PRINT<->? na saída.",
			Enum:        []string{"NONE", "TO_QUESTION_MARK", "TO_PRINT"}}, // convert_print
		{Category: CategoryMSX, Key: "STRIP_THEN_GOTO", Type: TypeEnum, Default: "NONE",
			Description: "Remove um THEN ou GOTO adjacente redundante.",
			Enum:        []string{"NONE", "THEN", "GOTO"}}, // strip_then_goto
		{Category: CategoryMSX, Key: "TOKENIZE", Type: TypeBool, Default: "false",
			Description: "Gera o binário .bmx tokenizado além da listagem Classic."}, // tokenize
		{Category: CategoryMSX, Key: "TOKENIZE_LIST_WIDTH", Type: TypeInt, Default: "0",
			Description: "Bytes por linha no arquivo de listagem .lmx (0 desliga a listagem).", Min: new(0)}, // list
		{Category: CategoryMSX, Key: "DELETE_ASCII_AFTER_TOKENIZE", Type: TypeBool, Default: "false",
			Description: "Apaga o .amx depois de gerar o .bmx tokenizado."}, // del_ascii
		{Category: CategoryMSX, Key: "TOKENIZER_VERBOSE", Type: TypeInt, Default: "3",
			Description: "Nível de detalhe das mensagens do tokenizer.", Min: new(0), Max: new(5)}, // verbose (tokenizer)
		{Category: CategoryMSX, Key: "EMULATOR_RUN", Type: TypeBool, Default: "false",
			Description: "Abre o emulador openMSX depois de converter."}, // run
		{Category: CategoryMSX, Key: "EMULATOR_SETTING_FILE", Type: TypeString, Default: "",
			Description: "Arquivo de configuração do openMSX a usar."}, // setting
		{Category: CategoryMSX, Key: "EMULATOR_MACHINE", Type: TypeString, Default: "",
			Description: "Máquina MSX a emular (ex. Philips_NMS_8250)."}, // machine
		{Category: CategoryMSX, Key: "EMULATOR_EXTENSION", Type: TypeString, Default: "",
			Description: "Extensão de disco e slot a montar (formato nome:slot)."}, // extension
		{Category: CategoryMSX, Key: "EMULATOR_NO_THROTTLE", Type: TypeBool, Default: "false",
			Description: "Desliga o throttle (roda o mais rápido possível, sem respeitar velocidade real)."}, // nothrottle
		{Category: CategoryMSX, Key: "EMULATOR_MONITOR", Type: TypeBool, Default: "true",
			Description: "Monitora a execução no emulador e reporta erro/linha."}, // monitor
		{Category: CategoryMSX, Key: "EMULATOR_VERBOSE", Type: TypeInt, Default: "3",
			Description: "Nível de detalhe das mensagens do emulador.", Min: new(0), Max: new(4)}, // verbose (emulator)
		{Category: CategoryMSX, Key: "EMULATOR_PATH", Type: TypeString, Default: "",
			Description: "Caminho do executável do openMSX."}, // emulator_path

		// --- zx81 (zx81/tokenizer_interface.py + emulator_interface.py; badig_zx81.py has no dialect-specific options of its own) ---
		{Category: CategoryZX81, Key: "TOKENIZE", Type: TypeBool, Default: "true",
			Description: "Gera o binário .p tokenizado além da listagem Classic."}, // tokenize (default IS true for ZX81, unlike MSX)
		{Category: CategoryZX81, Key: "DELETE_ASCII_AFTER_TOKENIZE", Type: TypeBool, Default: "false",
			Description: "Apaga o arquivo ASCII depois de gerar o .p tokenizado."}, // del_ascii
		{Category: CategoryZX81, Key: "EMULATOR_RUN", Type: TypeBool, Default: "false",
			Description: "Abre o emulador EightyOne depois de converter."}, // run
		{Category: CategoryZX81, Key: "EMULATOR_PATH", Type: TypeString, Default: `C:\zeddy\EightyOne.exe`,
			Description: "Caminho do executável do EightyOne."}, // emulator_path

		// --- zxspectrum (zxspectrum/badig_zxspectrum.py + tokenizer_interface.py + emulator_interface.py) ---
		{Category: CategoryZXSpectrum, Key: "SPECTRUM_MODEL", Type: TypeEnum, Default: "48",
			Description: "Modelo de ROM alvo.", Enum: []string{"48", "128"}}, // sp_model
		{Category: CategoryZXSpectrum, Key: "TOKENIZE", Type: TypeBool, Default: "false",
			Description: "Gera a fita tokenizada além da listagem Classic."}, // tokenize (Python itself has no clean default here - chosen defensively)
		{Category: CategoryZXSpectrum, Key: "TAPE_NAME", Type: TypeString, Default: "",
			Description: "Nome gravado no cabeçalho da fita (até 10 caracteres ASCII)."}, // name
		{Category: CategoryZXSpectrum, Key: "TAPE_FORMAT", Type: TypeEnum, Default: "TAP",
			Description: "Formato do arquivo de fita/disco gerado.",
			Enum:        []string{"TAP", "TZX", "BOTH", "Z80", "SNA", "ALL"}}, // format
		{Category: CategoryZXSpectrum, Key: "AUTORUN_LINE", Type: TypeInt, Default: "-1",
			Description: "Linha de autorun gravada na fita (-1 desliga).", Min: new(-1), Max: new(9999)}, // autorun
		{Category: CategoryZXSpectrum, Key: "EMULATOR_RUN", Type: TypeBool, Default: "false",
			Description: "Abre o emulador EightyOne depois de converter."}, // run
		{Category: CategoryZXSpectrum, Key: "EMULATOR_PATH", Type: TypeString, Default: "",
			Description: "Caminho do executável do EightyOne."}, // path/emulator_path
		{Category: CategoryZXSpectrum, Key: "EMULATOR_SETTING_FILE", Type: TypeString, Default: "",
			Description: "Arquivo de configuração do emulador a usar."}, // setting
	}
}
