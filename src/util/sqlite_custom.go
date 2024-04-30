package util

import (
	"database/sql/driver"
	"fmt"
	"slices"
	"strings"

	"modernc.org/sqlite"
)

type SortedConcatenate struct {
	ans         map[int]string
	sep         string
	finalCalled bool
}

var SortConcatFinalCalled bool

func NewSortedConcatenate(sep string) *SortedConcatenate {
	return &SortedConcatenate{ans: make(map[int]string), sep: sep, finalCalled: false}
}

func (sc *SortedConcatenate) Step(ctx *sqlite.FunctionContext, rowArgs []driver.Value) error {
	ndx := 0
	switch rowArgs[0].(type) {
	case int64:
		ndx = int(rowArgs[0].(int64))
	default:
		return fmt.Errorf("invalid type: %T", rowArgs[0])
	}
	value := ""
	switch rowArgs[1].(type) {
	case string:
	    value = rowArgs[1].(string)
	default:
		return fmt.Errorf("invalid type: %T", rowArgs[1])
	}
	if sc.ans != nil && ndx != 0 && value != ""{
		sc.ans[ndx] = value
	}
	return nil
}

func (sc *SortedConcatenate) WindowValue(ctx *sqlite.FunctionContext) (driver.Value, error) {
	if len(sc.ans) == 0 {
		return "", nil
	}

	keys := make([]int, 0, len(sc.ans))
	for k := range sc.ans {
		keys = append(keys, k)
	}
	slices.Sort(keys)

	var values []string
	for _, k := range keys {
		values = append(values, sc.ans[k])
	}
	return strings.Join(values, sc.sep), nil
}

func (sc *SortedConcatenate) WindowInverse(ctx *sqlite.FunctionContext, args []driver.Value) error {
	return nil
}

func (sc *SortedConcatenate) Final(ctx *sqlite.FunctionContext) {
	sc.finalCalled = true
}

func (sc *SortedConcatenate) GetFinalCalled() bool {
	return sc.finalCalled
}

type Concatenate struct {
	ans         []string
	sep         string
}

func NewConcatenate(sep string) *Concatenate {
	return &Concatenate{sep: sep, ans: []string{}}
}

func (c *Concatenate) Step(ctx *sqlite.FunctionContext, rowArgs []driver.Value) error {
    value := ""
	switch rowArgs[0].(type) {
	case string:
		value = rowArgs[0].(string)
	default:
		return fmt.Errorf("invalid type: %T", rowArgs[0])
	}
	if value != "" {
		c.ans = append(c.ans, value)
	}
	return nil
}

func (c *Concatenate) WindowValue(ctx *sqlite.FunctionContext) (driver.Value, error) {
	return strings.Join(c.ans, c.sep), nil
}

func (c *Concatenate) WindowInverse(ctx *sqlite.FunctionContext, args []driver.Value) error {
	return nil
}

func (c *Concatenate) Final(ctx *sqlite.FunctionContext) {
}
