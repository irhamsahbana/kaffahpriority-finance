package pkg

import (
	"encoding/json"
	"fmt"
)

func PrettyPrint(identifier string, v interface{}) {
	bytes, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Println("Error marshalling JSON:", err)
		return
	}
	fmt.Println(identifier + ":")
	fmt.Println(string(bytes))
}
