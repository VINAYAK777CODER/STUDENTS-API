package response // response package holds helper methods to send JSON response

/*
   ---------------------------------------------------------
   IMPORTS
   ---------------------------------------------------------
   - encoding/json → used to encode Go structs or maps into JSON.
   - fmt           → used for building formatted error messages.
   - net/http      → used to set headers & manage HTTP response codes.
   - strings       → used to join error messages for validation.
   - validator/v10 → used to detect validation errors returned by validator.
*/
import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
)

/*
Response STRUCT
-------------------------------------------------------------
   - This struct defines how an error response will look in JSON.
   - Fields are exported (capital letter) so JSON encoder can access them.
   - json:"status" → key inside the JSON output will be "status".
   - json:"error"  → key inside JSON output will be "error".
*/
type Response struct {
	Status string `json:"status"`
	Error  string `json:"error"`
}

/*
CONSTANT VALUES
-------------------------------------------------------------
   - Constant predefined status messages.
   - Avoids hardcoding "OK" or "Error" everywhere.
*/
const (
	StatusOk    = "OK"
	StatusError = "Error"
)

/*
WriteJson()
-------------------------------------------------------------
   PURPOSE:
     → Converts any Go value (struct/map/string) into JSON.
     → Writes JSON to the HTTP response.
     → Sets proper headers.
     → Writes desired HTTP status code.

   PARAMETERS:
     - w      : http.ResponseWriter → used to send output to client.
     - status : integer → HTTP status code (200, 201, 400 etc.)
     - data   : interface{} → any data you want to send as JSON.
*/
func WriteJson(w http.ResponseWriter, status int, data interface{}) error {

	// 1. Set header so browser/Postman knows data is JSON
	w.Header().Set("Content-Type", "application/json")

	// 2. Must write HTTP status before writing the body
	w.WriteHeader(status)

	/*
	   3. Encode the data into JSON:
	      json.NewEncoder(w).Encode(data)
	      - Converts "data" into JSON.
	      - Writes JSON directly into HTTP response.
	      - If any error occurs, Encode returns an error.
	*/
	return json.NewEncoder(w).Encode(data)
}

/*
GeneralError()
-------------------------------------------------------------
   PURPOSE:
     → Prepare a standard JSON error response.

   INPUT:
     - err → error object

   RETURN:
     → Response struct:
       {
         "status": "Error",
         "error": "<error message>"
       }
*/
func GeneralError(err error) Response {
	return Response{
		Status: StatusError,
		Error:  err.Error(), // converts actual error into string
	}
}

/*
ValidationError()
-------------------------------------------------------------
   PURPOSE:
     → Converts validator.ValidationErrors into readable JSON.

   FLOW:
     - Loop through all validation errors.
     - Check validation type using err.ActualTag() (“required”, “email”, etc.)
     - Build a user-friendly error message.
     - Combine all messages into a single string.

   RETURNS:
     Response{
         Status: "Error",
         Error:  "field X is required, field Y is invalid"
     }
*/
func ValidationError(errs validator.ValidationErrors) Response {
	var errMsg []string // slice to collect all error messages

	for _, err := range errs {
		switch err.ActualTag() {

		// If struct tag validation = required
		case "required":
			errMsg = append(errMsg,
				fmt.Sprintf("field %s is required field", err.Field()))

		// For all other validation types
		default:
			errMsg = append(errMsg,
				fmt.Sprintf("field %s is invalid", err.Field()))
		}
	}

	// Join messages into single string:  "msg1, msg2, msg3"
	return Response{
		Status: StatusError,
		Error:  strings.Join(errMsg, ", "),
	}
}
