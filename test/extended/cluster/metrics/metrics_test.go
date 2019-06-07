package metrics

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

func TestSum(t *testing.T) {
	td := TestDuration{
		TestDuration: 2 * time.Second,
	}

	bolB, err := json.Marshal(td)
	fmt.Println(string(bolB))

	if err != nil {
		t.Errorf("error: %s", err.Error())
	}
}
