package builtin

import (
    mgRct "mingle/reactor"
    "mingle/parser"
    "mingle/bind"
    mg "mingle"
    "bitgirder/objpath"
//    "log"
)

// if err is something that should be sent to caller as a value error, a value
// error is returned; otherwise err is returned unchanged
func asValueError( ve mgRct.ReactorEvent, err error ) error {
    switch v := err.( type ) {
    case *parser.ParseError:
        err = mg.NewValueCastError( ve.GetPath(), v.Error() )
    case *mg.BinIoError: err = mg.NewValueCastError( ve.GetPath(), v.Error() )
    }
    return err
}

func setStructFunc( 
    b *mgRct.FunctionsBuilderFactory,
    reg *bind.Registry,
    f func( *bind.Registry ) mgRct.FieldSetBuilder ) {

    b.StructFunc = func( 
        _ *mgRct.StructStartEvent ) ( mgRct.FieldSetBuilder, error ) {

        return f( reg ), nil
    }
}

func setListFunc(
    b *mgRct.FunctionsBuilderFactory, 
    valFact func() interface{},
    addVal func( val, acc interface{} ) interface{},
    nextFunc func() mgRct.BuilderFactory ) {

    b.ListFunc = func( _ *mgRct.ListStartEvent ) ( mgRct.ListBuilder, error ) {
        lb := bind.NewFunctionsListBuilder()
        lb.Value = valFact()
        lb.AddFunc = func( val interface{}, path objpath.PathNode ) error {
            lb.Value = addVal( val, lb.Value )
            return nil
        }
        lb.NextFunc = nextFunc
        return lb, nil
    }
}

func registerBoundField0( 
    fsb *mgRct.FunctionsFieldSetBuilder,
    fld *mg.Identifier,
    typ mg.TypeReference,
    set func( fldVal, val interface{} ),
    reg *bind.Registry ) {

    fsb.RegisterField(
        fld,
        func( path objpath.PathNode ) ( mgRct.BuilderFactory, error ) {
            return reg.MustBuilderFactoryForType( typ ), nil
        },
        func( val interface{}, path objpath.PathNode ) error {
            set( val, fsb.Value )
            return nil
        },
    )
}

func createIdSliceBuilderFactory( reg *bind.Registry ) mgRct.BuilderFactory {
    res := bind.NewFunctionsBuilderFactory()
    setListFunc(
        res,
        func() interface{} { return make( []*mg.Identifier, 0, 2 ) },
        func( val, acc interface{} ) interface{} {
            return append( acc.( []*mg.Identifier ), val.( *mg.Identifier ) )
        },
        builderFactFuncForType( mg.TypeIdentifier, reg ),
    )
    return res
}

func registerIdSliceField(
    fsb *mgRct.FunctionsFieldSetBuilder,
    fld *mg.Identifier,
    reg *bind.Registry,
    set func( val interface{}, path objpath.PathNode ) error ) {
    
    fsb.RegisterField(
        fld,
        func( path objpath.PathNode ) ( mgRct.BuilderFactory, error ) {
            return createIdSliceBuilderFactory( reg ), nil
        },
        set,
    )
}

func builderFactFuncForType( 
    typ mg.TypeReference, reg *bind.Registry ) func() mgRct.BuilderFactory {

    return func() mgRct.BuilderFactory {
        if bf, ok := reg.BuilderFactoryForType( typ ); ok { return bf }
        return nil
    }
}

func idPartsBuilderFactory( reg *bind.Registry ) mgRct.BuilderFactory {
    res := bind.NewFunctionsBuilderFactory()
    setListFunc(
        res,
        func() interface{} { return make( []string, 0, 2 ) },
        func( val, acc interface{} ) interface{} {
            return append( acc.( []string ), val.( string ) )
        },
        builderFactFuncForType( mg.TypeString, reg ),
    )
    return res
}

func idBuilderForStruct( reg *bind.Registry ) mgRct.FieldSetBuilder {
    idBuilder := bind.NewFunctionsFieldSetBuilder()
    idBuilder.RegisterField( 
        identifierParts,
        func( path objpath.PathNode ) ( mgRct.BuilderFactory, error ) {
            return idPartsBuilderFactory( reg ), nil
        },
        func( val interface{}, path objpath.PathNode ) error {
            idBuilder.Value = mg.NewIdentifierUnsafe( val.( []string ) )
            return nil
        },
    )
    return idBuilder
}

