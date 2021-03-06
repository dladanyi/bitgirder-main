package service

import (
    mgRct "mingle/reactor"
)

type proxyProc struct {
    proc mgRct.EventProcessor
    dt *mgRct.DepthTracker
}

func newProxyProc( proc mgRct.EventProcessor ) *proxyProc {
    return &proxyProc{ proc, mgRct.NewDepthTracker() }
}

func ( p *proxyProc ) isDone() bool { return p.dt.Depth() == 0 }

func ( p *proxyProc ) ProcessEvent( ev mgRct.Event ) error {
    if err := p.dt.ProcessEvent( ev ); err != nil { return err }
    return p.proc.ProcessEvent( ev )
}
