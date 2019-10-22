package util

import "io/ioutil"

func ReadLines(filename string) ([][]byte, error) {
	bs, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	return SplitLines(bs)
}

func ReadStringLines(filename string, ignoreEmpty bool) ([]string, error) {
	bs, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	lines, err := SplitLines(bs)
	if err != nil {
		return nil, err
	}

	ss := make([]string, 0, len(lines))
	for idx := range lines {
		if ignoreEmpty {
			if len(lines[idx]) == 0 {
				continue
			}
		}

		ss = append(ss, string(lines[idx]))
	}
	return ss, nil
}