func idFromBytes( ve *mgRct.ValueEvent ) ( interface{}, error, bool ) {
    if b, ok := ve.Val.( mg.Buffer ); ok {
        res, err := mg.IdentifierFromBytes( []byte( b ) )
        if err != nil { err = asValueError( ve, err ) }
        return res, err, true
    }
    return nil, nil, false
}

func idFromString( ve *mgRct.ValueEvent ) ( interface{}, error, bool ) {
    if s, ok := ve.Val.( mg.String ); ok {
        res, err := parser.ParseIdentifier( string( s ) )
        if err != nil { err = asValueError( ve, err ) }
        return res, err, true
    }
    return nil, nil, false
}

func newIdBuilderFactory( reg *bind.Registry ) mgRct.BuilderFactory {
    res := bind.NewFunctionsBuilderFactory()
    setStructFunc( res, reg, idBuilderForStruct )
    res.ValueFunc = 
        mgRct.NewBuildValueOkFunctionSequence( idFromBytes, idFromString )
    return res
}

func nsBuilderForStruct( reg *bind.Registry ) mgRct.FieldSetBuilder {
    res := bind.NewFunctionsFieldSetBuilder()
    res.Value = new( mg.Namespace )
    res.RegisterField(
        idUnsafe( "version" ),
        func( path objpath.PathNode ) ( mgRct.BuilderFactory, error ) {
            if bf, ok := reg.BuilderFactoryForType( mg.TypeIdentifier ); ok {
                return bf, nil
            }
            return nil, nil
        },
        func( val interface{}, path objpath.PathNode ) error {
            res.Value.( *mg.Namespace ).Version = val.( *mg.Identifier )
            return nil
        },
    )
    registerIdSliceField( res, idUnsafe( "parts" ), reg, 
        func( val interface{}, path objpath.PathNode ) error {
            res.Value.( *mg.Namespace ).Parts = val.( []*mg.Identifier )
            return nil
        },
    )
    return res
}

func nsFromBytes( ve *mgRct.ValueEvent ) ( interface{}, error, bool ) {
    if b, ok := ve.Val.( mg.Buffer ); ok {
        res, err := mg.NamespaceFromBytes( []byte( b ) )
        if err != nil { err = asValueError( ve, err ) }
        return res, err, true
    }
    return nil, nil, false
}

func nsFromString( ve *mgRct.ValueEvent ) ( interface{}, error, bool ) {
    if s, ok := ve.Val.( mg.String ); ok {
        res, err := parser.ParseNamespace( string( s ) )
        if err != nil { err = asValueError( ve, err ) }
        return res, err, true
    }
    return nil, nil, false
}

func newNsBuilderFactory( reg *bind.Registry ) mgRct.BuilderFactory {
    res := bind.NewFunctionsBuilderFactory()
    setStructFunc( res, reg, nsBuilderForStruct )
    res.ValueFunc = 
        mgRct.NewBuildValueOkFunctionSequence( nsFromBytes, nsFromString )
    return res
}

func idPathPartFromValue( ve *mgRct.ValueEvent ) ( interface{}, error, bool ) {
    negErr := func() error {
        return mg.NewValueCastError( ve.GetPath(), "value is negative" )
    }
    switch v := ve.Val.( type ) {
    case mg.Int32:
        if int32( v ) < 0 { return nil, negErr(), true }
        return uint64( int32( v ) ), nil, true
    case mg.Int64:
        if int64( v ) < 0 { return nil, negErr(), true }
        return uint64( int64( v ) ), nil, true
    case mg.Uint32: return uint64( uint32( v ) ), nil, true
    case mg.Uint64: return uint64( v ), nil, true
    }
    return nil, nil, false
}

func idPathPartFailBadVal( ve *mgRct.ValueEvent ) ( interface{}, error, bool ) {
    tmpl := "invalid value for identifier path part: %s"
    err := mg.NewValueCastErrorf( ve.GetPath(), tmpl, mgRct.TypeOfEvent( ve ) )
    return nil, err, true
}

// note that we have ValueFunc end with idPathPartFailBadVal so that we can fail
// with a ValueCastError instead of the default error. This is to reflect the
// intent of IdentifierPart.parts being typed as Value+, but where the values
// themselves are expected to be of a finite set of types (if we had union types
// we would use that)
func idPathPartBuilderFactory( reg *bind.Registry ) mgRct.BuilderFactory {
    res := bind.NewFunctionsBuilderFactory()
    res.StructFunc = func( 
        sse *mgRct.StructStartEvent ) ( mgRct.FieldSetBuilder, error ) {

        if qn := sse.Type; qn.Equals( mg.QnameIdentifier ) {
            if bf, ok := reg.BuilderFactoryForName( qn ); ok {
                return bf.StartStruct( sse )
            }
        }
        return nil, nil
    }
    res.ValueFunc = mgRct.NewBuildValueOkFunctionSequence(
        idFromBytes, idFromString, idPathPartFromValue, idPathPartFailBadVal )
    return res
}

