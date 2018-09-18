package util

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

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

func SplitLines(bs []byte) [][]byte {
	if len(bs) == 0 {
		return nil
	}

	r := InplaceReader(len(bs))
	scanner := bufio.NewScanner(&r)
	scanner.Buffer(bs, len(bs))

	lines := make([][]byte, 0, 10)
	for scanner.Scan() {
		lines = append(lines, scanner.Bytes())
	}

	if nil != scanner.Err() {
		panic(scanner.Err())
	}
	return lines
}

func NewScanner(bs []byte) *bufio.Scanner {
	r := InplaceReader(len(bs))
	scanner := bufio.NewScanner(&r)
	scanner.Buffer(bs, len(bs))
	return scanner
}

type UndoScanner struct {
	Scanner *bufio.Scanner
	IsUndo  bool
	Offset  int
	Error   error
}

func (s *UndoScanner) Undo() {
	if s.IsUndo {
		panic(errors.New("scanner is already undo"))
	}
	s.IsUndo = true
}

func (s *UndoScanner) Err() error {
	if s.Error != nil {
		return s.Error
	}
	return s.Scanner.Err()
}

func (s *UndoScanner) Bytes() []byte {
	return s.Scanner.Bytes()
}

func (s *UndoScanner) Text() string {
	return s.Scanner.Text()
}

func (s *UndoScanner) Scan() bool {
	if s.IsUndo {
		s.IsUndo = false
		return true
	}

	if s.Error != nil {
		return false
	}

	hasNext := s.Scanner.Scan()
	if hasNext {
		s.Offset += len(s.Scanner.Bytes())
	}
	return hasNext
}

func (s *UndoScanner) Buffer(buf []byte, max int) {
	s.Scanner.Buffer(buf, max)
}

func (s *UndoScanner) Split(split bufio.SplitFunc) {
	s.Scanner.Split(split)
}

type LineParser struct {
	scanner    UndoScanner
	lineNumber int
}

func NewLineParser(scanner *bufio.Scanner) *LineParser {
	return &LineParser{scanner: UndoScanner{Scanner: scanner}}
}

func (p *LineParser) Scan(skipEmptyLine ...bool) bool {
	if len(skipEmptyLine) > 0 && skipEmptyLine[0] {
		for p.scanner.Scan() {
			p.lineNumber++
			bs := p.scanner.Bytes()
			if len(bs) == 0 {
				continue
			}
			bs = bytes.TrimSpace(bs)
			if len(bs) != 0 {
				return true
			}
		}
		return false
	}

	hasNext := p.scanner.Scan()
	if hasNext {
		p.lineNumber++
	}
	return hasNext
}

func (s *LineParser) LineNumber() int {
	return s.lineNumber
}

func (s *LineParser) Err() error {
	return s.scanner.Err()
}

func (s *LineParser) SetError(err error) {
	s.scanner.Error = err
}

func (s *LineParser) Bytes() []byte {
	return s.scanner.Bytes()
}

func (s *LineParser) Text() string {
	return s.scanner.Text()
}

func (s *LineParser) NewParser(value map[string]interface{}) *Parser {
	p := NewParser(NewScanner(s.Bytes()), value)
	p.LineNumber = s.lineNumber
	return p
}

type Parser struct {
	LineNumber int
	Scanner    UndoScanner
	value      map[string]interface{}
}

func NewParser(scanner *bufio.Scanner, value map[string]interface{}) *Parser {
	return &Parser{Scanner: UndoScanner{Scanner: scanner},
		value: value}
}

func (s *Parser) undo() *Parser {
	s.Scanner.Undo()
	return s
}

func (s *Parser) To(value map[string]interface{}) *Parser {
	s.value = value
	return s
}

