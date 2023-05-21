package utils

import (
	"context"
	"eltneg/goliltemp/src/config"
	"eltneg/goliltemp/src/dependencies"
	"encoding/json"
	"errors"
	"net/http"
	"reflect"
	"strings"

	validator "github.com/go-playground/validator/v10"
	log "github.com/sirupsen/logrus"
)

type RequestDataModel[R any] interface {
	Controller(ctx context.Context, cfg *config.Config, dpc *dependencies.Dependencies) (status int, msg string, data R, err error)
}

type HandlerMiddleware = func(ctx context.Context, cfg *config.Config, r *http.Request) (context.Context, error)

type H[T RequestDataModel[R], R any] struct{}

func (h *H[M, R]) Handler(cfg *config.Config, d *dependencies.Dependencies, data M, middlewares ...HandlerMiddleware) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		var reqData M

		for _, m := range middlewares {
			_ctx, err := m(ctx, cfg, r)
			if err != nil {
				JSONResponse(w, http.StatusBadRequest, &Response{Success: false, Message: err.Error(), Data: nil})
				return
			}
			ctx = _ctx
		}

		if r.Method == http.MethodGet || r.Method == http.MethodDelete {
			log.Debug("Processing request params")
			queryData := map[string]interface{}{}
			queryParams := h.getQueryKeys(data)
			for _, k := range queryParams {
				queryData[k] = r.URL.Query().Get(k)
			}
			log.WithField("query data", queryData).Info("query data from request")

			jsonBody, err := json.Marshal(queryData)

			if err != nil {
				JSONResponse(w, http.StatusBadRequest, &Response{Success: false, Message: "Encode query error: " + err.Error(), Data: nil})
				return
			}
			if err := json.Unmarshal(jsonBody, &reqData); err != nil {
				log.WithError(err).Error("DecodingRequestError")
				JSONResponse(w, http.StatusBadRequest, &Response{Success: false, Message: "Decode query error: " + err.Error(), Data: nil})
				return
			}

		} else {
			log.Debug("Processing request body")
			if err := DecodeJSONBody(w, r, &reqData); err != nil {
				log.WithError(err).Error("DecodingRequestError")
				JSONResponse(w, http.StatusBadRequest, &Response{Success: false, Message: "Decode req body error: " + err.Error(), Data: nil})
				return
			}
		}

		log.WithField("Request data", reqData).Debug("Processed request param or body")

		validate := validator.New()
		err := validate.Struct(reqData)
		log.WithError(err).WithField("reqdata", reqData).Info("validation error")
		if err != nil {
			var errs []error
			if err, ok := err.(*validator.InvalidValidationError); ok {
				errs = append(errs, errors.New("validation error: "+err.Error()))
				JSONResponse(w, http.StatusBadRequest, &Response{Success: false, Message: "invalid request data", Data: ResponseErrors{Errors: CreateErrorResponse(errs...)}})
				return
			}
			for _, err := range err.(validator.ValidationErrors) {
				errs = append(errs, errors.New("validation error: invalid "+err.StructNamespace()))
			}
			JSONResponse(w, http.StatusBadRequest, &Response{Success: false, Message: "invalid request data", Data: ResponseErrors{Errors: CreateErrorResponse(errs...)}})
			return
		}

		status, msg, data, err := reqData.Controller(ctx, cfg, d)
		if err != nil {
			log.WithError(err).Error("Request failed")
			JSONResponse(w, status, &Response{Success: false, Message: msg, Data: ResponseErrors{Errors: CreateErrorResponse(err)}})
			return
		}

		if status >= 400 {
			log.WithError(err).Error("Request failed")
			JSONResponse(w, status, &Response{Success: false, Message: msg, Data: nil})
			return
		}

		if status == http.StatusNoContent {
			JSONResponse(w, status, nil)
			return
		}

		JSONResponse(w, status, &Response{Success: true, Message: msg, Data: data})
	}
}

func (h *H[M, R]) getQueryKeys(d M) []string {
	val := reflect.Indirect(reflect.ValueOf(d))
	queryKeys := []string{}
	for i := 0; i < val.Type().NumField(); i++ {
		t := val.Type().Field(i)
		fieldName := t.Name

		switch jsonTag := t.Tag.Get("json"); jsonTag {
		case "-":
		case "":
			queryKeys = append(queryKeys, fieldName)
		default:
			parts := strings.Split(jsonTag, ",")
			name := parts[0]
			if name == "" {
				name = fieldName
			}
			queryKeys = append(queryKeys, name)
		}
	}
	return queryKeys
}
