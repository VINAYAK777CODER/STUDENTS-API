package student // package groups all student-related API handlers

/*
   ---------------------------------------------------------
   IMPORTS
   ---------------------------------------------------------
   - encoding/json → decode JSON request body into Go struct
   - errors        → used to check specific errors (like io.EOF)
   - fmt           → formatting messages
   - io            → used for detecting empty request body (io.EOF)
   - slog          → structured logging (new standard logger)
   - net/http      → for HTTP handler, status codes

   - types         → your custom Student struct (from internal/types)
   - response      → custom helper for sending JSON responses
   - validator/v10 → for struct validation (required fields etc.)
*/
import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/VINAYAK777CODER/STUDENTS-API/internal/types"
	"github.com/VINAYAK777CODER/STUDENTS-API/internal/utils/response"
	"github.com/go-playground/validator/v10"
)

/*
New()
-------------------------------------------------------------

	PURPOSE:
	  → Returns an http.HandlerFunc
	  → This handler will process "create student" API requests.

	WHY RETURN A FUNCTION?
	  → Useful pattern to add dependencies later (DB, services…)
	  → Example: func New(db *sql.DB) http.HandlerFunc

	RETURN VALUE:
	  func(w http.ResponseWriter, r *http.Request)
*/
func New() http.HandlerFunc {

	// This anonymous function IS the real request handler
	return func(w http.ResponseWriter, r *http.Request) {

		// Log API call (server console)
		slog.Info("creating a student api")

		/*
		   STEP 1:
		   Create a student variable that will store the JSON body.

		   types.Student:
		     - Your custom struct
		     - It will receive values according to JSON keys sent by client
		*/
		var student types.Student

		/*
		   STEP 2:
		   Decode JSON request body into "student" struct.
		   json.NewDecoder(r.Body) reads raw JSON from the HTTP request.

		   Decode(&student):
		     - Converts JSON → Go struct
		     - Fills student.Name, student.Age, student.Email, etc.

		   POSSIBLE ERRORS:
		     - io.EOF → body is empty ({} or nothing)
		     - invalid JSON format → {"name":123}
		     - wrong types
		*/
		err := json.NewDecoder(r.Body).Decode(&student)

		/*
		   STEP 3: Handle EMPTY BODY
		   --------------------------------------------------
		   - If the client sends empty request body
		   - json.Decode() returns io.EOF error
		   - errors.Is(err, io.EOF) checks exact error type
		*/
		if errors.Is(err, io.EOF) {

			// Send nice JSON error
			response.WriteJson(
				w,
				http.StatusBadRequest,
				response.GeneralError(fmt.Errorf("empty body")),
			)
			return // STOP further execution
		}

		/*
		   STEP 4: Handle ANY OTHER JSON PARSING ERROR
		   --------------------------------------------------
		   Examples:
		     - Missing commas
		     - Wrong JSON syntax
		     - Type mismatch
		*/
		if err != nil {
			response.WriteJson(
				w,
				http.StatusBadRequest,
				response.GeneralError(err),
			)
			return
		}

		/*
		   STEP 5: STRUCT VALIDATION USING validator/v10
		   --------------------------------------------------
		   - Student struct likely contains tags like:
		         Name  string `validate:"required"`
		         Age   int    `validate:"required"`
		   - validator.New().Struct(student)
		         → checks all tags
		         → returns error if validation fails
		*/
		if err := validator.New().Struct(student); err != nil {

			// Convert validation errors into readable JSON
			validateErrs := err.(validator.ValidationErrors)

			response.WriteJson(
				w,
				http.StatusBadRequest,
				response.ValidationError(validateErrs),
			)
			return
		}

		/*
		   STEP 6: SUCCESS RESPONSE
		   --------------------------------------------------
		   - No JSON decode error
		   - No validation error
		   - So we return HTTP status 201 (Created)
		   - Body is a simple JSON map: {"Success":"ok"}
		*/
		response.WriteJson(w, http.StatusCreated, map[string]string{
			"Success": "ok",
		})
	}
}
