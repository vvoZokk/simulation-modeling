package main

import (
	//"chain"
	"fmt"
	"math/rand"
	"os"
	"sim"
	"time"
	"transaction"
)

const Points = 8
const ( // List of points
	Point0 = iota
	PointA
	PointB
	PointCw
	PointC
	PointAC
	PointBC
	ClockPoint
)
const ( // List of actions
	Generate = iota
	Wait
	Use
	Terminate
)
const ( // List of limits
	Station = iota
	AC
	BC
	Timer
)

type Checks struct {
	cur, next int
	check     bool
}

type Action struct {
	Type      int
	Arguments []int
}

func GenerateUniform(S *sim.Sim, R *rand.Rand, Limits sim.Pair, PointList []int) {
	for _, point := range PointList {
		if time, err := sim.Uniform(R, Limits); err != nil {
			fmt.Println(err)
			os.Exit(1)
		} else {
			S.Generate(time, point)
		}
	}
}

func UseBlock(S *sim.Sim, Tr *transaction.Transaction, Time float64, NextPoint int) {
	if err := S.UsePoint(Tr, Time, NextPoint); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func Phase(S *sim.Sim, R *rand.Rand, TimeTable map[int]sim.Pair, CheckTable map[transaction.Points][]int, RoadMap map[Checks][]Action) {
	cec, err := S.Extraction()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	for _, tr := range cec {
		points := transaction.GetPoints(*tr)
		check, err := S.Test(CheckTable[points])
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		actions := RoadMap[Checks{points.Current, points.Next, check}]
		for _, action := range actions {
			if action.Type == Wait {
				// GEBUG PRINT
				fmt.Println("WAIT ACTION")
			}
			if action.Type == Generate {
				// GEBUG PRINT
				fmt.Println("GENERATE ACTION")
				GenerateUniform(S, R, TimeTable[action.Arguments[0]], []int{action.Arguments[1]})
			}
			if action.Type == Use {
				// GEBUG PRINT
				fmt.Println("USE ACTION")
				switch {
				case action.Arguments[0] == 0:
					UseBlock(S, tr, 0.0, action.Arguments[1])
				default:
					if time, err := sim.Uniform(R, TimeTable[action.Arguments[0]]); err != nil {
						fmt.Println(err)
						os.Exit(1)
					} else {
						UseBlock(S, tr, time, action.Arguments[1])
					}
				}
			}
			if action.Type == Terminate {
				S.Terminate()
			}
		}
	}
}

func main() {

	// Init section

	timings := map[int]sim.Pair{
		Station: sim.Pair{35, 55},
		AC:      sim.Pair{12, 18},
		BC:      sim.Pair{17, 23},
		Timer:   sim.Pair{1440, 1440}}

	checks := map[transaction.Points][]int{
		transaction.Points{Point0, PointA}:   []int{PointAC, PointCw},
		transaction.Points{Point0, PointB}:   []int{PointBC, PointCw},
		transaction.Points{PointAC, PointC}:  []int{PointBC},
		transaction.Points{PointCw, PointBC}: []int{PointBC},
		transaction.Points{PointBC, PointC}:  []int{PointAC},
		transaction.Points{PointCw, PointAC}: []int{PointAC},
	}

	transfers := map[Checks][]Action{
		{Point0, PointA, false}:   []Action{Action{Wait, []int{}}},                                                    // >A****C****B
		{Point0, PointA, true}:    []Action{Action{Use, []int{0, PointAC}}, Action{Generate, []int{Station, PointA}}}, // >A****C****B
		{PointA, PointAC, true}:   []Action{Action{Use, []int{0, PointC}}},                                            // A->***C****B
		{PointAC, PointC, true}:   []Action{Action{Use, []int{AC, PointBC}}},                                          // A***->C****B
		{PointAC, PointC, false}:  []Action{Action{Use, []int{AC, PointCw}}},                                          // A***->W****B
		{PointC, PointCw, true}:   []Action{Action{Use, []int{0, PointBC}}},                                           // A****>W<****B
		{PointCw, PointBC, false}: []Action{Action{Wait, []int{}}},                                                    // A****>W<****B
		{PointCw, PointBC, true}:  []Action{Action{Use, []int{BC, PointB}}},                                           // A****W->***B
		{PointC, PointBC, true}:   []Action{Action{Use, []int{BC, PointB}}},                                           // A****C->***B
		{PointBC, PointB, true}:   []Action{Action{Use, []int{0, Point0}}},                                            // A****C***->B

		{Point0, PointB, false}:   []Action{Action{Wait, []int{}}},                                                    // A****C****B<
		{Point0, PointB, true}:    []Action{Action{Use, []int{0, PointBC}}, Action{Generate, []int{Station, PointB}}}, // A****C****B<
		{PointB, PointBC, true}:   []Action{Action{Use, []int{0, PointC}}},                                            // A****C***<-B
		{PointBC, PointC, true}:   []Action{Action{Use, []int{BC, PointBC}}},                                          // A****C<-***B
		{PointBC, PointC, false}:  []Action{Action{Use, []int{BC, PointCw}}},                                          // A****W<-***B
		{PointC, PointCw, true}:   []Action{Action{Use, []int{0, PointAC}}},                                           // A****>W<****B
		{PointCw, PointAC, false}: []Action{Action{Wait, []int{}}},                                                    // A****>W<****B
		{PointCw, PointAC, true}:  []Action{Action{Use, []int{AC, PointA}}},                                           // A***<-W****B
		{PointC, PointAC, true}:   []Action{Action{Use, []int{AC, PointA}}},                                           // A***<-C****B
		{PointAC, PointA, true}:   []Action{Action{Use, []int{0, Point0}}},                                            // A<-***C****B

		{Point0, ClockPoint, true}: []Action{Action{Terminate, []int{}}}, // Clock
	}

	rand := rand.New(rand.NewSource(time.Now().UnixNano()))
	CLSim := sim.New(Points)
	CLSim.Init()

	// Begin simulation

	GenerateUniform(CLSim, rand, timings[Timer], []int{Terminate})
	GenerateUniform(CLSim, rand, timings[Station], []int{PointA, PointB})
	fmt.Println(CLSim)

	Phase(CLSim, rand, timings, checks, transfers)

}
