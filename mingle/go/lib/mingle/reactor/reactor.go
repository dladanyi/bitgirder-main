package reactor

import (
    "fmt"
    "bitgirder/objpath"
    "bitgirder/pipeline"
    "strings"
    mg "mingle"
//    "log"
)

type ReactorError struct { 
    Location objpath.PathNode
    Message string
}

func ( e *ReactorError ) Error() string { 
    return mg.FormatError( e.Location, e.Message )
}

func NewReactorError( path objpath.PathNode, msg string ) *ReactorError { 
    return &ReactorError{ Location: path, Message: msg }
}

func NewReactorErrorf( 
    path objpath.PathNode, tmpl string, args ...interface{} ) *ReactorError {

    return NewReactorError( path, fmt.Sprintf( tmpl, args... ) )
}

type Event interface {

    // may be the empty path
    GetPath() objpath.PathNode

    SetPath( path objpath.PathNode )
}

type reactorEventImpl struct {
    path objpath.PathNode
}

func ( ri *reactorEventImpl ) GetPath() objpath.PathNode { return ri.path }

func ( ri *reactorEventImpl ) SetPath( path objpath.PathNode ) { 
    ri.path = path 
}

type ValueEvent struct { 
    *reactorEventImpl
    Val mg.Value 
}

func NewValueEvent( val mg.Value ) *ValueEvent { 
    return &ValueEvent{ Val: val, reactorEventImpl: &reactorEventImpl{} } 
}

type StructStartEvent struct { 
    *reactorEventImpl
    Type *mg.QualifiedTypeName 
}

func NewStructStartEvent( typ *mg.QualifiedTypeName ) *StructStartEvent {
    return &StructStartEvent{ Type: typ, reactorEventImpl: &reactorEventImpl{} }
}

func isStructStart( ev Event ) bool {
    _, ok := ev.( *StructStartEvent )
    return ok
}

type MapStartEvent struct {
    *reactorEventImpl
}

func NewMapStartEvent() *MapStartEvent {
    return &MapStartEvent{ reactorEventImpl: &reactorEventImpl{} }
}

type FieldStartEvent struct { 
    *reactorEventImpl
    Field *mg.Identifier 
}

func NewFieldStartEvent( fld *mg.Identifier ) *FieldStartEvent {
    return &FieldStartEvent{ Field: fld, reactorEventImpl: &reactorEventImpl{} }
}

type ListStartEvent struct {
    *reactorEventImpl
    Type *mg.ListTypeReference // the element type
}

func NewListStartEvent( typ *mg.ListTypeReference ) *ListStartEvent {
    return &ListStartEvent{ 
        reactorEventImpl: &reactorEventImpl{}, 
        Type: typ,
    }
}

type EndEvent struct {
    *reactorEventImpl
}

func NewEndEvent() *EndEvent {
    return &EndEvent{ reactorEventImpl: &reactorEventImpl{} }
}

func EventToString( ev Event ) string {
    pairs := [][]string{ { "type", fmt.Sprintf( "%T", ev ) } }
    switch v := ev.( type ) {
    case *ValueEvent: 
        pairs = append( pairs, []string{ "value", mg.QuoteValue( v.Val ) } )
    case *StructStartEvent:
        pairs = append( pairs, []string{ "type", v.Type.ExternalForm() } )
    case *ListStartEvent:
        pairs = append( pairs, []string{ "type", v.Type.ExternalForm() } )
    case *FieldStartEvent:
        pairs = append( pairs, []string{ "field", v.Field.ExternalForm() } )
    }
    if p := ev.GetPath(); p != nil {
        pairs = append( pairs, []string{ "path", mg.FormatIdPath( p ) } )
    }
    elts := make( []string, len( pairs ) )
    for i, pair := range pairs { elts[ i ] = strings.Join( pair, " = " ) }
    return fmt.Sprintf( "[ %s ]", strings.Join( elts, ", " ) )
}

func CopyEvent( ev Event, withPath bool ) Event {
    var res Event
    switch v := ev.( type ) {
    case *ValueEvent: res = NewValueEvent( v.Val )
    case *ListStartEvent: res = NewListStartEvent( v.Type )
    case *MapStartEvent: res = NewMapStartEvent()
    case *StructStartEvent: res = NewStructStartEvent( v.Type )
    case *FieldStartEvent: res = NewFieldStartEvent( v.Field )
    case *EndEvent: res = NewEndEvent()
    default: panic( libErrorf( "unhandled copy target: %T", ev ) )
    }
    if withPath { res.SetPath( ev.GetPath() ) }
    return res
}

func TypeOfEvent( ev Event ) mg.TypeReference {
    switch v := ev.( type ) {
    case *ValueEvent: return mg.TypeOf( v.Val )
    case *ListStartEvent: return v.Type
    case *MapStartEvent: return mg.TypeSymbolMap
    case *StructStartEvent: return v.Type.AsAtomicType()
    }
    panic( libErrorf( "can't get type for: %T", ev ) )
}

type ReactorTopType int

