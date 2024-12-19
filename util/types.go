package util

// Represents a generic database row scanner.
// Useful for scanning multiple and single rowed resultsets.
type RowScanner interface {
	Scan(dest ...any) error
}
