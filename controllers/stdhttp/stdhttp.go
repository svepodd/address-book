package stdhttp

import (
	"HW/gate/psg"
	"HW/models/dto"
	"HW/pkg"
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

// Controller обрабатывает HTTP запросы для адресной книги.
type Controller struct {
	DB  *psg.Psg
	Srv *http.Server
}


// NewController создает новый Controller.
func NewController(addr string, db *psg.Psg) (cont *Controller) {
	cont = new(Controller)
	cont.Srv = &http.Server{}
	mux := http.NewServeMux()
	mux.Handle("/create", http.HandlerFunc(cont.RecordAdd))
	mux.Handle("/get", http.HandlerFunc(cont.RecordsGet))
	mux.Handle("/update", http.HandlerFunc(cont.RecordUpdate))
	mux.Handle("/delete", http.HandlerFunc(cont.RecordDeleteByPhone))
	cont.Srv.Handler = mux
	cont.Srv.Addr = addr
	cont.DB = db
	return cont
}

func (cont *Controller) Start() (err error) { // запускаем сервер
	eW := pkg.NewEWrapper("(cont *Controller) Start()")

	err = cont.Srv.ListenAndServe()
	if err != nil {
		err = eW.WrapError(err, "cont.Srv.ListenAndServe()")
		return
	}
	return
}


// RecordAdd обрабатывает HTTP запрос для добавления новой записи.
func (cont *Controller) RecordAdd(w http.ResponseWriter, r *http.Request) {
	var err error
	eW := pkg.NewEWrapper("(cont *Controller) RecordAdd()")

	resp := &dto.Response{}
	defer responseReturn(w, eW, resp)

	if r.Method != http.MethodPost {
        http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
        return
    }

	record := dto.Record{}
	byteReq, err := io.ReadAll(r.Body)
	if err != nil {
		resp.Wrap("Error reading request", nil, err.Error())
		eW.LogError(err, "io.ReadAll(req.Body)")
		return
	}
	err = json.Unmarshal(byteReq, &record)
	if err != nil {
		resp.Wrap("Error JSON", nil, err.Error())
		eW.LogError(err, "json.Unmarshal()")
		return
	}

	if record.Name == "" || record.LastName == "" || record.Address == "" || record.Phone == "" {
		err = errors.New("required data is missing")
		resp.Wrap("Required data is missing", nil, err.Error())
		eW.LogError(err, "json.Unmarshal")
		return
	}

	record.Phone, err = pkg.PhoneNormalize(record.Phone)
	if err != nil {
		resp.Wrap("Error: wrong Phone", nil, err.Error())
		eW.LogError(err, "pkg.PhoneNormalize(record.Phone)")
		return
	}

	err = cont.DB.RecordAdd(record)

	if err != nil {
		resp.Wrap("Error in saving record", nil, err.Error())
		eW.LogError(err, "cont.DB.RecordAdd(record)")
		return
	}

	resp.Wrap("Successfully added", nil, "")
}

// RecordsGet обрабатывает HTTP запрос для получения записей на основе предоставленных полей Record.
func (c *Controller) RecordsGet(w http.ResponseWriter, r *http.Request) {
	var err error
	eW := pkg.NewEWrapper("(c *Controller) RecordsGet")

	resp := &dto.Response{} 
	defer responseReturn(w, eW, resp) 

	if r.Method != http.MethodPost { 
        http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
        return
    }

	record := dto.Record{}
	byteReq, err := io.ReadAll(r.Body) 
	if err != nil {
		resp.Wrap("Error reading request", nil, err.Error())
		eW.LogError(err, "io.ReadAll(r.Body)")
		return
	}

	err = json.Unmarshal(byteReq, &record) 
	if err != nil {
		resp.Wrap("Error JSON", nil, err.Error())
		eW.LogError(err, "json.Unmarshal(r)")
		return
	}

	if record.Phone != ""{
		record.Phone, err = pkg.PhoneNormalize(record.Phone)
		if err != nil {
			resp.Wrap("Error: wrong Phone", nil, err.Error())
			eW.LogError(err, "pkg.PhoneNormalize(record.Phone)")
			return
		}
	}

	records, err := c.DB.RecordsGet(record)
	if err != nil {
		resp.Wrap("Error in finding records", nil, err.Error())
		eW.LogError(err, "c.DB.RecordsGet(record)")
		return
	}

	recordsJSON, err := json.Marshal(records)
	if err != nil {
		resp.Wrap("Error JSON", nil, err.Error())
		eW.LogError(err, "json.Marshal(records)")
		return
	}

	resp.Wrap("OK", recordsJSON, "")
}

// RecordUpdate обрабатывает HTTP запрос для обновления записи.
func (c *Controller) RecordUpdate(w http.ResponseWriter, r *http.Request) {
	var err error
	eW := pkg.NewEWrapper("(c *Controller) RecordUpdate()")

	resp := &dto.Response{}
	defer responseReturn(w, eW, resp)

	if r.Method != http.MethodPost {
        http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
        return
    }

	record := dto.Record{}
	byteReq, err := io.ReadAll(r.Body)
	if err != nil {
		resp.Wrap("Error reading request", nil, err.Error())
		eW.LogError(err, "io.ReadAll(req.Body)")
		return
	}
	err = json.Unmarshal(byteReq, &record)
	if err != nil {
		resp.Wrap("Error JSON", nil, err.Error())
		eW.LogError(err, "json.Unmarshal(byteReq, &record)")
		return
	}

	if (record.Name == "" && record.LastName == "" && record.MiddleName == "" && record.Address == "") || record.Phone == "" {
		err = errors.New("required data is missing") 
		resp.Wrap("Required data is missing", nil, err.Error())
		eW.LogError(err, "json.Unmarshal")
		return
	}

	record.Phone, err = pkg.PhoneNormalize(record.Phone)
	if err != nil {
		resp.Wrap("Error: wrong Phone", nil, err.Error())
		eW.LogError(err, "pkg.PhoneNormalize(record.Phone)")
		return
	}

	err = c.DB.RecordUpdate(record)
	if err != nil {
		resp.Wrap("Error in updating record", nil, err.Error()) 
		eW.LogError(err, "c.DB.RecordUpdate(record)")
		return
	}
	resp.Wrap("OK", nil, "")
}

// RecordDeleteByPhone обрабатывает HTTP запрос для удаления записи по номеру телефона.
func (c *Controller) RecordDeleteByPhone(w http.ResponseWriter, r *http.Request) {
	var err error
	eW := pkg.NewEWrapper("(c *Controller) RecordDeleteByPhone()")

	resp := &dto.Response{}
	defer responseReturn(w, eW, resp)

	if r.Method != http.MethodPost {
        http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
        return
    }

	record := dto.Record{}
	byteReq, err := io.ReadAll(r.Body)
	if err != nil {
		resp.Wrap("Error reading request", nil, err.Error())
		eW.LogError(err, "io.ReadAll(r.Body)")
		return
	}
	err = json.Unmarshal(byteReq, &record)
	if err != nil {
		resp.Wrap("Error JSON", nil, err.Error())
		eW.LogError(err, "json.Unmarshal(byteReq, &record)")
		return
	}

	if record.Phone == "" {
		err = errors.New("phone data is missing")
		resp.Wrap("Phone data is missing", nil, err.Error())
		eW.LogError(err, "json.Unmarshal")
		return
	}

	record.Phone, err = pkg.PhoneNormalize(record.Phone)
	if err != nil {
		resp.Wrap("Error: wrong Phone", nil, err.Error())
		eW.LogError(err, "pkg.PhoneNormalize(record.Phone)")
		return
	}

	err = c.DB.RecordDeleteByPhone(record.Phone)
	if err != nil {
		resp.Wrap("Error in deleting record", nil, err.Error())
		eW.LogError(err, "c.DB.RecordDeleteByPhone(record.Phone)")
		return
	}
	resp.Wrap("OK", nil, "")
}


func responseReturn(w http.ResponseWriter, eW *pkg.EWrapper, resp *dto.Response){
	err_encode := json.NewEncoder(w).Encode(resp) 
	if err_encode != nil {
		eW.LogError(err_encode, "json.NewEncoder(w).Encode(resp)")
		w.WriteHeader(http.StatusPaymentRequired)
		return
	}
}