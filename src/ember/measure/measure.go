package measure

import (
	"time"
	"sync"
)

func (p *Measure) Sync(time int64) (ret MeasureData, err error) {
	p.locker.Lock()
	defer p.locker.Unlock()
	ret = p.data.After(p.round(time))
	return
}

func (p *Measure) Dump() MeasureData {
	p.locker.Lock()
	defer p.locker.Unlock()
	return p.data.Copy()
}

func (p *Measure) Record(name string, value int64) {
	p.locker.Lock()
	defer p.locker.Unlock()

	now := p.round(time.Now().UnixNano())
	last := p.data.LastTime()
	if last < now {
		p.data.Padding(int((now - last) / p.interval))
	}
	p.data.Record(now, name, value)
}

func (p *Measure) round(time int64) int64 {
	return time / p.interval * p.interval
}

func (p *Measure) SetInterval(interval time.Duration, keep time.Duration) {
	p.interval = int64(interval)
	p.data = NewMeasureData(int(keep / interval))
}

func NewMeasure(interval time.Duration, keep time.Duration) *Measure {
	return &Measure {
		interval: int64(interval),
		data: NewMeasureData(int(keep / interval)),
	}
}

type Measure struct {
	interval int64
	data MeasureData
	locker sync.Mutex
}
