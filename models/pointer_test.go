package models

import (
	"testing"

	"github.com/nidhik/backend/db"
)

var pointerTests = []struct {
	className string
	objectId  string
	expected  db.Model
	err       error
}{
	{"gooaklsdjnf", "3FanK38ZTb", nil, ERR_UNKNOWN_COLL_NAME},
	{CollectionRole, "123", NewRole("123"), nil},
	{CollectionUser, "789", NewUser("789"), nil},
	{CollectionTask, "123456", NewTask("123456"), nil},
}

func TestPointer(t *testing.T) {

	for _, test := range pointerTests {
		p := NewPointer(test.className, test.objectId)
		if p.TypeName != "Pointer" {
			t.Fatal("Expected type name Pointer. Actual:", p.TypeName)
		}

		retModel, err := p.model()

		if err != test.err {
			t.Fatal("Expected error:", test.err, "Actual:", err)
		}

		if test.err == nil {
			if retModel.ObjectId() != test.expected.ObjectId() {
				t.Fatal("Expected objectId:", test.expected.ObjectId(), "Actual:", retModel.ObjectId())
			}

			if retModel.Collection() != test.expected.Collection() {
				t.Fatal("Expected collection name:", test.expected.Collection(), "Actual:", retModel.Collection())
			}

		}

	}

}

func TestUnknownTypeName(t *testing.T) {
	p := &Pointer{TypeName: "foobar", ClassName: CollectionRole, ObjectId: "123"}
	ret, err := p.model()

	if ret != nil {
		t.Fatal("Returned model when should have been nil:", ret)
	}

	if err != ERR_UNKNOWN_POINTER_TYPE {
		t.Fatal("Expected:", ERR_UNKNOWN_POINTER_TYPE, "Actual:", err)
	}
}
