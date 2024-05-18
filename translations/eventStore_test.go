package translations

import (
	"fmt"
	"testing"
)

type Struct1 struct {
	Message string `json:"message"`
}

type Struct2 struct {
	Value int `json:"value"`
}

func TestSerialization(t *testing.T) {
	s := Struct1{
		Message: "hello",
	}

	serialized, err := Serialize(s)
	if err != nil {
		t.Fatal(err)
	}

	deserialized, err := Deserialize(serialized, Struct1{}, Struct2{})
	if err != nil {
		t.Fatal(err)
	}

	switch data := deserialized.(type) {
	case Struct1:
		fmt.Println("got Struct1")
		fmt.Println(data.Message)
		if data.Message != "hello" {
			t.Error("wrong message value")
		}
	case Struct2:
		fmt.Println("got Struct2")
		fmt.Println(data.Value)
	default:
		t.Fatal("no type match")
	}
}
