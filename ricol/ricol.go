package ricol

type Ricol struct {
	filePath string
}

func NewRicol(filePath string) *Ricol {
	return &Ricol{
		filePath: filePath,
	}
}

func (r *Ricol) Run() error {
	return nil
}
