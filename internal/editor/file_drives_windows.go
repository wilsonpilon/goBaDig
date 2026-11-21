package edit

import (
	"fmt"

	"golang.org/x/sys/windows"
)

func editorDrives() ([]string, error) {
	mask, err := windows.GetLogicalDrives()
	if err != nil {
		return nil, err
	}
	var drives []string
	for index := 0; index < 26; index++ {
		if mask&(1<<index) != 0 {
			drives = append(drives, fmt.Sprintf("%c:\\", 'A'+index))
		}
	}
	return drives, nil
}
