package main

import (
    "fmt"
    "testing"
)


// Utilites


// Transform funcitons

func TestScale(t *testing.T) {
    var tests = []struct{
        inputSd []SingleData
        delta float64
        output []SingleData
    }{
        {
            inputSd: []SingleData{
                {
                    Times: TimeArrayHelper(0,6),
                    Values: []float64{1,1,2,3,5,8},
                },
            },
            delta: 2,
            output: []SingleData{
                {
                    Times: TimeArrayHelper(0,6),
                    Values: []float64{2,2,4,6,10,16},
                },
            },
        },
    }
    for tdx, testCase := range tests {
        testName := fmt.Sprintf("case %d: %v", tdx, testCase.output)
        t.Run(testName, func(t *testing.T) {
            result := Scale(testCase.inputSd, testCase.delta)
            SingleDataCompareHelper(result, testCase.output, t)
        })
    }
}

func TestOffset(t *testing.T) {
    var tests = []struct{
        inputSd []SingleData
        delta float64
        output []SingleData
    }{
        {
            inputSd: []SingleData{
                {
                    Times: TimeArrayHelper(0,6),
                    Values: []float64{1,1,2,3,5,8},
                },
            },
            delta: 2,
            output: []SingleData{
                {
                    Times: TimeArrayHelper(0,6),
                    Values: []float64{3,3,4,5,7,10},
                },
            },
        },
    }
    for tdx, testCase := range tests {
        testName := fmt.Sprintf("case %d: %v", tdx, testCase.output)
        t.Run(testName, func(t *testing.T) {
            result := Offset(testCase.inputSd, testCase.delta)
            SingleDataCompareHelper(result, testCase.output, t)
        })
    }
}

func TestDelta(t *testing.T) {
    var tests = []struct{
        inputSd []SingleData
        output []SingleData
    }{
        {
            inputSd: []SingleData{
                {
                    Times: TimeArrayHelper(0,6),
                    Values: []float64{1,1,2,3,5,8},
                },
            },
            output: []SingleData{
                {
                    Times: TimeArrayHelper(1,6),
                    Values: []float64{0,1,1,2,3},
                },
            },
        },
    }
    for tdx, testCase := range tests {
        testName := fmt.Sprintf("case: %d", tdx)
        t.Run(testName, func(t *testing.T) {
            result := Delta(testCase.inputSd)
            SingleDataCompareHelper(result, testCase.output, t)
        })
    }
}

func TestFluctuation(t *testing.T) {
    var tests = []struct{
        inputSd []SingleData
        output []SingleData
    }{
        {
            inputSd: []SingleData{
                {
                    Times: TimeArrayHelper(0,6),
                    Values: []float64{1,1,2,3,5,8},
                },
            },
            output: []SingleData{
                {
                    Times: TimeArrayHelper(1,6),
                    Values: []float64{0,0,1,2,4,7},
                },
            },
        },
    }
    for tdx, testCase := range tests {
        testName := fmt.Sprintf("case: %d", tdx)
        t.Run(testName, func(t *testing.T) {
            result := Fluctuation(testCase.inputSd)
            SingleDataCompareHelper(result, testCase.output, t)
        })
    }
}

func TestMovingAverage(t *testing.T) {
    var tests = []struct{
        inputSd []SingleData
        windowSize int
        output []SingleData
    }{
        {
            inputSd: []SingleData{
                {
                    Times: TimeArrayHelper(0,6),
                    Values: []float64{3,9,6,3,3,6},
                },
            },
            windowSize: 3,
            output: []SingleData{
                {
                    Times: TimeArrayHelper(0,6),
                    Values: []float64{3,6,6,6,4,4},
                },
            },
        },
    }
    for tdx, testCase := range tests {
        testName := fmt.Sprintf("case: %d", tdx)
        t.Run(testName, func(t *testing.T) {
            result := MovingAverage(testCase.inputSd, testCase.windowSize)
            SingleDataCompareHelper(result, testCase.output, t)
        })
    }
}


