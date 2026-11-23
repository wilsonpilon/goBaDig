// Package dsk reads and writes MSX-DOS FAT12 floppy disk images
// (.DSK): a raw sector dump of a 360KB/720KB disk formatted with the
// classic MS-DOS 2.0-style BPB/FAT12 filesystem MSX-DOS itself uses.
//
// Ported from E:\msxdiskutil\MSXDisk.pbi (the user's own, already
// working PureBasic reimplementation of Arnold Metselaar's original C
// rddsk/wrdsk tools - confirmed by grepping badig/ and this project's
// own internal/ that NEITHER had ever implemented any .DSK read/write
// before this; docs/RESOURCES.md:161-168 and OUTLINE.md:2315-2316
// both already named this as the pending, not-yet-started piece).
// Ported the real, tested reference instead of only the general
// FAT12/MSX-DOS specification researched separately - same principle
// used for badig.py throughout this whole project.
//
// Deliberate structural change from the original: PureBasic's own
// module keeps its state (open file, boot block, FAT/dir buffers) in
// package-level globals - only one disk open at a time, process-wide.
// This package instead returns an independent *Image per Open/Create
// call, with zero package-level state - not just an idiom choice, a
// real requirement: the planned two-pane file manager (ver OUTLINE.md)
// needs each panel able to hold its own open .DSK at the same time.
// Likewise, every method returns a real `error` instead of a bool
// alongside a global GetLastErrorMessage() - same reasoning, applied
// throughout this whole project already.
package dsk

import (
	"fmt"
	"os"
)

const sectorSize = 512

// FileInfo describes one file found by ListFiles.
type FileInfo struct {
	// Name is the display name, lower-cased (e.g. "autoexec.bas") -
	// mirrors GetEntryName's own LCase(name$+"."+ext$): MSX-DOS
	// stores names upper-case on disk but every real tool (including
	// this one's own reference) displays them lower-case.
	Name         string
	Size         uint32
	Attrib       byte
	FirstCluster uint16
	// MSXTime/MSXDate are the raw on-disk MS-DOS timestamp fields
	// (hour<<11|min<<5|sec/2 and (year-1980)<<9|month<<5|day) - kept
	// raw rather than converted to time.Time here, since callers that
	// just want to display or round-trip them (the planned file
	// manager) don't need a timezone-less time.Time: ToTime()/
	// FromTime() below do the conversion when it's actually wanted.
	MSXTime uint16
	MSXDate uint16
}

// Image is one open MSX-DOS disk image, created by Create or Open.
// Every field here was a package-level Global in the original -
// moved onto this struct so a caller can hold more than one Image
// open at once (ver this package's own doc comment).
type Image struct {
	file *os.File
	path string

	bootBlock [sectorSize]byte
	fat       []byte
	dir       []byte

	// Geometry, all derived from bootBlock by parseBootBlock -
	// mirrors ParseBootBlock's own local variables exactly.
	sectorsPerCluster int
	clusterLen        int // bytes per cluster = sectorSize * sectorsPerCluster
	numDirEntries     int
	numFATs           int
	sectorsPerFAT     int
	dirOffset         int64 // byte offset of the root directory
	dataOffset        int64 // byte offset of cluster 2's own data
	totalSectors      int
	maxCluster        int
	fatLen            int
	dirLen            int
}

// parseBootBlock derives this Image's geometry from its own 512-byte
// bootBlock - mirrors ParseBootBlock(), including its exact
// operator-precedence-sensitive maxcl formula.
func (img *Image) parseBootBlock() error {
	b := &img.bootBlock
	bytesPerSector := int(b[11]) | int(b[12])<<8
	sectorsPerCluster := int(b[13])
	if bytesPerSector != sectorSize || sectorsPerCluster == 0 {
		return fmt.Errorf("dsk: boot block does not contain valid sector/cluster parameters")
	}

	img.sectorsPerCluster = sectorsPerCluster
	img.clusterLen = sectorSize * sectorsPerCluster
	img.numDirEntries = int(b[17]) | int(b[18])<<8
	img.numFATs = int(b[16])
	img.sectorsPerFAT = int(b[22]) | int(b[23])<<8
	img.totalSectors = int(b[19]) | int(b[20])<<8

	img.dirOffset = int64(sectorSize * (1 + img.numFATs*img.sectorsPerFAT))
	img.dataOffset = img.dirOffset + int64(img.numDirEntries*32)

	// mirrors `maxcl = (total_sectors - data_ofs/seclen) / cl + 1`:
	// PureBasic's own operator precedence (/ before -, before the
	// trailing + 1) matches Go's, so this reads identically.
	img.maxCluster = (img.totalSectors-int(img.dataOffset)/sectorSize)/sectorsPerCluster + 1

	img.fatLen = sectorSize * img.sectorsPerFAT
	img.dirLen = img.numDirEntries * 32
	return nil
}

