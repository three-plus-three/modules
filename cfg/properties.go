package cfg

import (
	"bufio"
	"io"
	"os"
	"strings"
)

func ReadProperties(nm string) (map[string]string, error) {
	f, e := os.Open(nm)
	if nil != e {
		return nil, e
	}
	defer f.Close()

	cfg := map[string]string{}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		ss := strings.SplitN(scanner.Text(), "#", 2)
		//ss = strings.SplitN(ss[0], "//", 2)
		s := strings.TrimSpace(ss[0])
		if 0 == len(s) {
			continue
		}
		ss = strings.SplitN(s, "=", 2)
		if 2 != len(ss) {
			continue
		}

		key := strings.TrimLeft(strings.TrimSpace(ss[0]), ".")
		value := strings.TrimSpace(ss[1])
		if 0 == len(key) {
			continue
		}
		if 0 == len(value) {
			continue
		}
		cfg[key] = os.ExpandEnv(value)
	}

	return expandAll(cfg), nil
}

func expandAll(cfg map[string]string) map[string]string {
	remain := 0
	expend := func(key string) string {
		if value, ok := cfg[key]; ok {
			return value
		}
		remain++
		return key
	}

	for i := 0; i < 100; i++ {
		for k, v := range cfg {
			cfg[k] = os.Expand(v, expend)
		}
		if 0 == remain {
			break
		}
	}
	return cfg
}

func WriteWith(w io.Writer, values map[string]string) error {
	var err error
	for k, v := range values {
		io.WriteString(w, k)
		io.WriteString(w, "=")
		io.WriteString(w, v)
		_, err = io.WriteString(w, "\r\n")
	}
	return err
}

func WriteProperties(nm string, values map[string]string) error {
	if len(values) == 0 {
		return nil
	}
	f, e := os.Create(nm)
	if nil != e {
		return e
	}
	defer f.Close()
	return WriteWith(f, values)
}

func UpdateWith(r io.Reader, w io.Writer, updated map[string]string) error {
	updatedCopy := map[string]string{}
	for k, v := range updated {
		updatedCopy[k] = v
	}

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		txt := scanner.Text()

		for k, v := range updated {
			if strings.Contains(txt, k) {
				ss := strings.SplitN(txt, "=", 2)
				if 2 == len(ss) {
					key := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(ss[0]), "#"))
					if key == k {
						if ss = strings.SplitN(ss[1], "#", 2); 2 == len(ss) {
							txt = k + "=" + v + " #" + ss[1]
						} else {
							txt = k + "=" + v
						}
						delete(updatedCopy, k)
						break
					}
				}
			}
		}
		io.WriteString(w, txt)
		io.WriteString(w, "\r\n")
	}

	if scanner.Err() != nil {
		return scanner.Err()
	}

	var err error
	for k, v := range updatedCopy {
		io.WriteString(w, k)
		io.WriteString(w, "=")
		io.WriteString(w, v)
		_, err = io.WriteString(w, "\r\n")
	}
	return err
}

func UpdateProperties(nm string, updated map[string]string) error {
	if len(updated) == 0 {
		return nil
	}
	f, e := os.Open(nm)
	if nil != e {
		return e
	}
	defer f.Close()

	out, e := os.Create(nm + ".tmp")
	if nil != e {
		return e
	}
	defer out.Close()

	if e := UpdateWith(f, out, updated); nil != e {
		return e
	}

	if e := out.Close(); nil != e {
		return e
	}
	if e := f.Close(); nil != e {
		return e
	}
	if e := os.Remove(nm); nil != e {
		return e
	}
	if e := os.Rename(nm+".tmp", nm); nil != e {
		return e
	}
	return nil
}
