package parser

import (
    "bitgirder/assert"
    "bytes"
    mg "mingle"
//    "log"
)

func id( strs ...string ) *mg.Identifier {
    return mg.NewIdentifierUnsafe( strs )
}

func ws( str string ) WhitespaceToken { return WhitespaceToken( str ) }

var makeTypeName = mg.NewDeclaredTypeNameUnsafe

type ParseErrorExpect struct {
    Col int
    Message string
}

func AssertParseError(
    err error, errExpct *ParseErrorExpect, a *assert.PathAsserter ) {

    pErr, ok := err.( *ParseError)
    if ! ok { a.Fatal( err ) }
    a.Descend( "Message" ).Equal( errExpct.Message, pErr.Message )
    aLoc := a.Descend( "Loc" )
    aLoc.Descend( "Col" ).Equal( errExpct.Col, pErr.Loc.Col )
    aLoc.Descend( "Line" ).Equal( 1, pErr.Loc.Line )
    aLoc.Descend( "Source" ).Equal( ParseSourceInput, pErr.Loc.Source )
}

func newTestLexer( in string, strip bool ) *Lexer {
    return New(
        &Options{
            Reader: bytes.NewBufferString( in ),
            SourceName: ParseSourceInput,
            Strip: strip,
        },
    )
}

func assertRegexRestriction( 
    expct *RegexRestrictionSyntax, 
    rs RestrictionSyntax,
    a *assert.PathAsserter ) {

    act, ok := rs.( *RegexRestrictionSyntax )
    a.Truef( ok, "not a regex restriction: %T", rs )
    a.Descend( "Loc" ).Equal( expct.Loc, act.Loc )
}

func assertRangeRestriction(
    expct *RangeRestrictionSyntax,
    rs RestrictionSyntax,
    a *assert.PathAsserter ) {

    act, ok := rs.( *RangeRestrictionSyntax )
    a.Truef( ok, "not a range restriction: %T", rs )
    a.Descend( "Loc" ).Equal( expct.Loc, act.Loc )
    a.Descend( "LeftClosed" ).Equal( expct.LeftClosed, act.LeftClosed )
    assertRestriction( expct.Left, act.Left, a.Descend( "Left" ) )
    assertRestriction( expct.Right, act.Right, a.Descend( "Right" ) )
    a.Descend( "RightClosed" ).Equal( expct.RightClosed, act.RightClosed )
}

func assertNumRestriction( 
    expct *NumRestrictionSyntax,
    rs RestrictionSyntax,
    a *assert.PathAsserter ) {

    act, ok := rs.( *NumRestrictionSyntax )
    a.Truef( ok, "not a num restriction: %T", rs )
    a.Descend( "IsNeg" ).Equal( expct.IsNeg, act.IsNeg )
    a.Descend( "Num" ).Equal( expct.Num, act.Num )
    a.Descend( "Loc" ).Equal( expct.Loc, act.Loc )
}

func assertRestriction( expct, act RestrictionSyntax, a *assert.PathAsserter ) {
    if expct == nil {
        a.Truef( act == nil, "got non-nil restriction: %s", act )
        return
    }
    switch v := expct.( type ) {
    case *RegexRestrictionSyntax: assertRegexRestriction( v, act, a )
    case *RangeRestrictionSyntax: assertRangeRestriction( v, act, a )
    case *NumRestrictionSyntax: assertNumRestriction( v, act, a )
    default: a.Fatalf( "unhandled restriction: %T", expct )
    }
}

func AssertCompletableTypeReference(
    expct, act *CompletableTypeReference, a *assert.PathAsserter ) {

    if expct == nil {
        a.Truef( act == nil, "expected nil, got %s", act )
        return
    }
    a.Descend( "ErrLoc" ).Equalf( expct.ErrLoc, act.ErrLoc,
        "%s != %s", expct.ErrLoc, act.ErrLoc )
    a.Descend( "Name" ).Equal( expct.Name, act.Name )
    assertRestriction( 
        expct.Restriction, act.Restriction, a.Descend( "Restriction" ) )
    a.Descend( "ptrDepth" ).Equal( expct.ptrDepth, act.ptrDepth )
    a.Descend( "quants" ).Equal( expct.quants, act.quants )
}

func MustIdentifier( s string ) *mg.Identifier {
    id, err := ParseIdentifier( s )
    if err == nil { return id }
    panic( err )
}

func MustNamespace( s string ) *mg.Namespace {
    ns, err := ParseNamespace( s )
    if err == nil { return ns }
    panic( err )
}

func MustDeclaredTypeName( s string ) *mg.DeclaredTypeName {
    nm, err := ParseDeclaredTypeName( s )
    if err == nil { return nm }
    panic( err )
}

