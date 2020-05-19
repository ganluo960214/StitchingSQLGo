package StitchingSQLGo

type Returning SQLFields

func (r Returning) Returning(s *sql) error {
	if s == nil {
		return ErrorNilSQL
	}
	if r == nil {
		return nil
	}

	s.WriteString(" returning")

	return SQLFields(r).sqlFields(s)
}
