package routes

import (
	"fmt"
	"net/http"

	"github.com/nuric/go-api-template/utils"
)

func SetupRoutes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /greetings/{firstName}", GreetingHandler)
	return mux
}

/* Key things to note:
- Request and response types are nearby the handler for easy debugging, you know
what is coming and going.
- We validate the request explicitly so it knows what is expected.
*/

type GreetingRequest struct {
	LastName string `json:"lastName"`
}

type GreetingResponse struct {
	Greeting string `json:"greeting"`
}

func (r GreetingRequest) Validate() error {
	if r.LastName == "" {
		return fmt.Errorf("last name is required")
	}
	return nil
}

func GreetingHandler(w http.ResponseWriter, r *http.Request) {
	req, err := utils.DecodeValid[GreetingRequest](r)
	/* You'll find this error checking pattern repeats in all handlers. It's not
	 * that much if it really annoys you then you can do it inside the
	 * DecodeValid function. Either way is fine. I prefer explicit error checking
	 * in case we can do something else. */
	if err != nil {
		utils.Encode(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	// Do path validation here.
	firstName := r.PathValue("firstName")
	if len(firstName) < 2 {
		utils.Encode(w, http.StatusBadRequest, map[string]string{"error": "first name must be at least 2 characters"})
		return
	}
	// Construct the response and encode as JSON.
	response := GreetingResponse{
		Greeting: fmt.Sprintf("Hello, %s %s!", firstName, req.LastName),
	}
	utils.Encode(w, http.StatusOK, response)
}
