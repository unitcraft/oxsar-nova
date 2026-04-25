package goal

import "encoding/json"

// jsonMarshal — обёртка для тестируемости (мы можем подменить в тестах).
var jsonMarshal = json.Marshal
