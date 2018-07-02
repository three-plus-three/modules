package errors

type ErrorBuilder struct {
	code      int
	message   string
	fields    map[string]interface{}
	internals []*ApplicationError
}

func (err *ErrorBuilder) WithInternalError(e error) *ErrorBuilder {
	if rerr, ok := e.(*ApplicationError); ok {
		if rerr.HTTPCode() == ToHttpStatus(ErrCodeMultipleError) {
			return err.WithInternalErrors(rerr.Internals)
		}
		err.internals = append(err.internals, rerr)
		return err
	}
	err.internals = append(err.internals, ToApplicationError(e))
	return err
}

func (err *ErrorBuilder) WithInternalErrors(internals []*ApplicationError) *ErrorBuilder {
	if len(err.internals) <= 0 {
		err.internals = internals
		return err
	}
	if len(internals) > 0 {
		err.internals = append(err.internals, internals...)
	}
	return err
}

func (err *ErrorBuilder) WithField(nm string, v interface{}) *ErrorBuilder {
	if nil == err.fields {
		err.fields = map[string]interface{}{}
	}
	err.fields[nm] = v
	return err
}

func (err *ErrorBuilder) Fields() map[string]interface{} {
	return err.fields
}

func (err *ErrorBuilder) FieldsWithDefault() map[string]interface{} {
	if nil == err.fields {
		err.fields = map[string]interface{}{}
	}
	return err.fields
}

func (err *ErrorBuilder) Build() *ApplicationError {
	var fields map[string]interface{}
	var internals []*ApplicationError
	if len(err.fields) > 0 {
		fields = err.fields
	}

	if len(err.internals) > 0 {
		internals = err.internals
	}

	return &ApplicationError{
		ErrCode:    err.code,
		ErrMessage: err.message,
		Fields:     fields,
		Internals:  internals,
	}
}

func Build(code int, msg string) *ErrorBuilder {
	return &ErrorBuilder{
		code:    code,
		message: msg,
	}
}

func ReBuildFromRuntimeError(e RuntimeError) *ErrorBuilder {
	var fields map[string]interface{}
	var internals []*ApplicationError
	if err, ok := e.(*ApplicationError); ok {
		if len(err.Fields) > 0 {
			fields = map[string]interface{}{}
			for k, v := range err.Fields {
				fields[k] = v
			}
		}

		if len(err.Internals) > 0 {
			internals = make([]*ApplicationError, len(err.Internals))
			copy(internals, err.Internals)
		}
	}
	return &ErrorBuilder{
		code:      e.Code(),
		message:   e.Error(),
		fields:    fields,
		internals: internals,
	}
}

func ReBuildFromError(e error, code int) *ErrorBuilder {
	if err, ok := e.(RuntimeError); ok {
		return ReBuildFromRuntimeError(err)
	}
	return &ErrorBuilder{
		code:    code,
		message: e.Error(),
	}
}

func BuildApplicationErrorFromError(e error, code int) *ApplicationError {
	return ToApplicationError(e, code)
}
