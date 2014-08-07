package service

import (
    mg "mingle"
    "mingle/types"
    "bitgirder/objpath"
)

var (
    NsService *mg.Namespace
    QnameRequest *mg.QualifiedTypeName
    TypeRequest *mg.AtomicTypeReference
    QnameResponse *mg.QualifiedTypeName
    TypeResponse *mg.AtomicTypeReference
    IdNamespace *mg.Identifier
    IdService *mg.Identifier
    IdOperation *mg.Identifier
    IdAuthentication *mg.Identifier
    IdParameters *mg.Identifier
    IdResult *mg.Identifier
    IdError *mg.Identifier
)

type RequestContext struct {
    Namespace *mg.Namespace
    Service *mg.Identifier
    Operation *mg.Identifier
}

func initNames() {
    mkId := func( parts ...string ) *mg.Identifier {
        return mg.NewIdentifierUnsafe( parts )
    }
    declNm := mg.NewDeclaredTypeNameUnsafe
    NsService = &mg.Namespace{
        Parts: []*mg.Identifier{ mkId( "mingle" ), mkId( "service" ) },
        Version: mkId( "v1" ),
    }
    QnameRequest = &mg.QualifiedTypeName{
        Namespace: NsService,
        Name: declNm( "Request" ),
    }
    TypeRequest = QnameRequest.AsAtomicType()
    QnameResponse = &mg.QualifiedTypeName{
        Namespace: NsService,
        Name: declNm( "Response" ),
    }
    TypeResponse = QnameResponse.AsAtomicType()
    IdNamespace = mkId( "namespace" )
    IdService = mkId( "service" )
    IdOperation = mkId( "operation" )
    IdAuthentication = mkId( "authentication" )
    IdParameters = mkId( "parameters" )
    IdResult = mkId( "result" )
    IdError = mkId( "error" )
}

func initReqType() {
    sd := types.NewStructDefinition()
    sd.Name = QnameRequest
    addFld := func( nm *mg.Identifier, typ mg.TypeReference ) {
        sd.Fields.Add( &types.FieldDefinition{ Name: nm, Type: typ } )
    }
    idPtr := mg.NewPointerTypeReference( types.TypeIdentifier )
    addFld( IdNamespace, mg.NewPointerTypeReference( types.TypeNamespace ) )
    addFld( IdService, idPtr )
    addFld( IdOperation, idPtr )
    addFld( IdAuthentication, mg.TypeNullableValue )
    addFld( IdParameters, mg.MustNullableTypeReference( mg.TypeSymbolMap ) )
    types.MustAddBuiltinType( sd )
}

func initRespType() {
    sd := types.NewStructDefinition()
    sd.Name = QnameResponse
    addFld := func( nm *mg.Identifier ) {
        sd.Fields.Add(
            &types.FieldDefinition{ Name: nm, Type: mg.TypeNullableValue } )
    }
    addFld( IdResult )
    addFld( IdError )
    types.MustAddBuiltinType( sd )
}

func initTypes() {
    initReqType()
    initRespType()
}

type RequestError struct {
    Path objpath.PathNode
    Message string
}

func ( e *RequestError ) Error() string {
    return mg.FormatError( e.Path, e.Message )
}

func NewRequestError( path objpath.PathNode, msg string ) *RequestError {
    return &RequestError{ Path: path, Message: msg }
}

type ResponseError struct { 
    Path objpath.PathNode
    Message string
}

func ( e *ResponseError ) Error() string {
    return mg.FormatError( e.Path, e.Message )
}

func NewResponseError( path objpath.PathNode, msg string ) *ResponseError {
    return &ResponseError{ Path: path, Message: msg }
}

const respErrMsgMultipleResponseFields =
    "response contains both a result and an error"
