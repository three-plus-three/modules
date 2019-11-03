package assert

import (
	"fmt"
	"runtime"
	"strings"
	"testing"
)

// -----------------------------------------------------------------------
// Generic checks and assertions based on checkers.

// Verify if the first value matches with the expected value.  What
// matching means is defined by the provided checker. In case they do not
// match, an error will be logged, the test will be marked as failed, and
// the test execution will continue.  Some checkers may not need the expected
// argument (e.g. IsNil).  In either case, any extra arguments provided to
// the function will be logged next to the reported problem when the
// matching fails.  This is a handy way to provide problem-specific hints.
func Check(t *testing.T, obtained interface{}, checker Checker, args ...interface{}) bool {
	return internalCheck(t, 2, "Check", obtained, checker, args...)
}

// Ensure that the first value matches with the expected value.  What
// matching means is defined by the provided checker. In case they do not
// match, an error will be logged, the test will be marked as failed, and
// the test execution will stop.  Some checkers may not need the expected
// argument (e.g. IsNil).  In either case, any extra arguments provided to
// the function will be logged next to the reported problem when the
// matching fails.  This is a handy way to provide problem-specific hints.
func Assert(t *testing.T, obtained interface{}, checker Checker, args ...interface{}) {
	if !internalCheck(t, 2, "Assert", obtained, checker, args...) {
		t.FailNow()
	}
}

func internalCheck(t *testing.T, cd int, funcName string, obtained interface{}, checker Checker, args ...interface{}) bool {
	if checker == nil {
		_, file, line, _ := runtime.Caller(cd + 1)
		t.Errorf("%s:%d", file, line)
		t.Logf("%s(obtained, nil!?, ...):", funcName)
		t.Log("Oops.. you've provided a nil checker!")
		t.FailNow()
		return false
	}

	// If the last argument is a bug info, extract it out.
	var comment CommentInterface
	if len(args) > 0 {
		if c, ok := args[len(args)-1].(CommentInterface); ok {
			comment = c
			args = args[:len(args)-1]
		}
	}

	params := append([]interface{}{obtained}, args...)
	info := checker.Info()

	if len(params) != len(info.Params) {
		names := append([]string{info.Params[0], info.Name}, info.Params[1:]...)
		_, file, line, _ := runtime.Caller(cd + 1)
		t.Errorf("%s:%d", file, line)
		t.Log(fmt.Sprintf("%s(%s):", funcName, strings.Join(names, ", ")))
		t.Log(fmt.Sprintf("Wrong number of parameters for %s: want %d, got %d", info.Name, len(names), len(params)+1))
		t.FailNow()
		return false
	}

	// Copy since it may be mutated by Check.
	names := append([]string{}, info.Params...)

	// Do the actual check.
	result, error := checker.Check(params, names)
	if !result || error != "" {
		for i := 0; i != len(params); i++ {
			t.Logf("%s = %v", names[i], params[i])
		}
		if comment != nil {
			t.Log(comment.CheckCommentString())
		}
		if error != "" {
			t.Log(error)
		}
		t.FailNow()
		return false
	}
	return true
}
