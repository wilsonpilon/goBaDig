package edit

import (
	"errors"
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	winspool              = windows.NewLazySystemDLL("winspool.drv")
	getDefaultPrinterProc = winspool.NewProc("GetDefaultPrinterW")
	openPrinterProc       = winspool.NewProc("OpenPrinterW")
	closePrinterProc      = winspool.NewProc("ClosePrinter")
	startDocPrinterProc   = winspool.NewProc("StartDocPrinterW")
	startPagePrinterProc  = winspool.NewProc("StartPagePrinter")
	writePrinterProc      = winspool.NewProc("WritePrinter")
	endPagePrinterProc    = winspool.NewProc("EndPagePrinter")
	endDocPrinterProc     = winspool.NewProc("EndDocPrinter")
)

type printerDocumentInfo struct {
	name       *uint16
	outputFile *uint16
	dataType   *uint16
}

func sendToDefaultPrinter(stream []byte) (err error) {
	size := uint32(0)
	result, _, callErr := getDefaultPrinterProc.Call(0, uintptr(unsafe.Pointer(&size)))
	if result == 0 && size == 0 {
		return winspoolError("getting the default printer", callErr)
	}
	if size == 0 {
		return fmt.Errorf("getting the default printer returned an empty name")
	}
	printerName := make([]uint16, size)
	if err := winspoolCall("getting the default printer", getDefaultPrinterProc,
		uintptr(unsafe.Pointer(&printerName[0])), uintptr(unsafe.Pointer(&size))); err != nil {
		return err
	}

	printerNamePtr := &printerName[0]
	var printer windows.Handle
	if err := winspoolCall("opening the default printer", openPrinterProc,
		uintptr(unsafe.Pointer(printerNamePtr)), uintptr(unsafe.Pointer(&printer)), 0); err != nil {
		return err
	}
	defer func() {
		err = errors.Join(err, winspoolCall("closing the default printer", closePrinterProc, uintptr(printer)))
	}()

	documentName, err := windows.UTF16PtrFromString("WibEdit")
	if err != nil {
		return err
	}
	dataType, err := windows.UTF16PtrFromString("RAW")
	if err != nil {
		return err
	}
	documentInfo := printerDocumentInfo{name: documentName, dataType: dataType}
	if err := winspoolCall("starting the printer job", startDocPrinterProc,
		uintptr(printer), 1, uintptr(unsafe.Pointer(&documentInfo))); err != nil {
		return err
	}
	docStarted := true
	defer func() {
		if docStarted {
			err = errors.Join(err, winspoolCall("ending the printer job", endDocPrinterProc, uintptr(printer)))
		}
	}()

	if err := winspoolCall("starting the printer page", startPagePrinterProc, uintptr(printer)); err != nil {
		return err
	}
	pageStarted := true
	defer func() {
		if pageStarted {
			err = errors.Join(err, winspoolCall("ending the printer page", endPagePrinterProc, uintptr(printer)))
		}
	}()

	for offset := 0; offset < len(stream); {
		chunkSize := min(len(stream)-offset, int(^uint32(0)))
		written := uint32(0)
		if err := winspoolCall("writing to the printer", writePrinterProc,
			uintptr(printer), uintptr(unsafe.Pointer(&stream[offset])), uintptr(uint32(chunkSize)),
			uintptr(unsafe.Pointer(&written))); err != nil {
			return err
		}
		if written == 0 || int(written) > chunkSize {
			return fmt.Errorf("writing to the printer returned an invalid byte count: %d", written)
		}
		offset += int(written)
	}

	if err := winspoolCall("ending the printer page", endPagePrinterProc, uintptr(printer)); err != nil {
		return err
	}
	pageStarted = false
	if err := winspoolCall("ending the printer job", endDocPrinterProc, uintptr(printer)); err != nil {
		return err
	}
	docStarted = false
	return nil
}

func winspoolCall(name string, proc *windows.LazyProc, arguments ...uintptr) error {
	result, _, callErr := proc.Call(arguments...)
	if result != 0 {
		return nil
	}
	return winspoolError(name, callErr)
}

func winspoolError(name string, callErr error) error {
	if callErr == nil || callErr == windows.ERROR_SUCCESS {
		callErr = windows.GetLastError()
	}
	if callErr == nil || callErr == windows.ERROR_SUCCESS {
		return fmt.Errorf("%s failed without an operating-system error", name)
	}
	return fmt.Errorf("%s: %w", name, callErr)
}
