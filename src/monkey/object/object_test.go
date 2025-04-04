package object

import "testing"

func TestStringHashKey(t *testing.T) {
	hello1 := &String{Value: "Hello World"}
	hello2 := &String{Value: "Hello World"}
	diff1 := &String{Value: "my name is johnny"}
	diff2 := &String{Value: "my name is johnny"}

	if hello1.HashKey() != hello2.HashKey() {
		t.Errorf("string with the same contents have different hash keys")
	}

	if diff1.HashKey() != diff2.HashKey() {
		t.Errorf("string with the same content have different hash keys")
	}

	if hello1.HashKey() == diff1.HashKey() {
		t.Errorf("strins with different content have the same hash keys")
	}
}