// Array to Scalar Functions

func TestToScalarByAvg(t *testing.T) {
    t.Skipf("Not Implemeneted")
}

func TestToScalarByMax(t *testing.T) {
    t.Skipf("Not Implemeneted")
}

func TestToScalarByMin(t *testing.T) {
    t.Skipf("Not Implemeneted")
}

func TestToScalarBySum(t *testing.T) {
    t.Skipf("Not Implemeneted")
}

func TestToScalarByMed(t *testing.T) {
    t.Skipf("Not Implemeneted")
}

func TestToScalarByStd(t *testing.T) {
    t.Skipf("Not Implemeneted")
}


// Filter Series Functions 

func TestTop(t *testing.T) {
    var tests = []struct{
        inputSd []SingleData
        number int
        value string
        output []SingleData
    }{
        {
            inputSd: []SingleData{
                {
                    Times: TimeArrayHelper(0,6),
                    Values: []float64{1,1,2,3,5,8},
                },
                {
                    Times: TimeArrayHelper(0,6),
                    Values: []float64{10,10,20,30,50,80},
                },
            },
            number: 1,
            value: "avg",
            output: []SingleData{
                {
                    Times: TimeArrayHelper(0,6),
                    Values: []float64{10,10,20,30,50,80},
                },
            },
        },
        {
            inputSd: []SingleData{
                {
                    Times: TimeArrayHelper(0,6),
                    Values: []float64{1,1,2,3,5,8},
                },
                {
                    Times: TimeArrayHelper(0,6),
                    Values: []float64{10,10,20,30,50,80},
                },
            },
            number: 1,
            value: "min",
            output: []SingleData{
                {
                    Times: TimeArrayHelper(0,6),
                    Values: []float64{10,10,20,30,50,80},
                },
            },
        },
        {
            inputSd: []SingleData{
                {
                    Times: TimeArrayHelper(0,6),
                    Values: []float64{1,1,2,3,5,81},
                },
                {
                    Times: TimeArrayHelper(0,6),
                    Values: []float64{10,10,20,30,50,80},
                },
            },
            number: 1,
            value: "max",
            output: []SingleData{
                {
                    Times: TimeArrayHelper(0,6),
                    Values: []float64{1,1,2,3,5,81},
                },
            },
        },
        {
            inputSd: []SingleData{
                {
                    Times: TimeArrayHelper(0,6),
                    Values: []float64{1,1,2,3,5,8},
                },
                {
                    Times: TimeArrayHelper(0,6),
                    Values: []float64{0,10,20,30,50,80},
                },
            },
            number: 1,
            value: "absoluteMin",
            output: []SingleData{
                {
                    Times: TimeArrayHelper(0,6),
                    Values: []float64{1,1,2,3,5,8},
                },
            },
        },
        {
            inputSd: []SingleData{
                {
                    Times: TimeArrayHelper(0,6),
                    Values: []float64{1,1,2,3,5,8},
                },
                {
                    Times: TimeArrayHelper(0,6),
                    Values: []float64{10,10,20,30,50,80},
                },
            },
            number: 1,
            value: "absoluteMax",
            output: []SingleData{
                {
                    Times: TimeArrayHelper(0,6),
                    Values: []float64{10,10,20,30,50,80},
                },
            },
        },
        {
            inputSd: []SingleData{
                {
                    Times: TimeArrayHelper(0,6),
                    Values: []float64{1,1,2,3,5,8},
                },
                {
                    Times: TimeArrayHelper(0,6),
                    Values: []float64{10},
                },
            },
            number: 1,
            value: "sum",
            output: []SingleData{
                {
                    Times: TimeArrayHelper(0,6),
                    Values: []float64{1,1,2,3,5,8},
                },
            },
        },
    }
    for tdx, testCase := range tests {
        testName := fmt.Sprintf("case %d: %v", tdx, testCase.value)
        t.Run(testName, func(t *testing.T) {
            result, err := Top(testCase.inputSd, testCase.number, testCase.value)
            if err != nil {
                t.Errorf("Error not expected %v", err)
            }
            SingleDataCompareHelper(result, testCase.output, t)
        })
    }
}

