package market_data

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestRetrieveOptions(t *testing.T) {
	got, err := RetrieveOptions("AMD", "PUT", "OTM", "2023-06-30")
	if err != nil {
		t.Fatalf("RetrieveOptions() error = %v", err)
		return
	}
	if len(got) <= 0 {
		t.Fatal("nothing returned")
	}
	b, err := json.MarshalIndent(got, "", "  ")
	if err != nil {
		t.Fatalf("fail to marshalJson: %v", err)
	}
	fmt.Println(string(b))
}