func MustQualifiedTypeName( s string ) *mg.QualifiedTypeName {
    qn, err := ParseQualifiedTypeName( s )
    if err == nil { return qn }
    panic( err )
}

type unsafeTypeCompleter struct {}

func ( tc *unsafeTypeCompleter ) resolveName( 
    nm mg.TypeName ) *mg.QualifiedTypeName {

    if qn, ok := nm.( *mg.QualifiedTypeName ); ok { return qn }
    return nm.( *mg.DeclaredTypeName ).ResolveIn( mg.CoreNsV1 )
}

func ( tc *unsafeTypeCompleter ) setStringRestriction(
    at *mg.AtomicTypeReference, rx *RegexRestrictionSyntax ) {

    at.Restriction = mg.MustRegexRestriction( rx.Pat )
}

func ( tc *unsafeTypeCompleter ) setRangeValue(
    valPtr *mg.Value,
    qn *mg.QualifiedTypeName,
    rx RestrictionSyntax ) {

    sx, _ := rx.( *StringRestrictionSyntax )
    nx, _ := rx.( *NumRestrictionSyntax )
    switch {
    case mg.IsNumericTypeName( qn ):
        if num, err := mg.ParseNumber( nx.LiteralString(), qn ); err == nil {
            *valPtr = num
        } else { panic( err ) }
    case qn.Equals( mg.QnameTimestamp ):
        if tm, err := ParseTimestamp( sx.Str ); err == nil {
            *valPtr = tm
        } else { panic( err ) }
    case qn.Equals( mg.QnameString ): *valPtr = mg.String( sx.Str )
    default: panic( libErrorf( "unhandled range type: %s", qn ) )
    }
}

func ( tc *unsafeTypeCompleter ) setRangeRestriction(
    at *mg.AtomicTypeReference, rx *RangeRestrictionSyntax ) {

    rng := &mg.RangeRestriction{}
    rng.MinClosed = rx.LeftClosed
    if l := rx.Left; l != nil { tc.setRangeValue( &( rng.Min ), at.Name, l ) }
    if r := rx.Right; r != nil { tc.setRangeValue( &( rng.Max ), at.Name, r ) }
    rng.MaxClosed = rx.RightClosed

    at.Restriction = rng
}

func ( tc *unsafeTypeCompleter ) setRestriction(
    at *mg.AtomicTypeReference, rx RestrictionSyntax ) {

    if at.Name.Equals( mg.QnameString ) {
        if regx, ok := rx.( *RegexRestrictionSyntax ); ok {
            tc.setStringRestriction( at, regx )
            return
        }
    }
    tc.setRangeRestriction( at, rx.( *RangeRestrictionSyntax ) )
}

func ( tc *unsafeTypeCompleter ) CompleteBaseType(
    nm mg.TypeName, 
    rx RestrictionSyntax, 
    errLoc *Location ) ( mg.TypeReference, bool, error ) {

    at := &mg.AtomicTypeReference{ Name: tc.resolveName( nm ) }
    if rx != nil { tc.setRestriction( at, rx ) }
    return at, true, nil
}

func MustTypeReference( s string ) mg.TypeReference {
    ct, err := ParseTypeReference( s )
    if err != nil { panic( err ) }
    res, err := ct.CompleteType( &unsafeTypeCompleter{} )
    if err != nil { panic( err ) }
    return res
}

func MustTimestamp( s string ) mg.Timestamp {
    tm, err := ParseTimestamp( s )
    if err == nil { return tm }
    panic( err )
}

func mustAsQname( val interface{} ) *mg.QualifiedTypeName {
    if qn, ok := val.( *mg.QualifiedTypeName ); ok { return qn }
    return MustQualifiedTypeName( val.( string ) )
}

func mustAsId( val interface{} ) *mg.Identifier {
    if id, ok := val.( *mg.Identifier ); ok { return id }
    return MustIdentifier( val.( string ) )
}

func mustMapPairs( pairs []interface{} ) []interface{} {
    res := make( []interface{}, len( pairs ) )
    for i, e := 0, len( res ); i < e; i += 2 {
        res[ i ] = mustAsId( pairs[ i ] )
        res[ i + 1 ] = pairs[ i + 1 ]
    }
    return res
}

func MustSymbolMap( pairs ...interface{} ) *mg.SymbolMap {
    return mg.MustSymbolMap( mustMapPairs( pairs )... )
}

func MustStruct( typ interface{}, pairs ...interface{} ) *mg.Struct {
    return mg.MustStruct( mustAsQname( typ ), mustMapPairs( pairs )... )
}

func MustEnum( typ interface{}, val interface{} ) *mg.Enum {
    return &mg.Enum{ Type: mustAsQname( typ ), Value: mustAsId( val ) }
}
