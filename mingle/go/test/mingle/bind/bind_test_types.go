package bind

import (
    mg "mingle"
)

var domainPackageBindTest = mkId( "package-bind-test" )

type BindTestDirection int

const (
    BindTestDirectionRoundtrip = iota
    BindTestDirectionIn
    BindTestDirectionOut
)

func ( d BindTestDirection ) Includes( d2 BindTestDirection ) bool {
    return d == d2 || d == BindTestDirectionRoundtrip
}

type BindTest struct {
    Mingle mg.Value
    BoundId *mg.Identifier
    Direction BindTestDirection
    Type mg.TypeReference
    Domain *mg.Identifier
    SerialOptions *SerialOptions
    Error error
}
