package trafficmon

import (
	"bytes"
	"fmt"
	"math"
	"strconv"
	"sync/atomic"
	"time"
	"unsafe"
)

type trafficData struct {
	logarithm bool
	startTime int64
	lastTime  int64
	lastData  int64
	ticks     int64
	interval  int // seconds
	tmp       float64
	min, max  float64
	data      []float64
}

func (d *trafficData) Init(ln, intval int) {
	d.data = make([]float64, ln)
	d.interval = intval
	d.startTime = time.Now().UnixNano()
	d.lastTime = d.startTime
	d.ticks = 1
}

func (d *trafficData) Log(n float64) float64 {
	return math.Log2(n + 1)
}

func round(x float64) int64 {
	if int64(x+0.5) == int64(x) {
		return int64(x)
	}
	return int64(x) + 1
}

func (d *trafficData) Append(data int64) {
	now := time.Now().UnixNano()
	intv := int64(d.interval) * 1e9

AGAIN:
	if now <= d.startTime+d.ticks*intv {
		d.tmp += float64(data - d.lastData)
		d.lastData = data
		d.lastTime = now
	} else {
		ds := float64(data-d.lastData) / (float64(now-d.lastTime) / 1e9) // average data per second
		rem := ds * float64(d.startTime+d.ticks*intv-d.lastTime) / 1e9   // remaining data to next tick
		// fmt.Println(d.lastData, ds, rem)
		d.tmp += rem
		d.lastData += round(rem)
		d.tmp /= float64(d.interval)

		copy(d.data[1:], d.data)
		d.data[0] = d.tmp
		d.tmp = 0

		d.lastTime = d.startTime + d.ticks*intv
		d.ticks++
		goto AGAIN
	}
}

func (d *trafficData) Range() (min float64, avg float64, max float64) {
	min, max = 1e100, 0.0

	for i := len(d.data) - 1; i >= 0; i-- {
		f := d.data[i]
		avg += f

		if f > max {
			max = f
		} else if f < min {
			min = f
		}
	}

	d.min, d.max = min, max
	avg /= float64(len(d.data))
	return
}

func (d *trafficData) Get(index int) float64 {
	f := d.data[index]
	if d.logarithm {
		f = d.Log(f)
		return f - d.Log(d.min)
	}
	return f - d.min
}

type Survey struct {
	totalSent   int64
	totalRecved int64
	latency     float64
	latencyMin  int64
	latencyMax  int64
	sent        trafficData
	recved      trafficData
}

func (s *Survey) Init(length, intval int) {
	s.sent.Init(length/intval, intval)
	s.recved.Init(length/intval, intval)
	s.latencyMin = -1
}

func (s *Survey) Send(size int64) *Survey {
	atomic.AddInt64(&s.totalSent, size)
	return s
}

func (s *Survey) Recv(size int64) *Survey {
	atomic.AddInt64(&s.totalRecved, size)
	return s
}

func (s *Survey) Latency(nsec int64) {
	const N = 2
	for {
		o := s.latency
		n := o - o/N + float64(nsec)/N

		oi, ni := *(*uint64)(unsafe.Pointer(&o)), *(*uint64)(unsafe.Pointer(&n))
		if atomic.CompareAndSwapUint64((*uint64)(unsafe.Pointer(&s.latency)), oi, ni) {
			break
		}
	}

	if nsec > s.latencyMax {
		s.latencyMax = nsec
	}

	if nsec < s.latencyMin || s.latencyMin == -1 {
		s.latencyMin = nsec
	}
}

func (s *Survey) Update() {
	s.sent.Append(s.totalSent)
	s.recved.Append(s.totalRecved)
}

func (s *Survey) Data() (int64, int64) {
	return s.totalRecved, s.totalSent
}

