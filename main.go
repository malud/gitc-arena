package main

import (
	"bytes"
	"flag"
	"fmt"
	"math"
	"math/rand"
	"os"
	"time"
)

const (
	Debug_AI                      = false
	Timeout                       = true
	N                             = 2
	FirstTurnTime                 = 100
	TimeLimit                     = 50
	extra_space_between_factories = 300
	W                             = 16000
	H                             = 6500
	Min_Production_Rate           = 4
	MOVE                          = moveType(0)
	BOMB                          = moveType(1)
	INCREASE                      = moveType(2)
)

var Move_Type_String = []string{"MOVE", "BOMB", "INC"}

type Vec struct {
	x int
	y int
}

func (a Vec) Sub(b Vec) Vec {
	return Vec{a.x - b.x, a.y - b.y}
}

type Factory struct {
	owner int
	units int
	prod  int
	turns int
	r     Vec
	l     []int
}

type Troop struct {
	owner  int
	source int
	target int
	units  int
	turns  int
}

type Bomb struct {
	owner  int
	source int
	target int
	turns  int
}

type State struct {
	F        []Factory
	T        []Troop
	B        map[int]Bomb
	N_Bombs  []int
	entityId int
}

type Play struct {
	moveType moveType
	from     int
	to       int
	amount   int
}

type moveType int

type strat []Vec

func main() {
	if !flag.Parsed() {
		flag.Parse()
	}

	if *aiFlag == "" {
		flag.Usage()
		os.Exit(1)
	}

	aiOne := NewAI()
	aiTwo := NewAI()

	for {
		PlayRound([]*AI{aiOne, aiTwo})
	}

}

func generator(min int, max int) int {
	if max == 0 {
		max = 1
	}
	rndSrc := rand.NewSource(time.Now().UnixNano())
	rnd := rand.New(rndSrc)
	if min > 0 {
		return rnd.Intn(max-min) + min
	} else {
		return rnd.Intn(max)
	}
}

func PlayRound(ai []*AI) int {
	swapDistrib := UniformIntDistribution{0, 1}
	playerSwap := swapDistrib.Gen() == 1
	if playerSwap {
		// TODO swap ai
	}
	state := State{}
	rndFactoryCount := generator(7, 15)
	if rndFactoryCount%2 == 0 { // must be odd
		rndFactoryCount++
	}
	state.F = make([]Factory, rndFactoryCount)

	factoryRadius := 700
	if rndFactoryCount > 10 {
		factoryRadius -= 100
	}
	minSpaceBetweenFactories := 2 * (factoryRadius + extra_space_between_factories)

	// center Factory
	state.F[0] = Factory{0, 0, 0, 0, Vec{W / 2, H / 2}, []int{}}

	// random distribution
	xDistrib := UniformIntDistribution{(0), (W/2 - 2*factoryRadius)}
	yDistrib := UniformIntDistribution{(0), (H - 2*factoryRadius)}
	productionDistrib := UniformIntDistribution{(0), (3)}
	initialUnitsDistrib := UniformIntDistribution{(15), (30)}

	for i := 1; i < len(state.F); i += 2 {
		var r Vec
		for !ValidSpawn(r, &state, i, minSpaceBetweenFactories) {
			r = Vec{xDistrib.Gen() + factoryRadius + extra_space_between_factories, yDistrib.Gen() + factoryRadius + extra_space_between_factories}
		}
		prod := productionDistrib.Gen()
		if i == 1 {
			units := initialUnitsDistrib.Gen()
			state.F[i] = Factory{1, units, prod, 0, r, make([]int, 0)}
			state.F[i+1] = Factory{-1, units, prod, 0, Vec{W, H}.Sub(r), make([]int, 0)}
		} else {
			initialNeutralsDistrib := UniformIntDistribution{(0), (5 * prod)}
			units := initialNeutralsDistrib.Gen()
			state.F[i] = Factory{0, units, prod, 0, r, make([]int, 0)}
			state.F[i+1] = Factory{0, units, prod, 0, Vec{W, H}.Sub(r), make([]int, 0)}
		}
	}

	total_prod := 0
	for i := 0; i < len(state.F); i++ {
		total_prod += state.F[i].prod
		state.F[i].l = make([]int, len(state.F))
	}
	for i := 0; i < len(state.F); i++ {
		for j := i + 1; j < len(state.F); j++ {
			d := int((Distance(state.F[i].r, state.F[j].r) - 2*factoryRadius) / 800.0)
			state.F[i].l[j] = d
			state.F[j].l[i] = d
		}
	}
	for i := 1; total_prod < Min_Production_Rate && i < len(state.F); i++ {
		f := state.F[i]
		if f.prod < 3 {
			f.prod++
			total_prod++
		}
	}
	state.N_Bombs = []int{2, 2}
	//println("PlayGame")
	winner := Play_Game(ai, &state)
	if playerSwap {
		if winner == -1 {
			return -1
		} else if winner == 0 {
			return 1
		}
		return 0
	} else {
		return winner
	}
}

