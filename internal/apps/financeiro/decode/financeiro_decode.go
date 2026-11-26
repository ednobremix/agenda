package decode

import (
	"encoding/json"
	"errors"
	"massage-scheduler/internal/apps/financeiro/models"
	common_models "massage-scheduler/internal/common/models"
	"net/http"
)

func ParseRequiredEmailParam(r *http.Request) (string, error) {
	email := r.URL.Query().Get("email")
	if email == "" {
		return "", errors.New("Email missing")
	}
	return email, nil
}

func ParseOptionalEmailParam(r *http.Request) string {
	return r.URL.Query().Get("email")
}

func ParseRequiredIDParam(r *http.Request) (string, error) {
	id := r.URL.Query().Get("id")
	if id == "" {
		return "", errors.New("ID missing")
	}
	return id, nil
}

func ParseOptionalIDParam(r *http.Request) string {
	return r.URL.Query().Get("id")
}

func ParseDeleteUserParams(r *http.Request) (email, currentUserEmail string, err error) {
	email = r.URL.Query().Get("email")
	if email == "" {
		return "", "", errors.New("Email missing")
	}
	return email, r.URL.Query().Get("currentUserEmail"), nil
}

func DecodeExpense(r *http.Request) (common_models.Expense, error) {
	var expense common_models.Expense
	if err := json.NewDecoder(r.Body).Decode(&expense); err != nil {
		return common_models.Expense{}, err
	}
	if expense.UserEmail == "" {
		return common_models.Expense{}, errors.New("User email missing")
	}
	return expense, nil
}

func DecodeIncome(r *http.Request) (common_models.Income, error) {
	var income common_models.Income
	if err := json.NewDecoder(r.Body).Decode(&income); err != nil {
		return common_models.Income{}, err
	}
	if income.UserEmail == "" {
		return common_models.Income{}, errors.New("User email missing")
	}
	return income, nil
}

func DecodeBank(r *http.Request) (common_models.Bank, error) {
	var bank common_models.Bank
	if err := json.NewDecoder(r.Body).Decode(&bank); err != nil {
		return common_models.Bank{}, err
	}
	return bank, nil
}

func DecodePaymentType(r *http.Request) (common_models.PaymentType, error) {
	var pt common_models.PaymentType
	if err := json.NewDecoder(r.Body).Decode(&pt); err != nil {
		return common_models.PaymentType{}, err
	}
	return pt, nil
}

func DecodePayment(r *http.Request) (common_models.UserPayment, error) {
	var payment common_models.UserPayment
	if err := json.NewDecoder(r.Body).Decode(&payment); err != nil {
		return common_models.UserPayment{}, err
	}
	return payment, nil
}

func DecodeUpdatePayment(r *http.Request) (models.UpdatePaymentInput, error) {
	var input models.UpdatePaymentInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		return models.UpdatePaymentInput{}, err
	}
	return input, nil
}

func DecodeUpdateUserPermission(r *http.Request) (models.UpdateUserPermissionInput, error) {
	var input models.UpdateUserPermissionInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		return models.UpdateUserPermissionInput{}, err
	}
	return input, nil
}

func DecodeUpdateUserAccess(r *http.Request) (models.UpdateUserAccessInput, error) {
	var input models.UpdateUserAccessInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		return models.UpdateUserAccessInput{}, err
	}
	return input, nil
}
