package main

import (
    "fmt"
    "encoding/json"
)

type Request struct {
    Operation string      //`json:"operation"`
    Key string            `json:"key"`
    Value string          `json:"value"`
}

func main() {
    s := string(`{"operation": "get", "key": "example"}`)
    data := Request{}
    err := json.Unmarshal([]byte(s), &data)

	if err != nil {
	    fmt.Println(err.Error()) 
	    //invalid character '\'' looking for beginning of object key string
	}
	fmt.Printf("Unmarshaling Data [%+v] Operation [%s]\n", data, data.Operation)
	
	data2 := Request{}
	j2, err := json.Marshal(data2)

	fmt.Printf("Marshaling Data2 [%+v] Json [%s]\n", data2, j2)
}