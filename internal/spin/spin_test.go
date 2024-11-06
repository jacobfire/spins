package spin

import (
	"strconv"
	"strings"
	"testing"
)

func TestSpin_Spin(t *testing.T) {
	t.Run("check combinations", func(t *testing.T) {
		s := Spin{}
		res, comb := s.Spin()
		parts := strings.Split(comb, "-")
		slots := make([]int, 0, 3)
		var (
			temp int
			err  error
		)
		for _, v := range parts {
			temp, err = strconv.Atoi(v)
			if err != nil {
				t.Fatal("convertion error")
			}
			slots = append(slots, temp)
		}
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

		if globmax != res {
			t.Fatal("not correct")
		}
	})
}
