package gorules

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const RulesFileSuffix = "-rules.json"

func ReadFileToString(path string) ([]byte, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Print(err)
		return nil, err
	}
	return data, nil
}

type RulesFile struct {
	Event           string `json:"event"`
	ApplicationRule string `json:"application"`
	Rules           []struct {
		ID         string   `json:"id"`
		Message    string   `json:"message"`
		Conditions []string `json:"conditions"`
		StartDate  string   `json:"start_date"`
		EndDate    string   `json:"end_date"`
		Reward     int32    `json:"reward"`
	} `json:"rules"`
}

func LoadRulesFromFolder(path string) map[string]RulesList {
	files, _ := ioutil.ReadDir(path)

	rules := make(map[string]RulesList)
	for _, f := range files {
		if strings.HasSuffix(f.Name(), RulesFileSuffix) && !f.IsDir() {
			rulesList, err := LoadRulesFromFile(filepath.Join(path, f.Name()))
			if err != nil {
				log.Println("Error reading file", f.Name(), err)
				return nil
			}
			// fmt.Println("rulesList.EventName", rulesList.EventName)
			// fmt.Println("rulesList", rulesList)
			rules[rulesList.EventName] = rulesList
		}
	}
	// fmt.Println(rules)
	return rules
}

func LoadRulesFromFile(filePath string) (RulesList, error) {
	fileText, err := ReadFileToString(filePath)
	if err != nil {
		fmt.Println("Error Reading File:", err)
		return RulesList{}, err
	}
	var rulesFile RulesFile
	err = json.Unmarshal(fileText, &rulesFile)
	if err != nil {
		fmt.Println("error:", err)
		return RulesList{}, err
	}
	fmt.Println(rulesFile)

	rules, err := extractRulesFile(rulesFile)

	return rules, err
}

func extractRulesFile(rulesFile RulesFile) (RulesList, error) {
	rules := make([]Rule, 0)
	for _, rule := range rulesFile.Rules {
		conditionList := make([]Condition, 0)
		for _, conditionStr := range rule.Conditions {
			tokens := strings.SplitN(conditionStr, " ", 3)
			op, err := stringToOperator(tokens[1])
			if err != nil {
				return RulesList{}, err
			}
			valueArr := strings.Split(tokens[2], ",")
			condition := Condition{tokens[0], op, tokens[2], valueArr}
			conditionList = append(conditionList, condition)
		}
		dateLayout := "2006-01-02 15:04"
		startDate, err := time.Parse(dateLayout, rule.StartDate)
		endDate, err2 := time.Parse(dateLayout, rule.StartDate)

		if err != nil || err2 != nil {
			log.Printf("Error parsing date in Rules %s | %s", rule.StartDate, rule.EndDate)
			return RulesList{}, fmt.Errorf("Error parsing date in Rules %s | %s", rule.StartDate, rule.EndDate)
		}

		rule := Rule{rule.ID, rule.Message, conditionList, rule.Reward, startDate, endDate}
		// fmt.Println("Found Rule", rule, len(rule.Conditions))
		rules = append(rules, rule)
	}
	sort.Slice(rules, func(i, j int) bool { return rules[i].Reward > rules[j].Reward })
	return RulesList{rulesFile.Event, rulesFile.ApplicationRule, rules}, nil
}

var operatorLabels = map[string]Operator{
	"=":     Equal,
	"!=":    NotEqual,
	"<":     LessThan,
	"<=":    LessThanEqual,
	">":     GreaterThan,
	">=":    GreaterThanEqual,
	"in":    In,
	"notin": NotIn,
}

func stringToOperator(label string) (Operator, error) {
	elem, ok := operatorLabels[label]
	if ok {
		return elem, nil
	} else {
		return -1, fmt.Errorf("Unknown Operator |" + label + "|")
	}
}
