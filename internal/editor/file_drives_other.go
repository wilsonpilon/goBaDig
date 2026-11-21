//go:build !windows

package edit

func editorDrives() ([]string, error) {
	return []string{"/"}, nil
}
