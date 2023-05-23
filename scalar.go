package stdgraph

import (
	"fmt"
	"time"
)

type Date struct {
	time.Time
}

func (Date) ImplementsGraphQLType(name string) bool {
	return name == "Date"
}

func (ts *Date) UnmarshalGraphQL(input interface{}) error {
	switch t := input.(type) {
	case string:
		tt, err := time.Parse("2006-01-02", t)
		if err != nil {
			return err
		}

		ts.Time = tt
	default:
		return fmt.Errorf("unknown Date unmarshal type: %T", t)
	}

	return nil
}

type DateTime struct {
	time.Time
}

func (DateTime) ImplementsGraphQLType(name string) bool {
	return name == "DateTime"
}

func (ts *DateTime) UnmarshalGraphQL(input interface{}) error {
	switch t := input.(type) {
	case string:
		tt, err := time.Parse("2006-01-02T15:04:05Z", t)
		if err != nil {
			return err
		}

		ts.Time = tt
	default:
		return fmt.Errorf("unknown DateTime unmarshal type: %T", t)
	}

	return nil
}
