package gox

import "reflect"

// Merge combines multiple styles. Later values override earlier ones.
// Non-zero fields in later styles overwrite fields in earlier styles.
//
//	gox.Merge(styles["base"], styles["active"])
func Merge(styles ...Style) Style {
	var result Style
	for _, s := range styles {
		mergeStyle(&result, &s)
	}
	return result
}

// When returns ifTrue if condition is true, ifFalse otherwise.
// Generic conditional helper for use in view expressions.
//
//	gox.When(active, 1.0, 0.5)
func When[T any](condition bool, ifTrue T, ifFalse T) T {
	if condition {
		return ifTrue
	}
	return ifFalse
}

// mergeStyle copies non-zero fields from src into dst.
func mergeStyle(dst, src *Style) {
	dv := reflect.ValueOf(dst).Elem()
	sv := reflect.ValueOf(src).Elem()

	for i := range dv.NumField() {
		sf := sv.Field(i)
		if !sf.IsZero() {
			dv.Field(i).Set(sf)
		}
	}
}
