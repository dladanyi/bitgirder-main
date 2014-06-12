package mingle

import (
    "strings"
    "encoding/base64"
    "bitgirder/objpath"
    "fmt"
)

func strToBool( s String, path objpath.PathNode ) ( Value, error ) {
    switch lc := strings.ToLower( string( s ) ); lc { 
    case "true": return Boolean( true ), nil
    case "false": return Boolean( false ), nil
    }
    errTmpl :="Invalid boolean value: %s"
    errStr := QuoteValue( s )
    return nil, NewValueCastErrorf( path, errTmpl, errStr )
}

func castBoolean( 
    mgVal Value,
    at *AtomicTypeReference,
    callTyp TypeReference,
    path objpath.PathNode ) ( Value, error ) {

    switch v := mgVal.( type ) {
    case Boolean: return v, nil
    case String: return strToBool( v, path )
    }
    return nil, NewTypeCastErrorValue( callTyp, mgVal, path )
}

func castBuffer( 
    mgVal Value,
    at *AtomicTypeReference,
    callTyp TypeReference,
    path objpath.PathNode ) ( Value, error ) {

    switch v := mgVal.( type ) {
    case Buffer: return v, nil
    case String: 
        buf, err := base64.StdEncoding.DecodeString( string( v ) )
        if err == nil { return Buffer( buf ), nil }
        msg := "Invalid base64 string: %s"
        return nil, NewValueCastErrorf( path, msg, err.Error() )
    }
    return nil, NewTypeCastErrorValue( callTyp, mgVal, path )
}

func castString( 
    mgVal Value,
    at *AtomicTypeReference,
    callTyp TypeReference,
    path objpath.PathNode ) ( Value, error ) {

    switch v := mgVal.( type ) {
    case String: return mgVal, nil
    case Boolean, Int32, Int64, Uint32, Uint64, Float32, Float64:
        return String( v.( fmt.Stringer ).String() ), nil
    case Timestamp: return String( v.Rfc3339Nano() ), nil
    case Buffer:
        return String( base64.StdEncoding.EncodeToString( []byte( v ) ) ), nil
    case *Enum: return String( v.Value.ExternalForm() ), nil
    }
    return nil, NewTypeCastErrorValue( callTyp, mgVal, path )
}

func isDecimalNumString( s String ) bool {
    return strings.IndexAny( string( s ), "eE." ) >= 0
}

func parseNumberForCast( 
    s String, 
    numTyp *QualifiedTypeName, 
    path objpath.PathNode ) ( Value, error ) {

    asFloat := IsIntegerTypeName( numTyp ) && isDecimalNumString( s )
    parseTyp := numTyp
    if asFloat { parseTyp = QnameFloat64 }
    val, err := ParseNumber( string( s ), parseTyp )
    if ne, ok := err.( *NumberFormatError ); ok {
        err = NewValueCastError( path, ne.Error() )
    }
    if err != nil || ( ! asFloat ) { return val, err }
    f64 := float64( val.( Float64 ) )
    switch {
    case numTyp.Equals( QnameInt32 ): val = Int32( int32( f64 ) )
    case numTyp.Equals( QnameUint32 ): val = Uint32( uint32( f64 ) )
    case numTyp.Equals( QnameInt64 ): val = Int64( int64( f64 ) )
    case numTyp.Equals( QnameUint64 ): val = Uint64( uint64( f64 ) )
    }
    return val, nil
}

func castInt32( 
    mgVal Value,
    at *AtomicTypeReference,
    callTyp TypeReference,
    path objpath.PathNode ) ( Value, error ) {

    switch v := mgVal.( type ) {
    case Int32: return v, nil
    case Int64: return Int32( v ), nil
    case Uint32: return Int32( int32( v ) ), nil
    case Uint64: return Int32( int32( v ) ), nil
    case Float32: return Int32( int32( v ) ), nil
    case Float64: return Int32( int32( v ) ), nil
    case String: return parseNumberForCast( v, QnameInt32, path )
    }
    return nil, NewTypeCastErrorValue( callTyp, mgVal, path )
}

