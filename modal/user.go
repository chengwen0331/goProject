package modal

// Define the response structure for the product
type User struct {
	CompanyID   int8   `json:"company_id"`
	Permission  string `json:"permission"`
	Email       string `json:"email"`
	Address     string `json:"address"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	Nationality string `json:"nationality"`
	Role        string `json:"role"`
	City        string `json:"city"`
	Gender      string `json:"gender"`
	Phone       string `json:"phone"`
	Status      bool   `json:"status"`
	Zipcode     string `json:"zipcode"`
}
