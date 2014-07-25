package types

import ( 
    mg "mingle"
    "mingle/bind"
    mgRct "mingle/reactor"
)

func ( t *CastReactorTest ) Call( c *mgRct.ReactorTestCall ) {
    rcts := []interface{}{}
    if p := t.Path; p != nil {
        rcts = append( rcts, mgRct.NewPathSettingProcessorPath( p ) )
    }
    rcts = append( rcts, mgRct.NewDebugReactor( c ) )
    rcts = append( rcts, NewCastReactor( t.Type, t.Map ) )
    vb := mgRct.NewValueBuilder()
    rcts = append( rcts, vb )
    pip := mgRct.InitReactorPipeline( rcts... )
    c.Logf( "casting as %s: %s", t.Type, mg.QuoteValue( t.In ) )
    if err := mgRct.VisitValue( t.In, pip ); err == nil {
        mgRct.CheckNoError( t.Err, c )
        c.Logf( "got %s, expect %s", mg.QuoteValue( vb.GetValue() ),
            mg.QuoteValue( t.Expect ) )
        mg.AssertEqualValues( t.Expect, vb.GetValue(), c )
    } else { 
        cae := mg.CastErrorAssert{ 
            ErrExpect: t.Err, ErrAct: err, PathAsserter: c.PathAsserter }
        cae.Call()
    }
}

func ( t *EventPathTest ) Call( c *mgRct.ReactorTestCall ) {
    chk := mgRct.NewEventPathCheckReactor( t.Expect, c.PathAsserter )
    rct := NewCastReactor( t.Type, t.Map )
    pip := mgRct.InitReactorPipeline( rct, chk )
    mgRct.AssertFeedSource( t.Source, pip, c )
    chk.Complete()
}

func binderFactoryForTest( t *BuiltinTypeTest ) bind.BinderFactory {
    switch {
    case t.Type.Equals( mg.TypeIdentifier ): return IdentifierBinderFactory
    }
    panic( libErrorf( "unhandled built in type: %s", t.Type ) )
}

func ( t *BuiltinTypeTest ) Call( c *mgRct.ReactorTestCall ) {
    c.Logf( "expcting %s as type: %s", mg.QuoteValue( t.In ), t.Type )
    br := bind.NewBindReactor( binderFactoryForTest( t ) )
    cr := NewCastReactor( t.Type, V1Types() )
    pip := mgRct.InitReactorPipeline( cr, mgRct.NewDebugReactor( c ), br )
    if err := mgRct.VisitValue( t.In, pip ); err == nil {
        c.Equal( t.Expect, br.GetValue() )
    } else { c.EqualErrors( t.Err, err ) }
}
