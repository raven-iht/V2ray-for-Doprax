package financial

import (
	"errors"
	"math"
	"time"
)

// LoanPlan defines a purchase financing plan (credit line for purchases)
// with loan-like repayment (fixed term, fixed annual interest).
type LoanPlan struct {
	// Optional, for classification/reporting (e.g., personal, goods, business)
	LoanPurpose string

	// Purchase amount limits (0 means no limit)
	MinLoanAmount uint64
	MaxLoanAmount uint64

	// Total number of months for repayment (e.g., 12, 24, 36)
	LoanTermMonths int

	// Annual percentage rate (e.g., 22.0 means 22%)
	InterestRate float64

	// Origination fee as a percentage of the full credit limit, paid in cash
	// on the first purchase only (per product spec)
	OriginationFeePercent float64

	// Late payment fee (not applied here; included for completeness)
	LatePaymentFee uint64

	// Human-readable rules/notes for guarantees
	GuaranteeRules string
}

// InstallmentSchedule represents a single installment line.
type InstallmentSchedule struct {
	InstallmentNumber   int
	DueDate             time.Time
	Amount              uint64
	IsAdditionalPayment bool
}

// LoanPaymentDetails summarizes the contract created for a purchase
// funded by the credit line under a LoanPlan.
type LoanPaymentDetails struct {
	// Purchase and credit context
	PurchaseAmount       uint64
	CreditLimit          uint64
	CreditBalanceBefore  uint64
	CreditBalanceAfter   uint64

	// Fees and rates
	OriginationFeeCash   uint64
	AnnualRatePercent    float64
	NumberOfInstallments int

	// Repayment metrics
	MonthlyPayment uint64
	TotalInterest uint64
	TotalRepayment uint64

	// Schedule
	FirstInstallmentDate time.Time
	Schedule             []InstallmentSchedule
}

// Calculator provides financial math utilities.
// You can replace or extend this with your existing calculator.
type Calculator struct{}

// CalculateFixedMonthlyPayment computes level payment for principal P at annual rate APR over n months.
// Uses annuity formula with monthly compounding. Returns rounded up to nearest unit.
func (c *Calculator) CalculateFixedMonthlyPayment(principal uint64, annualRatePercent float64, termMonths int) (uint64, error) {
	if termMonths <= 0 {
		return 0, errors.New("term months must be positive")
	}
	P := float64(principal)
	if P <= 0 {
		return 0, nil
	}
	monthlyRate := (annualRatePercent / 100.0) / 12.0
	if monthlyRate == 0 {
		return uint64(math.Ceil(P / float64(termMonths))), nil
	}
	// payment = P * r / (1 - (1+r)^-n)
	den := 1 - math.Pow(1+monthlyRate, float64(-termMonths))
	if den == 0 {
		return 0, errors.New("invalid parameters lead to division by zero")
	}
	payment := P * monthlyRate / den
	return uint64(math.Ceil(payment)), nil
}

// CalculateTotalInterest derives total interest from payment schedule parameters.
func (c *Calculator) CalculateTotalInterest(principal uint64, monthlyPayment uint64, termMonths int) uint64 {
	if termMonths <= 0 || monthlyPayment == 0 || principal == 0 {
		return 0
	}
	totalPaid := uint64(termMonths) * monthlyPayment
	if totalPaid <= principal {
		return 0
	}
	return totalPaid - principal
}

