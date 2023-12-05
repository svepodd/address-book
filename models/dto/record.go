package dto

import "encoding/json"

type Record struct {
	ID         int64  `json:"-" sql.field:"id"`
	Name       string `json:"name" sql.field:"name"`
	LastName   string `json:"last_name" sql.field:"last_name"`
	MiddleName string `json:"middle_name,omitempty" sql.field:"middle_name"`
	Address    string `json:"address" sql.field:"address"`
	Phone      string `json:"phone" sql.field:"phone"`
}

type Cond struct {
	Lop    string
	PgxInd string
	Field  string
	Value  any
}

// структура ответа сервера
type Response struct {
	Result string 			`json:"result"`
    Data   json.RawMessage	`json:"data"`
    Error string 			`json:"error"`
}

func (r *Response) Wrap(result string, data json.RawMessage, error string) {
    r.Result = result
    r.Data = data
	r.Error = error
}
