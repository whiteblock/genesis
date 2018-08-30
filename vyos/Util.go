package vyos


import (
	"fmt"
	"regexp"
)


func GrabValue(indent int,term string,data string) string{

	initPattern := regexp.MustCompile(fmt.Sprintf("(?m)^( ){%d}%s( )+([0-9|A-z]|\\.|\\/|:|\\-|_)*",indent,term))
	initResults := initPattern.FindAllString(data,-1)
	if initResults == nil {
		return ""
	}
	extractPattern := regexp.MustCompile(`([0-9|A-z]|\.|\/|:|\-|_)+`)
	results := extractPattern.FindAllString(initResults[0],2)

	return results[1]
}

func GrabPairs(indent int,data string) map[string]string {
	initPattern := regexp.MustCompile(fmt.Sprintf("(?m)^( ){%d}([0-9|A-z]|\\.|\\/|:|\\-|_)*( )+([0-9|A-z]|\\.|\\/|:|\\-|_)*",indent))
	initResults := initPattern.FindAllString(data,-1)
	out := make(map[string]string)
	if initResults == nil {
		return out
	}
	extractPattern := regexp.MustCompile(`([0-9|A-z]|\.|\/|:|\-|_)+`)
	for _,line := range initResults {
		pair := extractPattern.FindAllString(line,2)
		out[pair[0]] = pair[1]
	}
	return out;

} 

func MakeIndent(indent int) string {
	out := ""
	for i := 0; i < indent; i++ {
		out += " "
	}
	return out;
}