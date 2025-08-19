package span

type Span struct {
	From        Position
	To          Position
	SourceLines *[]string
	Source      *string
	Path        *string
}

type Position struct {
	Index     uint
	LineIndex uint
	Lineno    uint
}

func (s *Span) String() string {
	if s.Source == nil {
		return "(no source)"
	}
	return (*s.Source)[s.From.Index:s.To.Index]
}

func (s *Span) Merge(t Span) Span {
	var from, to Position
	if s.From.Index < t.From.Index {
		from = s.From
	} else {
		from = t.From
	}
	if s.To.Index > t.To.Index {
		to = s.To
	} else {
		to = t.To
	}
	return Span{
		From: from,
		To:   to,
	}
}

func Merge(lhs, rhs Span) Span {
	return lhs.Merge(rhs)
}