func idPathPartsBuilder( reg *bind.Registry ) mgRct.BuilderFactory {
    res := bind.NewFunctionsBuilderFactory()
    setListFunc( 
        res,
        func() interface{} { return make( []interface{}, 0, 4 ) },
        func( val, acc interface{} ) interface{} {
            return append( acc.( []interface{} ), val )
        },
        func() mgRct.BuilderFactory { return idPathPartBuilderFactory( reg ) },
    )
    return res
}

func buildIdPath( parts []interface{} ) objpath.PathNode {
    var res objpath.PathNode
    for _, part := range parts {
        switch v := part.( type ) {
        case uint64:
            if res == nil { 
                res = objpath.RootedAtList().SetIndex( v )
            } else {
                res = res.StartList().SetIndex( v )
            }
        case *mg.Identifier:
            if res == nil {
                res = objpath.RootedAt( v )
            } else {
                res = res.Descend( v )
            }
        default: panic( libErrorf( "unhandled id path part: %T", part ) )
        }
    }
    return res
}

func idPathFromString( ve *mgRct.ValueEvent ) ( interface{}, error, bool ) {
    if s, ok := ve.Val.( mg.String ); ok {
        res, err := parser.ParseIdentifierPath( string( s ) )
        if err != nil { err = asValueError( ve, err ) }
        return res, err, true
    }
    return nil, nil, false
}

func newIdPathBuilderFactory( reg *bind.Registry ) mgRct.BuilderFactory {
    res := bind.NewFunctionsBuilderFactory()
    setStructFunc( res, reg, func( reg *bind.Registry ) mgRct.FieldSetBuilder {
        res := bind.NewFunctionsFieldSetBuilder()
        res.RegisterField(
            idUnsafe( "parts" ),
            func( path objpath.PathNode ) ( mgRct.BuilderFactory, error ) {
                return idPathPartsBuilder( reg ), nil
            },
            func( val interface{}, path objpath.PathNode ) error {
                res.Value = buildIdPath( val.( []interface{} ) )
                return nil
            },
        )
        return res
    })
    res.ValueFunc = idPathFromString
    return res
}

func newLocatableErrorBuilderFactory( 
    qn *mg.QualifiedTypeName, 
    instFact func() interface{},
    msgSet, locSet func( fldVal, err interface{} ),
    addFlds func( fsb *mgRct.FunctionsFieldSetBuilder ),
    reg *bind.Registry ) *mgRct.FunctionsBuilderFactory {

    res := bind.NewFunctionsBuilderFactory()
    setStructFunc( res, reg, func( reg *bind.Registry ) mgRct.FieldSetBuilder {
        errBldr := bind.NewFunctionsFieldSetBuilder()
        errBldr.Value = instFact()
        registerBoundField0( 
            errBldr, idUnsafe( "message" ), mg.TypeString, msgSet, reg )
        registerBoundField0(
            errBldr, idUnsafe( "location" ), mg.TypeIdentifierPath, locSet, 
            reg )
        if addFlds != nil { addFlds( errBldr ) }
        return errBldr
    })
    return res
}

func newCastErrorBuilderFactory( reg *bind.Registry ) mgRct.BuilderFactory {
    return newLocatableErrorBuilderFactory(     
        mg.QnameCastError, 
        func() interface{} { return new( mg.ValueCastError ) },
        func( fldVal, err interface{} ) {
            err.( *mg.ValueCastError ).Message = fldVal.( string )
        },
        func( fldVal, err interface{} ) {
            err.( *mg.ValueCastError ).Location = fldVal.( objpath.PathNode )
        },
        nil,
        reg,
    )
}

func newUnrecognizedFieldErrorBuilderFactory( 
    reg *bind.Registry ) mgRct.BuilderFactory {

    return newLocatableErrorBuilderFactory(
        mg.QnameUnrecognizedFieldError,
        func() interface{} { return new( mg.UnrecognizedFieldError ) },
        func( fldVal, err interface{} ) {
            err.( *mg.UnrecognizedFieldError ).Message = fldVal.( string )
        },
        func( fldVal, err interface{} ) {
            err.( *mg.UnrecognizedFieldError ).Location = 
                fldVal.( objpath.PathNode )
        },
        func( fsb *mgRct.FunctionsFieldSetBuilder ) {
            set := func( val, err interface{} ) {
                err.( *mg.UnrecognizedFieldError ).Field = 
                    val.( *mg.Identifier )
            }
            registerBoundField0( 
                fsb, idUnsafe( "field" ), mg.TypeIdentifier, set, reg )
        },
        reg,
    )
}

