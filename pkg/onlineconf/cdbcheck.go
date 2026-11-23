package onlineconf

import (
	"encoding/binary"
	"fmt"
	"io"
)

const (
	// cdbTableCount is the number of hash tables of a CDB file. It is a part of the format
	// and is not configurable: a key is routed to a table by the low byte of its hash.
	cdbTableCount = 256

	// cdbIndexEntrySize is the size of a single entry of the index: (offset uint32,
	// length uint32) of one hash table.
	cdbIndexEntrySize = 8

	// cdbIndexSize is the size of the index at the head of a CDB file: one entry per hash
	// table.
	cdbIndexSize = cdbTableCount * cdbIndexEntrySize

	// cdbSlotSize is the size of a single slot of a hash table: (hash uint32, offset uint32).
	cdbSlotSize = 8
)

// validateCDB reports whether a file of the given size is usable as a CDB.
//
// It is a structural check of the index: every hash table must lie between the index and the
// end of the file. The CDB reader itself only reads the index when the file is opened and
// discovers everything else on a lookup, so without this check a file whose index survived
// but whose tables did not — the typical result of an interrupted or partially written
// rebuild — is accepted and then fails on every read.
//
// The two shapes of a broken file it recognizes:
//
//   - a table that ends past the end of the file: the tables are written after the data, so
//     a file truncated anywhere loses a part of them;
//   - a table that starts inside the index: a writer fills the index with zeroes and
//     rewrites it only after the data has been flushed, so a zeroed index means the file was
//     never finished.
//
// Validating the records themselves is deliberately left out: they precede the tables, so a
// truncated file always trips one of the checks above, and a reader never walks the records
// outside of a lookup.
func validateCDB(r io.ReaderAt, size int) error {
	if size < cdbIndexSize {
		return fmt.Errorf("%w: file is %d bytes long, the index alone takes %d", ErrInvalidCDB, size, cdbIndexSize)
	}

	index := make([]byte, cdbIndexSize)
	if _, err := r.ReadAt(index, 0); err != nil {
		return fmt.Errorf("%w: can't read the index: %w", ErrInvalidCDB, err)
	}

	for table := 0; table < cdbTableCount; table++ {
		pos := table * cdbIndexEntrySize
		offset := binary.LittleEndian.Uint32(index[pos : pos+4])
		length := binary.LittleEndian.Uint32(index[pos+4 : pos+8])

		if offset < cdbIndexSize {
			return fmt.Errorf("%w: hash table %d starts at %d, inside the %d byte index: the file is not complete",
				ErrInvalidCDB, table, offset, cdbIndexSize)
		}

		if end := uint64(offset) + uint64(length)*cdbSlotSize; end > uint64(size) {
			return fmt.Errorf("%w: hash table %d ends at %d, past the end of the %d byte file: the file is truncated",
				ErrInvalidCDB, table, end, size)
		}
	}

	return nil
}
