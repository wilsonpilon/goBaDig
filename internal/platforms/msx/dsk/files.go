package dsk

import (
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// ErrClosed is returned by every Image method once Close has been
// called on it.
var ErrClosed = errors.New("dsk: image is closed")

// entryUsed reports whether a 32-byte directory entry slice holds a
// real file - mirrors IsEntryUsed (0xE5 = deleted, 0x00 = never
// used).
func entryUsed(e []byte) bool {
	return e[0] != 0xE5 && e[0] != 0x00
}

// findEntry locates the directory entry for msxName, returning its
// index and its own 32-byte slice (a view into img.dir, so writes
// through it persist), or -1/nil if there is no such file.
func (img *Image) findEntry(msxName string) (int, []byte) {
	target := ConvertToFAT11(msxName)
	for i := 0; i < img.numDirEntries; i++ {
		e := img.dir[i*32 : i*32+32]
		if entryUsed(e) && string(e[0:11]) == target {
			return i, e
		}
	}
	return -1, nil
}

// freeChain walks a cluster chain in the FAT starting at cluster,
// zeroing every entry it visits - mirrors the cluster-freeing loop
// shared by DeleteMSXFile and AddFile's own "roll back on disk full"
// path.
func (img *Image) freeChain(cluster int) {
	for cluster >= 2 && cluster <= img.maxCluster {
		next := int(readFAT12(img.fat, cluster))
		writeFAT12(img.fat, cluster, 0)
		cluster = next
	}
}

// ListFiles returns every real file in the image's root directory,
// in on-disk order. Mirrors ListFiles.
func (img *Image) ListFiles() ([]FileInfo, error) {
	if img.file == nil {
		return nil, ErrClosed
	}
	var out []FileInfo
	for i := 0; i < img.numDirEntries; i++ {
		e := img.dir[i*32 : i*32+32]
		if !entryUsed(e) {
			continue
		}
		var name [8]byte
		var ext [3]byte
		copy(name[:], e[0:8])
		copy(ext[:], e[8:11])
		out = append(out, FileInfo{
			Name:         decodeName(name, ext),
			Size:         binary.LittleEndian.Uint32(e[28:32]),
			Attrib:       e[11],
			FirstCluster: binary.LittleEndian.Uint16(e[26:28]),
			MSXTime:      binary.LittleEndian.Uint16(e[22:24]),
			MSXDate:      binary.LittleEndian.Uint16(e[24:26]),
		})
	}
	return out, nil
}

// ExtractFile copies msxName out of the image into destPath on the
// local filesystem, setting destPath's modification time from the
// image's own stored timestamp. Mirrors ExtractFile.
//
// Deliberate correction, documented: the original sets a warning
// message ("File was truncated") when the cluster chain runs out
// before the declared size is reached, but still returns success -
// and its own CLI (msxdisk.pb) never checks that message on a
// success return, making the warning dead code in practice. This
// case is corrupt-image territory (the size field and the real
// chain length disagree), so this port returns a real error instead
// of silently reporting success.
func (img *Image) ExtractFile(msxName, destPath string) error {
	if img.file == nil {
		return ErrClosed
	}
	_, e := img.findEntry(msxName)
	if e == nil {
		return fmt.Errorf("dsk: file not found on disk image: %s", msxName)
	}

	size := int64(binary.LittleEndian.Uint32(e[28:32]))
	cluster := int(binary.LittleEndian.Uint16(e[26:28]))

	out, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("dsk: create local file %s: %w", destPath, err)
	}

	remaining := size
	for remaining > 0 && cluster >= 2 && cluster <= img.maxCluster {
		buf := make([]byte, img.clusterLen)
		offset := img.dataOffset + int64(img.clusterLen)*int64(cluster-2)
		if _, err := img.file.ReadAt(buf, offset); err != nil {
			out.Close()
			return fmt.Errorf("dsk: read cluster for %s: %w", msxName, err)
		}
		n := min(int64(len(buf)), remaining)
		if _, err := out.Write(buf[:n]); err != nil {
			out.Close()
			return fmt.Errorf("dsk: write local file %s: %w", destPath, err)
		}
		remaining -= n
		cluster = int(readFAT12(img.fat, cluster))
	}
	if err := out.Close(); err != nil {
		return fmt.Errorf("dsk: close local file %s: %w", destPath, err)
	}
	if remaining > 0 {
		return fmt.Errorf("dsk: file %s is truncated on the disk image (wrote %d of %d bytes) - the image may be corrupt",
			msxName, size-remaining, size)
	}

	modTime := ToTime(binary.LittleEndian.Uint16(e[22:24]), binary.LittleEndian.Uint16(e[24:26]))
	if err := os.Chtimes(destPath, modTime, modTime); err != nil {
		return fmt.Errorf("dsk: set modification time on %s: %w", destPath, err)
	}
	return nil
}

