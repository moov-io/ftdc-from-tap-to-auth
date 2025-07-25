package paycard

// KernelID represents EMV contactless kernel identifiers
type KernelID int

const (
	Kernel1ID KernelID = iota + 1 // Generic/Visa International
	Kernel2ID                     // Mastercard
	Kernel3ID                     // Visa (US)
	Kernel4ID                     // American Express
	Kernel5ID                     // JCB
	Kernel6ID                     // Discover
	Kernel7ID                     // UnionPay
)
