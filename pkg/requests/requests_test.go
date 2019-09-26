package requests

import (
	"strconv"
	"sync"
	"testing"
)

func TestGet(t *testing.T) {
	req := New()

	wg := &sync.WaitGroup{}
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			result := req.Get("http://httpbin.org/cookies/set").
				Params(Params{"id": strconv.Itoa(id)}).Session(true).Send()
			if result.Err != nil {
				text, _ := result.Text()
				t.Log(text)
			}
		}(i)
	}

	wg.Wait()
}