const (
    ReactorTopTypeValue = ReactorTopType( iota )
    ReactorTopTypeList
    ReactorTopTypeMap 
    ReactorTopTypeStruct 
)

func ( t ReactorTopType ) String() string {
    switch t {
    case ReactorTopTypeValue: return "value"
    case ReactorTopTypeList: return "list"
    case ReactorTopTypeMap: return "map"
    case ReactorTopTypeStruct: return "struct"
    }
    panic( libErrorf( "Unhandled reactor top type: %d", t ) )
}

type EventProcessor interface { ProcessEvent( Event ) error }

type EventProcessorFunc func( ev Event ) error

func ( f EventProcessorFunc ) ProcessEvent( ev Event ) error {
    return f( ev )
}

var DiscardProcessor = EventProcessorFunc( 
    func( ev Event ) error { return nil } )

type PipelineProcessor interface {
    ProcessEvent( ev Event, rep EventProcessor ) error
}

func makePipelineReactor( 
    elt interface{}, next EventProcessor ) EventProcessor {

    var f EventProcessorFunc

    switch v := elt.( type ) {
    case PipelineProcessor:
        f = func( ev Event ) error { return v.ProcessEvent( ev, next ) }
    case EventProcessor:
        f = func( ev Event ) error { 
            if err := v.ProcessEvent( ev ); err != nil { return err }
            return next.ProcessEvent( ev )
        }
    default: panic( libErrorf( "unhandled pipeline element: %T", elt ) )
    }

    return f
}

func InitReactorPipeline( elts ...interface{} ) EventProcessor {
    pip := pipeline.NewPipeline()
    for _, elt := range elts { pip.Add( elt ) }
    var res EventProcessor = DiscardProcessor
    pip.VisitReverse( func( elt interface{} ) {
        res = makePipelineReactor( elt, res ) 
    })
    return res
}

type EventSender struct {
    Destination EventProcessor
}

func EventSenderForReactor( rep EventProcessor ) EventSender {
    return EventSender{ Destination: rep }
}

func ( es EventSender ) processEvent( ev Event ) error {
    return es.Destination.ProcessEvent( ev )
}

func ( es EventSender ) StartStruct( qn *mg.QualifiedTypeName ) error {
    return es.processEvent( NewStructStartEvent( qn ) )
}

func ( es EventSender ) StartMap() error {
    return es.processEvent( NewMapStartEvent() )
}

func ( es EventSender ) StartList( lt *mg.ListTypeReference ) error {
    return es.processEvent( NewListStartEvent( lt ) )
}

func ( es EventSender ) StartField( fld *mg.Identifier ) error {
    return es.processEvent( NewFieldStartEvent( fld ) )
}

func ( es EventSender ) Value( mv mg.Value ) error {
    return es.processEvent( NewValueEvent( mv ) )
}

func ( es EventSender ) End() error { return es.processEvent( NewEndEvent() ) }

type valueVisit struct {
    es EventSender
}

func ( vv valueVisit ) visitSymbolMapFields( m *mg.SymbolMap ) error {
    err := m.EachPairError( func( fld *mg.Identifier, val mg.Value ) error {
        if err := vv.es.StartField( fld ); err != nil { return err }
        return vv.visitValue( val )
    })
    if err != nil { return err }
    return vv.es.End()
}

func ( vv valueVisit ) visitStruct( ms *mg.Struct ) error {
    if err := vv.es.StartStruct( ms.Type ); err != nil { return err }
    return vv.visitSymbolMapFields( ms.Fields )
}

func ( vv valueVisit ) visitList( ml *mg.List ) error {
    if err := vv.es.StartList( ml.Type ); err != nil { return err }
    for _, val := range ml.Values() {
        if err := vv.visitValue( val ); err != nil { return err }
    }
    return vv.es.End()
}

func ( vv valueVisit ) visitSymbolMap( sm *mg.SymbolMap ) error {
    if err := vv.es.StartMap(); err != nil { return err }
    return vv.visitSymbolMapFields( sm )
}

func ( vv valueVisit ) visitValue( mv mg.Value ) error {
    switch v := mv.( type ) {
    case *mg.Struct: return vv.visitStruct( v )
    case *mg.SymbolMap: return vv.visitSymbolMap( v )
    case *mg.List: return vv.visitList( v )
    }
    return vv.es.Value( mv )
}

type pathSetterCaller struct {
    ps *PathSettingProcessor
    rep EventProcessor
}

func ( c pathSetterCaller ) ProcessEvent( ev Event ) error {
    return c.ps.ProcessEvent( ev, c.rep )
}

func VisitValue( mv mg.Value, rep EventProcessor ) error {
    return ( valueVisit{ es: EventSenderForReactor( rep ) } ).visitValue( mv )
}

func VisitValuePath( 
    mv mg.Value, rep EventProcessor, path objpath.PathNode ) error {

    ps := NewPathSettingProcessor()
    if path != nil { ps.SetStartPath( path ) }
    return VisitValue( mv, pathSetterCaller{ ps, rep } )
}

func isAssignableValueType( typ mg.TypeReference ) bool {
    return mg.CanAssignType( mg.TypeValue, typ )
}
