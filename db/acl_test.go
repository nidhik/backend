package db

import (
	"fmt"
	"testing"
)

var aclTests = []struct {
	acl        *ACL
	readers    []string
	writers    []string
	notReaders []string
	notWriters []string
}{
	// None
	{NewACL(), nil, nil, []string{"foo"}, []string{"foo", "bar", "*"}},
	// Public Read
	{&ACL{ACL: map[string]Permission{"*": Permission{Read: true}}}, []string{"*", "foobarbaz"}, nil, nil, []string{"foo", "bar", "*"}},
	// Pubic Write
	{&ACL{ACL: map[string]Permission{"*": Permission{Write: true}}}, nil, []string{"*"}, []string{"foo", "bar", "*"}, nil},
	// Public Read/Write
	{&ACL{ACL: map[string]Permission{"*": Permission{Read: true, Write: true}}}, []string{"*", "foo"}, []string{"*", "bar"}, nil, nil},

	{&ACL{ACL: map[string]Permission{"foo": Permission{Read: true, Write: true}, "bar": Permission{Read: true}, "baz": Permission{Write: true}}}, []string{"foo", "bar"}, []string{"baz", "foo"}, []string{"baz", "random"}, []string{"bar", "random"}},
}

func TestReadACL(t *testing.T) {
	for _, test := range aclTests {
		acl := test.acl

		for _, reader := range test.readers {
			if !acl.CanRead(reader) {
				t.Fatal("Expected canRead() to be true, got false:", reader)
			}
		}

		for _, writer := range test.writers {
			if !acl.CanWrite(writer) {
				t.Fatal("Expected canWrite() to be true, got false:", writer)
			}
		}

		for _, notr := range test.notReaders {
			if acl.CanRead(notr) {
				t.Fatal("Expected canRead() to be false, got true:", notr)
			}
		}

		for _, notw := range test.notWriters {
			if acl.CanWrite(notw) {
				t.Fatal("Expected canWrite() to be false, got true:", notw)
			}
		}
	}
}

func TestBuildACL(t *testing.T) {
	// 1.
	acl := NewACL()
	fmt.Printf("1. New ACL: %s \n", acl)

	if acl.CanRead("blah") {
		t.Fatal("Expected canRead() to be false, got true.")
	}

	if acl.CanWrite("blah") {
		t.Fatal("Expected canWrite() to be false, got true.")
	}

	acl.SetPublicRead()
	fmt.Printf("Add Public Read. ACL: %s \n", acl)

	if !acl.CanRead("blah") {
		t.Fatal("Expected canRead() to be true, got false.")
	}

	if acl.CanWrite("blah") {
		t.Fatal("Expected canWrite() to be false, got true.")
	}

	acl.SetPublicWrite()
	fmt.Printf("Add Public Write. ACL: %s \n", acl)

	if !acl.CanRead("blah") {
		t.Fatal("Expected canRead() to be true, got false.")
	}

	if !acl.CanWrite("blah") {
		t.Fatal("Expected canWrite() to be true, got false.")
	}

	acl.SetPublicReadWrite()
	fmt.Printf("Add Public ReadWrite ACL: %s \n", acl)

	if !acl.CanRead("blah") {
		t.Fatal("Expected canRead() to be true, got false.")
	}

	if !acl.CanWrite("blah") {
		t.Fatal("Expected canWrite() to be true, got false.")
	}

	// 2.
	acl2 := NewACL()
	fmt.Printf("2. New ACL: %s \n", acl2)

	acl2.SetPublicReadWrite()
	fmt.Printf("Add Public ReadWrite ACL: %s \n", acl2)

	if !acl2.CanRead("blah") {
		t.Fatal("Expected canRead() to be true, got false.")
	}

	if !acl2.CanWrite("blah") {
		t.Fatal("Expected canWrite() to be true, got false.")
	}

	acl2.AddRead("testname")
	fmt.Printf("Add Read for testname. ACL: %s \n", acl2)

	if !acl2.CanRead("testname") {
		t.Fatal("Expected canRead() to be true, got false.")
	}

	if !acl2.CanWrite("testname") {
		t.Fatal("Expected canWrite() to be true, got false.")
	}

	acl2.AddWrite("testname")
	fmt.Printf("Add Read for testname. ACL: %s \n", acl2)

	if !acl2.CanRead("testname") {
		t.Fatal("Expected canRead() to be true, got false.")
	}

	if !acl2.CanWrite("testname") {
		t.Fatal("Expected canWrite() to be true, got false.")
	}

	// 3.
	acl3 := NewACL()
	fmt.Printf("3. New ACL: %s \n", acl3)

	acl3.AddRead("foobar")
	acl3.AddRead("barbaz")
	acl3.AddWrite("foofoofoo")

	fmt.Printf("Add Read for foobar and barbaz, write for foofoofoo. ACL: %s \n", acl3)

	if !acl3.CanRead("foobar") {
		t.Fatal("Expected canRead() to be true, got false.")
	}

	if acl3.CanWrite("foobar") {
		t.Fatal("Expected canWrite() to be false, got true.")
	}

	if !acl3.CanRead("barbaz") {
		t.Fatal("Expected canRead() to be true, got false.")
	}

	if acl3.CanWrite("barbaz") {
		t.Fatal("Expected canWrite() to be false, got true.")
	}

	if acl3.CanRead("foofoofoo") {
		t.Fatal("Expected canRead() to be false, got true.")
	}

	if !acl3.CanWrite("foofoofoo") {
		t.Fatal("Expected canWrite() to be true, got false.")
	}

	if acl3.CanRead("random") {
		t.Fatal("Expected canRead() to be false, got true.")
	}

	if acl3.CanWrite("random") {
		t.Fatal("Expected canWrite() to be false, got true.")
	}

}
