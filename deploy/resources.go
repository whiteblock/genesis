package deploy

import(
    "strconv"
    "strings"
)

type Resources struct{
    Cpus    string  `json:"cpus"`
    Memory  string  `json:"memory"`
}

/**
 * Return the memory value as an integer
 */
func (this Resources) GetMemory() (int64,error) {
    m := strings.ToLower(this.Memory)
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

func (this Resources) NoLimits() bool{
    return len(this.Memory) == 0 && len(this.Cpus) == 0
}

func (this Resources) NoCpuLimits() bool {
    return len(this.Cpus) == 0
}

func (this Resources) NoMemoryLimits() bool {
    return len(this.Memory) == 0
}