func Play_Game(ai []*AI, state *State) int {
	for i := 0; i < N; i++ {
		ss := &bytes.Buffer{}
		fmt.Fprintf(ss, "%d %d\n", len(state.F), len(state.F)*len(state.F-1)/2)
		for j := 0; j < len(state.F); j++ {
			for k := j + 1; k < len(state.F); k++ {
				fmt.Fprintf(ss, "%d %d %d\n", j, k, state.F[j].l[k])
			}
		}
		ai[i].Feed_Inputs(ss)
	}
	state.entityId = 0
	for turn := 1; turn <= 400; turn++ {
		//println("Turn:", turn)
		M := []string{"WAIT", "WAIT"}
		for i := 0; i < N; i++ {
			if ai[i].Alive() {
				if Has_Won(ai, i) {
					//cerr << i << " has won in " << turn << " turns" << endl;
					return i
				}
				color := 1
				if i != 0 {
					color = -1
				}
				ss := &bytes.Buffer{}
				fmt.Fprintf(ss, "%d %d %d\n", len(state.F), len(state.T), len(state.B))
				for j := 0; j < len(state.F); j++ {
					fmt.Fprintf(ss, "%d FACTORY %d %d %d %d %d\n", j, color*state.F[j].owner, state.F[j].units, state.F[j].prod, state.F[j].turns, 0)
				}
				for j := 0; j < len(state.T); j++ {
					fmt.Fprintf(ss, "%d TROOP %d %d %d %d %d\n", j, color*state.T[j].owner, state.T[j].source, state.T[j].target, state.T[j].units, state.T[j].turns)
				}
				//for (const auto &it:S.B){
				//const bomb &b{it.second};
				//ss << it.first << " BOMB " << color*b.owner << " " << b.source << " " << (b.owner==color?b.target:-1) << " " << (b.owner==color?b.turns:-1) << " " << 0 << endl;
				//}
				for idx, b := range state.B {
					// TODO see above
					fmt.Fprintf(ss, "%d BOMB %d %d %d %d %d", idx, color*b.owner, b.source, -1, -1, 0)
				}
				ai[i].Feed_Inputs(ss)
				M[i] = GetMove(ai[i], turn)
			}
		}
		for i := 0; i < 2; i++ {
			Play_Move(state, ai[i], M[i])
		}
		Simulate(state)
		//for i := 0; i < N; i++ {
		//	if (!Player_Alive(state, i==0?1:-1)){
		//		ai[i].stop()
		//	}
		//}
		if turn == 200 {
			Units := []int{0, 0}
			for _, f := range state.F {
				if f.owner == 1 {
					Units[0] += f.units
				} else if f.owner == -1 {
					Units[1] += f.units
				}
			}
			for _, t := range state.T {
				if t.owner == 1 {
					Units[0] += t.units
				} else if t.owner == -1 {
					Units[1] += t.units
				}
			}
			if Units[0] == Units[1] {
				//cerr << "Draw" << endl;
				return -1
			} else if Units[0] > Units[1] {
				//cerr << 0 << " has won in " << turn << " turns" << endl;
				return 0
			} else {
				//cerr << 1 << " has won in " << turn << " turns" << endl;
				return 1
			}
		}

		if All_Dead(ai) {
			os.Exit(1)
			//return -1
		}
	}
	return -2
}

func Distance(a Vec, b Vec) int {
	sqrt := math.Sqrt(math.Pow(float64(a.x-b.x), 2.0) + math.Pow(float64(a.y-b.y), 2.0))
	return int(sqrt)
}

func ValidSpawn(r Vec, state *State, id int, minSpaceBetweenFactories int) bool {
	for j := 0; j < id; j++ {
		if Distance(r, state.F[j].r) < minSpaceBetweenFactories {
			return false
		}
	}
	return true
}

func Has_Won(bot []*AI, idx int) bool {
	for i := 0; i < N; i++ {
		if i != idx && bot[i].Alive() {
			return true
		}
	}
	return false
}

func GetMove(ai *AI, turn int) string {
	buf := &bytes.Buffer{}
	duration := TimeLimit * time.Second
	if turn == 1 {
		duration = FirstTurnTime * time.Millisecond
	}
	timer := time.NewTimer(duration)
	<-timer.C
	n, err := buf.ReadFrom(ai.outPipe)
	if err != nil {
		println(err)
		os.Exit(1)
	}

	//println("AI response:", buf.String())
	if n == 0 {
		println("Loss by Timeout of AI", ai.id, ai.name)
		os.Exit(1)
	}

	return buf.String()
}

func Simulate(s *State) {
	//println("Simulate++")
}

func Play_Move(state *State, ai *AI, m string) {
	//println("PlayMove")
}

func All_Dead(ai []*AI) bool {
	for _, bot := range ai {
		if bot.Alive() {
			return false
		}
	}
	return true
}