func TestBottom(t *testing.T) {
    var tests = []struct{
        inputSd []SingleData
        number int
        value string
        output []SingleData
    }{
        {
            inputSd: []SingleData{
                {
                    Times: TimeArrayHelper(0,6),
                    Values: []float64{1,1,2,3,5,8},
                },
                {
                    Times: TimeArrayHelper(0,6),
                    Values: []float64{10,10,20,30,50,80},
                },
            },
            number: 1,
            value: "avg",
            output: []SingleData{
                {
                    Times: TimeArrayHelper(0,6),
                    Values: []float64{1,1,2,3,5,8},
                },
            },
        },
        {
            inputSd: []SingleData{
                {
                    Times: TimeArrayHelper(0,6),
                    Values: []float64{1,1,2,3,5,8},
                },
                {
                    Times: TimeArrayHelper(0,6),
                    Values: []float64{10,10,20,30,50,80},
                },
            },
            number: 1,
            value: "min",
            output: []SingleData{
                {
                    Times: TimeArrayHelper(0,6),
                    Values: []float64{1,1,2,3,5,8},
                },
            },
        },
        {
            inputSd: []SingleData{
                {
                    Times: TimeArrayHelper(0,6),
                    Values: []float64{1,1,2,3,5,8},
                },
                {
                    Times: TimeArrayHelper(0,6),
                    Values: []float64{10,10,20,30,50,80},
                },
            },
            number: 1,
            value: "max",
            output: []SingleData{
                {
                    Times: TimeArrayHelper(0,6),
                    Values: []float64{1,1,2,3,5,8},
                },
            },
        },
        {
            inputSd: []SingleData{
                {
                    Times: TimeArrayHelper(0,6),
                    Values: []float64{1,1,2,3,5,8},
                },
                {
                    Times: TimeArrayHelper(0,6),
                    Values: []float64{0,10,20,30,50,80},
                },
            },
            number: 1,
            value: "absoluteMin",
            output: []SingleData{
                {
                    Times: TimeArrayHelper(0,6),
                    Values: []float64{0,10,20,30,50,80},
                },
            },
        },
        {
            inputSd: []SingleData{
                {
                    Times: TimeArrayHelper(0,6),
                    Values: []float64{1,1,2,3,5,8},
                },
                {
                    Times: TimeArrayHelper(0,6),
                    Values: []float64{10,10,20,30,50,80},
                },
            },
            number: 1,
            value: "absoluteMax",
            output: []SingleData{
                {
                    Times: TimeArrayHelper(0,6),
                    Values: []float64{1,1,2,3,5,8},
                },
            },
        },
        {
            inputSd: []SingleData{
                {
                    Times: TimeArrayHelper(0,6),
                    Values: []float64{1,1,2,3,5,8},
                },
                {
                    Times: TimeArrayHelper(0,6),
                    Values: []float64{10},
                },
            },
            number: 1,
            value: "sum",
            output: []SingleData{
                {
                    Times: TimeArrayHelper(0,6),
                    Values: []float64{10},
                },
            },
        },
    }
    for tdx, testCase := range tests {
        testName := fmt.Sprintf("case %d: %v", tdx, testCase.value)
        t.Run(testName, func(t *testing.T) {
            result, err := Bottom(testCase.inputSd, testCase.number, testCase.value)
            if err != nil {
                t.Errorf("Error not expected %v", err)
            }
            SingleDataCompareHelper(result, testCase.output, t)
        })
    }
}

