package com.bitgirder.concurrent;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.test.Test;

import java.util.concurrent.Callable;

@Test
final
class LazyValueTests
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static class MarkerException extends Exception {}

    private
    final
    static
    class StringCall
    extends LazyValue< String >
    {
        private final String s;
        private final boolean fail;

        private int calls;

        private
        StringCall( String s,
                    boolean fail )
        {
            this.s = s;
            this.fail = fail;
        }

        public
        String
        call()
            throws Exception
        {
            ++calls;
            if ( fail ) throw new MarkerException();
            return s;
        }
    }

    @Test
    public
    void
    testSuccess()
        throws Exception
    {
        StringCall sc = new StringCall( "hello", false );

        // ensure sc is called only once
        for ( int i = 0; i < 2; ++i )
        {
            state.equal( "hello", sc.get() );
            state.equalInt( 1, sc.calls );
        }
    }

    @Test
    public
    void
    testFailure()
        throws Exception
    {
        StringCall sc = new StringCall( "bad", true );

        for ( int i = 0; i < 2; ++i )
        {
            try
            {
                sc.get();
                state.fail( "No failure" );
            }
            catch ( MarkerException ok ) {}

            state.equalInt( 1, sc.calls );
        }
    }

    private
    final
    static
    class HelloCall
    implements Callable< String >
    {
        private final boolean fail;

        private HelloCall( boolean fail ) { this.fail = fail; }

        public
        String
        call()
            throws Exception
        {
            if ( fail ) throw new MarkerException();
            return "hello";
        }
    }

    @Test
    private
    void
    testForCallSuccess()
        throws Exception
    {
        state.equal(
            "hello",
            LazyValue.< String >forCall( new HelloCall( false ) ).get()
        );
    }

    @Test( expected = MarkerException.class )
    private
    void
    testForCallFailure()
        throws Exception
    {
        LazyValue.< String >forCall( new HelloCall( true ) ).get();
    }
}
