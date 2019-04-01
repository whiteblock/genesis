package util

import(
    "strconv"
    "strings"
    "fmt"
)

/*
    Resources represents the maximum amount of resources 
    that a node can use. 
 */
type Resources struct{
    /*
        Cpus should be a floating point value represented as a string, and
        is  equivalent to the percentage of a single cores time which can be used
        by a node. Can be more than 1.0, meaning the node can use multiple cores at 
        a time.
     */
    Cpus    string  `json:"cpus"`
    /*
        Memory supports values up to Terrabytes (tb). If the unit is ommited, then it
        is assumed to be bytes. This is not case sensitive.
     */
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
/*
    GetMemory gets the memory value as an integer.
 */
func (this Resources) GetMemory() (int64,error) {
    return memconv(this.Memory)
}

/*
    Validate ensures that the given resource object is valid, and
    allowable.
 */
func (this Resources) Validate() error {
    if this.NoLimits() {
        return nil
    }

    err := ValidateCommandLine(this.Memory)
    if err != nil {
        return err
    }

    err = ValidateCommandLine(this.Cpus)
    if err != nil {
        return err
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
            return fmt.Errorf("Cannot give each node that much RAM, max is %s",conf.MaxNodeMemory)
        }
    }
    
    if !this.NoCpuLimits() {
        c1 := conf.MaxNodeCpu
        c2,err := strconv.ParseFloat(this.Cpus,64)
        if err != nil{
            return err
        }
        if c2 > c1 {
            return fmt.Errorf("Cannot give each node that much CPU, max is %f",conf.MaxNodeCpu)
        }
    }
    
    return nil
}

/*
    ValidateAndSetDefaults calls Validate, and if it is valid, fills any missing
    information. Helps to ensure that the Maximum limits are enforced.
 */
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

/*
    NoLimits checks if the resources object doesn't specify any limits
 */
func (this Resources) NoLimits() bool{
    return len(this.Memory) == 0 && len(this.Cpus) == 0
}

/*
    NoCpuLimits checks if the resources object doesn't specify any cpu limits
 */
func (this Resources) NoCpuLimits() bool {
    return len(this.Cpus) == 0
}

/*
    NoMemoryLimits checks if the resources object doesn't specify any memory limits
 */
func (this Resources) NoMemoryLimits() bool {
    return len(this.Memory) == 0
}