func castInt64( 
    mgVal Value,
    at *AtomicTypeReference,
    callTyp TypeReference,
    path objpath.PathNode ) ( Value, error ) {

    switch v := mgVal.( type ) {
    case Int32: return Int64( v ), nil
    case Int64: return v, nil
    case Uint32: return Int64( int64( v ) ), nil
    case Uint64: return Int64( int64( v ) ), nil
    case Float32: return Int64( int64( v ) ), nil
    case Float64: return Int64( int64( v ) ), nil
    case String: return parseNumberForCast( v, QnameInt64, path )
    }
    return nil, NewTypeCastErrorValue( callTyp, mgVal, path )
}

func castUint32(
    mgVal Value,
    at *AtomicTypeReference,
    callTyp TypeReference,
    path objpath.PathNode ) ( Value, error ) {

    switch v := mgVal.( type ) {
    case Int32: return Uint32( uint32( v ) ), nil
    case Uint32: return v, nil
    case Int64: return Uint32( uint32( v ) ), nil
    case Uint64: return Uint32( uint32( v ) ), nil
    case Float32: return Uint32( uint32( v ) ), nil
    case Float64: return Uint32( uint32( v ) ), nil
    case String: return parseNumberForCast( v, QnameUint32, path )
    }
    return nil, NewTypeCastErrorValue( callTyp, mgVal, path )
}

func castUint64(
    mgVal Value,
    at *AtomicTypeReference,
    callTyp TypeReference,
    path objpath.PathNode ) ( Value, error ) {

    switch v := mgVal.( type ) {
    case Int32: return Uint64( uint64( v ) ), nil
    case Uint32: return Uint64( uint64( v ) ), nil
    case Int64: return Uint64( uint64( v ) ), nil
    case Uint64: return v, nil
    case Float32: return Uint64( uint64( v ) ), nil
    case Float64: return Uint64( uint64( v ) ), nil
    case String: return parseNumberForCast( v, QnameUint64, path )
    }
    return nil, NewTypeCastErrorValue( callTyp, mgVal, path )
}

func castFloat32( 
    mgVal Value,
    at *AtomicTypeReference,
    callTyp TypeReference,
    path objpath.PathNode ) ( Value, error ) {

    switch v := mgVal.( type ) {
    case Int32: return Float32( float32( v ) ), nil
    case Int64: return Float32( float32( v ) ), nil
    case Uint32: return Float32( float32( v ) ), nil
    case Uint64: return Float32( float32( v ) ), nil
    case Float32: return v, nil
    case Float64: return Float32( float32( v ) ), nil
    case String: return parseNumberForCast( v, QnameFloat32, path )
    }
    return nil, NewTypeCastErrorValue( callTyp, mgVal, path )
}

func castFloat64( 
    mgVal Value,
    at *AtomicTypeReference,
    callTyp TypeReference,
    path objpath.PathNode ) ( Value, error ) {

    switch v := mgVal.( type ) {
    case Int32: return Float64( float64( v ) ), nil
    case Int64: return Float64( float64( v ) ), nil
    case Uint32: return Float64( float64( v ) ), nil
    case Uint64: return Float64( float64( v ) ), nil
    case Float32: return Float64( float64( v ) ), nil
    case Float64: return v, nil
    case String: return parseNumberForCast( v, QnameFloat64, path )
    }
    return nil, NewTypeCastErrorValue( callTyp, mgVal, path )
}

func castTimestamp( 
    mgVal Value,
    at *AtomicTypeReference,
    callTyp TypeReference,
    path objpath.PathNode ) ( Value, error ) {

    switch v := mgVal.( type ) {
    case Timestamp: return v, nil
    case String:
//        tm, err := ParseTimestamp( string( v ) )
//        if err == nil { return tm, nil }
//        msg := "Invalid timestamp: %s"
//        return nil, NewValueCastErrorf( path, msg, err.Error() )
        return nil, NewValueCastError( path, "castTimestamp() unimplemented" )
    }
    return nil, NewTypeCastErrorValue( callTyp, mgVal, path )
}

