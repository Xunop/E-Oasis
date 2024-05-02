package util

import (
	"database/sql/driver"
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/google/uuid"
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
	if sc.ans != nil && ndx != 0 && value != "" {
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
	ans []string
	sep string
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

func TitleSort(title string) string {
	// calibre sort stuff
	// ^(A|The|An|Der|Die|Das|Den|Ein|Eine|Einen|Dem|Des|Einem|Eines|Le|La|Les|L\'|Un|Une)\s+
	match := TitleSortMatcher.FindStringSubmatch(title)
	if match != nil {
		prep := match[1]
		title = strings.TrimPrefix(title, prep) + ", " + prep
	}
	return strings.TrimSpace(title)
}

func GetSortedAuthor(value string) string {
	var value2 string
	regexes := []string{"^(JR|SR)\\.?$", "^I{1,3}\\.?$", "^IV\\.?$"}
	combined := "(" + strings.Join(regexes, "|") + ")"
	values := strings.Split(value, " ")

	if strings.Index(value, ",") == -1 {
		if match, _ := regexp.MatchString(combined, strings.ToUpper(values[len(values)-1])); match {
			if len(values) > 1 {
				value2 = values[len(values)-2] + ", " + strings.Join(values[:len(values)-2], " ") + " " + values[len(values)-1]
			} else {
				value2 = values[0]
			}
		} else if len(values) == 1 {
			value2 = values[0]
		} else {
			value2 = values[len(values)-1] + ", " + strings.Join(values[:len(values)-1], " ")
		}
	} else {
		value2 = value
	}

	return value2
}

func UUID4() (string, error) {
	uuidObj, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}
	return uuidObj.String(), nil
}
