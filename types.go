package gorules

import (
	"time"
)

type Event struct {
	EventName string
	Meta      map[string]string
	Timestamp time.Time
}

type Reward struct {
	Event  Event
	Amount int32
	Rule   Rule
}

type RulesList struct {
	EventName       string
	ApplicationRule string
	Rules           []Rule
}

type Rule struct {
	Id         string
	Message    string
	Conditions []Condition
	Reward     int32
	StartTime  time.Time
	EndTime    time.Time
}

type Condition struct {
	Field    string
	Op       Operator
	Value    string
	ValueArr []string
}

type Operator int

const (
	Equal Operator = iota
	NotEqual
	LessThan
	LessThanEqual
	GreaterThan
	GreaterThanEqual
	In
	NotIn
)
