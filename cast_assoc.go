package changeset

import (
	"reflect"
	"strconv"
	"strings"

	"github.com/Fs02/changeset/params"
)

// CastAssocErrorMessage is the default error message for CastAssoc when its invalid.
var CastAssocErrorMessage = "{field} is invalid"

// CastAssocRequiredMessage is the default error message for CastAssoc when its missing.
var CastAssocRequiredMessage = "{field} is required"

// ChangeFunc is changeset function.
type ChangeFunc func(interface{}, params.Params) *Changeset

// CastAssoc casts association changes using changeset function.
// Repo insert or update won't persist any changes generated by CastAssoc.
func CastAssoc(ch *Changeset, field string, fn ChangeFunc, opts ...Option) {
	options := Options{
		message: CastAssocErrorMessage,
	}
	options.apply(opts)

	sourceField := options.sourceField
	if sourceField == "" {
		sourceField = field
	}

	typ, texist := ch.types[field]
	valid := true
	if texist && ch.params.Exists(sourceField) {
		if typ.Kind() == reflect.Struct {
			valid = castOne(ch, sourceField, field, fn)
		} else if typ.Kind() == reflect.Slice && typ.Elem().Kind() == reflect.Struct {
			valid = castMany(ch, sourceField, field, fn)
		}
	}

	if !valid {
		msg := strings.Replace(options.message, "{field}", field, 1)
		AddError(ch, field, msg)
	}

	_, found := ch.changes[field]
	if options.required && !found {
		options.message = CastAssocRequiredMessage
		msg := strings.Replace(options.message, "{field}", field, 1)
		AddError(ch, field, msg)
	}
}

func castOne(ch *Changeset, fieldSource string, fieldTarget string, fn ChangeFunc) bool {
	par, valid := ch.params.GetParams(fieldSource)
	if !valid {
		return false
	}

	var innerch *Changeset

	if val, exist := ch.values[fieldTarget]; exist && val != nil {
		innerch = fn(val, par)
	} else {
		innerch = fn(reflect.Zero(ch.types[fieldTarget]).Interface(), par)
	}

	ch.changes[fieldTarget] = innerch

	// add errors to main errors
	mergeErrors(ch, innerch, fieldTarget+".")

	return true
}

func castMany(ch *Changeset, fieldSource string, fieldTarget string, fn ChangeFunc) bool {
	spar, valid := ch.params.GetParamsSlice(fieldSource)
	if !valid {
		return false
	}

	data := reflect.Zero(ch.types[fieldTarget].Elem()).Interface()

	chs := make([]*Changeset, len(spar))
	for i, par := range spar {
		innerch := fn(data, par)
		chs[i] = innerch

		// add errors to main errors
		mergeErrors(ch, innerch, fieldTarget+"["+strconv.Itoa(i)+"].")
	}
	ch.changes[fieldTarget] = chs

	return true
}

func mergeErrors(parent *Changeset, child *Changeset, prefix string) {
	for _, err := range child.errors {
		e := err.(Error)
		AddError(parent, prefix+e.Field, e.Message)
	}
}
