package sqldb

import "testing"

func TestCondition(t *testing.T) {
	cs := NewConditions()
	c := NewCondition()
	c.Field("name").Op("=").Value("test")
	cs.Add(c)
	var whereClause string
	var args []any
	conditionsCalsue(PostgreSQL, cs, &whereClause, &args)
	t.Logf("whereClause: %s", whereClause)
	t.Logf("args: %v", args)
}
