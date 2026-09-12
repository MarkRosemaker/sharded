```bash
go get github.com/MarkRosemaker/sharded
```


```go
package main

import (
	"fmt"

	"github.com/MarkRosemaker/sharded"
)

func main() {
	m := sharded.NewStringMap[int]()

	m.Set("key", 3)

	v, ok := m.Get("key")
	fmt.Println(v, ok) // 3 true

	m.Delete("key")
	fmt.Println(m.Len()) // 0
}
```
