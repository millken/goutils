package sqldb

import (
	"fmt"
	"sync"
)

type Conditions []*Condition

type Condition struct {
	field     string
	operate   string
	value     any
	alias     string
	connector string
	sub       Conditions
}

var conditionPool = sync.Pool{
	New: func() any {
		return &Condition{}
	},
}

func acquireCondition() *Condition {
	return conditionPool.Get().(*Condition)
}

func releaseCondition(c *Condition) {
	c.field = ""
	c.operate = ""
	c.value = nil
	c.alias = ""
	c.connector = ""
	c.sub = c.sub[:0]
	conditionPool.Put(c)
}

func NewConditions(conditions ...*Condition) Conditions {
	if len(conditions) == 0 {
		return make([]*Condition, 0)
	}
	return Conditions(conditions)
}

func (c *Conditions) Add(condition *Condition) {
	*c = append(*c, condition)
}

func conditionsCalsue(flavor Flavor, c Conditions, whereClause *string, args *[]any) {
	if len(c) == 0 {
		return
	}
	for i, condition := range c {
		if i > 0 {
			connector := "AND"
			if condition.connector != "" {
				connector = condition.connector
			}
			*whereClause += fmt.Sprintf(" %s ", connector)
		}
		conditionCalsue(flavor, condition, whereClause, args)
	}
}

func conditionCalsue(flavor Flavor, c *Condition, whereClause *string, args *[]any) {
	defer releaseCondition(c)
	if len(c.sub) > 0 {
		conditionsCalsue(flavor, c.sub, whereClause, args)
	} else {
		if c.alias != "" {
			*whereClause += fmt.Sprintf("%s.%s%s?", flavor.columnQuote(c.alias), c.field, c.operate)
		} else {
			*whereClause += fmt.Sprintf("%s%s?", flavor.columnQuote(c.field), c.operate)
		}
		*args = append(*args, c.value)
	}
}

func NewCondition() *Condition {
	return acquireCondition()
}

func (c *Condition) Field(field string) *Condition {
	c.field = field
	return c
}

func (c *Condition) Op(operate string) *Condition {
	c.operate = operate
	return c
}

func (c *Condition) Value(value any) *Condition {
	c.value = value
	return c
}

func (c *Condition) Alias(alias string) *Condition {
	c.alias = alias
	return c
}

func (c *Condition) Connector(connector string) *Condition {
	c.connector = connector
	return c
}

func (c *Condition) Sub(sub ...*Condition) *Condition {
	c.sub = sub
	return c
}
