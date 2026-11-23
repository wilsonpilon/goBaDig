package dsk

import (
	"strings"
	"time"
)

// decodeName decodes an 8-byte name field and a 3-byte extension
// field (the raw on-disk layout) into a lower-case display string -
// mirrors GetEntryName exactly: each field stops at its first 0x00
// or space byte (not a trailing-space trim, though the two coincide
// for every legal FAT11 name actually encountered in practice).
func decodeName(name [8]byte, ext [3]byte) string {
	var b strings.Builder
	for _, c := range name {
		if c == 0 || c == ' ' {
			break
		}
		b.WriteByte(c)
	}
	var e strings.Builder
	for _, c := range ext {
		if c == 0 || c == ' ' {
			break
		}
		e.WriteByte(c)
	}
	if e.Len() > 0 {
		return strings.ToLower(b.String() + "." + e.String())
	}
	return strings.ToLower(b.String())
}

// ConvertToFAT11 converts a display name (e.g. "file.txt", or a
// wildcard mask like "*.bas") into its padded, upper-case, 11-byte
// FAT directory representation ("FILE    TXT"). Mirrors
// ConvertToFAT11 field-by-field, including its two non-obvious
// quirks: a "." only actually advances past itself once the name
// field has been padded out to all 8 characters (so "FILE.TXT"'s
// dot gets "consumed" only after 4 padding spaces fill the rest of
// the name slot), and a "*" expands to fill the REST of its current
// 8.3 segment with "?" without consuming more than the one input
// character (classic DOS wildcard expansion).
func ConvertToFAT11(name string) string {
	res := make([]byte, 11)
	for i := range res {
		res[i] = ' '
	}
	runes := []rune(strings.ToUpper(name))
	k, i := 0, 0
	for k < 11 {
		var c rune
		hasChar := i < len(runes)
		if hasChar {
			c = runes[i]
		}
		switch {
		case hasChar && c == '.':
			if k < 8 {
				res[k] = ' '
				k++
			} else {
				i++
			}
		case hasChar && c == '*':
			res[k] = '?'
			k++
			if k == 8 {
				i++
			}
		case !hasChar:
			res[k] = ' '
			k++
		default:
			res[k] = byte(c)
			k++
			i++
		}
	}
	return string(res)
}

// MatchesFAT11 reports whether a FAT11-encoded entry name matches a
// FAT11-encoded mask (as produced by ConvertToFAT11), where a mask
// byte of '?' matches any entry byte. Mirrors MatchesFAT11.
func MatchesFAT11(entry, mask string) bool {
	for k := range 11 {
		if entry[k] != mask[k] && mask[k] != '?' {
			return false
		}
	}
	return true
}

// ToTime converts an MSX/MS-DOS packed time+date pair into a
// time.Time - mirrors MSXTimeToPBDate's own field extraction and
// out-of-range clamping (a corrupt/zero timestamp becomes
// 1980-01-01T00:00:00 rather than an invalid date).
func ToTime(msxTime, msxDate uint16) time.Time {
	day := int(msxDate & 0x1F)
	month := time.Month((msxDate >> 5) & 0x0F)
	year := 1980 + int((msxDate>>9)&0x7F)
	sec := int(msxTime&0x1F) * 2
	minute := int((msxTime >> 5) & 0x3F)
	hour := int((msxTime >> 11) & 0x1F)

	if month < 1 || month > 12 {
		month = 1
	}
	if day < 1 || day > 31 {
		day = 1
	}
	if hour > 23 {
		hour = 0
	}
	if minute > 59 {
		minute = 0
	}
	if sec > 59 {
		sec = 0
	}
	return time.Date(year, month, day, hour, minute, sec, 0, time.Local)
}

// FromTime converts a time.Time into its packed MSX/MS-DOS time and
// date fields - mirrors PBDateToMSXTime/PBDateToMSXDate.
func FromTime(t time.Time) (msxTime, msxDate uint16) {
	hour, minute, sec := t.Hour(), t.Minute(), t.Second()/2
	msxTime = uint16(hour<<11) | uint16(minute<<5) | uint16(sec)

	year := max(min(t.Year()-1980, 127), 0)
	msxDate = uint16(year<<9) | uint16(int(t.Month())<<5) | uint16(t.Day())
	return msxTime, msxDate
}
