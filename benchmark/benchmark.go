package benchmark

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"parser/config"
	"parser/entity"
	"parser/logger"
	"sync"
	"sync/atomic"
	"time"
)

type benchmarkStruct struct {
	lenHosts      int                   // len(hosts), for convenience
	hosts         []entity.ResponseItem // slice of hosts
	recommendNums []int32               // len(nums) = len(hosts), recommend number for each host
	t             *http.Transport       // transport
}

func GetRecommendNums(ctx context.Context, hosts []entity.ResponseItem) (jsonBt []byte, err error) {
	// create one transport for all get requests
	t := &http.Transport{}

	// create a new benchmarkStruct
	var r *benchmarkStruct
	r, err = newBenchmarkStruct(hosts, t)
	if err != nil {
		return
	}

	// calc - concurrently determines recommend number
	r.calc(ctx)

	// for fast json response
	jsonBt, err = r.toJson()
	if err != nil {
		return
	}
	return
}

func newBenchmarkStruct(hosts []entity.ResponseItem, t *http.Transport) (r *benchmarkStruct, err error) {
	if hosts == nil {
		err = fmt.Errorf("hosts is nil")
		return
	}
	lenHosts := len(hosts)
	r = &benchmarkStruct{
		lenHosts:      lenHosts,
		hosts:         hosts,
		recommendNums: make([]int32, lenHosts),
		t:             t,
	}
	return
}

func (r *benchmarkStruct) calc(ctx context.Context) {
	// run binary func for each host and wait until all ends
	wg := sync.WaitGroup{}
	wg.Add(r.lenHosts)
	for i := 0; i < r.lenHosts; i++ {
		go func(inx int) {
			defer wg.Done()
			binary(ctx, r.hosts[inx].Host, r.hosts[inx].Url, &r.recommendNums[inx], r.t)
		}(i)
	}
	wg.Wait()
}

func (r *benchmarkStruct) toJson() (jsonBt []byte, err error) {
	// for a fast json response
	var buf bytes.Buffer
	_, err = fmt.Fprintf(&buf, `{`)
	if err != nil {
		return
	}
	for i := 0; i < r.lenHosts; i++ {
		_, err = fmt.Fprintf(&buf, `"%s":"%d",`, r.hosts[i].Host, r.recommendNums[i])
		if err != nil {
			return
		}
	}
	buf.Truncate(buf.Len() - 1)
	_, err = fmt.Fprintf(&buf, `}`)
	if err != nil {
		return
	}
	jsonBt = buf.Bytes()
	return
}

func binary(ctx context.Context, host, url string, recommendNumber *int32, t *http.Transport) {
	// binary search for correct number

	var err error
	var beg, end, i int

	// loop for determining range
	//begin of range - with no mistakes, end of range - with mistakes
	i = 1
	for {
		logger.Write(fmt.Sprintf("%s with %d\n", host, i))
		err = getNRequest(ctx, i, url, t)
		if err != nil {
			end = i
			break
		} else {
			beg = i
		}
		// let's try two times more
		i *= 2
	}

	// loop for determining one number in range, binary search
	for {
		if (end - beg) < 2 {
			atomic.AddInt32(recommendNumber, int32(beg))
			break
		}
		i = (beg + end) / 2
		logger.Write(fmt.Sprintf("%s with %d\n", host, i))
		err = getNRequest(ctx, i, url, t)
		if err != nil {
			end = i
		} else {
			beg = i
		}
	}
	return
}

func getRequest(parentCtx context.Context, url string, t *http.Transport) (err error) {
	// one request function

	// create a context with configs per request timeout
	ctx, cancel := context.WithTimeout(parentCtx, config.Values().GetPerUrlTimeout()*time.Second)
	defer cancel()

	// create a request with context
	var req *http.Request
	req, err = http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		logger.Write(fmt.Sprintf("getRequest: url:%s err1:%s\n", url, err))
		return
	}

	// run get request
	var resp *http.Response
	resp, err = t.RoundTrip(req)
	if err != nil {
		logger.Write(fmt.Sprintf("getRequest: url:%s err2:%s\n", url, err))
		return
	}

	// close the response body
	err = resp.Body.Close()
	if err != nil {
		logger.Write(fmt.Sprintf("getRequest: url:%s err3:%s\n", url, err))
		return
	}

	// if status code not 200, throw an error
	if resp.StatusCode != http.StatusOK {
		logger.Write(fmt.Sprintf("getRequest: url:%s badStatusCode:%d\n", url, resp.StatusCode))
		err = fmt.Errorf("bad status code: %d", resp.StatusCode)
		return
	}
	return
}

func getNRequest(parentCtx context.Context, n int, url string, t *http.Transport) (err error) {
	// run n requests concurrently

	// create a context with cancel
	ctx, cancel := context.WithCancel(parentCtx)
	defer cancel()

	wg := sync.WaitGroup{}
	wg.Add(n)

	// variable for atomic change
	var isErrAppeared int32

	// run n requests concurrently
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			err1 := getRequest(ctx, url, t)
			if err1 != nil {
				// is error received, atomic add 1 to isErrAppeared, stop all n requests with cancel()
				atomic.AddInt32(&isErrAppeared, 1)
				cancel()
				return
			}
		}()
	}

	// wait for all requests
	wg.Wait()
	if isErrAppeared > 0 {
		// at least one error received, return an error
		err = fmt.Errorf("some error")
		return
	}
	return
}
