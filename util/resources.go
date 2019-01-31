package util

import(
    "strconv"
    "strings"
    "errors"
    "fmt"
)

type Resources struct{
    Cpus    string  `json:"cpus"`
    Memory  string  `json:"memory"`
}

func memconv(mem string) (int64,error){
    m := strings.ToLower(mem)
    var multiplier int64 = -1
     
    switch m[len(m)-2:] {
        case "kb":
            multiplier = 1000
        
        case "mb":
            multiplier = 1000000
            
        case "gb":
            multiplier = 1000000000

        case "tb":
            multiplier = 1000000000000
    }
    
    if multiplier == -1 {
        return strconv.ParseInt(m, 10, 64)
    }

    i, err := strconv.ParseInt(m[:len(m)-2], 10, 64)
    if err != nil {
        return -1, err
    }
    return i*multiplier,nil
}
/**
 * Return the memory value as an integer
 */
func (this Resources) GetMemory() (int64,error) {
    return memconv(this.Memory)
}

func (this Resources) Validate() error {
    if this.NoLimits() {
        return nil
    }
    if !this.NoMemoryLimits() {
        m1,err   := memconv(conf.MaxNodeMemory)
        if err != nil {
            panic(err)
        }
        m2,err := this.GetMemory()
        fmt.Printf("m2 = %d\n",m2)
        if err != nil{
            return err
        }
        if m2 > m1 {
            return errors.New("Cannot give each node that much RAM, max is "+conf.MaxNodeMemory)
        }
    }
    
    if !this.NoCpuLimits() {
        c1 := conf.MaxNodeCpu
        c2,err := strconv.ParseFloat(this.Cpus,64)
        if err != nil{
            return err
        }
        if c2 > c1 {
            return errors.New(fmt.Sprintf("Cannot give each node that much CPU, max is %f",conf.MaxNodeCpu))
        }
    }
    
    return nil
}

func (this Resources) ValidateAndSetDefaults() error {
    err := this.Validate()
    if err != nil {
        return err
    }
    if this.NoCpuLimits() {
        this.Cpus = fmt.Sprintf("%f",conf.MaxNodeCpu)
    }
    if this.NoMemoryLimits() {
        this.Memory = conf.MaxNodeMemory
    }
    return nil
}

func (this Resources) NoLimits() bool{
    return len(this.Memory) == 0 && len(this.Cpus) == 0
}

func (this Resources) NoCpuLimits() bool {
    return len(this.Cpus) == 0
}

func (this Resources) NoMemoryLimits() bool {
    return len(this.Memory) == 0
}