func TestExclude(t *testing.T) {
    var tests = []struct{
        inputSd []SingleData
        pattern string
        err bool
        output []SingleData
    }{
        {
            inputSd: []SingleData{
                {
                    Name: "Hello there",
                },
                {
                    Name: "General Kenobi!",
                },
            },
            pattern: "Hello",
            err: false,
            output: []SingleData{
                {
                    Name: "General Kenobi!",
                },
            },
        },
        {
            inputSd: []SingleData{
                {
                    Name: "this is a phrase",
                },
                {
                    Name: "a phrase containing this",
                },
            },
            pattern: "^this",
            err: false,
            output: []SingleData{
                {
                    Name: "a phrase containing this",
                },
            },
        },
        {
            inputSd: []SingleData{
                {
                    Name: "abcdef12ghi",
                },
                {
                    Name: "nonumbers",
                },
            },
            pattern: "f[1-9]*",
            err: false,
            output: []SingleData{
                {
                    Name: "nonumbers",
                },
            },
        },
        {
            inputSd: []SingleData{
                {
                    Name: "abcdef12ghi",
                },
                {
                    Name: "nonumbers",
                },
            },
            pattern: "***Hello",
            err: true,
            output: []SingleData{
                {
                    Name: "abcdef12ghi",
                },
                {
                    Name: "nonumbers",
                },
            },
        },
    }
    for tdx, testCase := range tests {
        testName := fmt.Sprintf("case %d: %v", tdx, testCase.pattern)
        t.Run(testName, func(t *testing.T) {
            result, err := Exclude(testCase.inputSd, testCase.pattern)
            if testCase.err {
                if err == nil {
                    t.Errorf("Error expected but not received %v", err)
                }
            } else {
                if err != nil {
                    t.Errorf("Error not expected %v", err)
                }
            }
            SingleDataCompareHelper(result, testCase.output, t)
        })
    }
}


// Sort Functions 

func TestSortByAvg(t *testing.T) {
    var tests = []struct{
        inputSd []SingleData
        order string
        output []SingleData
    }{
        {
            inputSd: []SingleData{
                {
                    Times: TimeArrayHelper(0,6),
                    Values: []float64{10,10,20,30,50,80},
                },
                {
                    Times: TimeArrayHelper(0,6),
                    Values: []float64{1,1,2,3,5,8},
                },
            },
            order: "asc",
            output: []SingleData{
                {
                    Times: TimeArrayHelper(0,6),
                    Values: []float64{1,1,2,3,5,8},
                },
                {
                    Times: TimeArrayHelper(0,6),
                    Values: []float64{10,10,20,30,50,80},
                },
            },
        },
        {
            inputSd: []SingleData{
                {
                    Times: TimeArrayHelper(0,6),
                    Values: []float64{1,1,2,3,5,8},
                },
                {
                    Times: TimeArrayHelper(0,6),
                    Values: []float64{10,10,20,30,50,80},
                },
            },
            order: "desc",
            output: []SingleData{
                {
                    Times: TimeArrayHelper(0,6),
                    Values: []float64{10,10,20,30,50,80},
                },
                {
                    Times: TimeArrayHelper(0,6),
                    Values: []float64{1,1,2,3,5,8},
                },
            },
        },
    }
    for tdx, testCase := range tests {
        testName := fmt.Sprintf("case %d: %v", tdx, testCase.order)
        t.Run(testName, func(t *testing.T) {
            result, err := SortByAvg(testCase.inputSd, testCase.order)
            if err != nil {
                t.Errorf("Error not expected %v", err)
            }
            SingleDataCompareHelper(result, testCase.output, t)
        })
    }
}

