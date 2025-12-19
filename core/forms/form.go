package forms

// RequestForm is implemented by forms that create/update models
// T is the model type
type RequestForm[T any] interface {
	// ToModel creates a new model from the form data
	ToModel() *T

	// ApplyTo applies the form data to an existing model
	ApplyTo(model *T)
}

// ResponseForm is implemented by forms that serialize models for API responses
// T is the model type
type ResponseForm[T any] interface {
	// FromModel populates the form from a model
	FromModel(model *T) ResponseForm[T]
}

// CreateForm is a convenience interface combining request form requirements
type CreateForm[T any] interface {
	RequestForm[T]
	Validatable
}

// UpdateForm is a convenience interface for partial updates
type UpdateForm[T any] interface {
	// ApplyTo applies only the non-nil fields to the model
	ApplyTo(model *T)
	Validatable
}

// Validatable can be validated
type Validatable interface {
	Validate() error
}

// Mapper provides bidirectional mapping between forms and models
type Mapper[F any, M any] interface {
	ToModel(form *F) *M
	ToForm(model *M) *F
	ToModels(forms []*F) []*M
	ToForms(models []*M) []*F
}
