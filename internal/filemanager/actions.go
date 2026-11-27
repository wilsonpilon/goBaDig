package filemanager

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/wilsonpilon/goBaDig/internal/dialect"
	"github.com/wilsonpilon/goBaDig/internal/platforms/msx"
)

// compileSelected dispatches by extension: the three Dignified
// dialects go through dialect.CompileByExtension, and ".asm" goes
// through ASMSX instead (Fase 4 of the autocomplete+man/help+ASMSX
// plan, ver OUTLINE.md) - not part of dialect.CompileByExtension's
// own switch, since ASMSX is an MSX-specific tool, not a Dignified
// dialect.
func (fm *FileManager) compileSelected() {
	path, ok := fm.current().selectedPath()
	if !ok {
		fm.setStatus(" selecione um arquivo pra compilar")
		return
	}
	if strings.EqualFold(filepath.Ext(path), ".asm") {
		fm.compileAsmSelected(path)
		return
	}
	classicPath, binaryPath, err := dialect.CompileByExtension(path)
	if err != nil {
		fm.setStatus(fmt.Sprintf(" erro ao compilar: %v", err))
		return
	}
	_ = fm.current().reload()
	fm.setStatus(fmt.Sprintf(" compilado: %s + %s", filepath.Base(classicPath), filepath.Base(binaryPath)))
}

// compileAsmSelected resolves the asMSX executable (internal/options'
// ASMSX_PATH, via the file manager's own settings store - ver
// internal/platforms/msx's ResolveAsmsxPath) and compiles path with it.
func (fm *FileManager) compileAsmSelected(path string) {
	asmsxExe, err := msx.ResolveAsmsxPath(fm.store)
	if err != nil {
		fm.setStatus(fmt.Sprintf(" erro ao compilar: %v", err))
		return
	}
	binPath, _, err := msx.CompileAsm(path, asmsxExe)
	if err != nil {
		fm.setStatus(fmt.Sprintf(" erro ao compilar: %v", err))
		return
	}
	_ = fm.current().reload()
	fm.setStatus(fmt.Sprintf(" compilado: %s", filepath.Base(binPath)))
}

// executeSelected uses the platform's default emulator (empty
// backend name) - a per-project default emulator is explicitly out
// of scope for this round (see the plan's "Fora de escopo").
func (fm *FileManager) executeSelected() {
	path, ok := fm.current().selectedPath()
	if !ok {
		fm.setStatus(" selecione um arquivo pra executar")
		return
	}
	artifactPath, err := dialect.RunInEmulatorByExtension(path, "")
	if err != nil {
		fm.setStatus(fmt.Sprintf(" erro ao executar: %v", err))
		return
	}
	_ = fm.current().reload()
	fm.setStatus(fmt.Sprintf(" executando %s no emulador", filepath.Base(artifactPath)))
}

func (fm *FileManager) editSelected() {
	path, ok := fm.current().selectedPath()
	if !ok {
		fm.setStatus(" selecione um arquivo pra editar")
		return
	}
	fm.onOpenInEditor(path)
}

// copySelected copies the active panel's selected FILE into the
// other panel's current directory. Copying a directory would need
// recursion, which this round deliberately doesn't implement (see
// the plan's "Fora de escopo" - local<->local file ops only, kept
// simple); moveSelected below covers directories instead, since
// os.Rename needs no recursion.
func (fm *FileManager) copySelected() {
	e, ok := fm.current().selected()
	if !ok || e.isParent() {
		fm.setStatus(" selecione um arquivo pra copiar")
		return
	}
	if e.isDir {
		fm.setStatus(" copiar diretorios ainda nao e suportado (use mover)")
		return
	}
	src := filepath.Join(fm.current().dir, e.name)
	dest := filepath.Join(fm.other().dir, e.name)
	if err := copyFile(src, dest); err != nil {
		fm.setStatus(fmt.Sprintf(" erro ao copiar: %v", err))
		return
	}
	_ = fm.other().reload()
	fm.setStatus(fmt.Sprintf(" copiado para %s", dest))
}

func (fm *FileManager) moveSelected() {
	e, ok := fm.current().selected()
	if !ok || e.isParent() {
		fm.setStatus(" selecione algo pra mover")
		return
	}
	src := filepath.Join(fm.current().dir, e.name)
	dest := filepath.Join(fm.other().dir, e.name)
	fm.confirm(fmt.Sprintf("Mover \"%s\" para\n%s ?", e.name, fm.other().dir), func() {
		if err := os.Rename(src, dest); err != nil {
			fm.setStatus(fmt.Sprintf(" erro ao mover: %v", err))
			return
		}
		_ = fm.current().reload()
		_ = fm.other().reload()
		fm.setStatus(" movido")
	})
}

func (fm *FileManager) promptDelete() {
	e, ok := fm.current().selected()
	if !ok || e.isParent() {
		fm.setStatus(" selecione algo pra apagar")
		return
	}
	path := filepath.Join(fm.current().dir, e.name)
	fm.confirm(fmt.Sprintf("Apagar \"%s\" ?", e.name), func() {
		// os.Remove refuses a non-empty directory on its own, which
		// is exactly the "no recursive delete" restriction this
		// round wants - no extra check needed here.
		if err := os.Remove(path); err != nil {
			fm.setStatus(fmt.Sprintf(" erro ao apagar: %v", err))
			return
		}
		_ = fm.current().reload()
		fm.setStatus(" apagado")
	})
}

func (fm *FileManager) mkdir(name string) {
	if name == "" {
		return
	}
	path := filepath.Join(fm.current().dir, name)
	if err := os.Mkdir(path, 0o755); err != nil {
		fm.setStatus(fmt.Sprintf(" erro ao criar diretorio: %v", err))
		return
	}
	_ = fm.current().reload()
	fm.setStatus(fmt.Sprintf(" diretorio criado: %s", name))
}

func copyFile(src, dest string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Close()
}
