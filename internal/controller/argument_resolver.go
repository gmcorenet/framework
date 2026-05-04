package controller

import (
	"context"
	"net/http"
	"reflect"
	"strconv"
)

type ArgumentResolver struct {
	container interface{}
}

func NewArgumentResolver(container interface{}) *ArgumentResolver {
	return &ArgumentResolver{
		container: container,
	}
}

func (r *ArgumentResolver) Resolve(method reflect.Method, ctx context.Context, req *http.Request, w http.ResponseWriter, params map[string]string) ([]reflect.Value, error) {
	methodType := method.Type
	numIn := methodType.NumIn()

	arguments := make([]reflect.Value, 0, numIn)

	for i := 1; i < numIn; i++ { // Start from 1, skip receiver
		argType := methodType.In(i)
		arg, err := r.resolveArgument(argType, ctx, req, w, params)
		if err != nil {
			return nil, err
		}
		arguments = append(arguments, arg)
	}

	return arguments, nil
}

func (r *ArgumentResolver) resolveArgument(argType reflect.Type, ctx context.Context, req *http.Request, w http.ResponseWriter, params map[string]string) (reflect.Value, error) {
	switch argType {
	case reflect.TypeOf((*http.Request)(nil)):
		return reflect.ValueOf(req), nil
	case reflect.TypeOf((*http.ResponseWriter)(nil)):
		return reflect.ValueOf(w), nil
	case reflect.TypeOf((*context.Context)(nil)):
		return reflect.ValueOf(ctx), nil
	case reflect.TypeOf(""):
		if param := ctx.Value("params").(map[string]string); param != nil {
			return reflect.ValueOf(params), nil
		}
		return reflect.ValueOf(""), nil
	case reflect.TypeOf(0):
		return reflect.ValueOf(0), nil
	case reflect.TypeOf(int64(0)):
		return reflect.ValueOf(int64(0)), nil
	}

	if argType.Kind() == reflect.Ptr {
		argType = argType.Elem()
	}

	if argType.Kind() == reflect.Struct {
		return r.resolveStruct(argType, req)
	}

	if argType.Kind() == reflect.Map && argType.Key().Kind() == reflect.String {
		return r.resolveMap(argType, req)
	}

	return reflect.Zero(argType), nil
}

func (r *ArgumentResolver) resolveStruct(argType reflect.Type, req *http.Request) (reflect.Value, error) {
	structVal := reflect.New(argType)
	structElem := structVal.Elem()

	for i := 0; i < argType.NumField(); i++ {
		field := argType.Field(i)
		formTag := field.Tag.Get("form")
		queryTag := field.Tag.Get("query")
		pathTag := field.Tag.Get("path")

		var fieldName string
		if formTag != "" {
			fieldName = formTag
		} else if queryTag != "" {
			fieldName = queryTag
		} else if pathTag != "" {
			fieldName = pathTag
		} else {
			fieldName = field.Name
		}

		var value string
		if pathTag != "" {
			if params := req.Context().Value("params").(map[string]string); params != nil {
				value = params[fieldName]
			}
		} else if formTag != "" {
			value = req.FormValue(fieldName)
		} else if queryTag != "" {
			value = req.URL.Query().Get(fieldName)
		}

		if value == "" {
			continue
		}

		fieldVal := structElem.Field(i)
		if err := setFieldValue(fieldVal, value); err != nil {
			continue
		}
	}

	return structVal, nil
}

func (r *ArgumentResolver) resolveMap(argType reflect.Type, req *http.Request) (reflect.Value, error) {
	mapVal := reflect.MakeMap(argType)

	if req.Form != nil {
		for key, values := range req.Form {
			if len(values) > 0 {
				mapVal.SetMapIndex(reflect.ValueOf(key), reflect.ValueOf(values[0]))
			}
		}
	}

	return mapVal, nil
}

func setFieldValue(fieldVal reflect.Value, value string) error {
	switch fieldVal.Kind() {
	case reflect.String:
		fieldVal.SetString(value)
	case reflect.Int:
		if intVal, err := strconv.Atoi(value); err == nil {
			fieldVal.SetInt(int64(intVal))
		}
	case reflect.Int64:
		if intVal, err := strconv.ParseInt(value, 10, 64); err == nil {
			fieldVal.SetInt(intVal)
		}
	case reflect.Float64:
		if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
			fieldVal.SetFloat(floatVal)
		}
	case reflect.Bool:
		fieldVal.SetBool(value == "true" || value == "1")
	case reflect.Slice:
		if fieldVal.Type().Elem().Kind() == reflect.String {
			fieldVal.Set(reflect.ValueOf([]string{value}))
		}
	}
	return nil
}