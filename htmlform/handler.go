package htmlform

import (
	"fmt"
	"html/template"
	"net/http"
	"reflect"

	"github.com/domonda/go-types"

	"github.com/ungerik/go-command"
	"github.com/ungerik/go-fs"
	"github.com/ungerik/go-fs/multipartfs"
	"github.com/ungerik/go-httpx/httperr"
)

var typeOfFileReader = reflect.TypeOf((*fs.FileReader)(nil)).Elem()

type Option struct {
	Label string
	Value interface{}
}

type formField struct {
	Name    string
	Label   string
	Type    string
	Value   string
	Options []Option
}

type Handler struct {
	cmdFunc         command.StringMapArgsFunc
	args            command.Args
	argValidator    map[string]types.ValidatErr
	argOptions      map[string][]Option
	argDefaultValue map[string]interface{}
	form            struct {
		Title            string
		Fields           []formField
		SubmitButtonText string
	}
	template       *template.Template
	successHandler http.Handler
}

func NewHandler(commandFunc interface{}, args command.Args, title string, successHandler http.Handler) (handler *Handler, err error) {
	handler = &Handler{
		args:            args,
		argValidator:    make(map[string]types.ValidatErr),
		argOptions:      make(map[string][]Option),
		argDefaultValue: make(map[string]interface{}),
		successHandler:  successHandler,
	}
	handler.form.Title = title
	handler.form.SubmitButtonText = "Submit"
	handler.cmdFunc, err = command.GetStringMapArgsFunc(commandFunc, args)
	if err != nil {
		return nil, err
	}
	handler.template, err = template.New("form").Parse(FormTemplate)
	if err != nil {
		return nil, err
	}
	return handler, nil
}

func MustNewHandler(commandFunc interface{}, args command.Args, title string, successHandler http.Handler) (handler *Handler) {
	handler, err := NewHandler(commandFunc, args, title, successHandler)
	if err != nil {
		panic(err)
	}
	return handler
}

func (handler *Handler) SetArgValidator(arg string, validator types.ValidatErr) {
	handler.argValidator[arg] = validator
}

func (handler *Handler) SetArgOptions(arg string, options []Option) {
	handler.argOptions[arg] = options
}

func (handler *Handler) SetArgDefaultValue(arg string, value interface{}) {
	handler.argDefaultValue[arg] = value
}

func (handler *Handler) SetSubmitButtonText(text string) {
	handler.form.SubmitButtonText = text
}

func (handler *Handler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			httperr.Handle(httperr.AsError(r), response, request)
		}
	}()

	switch request.Method {
	case "GET":
		handler.get(response, request)
	case "POST":
		handler.post(response, request)
	default:
		http.Error(response, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
}

func (handler *Handler) get(response http.ResponseWriter, request *http.Request) {
	handler.form.Fields = nil
	for _, arg := range handler.args.Args() {
		field := formField{
			Name:  arg.Name,
			Label: arg.Description,
			Type:  "text",
		}
		if field.Label == "" {
			field.Label = arg.Name
		}
		if defaultValue, ok := handler.argDefaultValue[arg.Name]; ok {
			field.Value = fmt.Sprint(defaultValue)
		}
		options, isSelect := handler.argOptions[arg.Name]
		switch {
		case isSelect:
			field.Type = "select"
			field.Options = options

		case arg.Type.Implements(typeOfFileReader):
			field.Type = "file"

		// case arg.Type == reflect.TypeOf(date.Date("")) || arg.Type == reflect.TypeOf(date.NullableDate("")):
		// 	field.Type = "date"

		// case arg.Type == reflect.TypeOf(time.Time{}):
		// 	field.Type = "datetime-local"

		default:
			switch arg.Type.Kind() {
			case reflect.Bool:
				field.Type = "checkbox"
			case reflect.Float32, reflect.Float64:
				field.Type = "number"
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				field.Type = "number"
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				field.Type = "number"
			}
		}

		handler.form.Fields = append(handler.form.Fields, field)
	}

	err := handler.template.Execute(response, &handler.form)
	if err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
	}
}

func (handler *Handler) post(response http.ResponseWriter, request *http.Request) {
	formfs, err := multipartfs.FromRequestForm(request, 100*1024*1024)
	if err != nil {
		http.Error(response, err.Error(), http.StatusBadRequest)
		return
	}
	defer formfs.Close()

	argsMap := make(map[string]string)
	for key, vals := range formfs.Form.Value {
		argsMap[key] = vals[0]
	}
	for key := range formfs.Form.File {
		file, _ := formfs.FormFile(key)
		argsMap[key] = string(file)
	}

	err = handler.cmdFunc(request.Context(), argsMap)
	if err != nil {
		httperr.Handle(err, response, request)
		return
	}

	handler.successHandler.ServeHTTP(response, request)
}