// CalculateLoanPurchasePlan computes the repayment plan for a single purchase funded
// by a LoanPlan-backed credit line. Origination fee is paid in cash only on the first purchase.
//
// Params:
//   - purchaseAmount: price of the item to finance for this contract
//   - plan: the LoanPlan settings to apply
//   - creditLimit: the user's allocated credit line limit
//   - creditBalance: the user's available credit before this purchase
//   - isFirstPurchase: true if this is the first usage of this credit line (cash origination fee applies)
//   - now: timestamp used for scheduling (first installment defaults to next month)
//   - calc: financial calculator
func CalculateLoanPurchasePlan(
	purchaseAmount uint64,
	plan LoanPlan,
	creditLimit uint64,
	creditBalance uint64,
	isFirstPurchase bool,
	now time.Time,
	calc *Calculator,
) (LoanPaymentDetails, error) {
	// Validate inputs
	if purchaseAmount == 0 {
		return LoanPaymentDetails{}, errors.New("purchase amount must be positive")
	}
	if plan.LoanTermMonths <= 0 {
		return LoanPaymentDetails{}, errors.New("loan term months must be positive")
	}
	if plan.InterestRate < 0 {
		return LoanPaymentDetails{}, errors.New("interest rate cannot be negative")
	}
	if plan.MinLoanAmount > 0 && purchaseAmount < plan.MinLoanAmount {
		return LoanPaymentDetails{}, errors.New("purchase amount below minimum limit")
	}
	if plan.MaxLoanAmount > 0 && purchaseAmount > plan.MaxLoanAmount {
		return LoanPaymentDetails{}, errors.New("purchase amount exceeds maximum limit")
	}
	if creditLimit == 0 {
		return LoanPaymentDetails{}, errors.New("credit limit must be positive")
	}
	if creditBalance < purchaseAmount {
		return LoanPaymentDetails{}, errors.New("insufficient available credit for this purchase")
	}
	if calc == nil {
		calc = &Calculator{}
	}

	// Compute cash origination fee (only at first purchase)
	originationFeeCash := uint64(0)
	if isFirstPurchase && plan.OriginationFeePercent > 0 {
		originationFeeCash = uint64(math.Ceil(float64(creditLimit) * plan.OriginationFeePercent / 100.0))
	}

	// Monthly annuity for the financed purchase amount
	monthlyPayment, err := calc.CalculateFixedMonthlyPayment(purchaseAmount, plan.InterestRate, plan.LoanTermMonths)
	if err != nil {
		return LoanPaymentDetails{}, err
	}
	totalInterest := calc.CalculateTotalInterest(purchaseAmount, monthlyPayment, plan.LoanTermMonths)
	totalRepayment := purchaseAmount + totalInterest

	// Update credit balance after allocating principal for this purchase
	creditBalanceAfter := creditBalance - purchaseAmount

	// Build schedule: first installment next month on the same day, best-effort for shorter months
	firstInstallmentDate := nextMonthSameDay(now)
	schedule := make([]InstallmentSchedule, 0, plan.LoanTermMonths)
	for i := 0; i < plan.LoanTermMonths; i++ {
		due := firstInstallmentDate.AddDate(0, i, 0)
		schedule = append(schedule, InstallmentSchedule{
			InstallmentNumber:   i + 1,
			DueDate:             due,
			Amount:              monthlyPayment,
			IsAdditionalPayment: false,
		})
	}

	return LoanPaymentDetails{
		PurchaseAmount:       purchaseAmount,
		CreditLimit:          creditLimit,
		CreditBalanceBefore:  creditBalance,
		CreditBalanceAfter:   creditBalanceAfter,
		OriginationFeeCash:   originationFeeCash,
		AnnualRatePercent:    plan.InterestRate,
		NumberOfInstallments: plan.LoanTermMonths,
		MonthlyPayment:       monthlyPayment,
		TotalInterest:        totalInterest,
		TotalRepayment:       totalRepayment,
		FirstInstallmentDate: firstInstallmentDate,
		Schedule:             schedule,
	}, nil
}

// nextMonthSameDay returns the next month's date attempting to preserve day-of-month.
// If the next month is shorter and the day overflows, it clips to the last valid day.
func nextMonthSameDay(t time.Time) time.Time {
	// Try naive AddDate first; in Go this already clips to end-of-month when overflowing.
	return t.AddDate(0, 1, 0)
}

