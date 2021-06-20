package gorules

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"strconv"
	"time"

	dbqueue "github.com/a8a-io/go-dbqueue"
	"github.com/a8a-io/gorules/utils"
)

const RuleEvent = "RuleEvents"

type RuleEngine interface {
	Start()
	// AwardCoins(userId string, eventName string, fieldValues map[string]string, timestamp time.Time)
	setRules(rules map[string]RulesList)
	getRules(eventName string) (RulesList, bool)
}

type _RuleEngine struct {
	rulesFolderPath string
	Rules           map[string]RulesList
	EventQ          *dbqueue.DBQueue
	RewardQ         *dbqueue.DBQueue
}

func NewRuleEngine(rulesFolderPath string) RuleEngine {
	return _RuleEngine{rulesFolderPath: rulesFolderPath}
}

func (this _RuleEngine) setRules(rulesMap map[string]RulesList) {
	this.Rules = rulesMap
}

func (this _RuleEngine) Start() {
	ticker := time.NewTicker(5 * time.Minute)
	quit := make(chan struct{})
	go refreshRules(this.rulesFolderPath, this, ticker, quit)
	go calculateCoins(this)
}

func (this _RuleEngine) getRules(eventName string) (RulesList, bool) {
	rl, ok := this.Rules[eventName]
	return rl, ok
}

func refreshRules(rulesFolderPath string, ruleEngine RuleEngine, ticker *time.Ticker, quit <-chan struct{}) {
	for {
		select {
		case <-ticker.C:
			rulesMap := LoadRulesFromFolder(rulesFolderPath)
			ruleEngine.setRules(rulesMap)
		case <-quit:
			ticker.Stop()
			return
		}
	}
}

func calculateCoins(re _RuleEngine) int {
	for {
		event, err := DequeEvent(re.EventQ)
		if err != nil {
			fmt.Println(err)
			time.Sleep(1 * time.Second)
			continue
		}

		rulesList, ok := re.getRules(event.EventName)
		if !ok {
			log.Printf("No rule found for event %s", event.EventName)
		}

		var reward Reward
		for _, rule := range rulesList.Rules {
			if validateEvent(rule, event) {
				reward = Reward{event, rule.Reward, rule}
				break
			}
		}
		var rewardBytes bytes.Buffer
		err = gob.NewEncoder(&rewardBytes).Encode(reward)
		if err != nil {
			fmt.Println(err)
			continue
		}
		err = (*re.RewardQ).Enque("Rewards", nil, rewardBytes.Bytes())
		if err != nil {
			fmt.Println(err)
		}
	}
}

func DequeEvent(EventQueue *dbqueue.DBQueue) (Event, error) {
	var event Event
	msg, err := (*EventQueue).Deque(RuleEvent)
	if err != nil {
		return Event{}, err
	}
	z := bytes.NewBuffer(msg.Body)
	err = gob.NewDecoder(z).Decode(&event)
	if err != nil {
		return Event{}, err
	}

	return event, nil
}

func validateEvent(rule Rule, event Event) bool {
	if !utils.InBetween(event.Timestamp, rule.StartTime, rule.EndTime) {
		return false
	}
	var allConditionMatch bool = true

	for _, condition := range rule.Conditions {
		field := condition.Field
		val, ok := event.Meta[field]
		if !ok {
			return false
		}
		switch condition.Op {
		case LessThan, LessThanEqual, GreaterThan, GreaterThanEqual:
			vi, err := strconv.ParseInt(val, 10, 32)
			cvi, err2 := strconv.ParseInt(condition.Value, 10, 32)
			if err != nil || err2 != nil {
				allConditionMatch = false
			} else {
				switch condition.Op {
				case LessThan:
					allConditionMatch = allConditionMatch && (vi < cvi)
				case LessThanEqual:
					allConditionMatch = allConditionMatch && (vi <= cvi)
				case GreaterThan:
					allConditionMatch = allConditionMatch && (vi > cvi)
				case GreaterThanEqual:
					allConditionMatch = allConditionMatch && (vi >= cvi)
				}
			}
		case Equal:
			allConditionMatch = allConditionMatch && (val == condition.Value)
		case NotEqual:
			allConditionMatch = allConditionMatch && (val != condition.Value)
		case In:
			allConditionMatch = allConditionMatch && utils.Contains(condition.ValueArr, val)
		case NotIn:
			allConditionMatch = allConditionMatch && !utils.Contains(condition.ValueArr, val)
		}

		if !allConditionMatch {
			break
		}
	}
	return allConditionMatch
}
