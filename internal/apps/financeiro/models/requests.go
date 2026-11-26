package models

type UpdatePaymentInput struct {
	ID     uint   `json:"id"`
	Status string `json:"status"`
}

type UpdateUserPermissionInput struct {
	UserID        uint   `json:"user_id"`
	HasFinanceiro bool   `json:"has_financeiro"`
	HasEstoque    bool   `json:"has_estoque"`
	Role          string `json:"role"`
}

type UpdateUserAccessInput struct {
	Email         string `json:"email"`
	HasAgenda     bool   `json:"has_agenda"`
	HasFinanceiro bool   `json:"has_financeiro"`
	HasEstoque    bool   `json:"has_estoque"`
}
