package accumulator

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestAccu(t *testing.T) {
	r := require.New(t)
	accu := NewAccumulator()
	if accu == nil {
		t.Fatal("accumulator is nil")
	}
	key1 := "key1"
	i := 0
	for {
		if !accu.AllowN(key1, 1, 1, time.Second) {
			break
		}
		time.Sleep(100 * time.Millisecond)
		i++
	}
	r.Equal(1, i)
	for {
		if !accu.AllowN(key1, 1, 1, time.Second) {
			break
		}
		i++
	}
	r.Equal(1, i)
	time.Sleep(1 * time.Second)
	for {
		if !accu.AllowN(key1, 1, 1, time.Second) {
			break
		}
		i++
	}
	r.Equal(2, i)
}

func TestMain(t *testing.T) {
	t.Log(float64(time.Duration(1) * time.Second))
}

func TestAccumulator(t *testing.T) {
	key := "key1"
	succ := 0
	for i := 0; i < 10; i++ {
		ok, fail := 0, 0
		for j := 0; j < 200; j++ {
			if Allow(key, 240, 3*time.Second) {
				ok++
				succ++
			} else {
				fail++
			}
			time.Sleep(5 * time.Millisecond)
		}
		fmt.Printf("%s ok:%d, fail:%d\n", time.Now(), ok, fail)
		// time.Sleep(time.Second)
	}
	fmt.Println("succ", succ)
}

func TestMaina(t *testing.T) {
	currentTime := 1723974060
	lastTime := 1723974030
	seconds := 30
	a := time.Unix(int64(currentTime), 0)
	t.Log(a, a.Truncate(30*time.Second))

	sizeAlignedTime := currentTime - (currentTime % seconds)
	timeSinceStart := sizeAlignedTime - lastTime
	nSlides := timeSinceStart / seconds
	fmt.Println("sizeAlignedTime", sizeAlignedTime, "timeSinceStart", timeSinceStart, "nSlides", nSlides)
}

func BenchmarkXxx(b *testing.B) {
	b.ReportAllocs()
	accu := NewAccumulator()
	key1 := "key1"
	for i := 0; i < b.N; i++ {
		accu.AllowN(key1, 1, 3, 1)
	}
}
