package vyos

import (
    "regexp"
    "fmt"
    "log"
)

type Service struct{
    name    string
    values  map[string]string
}

func LoadServices(data string) []*Service {
    startPattern := regexp.MustCompile("(?m)^( ){3,5}[A-z]([A-z|0-9| |\t])*{")
    startIndexes := startPattern.FindAllIndex([]byte(data),-1)
    endPattern := regexp.MustCompile(`(?m)^( ){3,5}\}`)
    endIndexes := endPattern.FindAllIndex([]byte(data),-1)
    //fmt.Printf("Start indexes are %v\n",startIndexes)
    //fmt.Printf("End index are %v\n",endIndexes)
    services := []*Service{}

    for i,start := range startIndexes {
        
        if len(start) > 1 && len(endIndexes[i]) > 0  {
            services = append(services,NewService(data[start[1]:endIndexes[i][0]],data[start[0]:start[1]],8))
        }else{
            log.Println("Out of range")
            panic(1)
        }
    }

    return services;
}

func ServicesToString(services []*Service) string {
    out := ""
    for _,service := range services {
        out += service.ToString()
    }
    return out
}

func (this *Service) ToString() string {
    shallowIndent := MakeIndent(4)
    mediumIndent := MakeIndent(8)

    out := fmt.Sprintf("%s%s {\n",shallowIndent,this.name)
    for key,value := range this.values {
        out += fmt.Sprintf("%s%s %s\n",mediumIndent,key,value)
    }

    out += fmt.Sprintf("%s}\n",shallowIndent);
    //fmt.Printf("OutToString:\n%s\n",out)
    return out
}


func NewService(data string,name string,indent int) *Service {
    //fmt.Printf("Given data: %s\n\nname:%s\n",data,name)
    namePattern := regexp.MustCompile(`[A-z|0-9]+`)
    extractedName := namePattern.FindAllString(name,1)[0]
    service := new(Service)
    service.name = extractedName
    service.values = GrabPairs(indent,data)
    //fmt.Printf("Out:\n%#v\n",service)
    return service
}
