package StitchingSQLGo

import (
	"strconv"
	"strings"
)

type sql struct {
	strings.Builder
	args []interface{}
}

func (s *sql) push(arg interface{}) {

	l := len(s.args) + 1
	lS := strconv.FormatInt(int64(l), 10)
	s.WriteString(" $")
	s.WriteString(lS)
	s.args = append(s.args, arg)

}