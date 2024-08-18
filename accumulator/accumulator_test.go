package accumulator

import (
	"fmt"
	"testing"
	"time"
)

func TestAccu(t *testing.T) {
	accu := NewAccumulator()
	if accu == nil {
		t.Fatal("accumulator is nil")
	}
	key1 := "key1"
	i := 0
	for {
		if !accu.AllowN(key1, 1, 3, 1) {
			break
		}
		i++
	}
	if i != 3 {
		t.Fatal("i != 1")
	}
	for {
		if !accu.AllowN(key1, 1, 3, 1) {
			break
		}
		i++
	}
	if i != 3 {
		t.Fatal("i != 1")
	}
	time.Sleep(1 * time.Second)
	for {
		if !accu.AllowN(key1, 1, 3, 1) {
			break
		}
		i++
	}
	if i != 3 {
		t.Fatal("i != 1")
	}
}

func debug(sliding *sliding, currentTime, n uint64) bool {

	seconds := sliding.seconds

	sizeAlignedTime := currentTime - (currentTime % seconds)
	timeSinceStart := sizeAlignedTime - sliding.current.getStartTime()
	nSlides := timeSinceStart / seconds

	// window slide shares both current and previous windows.
	if nSlides == 1 {
		sliding.previous.setToState(sizeAlignedTime-seconds, sliding.current.count)
		sliding.current.resetToTime(sizeAlignedTime)

	} else if nSlides > 1 {
		sliding.previous.resetToTime(sizeAlignedTime - seconds)
		sliding.current.resetToTime(sizeAlignedTime)
	}

	// currentWindowBoundary := currentTime - sliding.current.getStartTime()
	b := seconds - (currentTime - sizeAlignedTime)
	_ = b
	if false {
		a := uint64((sliding.previous.count/sliding.seconds)*(seconds-(currentTime-sizeAlignedTime))) + sliding.current.count
		// fmt.Printf("[%d] %d\n", uint64(sliding.previous.count/sliding.seconds*(seconds-(currentTime-sizeAlignedTime))), sliding.current.count)
		// currentSlidingRequests := uint64(w*float64(sliding.previous.count)) + sliding.current.count
		// diff := currentTime - sizeAlignedTime
		// rate := float64(sliding.previous.count)*(float64(sliding.seconds)-float64(diff))/float64(sliding.seconds) + float64(sliding.current.count)
		// fmt.Println("rate", rate, "currentcount", sliding.current.count, "previouscount", sliding.previous.count, "diff", diff, "currentSlidingRequests", currentSlidingRequests, "n", n)
		if a+n > sliding.limit {
			return false
		}
	}
	if true {
		// Calculate the number of requests in the current sliding window
		currentWindowBoundary := currentTime - sliding.current.getStartTime()
		w := float64(sliding.seconds-currentWindowBoundary) / float64(sliding.seconds)
		currentSlidingRequests := uint64(w*float64(sliding.previous.count)) + sliding.current.count

		if currentSlidingRequests+n > sliding.limit {
			return false
		}
	}

	// add current request count to window of current count
	sliding.current.updateCount(n)
	return true
}

func TestMain(t *testing.T) {
	t.Log(float64(time.Duration(1) * time.Second))
}

func TestAccumulatorDebug(t *testing.T) {
	sliding := newSliding(3, 3)
	currentTime := uint64(1723982427)
	//当前时间运行两次
	n := uint64(1)
	for i := 0; i < 5; i++ {
		fmt.Printf("%d:%v", n, debug(sliding, currentTime, 1))
		n++
		currentTime++
	}
}

func TestAccumulator(t *testing.T) {
	sliding := newSliding(100, 3)
	currentTime := uint64(1723982427)
	for i := 0; i < 40; i++ {
		fmt.Printf("%d ", currentTime)
		ok, fail := 0, 0
		for j := 0; j < 83; j++ {
			if debug(sliding, currentTime, 1) {
				ok++
			} else {
				fail++
			}
		}
		fmt.Printf("ok:%d, fail:%d\n", ok, fail)
		currentTime++
	}
	// currentTime = currentTime + 5
	// for i := 0; i < 3; i++ {
	// 	fmt.Printf("%d ", currentTime)
	// 	ok, fail := 0, 0
	// 	for j := 0; j < 12; j++ {
	// 		if debug(sliding, currentTime, 1) {
	// 			ok++
	// 		} else {
	// 			fail++
	// 		}
	// 	}
	// 	fmt.Printf("ok:%d, fail:%d\n", ok, fail)
	// 	currentTime++
	// }

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
