package models

type UpdateUserAccessRequest struct {
	Email         string `json:"email"`
	HasAgenda     bool   `json:"has_agenda"`
	HasFinanceiro bool   `json:"has_financeiro"`
	HasEstoque    bool   `json:"has_estoque"`
}

type UpdatePaymentRequest struct {
	ID     any    `json:"id"`
	Status string `json:"status"`
}

type UpdateUserPermissionRequest struct {
	Email            string `json:"email"`
	MaxProfessionals int    `json:"max_professionals"`
	FreeScheduling   *bool  `json:"free_scheduling"`
}

type GoogleAuthRequest struct {
	Credential string `json:"credential"`
}

type LinkGoogleRequest struct {
	Credential string `json:"credential"`
}