// AddFile copies localPath into the image under msxName (the local
// file's own base name if msxName is empty), overwriting any existing
// entry of the same name. Mirrors AddFile, including its unconditional
// delete-then-add overwrite semantics.
func (img *Image) AddFile(localPath, msxName string) error {
	if img.file == nil {
		return ErrClosed
	}
	info, err := os.Stat(localPath)
	if err != nil {
		return fmt.Errorf("dsk: stat local file %s: %w", localPath, err)
	}
	data, err := os.ReadFile(localPath)
	if err != nil {
		return fmt.Errorf("dsk: read local file %s: %w", localPath, err)
	}
	if msxName == "" {
		msxName = filepath.Base(localPath)
	}

	if err := img.DeleteFile(msxName); err != nil {
		return err
	}

	entryIdx := -1
	for i := 0; i < img.numDirEntries; i++ {
		if !entryUsed(img.dir[i*32 : i*32+32]) {
			entryIdx = i
			break
		}
	}
	if entryIdx == -1 {
		return fmt.Errorf("dsk: directory is full, cannot add %s", msxName)
	}

	size := len(data)
	bytesLeft := size
	pos := 0
	prevCluster, firstCluster := 0, 0

	for bytesLeft > 0 {
		curCluster := 0
		for c := 2; c <= img.maxCluster; c++ {
			if readFAT12(img.fat, c) == 0 {
				curCluster = c
				break
			}
		}
		if curCluster == 0 {
			if firstCluster != 0 {
				img.freeChain(firstCluster)
			}
			return fmt.Errorf("dsk: disk full, cannot write entire file %s", msxName)
		}
		if firstCluster == 0 {
			firstCluster = curCluster
		}
		writeFAT12(img.fat, curCluster, 0xFFF)
		if prevCluster != 0 {
			writeFAT12(img.fat, prevCluster, uint16(curCluster))
		}

		buf := make([]byte, img.clusterLen)
		n := copy(buf, data[pos:])
		pos += n
		bytesLeft -= n
		offset := img.dataOffset + int64(img.clusterLen)*int64(curCluster-2)
		if _, err := img.file.WriteAt(buf, offset); err != nil {
			return fmt.Errorf("dsk: write cluster for %s: %w", msxName, err)
		}
		prevCluster = curCluster
	}

	e := img.dir[entryIdx*32 : entryIdx*32+32]
	for i := range e {
		e[i] = 0
	}
	copy(e[0:11], ConvertToFAT11(msxName))
	msxTime, msxDate := FromTime(info.ModTime())
	binary.LittleEndian.PutUint16(e[22:24], msxTime)
	binary.LittleEndian.PutUint16(e[24:26], msxDate)
	binary.LittleEndian.PutUint16(e[26:28], uint16(firstCluster))
	binary.LittleEndian.PutUint32(e[28:32], uint32(size))
	return nil
}

// DeleteFile removes msxName from the image, freeing its cluster
// chain in the FAT and marking its directory entry deleted. Mirrors
// DeleteMSXFile - including its "not found is not an error" behavior
// (AddFile relies on this to implement overwrite-by-delete-then-add).
func (img *Image) DeleteFile(msxName string) error {
	if img.file == nil {
		return ErrClosed
	}
	_, e := img.findEntry(msxName)
	if e == nil {
		return nil
	}
	img.freeChain(int(binary.LittleEndian.Uint16(e[26:28])))
	e[0] = 0xE5
	return nil
}
