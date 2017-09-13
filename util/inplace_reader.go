package util

import "io"

// InplaceReader a inplace reader for bufio.Scanner
type InplaceReader int

func (p *InplaceReader) Read([]byte) (int, error) {
	if *p == 0 {
		return 0, io.EOF
	}
	ret := int(*p)
	*p = 0
	return ret, io.EOF
}