func (s *Parser) Except(word []byte) *Parser {
	if !s.Scanner.Scan() {
		if err := s.Scanner.Err(); err == nil || err == io.EOF {
			s.Scanner.Error = &ErrExcept{LineNumber: s.LineNumber, Offset: s.Scanner.Offset, Excepted: word, Err: io.EOF}
		}
		return s
	}
	if !bytes.Equal(word, bytes.ToLower(s.Scanner.Bytes())) {
		s.Scanner.Error = &ErrExcept{LineNumber: s.LineNumber, Offset: s.Scanner.Offset, Excepted: word, Actual: s.Scanner.Bytes()}
	}
	return s
}

func (s *Parser) ExceptAll(words ...[]byte) *Parser {
	for _, word := range words {
		if !s.Scanner.Scan() {
			if err := s.Scanner.Err(); err == nil || err == io.EOF {
				s.Scanner.Error = &ErrExcept{LineNumber: s.LineNumber, Offset: s.Scanner.Offset, Excepted: word, Err: io.EOF}
			}
			return s
		}
		if !bytes.Equal(word, bytes.ToLower(s.Scanner.Bytes())) {
			s.Scanner.Error = &ErrExcept{LineNumber: s.LineNumber, Offset: s.Scanner.Offset, Excepted: word, Actual: s.Scanner.Bytes()}
		}
	}
	return s
}

func (s *Parser) ExceptOptional(values ...interface{}) *Parser {
	for idx, value := range values {
		if !s.Scanner.Scan() {
			if err := s.Scanner.Err(); err == nil || err == io.EOF {
				s.Scanner.Error = &ErrExcept{LineNumber: s.LineNumber, Offset: s.Scanner.Offset, Excepted: []byte(fmt.Sprint(value)), Err: io.EOF}
			}
			return s
		}

		switch v := value.(type) {
		case []byte:
			if !bytes.Equal(v, bytes.ToLower(s.Scanner.Bytes())) {
				if idx != 0 {
					s.Scanner.Error = &ErrExcept{LineNumber: s.LineNumber, Offset: s.Scanner.Offset, Excepted: v, Actual: s.Scanner.Bytes()}
					return s
				}

				s.undo()
				return s
			}
		case string:
			if !bytes.Equal([]byte(v), bytes.ToLower(s.Scanner.Bytes())) {
				if idx != 0 {
					s.Scanner.Error = &ErrExcept{LineNumber: s.LineNumber, Offset: s.Scanner.Offset, Excepted: []byte(v), Actual: s.Scanner.Bytes()}
					return s
				}

				s.undo()
				return s
			}
		case func([]byte) (string, interface{}, error):
			key, kvalue, err := v(s.Scanner.Bytes())
			if err != nil {
				if idx != 0 {
					s.Scanner.Error = err // &ErrExcept{LineNumber: s.LineNumber, Excepted: "call", Actual: s.Scanner.Bytes(), Offset: s.Scanner.Offset}
					return s
				}
				s.undo()
				return s
			}

			if key != "" {
				s.value[key] = kvalue
			}

		case func(*Parser, []byte) (string, interface{}, error):
			key, kvalue, err := v(s, s.Scanner.Bytes())
			if err != nil {
				if idx != 0 {
					s.Scanner.Error = err // &ErrExcept{LineNumber: s.LineNumber, Excepted: "call", Actual: s.Scanner.Bytes(), Offset: s.Scanner.Offset}
					return s
				}
				s.undo()
				return s
			}

			if key != "" {
				s.value[key] = kvalue
			}
		default:
			s.Scanner.Error = errors.New("unsupported type")
			return s
		}
	}
	return s
}

func (s *Parser) ExceptAny(words ...[]byte) *Parser {
	if !s.Scanner.Scan() {
		if err := s.Scanner.Err(); err == nil || err == io.EOF {
			s.Scanner.Error = &ErrExcept{LineNumber: s.LineNumber,
				Offset:   s.Scanner.Offset,
				Excepted: []byte("\"" + string(bytes.Join(words, []byte("\", \""))) + "\""),
				Err:      io.EOF}
		}
		return s
	}
	for _, word := range words {
		if bytes.Equal(word, bytes.ToLower(s.Scanner.Bytes())) {
			return s
		}
	}

	s.Scanner.Error = &ErrExcept{LineNumber: s.LineNumber,
		Offset:   s.Scanner.Offset,
		Excepted: []byte("\"" + string(bytes.Join(words, []byte("\", \""))) + "\"")}
	return s
}