// Open reads an existing .DSK image from path.
func Open(path string) (*Image, error) {
	f, err := os.OpenFile(path, os.O_RDWR, 0)
	if err != nil {
		return nil, fmt.Errorf("dsk: open %s: %w", path, err)
	}

	img := &Image{file: f, path: path}
	if _, err := f.ReadAt(img.bootBlock[:], 0); err != nil {
		f.Close()
		return nil, fmt.Errorf("dsk: read boot sector from %s: %w", path, err)
	}
	if err := img.parseBootBlock(); err != nil {
		f.Close()
		return nil, err
	}

	info, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, fmt.Errorf("dsk: stat %s: %w", path, err)
	}
	expectedSize := int64(img.totalSectors) * sectorSize
	if info.Size() != expectedSize {
		f.Close()
		return nil, fmt.Errorf("dsk: %s size (%d) does not match boot sector's expected size (%d)", path, info.Size(), expectedSize)
	}

	img.fat = make([]byte, img.fatLen)
	if _, err := f.ReadAt(img.fat, sectorSize); err != nil {
		f.Close()
		return nil, fmt.Errorf("dsk: read FAT from %s: %w", path, err)
	}
	img.dir = make([]byte, img.dirLen)
	if _, err := f.ReadAt(img.dir, img.dirOffset); err != nil {
		f.Close()
		return nil, fmt.Errorf("dsk: read directory from %s: %w", path, err)
	}

	return img, nil
}

// Create formats a new, blank .DSK image at path. bootSector, if
// non-nil, must be exactly 512 bytes and replaces the embedded
// default (ver defaultBootSector) - same optional-custom-boot-sector
// parameter CreateDisk's own BootSectorPath$ argument supports.
//
// Deliberate improvement, documented: the original grows its file to
// its final size by seeking to the LAST sector and writing just that
// (relying on the OS/filesystem reading the resulting gap back as
// zero bytes - true on most modern filesystems, but a "sparse file"
// implementation detail, not a guaranteed cross-platform behavior).
// 720KB is tiny, so this writes the file's FULL zeroed contents
// explicitly instead - correct regardless of filesystem.
func Create(path string, bootSector []byte) (*Image, error) {
	var boot [sectorSize]byte
	if bootSector != nil {
		if len(bootSector) != sectorSize {
			return nil, fmt.Errorf("dsk: custom boot sector must be exactly %d bytes, got %d", sectorSize, len(bootSector))
		}
		copy(boot[:], bootSector)
	} else {
		copy(boot[:], defaultBootSector[:])
	}

	img := &Image{bootBlock: boot}
	if err := img.parseBootBlock(); err != nil {
		return nil, err
	}

	f, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("dsk: create %s: %w", path, err)
	}

	totalSize := int64(img.totalSectors) * sectorSize
	if err := f.Truncate(totalSize); err != nil {
		f.Close()
		return nil, fmt.Errorf("dsk: allocate %s to %d bytes: %w", path, totalSize, err)
	}
	zeros := make([]byte, totalSize)
	copy(zeros, img.bootBlock[:])
	if _, err := f.WriteAt(zeros, 0); err != nil {
		f.Close()
		return nil, fmt.Errorf("dsk: zero-fill %s: %w", path, err)
	}

	// Format the FAT: cluster 0 carries the media descriptor (low
	// byte) with its high nibble set to 1 (mirrors `mediaDescriptor |
	// $F00`), cluster 1 is a fixed end-of-chain marker - both
	// reserved, never allocated to a real file.
	img.fat = make([]byte, img.fatLen)
	mediaDescriptor := uint16(img.bootBlock[21])
	writeFAT12(img.fat, 0, mediaDescriptor|0xF00)
	writeFAT12(img.fat, 1, 0xFFF)
	img.dir = make([]byte, img.dirLen)

	if _, err := f.WriteAt(img.bootBlock[:], 0); err != nil {
		f.Close()
		return nil, fmt.Errorf("dsk: write boot sector to %s: %w", path, err)
	}
	offset := int64(sectorSize)
	for i := 0; i < img.numFATs; i++ {
		if _, err := f.WriteAt(img.fat, offset); err != nil {
			f.Close()
			return nil, fmt.Errorf("dsk: write FAT copy %d to %s: %w", i+1, path, err)
		}
		offset += int64(img.fatLen)
	}
	if _, err := f.WriteAt(img.dir, img.dirOffset); err != nil {
		f.Close()
		return nil, fmt.Errorf("dsk: write directory to %s: %w", path, err)
	}

	img.file = f
	img.path = path
	return img, nil
}

// Close flushes the FAT (every copy) and the directory back to disk
// and releases the underlying file handle.
func (img *Image) Close() error {
	if img.file == nil {
		return nil
	}
	offset := int64(sectorSize)
	for i := 0; i < img.numFATs; i++ {
		if _, err := img.file.WriteAt(img.fat, offset); err != nil {
			return fmt.Errorf("dsk: write FAT copy %d to %s: %w", i+1, img.path, err)
		}
		offset += int64(img.fatLen)
	}
	if _, err := img.file.WriteAt(img.dir, img.dirOffset); err != nil {
		return fmt.Errorf("dsk: write directory to %s: %w", img.path, err)
	}
	err := img.file.Close()
	img.file = nil
	if err != nil {
		return fmt.Errorf("dsk: close %s: %w", img.path, err)
	}
	return nil
}
