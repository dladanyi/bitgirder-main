package main

import (
    "mingle/testgen"
    "bytes"
    "fmt"
    mg "mingle"
    mgIo "mingle/io"
)

type typeCode int8

const (
    tcEnd = typeCode( iota )
    tcInvalidDataTest
    tcRoundtripTest
    tcSequenceRoundtripTest
)

const valFileHeader = int32( 1 )

func writeTypeCode( tc typeCode, w *mgIo.BinWriter ) error {
    return w.WriteInt8( int8( tc ) )
}

func writeValue( val interface{}, bb *bytes.Buffer ) error {
    return mgIo.WriteBinIoTestValue( val, bb )
}

func writeTestName( test interface{}, w *mgIo.BinWriter ) error {
    return w.WriteUtf8( mg.CoreIoTestNameFor( test ) )
}

func writeInvalidDataTest( 
    t *mg.BinIoInvalidDataTest, w *mgIo.BinWriter ) error {

    if err := writeTypeCode( tcInvalidDataTest, w ); err != nil { return err }
    if err := writeTestName( t, w ); err != nil { return err }
    if err := w.WriteUtf8( t.ErrMsg ); err != nil { return err }
    if err := w.WriteBuffer32( t.Input ); err != nil { return err }
    return nil
}

func writeRoundtripTest( t *mg.BinIoRoundtripTest, w *mgIo.BinWriter ) error {
    if err := writeTypeCode( tcRoundtripTest, w ); err != nil { return err }
    if err := writeTestName( t, w ); err != nil { return err }
    bb := &bytes.Buffer{}
    if err := writeValue( t.Val, bb ); err != nil { return err }
    if err := w.WriteBuffer32( bb.Bytes() ); err != nil { return err }
    return nil
}

func writeSequenceTest(
    t *mg.BinIoSequenceRoundtripTest, w *mgIo.BinWriter,) error {

    if err := writeTypeCode( tcSequenceRoundtripTest, w ); err != nil { 
        return err 
    }
    if err := writeTestName( t, w ); err != nil { return err }
    bb := &bytes.Buffer{}
    for _, val := range t.Seq {
        if err := writeValue( val, bb ); err != nil { return err }
    }
    if err := w.WriteBuffer32( bb.Bytes() ); err != nil { return err }
    return nil
}

func writeTest( test interface{}, w *mgIo.BinWriter ) error { 
    switch v := test.( type ) {
    case *mg.BinIoInvalidDataTest: return writeInvalidDataTest( v, w )
    case *mg.BinIoRoundtripTest: return writeRoundtripTest( v, w )
    case *mg.BinIoSequenceRoundtripTest: return writeSequenceTest( v, w )
    }
    return fmt.Errorf( "unhandled test type: %T", test )
}

func writeTests( w *mgIo.BinWriter ) error {
    if err := w.WriteInt32( valFileHeader ); err != nil { return err }
    for _, test := range mg.CreateCoreIoTests() {
        if err := writeTest( test, w ); err != nil { return err }
    }
    return writeTypeCode( tcEnd, w )
}

func main() { testgen.WriteOutFile( writeTests ) }
