package module

import (
	"github.com/jamesli0011/leaf/conf"
	"github.com/jamesli0011/leaf/log"
	"runtime"
	"sync"
)

type Module interface {
	OnInit()
	OnDestroy()
	OnSetupHandlers()
	Run(closeSig chan bool)
	AfterRun(args[] interface{})
	SignalAfterRun()
	GetId() int
	SetId(n int)
	GetName() string
	SetName(s string)
}

type module struct {
	mi       Module
	closeSig chan bool
	wg       sync.WaitGroup
}

var mods []*module

func Register(mi Module) {
	m := new(module)
	m.mi = mi
	m.closeSig = make(chan bool, 1)

	mods = append(mods, m)
}

func Init() {
	for i := 0; i < len(mods); i++ {
		mods[i].mi.OnSetupHandlers()
		log.Release("service [ %s : %d ] OnInit begin\n", mods[i].mi.GetName(), mods[i].mi.GetId())
		mods[i].mi.OnInit()
		log.Release("service [ %s : %d ] OnInit end", mods[i].mi.GetName(), mods[i].mi.GetId())
	}

	for i := 0; i < len(mods); i++ {
		m := mods[i]
		m.wg.Add(1)
		go run(m)
	}

	for i := 0; i < len(mods); i++ {
		m := mods[i]
		log.Release("service [ %s : %d ] SignalAfterRun", m.mi.GetName(), m.mi.GetId())
		go m.mi.SignalAfterRun()
	}
}

func Destroy() {
	for i := len(mods) - 1; i >= 0; i-- {
		m := mods[i]
		m.closeSig <- true
		m.wg.Wait()
		destroy(m)
	}
}

func run(m *module) {
	m.mi.Run(m.closeSig)
	m.wg.Done()
}

func destroy(m *module) {
	defer func() {
		if r := recover(); r != nil {
			if conf.LenStackBuf > 0 {
				buf := make([]byte, conf.LenStackBuf)
				l := runtime.Stack(buf, false)
				log.Error("%v: %s", r, buf[:l])
			} else {
				log.Error("%v", r)
			}
		}
	}()

	log.Debug("service [ %s : %d ] OnDestory begin", m.mi.GetName(), m.mi.GetId())
	m.mi.OnDestroy()
	log.Debug("service [ %s : %d ] OnDestory end", m.mi.GetName(), m.mi.GetId())
}
