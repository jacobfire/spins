package spin

import (
	"fmt"
	"math/rand"
)

type Spin struct{}

func (s *Spin) Spin() (int, string) {
	slots := make([]int, 0, 3)
	slots = append(slots, rand.Intn(9))
	slots = append(slots, rand.Intn(9))
	slots = append(slots, rand.Intn(9))

	fmt.Printf("GLOBMAX: %v\n", slots)
	max := 0
	globmax := 0
	for i := 0; i < 3; i++ {
		for j := i + 1; j < 3; j++ {
			if slots[i] == slots[j] {
				max++
			}
		}
		if max > globmax {
			globmax = max
		}
	}

	return globmax, fmt.Sprintf("%d-%d-%d", slots[0], slots[1], slots[2])
}
