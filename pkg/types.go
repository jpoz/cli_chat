package clichat

type (
	errMsg error

	AIMsg      struct{ text string }
	AgentMsg   struct{ text string }
	BackendMsg struct{ text string }

	Message struct {
		Sender string
		Text   string
		Input  string
	}

	Messages []Message

	MessageContext struct {
		Current Message
		History []Message
	}

	orderItem struct {
		ProductDesc string `json:"product_desc"`
		Quantity    int    `json:"quantity"`
	}

	order struct {
		OrderNumber string      `json:"order_number"`
		Items       []orderItem `json:"items"`
		OrderDate   string      `json:"order_date"`
	}

	orderResults struct {
		ResultsFor string  `json:"results_for"`
		LookupBy   string  `json:"lookup_by"`
		Orders     []order `json:"orders"`
	}

	user struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Email     string `json:"email"`
		Phone     string `json:"phone"`
	}

	userResult struct {
		LookupUserBy string `json:"lookup_user_by"`
		User         *user  `json:"user"`
	}
)
