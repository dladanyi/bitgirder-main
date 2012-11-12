package com.bitgirder.mingle.model;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.parser.SyntaxException;

public
final
class MingleBoolean
implements MingleValue
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    public final static MingleBoolean TRUE = new MingleBoolean( Boolean.TRUE );

    public final static MingleBoolean FALSE = 
        new MingleBoolean( Boolean.FALSE );

    private final Boolean b;

    private MingleBoolean( Boolean b ) { this.b = b; }

    public boolean booleanValue() { return b.booleanValue(); }

    public int hashCode() { return b.hashCode(); }

    public
    boolean
    equals( Object other )
    {
        return 
            other == this ||
            ( other instanceof MingleBoolean &&
              b.equals( ( (MingleBoolean) other ).b ) );
    }

    @Override public String toString() { return b.toString(); }

    public
    static
    MingleBoolean
    valueOf( boolean b )
    {
        return b ? TRUE : FALSE; 
    }

    public
    static
    MingleBoolean
    parse( CharSequence str )
        throws SyntaxException
    {
        inputs.notNull( str, "str" );

        String s = str.toString();

        if ( s.equals( "true" ) ) return TRUE;
        else if ( s.equals( "false" ) ) return FALSE;
        else throw new SyntaxException( "Invalid boolean string: " + str );
    }
}