func TestSortByMax(t *testing.T) {
    var tests = []struct{
        inputSd []SingleData
        order string
        output []SingleData
    }{
        {
            inputSd: []SingleData{
                {
                    Times: TimeArrayHelper(0,6),
                    Values: []float64{10,10,20,30,50,80},
                },
                {
                    Times: TimeArrayHelper(0,6),
                    Values: []float64{1,1,2,3,5,81},
                },
            },
            order: "desc",
            output: []SingleData{
                {
                    Times: TimeArrayHelper(0,6),
                    Values: []float64{1,1,2,3,5,81},
                },
                {
                    Times: TimeArrayHelper(0,6),
                    Values: []float64{10,10,20,30,50,80},
                },
            },
        },
        {
            inputSd: []SingleData{
                {
                    Times: TimeArrayHelper(0,6),
                    Values: []float64{1,1,2,3,5,81},
                },
                {
                    Times: TimeArrayHelper(0,6),
                    Values: []float64{10,10,20,30,50,80},
                },
            },
            order: "asc",
            output: []SingleData{
                {
                    Times: TimeArrayHelper(0,6),
                    Values: []float64{10,10,20,30,50,80},
                },
                {
                    Times: TimeArrayHelper(0,6),
                    Values: []float64{1,1,2,3,5,81},
                },
            },
        },
    }
    for tdx, testCase := range tests {
        testName := fmt.Sprintf("case %d: %v", tdx, testCase.order)
        t.Run(testName, func(t *testing.T) {
            result, err := SortByMax(testCase.inputSd, testCase.order)
            if err != nil {
                t.Errorf("Error not expected %v", err)
            }
            SingleDataCompareHelper(result, testCase.output, t)
        })
    }
}

func TestSortByMin(t *testing.T) {
    var tests = []struct{
        inputSd []SingleData
        order string
        output []SingleData
    }{
        {
            inputSd: []SingleData{
                {
                    Times: TimeArrayHelper(0,6),
                    Values: []float64{10,10,20,30,50,80},
                },
                {
                    Times: TimeArrayHelper(0,6),
                    Values: []float64{-100,1,2,3,5,8},
                },
            },
            order: "asc",
            output: []SingleData{
                {
                    Times: TimeArrayHelper(0,6),
                    Values: []float64{-100,1,2,3,5,8},
                },
                {
                    Times: TimeArrayHelper(0,6),
                    Values: []float64{10,10,20,30,50,80},
                },
            },
        },
        {
            inputSd: []SingleData{
                {
                    Times: TimeArrayHelper(0,6),
                    Values: []float64{-100,1,2,3,5,8},
                },
                {
                    Times: TimeArrayHelper(0,6),
                    Values: []float64{10,10,20,30,50,80},
                },
            },
            order: "desc",
            output: []SingleData{
                {
                    Times: TimeArrayHelper(0,6),
                    Values: []float64{10,10,20,30,50,80},
                },
                {
                    Times: TimeArrayHelper(0,6),
                    Values: []float64{-100,1,2,3,5,8},
                },
            },
        },
    }
    for tdx, testCase := range tests {
        testName := fmt.Sprintf("case %d: %v", tdx, testCase.order)
        t.Run(testName, func(t *testing.T) {
            result, err := SortByMin(testCase.inputSd, testCase.order)
            if err != nil {
                t.Errorf("Error not expected %v", err)
            }
            SingleDataCompareHelper(result, testCase.output, t)
        })
    }
}

func TestSortBySum(t *testing.T) {
    var tests = []struct{
        inputSd []SingleData
        order string
        output []SingleData
    }{
        {
            inputSd: []SingleData{
                {
                    Times: TimeArrayHelper(0,6),
                    Values: []float64{10,10,20,30,50,80},
                },
                {
                    Times: TimeArrayHelper(0,6),
                    Values: []float64{1,1,2,3,5,8000},
                },
            },
            order: "desc",
            output: []SingleData{
                {
                    Times: TimeArrayHelper(0,6),
                    Values: []float64{1,1,2,3,5,8000},
                },
                {
                    Times: TimeArrayHelper(0,6),
                    Values: []float64{10,10,20,30,50,80},
                },
            },
        },
        {
            inputSd: []SingleData{
                {
                    Times: TimeArrayHelper(0,6),
                    Values: []float64{1,1,2,3,5,8000},
                },
                {
                    Times: TimeArrayHelper(0,6),
                    Values: []float64{10,10,20,30,50,80},
                },
            },
            order: "asc",
            output: []SingleData{
                {
                    Times: TimeArrayHelper(0,6),
                    Values: []float64{10,10,20,30,50,80},
                },
                {
                    Times: TimeArrayHelper(0,6),
                    Values: []float64{1,1,2,3,5,8000},
                },
            },
        },
    }
    for tdx, testCase := range tests {
        testName := fmt.Sprintf("case %d: %v", tdx, testCase.order)
        t.Run(testName, func(t *testing.T) {
            result, err := SortBySum(testCase.inputSd, testCase.order)
            if err != nil {
                t.Errorf("Error not expected %v", err)
            }
            SingleDataCompareHelper(result, testCase.output, t)
        })
    }
}

