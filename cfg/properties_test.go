package cfg

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

func TestUpdate(t *testing.T) {
	input := `
# asdfsd
# a=b
# sdfasdf
c=d
noupdate=a
`

	excepted := "\r\n" +
		"# asdfsd\r\n" +
		"a=1\r\n" +
		"# sdfasdf\r\n" +
		"c=2\r\n" +
		"noupdate=a\r\n" +
		"notExists=3\r\n"

	var buf bytes.Buffer

	e := UpdateWith(strings.NewReader(input), &buf, map[string]string{"a": "1", "c": "2", "notExists": "3"})
	if e != nil {
		t.Error(e)
		return
	}

	if excepted != buf.String() {
		fmt.Println("actual   =============")
		fmt.Println(buf.String())
		fmt.Println("excepted =============")
		fmt.Println(excepted)
		t.Error(buf.String())
	}
}
