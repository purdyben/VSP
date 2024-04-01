package vsp

import (
	"errors"
	"math"

	"gopkg.in/karalabe/cookiejar.v2/collections/stack"
)

// including the return to (0,0), must be less than 12*60
const MaxDistance = float64(720)

var ThreadholdErr error = errors.New("threshold reached")

// Idea 1
// Vary simple idea, Fron a given point pick the closes load,
// This solution maximizes the drivers capacity as if the closes node cannot be added return to depo.
//
// This does nto guarantee using all the nodes
func OptimizeClosetPath(loads []Load) ([]Load, []Load) {
	// finalLoads := [][]Load{}
	// 1. get the current toal distence
	// 2. add the closest node
	// 3 check if we can
	// 	idea look ahead if we need to return back to depo
	// 	if all nodes are not used create a new cluster and start again

	path := []Load{}
	currDis := float64(0)
	currPoint := startnode // <- current starting point is 0,0
	// fmt.Println(loads)
	for len(loads) > 0 {
		loads = Sort(currPoint, loads) //<- sort by closes node first
		if len(loads) == 0 {
			break
		}
		nextload := loads[0]

		loads = loads[1:]
		err := TestNextLoad(currPoint, currDis, nextload)
		if err != nil {
			loads = append(loads, nextload)
			// fmt.Println("TestNextLoad failed", loads)
			break
		}

		path = append(path, nextload)
		currDis += LookAheadDist(currPoint, nextload)
		currPoint = nextload.Dropoff
	}
	return path, loads
}

// Idea 2 Fail Result Is Worse
// // 1. Go to the farthest nodes first and work backwards.
func OptimizeFurthestPath(loads []Load) ([]Load, []Load) {
	s := stack.New()
	for i := 0; i < len(loads); i++ {
		s.Push(loads[i])
	}
	path := []Load{}
	currDis := float64(0)
	currPoint := startnode // <- current starting point is 0,0
	// fmt.Println(loads)

	// Get furthest first
	loads = Sort(currPoint, loads) //<- sort by closes node first
	nextload := loads[len(loads)-1]

	loads = loads[:len(loads)-1]
	err := TestNextLoad(currPoint, currDis, nextload)
	if err != nil {
		loads = append(loads, nextload)
	}
	path = append(path, nextload)
	currDis += LookAheadDist(currPoint, nextload)
	currPoint = nextload.Dropoff

	for !s.Empty() {
		nextload := s.Pop().(Load)
		// nextload := loads[0]

		// loads = loads[1:]

		err := TestNextLoad(currPoint, currDis, nextload)
		if err != nil {
			// loads = append(loads, nextload)
			break
		}

		path = append(path, nextload)
		currDis += LookAheadDist(currPoint, nextload)
		currPoint = nextload.Dropoff
	}
	return path, loads
}

func LookAheadDist(curr Point, l Load) float64 {
	dispickup := EuclideanDistance(curr, l.Pickup)
	disLoad := l.GetDistance()
	return dispickup + disLoad
}

// currDistance includes starting point,
func TestNextLoad(curr Point, currDistance float64, l Load) error {
	dispickup := EuclideanDistance(curr, l.Pickup)
	disLoad := l.GetDistance()

	// Distence to return to the depo
	if currDistance+dispickup+disLoad+DistanceFromDepo(l.Dropoff) > MaxDistance {
		return ThreadholdErr
	}
	return nil
}

type Bucket struct {
	Loads     []Load
	currDis   float64
	currPoint Point
}

// Idea 3 Fail
// Given a Load adds it to a bucket which returns the lost cost,
func Greedy(driverNum int, loads []Load) ([][]Load, error) {
	s := stack.New()
	for i := 0; i < len(loads); i++ {
		s.Push(loads[i])
	}

	res := [][]Load{}
	buckets := []*Bucket{}
	for range driverNum {
		buckets = append(buckets, &Bucket{
			Loads:     []Load{},
			currPoint: startnode,
		})
	}

	for !s.Empty() {
		nextload := s.Pop().(Load)
		selectedBucket := -1
		selectedBucketDistence := math.MaxFloat64

		for i := range buckets {

			b := buckets[i]
			dis := GetDistanceWithNextNode(b.currDis, b.currPoint, nextload)
			if dis+DistanceFromDepo(nextload.Dropoff) > MaxDistance {
				continue
			}
			if dis < selectedBucketDistence {
				selectedBucketDistence = dis
				selectedBucket = i
			}
		}
		if selectedBucket == -1 {
			return nil, errors.New("unable to proceed with this number of buckets")
		}

		b := buckets[selectedBucket]
		b.Loads = append(b.Loads, nextload)
		b.currDis = GetDistanceWithNextNode(b.currDis, b.currPoint, nextload)
		b.currPoint = nextload.Dropoff
	}

	for i := range buckets {
		res = append(res, buckets[i].Loads)
	}
	return res, nil
}

func GetDistanceWithNextNode(currDis float64, curr Point, l Load) float64 {
	dispickup := EuclideanDistance(curr, l.Pickup)
	disLoad := l.GetDistance()
	return currDis + dispickup + disLoad
}

// Idea 4 Fail
// Did not improve results under 8 loads in a row.
func OptimizePath(loads []Load) []Load {
	// Optimize path via permutations
	opLoads := loads
	newpaths := permutations(loads)
	cost := math.MaxFloat64
	for _, p := range newpaths {
		path := ToPath(p)
		path = AddStartAndEndPoints(path)
		if c := TotalCost(1, CalcTotalDistance(path)); c < cost {
			opLoads = p
			cost = c
		}
	}
	return opLoads
}

func permutations(arr []Load) [][]Load {
	var helper func([]Load, int)
	res := [][]Load{}

	helper = func(arr []Load, n int) {
		if n == 1 {
			tmp := make([]Load, len(arr))
			copy(tmp, arr)
			res = append(res, tmp)
		} else {
			for i := 0; i < n; i++ {
				helper(arr, n-1)
				if n%2 == 1 {
					tmp := arr[i]
					arr[i] = arr[n-1]
					arr[n-1] = tmp
				} else {
					tmp := arr[0]
					arr[0] = arr[n-1]
					arr[n-1] = tmp
				}
			}
		}
	}
	helper(arr, len(arr))
	return res
}