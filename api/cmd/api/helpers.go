//Filename: cmd/api/helpers.go

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"remindMe/federicorosado.net/internal/validator"

	"github.com/julienschmidt/httprouter"
)

func (app *application) readIDParam(r *http.Request) (int64, error) {
	//Use the "ParamsFromContext()" function to get the request context as a slice
	params := httprouter.ParamsFromContext(r.Context())
	//Get the value of the "id" parameter
	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	if err != nil || id < 1 {
		return 0, errors.New("invalid id parameter")
	}
	return id, nil
}

//Define a new type named envelope
type envelope map[string]interface {
}

func (app *application) writeJSON(w http.ResponseWriter, status int, data envelope, headers http.Header) error {
	//Convert our map into a JSON object
	// js, err := json.Marshal(data)
	js, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}
	//Add a newline to make viewing on the terminal easier
	js = append(js, '\n')
	//Add the headers
	for key, value := range headers {
		w.Header()[key] = value
	}
	//Specify that we will server our responses using JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	//Write the []byte slice containing the JSON response body
	w.Write(js)
	return nil
}

//Helper to help decode json
func (app *application) readJSON(w http.ResponseWriter, r *http.Request, dst interface{}) error {
	//Use http.Max.BytesReader() to limit the size of the request body to
	//1 MB 2^20
	maxBytes := 1_048_576

	//Decode the request body into the target distination
	// err := json.NewDecoder(r.Body).Decode(dst)
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	err := dec.Decode(dst)
	//check for a bad request
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError

		//switch to check for erros
		switch {
		//check for syntax error
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains badly-formed JSON(at character %d)", syntaxError.Offset)
		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly-formed JSON")
		//Check for wrong types passed by the client
		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("body contains incorrect JSON type (at character %d)", unmarshalTypeError.Offset)
		//Empty Body
		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")

		//Unmappable fields
		case strings.HasPrefix(err.Error(), "json: unkown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unkown field")
			return fmt.Errorf("body contains unkown key %s", fieldName)
		//Too large
		case err.Error() == "http: request body too large":
			return fmt.Errorf("body must not be larger than %d bytes", maxBytes)

		//Pass non-nil pointer error
		case errors.As(err, &invalidUnmarshalError):
			panic(err)
		//default
		default:
			return err
		}
	}
	//Call decode again
	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("body must only contain a single JSON value")
	}
	return nil
}

// the readString()  method returns a string value from the query parameters
//string or returns a default value if no matching key is found
func (app *application) readString(qs url.Values, key string, defaultValue string) string {
	// Get the value
	value := qs.Get(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// The readCSV() method splits a value into a slice based on the comma separator
// If no matching key is found then the default value is returned.
func (app *application) readCSV(qs url.Values, key string, defaultValue []string) []string {
	// Get the value
	value := qs.Get(key)
	if value == "" {
		return defaultValue
	}
	// Split the string based on the ',' delimeter
	return strings.Split(value, ",")
}

// the readInt() method converts a string value from the query string to an integer value
// If the value cannot be converted to an integer then a validation error is added to the validation errors map
func (app *application) readInt(qs url.Values, key string, defaultValue int, v *validator.Validator) int {
	//Get the value
	value := qs.Get(key)
	if value == "" {
		return defaultValue
	}
	//perform the conversiomn to an int
	intValue, err := strconv.Atoi(value)
	if err != nil {
		v.AddError(key, "must ve an integer value")
		return defaultValue
	}
	return intValue
}
