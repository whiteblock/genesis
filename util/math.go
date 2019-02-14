package util

import(
    "math"
)

type Point struct{
    X   int
    Y   int
}

func Distances(pnts []Point) [][]float64 {
    out := make([][]float64,len(pnts))
    for i,_ := range pnts{
        out[i] = make([]float64,len(pnts))
    }

    for i,ipnt := range pnts{
        for j,jpnt := range pnts{
            if j == i {
                continue;
            }
            diffX := math.Abs(float64(ipnt.X - jpnt.X))
            diffY := math.Abs(float64(ipnt.Y - jpnt.Y))
            out[i][j] = math.Sqrt(math.Pow(diffX,2)+math.Pow(diffY,2))
        }
    }
    return out
}

