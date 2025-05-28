package routes
import (
	"net/http"
	"littleeinsteinchildcare/backend/internal/handlers"
)

func RegisterEmailRoutes(routes *http.ServeMux, emailHandler *handlers.EmailHandler) {
	routes.HandleFunc("POST /api/send-invite", emailHandler.SendInvite)
}