func newMissingFieldsErrorBuilderFactory( 
    reg *bind.Registry ) mgRct.BuilderFactory {

    return newLocatableErrorBuilderFactory(
        mg.QnameMissingFieldsError,
        func() interface{} { return new( mg.MissingFieldsError ) },
        func( fldVal, err interface{} ) {
            err.( *mg.MissingFieldsError ).Message = fldVal.( string )
        },
        func( fldVal, err interface{} ) {
            err.( *mg.MissingFieldsError ).Location = 
                fldVal.( objpath.PathNode )
        },
        func( fsb *mgRct.FunctionsFieldSetBuilder ) {
            registerIdSliceField( fsb, idUnsafe( "fields" ), reg,
                func( val interface{}, _ objpath.PathNode ) error {
                    flds := val.( []*mg.Identifier )
                    fsb.Value.( *mg.MissingFieldsError ).SetFields( flds )
                    return nil
                },
            )
        },
        reg,
    )
}

type eventSend struct {
    out mgRct.ReactorEventProcessor
    bc *bind.BindContext
}

func ( es eventSend ) startStruct( qn *mg.QualifiedTypeName ) error {
    return es.out.ProcessEvent( mgRct.NewStructStartEvent( qn ) )
}

func ( es eventSend ) startField( fld *mg.Identifier ) error {
    return es.out.ProcessEvent( mgRct.NewFieldStartEvent( fld ) )
}

func ( es eventSend ) startList( lt *mg.ListTypeReference ) error {
    return es.out.ProcessEvent( mgRct.NewListStartEvent( lt ) )
}

func ( es eventSend ) value( v mg.Value ) error {
    return es.out.ProcessEvent( mgRct.NewValueEvent( v ) )
}

func ( es eventSend ) end() error {
    return es.out.ProcessEvent( mgRct.NewEndEvent() )
}

func visitIdentifierAsStruct( id *mg.Identifier, es eventSend ) error {
    if err := es.startStruct( mg.QnameIdentifier ); err != nil { return err }
    if err := es.startField( identifierParts ); err != nil { return err }
    if err := es.startList( typeIdentifierPartsList ); err != nil { return err }
    for _, part := range id.GetPartsUnsafe() {
        if err := es.value( mg.String( part ) ); err != nil { return err }
    }
    if err := es.end(); err != nil { return err } // parts
    if err := es.end(); err != nil { return err } // struct
    return nil
}

func visitIdentifier( id *mg.Identifier, es eventSend ) error {
    switch opts := es.bc.SerialOptions; opts.Format {
    case bind.SerialFormatBinary:
        return es.value( mg.Buffer( mg.IdentifierAsBytes( id ) ) )
    case bind.SerialFormatText:
        return es.value( mg.String( id.Format( opts.Identifiers ) ) )
    }
    return visitIdentifierAsStruct( id, es )
}

func visitIdentifierList( ids []*mg.Identifier, es eventSend ) ( err error ) {
    lt := typeIdentifierPointerList
    switch es.bc.SerialOptions.Format {
    case bind.SerialFormatText: lt = typeNonEmptyStringList
    case bind.SerialFormatBinary: lt = typeNonEmptyBufferList
    }
    if err = es.startList( lt ); err != nil { return }
    for _, id := range ids {
        if err = visitIdentifier( id, es ); err != nil { return }
    }
    if err = es.end(); err != nil { return }
    return
}

func visitNamespaceAsStruct( ns *mg.Namespace, es eventSend ) ( err error ) {
    if err = es.startStruct( mg.QnameNamespace ); err != nil { return }
    if err = es.startField( identifierParts ); err != nil { return }
    if err = visitIdentifierList( ns.Parts, es ); err != nil { return }
    if err = es.startField( identifierVersion ); err != nil { return }
    if err = visitIdentifierAsStruct( ns.Version, es ); err != nil { return }
    if err = es.end(); err != nil { return }
    return
}

func visitNamespace( ns *mg.Namespace, es eventSend ) error {
    switch opts := es.bc.SerialOptions; opts.Format {
    case bind.SerialFormatText:
        return es.value( mg.String( ns.ExternalForm() ) )
    case bind.SerialFormatBinary:
        return es.value( mg.Buffer( mg.NamespaceAsBytes( ns ) ) )
    }
    return visitNamespaceAsStruct( ns, es )
}