func (s *Parser) ExceptDate(nm string, layout string, layouts ...string) *Parser {
	if !s.Scanner.Scan() {
		if err := s.Scanner.Err(); err == nil || err == io.EOF {
			s.Scanner.Error = &ErrExcept{LineNumber: s.LineNumber,
				Offset:   s.Scanner.Offset,
				Excepted: []byte("field[" + nm + "] "),
				Err:      io.EOF}
		}
		return s
	}
	str := s.Scanner.Text()

	t, e := time.Parse(layout, str)
	if e == nil {
		s.value[nm] = t
		return s
	}

	for _, layout := range layouts {
		t, err := time.Parse(layout, str)
		if err == nil {
			s.value[nm] = t
			return s
		}
	}

	s.Scanner.Error = &ErrExcept{LineNumber: s.LineNumber,
		Offset:   s.Scanner.Offset,
		Excepted: []byte("field[" + nm + "] "),
		Err:      e}
	return s
}

func (s *Parser) ExceptInt(nm string) *Parser {
	if !s.Scanner.Scan() {
		if err := s.Scanner.Err(); err == nil || err == io.EOF {
			s.Scanner.Error = &ErrExcept{LineNumber: s.LineNumber,
				Offset:   s.Scanner.Offset,
				Excepted: []byte("field[" + nm + "] "),
				Err:      io.EOF}
		}
		return s
	}
	if i, e := strconv.ParseInt(string(s.Scanner.Bytes()), 10, 64); nil != e {
		s.Scanner.Error = &ErrExcept{LineNumber: s.LineNumber,
			Offset:   s.Scanner.Offset,
			Excepted: []byte("field[" + nm + "] "),
			Err:      e}
	} else {
		s.value[nm] = i
	}

	return s
}

func (s *Parser) ExceptFloat(nm string) *Parser {
	if !s.Scanner.Scan() {
		if err := s.Scanner.Err(); err == nil || err == io.EOF {
			s.Scanner.Error = &ErrExcept{LineNumber: s.LineNumber,
				Offset:   s.Scanner.Offset,
				Excepted: []byte("field[" + nm + "] "),
				Err:      io.EOF}
		}
		return s
	}
	if f, e := strconv.ParseFloat(string(s.Scanner.Bytes()), 10); nil != e {
		s.Scanner.Error = &ErrExcept{LineNumber: s.LineNumber,
			Offset:   s.Scanner.Offset,
			Excepted: []byte("field[" + nm + "] "),
			Err:      e}
	} else {
		s.value[nm] = f
	}

	return s
}

func TimeUnit(s *Parser, value interface{}) (interface{}, error) {
	if !s.Scanner.Scan() {
		return nil, s.Err()
	}
	return parseWithTimeUnit(s.Scanner.Text(), value)
}

func parseWithTimeUnit(s string, value interface{}) (interface{}, error) {
	unit := int64(0)
	switch strings.ToLower(s) {
	case "ns":
	case "us":
		unit = int64(time.Microsecond)
	case "ms":
		unit = int64(time.Millisecond)
	case "s":
		unit = int64(time.Second)
	case "m":
		unit = int64(time.Minute)
	case "h":
		unit = int64(time.Hour)
	default:
		return nil, fmt.Errorf("unknown unit - %v ", s)
	}

	switch v := value.(type) {
	case int64:
		return v * unit, nil
	case uint64:
		return v * uint64(unit), nil
	case float64:
		return uint64(v * float64(unit)), nil
	default:
		return nil, fmt.Errorf("unknown type - [%T] %v ", value, value)
	}
}

