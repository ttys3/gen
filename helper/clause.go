package helper

import (
	"bytes"
	"fmt"
	"log"
	"reflect"
	"regexp"
	"strings"

	"code.byted.org/security/go-polaris/sql"
)

type Cond struct {
	Cond   bool
	Result string
}

func IfClause(conds []Cond) string {
	judge := func(c Cond) string {
		if c.Cond {
			return c.Result
		}
		return ""
	}

	clauses := make([]string, len(conds))
	for i, cond := range conds {
		clauses[i] = strings.Trim(judge(cond), " ")
	}
	return " " + strings.Join(clauses, " ")
}

func WhereClause(conds []string) string {
	return joinClause(conds, "WHERE", whereValue, " ")
}

func SetClause(conds []string) string {
	return joinClause(conds, "SET", setValue, ",")
}

func joinClause(conds []string, keyword string, deal func(string) string, sep string) string {
	clauses := make([]string, len(conds))
	for i, clause := range conds {
		clauses[i] = deal(clause)
	}

	sql := trimAll(strings.Join(clauses, sep))
	if sql != "" {
		sql = " " + keyword + " " + sql
	}
	return sql
}

func trimAll(input string) string {
	input = strings.Trim(input, " ")
	lowercase := strings.ToLower(input)
	switch {
	case strings.HasPrefix(lowercase, "and "):
		return input[4:]
	case strings.HasPrefix(lowercase, "or "):
		return input[3:]
	case strings.HasPrefix(lowercase, "xor "):
		return input[4:]
	case strings.HasPrefix(lowercase, ","):
		return input[1:]
	default:
		return input
	}
}

// whereValue append a new condition with prefix "AND"
func whereValue(value string) string {
	value = strings.Trim(value, " ")
	lowercase := strings.ToLower(value)
	switch {
	case lowercase == "":
		return ""
	case strings.HasPrefix(lowercase, "and "):
		return value
	case strings.HasPrefix(lowercase, "or "):
		return value
	case strings.HasPrefix(lowercase, "xor "):
		return value
	default:
		return "AND " + value
	}
}

func setValue(value string) string {
	return strings.Trim(value, ", ")
}

// list slice map
func ForClause(params map[string]interface{}, list interface{}, tmpl, clauseName, rangeKey, rangeValue string) string {
	var buf bytes.Buffer
	inputList := interfaceToList(list)
	members := getFieldsName(tmpl, rangeValue)
	for key, val := range inputList {
		tempValue := tmpl
		for _, member := range members {

			replaceKey := rangeKey
			replaceValue := rangeValue
			if member != "" {
				replaceKey = fmt.Sprintf("%s.%s", rangeKey, member)
				replaceValue = fmt.Sprintf("%s.%s", rangeValue, member)
			}

			if strings.Contains(tempValue, "@@"+replaceKey) {
				tempValue = strings.Replace(tempValue, "@@"+replaceKey, sql.EscapeSQLName(fmt.Sprintf("%v", getReplace(key, member))), -1)
			}
			if strings.Contains(tempValue, "@@"+replaceValue) {
				tempValue = strings.Replace(tempValue, "@@"+replaceValue, sql.EscapeSQLName(fmt.Sprintf("%v", getReplace(val, member))), -1)
			}
			if strings.Contains(tempValue, "@"+replaceKey) {
				tempDataName := fmt.Sprintf("%s%s%s_%v", rangeKey, member, strings.Title(clauseName), key)
				tempValue = strings.Replace(tempValue, "@"+replaceKey, "@"+tempDataName, -1)
				params[tempDataName] = getReplace(key, member)
			}
			if strings.Contains(tempValue, "@"+replaceValue) {
				tempDataName := fmt.Sprintf("%s%s%s_%v", rangeValue, member, strings.Title(clauseName), key)
				tempValue = strings.Replace(tempValue, "@"+replaceValue, "@"+tempDataName, -1)
				params[tempDataName] = getReplace(val, member)
			}
		}
		buf.WriteString(" ")
		buf.WriteString(strings.TrimSpace(tempValue))
	}
	return buf.String()
}

// getFieldsName get fields name from tmpl
func getFieldsName(tmpl, structName string) []string {
	var fieldNameMap = make(map[string]int, 0)
	var res []string
	if strings.Contains(tmpl, fmt.Sprintf("@%s.", structName)) {
		fieldRegexp := regexp.MustCompile(fmt.Sprintf("@%s\\.(\\w*)", structName))
		fields := fieldRegexp.FindAllStringSubmatch(tmpl, -1)
		for _, field := range fields {
			fieldNameMap[field[1]] = 1
		}
		for f := range fieldNameMap {
			res = append(res, f)
		}
	}
	// if without field or index is key ,add "" in last
	res = append(res, "")
	return res
}

// getReplace get replace value by field
func getReplace(field interface{}, key string) interface{} {
	if key == "" {
		return field
	}
	if _, ok := reflect.TypeOf(field).FieldByName(key); ok {
		v := reflect.ValueOf(field)
		return v.FieldByName(key).Interface()
	} else {
		log.Fatal("can't find field %s")
	}
	return nil
}

// interfaceToList Adjust interface format，slice/map to map[interface{}]interface{}
func interfaceToList(input interface{}) (inputList []interface{}) {
	inputList = make(map[interface{}]interface{}, 100)
	//// map
	//if reflect.TypeOf(input).Kind().String() == "map" {
	//	v := reflect.ValueOf(input)
	//	m := v.MapRange()
	//	for i := 0; i < v.Len(); i++ {
	//		if m.Next() {
	//			inputList[m.Key()] = m.Value().Interface()
	//		}
	//	}
	//	return inputList
	//}
	// slice
	// TODO add
	v := reflect.ValueOf(input)
	for i := 0; i < v.Len(); i++ {
		inputList[i] = v.Index(i).Interface()
	}
	return inputList
}
