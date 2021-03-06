package com.bitgirder.lang.path;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;
import com.bitgirder.lang.TypedString;
import com.bitgirder.lang.ObjectReceiver;

import com.bitgirder.test.Test;

import java.util.List;

@Test
final
class PathTests
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private
    final
    static
    class Node
    extends TypedString< Node >
    {
        private Node( CharSequence str ) { super( str, "str" ); }
    }

    private
    final
    static
    class FormatterImpl
    implements ObjectPathFormatter< Node >
    {
        public void formatPathStart( StringBuilder sb ) { sb.append( '/' ); }
        public void formatSeparator( StringBuilder sb ) { sb.append( '/' ); }

        public
        void
        formatDictionaryKey( StringBuilder sb,
                             Node n )
        {
            sb.append( n );
        }

        public
        void
        formatListIndex( StringBuilder sb,
                         int indx )
        {
            sb.append( "[ " ).append( indx ).append( " ]" );
        }
    }

    private
    < V >
    CharSequence
    format( ObjectPath< V > p )
    {
        return ObjectPaths.format( p, ObjectPaths.DOT_FORMATTER );
    }

    private
    void
    assertFormat( CharSequence expct,
                  ObjectPath< Node > p )
    {
        state.equalString( 
            expct, ObjectPaths.format( p, new FormatterImpl() ) );
    }

    @Test
    private
    void
    testFormat0()
    {
        assertFormat( 
            "/node1/node2[ 2 ]/node3", 
            ObjectPath.< Node >getRoot().
                descend( new Node( "node1" ) ).
                descend( new Node( "node2" ) ).
                startImmutableList().
                next().
                next().
                descend( new Node( "node3" ) ) );
    }

    @Test
    private
    void
    testFormat1()
    {
        assertFormat( 
            "/node1[ 1 ][ 2 ]/node2/node3",
            ObjectPath.< Node >getRoot().
                descend( new Node( "node1" ) ).
                startImmutableList().
                next().
                startImmutableList().
                next().
                next().
                descend( new Node( "node2" ) ).
                descend( new Node( "node3" ) ) );
    }

    @Test
    private
    void
    testDefaultFormats()
    {
        ObjectPath< Node > path =
            ObjectPath.< Node >getRoot().
                descend( new Node( "node1" ) ).
                startImmutableList().
                next().
                startImmutableList().
                next().
                next().
                descend( new Node( "node2" ) ).
                descend( new Node( "node3" ) );

        state.equalString(
            "node1[ 1 ][ 2 ].node2.node3",
            ObjectPaths.format( path, ObjectPaths.DOT_FORMATTER ) );
        
        state.equalString(
            "node1[ 1 ][ 2 ]/node2/node3",
            ObjectPaths.format( path, ObjectPaths.SLASH_FORMATTER ) );
    }

    @Test
    private
    void
    testMutableListPath()
    {
        ObjectPath< Node > p = ObjectPath.getRoot( new Node( "n1" ) );
        MutableListPath< Node> lp = p.startMutableList( 4 );

        assertFormat( "/n1[ 4 ]/n2", lp.descend( new Node( "n2" ) ) );
        assertFormat( "/n1[ 7 ]", lp.setIndex( 7 ) );
        assertFormat( "/n1[ 8 ]", lp.increment() );
    }

    private
    < V >
    void
    assertRoot( ObjectPath< V > rootExpct,
                ObjectPath< V > path )
    {
        state.equal( rootExpct, ObjectPaths.rootOf( path ) );
    }

    @Test
    private
    void
    testPathRootIdentities()
    {
        ObjectPath< String > root = ObjectPath.getRoot();

        assertRoot( root, root );
        assertRoot( root, root.descend( "p1" ) );
        assertRoot( root, root.descend( "p1" ).descend( "p2" ) );

        assertRoot( root, 
            root.descend( "p1" ).startImmutableList().next().next() );
    }

    @Test
    private
    void
    testStartListWithIndex()
    {
        ImmutableListPath< Node > p = 
            ObjectPath.< Node >getRoot().
                descend( new Node( "a" ) ).
                descend( new Node( "b" ) ).
                startImmutableList( 5 );
        
        assertFormat( "/a/b[ 5 ]", p );
        assertFormat( "/a/b[ 6 ]", p.next() );
    }

    private
    < V >
    void
    assertPathsEq( ObjectPath< V > p1,
                   ObjectPath< V > p2 )
    {
        ObjectPaths.areEqual( p1, p2 );
        ObjectPaths.areEqual( p2, p1 );

        ObjectPaths.areEqual( p1, p1 ); 
        ObjectPaths.areEqual( p2, p2 ); 
    }

    private
    < V >
    void
    assertPathsNeq( ObjectPath< V > p1,
                    ObjectPath< V > p2 )
    {
        boolean res = ObjectPaths.areEqual( p1, p2 );
        state.isFalse( ObjectPaths.areEqual( p1, p2 ) );
        state.isFalse( ObjectPaths.areEqual( p2, p1 ) );
    }

    @Test
    private
    void
    testPathEquals()
    {
        ObjectPath< String > p1 = ObjectPath.getRoot();
        ObjectPath< String > p2 = ObjectPath.getRoot();
        assertPathsEq( p1, p2 );
        p2 = p2.descend( "n1" );
        assertPathsNeq( p1, p2 );
        p1 = p1.descend( "n1" );
        assertPathsEq( p1, p2 );
        p1 = p1.descend( "n2" ).startImmutableList().next().next();
        assertPathsNeq( p1, p2 );
        state.isFalse( p2.equals( p1 ) );
        p2 = p2.descend( "n2" ).startImmutableList().next();
        assertPathsNeq( p1, p2 );
        ImmutableListPath< String > p2Cast = Lang.castUnchecked( p2 );
        p2 = p2Cast.next();
        assertPathsEq( p1, p2 );
    }

    private
    < V >
    void
    assertVisit( ObjectPath< V > p,
                 List< ObjectPath< V > > expct )
        throws Exception
    {
        final List< ObjectPath< V > > act = Lang.newList();

        p.visitDescent( new ObjectReceiver< ObjectPath< V > >() {
            public void receive( ObjectPath< V > v ) { act.add( v ); }
        });

        state.equal( expct, act );
    }

    @Test
    public
    void
    testVisitDescentEmpty()
        throws Exception
    {
        assertVisit( 
            ObjectPath.< String >getRoot(), 
            Lang.< ObjectPath< String > >newList() 
        );
    }

    @Test
    public
    void
    testVisitDescentNonTrivial()
        throws Exception
    {
        ObjectPath< String > p = ObjectPath.getRoot( "p1" );

        List< ObjectPath< String > > expct = Lang.newList();
        expct.add( p );
        expct.add( p = p.descend( "p2" ) );
        expct.add( p = p.startImmutableList() );
        expct.add( p = p.descend( "p3" ) );
        expct.add( p = p.startImmutableList( 9 ) );

        assertVisit( p, expct );
    }

    private
    void
    assertMutableInstance( ObjectPath< ? > p,
                           boolean wantMutable )
    {
        // we don't yet have mutable dictionary paths, but if/when we do we'd
        // check that here
        if ( p instanceof DictionaryPath ) return;

        ListPath< ? > lp = state.cast( ListPath.class, p );

        if ( wantMutable ) state.cast( MutableListPath.class, lp );
        else state.cast( ImmutableListPath.class, lp );
    }

    private
    < V >
    void
    assertCopy( ObjectPath< V > expct,
                boolean mutable )   
    {
        ObjectPath< V > act = mutable ?
            ObjectPaths.asMutableCopy( expct ) :
            ObjectPaths.asImmutableCopy( expct );

        for ( ObjectPath< V > p : act.collectDescent() ) {
            assertMutableInstance( p, mutable );
        }

        state.isTrue( ObjectPaths.areEqual( expct, act ) );
    }

    private
    < V >
    void
    assertCopy( ObjectPath< V > expct )
    {
        assertCopy( expct, true );
        assertCopy( expct, false );
    }

    @Test
    public
    void
    testCopies()
    {
        assertCopy( ObjectPath.< String >getRoot() );

        ObjectPath< String > p = ObjectPath.getRoot( "n1" ).
            descend( "n2" ).
            startImmutableList( 5 ).
            descend( "n3" ).
            startMutableList( 3 ).
            descend( "n4" );
        
        for ( ; p != null; p = p.getParent() ) assertCopy( p );
    }

    @Test
    public
    void
    testNullSafeMethods()
    {
        ObjectPath< String > p = 
            ObjectPaths.descend( ObjectPaths.descend( null, "p1" ), "p2" );
        
        state.equalString( "p1.p2", format( p ) );
    }
}
