package gorules

func MakeRuleEngine() RuleEngine {
	re := _RuleEngine{}
	re.init("./rules")
	return re
}
