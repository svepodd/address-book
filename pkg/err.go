package pkg

import (
	"errors"
	"log"
) 

type EWrapper struct {
	functionName string
	comment  string
	err error
}

func NewEWrapper(funcName string) *EWrapper {
	return &EWrapper{funcName, "", nil}
}

func (e *EWrapper) Wrap(err error, comment string) *EWrapper {
	if err != nil {
		e.err = err
		e.comment = comment
	}
	return e
}

func (e *EWrapper) Error() error {
	if e.err == nil {
		return nil
	}
	return errors.New(e.comment + "    __IN__    " + e.functionName + ":\n" + e.err.Error() + "\n")
}

func (e *EWrapper) WrapError (err error, comment string) error{
	if err != nil {
		return e.Wrap(err, comment).Error()
	}
	return nil
}

func (e *EWrapper) LogError(err error, comment string){
	if err != nil {
		e.comment = comment
		log.Println(e.Wrap(err, comment).Error())
	}
}

