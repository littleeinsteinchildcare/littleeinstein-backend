package routes
import (
	"net/http"
	"littleeinsteinchildcare/backend/internal/handlers"
)

func RegisterProtectedEmailRoutes(routes *http.ServeMux, emailHandler *handlers.EmailHandler) {
	routes.HandleFunc("POST /api/send-invite", emailHandler.SendInvite)
}

func RegisterUnprotectedEmailRoutes(routes *http.ServeMux, emailHandler *handlers.EmailHandler) {
	routes.HandleFunc("GET /check-invited", emailHandler.CheckIfInvited)
}