func (s *Survey) SVG(w, h int, logarithm bool) *bytes.Buffer {
	ret := &bytes.Buffer{}
	ret.WriteString(fmt.Sprintf("<svg xmlns=\"http://www.w3.org/2000/svg\" version=\"1.1\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"0 0 %d %d\">", w, h))
	ret.WriteString("<style>*{ font-family: \"Lucida Console\", Monaco, monospace; box-sizing: border-box; }</style>")
	ret.WriteString("<defs>")
	id := strconv.FormatInt(time.Now().Unix(), 16)
	ret.WriteString("<linearGradient id=\"traffic-" + id + "-i\" x1=\"0\" x2=\"1\" y1=\"0\" y2=\"0\"><stop offset=\"0%\" stop-color=\"white\" stop-opacity=\"0.7\"/><stop offset=\"100%\" stop-color=\"white\" stop-opacity=\"0\"/></linearGradient>")
	ret.WriteString("<clipPath id=\"traffic-" + id + "-c\"><rect width=\"100%\" height=\"100%\" fill=\"none\" stroke=\"none\"/></clipPath>")
	ret.WriteString("</defs><g clip-path=\"url(#traffic-" + id + "-c)\">")

	wTick := float64(w) / float64(len(s.sent.data)-1)
	s.sent.logarithm, s.recved.logarithm = logarithm, logarithm
	smin, savg, smax := s.sent.Range()
	rmin, ravg, rmax := s.recved.Range()
	margin := h / 10
	tick := 60 / s.sent.interval
	minutes, m := len(s.sent.data)/tick, -1

	for i := 0; i < len(s.sent.data); i += tick {
		x := float64(i) * wTick
		m++
		if m%2 == 1 {
			ret.WriteString(fmt.Sprintf("<rect x=\"%f\" y=\"0\" width=\"%f\" height=\"%d\" fill=\"#f7f8f9\"/>",
				x-wTick, wTick*float64(tick), h))
		}
		if m%5 == 0 {
			ret.WriteString(fmt.Sprintf("<text x=\"%f\" y=\"%d\" font-size=\".3em\">-%d</text>", x+1, h-2, minutes-m))
		}
	}

	polybegin := func(c string) {
		ret.WriteString("<polyline stroke=\"" + c + "\" fill=\"" + c + "\" ")
		ret.WriteString("fill-opacity=\"0.5\" stroke-width=\"0.5px\" ")
		ret.WriteString("points=\"")
	}

	if delta := smax - smin; delta > 0 {
		if logarithm {
			delta = s.sent.Log(smax) - s.sent.Log(smin)
			margin = h/2 + 1
		}

		hScale := float64(h-margin) / delta
		polybegin("#F44336")

		x := 0.0
		for i := len(s.sent.data) - 1; i >= 0; i-- {
			f := int(s.sent.Get(i) * hScale)

			if x1, x2 := x-wTick/2, x+wTick/2; logarithm {
				ret.WriteString(fmt.Sprintf("%f,%d %f,%d ", x1, h/2-f, x2, h/2-f))
			} else {
				ret.WriteString(fmt.Sprintf("%f,%d %f,%d ", x1, f, x2, f))
			}
			x += wTick
		}

		if logarithm {
			ret.WriteString(fmt.Sprintf(" %d,%d 0,%d\"/>", w, h/2, h/2))
		} else {
			ret.WriteString(fmt.Sprintf(" %d,%d %d,%d %d,%d -1,-1 -1,0\"/>", w, 0, w+1, 0, w+1, -1))
		}
	}

	if delta := rmax - rmin; delta > 0 {
		if logarithm {
			delta = s.sent.Log(rmax) - s.sent.Log(rmin)
			margin = h/2 + 1
		}

		hScale := float64(h-margin) / delta
		polybegin("#00796B")

		x := 0.0
		for i := len(s.recved.data) - 1; i >= 0; i-- {
			f := int(s.recved.Get(i) * hScale)
			ret.WriteString(fmt.Sprintf("%f,%d %f,%d ", x-wTick/2, h-f, x+wTick/2, h-f))
			x += wTick
		}

		ret.WriteString(fmt.Sprintf(" %d,%d %d,%d %d,%d -1,%d -1,%d\"/>", w, h, w+1, h, w+1, h+1, h+1, h))
	}

	ret.WriteString("<rect width=\"100%\" height=\"100%\" fill=\"url(#traffic-" + id + "-i)\"/>")

	ret.WriteString("<text font-size=\"0.33em\" style='text-shadow: 0 0 1px #ccc'>")

	format := func(f float64) string {
		if f < 10 {
			return strconv.FormatFloat(f, 'f', 3, 64)
		} else if f < 100 {
			return strconv.FormatFloat(f, 'f', 2, 64)
		}
		return strconv.FormatFloat(f, 'f', 1, 64)
	}

	sText := "<tspan fill=\"#303F9F\" x=\".4em\" dy=\"1.2em\">Lt %d ms %d ms %d ms</tspan><tspan fill=\"#F44336\" x=\".4em\" dy=\"1.2em\">Tx %s KB/s %s KB/s %s KB/s %.2f MB</tspan>"
	rText := "<tspan fill=\"#00796B\" x=\".4em\" dy=\"1.2em\">Rx %s KB/s %s KB/s %s KB/s %.2f MB</tspan>"
	if logarithm {
		rText = "<tspan y=\"50%%\" style=\"visibility:hidden\">a</tspan>" + rText
	}

	ret.WriteString(fmt.Sprintf(sText, s.latencyMin/1e6, int(s.latency/1e6), s.latencyMax/1e6,
		format(s.sent.data[0]/1024), format(savg/1024), format(smax/1024), float64(s.totalSent)/1024/1024))

	ret.WriteString(fmt.Sprintf(rText, format(s.recved.data[0]/1024), format(ravg/1024), format(rmax/1024), float64(s.totalRecved)/1024/1024))

	ret.WriteString("</text>")

	ret.WriteString(fmt.Sprintf("<polyline stroke-width=\"1px\" stroke=\"#d1d2d3\" fill=\"none\" points=\"0,0 %d,%d %d,%d %d,%d 0,0\"/>",
		w, 0, w, h, 0, h))

	if logarithm {
		ret.WriteString("<line x1=\"0\" y1=\"50%\" x2=\"100%\" y2=\"50%\" stroke-width=\"0.5px\" stroke=\"#d7d8d9\"/>")
	}

	ret.WriteString("</g></svg>")
	return ret
}