type idPathPartsEventSendVisitor struct {
    es eventSend
}

func ( vis idPathPartsEventSendVisitor ) Descend( elt interface{} ) error {
    return visitIdentifier( elt.( *mg.Identifier ), vis.es )
}

func ( vis idPathPartsEventSendVisitor ) List( idx uint64 ) error {
    return vis.es.value( mg.Uint64( idx ) )
}

func visitIdPathAsStruct( p objpath.PathNode, es eventSend ) ( err error ) {
    if err = es.startStruct( mg.QnameIdentifierPath ); err != nil { return }
    if err = es.startField( identifierParts ); err != nil { return }
    if err = es.startList( typeIdentifierPathPartsList ); err != nil { return }
    if err = objpath.Visit( p, idPathPartsEventSendVisitor{ es } ); err != nil {
        return
    }
    if err = es.end(); err != nil { return } // parts
    if err = es.end(); err != nil { return } // struct
    return
}

func visitIdPath( p objpath.PathNode, es eventSend ) error {
    if es.bc.SerialOptions.Format == bind.SerialFormatText {
        return es.value( mg.String( mg.FormatIdPath( p ) ) )
    }
    return visitIdPathAsStruct( p, es )
}

func visitLocatableError( 
    loc objpath.PathNode, msg string, es eventSend ) ( err error ) {

    if loc != nil {
        if err = es.startField( identifierLocation ); err != nil { return }
        if err = visitIdPath( loc, es ); err != nil { return }
    }
    if err = es.startField( identifierMessage ); err != nil { return }
    if err = es.value( mg.String( msg ) ); err != nil { return }
    return
}

func visitValueCastError( e *mg.ValueCastError, es eventSend ) ( err error ) {
    if err = es.startStruct( mg.QnameCastError ); err != nil { return }
    if err = visitLocatableError( e.Location, e.Message, es ); err != nil {
        return
    }
    if err = es.end(); err != nil { return }
    return
}

func visitUnrecognizedFieldError( 
    e *mg.UnrecognizedFieldError, es eventSend ) ( err error ) {

    if err = es.startStruct( mg.QnameUnrecognizedFieldError ); err != nil {
        return
    }
    if err = visitLocatableError( e.Location, e.Message, es ); err != nil {
        return
    }
    if err = es.startField( identifierField ); err != nil { return }
    if err = visitIdentifier( e.Field, es ); err != nil { return }
    if err = es.end(); err != nil { return }
    return
}

func visitMissingFieldsError( 
    e *mg.MissingFieldsError, es eventSend ) ( err error ) {

    if err = es.startStruct( mg.QnameMissingFieldsError ); err != nil { return }
    if err = visitLocatableError( e.Location, e.Message, es ); err != nil {
        return
    }
    if err = es.startField( identifierFields ); err != nil { return }
    if err = visitIdentifierList( e.Fields(), es ); err != nil { return }
    if err = es.end(); err != nil { return }
    return
}

func visitBuiltinTypeOk(
    val interface{},
    out mgRct.ReactorEventProcessor,
    bc *bind.BindContext,
    path objpath.PathNode ) ( error, bool ) {

    es := eventSend{ out, bc }
    switch v := val.( type ) {
    case *mg.Identifier: return visitIdentifier( v, es ), true
    case *mg.Namespace: return visitNamespace( v, es ), true
    case objpath.PathNode: return visitIdPath( v, es ), true
    case *mg.ValueCastError: return visitValueCastError( v, es ), true
    case *mg.UnrecognizedFieldError: 
        return visitUnrecognizedFieldError( v, es ), true
    case *mg.MissingFieldsError: return visitMissingFieldsError( v, es ), true
    }
    return nil, false
}

func initBind() {
    reg := bind.RegistryForDomain( bind.DomainDefault )
    reg.MustAddValue( mg.QnameIdentifier, newIdBuilderFactory( reg ) )
    reg.MustAddValue( mg.QnameNamespace, newNsBuilderFactory( reg ) )
    reg.MustAddValue( mg.QnameIdentifierPath, newIdPathBuilderFactory( reg ) )
    reg.MustAddValue( mg.QnameCastError, newCastErrorBuilderFactory( reg ) )
    reg.MustAddValue( 
        mg.QnameUnrecognizedFieldError,
        newUnrecognizedFieldErrorBuilderFactory( reg ),
    )
    reg.MustAddValue( 
        mg.QnameMissingFieldsError, newMissingFieldsErrorBuilderFactory( reg ) )
    reg.AddVisitValueOkFunc( visitBuiltinTypeOk )
}