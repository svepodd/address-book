package psg

import (
	"HW/models/dto"
	"HW/pkg"
	"context"
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"text/template"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Psg представляет гейт к базе данных PostgreSQL.
type Psg struct {
	Conn *pgxpool.Pool
}

// NewPsg создает новый экземпляр Psg.
func NewPsg(psgAddr string, login, password string) (psg *Psg, err error) {
	eW := pkg.NewEWrapper("NewPsg(psgAddr string, login, password string)")

	psg = &Psg{}
	psg.Conn, err = parseConnectionString(psgAddr, login, password)
	if err != nil {
		err = eW.WrapError(err, "parseConnectionString(dburl, login, pass)")
		return nil, err
	}

	err = psg.Conn.Ping(context.Background())
	if err != nil {
		err = eW.WrapError(err, "psg.conn.Ping(context.Background())")
		return nil, err
	}

	return
}

// RecordAdd добавляет новую запись в базу данных.
func (p *Psg) RecordAdd(record dto.Record) (err error) {
	eW := pkg.NewEWrapper("RecordAdd(record dto.Record)") 

	sqlCommand := `INSERT INTO address_book (name, last_name, middle_name, address, phone) VALUES ($1, $2, $3, $4, $5)`
	_, err = p.Conn.Exec(context.Background(), sqlCommand, record.Name, record.LastName, record.MiddleName, record.Address, record.Phone)
	if err != nil {
		return eW.WrapError(err, "p.Conn.Exec()")
	}
	return nil
}

// RecordsGet возвращает записи из базы данных на основе предоставленных полей Record.
func (p *Psg) RecordsGet(record dto.Record) (result []dto.Record, err error) {
	eW := pkg.NewEWrapper("RecordsGet(record dto.Record)") 

	sqlCommand, values, err := p.SelectRecord(record)
	if err != nil {
		return result, eW.WrapError(err, "p.SelectRecord(record)")
	}

	fmt.Println(sqlCommand)
	fmt.Println(values)

	rows, err := p.Conn.Query(context.Background(), sqlCommand, values...) 
	if err != nil {
		return result, eW.WrapError(err, "p.Conn.Query()")
	}
	defer rows.Close() 
	for rows.Next() {
		var r dto.Record
		if err := rows.Scan(&r.ID, &r.Name, &r.LastName, &r.MiddleName, &r.Address, &r.Phone); err != nil {
			return result, eW.WrapError(err, "rows.Scan()")
		}
		result = append(result, r)
	}

	if err := rows.Err(); err != nil {
		return result, eW.WrapError(err, "rows.Err()")
	}

	return result, nil

}

// RecordUpdate обновляет существующую запись в базе данных по номеру телефона.
func (p *Psg) RecordUpdate(record dto.Record) error {
	eW := pkg.NewEWrapper("RecordUpdate(record dto.Record)")
    fields := []string{} 
    values := []interface{}{} 
    index := 1

    if record.Name != "" {
        fields = append(fields, fmt.Sprintf("name=$%d", index)) 
        values = append(values, record.Name) 
        index++
    }
    if record.LastName != "" {
        fields = append(fields, fmt.Sprintf("last_name=$%d", index)) 
        values = append(values, record.LastName)
        index++
    }
    if record.MiddleName != "" {
        fields = append(fields, fmt.Sprintf("middle_name=$%d", index))
        values = append(values, record.MiddleName)
        index++
    }
    if record.Address != "" {
        fields = append(fields, fmt.Sprintf("address=$%d", index))
        values = append(values, record.Address)
        index++
    }

    values = append(values, record.Phone)
    sqlCommand := fmt.Sprintf(`UPDATE address_book SET %s WHERE phone=$%d`, strings.Join(fields, ", "), index) 

    _, err := p.Conn.Exec(context.Background(), sqlCommand, values...)
    if err != nil {
        return eW.WrapError(err, "p.Conn.Exec()")
    }
    return nil
}

// RecordDeleteByPhone удаляет запись из базы данных по номеру телефона.
func (p *Psg) RecordDeleteByPhone(phone string) error {
	eW := pkg.NewEWrapper("RecordDeleteByPhone(phone string)")
	sqlCommand := `DELETE FROM address_book WHERE phone=$1`
	_, err := p.Conn.Exec(context.Background(), sqlCommand, phone)
	if err != nil {
		return eW.WrapError(err, "p.Conn.Exec()")
	}
	return nil
}


func parseConnectionString(dburl, user, password string) (db *pgxpool.Pool, err error) {
	eW := pkg.NewEWrapper("parseConnectionString()")

	var u *url.URL
	if u, err = url.Parse(dburl); err != nil {
		err = eW.WrapError(err, "url.Parse(dburl)")
		return nil, err
	}
	u.User = url.UserPassword(user, password)
	db, err = pgxpool.New(context.Background(), u.String())
	if err != nil {
		err = eW.WrapError(err, "pgxpool.New(context.Background(), u.String())")
		return nil, err
	}
	return
}


func (p *Psg) SelectRecord(r dto.Record) (res_query string, values []any, err error) {
	sqlFields, values, err := structToFieldsValues(r, "sql.field")
	if err != nil {
		return
	}

	var conds []dto.Cond

	for i := range sqlFields {
		if i == 0 {
			conds = append(conds, dto.Cond{
				Lop:    "",
				PgxInd: "$" + strconv.Itoa(i+1),
				Field:  sqlFields[i],
				Value:  values[i],
			})
			continue
		}
		conds = append(conds, dto.Cond{
			Lop:    "AND",
			PgxInd: "$" + strconv.Itoa(i+1),
			Field:  sqlFields[i],
			Value:  values[i],
		})
	}

	query := `
	SELECT 
		id, name, last_name, middle_name, address, phone
	FROM
	    address_book
	WHERE
		{{range .}} {{.Lop}} {{.Field}} = {{.PgxInd}}{{end}}
;
`
	tmpl, err := template.New("").Parse(query)
	if err != nil {
		return
	}

	var sb strings.Builder
	err = tmpl.Execute(&sb, conds)
	if err != nil {
		return
	}
	res_query = sb.String()
	return
}

func structToFieldsValues(s any, tag string) (sqlFields []string, values []any, err error) {
	rv := reflect.ValueOf(s)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	if rv.Kind() != reflect.Struct {
		return nil, nil, errors.New("s must be a struct")
	}

	for i := 0; i < rv.NumField(); i++ {
		field := rv.Type().Field(i)
		tg := strings.TrimSpace(field.Tag.Get(tag))
		if tg == "" || tg == "-" {
			continue
		}
		tgs := strings.Split(tg, ",")
		tg = tgs[0]

		fv := rv.Field(i)
		isZero := false
		switch fv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			isZero = fv.Int() == 0
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			isZero = fv.Uint() == 0
		case reflect.Float32, reflect.Float64:
			isZero = fv.Float() == 0
		case reflect.Complex64, reflect.Complex128:
			isZero = fv.Complex() == complex(0, 0)
		case reflect.Bool:
			isZero = !fv.Bool()
		case reflect.String:
			isZero = fv.String() == ""
		case reflect.Array, reflect.Slice:
			isZero = fv.Len() == 0
		}

		if isZero {
			continue
		}

		sqlFields = append(sqlFields, tg)
		values = append(values, fv.Interface())
	}

	return
}
