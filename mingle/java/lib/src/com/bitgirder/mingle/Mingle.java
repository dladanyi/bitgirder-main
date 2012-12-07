package com.bitgirder.mingle;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import com.bitgirder.lang.path.ObjectPath;
import com.bitgirder.lang.path.ImmutableListPath;
import com.bitgirder.lang.path.ObjectPaths;
import com.bitgirder.lang.path.ObjectPathFormatter;

import com.bitgirder.io.Base64Exception;

import java.util.Map;
import java.util.List;
import java.util.Iterator;

public
final
class Mingle
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    final static MingleNameResolver CORE_NAME_RESOLVER =
        new MingleNameResolver() {
            public QualifiedTypeName resolve( DeclaredTypeName nm ) {
                return Mingle.resolveInCore( nm );
            }
        };
    
    public final static MingleNamespace NS_CORE;
    public final static QualifiedTypeName QNAME_BOOLEAN;
    public final static MingleTypeReference TYPE_BOOLEAN;
    public final static QualifiedTypeName QNAME_INT32;
    public final static MingleTypeReference TYPE_INT32;
    public final static QualifiedTypeName QNAME_INT64;
    public final static MingleTypeReference TYPE_INT64;
    public final static QualifiedTypeName QNAME_UINT32;
    public final static MingleTypeReference TYPE_UINT32;
    public final static QualifiedTypeName QNAME_UINT64;
    public final static MingleTypeReference TYPE_UINT64;
    public final static QualifiedTypeName QNAME_FLOAT32;
    public final static MingleTypeReference TYPE_FLOAT32;
    public final static QualifiedTypeName QNAME_FLOAT64;
    public final static MingleTypeReference TYPE_FLOAT64;
    public final static QualifiedTypeName QNAME_STRING;
    public final static MingleTypeReference TYPE_STRING;
    public final static QualifiedTypeName QNAME_BUFFER;
    public final static MingleTypeReference TYPE_BUFFER;
    public final static QualifiedTypeName QNAME_TIMESTAMP;
    public final static MingleTypeReference TYPE_TIMESTAMP;
    public final static QualifiedTypeName QNAME_SYMBOL_MAP;
    public final static MingleTypeReference TYPE_SYMBOL_MAP;
    public final static QualifiedTypeName QNAME_NULL;
    public final static MingleTypeReference TYPE_NULL;
    public final static QualifiedTypeName QNAME_VALUE;
    public final static MingleTypeReference TYPE_VALUE;

    private final static Map< DeclaredTypeName, QualifiedTypeName > 
        CORE_DECL_NAMES;

    private final static 
        Map< MingleTypeReference, Class< ? extends MingleValue > > 
            VALUE_CLASSES;

    private final static ListTypeReference TYPE_VALUE_LIST;

    private final static ObjectPathFormatter< MingleIdentifier > 
        PATH_FORMATTER = new PathFormatterImpl();

    private Mingle() {}

    public
    static
    ObjectPathFormatter< MingleIdentifier >
    getIdPathFormatter()
    {
        return PATH_FORMATTER;
    }
    
    public
    static
    AtomicTypeReference
    atomicTypeReferenceIn( MingleTypeReference ref )
    {
        inputs.notNull( ref, "ref" );

        if ( ref instanceof AtomicTypeReference )
        {
            return (AtomicTypeReference) ref;
        }
        else if ( ref instanceof ListTypeReference )
        {
            return 
                atomicTypeReferenceIn( 
                    ( (ListTypeReference) ref ).getElementTypeReference() );
        }
        else if ( ref instanceof NullableTypeReference )
        {
            return
                atomicTypeReferenceIn(
                    ( (NullableTypeReference) ref ).getTypeReference() );
        }
        else throw state.createFail( "Unexpected type ref:", ref );
    }

    static
    QualifiedTypeName
    resolveInCore( DeclaredTypeName nm )
    {
        inputs.notNull( nm, "nm" );
        return CORE_DECL_NAMES.get( nm );
    }

    static
    TypeName
    typeNameIn( MingleTypeReference t )
    {
        state.notNull( t, "t" );

        if ( t instanceof AtomicTypeReference )
        {
            return ( (AtomicTypeReference) t ).getName();
        }
        else if ( t instanceof ListTypeReference )
        {
            return typeNameIn( 
                ( (ListTypeReference) t ).getElementTypeReference() );
        }
        else if ( t instanceof NullableTypeReference )
        {
            return typeNameIn(
                ( (NullableTypeReference) t ).getTypeReference() );
        }
        else throw state.createFail( "Unhandled type reference:", t );
    }

    public
    static
    boolean
    isIntegralType( MingleTypeReference t )
    {
        TypeName n = typeNameIn( inputs.notNull( t, "t" ) );

        return n.equals( QNAME_INT32 ) ||
               n.equals( QNAME_UINT32 ) ||
               n.equals( QNAME_INT64 ) ||
               n.equals( QNAME_UINT64 );
    }

    public
    static
    boolean
    isDecimalType( MingleTypeReference t )
    {
        TypeName n = typeNameIn( inputs.notNull( t, "t" ) );
        
        return n.equals( QNAME_FLOAT32 ) || n.equals( QNAME_FLOAT64 );
    }

    public
    static
    boolean
    isNumberType( MingleTypeReference t )
    {
        return isIntegralType( t ) || isDecimalType( t );
    }

    public
    static
    boolean
    isNumericValue( MingleValue v )
    {
        return ( v instanceof MingleInt32 ) ||
               ( v instanceof MingleInt64 ) ||
               ( v instanceof MingleUint32 ) ||
               ( v instanceof MingleUint64 ) ||
               ( v instanceof MingleFloat32 ) ||
               ( v instanceof MingleFloat64 );
    }

    public
    static
    Class< ? extends MingleValue >
    valueClassFor( MingleTypeReference typ )
    {
        inputs.notNull( typ, "typ" );

        return VALUE_CLASSES.get( typ );
    }

    // We could add other resolution and such here if needed; for now only
    // matches when nm is a qn we know about
    static
    Class< ? extends MingleValue >
    valueClassFor( TypeName nm )
    {
        inputs.notNull( nm, "nm" );

        if ( nm instanceof QualifiedTypeName )
        {
            return valueClassFor( new AtomicTypeReference( nm, null ) );
        }

        return null;
    }

    public
    static
    Class< ? extends MingleValue >
    expectValueClassFor( MingleTypeReference typ )
    {
        Class< ? extends MingleValue > res = valueClassFor( typ );

        if ( res == null )
        {
            throw inputs.createFail( "No value class known for", typ );
        }

        return res;
    }

    private
    static
    void
    implInspectBuffer( StringBuilder sb,
                       MingleBuffer mb )
    {
        sb.append( "buffer:[" );
        sb.append( mb.asHexString() );
        sb.append( "]" );
    }

    private
    static
    void
    implInspectEnum( StringBuilder sb,
                     MingleEnum me )
    {
        sb.append( me.getType().getExternalForm() );
        sb.append( "." );
        sb.append( me.getValue().getExternalForm() );
    }

    private
    static
    void
    implInspectMap( StringBuilder sb,
                    MingleSymbolMap m )
    {
        sb.append( "{" );

        Iterator< Map.Entry< MingleIdentifier, MingleValue > > it =
            m.entrySet().iterator();
        
        while ( it.hasNext() )
        {
            Map.Entry< MingleIdentifier, MingleValue > e = it.next();
            sb.append( e.getKey().getExternalForm() );
            sb.append( ":" );
            implInspect( sb, e.getValue() );

            if ( it.hasNext() ) sb.append( ", " );
        }

        sb.append( "}" );
    }

    private
    static
    void
    implInspectStruct( StringBuilder sb,
                       MingleStruct ms )
    {
        sb.append( ms.getType().getExternalForm() );
        implInspectMap( sb, ms.getFields() );
    }

    private
    static
    void
    implInspectList( StringBuilder sb,
                     MingleList ml )
    {
        sb.append( "[" );

        for ( Iterator< MingleValue > it = ml.iterator(); it.hasNext(); )
        {
            implInspect( sb, it.next() );
            if ( it.hasNext() ) sb.append( ", " );
        }

        sb.append( "]" );
    }

    private
    static
    StringBuilder
    implInspect( StringBuilder sb,
                 MingleValue mv )
    {
        if ( mv instanceof MingleNull ) sb.append( "null" );
        else if ( mv instanceof MingleBoolean ||
                  mv instanceof MingleUint32 ||
                  mv instanceof MingleUint64 ||
                  mv instanceof MingleInt32 ||
                  mv instanceof MingleInt64 ||
                  mv instanceof MingleFloat32 ||
                  mv instanceof MingleFloat64 ) 
        {
            sb.append( mv.toString() );
        }
        else if ( mv instanceof MingleString )
        {
            Lang.appendRfc4627String( sb, (MingleString) mv );
        }
        else if ( mv instanceof MingleTimestamp )
        {
            sb.append( ( (MingleTimestamp) mv ).getRfc3339String() );
        }
        else if ( mv instanceof MingleBuffer )
        {
            implInspectBuffer( sb, (MingleBuffer) mv );
        }
        else if ( mv instanceof MingleEnum )
        {
            implInspectEnum( sb, (MingleEnum) mv );
        }
        else if ( mv instanceof MingleList )
        {
            implInspectList( sb, (MingleList) mv );
        }
        else if ( mv instanceof MingleSymbolMap )
        {
            implInspectMap( sb, (MingleSymbolMap) mv );
        }
        else if ( mv instanceof MingleStruct )
        {
            implInspectStruct( sb, (MingleStruct) mv );
        }
        else state.fail( "Unhandled inspect type:", mv.getClass().getName() );
        
        return sb;
    }

    public
    static
    StringBuilder
    appendInspection( StringBuilder sb,
                      MingleValue mv )
    {
        inputs.notNull( sb, "sb" );
        inputs.notNull( mv, "mv" );

        return implInspect( sb, mv );
    }

    public
    static
    CharSequence
    inspect( MingleValue mv )
    {
        inputs.notNull( mv, "mv" );
        return implInspect( new StringBuilder(), mv );
    }

    private
    static
    MingleTypeReference
    inferredTypeOf( MingleValue mv )
    {   
        state.notNull( mv );

        if ( mv instanceof MingleBoolean ) return TYPE_BOOLEAN;
        if ( mv instanceof MingleInt32 ) return TYPE_INT32;
        if ( mv instanceof MingleInt64 ) return TYPE_INT64;
        if ( mv instanceof MingleUint32 ) return TYPE_UINT32;
        if ( mv instanceof MingleUint64 ) return TYPE_UINT64;
        if ( mv instanceof MingleFloat32 ) return TYPE_FLOAT32;
        if ( mv instanceof MingleFloat64 ) return TYPE_FLOAT64;
        if ( mv instanceof MingleString ) return TYPE_STRING;
        if ( mv instanceof MingleBuffer ) return TYPE_BUFFER;
        if ( mv instanceof MingleTimestamp ) return TYPE_TIMESTAMP;
        if ( mv instanceof MingleSymbolMap ) return TYPE_SYMBOL_MAP;
        if ( mv instanceof MingleList ) return TYPE_VALUE_LIST;
        if ( mv instanceof MingleNull ) return TYPE_NULL;

        if ( mv instanceof TypedMingleValue )
        {
            return ( (TypedMingleValue) mv ).getType();
        }

        throw state.createFail( "Unhandled value:", mv.getClass() );
    }

    private
    static
    MingleTypeCastException
    failCastType( MingleTypeReference expct,
                  MingleValue act,
                  ObjectPath< MingleIdentifier > loc )
    {
        return new MingleTypeCastException( expct, inferredTypeOf( act ), loc );
    }

    private
    static
    MingleValueCastException
    failCastValue( String msg,
                   ObjectPath< MingleIdentifier > loc )
    {
        return new MingleValueCastException( msg, loc );
    }

    private
    static
    MingleValueException
    initCause( MingleValueException mve,
               Throwable cause )
    {
        mve.initCause( cause );
        return mve;
    }

    private
    static
    MingleString
    castAsString( MingleValue mv,
                  MingleTypeReference tcErrTyp,
                  ObjectPath< MingleIdentifier > loc )
    {
        if ( mv instanceof MingleString ) return (MingleString) mv;

        if ( isNumericValue( mv ) || ( mv instanceof MingleBoolean ) )
        {
            return new MingleString( mv.toString() );
        }

        if ( mv instanceof MingleBuffer ) 
        {
            return new MingleString( ( (MingleBuffer) mv ).asBase64String() );
        }

        if ( mv instanceof MingleTimestamp )
        {
            return new MingleString( 
                ( (MingleTimestamp) mv ).getRfc3339String() );
        }

        if ( mv instanceof MingleEnum )
        {
            MingleEnum me = (MingleEnum) mv;
            return new MingleString( me.getValue().getExternalForm() );
        }

        throw failCastType( tcErrTyp, mv, loc );
    }

    private
    static
    RuntimeException
    numTypeNotHandled( QualifiedTypeName qn )
    {
        return state.createFail( "Unhandled number:", qn );
    }

    private
    static
    MingleValue
    castAsNumber( MingleNumber n,
                  QualifiedTypeName qn )
    {
        if ( qn.equals( QNAME_INT32 ) ) return new MingleInt32( n.intValue() );
        else if ( qn.equals( QNAME_INT64 ) )
        {
            return new MingleInt64( n.longValue() );
        }
        else if ( qn.equals( QNAME_UINT32 ) )
        {
            return new MingleUint32( n.intValue() );
        }
        else if ( qn.equals( QNAME_UINT64 ) )
        {
            return new MingleUint64( n.longValue() );
        }
        else if ( qn.equals( QNAME_FLOAT32 ) )
        {
            return new MingleFloat32( n.floatValue() );
        }
        else if ( qn.equals( QNAME_FLOAT64 ) )
        {
            return new MingleFloat64( n.doubleValue() );
        }

        throw numTypeNotHandled( qn );
    }

    private
    static
    boolean
    isDecimalString( MingleString s )
    {
        for ( int i = 0, e = s.length(); i < e; ++i )
        {
            char ch = s.charAt( i );
            if ( ch == '.' || ch == 'e' || ch == 'E' ) return true;
        }

        return false;
    }

    private
    static
    MingleNumber
    parseNumInitial( MingleString s,
                     QualifiedTypeName qn )
    {
        if ( qn.equals( QNAME_FLOAT64 ) || isDecimalString( s ) )
        {
            return MingleFloat64.parseFloat( s );
        }

        if ( qn.equals( QNAME_INT32 ) ) return MingleInt32.parseInt( s );
        else if ( qn.equals( QNAME_INT64 ) ) return MingleInt64.parseInt( s );
        else if ( qn.equals( QNAME_UINT32 ) )
        {
            return MingleUint32.parseUint( s );
        }
        else if ( qn.equals( QNAME_UINT64 ) )
        {
            return MingleUint64.parseUint( s );
        }
        else if ( qn.equals( QNAME_FLOAT32 ) )
        {
            return MingleFloat32.parseFloat( s );
        }

        throw numTypeNotHandled( qn );
    }

    private
    static
    MingleValueCastException
    failCastNumberFormat( NumberFormatException nfe,
                          MingleString in,
                          ObjectPath< MingleIdentifier > loc )
    {   
        String msg = nfe.getMessage();

        String pref = "For input string: ";
        if ( msg.startsWith( pref ) ) msg = "invalid syntax: " + in;
 
        return failCastValue( msg, loc );
    }

    private
    static
    MingleValue
    parseNumber( MingleString ms,
                 QualifiedTypeName qn,
                 ObjectPath< MingleIdentifier > loc )
    {
        try 
        { 
            MingleNumber res = parseNumInitial( ms, qn ); 
            return castAsNumber( res, qn );
        }
        catch ( NumberFormatException nfe )
        {
            throw failCastNumberFormat( nfe, ms, loc );
        }
    }

    private
    static
    MingleValue
    castAsNumber( MingleValue mv,
                  AtomicTypeReference at,
                  MingleTypeReference tcErrTyp,
                  ObjectPath< MingleIdentifier > loc )
    {
        QualifiedTypeName qn = (QualifiedTypeName) at.getName();

        if ( mv instanceof MingleString )
        {
            return parseNumber( (MingleString) mv, qn, loc );
        }
        else if ( mv instanceof MingleNumber ) 
        {
            return castAsNumber( (MingleNumber) mv, qn );
        }

        throw failCastType( tcErrTyp, mv, loc );
    }

    private
    static
    MingleValue
    castTypedValue( TypedMingleValue mv,
                    AtomicTypeReference at,
                    MingleTypeReference tcErrTyp,
                    ObjectPath< MingleIdentifier > loc )
    {
        if ( mv.getType().equals( at ) ) return mv;
        throw new MingleTypeCastException( tcErrTyp, mv.getType(), loc );
    }

    private
    static
    MingleBuffer
    castAsBuffer( MingleValue mv,
                  MingleTypeReference tcErrTyp,
                  ObjectPath< MingleIdentifier > loc )  
    {
        if ( mv instanceof MingleString )
        {
            try { return MingleBuffer.fromBase64String( (MingleString) mv ); }
            catch ( Base64Exception ex )
            {
                throw initCause( failCastValue( ex.getMessage(), loc ), ex );
            }
        }

        throw failCastType( tcErrTyp, mv, loc );
    }

    private
    static
    MingleTimestamp
    castAsTimestamp( MingleValue mv,
                     MingleTypeReference tcErrTyp,
                     ObjectPath< MingleIdentifier > loc )
    {
        if ( mv instanceof MingleString )
        {
            try { return MingleTimestamp.parse( (MingleString) mv ); }
            catch ( MingleSyntaxException mse )
            {
                throw initCause( failCastValue( mse.getMessage(), loc ), mse );
            }
        }

        throw failCastType( tcErrTyp, mv, loc );
    }

    private
    static
    MingleValue
    castAsNull( MingleValue mv,
                MingleTypeReference tcErrTyp,
                ObjectPath< MingleIdentifier > loc )
    {
        if ( mv instanceof MingleNull ) return mv;

        throw failCastType( tcErrTyp, mv, loc );
    }

    private
    static
    MingleValue
    castAsBoolean( MingleValue mv,
                   MingleTypeReference tcErrTyp,
                   ObjectPath< MingleIdentifier > loc )
    {
        if ( mv instanceof MingleString )
        {
            try { return MingleBoolean.parse( (MingleString) mv ); }
            catch ( MingleSyntaxException mse )
            {
                throw initCause( failCastValue( mse.getMessage(), loc ), mse );
            }
        }

        throw failCastType( tcErrTyp, mv, loc );
    }

    private
    static
    MingleValue
    castAsUnrestrictedAtomic( MingleValue mv,
                              AtomicTypeReference at,
                              MingleTypeReference tcErrTyp,
                              ObjectPath< MingleIdentifier > loc )
    {
        TypeName nm = at.getName();

        Class< ? extends MingleValue > valCls = valueClassFor( nm );
        if ( valCls != null && valCls.isInstance( mv ) ) return mv;

        if ( nm.equals( QNAME_STRING ) )
        {
            return castAsString( mv, tcErrTyp, loc );
        }
        else if ( isNumberType( at ) ) 
        {
            return castAsNumber( mv, at, tcErrTyp, loc );
        }
        else if ( nm.equals( QNAME_BUFFER ) )
        {
            return castAsBuffer( mv, tcErrTyp, loc );
        }
        else if ( nm.equals( QNAME_TIMESTAMP ) )
        {
            return castAsTimestamp( mv, tcErrTyp, loc );
        }
        else if ( nm.equals( QNAME_BOOLEAN ) )
        {
            return castAsBoolean( mv, tcErrTyp, loc );
        }
        else if ( nm.equals( QNAME_NULL ) )
        {
            return castAsNull( mv, tcErrTyp, loc );
        }
        else if ( mv instanceof TypedMingleValue )
        {
            return castTypedValue( (TypedMingleValue) mv, at, tcErrTyp, loc );
        }
        else throw failCastType( tcErrTyp, mv, loc );
    }

    private
    static
    MingleValue
    castAsAtomic( MingleValue mv,
                  AtomicTypeReference at,
                  MingleTypeReference tcErrTyp,
                  ObjectPath< MingleIdentifier > loc )
    {
        if ( mv instanceof MingleNull )
        {
            if ( at.equals( TYPE_NULL ) ) return mv;
            throw failCastValue( "Value is null", loc );
        }

        mv = castAsUnrestrictedAtomic( mv, at, tcErrTyp, loc );

        MingleValueRestriction vr = at.getRestriction();
        if ( vr != null ) vr.validate( mv, loc );

        return mv;
    }

    // This method unconditionally returns a copy of the underlying list,
    // whether or not the original could be used as-is. A more optimized
    // strategy would be to lazily create a new list only when needed.
    private
    static
    MingleValue
    castAsList( MingleValue mv,
                ListTypeReference lt,
                MingleTypeReference tcErrTyp,
                ObjectPath< MingleIdentifier > loc )
    {
        if ( ! ( mv instanceof MingleList ) ) 
        {
            throw failCastType( tcErrTyp, mv, loc );
        }

        List< MingleValue > l2 = Lang.newList();
        ImmutableListPath< MingleIdentifier > lp = loc.startImmutableList();
        MingleTypeReference eltTyp = lt.getElementTypeReference();
        Iterator< MingleValue > it = ( (MingleList) mv ).iterator();

        if ( ! ( it.hasNext() || lt.allowsEmpty() ) )
        {
            throw failCastValue( "List is empty", loc );
        }

        while ( it.hasNext() )
        {
            l2.add( implCastValue( it.next(), eltTyp, eltTyp, lp ) );
            lp = lp.next();
        }

        return MingleList.createLive( l2 );
    }

    private
    static
    MingleValue
    castAsNullable( MingleValue mv,
                    NullableTypeReference nt,
                    MingleTypeReference tcErrTyp,
                    ObjectPath< MingleIdentifier > loc )
    {
        if ( mv instanceof MingleNull ) return mv;
        return implCastValue( mv, nt.getTypeReference(), tcErrTyp, loc );
    }

    private
    static
    MingleValue
    implCastValue( MingleValue mv,
                   MingleTypeReference typ,
                   MingleTypeReference tcErrTyp,
                   ObjectPath< MingleIdentifier > loc )
    {
        if ( typ instanceof AtomicTypeReference )
        {
            return castAsAtomic( mv, (AtomicTypeReference) typ, tcErrTyp, loc );
        }
        else if ( typ instanceof ListTypeReference )
        {
            return castAsList( mv, (ListTypeReference) typ, tcErrTyp, loc );
        }
        else if ( typ instanceof NullableTypeReference )
        {
            NullableTypeReference nt = (NullableTypeReference) typ;
            return castAsNullable( mv, nt, tcErrTyp, loc );
        }

        throw state.createFail( "Unrecognized type reference:", typ );
    }

    // Hardcoding impl right now that mv is string since we're only trying to
    // use this to get past type ref parse tests. Once those are done this will
    // be replaced with a more broadly applicable version.
    public
    static
    MingleValue
    castValue( MingleValue mv,
               MingleTypeReference typ,
               ObjectPath< MingleIdentifier > loc )
    {
        inputs.notNull( mv, "mv" );
        inputs.notNull( typ, "typ" );
        inputs.notNull( loc, "loc" );
 
        return implCastValue( mv, typ, typ, loc );
    }

    public
    static
    StringBuilder
    appendIdPath( ObjectPath< MingleIdentifier > p,
                  StringBuilder sb )
    {
        inputs.notNull( p, "p" );
        inputs.notNull( sb, "sb" );

        ObjectPaths.appendFormat( p, PATH_FORMATTER, sb );
        return sb;
    }

    public
    static
    CharSequence
    formatIdPath( ObjectPath< MingleIdentifier > p )
    {
        return appendIdPath( p, new StringBuilder() );
    }

    // Static class init follows

    private
    static
    MingleIdentifier
    initId( String... parts )
    {
        return new MingleIdentifier( parts );
    }

    private
    static
    MingleNamespace
    initNs( MingleIdentifier ver,
            MingleIdentifier... parts )
    {
        return new MingleNamespace( parts, ver );
    }

    private
    static
    QualifiedTypeName
    initCoreQname( String nm )
    {
        state.notNull( NS_CORE, "NS_CORE" );
        state.notNull( CORE_DECL_NAMES, "CORE_DECL_NAMES" );

        DeclaredTypeName dn = new DeclaredTypeName( nm );
        QualifiedTypeName res = new QualifiedTypeName( NS_CORE, dn );

        Lang.putUnique( CORE_DECL_NAMES, dn, res );

        return res;
    }

    private
    static
    AtomicTypeReference
    initCoreType( QualifiedTypeName qn )
    {
        state.notNull( VALUE_CLASSES, "VALUE_CLASSES" );
        AtomicTypeReference res = new AtomicTypeReference( qn, null );

        try
        {
            String nm = "com.bitgirder.mingle.Mingle" + qn.getName();
            Class< ? > c1 = Class.forName( nm );

            VALUE_CLASSES.put( res, c1.asSubclass( MingleValue.class ) );
        }
        catch ( Exception ex )
        {
            String msg = "Couldn't initialize core type class for " + qn;
            throw new RuntimeException( msg );
        }

        return res;
    }

    static
    {
        NS_CORE = 
            initNs( initId( "v1" ), initId( "mingle" ), initId( "core" ) );

        CORE_DECL_NAMES = Lang.newMap();
        VALUE_CLASSES = Lang.newMap();
        
        QNAME_BOOLEAN = initCoreQname( "Boolean" );
        TYPE_BOOLEAN = initCoreType( QNAME_BOOLEAN );
        QNAME_INT32 = initCoreQname( "Int32" );
        TYPE_INT32 = initCoreType( QNAME_INT32 );
        QNAME_INT64 = initCoreQname( "Int64" );
        TYPE_INT64 = initCoreType( QNAME_INT64 );
        QNAME_UINT32 = initCoreQname( "Uint32" );
        TYPE_UINT32 = initCoreType( QNAME_UINT32 );
        QNAME_UINT64 = initCoreQname( "Uint64" );
        TYPE_UINT64 = initCoreType( QNAME_UINT64 );
        QNAME_FLOAT32 = initCoreQname( "Float32" );
        TYPE_FLOAT32 = initCoreType( QNAME_FLOAT32 );
        QNAME_FLOAT64 = initCoreQname( "Float64" );
        TYPE_FLOAT64 = initCoreType( QNAME_FLOAT64 );
        QNAME_STRING = initCoreQname( "String" );
        TYPE_STRING = initCoreType( QNAME_STRING );
        QNAME_BUFFER = initCoreQname( "Buffer" );
        TYPE_BUFFER = initCoreType( QNAME_BUFFER );
        QNAME_TIMESTAMP = initCoreQname( "Timestamp" );
        TYPE_TIMESTAMP = initCoreType( QNAME_TIMESTAMP );
        QNAME_SYMBOL_MAP = initCoreQname( "SymbolMap" );
        TYPE_SYMBOL_MAP = initCoreType( QNAME_SYMBOL_MAP );
        QNAME_NULL = initCoreQname( "Null" );
        TYPE_NULL = initCoreType( QNAME_NULL );
        QNAME_VALUE = initCoreQname( "Value" );
        TYPE_VALUE = initCoreType( QNAME_VALUE );

        TYPE_VALUE_LIST = new ListTypeReference( TYPE_VALUE, true );
    }

    private
    final
    static
    class PathFormatterImpl
    implements ObjectPathFormatter< MingleIdentifier >
    {
        public void formatPathStart( StringBuilder sb ) {}

        public void formatSeparator( StringBuilder sb ) { sb.append( '.' ); }

        public
        void
        formatDictionaryKey( StringBuilder sb,
                             MingleIdentifier key )
        {
            sb.append( key.getExternalForm() );
        }

        public
        void
        formatListIndex( StringBuilder sb,
                         int indx )
        {
            sb.append( "[ " ).append( indx ).append( " ]" );
        }
    }
}
