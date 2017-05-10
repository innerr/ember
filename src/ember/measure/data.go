package measure

import (
	"fmt"
	"io"
	"math"
	"os"
	"sort"
	"time"
	"ember/base"
)

func (p *MeasureData) PrintReadable(maxKeyLen int) (err error) {
	return p.Dump(os.Stdout, true, maxKeyLen)
}

func (p *MeasureData) Dump(w io.Writer, readable bool, maxKeyLen int) (err error) {
	for _, it := range *p {
		err = it.Dump(w, readable, maxKeyLen)
		if err != nil {
			return
		}
	}
	return
}

func (p *MeasureData) Merge(x MeasureData, prefixA string, prefixB string) MeasureData {
	data := *p
	ret := MeasureData{}
	a := 0
	b := 0
	for true {
		if a >= len(x) {
			if b >= len(x) {
				break
			} else {
				ret.AppendSpan(x[b], prefixB)
				b += 1
			}
		} else {
			if b >= len(x) {
				ret.AppendSpan(data[a], prefixA)
				a += 1
			} else {
				if data[a].Time > x[b].Time {
					ret.AppendSpan(x[b], prefixB)
					b += 1
				} else {
					ret.AppendSpan(data[a], prefixA)
					a += 1
				}
			}
		}
	}
	return ret
}

func (p *MeasureData) AppendSpan(x *SpanData, prefix string) {
	data := *p
	last := data[len(data) - 1]

	if last.Time == 0 {
		last.Time = x.Time
	}

	if last.Time < x.Time {
		n := NewSpanData()
		n.Time = x.Time
		*p = append(*p, n)
		data = *p
		last = n
	}

	for k, v := range x.Data {
		k = prefix + k
		if _, ok := last.Data[k]; !ok {
			last.Data[k] = NewSpecData()
		}
		last.Data[k].Merge(v)
	}
	return
}

func (p *MeasureData) After(time int64) MeasureData {
	data := *p
	i := len(data) - 1
	if time > data[i].Time {
		return NewMeasureData(0)
	}

	for ; i >= 0; i-- {
		if data[i].Time == time {
			break
		}
	}

	after := MeasureData(data[i:])
	return after.Copy()
}

func (p *MeasureData) Copy() MeasureData {
	result := NewMeasureData(len(*p))
	for i, it := range (*p) {
		result[i] = it.Copy()
	}
	return result
}

func (p *MeasureData) LastTime() int64 {
	return (*p)[len(*p) - 1].Time
}

func (p *MeasureData) Padding(count int) {
	data := *p
	if count >= len(data) {
		*p = NewMeasureData(len(data))
	} else {
		for i := 0; i < count; i++ {
			*p = append(*p, NewSpanData())
		}
		*p = (*p)[count:]
	}
}

func (p *MeasureData) Record(time int64, name string, value int64) {
	data := *p
	last := data[len(data) - 1]

	if last.Time == 0 {
		last.Time = time
	}

	if last.Time < time {
		n := NewSpanData()
		n.Time = time
		*p = append(*p, n)
		*p = (*p)[1:]
		data = *p
	}

	for i := len(data) - 1; i >= 0; i-- {
		if data[i].Time != time {
			continue
		}
		data[i].Record(name, value)
		return
	}

	panic("Measure.Record: unexpect")
}

func NewMeasureData(count int) MeasureData {
	data := make([]*SpanData, count)
	for i, _ := range data {
		data[i] = NewSpanData()
	}
	return data
}

type MeasureData []*SpanData

func (p *SpanData) Print(readable bool, maxKeyLen int) (err error) {
	return p.Dump(os.Stdout, readable, maxKeyLen)
}

func (p *SpanData) Dump(w io.Writer, readable bool, maxKeyLen int) (err error) {
	if p.Time == 0 {
		return
	}
	if readable {
		_, err = w.Write([]byte(fmt.Sprintf("[Time Stamp: %d (%s)]\n", p.Time / 1e9, time.Unix(0, p.Time).Format(TimeFormat))))
	} else {
		_, err = w.Write([]byte(fmt.Sprintf("@%d\n", p.Time / 1e9)))
	}
	if err != nil {
		return
	}
	keys := []string{}
	for k, _ := range p.Data {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		if readable {
			_, err = w.Write([]byte(fmt.Sprintf("%s: %s\n", base.Rpad(k, maxKeyLen), p.Data[k].Dump(readable))))
		} else {
			_, err = w.Write([]byte(fmt.Sprintf("%s %s\n", k, p.Data[k].Dump(readable))))
		}
		if err != nil {
			return
		}
	}
	return
}

func (p *SpanData) Record(name string, value int64) {
	if _, ok := p.Data[name]; !ok {
		p.Data[name] = NewSpecData()
	}
	p.Data[name].Record(value)
}

func (p *SpanData) Clear() {
	p.Time = 0
	p.Data = make(map[string]*SpecData)
}

func (p *SpanData) Copy() *SpanData {
	result := &SpanData {
		p.Time,
		make(map[string]*SpecData),
	}
	for k, v := range p.Data {
		var cp = *v
		result.Data[k] = &cp
	}
	return result
}

func NewSpanData() *SpanData {
	return &SpanData {
		0,
		make(map[string]*SpecData),
	}
}

type SpanData struct {
	Time int64
	Data map[string]*SpecData
}

func (p *SpecData) Dump(readable bool) string {
	avg := int64(0)
	if p.Count != 0 {
		avg = p.Max / p.Count
	}

	if readable {
		return fmt.Sprintf("min:%s max:%s cnt:%s avg:%s",
			base.Nkmg(p.Min, 4), base.Nkmg(p.Max, 4), base.Nkmg(p.Count, 4), base.Nkmg(avg, 4))
	} else {
		return fmt.Sprintf("%d %d %d %d", p.Min, p.Max, p.Count, avg)
	}
}

func (p *SpecData) Merge(x *SpecData) {
	p.Max = Max(p.Max, x.Max)
	p.Min = Min(p.Min, x.Min)
	p.Sum = Sum(p.Sum, x.Sum)
	p.Count = Sum(p.Count, x.Count)
}

func (p *SpecData) Record(value int64) {
	p.Max = Max(p.Max, value)
	p.Min = Min(p.Min, value)
	p.Sum = Sum(p.Sum, value)
	p.Count = Count(p.Count, value)
}

func NewSpecData() *SpecData {
	return &SpecData{0, math.MaxInt64, 0, 0}
}

type SpecData struct {
	Max, Min, Sum, Count int64
}

const TimeFormat = "2006-01-02/15:04:05"