func BytesUnit(s *Parser, value interface{}) (interface{}, error) {
	if !s.Scanner.Scan() {
		return nil, s.Err()
	}
	unit := int64(0)
	switch strings.ToLower(s.Scanner.Text()) {
	case "b":
	case "kb":
		unit = 1024
	case "mb":
		unit = 1024 * 1024
	case "gb":
		unit = 1024 * 1024 * 1024
	default:
		return nil, fmt.Errorf("unknown unit - %v ", s.Scanner.Text())
	}

	switch v := value.(type) {
	case int64:
		return v * unit, nil
	case uint64:
		return v * uint64(unit), nil
	case float64:
		return v * float64(unit), nil
	default:
		return nil, fmt.Errorf("unknown type - [%T] %v ", value, value)
	}
}

func (s *Parser) ExceptIntWithUnit(nm string, unit func(s *Parser, value interface{}) (interface{}, error)) *Parser {
	if !s.Scanner.Scan() {
		if err := s.Scanner.Err(); err == nil || err == io.EOF {
			s.Scanner.Error = &ErrExcept{LineNumber: s.LineNumber,
				Offset:   s.Scanner.Offset,
				Excepted: []byte("field[" + nm + "] "),
				Err:      io.EOF}
		}
		return s
	}
	if i, e := strconv.ParseInt(string(s.Scanner.Bytes()), 10, 64); nil != e {
		s.Scanner.Error = &ErrExcept{LineNumber: s.LineNumber,
			Offset:   s.Scanner.Offset,
			Excepted: []byte("field[" + nm + "] "),
			Err:      e}
	} else if v, e := unit(s, i); nil != e {
		s.Scanner.Error = &ErrExcept{LineNumber: s.LineNumber,
			Offset:   s.Scanner.Offset,
			Excepted: []byte("field[" + nm + "] "),
			Err:      e}
	} else {
		s.value[nm] = v
	}

	return s
}

func (s *Parser) ExceptFloatWithUnit(nm string, unit func(s *Parser, value interface{}) (interface{}, error)) *Parser {
	if !s.Scanner.Scan() {
		return s
	}
	if f, e := strconv.ParseFloat(string(s.Scanner.Bytes()), 10); nil != e {
		s.Scanner.Error = &ErrExcept{LineNumber: s.LineNumber,
			Offset:   s.Scanner.Offset,
			Excepted: []byte("field[" + nm + "] "),
			Err:      e}
	} else if v, e := unit(s, f); nil != e {
		s.Scanner.Error = &ErrExcept{LineNumber: s.LineNumber,
			Offset:   s.Scanner.Offset,
			Excepted: []byte("field[" + nm + "] "),
			Err:      e}
	} else {
		s.value[nm] = v
	}

	return s
}

func (s *Parser) Err() error {
	return s.Scanner.Err()
}

func (s *Parser) Result() (map[string]interface{}, error) {
	return s.value, s.Scanner.Err()
}

type ErrExcept struct {
	Excepted   []byte
	Actual     []byte
	LineNumber int
	Offset     int
	Err        error
}

func (e *ErrExcept) Error() string {
	if nil != e.Err {
		if 0 == len(e.Actual) {
			return "want read '" + string(e.Excepted) + "' at " + strconv.Itoa(e.LineNumber) + ":" + strconv.Itoa(e.Offset) + ", " + e.Err.Error()
		}
		return "want read '" + string(e.Excepted) + "' at " + strconv.Itoa(e.LineNumber) + ":" + strconv.Itoa(e.Offset) + ", got '" + string(e.Actual) + "'," + e.Err.Error()
	}
	if 0 == len(e.Actual) {
		return "want read '" + string(e.Excepted) + "' at " + strconv.Itoa(e.LineNumber) + ":" + strconv.Itoa(e.Offset) + ", but fail"
	}
	return "want read '" + string(e.Excepted) + "' at " + strconv.Itoa(e.LineNumber) + ":" + strconv.Itoa(e.Offset) + ", got '" + string(e.Actual) + "'"
}