func TestSortByAbsMax(t *testing.T) {
    var tests = []struct{
        inputSd []SingleData
        order string
        output []SingleData
    }{
        {
            inputSd: []SingleData{
                {
                    Times: TimeArrayHelper(0,6),
                    Values: []float64{-10,-10,-20,-30,-50,-80},
                },
                {
                    Times: TimeArrayHelper(0,6),
                    Values: []float64{1,1,2,3,5,8},
                },
            },
            order: "asc",
            output: []SingleData{
                {
                    Times: TimeArrayHelper(0,6),
                    Values: []float64{1,1,2,3,5,8},
                },
                {
                    Times: TimeArrayHelper(0,6),
                    Values: []float64{-10,-10,-20,-30,-50,-80},
                },
            },
        },
        {
            inputSd: []SingleData{
                {
                    Times: TimeArrayHelper(0,6),
                    Values: []float64{1,1,2,3,5,8},
                },
                {
                    Times: TimeArrayHelper(0,6),
                    Values: []float64{-10,-10,-20,-30,-50,-80},
                },
            },
            order: "desc",
            output: []SingleData{
                {
                    Times: TimeArrayHelper(0,6),
                    Values: []float64{-10,-10,-20,-30,-50,-80},
                },
                {
                    Times: TimeArrayHelper(0,6),
                    Values: []float64{1,1,2,3,5,8},
                },
            },
        },
    }
    for tdx, testCase := range tests {
        testName := fmt.Sprintf("case %d: %v", tdx, testCase.order)
        t.Run(testName, func(t *testing.T) {
            result, err := SortByAbsMax(testCase.inputSd, testCase.order)
            if err != nil {
                t.Errorf("Error not expected %v", err)
            }
            SingleDataCompareHelper(result, testCase.output, t)
        })
    }
}

func TestSorByAbsMin(t *testing.T) {
    var tests = []struct{
        inputSd []SingleData
        order string
        output []SingleData
    }{
        {
            inputSd: []SingleData{
                {
                    Times: TimeArrayHelper(0,6),
                    Values: []float64{-10,10,20,30,50,80},
                },
                {
                    Times: TimeArrayHelper(0,6),
                    Values: []float64{1,1,2,3,5,8},
                },
            },
            order: "asc",
            output: []SingleData{
                {
                    Times: TimeArrayHelper(0,6),
                    Values: []float64{1,1,2,3,5,8},
                },
                {
                    Times: TimeArrayHelper(0,6),
                    Values: []float64{-10,10,20,30,50,80},
                },
            },
        },
        {
            inputSd: []SingleData{
                {
                    Times: TimeArrayHelper(0,6),
                    Values: []float64{1,1,2,3,5,8},
                },
                {
                    Times: TimeArrayHelper(0,6),
                    Values: []float64{-10,10,20,30,50,80},
                },
            },
            order: "desc",
            output: []SingleData{
                {
                    Times: TimeArrayHelper(0,6),
                    Values: []float64{-10,10,20,30,50,80},
                },
                {
                    Times: TimeArrayHelper(0,6),
                    Values: []float64{1,1,2,3,5,8},
                },
            },
        },
    }
    for tdx, testCase := range tests {
        testName := fmt.Sprintf("case %d: %v", tdx, testCase.order)
        t.Run(testName, func(t *testing.T) {
            result, err := SortByAbsMin(testCase.inputSd, testCase.order)
            if err != nil {
                t.Errorf("Error not expected %v", err)
            }
            SingleDataCompareHelper(result, testCase.output, t)
        })
    }
}