func castEnum( 
    mgVal Value,
    at *AtomicTypeReference,
    callTyp TypeReference,
    path objpath.PathNode ) ( Value, error ) {

    switch v := mgVal.( type ) {
    case *Enum: if v.Type.Equals( at.Name ) { return v, nil }
    }
    return nil, NewTypeCastErrorValue( callTyp, mgVal, path )
}

func castSymbolMap( 
    mgVal Value,
    at *AtomicTypeReference,
    callTyp TypeReference,
    path objpath.PathNode ) ( Value, error ) {

    switch v := mgVal.( type ) {
    case *SymbolMap: return v, nil
    }
    return nil, NewTypeCastErrorValue( callTyp, mgVal, path )
}

func NewNullValueCastError( path objpath.PathNode ) *ValueCastError {
    return NewValueCastErrorf( path, "Value is null" )
}

// switch compares based on qname not at itself since we may be dealing with
// restriction types, meaning that if at is mingle:core@v1/String~"a", it is a
// string (has qname mingle:core@v1/String) but will not equal TypeString itself
func castAtomicUnrestricted(
    mgVal Value, 
    at *AtomicTypeReference, 
    callTyp TypeReference,
    path objpath.PathNode ) ( Value, error ) {

    if _, ok := mgVal.( *Null ); ok {
        if at.Equals( TypeNull ) { return mgVal, nil }
        return nil, NewNullValueCastError( path )
    }
    switch nm := at.Name; {
    case nm.Equals( QnameValue ): return mgVal, nil
    case nm.Equals( QnameBoolean ): 
        return castBoolean( mgVal, at, callTyp, path )
    case nm.Equals( QnameBuffer ): return castBuffer( mgVal, at, callTyp, path )
    case nm.Equals( QnameString ): return castString( mgVal, at, callTyp, path )
    case nm.Equals( QnameInt32 ): return castInt32( mgVal, at, callTyp, path )
    case nm.Equals( QnameInt64 ): return castInt64( mgVal, at, callTyp, path )
    case nm.Equals( QnameUint32 ): return castUint32( mgVal, at, callTyp, path )
    case nm.Equals( QnameUint64 ): return castUint64( mgVal, at, callTyp, path )
    case nm.Equals( QnameFloat32 ): 
        return castFloat32( mgVal, at, callTyp, path )
    case nm.Equals( QnameFloat64 ): 
        return castFloat64( mgVal, at, callTyp, path )
    case nm.Equals( QnameTimestamp ): 
        return castTimestamp( mgVal, at, callTyp, path )
    case nm.Equals( QnameSymbolMap ): 
        return castSymbolMap( mgVal, at, callTyp, path )
    }
    if _, ok := mgVal.( *Enum ); ok { 
        return castEnum( mgVal, at, callTyp, path ) 
    }
    return nil, NewTypeCastErrorValue( at, mgVal, path )
}

func checkRestriction( 
    val Value, 
    at *AtomicTypeReference, 
    path objpath.PathNode ) error {

    if at.Restriction.AcceptsValue( val ) { return nil }
    return NewValueCastErrorf( 
        path, "Value %s does not satisfy restriction %s",
        QuoteValue( val ), at.Restriction.ExternalForm() )
}

func CastAtomicWithCallType(
    mgVal Value,
    at *AtomicTypeReference,
    callTyp TypeReference,
    path objpath.PathNode ) ( val Value, err error ) {

    val, err = castAtomicUnrestricted( mgVal, at, callTyp, path )
    if err == nil && at.Restriction != nil { 
        err = checkRestriction( val, at, path ) 
    }
    return
}

func CastAtomic(
    mgVal Value, 
    at *AtomicTypeReference,
    path objpath.PathNode ) ( val Value, err error ) {
    return CastAtomicWithCallType( mgVal, at, at, path )
}
